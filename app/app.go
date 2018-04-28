package app

import "database/sql"

import "github.com/rs/zerolog"
import "github.com/gorilla/mux"

import tsql "mailru/tgrmalerter/core/sql"
import "mailru/tgrmalerter/core/config"

import "github.com/go-telegram-bot-api/telegram-bot-api"

type App struct {
	log *zerolog.Logger
	conf *config.CoreConfig
	sqldb *sql.DB
	tbotApi *tgrmApi
	tgDispatcher *tgrmDispatcher
}


// application public API:
func (m *App) SetLogger(l *zerolog.Logger) *App { m.log = l; return m }
func (m *App) SetConfig(c *config.CoreConfig) *App { m.conf = c; return m }
func (m *App) SetSqlDriver(d tsql.SqlDriver) *App { m.sqldb = d.GetRawDBSession(); return m }
func (m *App) NewApplicationApi() *mux.Router { return new(api).setApp(m).getMuxRouter() }
func (m *App) SetTBot(t *tgbotapi.BotAPI) *App { m.tbotApi = new(tgrmApi).setTBot(t).setLogger(m.log); return m }

func (m *App) Construct() (*App, error) {
	m.tgDispatcher = &tgrmDispatcher{
		pool: make(chan chan *tgrmJob, m.conf.Base.Telegram.Queue.Worker_Capacity),
		jobQueue: make(chan *tgrmJob),
		done: make(chan struct{}, 1),
		workerDone: make(chan struct{}, 1) }
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
