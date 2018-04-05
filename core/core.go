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
	log *zerolog.Logger
	cfg *CoreConfig

	appWg sync.WaitGroup
	app *app.App
}


func (m *Core) SetLogger(l *zerolog.Logger) *Core { m.log = l; return m }
func (m *Core) SetConfig(c *CoreConfig) *Core { m.cfg = c; return m }
func (m *Core) Construct() (*Core, error) {
	var e error
	m.app,e = new(app.App).SetLogger(m.log).Construct(); if e != nil { return nil,e }

	return m,nil
}

func (m *Core) Bootstrap() error {
	// define kernel signal catcher:
	var kernSignal = make(chan os.Signal)
	signal.Notify(kernSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	// application bootstrap:
	var appErrPipe = make(chan error, 1)
	go func(errPipe chan error, wg sync.WaitGroup) {
		wg.Add(1)
		errPipe <- m.app.Bootstrap()
		wg.Done()
	}(appErrPipe, m.appWg)

	// define global error variable:
	var e error = nil

	// main application event loop:
LOOP:
	for {
		select {

		// kernel signal catcher:
		case <-kernSignal:
			m.log.Warn().Msg("Syscall.SIG* has been detected! Closing application...")
			break LOOP

		// application error catcher:
		case e = <-appErrPipe:
			if e != nil { m.log.Error().Err(e).Msg("Runtime error! Abnormal application closing!") }
			break LOOP
			// TODO: automatic application re-bootstrap

		}
	}

	return m.Destruct(&e)
}

func (m *Core) Destruct(e *error) error {
	// This comment contains "bad english", be careful with its reading!
	//
	// method get defined error variable, with or without app.Bootstrap error.
	// if m.app.Destruct() has errors, e variable would be rewritten
	// else in main() we return app.Bootstrap() error (if it was) or nil
	*e = m.app.Destruct()
	m.appWg.Wait()
	return *e
}
