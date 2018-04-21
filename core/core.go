package core

import (
	"sync"

	"os"
	"os/signal"
	"syscall"

	"mservice1/app"

	"github.com/rs/zerolog"
)


type Core struct {
	http *httpService
	sql sqlDriver

	log *zerolog.Logger
	cfg *CoreConfig

	appWg sync.WaitGroup
	app *app.App
}


func (m *Core) SetLogger(l *zerolog.Logger) *Core { m.log = l; return m }
func (m *Core) SetConfig(c *CoreConfig) *Core { m.cfg = c; return m }
func (m *Core) Construct() (*Core, error) {
	var e error


	// application configuration:
	m.app,e = new(app.App).SetLogger(m.log).Construct(); if e != nil { return nil,e }

	// internal resources configuration:
	m.http = new(httpService).setConfig(m.cfg).setLogger(m.log).construct(m.app.NewHttpRouter())
	m.sql,e = new(mysqlDriver).setConfig(m.cfg).construct(); if e != nil { return nil,e }

	return m,nil
}

func (m *Core) Bootstrap() error {
	// define kernel signal catcher:
	var kernSignal = make(chan os.Signal)
	signal.Notify(kernSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	// define global error variables:
	var e error
	var epipe = make(chan error)

	// http service bootstrap:
	go func(e chan error, wg sync.WaitGroup) {
		wg.Add(1); defer wg.Done()
		e <- m.http.bootstrap()
	}(epipe, m.appWg)

	// application bootstrap:
	go func(e chan error, wg sync.WaitGroup) {
		wg.Add(1); defer wg.Done()
		// TODO: epipe <- m.app.Bootstrap()
	}(epipe, m.appWg)

	// main application event loop:
LOOP:
	for {
		select {

		// kernel signal catcher:
		case <-kernSignal:
			m.log.Warn().Msg("Syscall.SIG* has been detected! Closing application...")
			break LOOP

		// application error catcher:
		case e = <-epipe:
			if e != nil { m.log.Error().Err(e).Msg("Runtime error! Abnormal application closing!") }
			break LOOP
			// TODO: automatic application re-bootstrap

		}
	}

	return m.Destruct(&e)
}

func (m *Core) Destruct(e *error) error {
	// application destruct:
	m.app.Destruct()

	// internal resources destruct:
	m.http.destruct()
	m.sql.destruct()

	m.appWg.Wait(); return *e
}
