package app

import "github.com/rs/zerolog"

type App struct {
	log *zerolog.Logger
}


// application public API:
func (m *App) Construct() (*App, error) {
	return m,nil
}

func (m *App) SetLogger(l *zerolog.Logger) *App { m.log = l ; return m }

func (m *App) Bootstrap() error { return nil }

func (m *App) Destruct() error { return nil }


// internal methods:
