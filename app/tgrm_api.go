package app

import "github.com/rs/zerolog"
import "github.com/go-telegram-bot-api/telegram-bot-api"

const (
	tgrmMsgRequestAgreement = "Для получения уведомлений о входе на сервера от меня, тебе необходимо зарегистрироваться. При нажатии на кнопку, телеграм расшарит твой телефон для меня. После я тебя запишу тебя в свою базу и буду отправлять уведомления по команде в мою апишку."
	tgrmMsgRegisterPhoneOwner = "Номер телефона, который ты мне дал, не принадлежит тебе! Просто нажми на кнопку ДА и не страдай фигнёй!"
	tgrmMsgRegisterSuccess = "Ты был зарегистрирован в системе, поздравляю! Теперь все уведомления о заходе на сервера по SSH будут приходить в этот чатик. Если есть вопросы спрашивай у Бажина."
)

type tgrmApi struct {
	log *zerolog.Logger
	tbot *tgbotapi.BotAPI
}

func (m *tgrmApi) setTBot(t *tgbotapi.BotAPI) *tgrmApi { m.tbot = t; return m }
func (m *tgrmApi) setLogger(l *zerolog.Logger) *tgrmApi { m.log = l; return m }

func (m *tgrmApi) commandRouter(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start": m.requestContact(msg.Chat.ID)
	default: m.log.Warn().Str("author", msg.From.String()).Msg("An empty command has been received!")
	}
}

func (m *tgrmApi) requestContact(chatId int64) {
	msgAgreement := tgbotapi.NewMessage(chatId, tgrmMsgRequestAgreement)
	msgAgreement.ReplyMarkup = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButtonContact("Заргеистрироваться в системе") })
	if _,e :=m.tbot.Send(msgAgreement); e != nil {
		m.log.Error().Err(e).Msg("Could not send msgAgreement!") }
}

func (m *tgrmApi) registerContact(chatId int64, formId int, contact *tgbotapi.Contact) {
	if contact.UserID != formId {
		if _,e := m.tbot.Send(tgbotapi.NewMessage(chatId, tgrmMsgRegisterPhoneOwner)); e != nil {
			m.log.Error().Err(e).Msg("Could not send message!") } }

	// SOME REGISTRATION

	msg := tgbotapi.NewMessage(chatId, tgrmMsgRegisterSuccess)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	if _,e := m.tbot.Send(msg); e != nil {
		m.log.Error().Err(e).Msg("Could not send message!") }
}

func (m *tgrmApi) sendMessage(uId, message string) {
	m.log.Debug().Str("uId", uId).Str("message", message).Msg("WTF?!")
	if _,e := m.tbot.Send(tgbotapi.NewMessageToChannel(uId, message)); e != nil {
		m.log.Error().Err(e).Msg("Could not send message!") }
}
