package core

import "sync"
import "time"
import "net/http"

import "github.com/gorilla/mux"
import "github.com/justinas/alice"

import "github.com/rs/zerolog"
import "github.com/rs/zerolog/hlog"


type httpService struct {
	log *zerolog.Logger
	conf *CoreConfig

	httpServer *http.Server

	done chan struct{}
}

func (m *httpService) setConfig(c *CoreConfig) *httpService { m.conf = c; return m }
func (m *httpService) setLogger(l *zerolog.Logger) *httpService { m.log = l; return m }

func (m *httpService) construct(router *mux.Router) *httpService {
	m.done = make(chan struct{}, 1)

	var chain = alice.New().Append(
		hlog.NewHandler(*m.log),
		hlog.RemoteAddrHandler("ip"),
		hlog.RequestHandler("request"),
		hlog.RefererHandler("referer"),
		hlog.UserAgentHandler("ua"))
	chain = chain.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).Msg("")
	}))

	m.httpServer = &http.Server{
		// TODO Handler: chain.Then(router),
		Handler: chain.Then(nil),
		Addr: m.conf.Base.Http.Listen,
		ReadTimeout: time.Duration(m.conf.Base.Http.Read_Timeout) * time.Millisecond,
		WriteTimeout: time.Duration(m.conf.Base.Http.Write_Timeout) * time.Millisecond }

	m.log.Debug().Msg("Http Service has been successfully configured!")
	return m
}

func (m *httpService) bootstrap() error {
	var e error
	var wg sync.WaitGroup

	go m.httpServe(&wg, &e)

LOOP:
	for {
		select {
		case <-m.done:
			m.log.Info().Msg("HttpService has caught DONE signal. Http Shutdown in progress ...")
			if e == nil {
				if e = m.httpServer.Shutdown(nil); e != nil {
					m.log.Error().Err(e).Msg("Could not shutdown http server correctly! Abnormal exit of http.Shutdown()!") }}
			break LOOP
		}
	}

	m.log.Debug().Msg("Http Service has been successfully bootstrapped!")
	return e
}

func (m *httpService) destruct() error {
	close(m.done)
	return nil
}

func (m *httpService) httpServe(wg *sync.WaitGroup, e *error) {
	wg.Add(1); defer wg.Done()
	m.log.Debug().Msg("http.ListenAndServe executing ...")
	if *e = m.httpServer.ListenAndServe(); *e != nil && *e != http.ErrServerClosed {
		m.log.Error().Err(*e).Msg("Http.ListenAndServe abnormal exit!"); close(m.done) }
}
