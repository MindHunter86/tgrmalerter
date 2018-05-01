package app

import "github.com/go-telegram-bot-api/telegram-bot-api"

const (
	tgrmMsgRequestAgreement = "Для получения уведомлений о входе на сервера от меня, тебе необходимо зарегистрироваться. При нажатии на кнопку, телеграм расшарит твой телефон для меня. После я тебя запишу тебя в свою базу и буду отправлять уведомления по команде в мою апишку."
	tgrmMsgRegisterPhoneOwner = "Номер телефона, который ты мне дал, не принадлежит тебе! Просто нажми на кнопку ДА и не страдай фигнёй!"
	tgrmMsgRegisterSuccess = "Ты был зарегистрирован в системе, поздравляю! Теперь все уведомления о заходе на сервера по SSH будут приходить в этот чатик. Если есть вопросы спрашивай у Бажина."
)

type tgrmApi struct {
	*tgbotapi.BotAPI
}

func (m *tgrmApi) construct(t *tgbotapi.BotAPI) *tgrmApi { m.BotAPI = t; return m }

func (m *tgrmApi) commandRouter(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start": m.requestContact(msg.Chat.ID)
	default: globLogger.Warn().Str("author", msg.From.String()).Msg("[TG-API]: An empty command has been received!")
	}
}

func (m *tgrmApi) requestContact(chatId int64) {
	msgAgreement := tgbotapi.NewMessage(chatId, tgrmMsgRequestAgreement)
	msgAgreement.ReplyMarkup = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButtonContact("Заргеистрироваться в системе") })
	if _,e :=m.Send(msgAgreement); e != nil {
		globLogger.Error().Err(e).Msg("[TG-API]: Could not send msgAgreement!") }
}

func (m *tgrmApi) registerContact(chatId int64, formId int, contact *tgbotapi.Contact) {
	if contact.UserID != formId {
		if _,e := m.Send(tgbotapi.NewMessage(chatId, tgrmMsgRegisterPhoneOwner)); e != nil {
			globLogger.Error().Err(e).Msg("[TG-API]: Could not send message!") } }

	// SOME REGISTRATION
	phone,_ := parseRawPhone(contact.PhoneNumber)
	if e := new(baseUser).createAndSave(phone, chatId); e != nil {
		globLogger.Error().Err(e).Msg("[TG-API]: User creation failed!")	}

	msg := tgbotapi.NewMessage(chatId, tgrmMsgRegisterSuccess)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	if _,e := m.Send(msg); e != nil {
		globLogger.Error().Err(e).Msg("[TG-API]: Could not send message!") }
}

func (m *tgrmApi) sendMessage(chId int64, message string) {
	if _,e := m.Send(tgbotapi.NewMessage(chId, message)); e != nil {
		globLogger.Error().Err(e).Msg("[TG-API]: Could not send message!") }
}
