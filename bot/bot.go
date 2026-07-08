package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot wraps the Telegram API client plus our state, storage, and
// optional translation client.
type Bot struct {
	api       *tgbotapi.BotAPI
	storage   *Storage
	translate *TranslateClient
}

// New creates the bot but does not start polling. Call Run() to start.
func New(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	api.Debug = false // flip to true while debugging
	return &Bot{
		api:     api,
		storage: NewStorage(),
	}, nil
}

// SetTranslate attaches a translation client. nil disables translation.
func (b *Bot) SetTranslate(t *TranslateClient) {
	b.translate = t
}

// Self exposes the underlying bot user (handy for logging).
func (b *Bot) Self() tgbotapi.User {
	return b.api.Self
}

// Run starts the long-poll loop. Blocks until the process is killed.
func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 50 // Telegram's max long-poll window

	updates := b.api.GetUpdatesChan(u)

	b.startWorkers()

	for upd := range updates {
		b.routeUpdate(upd)
	}
}

// routeUpdate dispatches an Update to the right handler.
func (b *Bot) routeUpdate(upd tgbotapi.Update) {
	switch {
	case upd.Message != nil:
		b.handleMessage(upd.Message)
	case upd.EditedMessage != nil:
		// Live location updates come through here. Treat them like any
		// other message from the user — if they're in a chat, relay.
		b.handleMessage(upd.EditedMessage)
	case upd.CallbackQuery != nil:
		b.handleCallback(upd.CallbackQuery)
	default:
		// ignore inline queries, etc.
	}
}

// answerCallback closes the loading spinner on an inline button press.
// We swallow errors — they're harmless UX-wise.
func (b *Bot) answerCallback(cqID string, text string) {
	if text == "" {
		text = " "
	}
	_, _ = b.api.Request(tgbotapi.NewCallback(cqID, text))
}

// send is a thin wrapper that logs send errors.
func (b *Bot) send(c tgbotapi.Chattable) {
	if _, err := b.api.Send(c); err != nil {
		log.Printf("[send] error: %v", err)
	}
}
