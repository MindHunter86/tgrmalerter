package app

import "database/sql"

import tsql "mailru/tgrmalerter/core/sql"
import "mailru/tgrmalerter/core/config"

import "github.com/rs/zerolog"
import "github.com/gorilla/mux"
import "github.com/go-telegram-bot-api/telegram-bot-api"


/// GLOBAL TODO:
// -/+ refactor with global variables
// -/+ add prefixes for logger
// - check Content Type
// - check Accept Header
// -/+ check type and link in the body of the request
// + refactor apiErrors and httpRequest (apiErrors must be as a part of httpRequest). Create api_request.go
// + check api structs on NIL pointers
// -/+ added method for updating dispatch status
// + fix panic on httpapi Content-Type set
// - errors refactor; add telegram errors
// - update status code in httpRequest struct after successfull operation; refactor with new changes


var (
	globSqlDB *sql.DB
	globLogger *zerolog.Logger
	globConfig *config.CoreConfig
	globTgApi *tgrmApi
	globTgDispatcher *tgrmDispatcher
)


type App struct {}

// application public API:
func (m *App) SetSqlDriver(d tsql.SqlDriver) *App { globSqlDB = d.GetRawDBSession(); return m }
func (m *App) SetConfig(c *config.CoreConfig) *App { globConfig = c; return m }
func (m *App) SetLogger(l *zerolog.Logger) *App { globLogger = l; return m }
func (m *App) NewHttpApi() *mux.Router { return new(api).getMuxRouter() }

func (m *App) SetTgBotApi(t *tgbotapi.BotAPI) *App {
	globTgApi = new(tgrmApi).construct(t)
	return m
}

func (m *App) Construct() (*App, error) {
	globTgDispatcher = &tgrmDispatcher {
		pool: make(chan chan *tgrmJob, globConfig.Base.Telegram.Queue.Worker_Capacity),
		jobQueue: make(chan *tgrmJob, globConfig.Base.Telegram.Queue.Chain_Buffer),
		done: make(chan struct{}, 1),
		workerDone: make(chan struct{}, 1) }
	return m,nil
}

func (m *App) Bootstrap() error {
	globTgDispatcher.bootstrap(
		globConfig.Base.Telegram.Queue.Workers, globConfig.Base.Telegram.Queue.Worker_Capacity)
	return nil
}

func (m *App) Destruct() error {
	globTgDispatcher.destruct()
	return nil
}

func (m *App) TelegramUpdateHandler(up *tgbotapi.Update) {
	globLogger.Debug().Str("message", up.Message.Text).Msg("[APP]: TelegramUpdateHandler has been triggered!")

	switch {
	case up.Message == nil:
		globLogger.Warn().Int("update_id", up.UpdateID).Msg("[APP]: An empty message has been received!")
		return
	case up.Message.IsCommand():
		globTgApi.commandRouter(up.Message)
	case up.Message.Contact != nil:
		globTgApi.registerContact(up.Message.Chat.ID, up.Message.From.ID, up.Message.Contact)
	default:
		globLogger.Debug().Str("msg", up.Message.Text).Str("author", up.Message.From.String()).Msg("[APP]: TelegramUpdateHandler: undefined message")
	}
}
