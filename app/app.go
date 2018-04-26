package app

import "github.com/rs/zerolog"
import "github.com/gorilla/mux"

import "mailru/tgrmalerter/core/config"

type App struct {
	log *zerolog.Logger
	conf *config.CoreConfig
	sql *sql.SqlDriver

	api *appApi
}


// application public API:
func (m *App) SetLogger(l *zerolog.Logger) *App { m.log = l; return m }
func (m *App) SetConfig(c *config.CoreConfig) *App { m.conf = c; return m }
func (m *App) SetSQLDriver(d *sql.SqlDriver) *App { m.sql = d; return m }
func (m *App) NewHttpRouter() *mux.Router {	return m.api.getMuxRouter() }

func (m *App) Construct() (*App, error) {
	return m,nil
}

func (m *App) Bootstrap() error { return nil }

func (m *App) Destruct() error { return nil }


// internal methods:
