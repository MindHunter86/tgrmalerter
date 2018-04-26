package app

import "net/http"
import "github.com/gorilla/mux"
import "mailru/tgrmalerter/core/config"


type appApi struct {
	conf *config.CoreConfig
}

func (m *appApi) setConfig(c *config.CoreConfig) *appApi { m.conf = c; return m }
func (m *appApi) getMuxRouter() *mux.Router {
	var r = mux.NewRouter()

	r.Host(m.conf.Base.Http.Host)
	r.Headers("Content-Type", "application/json")

	r.HandleFunc("/", m.httpRootHandler).Methods("GET")
	r.HandleFunc("/v1/alert/send", nil).Methods("POST")

	return r
}

func (m *appApi) httpRootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte('{"version": "1.0"}'))
}
func (m *appApi) httpAlertSendHandler (w http.ResponseWriter, r *http.Request) {
	var mv = mux.Vars(r)


}
