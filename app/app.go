package app

import "database/sql"

import "github.com/rs/zerolog"
import "github.com/gorilla/mux"

import tsql "mailru/tgrmalerter/core/sql"
import "mailru/tgrmalerter/core/config"

type App struct {
	log *zerolog.Logger
	conf *config.CoreConfig
	sqldb *sql.DB
}


// application public API:
func (m *App) SetLogger(l *zerolog.Logger) *App { m.log = l; return m }
func (m *App) SetConfig(c *config.CoreConfig) *App { m.conf = c; return m }
func (m *App) SetSqlDriver(d tsql.SqlDriver) *App { m.sqldb = d.GetRawDBSession(); return m }

func (m *App) NewApplicationApi(d tsql.SqlDriver) *mux.Router {
	m.SetSqlDriver(d)
	return new(api).setApp(m).getMuxRouter()
}

func (m *App) Construct() (*App, error) {
	return m,nil
}

func (m *App) Bootstrap() error { return nil }

func (m *App) Destruct() error { return nil }


// internal methods:
