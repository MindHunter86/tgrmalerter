package app

import "net/http"
import "github.com/gorilla/context"

type modelUser struct {
	// internal pointers:
	ap *api

	// payload:
	isRegistered bool
	phone string
}


func getUserByPhone(r *http.Request) *modelUser {
	var ap = context.Get(r, "internal_api").(*api)
	var errs = context.Get(r, "internal_errors").(*apiErrors)
	var phone = context.Get(r, "param_phone").(string)

	// check phone length:
	if ph_len := len(phone); ph_len < 10 || ph_len > 12 {
		errs.newError(errAlertsCreatePhoneLength)
		return nil }

	// check phone format and shorten up to 10 digits:
	switch phone[0:2] {
	case "+7":
		if phone[2:3] != "9" { errs.newError(errAlertsCreatePhoneFormat); return nil }
		phone = phone[2:]
	case "89":
		phone = phone[1:]
	default: errs.newError(errAlertsCreatePhoneFormat); return nil }

	// check phone for availability:
	stmt,e := ap.sqldb.Prepare("SELECT IF(COUNT(1) = 1, 'true', 'false') FROM users WHERE phone = ? LIMIT 1"); if e != nil {
		ap.log.Error().Err(e).Msg("Could not prepare DB statement!")
		errs.newError(errInternalSqlError)
		return nil }
	defer stmt.Close()

	rows,e := stmt.Query(phone); if e != nil {
		ap.log.Error().Err(e).Msg("Could not get result from DB!")
		errs.newError(errInternalSqlError)
		return nil }
	defer rows.Close()

	for rows.Next() {
		if reg := false; rows.Scan(&reg) != nil && reg {
			ap.log.Debug().Str("phone", phone).Msg("User has been successfully found!")
			return &modelUser{
				ap: ap,
				isRegistered: reg,
				phone: phone }
		}
	}

	errs.newError(errAlertsCreatePhoneNotFound).setParameter("phone")
	ap.log.Debug().Str("phone", phone).Msg("The requested phone was not found in the database!")
	return nil
}

func (m *modelUser) putAlertInQueue() *responseData { return nil }
