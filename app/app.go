package app

import "database/sql"

import "github.com/rs/zerolog"
import "github.com/gorilla/mux"

import tsql "mailru/tgrmalerter/core/sql"
import "mailru/tgrmalerter/core/config"

import "github.com/go-telegram-bot-api/telegram-bot-api"



/// GLOBAL TODO:
// -/+ refactor with global variables
// -/+ add prefixes for logger
// - check Content Type
// - check Accept Header
// - check type and link in the body of the request
// + refactor apiErrors and httpRequest (apiErrors must be as a part of httpRequest). Create api_request.go
// ! check api structs on NIL pointers


type App struct {
	log *zerolog.Logger
	conf *config.CoreConfig
	sqldb *sql.DB
	tbotApi *tgrmApi
	tgDispatcher *tgrmDispatcher
}

var (
	globSqlDB *sql.DB
	globLogger *zerolog.Logger
	globConfig *config.CoreConfig
	globTgJobChan chan *tgrmJob
)

type prefixLogger struct {}
func (m prefixLogger) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	msg = "fuck you"
	e.Str("wtf", msg)
}

// application public API:
func (m *App) SetConfig(c *config.CoreConfig) *App { m.conf = c; globConfig = c; return m }
func (m *App) NewApplicationApi() *mux.Router { return new(api).getMuxRouter() }

func (m *App) SetSqlDriver(d tsql.SqlDriver) *App {
	globSqlDB = d.GetRawDBSession()
	m.sqldb = d.GetRawDBSession()
	return m
}

func (m *App) SetLogger(l *zerolog.Logger) *App {
	globLogger = l
	m.log = l; return m
}

func (m *App) SetTBot(t *tgbotapi.BotAPI) *App {
	m.tbotApi = new(tgrmApi).construct(t)
	m.tgDispatcher.tgApi = m.tbotApi
	return m
}

func (m *App) Construct() (*App, error) {
	m.tgDispatcher = &tgrmDispatcher{
		tgApi: nil,
		pool: make(chan chan *tgrmJob, m.conf.Base.Telegram.Queue.Worker_Capacity),
		jobQueue: make(chan *tgrmJob, m.conf.Base.Telegram.Queue.Chain_Buffer),
		done: make(chan struct{}, 1),
		workerDone: make(chan struct{}, 1) }
	globTgJobChan = m.tgDispatcher.getQueueChan()
	return m,nil
}
func (m *App) Bootstrap() error {
	m.tgDispatcher.bootstrap(
		m.conf.Base.Telegram.Queue.Workers, m.conf.Base.Telegram.Queue.Worker_Capacity)
	return nil
}
func (m *App) Destruct() error {
	m.tgDispatcher.destruct()
	return nil
}

func (m *App) TelegramUpdateHandler(up *tgbotapi.Update) {
	m.log.Debug().Str("message", up.Message.Text).Msg("TelegramUpdateHandler has been triggered!")

	if up.Message == nil {
		m.log.Warn().Int("update_id", up.UpdateID).Msg("An empty message has been received!"); return }

	if up.Message.IsCommand() {
		m.tbotApi.commandRouter(up.Message)
	} else if up.Message.Contact != nil {
		m.tbotApi.registerContact(up.Message.Chat.ID, up.Message.From.ID, up.Message.Contact)
	} else {
		m.log.Debug().Str("msg", up.Message.Text).Str("author", up.Message.From.String()).Msg("TelegramUpdateHandler: undefined message")
	}

	//
}


// internal methods:
