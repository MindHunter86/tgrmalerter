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

const (
	tgJobStatusPending = uint8(iota)
	tgJobStatusFailed
	tgJobStatusSended
)

func (m *tgrmJob) setUserModel(u *userModel) *tgrmJob { m.um = u; return m }
func (m *tgrmJob) queueUp() { globTgDispatcher.getQueueChan() <-m }

func (m *tgrmJob) create(reqId, mess string, usr *baseUser) *tgrmJob {
	m.requestId = reqId
	m.message = mess
	m.user = usr
	m.id = uuid.NewV4().String()

	return m.save()
}

func (m *tgrmJob) save() *tgrmJob {
	stmt,e := globSqlDB.Prepare("INSERT INTO dispatch_reports (id,request_id,recipient,message) VALUES (?,?,?,?)")
	if e != nil { m.um.handleError(e, errInternalSqlError, "Could not prepare DB statement!"); return nil }
	defer stmt.Close()

	if _,e := stmt.Exec(m.id,m.requestId,m.user.phone,m.message); e != nil {
		m.um.handleError(e, errInternalSqlError, "Could not write job!"); return nil }

	return m
}

func (m *tgrmJob) statusUpdate(jobStatus uint8) {
	stmt,e := globSqlDB.Prepare("UPDATE dispatch_reports SET status_code = ? WHERE id = ?")
	if e != nil { globLogger.Error().Err(e).Msg("[TG-QUEUE]: Could not update job status!"); return } // TODO: added telegram errors with save in mysql
	defer stmt.Close()

	if _,e := stmt.Exec(jobStatus, m.id); e != nil {
		globLogger.Error().Err(e).Msg("Could not update job status!")
		return } // TODO: added telegram errors with save in mysql
}


type tgrmDispatcher struct {
	pool chan chan *tgrmJob

	jobQueue chan *tgrmJob

	done chan struct{}
	workerDone chan struct{}
}

func (m *tgrmDispatcher) getQueueChan() chan *tgrmJob { return m.jobQueue }

func (m *tgrmDispatcher) bootstrap(maxWorkers, workerCapacity int) {
	var wg sync.WaitGroup
	wg.Add(maxWorkers + 1)

	for i := 0; i < maxWorkers; i ++ {
		go func(wg sync.WaitGroup) {
			new(tgrmWorker).
				setPool(m.pool).
				setDonePipe(m.workerDone).
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
	pool chan chan *tgrmJob
	inbox chan *tgrmJob

	done chan struct{}
}

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

func (m *tgrmWorker) doJob(j *tgrmJob) { globTgApi.sendMessage(j) }

