package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// mainMenu is the persistent keyboard for a fully-registered user.
var mainMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("👤 My profile"),
		tgbotapi.NewKeyboardButton("🚻 Set my gender"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("📍 Update my location"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🔍 Find nearby friends"),
		tgbotapi.NewKeyboardButton("☕ Coffee chat (15 min)"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🌐 Language"),
		tgbotapi.NewKeyboardButton("⏰ Wake hours"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🔔 Notify me"),
		tgbotapi.NewKeyboardButton("🏆 Achievements"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("❌ End chat"),
	),
)

// chatKeyboard is shown while a chat is active.
var chatKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("❌ End chat"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🚫 Block"),
		tgbotapi.NewKeyboardButton("🚩 Report"),
	),
)

// requestLocationKeyboard asks the user to share their Telegram location.
var requestLocationKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.KeyboardButton{Text: "📎 Share my location", RequestLocation: true},
	),
)

// requestLiveLocationKeyboard asks the user to share their live location
// (15 min, 1h, or 8h — Telegram allows these via the attachment menu).
// We provide a plain button and an explanatory hint.
var requestLiveLocationKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.KeyboardButton{Text: "📎 Share my live location (15 min)", RequestLocation: true},
	),
)

// genderKeyboard is used for both self-gender and search-gender prompts.
func genderKeyboard(prefix string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👨 Male", prefix+":male"),
			tgbotapi.NewInlineKeyboardButtonData("👩 Female", prefix+":female"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌐 Both / Any", prefix+":both"),
		),
	)
}

// radiusKeyboard lets the user pick a search radius.
func radiusKeyboard() tgbotapi.InlineKeyboardMarkup {
	btn := func(label, val string) tgbotapi.InlineKeyboardButton {
		return tgbotapi.NewInlineKeyboardButtonData(label, "radius:"+val)
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn("1 km", "1"), btn("5 km", "5")),
		tgbotapi.NewInlineKeyboardRow(btn("10 km", "10"), btn("50 km", "50")),
		tgbotapi.NewInlineKeyboardRow(btn("100 km", "100"), btn("500 km", "500")),
	)
}

// profileKeyboard shows the profile menu.
func profileKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Edit alias", "profile:alias"),
			tgbotapi.NewInlineKeyboardButtonData("📝 Edit bio", "profile:bio"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏷️ Edit interests", "profile:interests"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🖼️ Set photo", "profile:photo"),
		),
	)
}

// interestsKeyboard returns a multi-select grid of interest tags. The
// "page" argument is 0-based.
func interestsKeyboard(selected map[string]bool, page int) tgbotapi.InlineKeyboardMarkup {
	const perPage = 12
	start := page * perPage
	end := start + perPage
	if end > len(AllInterests) {
		end = len(AllInterests)
	}
	if start > len(AllInterests) {
		start = len(AllInterests)
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for i := start; i < end; i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 2 && i+j < end; j++ {
			tag := AllInterests[i+j]
			label := tag
			if selected[tag] {
				label = "✅ " + tag
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				label, "interest:toggle:"+tag))
		}
		rows = append(rows, row)
	}

	// Pagination.
	var nav []tgbotapi.InlineKeyboardButton
	if page > 0 {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("⬅️", fmt.Sprintf("interest:page:%d", page-1)))
	}
	if end < len(AllInterests) {
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData("➡️", fmt.Sprintf("interest:page:%d", page+1)))
	}
	if len(nav) > 0 {
		rows = append(rows, nav)
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ Done", "interest:done"),
		tgbotapi.NewInlineKeyboardButtonData("🗑️ Clear all", "interest:clear"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// languageKeyboard returns the language picker. 4 per row.
func languageKeyboard() tgbotapi.InlineKeyboardMarkup {
	// Stable order: en, fa, tr, de, fr, es, ar, ru, zh, ja, ko, it, pt, hi, nl
	order := []string{"en", "fa", "tr", "de", "fr", "es", "ar", "ru", "zh", "ja", "ko", "it", "pt", "hi", "nl"}
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(order); i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 2 && i+j < len(order); j++ {
			code := order[i+j]
			name, _ := SupportedLanguages[code]
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(name, "lang:"+code))
		}
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// timezoneKeyboard returns the timezone picker. 2 per row.
func timezoneKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(CommonTimezones); i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 2 && i+j < len(CommonTimezones); j++ {
			tz := CommonTimezones[i+j]
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(tz, "tz:"+tz))
		}
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// hourKeyboard returns an hour picker (0-23). 6 per row.
func hourKeyboard(prefix string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for h := 0; h < 24; h += 6 {
		var row []tgbotapi.InlineKeyboardButton
		for i := 0; i < 6; i++ {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				strconv.Itoa(h+i), prefix+":"+strconv.Itoa(h+i)))
		}
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// notifyRadiusKeyboard lets the user pick the notification radius.
func notifyRadiusKeyboard() tgbotapi.InlineKeyboardMarkup {
	btn := func(label, val string) tgbotapi.InlineKeyboardButton {
		return tgbotapi.NewInlineKeyboardButtonData(label, "notify:radius:"+val)
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn("5 km", "5"), btn("10 km", "10")),
		tgbotapi.NewInlineKeyboardRow(btn("50 km", "50"), btn("100 km", "100")),
		tgbotapi.NewInlineKeyboardRow(btn("500 km", "500")),
	)
}

// ratingKeyboard returns 1..5 star buttons.
func ratingKeyboard(partnerID int64) tgbotapi.InlineKeyboardMarkup {
	var row []tgbotapi.InlineKeyboardButton
	for i := 1; i <= 5; i++ {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			strings.Repeat("⭐", i),
			fmt.Sprintf("rate:%d:%d", partnerID, i)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(row)
}

// continueKeyboard is shown when a coffee chat timer expires.
func continueKeyboard(partnerID int64, _ bool) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ Keep chatting", fmt.Sprintf("continue:%d", partnerID)),
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ End here", fmt.Sprintf("endhere:%d", partnerID)),
		),
	)
}

// confirmKeyboard is a Yes/No confirmation.
func confirmKeyboard(yes, no string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Yes", yes),
			tgbotapi.NewInlineKeyboardButtonData("❌ No", no),
		),
	)
}
