package main

import (
	"log"
	"os"

	"github.com/yourname/nearfriend-bot/bot"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("[nearfriend] TELEGRAM_BOT_TOKEN env var is required. " +
			"Talk to @BotFather on Telegram to create a bot and copy the token.")
	}

	b, err := bot.New(token)
	if err != nil {
		log.Fatalf("[nearfriend] failed to create bot: %v", err)
	}

	me := b.Self()
	log.Printf("[nearfriend] bot @%s is online", me.UserName)
	b.Run()
}
