package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// mainMenu is the persistent keyboard for a fully-registered user.
func mainMenu(lang string) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(t(lang, "btn_profile")),
			tgbotapi.NewKeyboardButton(t(lang, "btn_set_gender")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(t(lang, "btn_update_location")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(t(lang, "btn_find_nearby")),
			tgbotapi.NewKeyboardButton(t(lang, "btn_coffee_chat")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(t(lang, "btn_language")),
			tgbotapi.NewKeyboardButton(t(lang, "btn_wake_hours")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(t(lang, "btn_notify")),
			tgbotapi.NewKeyboardButton(t(lang, "btn_achievements")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(t(lang, "btn_end_chat")),
		),
	)
}

// chatKeyboard is shown while a chat is active.
func chatKeyboard(lang string) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(t(lang, "btn_end_chat")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(t(lang, "btn_block")),
			tgbotapi.NewKeyboardButton(t(lang, "btn_report")),
		),
	)
}

// requestLocationKeyboard asks the user to share their Telegram location.
func requestLocationKeyboard(lang string) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.KeyboardButton{
				Text:            t(lang, "btn_share_location"),
				RequestLocation: true,
			},
		),
	)
}

// uiLangKeyboard shows the 3 UI language options (used at /start and
// when changing language later).
func uiLangKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, code := range ValidUILangs {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(UILangNames[code], "lang:"+code),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// genderKeyboard is used for both self-gender and search-gender prompts.
func genderKeyboard(lang, prefix string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(t(lang, "btn_male"), prefix+":male"),
			tgbotapi.NewInlineKeyboardButtonData(t(lang, "btn_female"), prefix+":female"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(t(lang, "btn_both"), prefix+":both"),
		),
	)
}

// radiusKeyboard lets the user pick a search radius.
func radiusKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
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
func profileKeyboard(lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(t(lang, "btn_edit_alias"), "profile:alias"),
			tgbotapi.NewInlineKeyboardButtonData(t(lang, "btn_edit_bio"), "profile:bio"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(t(lang, "btn_edit_interests"), "profile:interests"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(t(lang, "btn_set_photo"), "profile:photo"),
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
func continueKeyboard(partnerID int64, lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				t(lang, "btn_keep_chatting"), fmt.Sprintf("continue:%d", partnerID)),
			tgbotapi.NewInlineKeyboardButtonData(
				t(lang, "btn_end_here"), fmt.Sprintf("endhere:%d", partnerID)),
		),
	)
}

// confirmKeyboard is a Yes/No confirmation.
func confirmKeyboard(yes, no, lang string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(t(lang, "btn_yes"), yes),
			tgbotapi.NewInlineKeyboardButtonData(t(lang, "btn_no"), no),
		),
	)
}
