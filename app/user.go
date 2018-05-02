package app

import "net/http"
import "github.com/gorilla/context"

type userModel struct {
	phone string
	request *httpRequest
	user *baseUser
}

type baseUser struct {
  phone string
	chatId int64
	username string
	registered bool
}

func (m *userModel) construct(r *http.Request) *userModel {
	m.request = context.Get(r, "internal_request").(*httpRequest)
	return m
}

func (m *userModel) handleError(e error, err uint8, msg string) {
	m.request.newError(err); globLogger.Error().Err(e).Msg(msg)
}

func (m *userModel) findUserByPhone(phone string) bool {
	stmt,e := globSqlDB.Prepare("SELECT IF(COUNT(1)=1, true, false) FROM users WHERE phone = ? LIMIT 1")
	if e != nil { m.handleError(e, errInternalSqlError, "[USER]: Could not prepare DB statement!"); return false }
	defer stmt.Close()

	rows,e := stmt.Query(phone)
	if e != nil { m.handleError(e, errInternalSqlError, "[USER]: Could not get result from DB!"); return false }
	defer rows.Close()

	if ! rows.Next() {
		m.handleError(nil, errInternalCommonError, "[USER]: Could not exec rows.News method!"); return false }

	var isPhoneExist bool
	if e := rows.Scan(&isPhoneExist); e != nil {
		m.handleError(e, errInternalSqlError, "[USER]: Could not scan the result from DB!"); return false }

	if ! isPhoneExist {
		m.request.newError(errAlertsCreatePhoneNotFound).setParameter("phone"); return false }

	m.phone = phone
	return true
}

func (m *userModel) getUserByPhone() *baseUser {
	stmt,e := globSqlDB.Prepare("SELECT phone,chat_id,registered FROM users WHERE phone = ? LIMIT 1")
	if e != nil { m.handleError(e, errInternalSqlError, "[USER]: Could not prepare DB statement!"); return nil }
	defer stmt.Close()

	rows,e := stmt.Query(m.phone)
	if e != nil { m.handleError(e, errInternalSqlError, "[USER]: Could not get result from DB!"); return nil }
	defer rows.Close()

	if ! rows.Next() {
		m.handleError(nil, errInternalCommonError, "[USER]: Could not exec rows.News method!"); return nil }

	m.user = new(baseUser)
	if e := rows.Scan(&m.user.phone, &m.user.chatId, &m.user.registered); e != nil {
		m.handleError(e, errInternalSqlError, "[USER]: Could not scan the result from DB!"); return nil }

	return m.user
}

func (m *userModel) sendAlertWithResponse(alert string) (*responseData,int) {
	if m.getUserByPhone() == nil { return nil,0 }
	if ! m.user.registered {
		m.request.newError(errAlertsCreatePhoneNotRegistered); return nil,0 }

	tgJob := new(tgrmJob).setUserModel(m).create(m.request.id, alert, m.user)
	if tgJob == nil || len(m.request.errors) != 0 { return nil,0 }
	tgJob.queueUp()

	return &responseData{
		Type: "alerts",
		Id: tgJob.id,
		Attributes: &dataAttributes{
			Status: "IN QUEUE" }, // TODO: ADD CONS MESSAGES IN TGRM_ERRORS
	}, http.StatusAccepted
}

// methods for baseUser:
func (m *baseUser) createAndSave(phone string, chatId int64) error {
	stmt,e := globSqlDB.Prepare("INSERT INTO users (phone,chat_id,registered) VALUES (?,?,?)")
	if e != nil { return e }
	defer stmt.Close()

	if _,e := stmt.Exec(phone, chatId, true); e != nil { return e }

	m.phone = phone
	m.chatId = chatId

	return nil
}

// internal helpers for userModel:
func parseRawPhone(phone string) (string, uint8) {
	// check length:
	if ph_len := len(phone); ph_len < 10 || ph_len > 12 {
		return "",errAlertsCreatePhoneLength }

	// check format and shorten up to 10 digits:
	switch phone[0:2] {
	case "+7":
		if phone[2:3] != "9" { return "",errAlertsCreatePhoneFormat }
		phone = phone[2:]
	case "89": phone = phone[1:]
	case "79": phone = phone[1:]
	default: return "",errAlertsCreatePhoneFormat }

	return phone,errNotError
}
