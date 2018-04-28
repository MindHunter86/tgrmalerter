package app

import "sync"
import "github.com/satori/go.uuid"


type tgrmJob struct {
	// internal pointers:
	um *userModel // XXX: think about it ...

	// payload:
	id string
	requestId string
	user *baseUser
	message string
	status bool
	status_code uint8
}

func (m *tgrmJob) setUserModel(u *userModel) *tgrmJob { m.um = u; return m }
func (m *tgrmJob) queueUp() { m.um.ap.tgDispatcher.getQueueChan() <-m }

func (m *tgrmJob) create(reqId, mess string, usr *baseUser) *tgrmJob {
	m.requestId = reqId
	m.message = mess
	m.user = usr
	m.id = uuid.NewV4().String()

	return m.save()
}

func (m *tgrmJob) save() *tgrmJob {
	stmt,e := m.um.ap.sqldb.Prepare("INSERT INTO dispatch_reports (id,request_id,recipient,message) VALUES (?,?,?,?)")
	if e != nil { m.um.handleErrors(e, errInternalSqlError, "Could not prepare DB statement!"); return nil }
	defer stmt.Close()

	if _,e := stmt.Exec(m.id,m.requestId,m.user.phone,m.message); e != nil {
		m.um.handleErrors(e, errInternalSqlError, "Could not write job!"); return nil }

	return m
}


type tgrmDispatcher struct {
	tgApi *tgrmApi

	pool chan chan *tgrmJob

	jobQueue chan *tgrmJob

	done chan struct{}
	workerDone chan struct{}
}

func (m *tgrmDispatcher) getQueueChan() chan *tgrmJob { return m.jobQueue }
func (m *tgrmDispatcher) setTgApiPointer(t *tgrmApi) *tgrmDispatcher { m.tgApi = t; return m }

func (m *tgrmDispatcher) bootstrap(maxWorkers, workerCapacity int) {
	var wg sync.WaitGroup
	wg.Add(maxWorkers + 1)

	for i := 0; i < maxWorkers; i ++ {
		go func(wg sync.WaitGroup) {
			new(tgrmWorker).
				setPool(m.pool).
				setDonePipe(m.workerDone).
				setTgApi(m.tgApi).
				spawn(workerCapacity)
			wg.Done() }(wg)
	}

	go func(wg sync.WaitGroup) {
		m.dispatch()
		wg.Done() }(wg)

	wg.Wait()
}

func (m *tgrmDispatcher) destruct() { close(m.done) }

func (m *tgrmDispatcher) dispatch() {
LOOP:
	for {
		select {
		case <-m.done: break LOOP
		case j := <-m.jobQueue:
			go func(job *tgrmJob) {
				workerPipe := <-m.pool
				workerPipe <- job }(j)
		}
	}

	close(m.workerDone)
}


type tgrmWorker struct {
	tgApi *tgrmApi

	pool chan chan *tgrmJob
	inbox chan *tgrmJob

	done chan struct{}
}

func (m *tgrmWorker) setTgApi(t *tgrmApi) *tgrmWorker { m.tgApi = t; return m }
func (m *tgrmWorker) setPool(p chan chan *tgrmJob) *tgrmWorker { m.pool = p; return m }
func (m *tgrmWorker) setDonePipe(d chan struct{}) *tgrmWorker { m.done = d; return m }

func (m *tgrmWorker) spawn(cap int) {
	m.inbox = make(chan *tgrmJob, cap)

LOOP:
	for {
		m.pool <- m.inbox

		select {
		case j := <-m.inbox: m.doJob(j)
		case <-m.done:
			close(m.inbox)
			break LOOP
		}
	}
}

func (m *tgrmWorker) doJob(j *tgrmJob) {
	j.um.ap.log.Debug().Msg("HOORAY OVER UNSAFE POINTER!")
	m.tgApi.log.Debug().Msg("tgApi debug message")
	m.tgApi.sendMessage(j.user.chatId, j.message)
}

