package main

import (
	"log"
	"os"

	"github.com/nimacode/nearfriend-bot/bot"
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

	if url := os.Getenv("TRANSLATION_API_URL"); url != "" {
		b.SetTranslate(bot.NewTranslateClient(url, os.Getenv("TRANSLATION_API_KEY")))
		log.Printf("[nearfriend] chat translation enabled (%s)", url)
	} else {
		log.Printf("[nearfriend] chat translation disabled (set TRANSLATION_API_URL to enable)")
	}

	me := b.Self()
	log.Printf("[nearfriend] bot @%s is online", me.UserName)
	b.Run()
}
