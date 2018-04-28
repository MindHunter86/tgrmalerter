package tgrm

import "mailru/tgrmalerter/app"
import "mailru/tgrmalerter/core/config"
import "github.com/go-telegram-bot-api/telegram-bot-api"

type TgrmService struct {
	ap *app.App
	conf *config.CoreConfig

	done chan struct{}

	tbot *tgbotapi.BotAPI
	tbotUpdates tgbotapi.UpdatesChannel
}

func (m *TgrmService) SetApp(a *app.App) *TgrmService { m.ap = a; return m }
func (m *TgrmService) SetConfig(c *config.CoreConfig) *TgrmService { m.conf = c; return m}

func (m *TgrmService) Construct() (*TgrmService,error) {
	var e error
	if m.tbot,e = tgbotapi.NewBotAPI(m.conf.Base.Telegram.Botapi.Token); e != nil { return nil,e }
	if m.conf.Base.Telegram.Botapi.Tgrm_Debug { m.tbot.Debug = true }

	m.done = make(chan struct{}, 1)

	tgUpdatesConfig := tgbotapi.NewUpdate(0)
	tgUpdatesConfig.Timeout = m.conf.Base.Telegram.Botapi.Timeout
	if m.tbotUpdates,e = m.tbot.GetUpdatesChan(tgUpdatesConfig); e != nil { return nil,e }

	return m,nil
}

func (m *TgrmService) Destruct() error { close(m.done); return nil }

func (m *TgrmService) Bootstrap() error {
LOOP:
	for {
		select {
		case <-m.done: break LOOP
		case u := <-m.tbotUpdates: m.ap.TelegramUpdateHandler(&u)
		}
	}

	return nil
}

func (m *TgrmService) GetTBot() *tgbotapi.BotAPI { return m.tbot }
