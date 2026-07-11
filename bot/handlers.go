package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ---------------------------------------------------------------------------
// Message dispatch
// ---------------------------------------------------------------------------

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	if msg.IsCommand() {
		b.handleCommand(msg)
		return
	}

	u, ok := b.storage.Get(msg.From.ID)
	if !ok {
		b.sendText(msg.Chat.ID, t(UILangEn, "msg_send_start"), nil, false)
		return
	}

	lang := u.Lang()

	// User hasn't picked a language yet.
	if u.State == StateNeedUILang {
		b.sendText(msg.Chat.ID, t(lang, "lang_select"), uiLangKeyboard(), false)
		return
	}

	// In chat: first check for chat-keyboard buttons (which would
	// otherwise be relayed as plain text to the partner).
	if u.State == StateInChat {
		switch strings.TrimSpace(msg.Text) {
		case t(lang, "btn_end_chat"):
			b.endChat(u.ID, msg.Chat.ID, true)
			return
		case t(lang, "btn_block"):
			b.onMenuBlock(msg)
			return
		case t(lang, "btn_report"):
			b.onMenuReport(msg)
			return
		}
		// Otherwise relay everything (text, photo, voice, location — including
		// live location, since CopyMessage preserves the live nature).
		b.relayToPartner(msg)
		return
	}

	// Not in chat: a location share updates the user's profile location.
	if msg.Location != nil {
		b.onLocation(msg)
		return
	}

	// State-based text / photo input.
	b.handleStatefulInput(msg, u)
}

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	cmd := msg.Command()

	u, ok := b.storage.Get(msg.From.ID)
	if !ok {
		if cmd == "start" {
			b.cmdStart(msg)
			return
		}
		b.sendText(msg.Chat.ID, t(UILangEn, "msg_send_start"), nil, false)
		return
	}

	lang := u.Lang()

	// While in chat only a few commands are meaningful.
	if u.State == StateInChat {
		switch cmd {
		case "end", "stop", "cancel":
			b.endChat(msg.From.ID, msg.Chat.ID, true)
		default:
			b.sendText(msg.Chat.ID,
				t(lang, "msg_in_chat_hint"),
				chatKeyboard(lang), false)
		}
		return
	}

	// If picking language, only /start is meaningful.
	if u.State == StateNeedUILang {
		if cmd == "start" {
			b.cmdStart(msg)
			return
		}
		b.sendText(msg.Chat.ID, t(lang, "lang_select"), uiLangKeyboard(), false)
		return
	}

	switch cmd {
	case "start":
		b.cmdStart(msg)
	case "menu":
		b.sendText(msg.Chat.ID, t(lang, "msg_main_menu"), mainMenu(lang), false)
	case "end", "stop", "cancel":
		b.sendText(msg.Chat.ID, t(lang, "msg_not_in_chat"), mainMenu(lang), false)
	case "profile":
		b.showProfile(msg.Chat.ID, u)
	case "achievements":
		b.sendText(msg.Chat.ID, achievementsText(u), nil, true)
	case "language":
		b.sendText(msg.Chat.ID, t(lang, "lang_select_short"), uiLangKeyboard(), false)
	case "hours":
		b.startWakeHoursFlow(msg.Chat.ID, u)
	case "notify":
		b.toggleNotifications(u, msg.Chat.ID)
	case "skip":
		b.handleSkip(msg.Chat.ID, u)
	default:
		b.sendText(msg.Chat.ID, t(lang, "msg_unknown_cmd"), nil, false)
	}
}

func (b *Bot) handleCallback(cq *tgbotapi.CallbackQuery) {
	// For inaccessible (very old) messages Telegram sends no Message.
	if cq.Message == nil {
		b.answerCallback(cq.ID, t(UILangEn, "msg_button_unavailable"))
		return
	}
	data := cq.Data
	parts := strings.SplitN(data, ":", 3)

	switch {
	case len(parts) >= 2 && parts[0] == "gender" && len(parts) == 3:
		b.onGenderPick(cq, parts[1], parts[2])

	case len(parts) == 2 && parts[0] == "radius":
		b.onRadiusPick(cq, parts[1])

	case len(parts) == 2 && parts[0] == "nearby":
		id, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			b.answerCallback(cq.ID, t(UILangEn, "err_bad_user_id"))
			return
		}
		b.onNearbyPick(cq, id)

	case parts[0] == "endchat":
		b.onEndChatButton(cq)

	case parts[0] == "profile" && len(parts) == 2:
		b.onProfileAction(cq, parts[1])

	case parts[0] == "interest" && len(parts) >= 2:
		b.onInterestAction(cq, parts[1:])

	case parts[0] == "lang" && len(parts) == 2:
		b.onLanguagePick(cq, parts[1])

	case parts[0] == "tz" && len(parts) == 2:
		b.onTimezonePick(cq, parts[1])

	case parts[0] == "wake" && len(parts) == 3:
		b.onWakePick(cq, parts[1], parts[2])

	case strings.HasPrefix(data, "notify:radius:") && len(parts) == 3:
		b.onNotifyRadiusPick(cq, parts[2])

	case strings.HasPrefix(data, "rate:") && len(parts) >= 2:
		b.onRateAction(cq, parts)

	case strings.HasPrefix(data, "continue:") && len(parts) == 2:
		b.onContinueAction(cq, parts[1])

	case strings.HasPrefix(data, "endhere:") && len(parts) == 2:
		b.onEndHereButton(cq, parts[1])

	case parts[0] == "report" || parts[0] == "block":
		b.handleConfirm(cq, parts)

	default:
		b.answerCallback(cq.ID, t(UILangEn, "err_unknown_action"))
	}
}

// ---------------------------------------------------------------------------
// /start and registration flow
// ---------------------------------------------------------------------------

func (b *Bot) cmdStart(msg *tgbotapi.Message) {
	uid := msg.From.ID

	existing, ok := b.storage.Get(uid)
	if ok && existing.Gender != "" && existing.Latitude != 0 && existing.UILang != "" {
		lang := existing.Lang()
		b.sendText(msg.Chat.ID, t(lang, "msg_welcome_back"), mainMenu(lang), false)
		return
	}

	// No language chosen yet → ask first.
	if !ok || existing.UILang == "" {
		u := &User{
			ID:        uid,
			FirstName: msg.From.FirstName,
			Username:  msg.From.UserName,
			State:     StateNeedUILang,
		}
		if ok {
			u.Gender = existing.Gender
			u.LookingFor = existing.LookingFor
		}
		b.storage.Upsert(u)
		b.sendText(msg.Chat.ID, t(UILangEn, "lang_select"), uiLangKeyboard(), false)
		return
	}

	// Has language but not fully registered → resume registration.
	lang := existing.Lang()
	b.storage.WithUser(uid, func(x *User) {
		x.State = StateNeedLocation
	})
	b.sendText(msg.Chat.ID,
		tf(lang, "msg_start_intro", msg.From.FirstName),
		requestLocationKeyboard(lang), true)
}

func (b *Bot) onLocation(msg *tgbotapi.Message) {
	u, ok := b.storage.Get(msg.From.ID)
	if !ok {
		u = &User{ID: msg.From.ID, FirstName: msg.From.FirstName, Username: msg.From.UserName}
	}
	lang := u.Lang()
	u.Latitude = msg.Location.Latitude
	u.Longitude = msg.Location.Longitude
	u.LocationAt = time.Now()

	if u.State == StateNeedLocation {
		u.State = StateNeedGender
		b.storage.Upsert(u)

		b.sendText(msg.Chat.ID,
			t(lang, "msg_got_location"),
			genderKeyboard(lang, "gender:self"), false)
		b.sendText(msg.Chat.ID, t(lang, "msg_main_menu"), mainMenu(lang), false)
		return
	}

	b.storage.Upsert(u)
	if msg.Location.LivePeriod > 0 {
		b.sendText(msg.Chat.ID,
			t(lang, "msg_live_location_updated"),
			mainMenu(lang), false)
		return
	}
	b.sendText(msg.Chat.ID, t(lang, "msg_location_updated"), mainMenu(lang), false)
}

func (b *Bot) promptLocation(chatID int64, lang string) {
	b.sendText(chatID, t(lang, "msg_share_location_prompt"),
		requestLocationKeyboard(lang), false)
}

// ---------------------------------------------------------------------------
// Stateful text / photo input
// ---------------------------------------------------------------------------

func (b *Bot) handleStatefulInput(msg *tgbotapi.Message, u *User) {
	lang := u.Lang()

	// Menu buttons work from any state — let users jump around.
	switch strings.TrimSpace(msg.Text) {
	case t(lang, "btn_profile"):
		b.showProfile(msg.Chat.ID, u)
		return
	case t(lang, "btn_set_gender"):
		b.sendText(msg.Chat.ID, t(lang, "msg_whats_gender"), genderKeyboard(lang, "gender:self"), false)
		return
	case t(lang, "btn_update_location"):
		b.promptLocation(msg.Chat.ID, lang)
		return
	case t(lang, "btn_find_nearby"):
		b.startSearch(u, msg.Chat.ID, false)
		return
	case t(lang, "btn_coffee_chat"):
		b.startSearch(u, msg.Chat.ID, true)
		return
	case t(lang, "btn_language"):
		b.sendText(msg.Chat.ID, t(lang, "lang_select_short"), uiLangKeyboard(), false)
		return
	case t(lang, "btn_wake_hours"):
		b.startWakeHoursFlow(msg.Chat.ID, u)
		return
	case t(lang, "btn_notify"):
		b.toggleNotifications(u, msg.Chat.ID)
		return
	case t(lang, "btn_achievements"):
		b.sendText(msg.Chat.ID, achievementsText(u), nil, true)
		return
	case t(lang, "btn_end_chat"):
		b.endChat(u.ID, msg.Chat.ID, true)
		return
	}

	switch u.State {
	case StateNeedUILang:
		b.sendText(msg.Chat.ID, t(lang, "lang_select"), uiLangKeyboard(), false)

	case StateNeedGender:
		b.sendText(msg.Chat.ID, t(lang, "msg_pick_gender_below"), nil, false)
		b.sendText(msg.Chat.ID, t(lang, "msg_whats_gender"), genderKeyboard(lang, "gender:self"), false)

	case StateNeedSearchGender:
		b.sendText(msg.Chat.ID, t(lang, "msg_pick_search_gender"), nil, false)

	case StateNeedRadius:
		b.sendText(msg.Chat.ID, t(lang, "msg_pick_radius"), nil, false)

	case StateNeedAlias:
		alias := strings.TrimSpace(msg.Text)
		if alias == "" {
			b.sendText(msg.Chat.ID, t(lang, "msg_alias_empty"), nil, false)
			return
		}
		if len(alias) > 32 {
			alias = alias[:32]
		}
		b.storage.WithUser(u.ID, func(x *User) {
			x.Alias = alias
			x.State = StateIdle
		})
		b.sendText(msg.Chat.ID, tf(lang, "fmt_alias_set", alias), mainMenu(lang), true)

	case StateNeedBio:
		bio := strings.TrimSpace(msg.Text)
		if len(bio) > 200 {
			bio = bio[:200]
		}
		b.storage.WithUser(u.ID, func(x *User) {
			x.Bio = bio
			x.State = StateIdle
		})
		b.sendText(msg.Chat.ID, t(lang, "msg_bio_saved"), mainMenu(lang), false)

	case StateNeedInterests:
		b.sendText(msg.Chat.ID,
			t(lang, "msg_interests_prompt_2"),
			interestsKeyboard(interestSet(u.Interests), 0), true)

	case StateNeedPhoto:
		b.handlePhotoInput(msg, u)

	case StateNeedLanguage:
		b.sendText(msg.Chat.ID, t(lang, "lang_select_short"), uiLangKeyboard(), false)

	case StateNeedTimezone:
		b.sendText(msg.Chat.ID, t(lang, "msg_pick_timezone"), timezoneKeyboard(), false)

	case StateNeedWakeFrom:
		h, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || h < 0 || h > 23 {
			b.sendText(msg.Chat.ID, t(lang, "msg_send_hour_0_23"), nil, false)
			return
		}
		b.storage.WithUser(u.ID, func(x *User) {
			x.WakeFrom = h
			x.State = StateNeedWakeTo
		})
		b.sendText(msg.Chat.ID, tf(lang, "fmt_wake_from_set_named", h), nil, true)

	case StateNeedWakeTo:
		h, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || h < 0 || h > 23 {
			b.sendText(msg.Chat.ID, t(lang, "msg_send_hour_0_23"), nil, false)
			return
		}
		b.storage.WithUser(u.ID, func(x *User) {
			x.WakeTo = h
			x.State = StateIdle
		})
		b.sendText(msg.Chat.ID,
			tf(lang, "fmt_wake_hours_set", u.WakeFrom, h),
			mainMenu(lang), true)

	case StateNeedNotifyRadius:
		b.sendText(msg.Chat.ID, t(lang, "msg_pick_notify_radius"), notifyRadiusKeyboard(), false)

	case StateRatePartner:
		b.sendText(msg.Chat.ID, t(lang, "msg_rate_chat"),
			ratingKeyboard(u.PendingRatingFor), false)

	default:
		b.sendText(msg.Chat.ID, t(lang, "msg_use_menu"), mainMenu(lang), false)
	}
}

func (b *Bot) handleSkip(chatID int64, u *User) {
	lang := u.Lang()
	switch u.State {
	case StateNeedAlias, StateNeedBio, StateNeedInterests,
		StateNeedLanguage, StateNeedTimezone, StateNeedWakeFrom, StateNeedWakeTo:
		b.storage.WithUser(u.ID, func(x *User) {
			x.State = StateIdle
		})
		b.sendText(chatID, t(lang, "msg_skipped"), mainMenu(lang), false)
	case StateRatePartner:
		b.storage.WithUser(u.ID, func(x *User) {
			x.PendingRatingFor = 0
			x.State = StateIdle
		})
		b.sendText(chatID, t(lang, "msg_skipped_rating"), mainMenu(lang), false)
	default:
		b.sendText(chatID, t(lang, "msg_nothing_to_skip"), mainMenu(lang), false)
	}
}

// ---------------------------------------------------------------------------
// Gender
// ---------------------------------------------------------------------------

func (b *Bot) onGenderPick(cq *tgbotapi.CallbackQuery, scope, value string) {
	if value != GenderMale && value != GenderFemale && value != GenderAny {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_value"))
		return
	}

	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(UILangEn, "err_please_start"))
		return
	}
	lang := u.Lang()

	switch scope {
	case "self":
		b.storage.WithUser(u.ID, func(x *User) {
			x.Gender = value
		})
		b.answerCallback(cq.ID, t(lang, "msg_saved"))

		edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
			tf(lang, "fmt_gender_set", prettyGender(value, lang)))
		edit.ParseMode = "Markdown"
		b.send(edit)
		b.sendText(cq.Message.Chat.ID,
			t(lang, "msg_gender_all_set"),
			mainMenu(lang), true)

	case "search":
		b.storage.WithUser(u.ID, func(x *User) {
			x.LookingFor = value
			x.State = StateNeedRadius
		})
		b.answerCallback(cq.ID, t(lang, "label_got_it"))

		edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
			tf(lang, "fmt_looking_for_radius", prettyGender(value, lang)))
		edit.ParseMode = "Markdown"
		b.send(edit)
		b.sendText(cq.Message.Chat.ID, t(lang, "msg_search_within"), radiusKeyboard(lang), false)

	default:
		b.answerCallback(cq.ID, t(lang, "err_unknown_gender_scope"))
	}
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func (b *Bot) startSearch(u *User, chatID int64, coffee bool) {
	lang := u.Lang()
	if u.Latitude == 0 && u.Longitude == 0 {
		b.sendText(chatID, t(lang, "msg_share_location_first"),
			requestLocationKeyboard(lang), false)
		return
	}
	if u.Gender == "" {
		b.sendText(chatID, t(lang, "msg_set_gender_first"),
			genderKeyboard(lang, "gender:self"), false)
		return
	}

	b.storage.WithUser(u.ID, func(x *User) {
		x.State = StateNeedSearchGender
		x.IsCoffeeSearch = coffee
	})

	prompt := t(lang, "msg_who_chat_with")
	if coffee {
		prompt = t(lang, "msg_coffee_who_chat_with")
	}
	b.sendText(chatID, prompt, genderKeyboard(lang, "gender:search"), false)
}

func (b *Bot) onRadiusPick(cq *tgbotapi.CallbackQuery, raw string) {
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(UILangEn, "err_please_start"))
		return
	}
	lang := u.Lang()
	if u.State != StateNeedRadius {
		b.answerCallback(cq.ID, t(lang, "err_not_waiting_radius"))
		return
	}

	radius, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		b.answerCallback(cq.ID, t(lang, "err_bad_radius"))
		return
	}

	candidates := b.filterMatches(u, radius)
	coffee := u.IsCoffeeSearch

	b.storage.WithUser(u.ID, func(x *User) {
		x.State = StateBrowsing
	})
	b.answerCallback(cq.ID, fmt.Sprintf("%.0f km", radius))

	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		tf(lang, "fmt_searching_within", radius))
	b.send(edit)

	if len(candidates) == 0 {
		hint := t(lang, "msg_no_matches")
		if coffee {
			hint = t(lang, "msg_no_coffee_matches")
		}
		b.sendText(cq.Message.Chat.ID, hint, mainMenu(lang), false)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, c := range candidates {
		dist := DistanceKm(u.Latitude, u.Longitude, c.Latitude, c.Longitude)
		label := formatMatchLabel(c, dist, lang)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label,
				"nearby:"+strconv.FormatInt(c.ID, 10)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(t(lang, "btn_cancel"), "endchat"),
	))

	title := tf(lang, "fmt_found_nearby", len(candidates), prettyGender(u.LookingFor, lang))
	if coffee {
		title = tf(lang, "fmt_found_coffee_nearby", len(candidates), prettyGender(u.LookingFor, lang))
	}
	b.sendText(cq.Message.Chat.ID, title, tgbotapi.NewInlineKeyboardMarkup(rows...), true)
}

func (b *Bot) onNearbyPick(cq *tgbotapi.CallbackQuery, targetID int64) {
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(UILangEn, "err_please_start"))
		return
	}
	lang := u.Lang()
	if u.State != StateBrowsing {
		b.answerCallback(cq.ID, t(lang, "err_pick_fresh_search"))
		return
	}
	target, ok := b.storage.Get(targetID)
	if !ok {
		b.answerCallback(cq.ID, t(lang, "err_user_unavailable"))
		return
	}
	if target.State == StateInChat {
		b.answerCallback(cq.ID, t(lang, "err_already_chatting"))
		return
	}
	if !target.IsAwake() {
		b.answerCallback(cq.ID, t(lang, "err_sleeping"))
		return
	}

	targetLang := target.Lang()
	coffee := u.IsCoffeeSearch
	b.storage.WithUser(u.ID, func(x *User) {
		x.PartnerID = targetID
		x.State = StateInChat
		x.IsCoffeeSearch = false
	})
	b.storage.WithUser(targetID, func(x *User) {
		x.PartnerID = u.ID
		x.State = StateInChat
	})
	if coffee {
		ends := time.Now().Add(CoffeeChatDuration)
		b.storage.WithUser(u.ID, func(x *User) { x.ChatEndsAt = ends; x.IsCoffeeChat = true })
		b.storage.WithUser(targetID, func(x *User) { x.ChatEndsAt = ends; x.IsCoffeeChat = true })
	}

	b.answerCallback(cq.ID, t(lang, "msg_connected"))

	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		tf(lang, "fmt_now_chatting_with", shortName(target, lang)))
	b.send(edit)

	b.sendText(targetID,
		tf(targetLang, "fmt_partner_wants_chat", shortName(u, targetLang)),
		chatKeyboard(targetLang), false)

	// Suggest sharing live location.
	b.sendText(u.ID,
		t(lang, "msg_live_loc_tip"),
		chatKeyboard(lang), true)
	b.sendText(targetID,
		t(targetLang, "msg_live_loc_tip"),
		chatKeyboard(targetLang), true)

	// Icebreaker
	idx := int(time.Now().UnixNano() % int64(len(Icebreakers)))
	b.sendText(u.ID, t(lang, "msg_icebreaker_prefix")+icebreaker(lang, idx), nil, true)
	b.sendText(targetID, t(targetLang, "msg_icebreaker_prefix")+icebreaker(targetLang, idx), nil, true)
}

// filterMatches returns nearby users that:
//   - are within radiusKm
//   - are not the requester
//   - have a gender the requester is looking for (u.LookingFor)
//   - are themselves looking for the requester's gender (mutual match)
//   - are not already in a chat
//   - are not blocked
//   - are awake in their local timezone
func (b *Bot) filterMatches(u *User, radiusKm float64) []*User {
	nearby := b.storage.Nearby(u.ID, u.Latitude, u.Longitude, radiusKm)
	out := make([]*User, 0, len(nearby))
	for _, c := range nearby {
		if !genderMatches(u.LookingFor, c.Gender) {
			continue
		}
		if !genderMatches(c.LookingFor, u.Gender) {
			continue
		}
		if !c.IsAwake() {
			continue
		}
		out = append(out, c)
	}
	return out
}

func genderMatches(filter, target string) bool {
	switch filter {
	case GenderAny, "":
		return true
	default:
		return filter == target
	}
}

// ---------------------------------------------------------------------------
// Chat
// ---------------------------------------------------------------------------

func (b *Bot) relayToPartner(msg *tgbotapi.Message) {
	u, ok := b.storage.Get(msg.From.ID)
	if !ok || u.PartnerID == 0 {
		return
	}
	lang := u.Lang()
	partner, ok := b.storage.Get(u.PartnerID)
	if !ok {
		b.sendText(msg.Chat.ID, t(lang, "msg_partner_left"), mainMenu(lang), false)
		b.storage.WithUser(u.ID, func(x *User) {
			x.State = StateIdle
			x.PartnerID = 0
		})
		return
	}

	// Translate text messages when languages differ.
	if msg.Text != "" && b.translate != nil &&
		u.Language != "" && partner.Language != "" && u.Language != partner.Language {
		translated, err := b.translate.Translate(
			context.Background(), msg.Text, u.Language, partner.Language)
		if err == nil && translated != "" {
			b.sendText(partner.ID, translated, nil, false)
			return
		}
		LogTranslateErr(err)
	}

	// Default: copy message as-is (text, photo, voice, location — including
	// live location, since CopyMessage preserves the message kind).
	copyMsg := tgbotapi.NewCopyMessage(partner.ID, msg.Chat.ID, msg.MessageID)
	if _, err := b.api.Send(copyMsg); err != nil {
		log.Printf("[relay] copy failed: %v", err)
		b.sendText(msg.Chat.ID, t(lang, "msg_msg_not_delivered"), nil, false)
	}
}

func (b *Bot) onEndChatButton(cq *tgbotapi.CallbackQuery) {
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, "")
		return
	}
	lang := u.Lang()
	if u.State == StateInChat {
		b.endChat(cq.From.ID, cq.Message.Chat.ID, false)
		b.answerCallback(cq.ID, t(lang, "label_cancelled"))
		return
	}
	// Search-list cancel button: just reset state.
	b.storage.WithUser(u.ID, func(x *User) {
		x.State = StateIdle
		x.IsCoffeeSearch = false
	})
	b.answerCallback(cq.ID, t(lang, "label_cancelled"))
	b.sendText(cq.Message.Chat.ID, t(lang, "msg_cancelled_dot"), mainMenu(lang), false)
}

// onEndHereButton ends the chat (with partner notification) from the
// post-coffee "Continue?" prompt.
func (b *Bot) onEndHereButton(cq *tgbotapi.CallbackQuery, partnerIDStr string) {
	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 64)
	if err != nil {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_id"))
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok || u.PartnerID != partnerID || u.State != StateInChat {
		b.answerCallback(cq.ID, t(u.Lang(), "msg_chat_already_ended"))
		return
	}
	b.answerCallback(cq.ID, t(u.Lang(), "label_cancelled"))
	b.endChat(cq.From.ID, cq.Message.Chat.ID, true)
}

func (b *Bot) endChat(userID, chatID int64, notifyPartner bool) {
	var (
		partnerID        int64
		partnerCopy      *User
		wasCoffeeChat    bool
		partnerWasCoffee bool
	)
	u, ok := b.storage.Get(userID)
	if !ok {
		return
	}
	lang := u.Lang()
	if u.State != StateInChat || u.PartnerID == 0 {
		b.sendText(chatID, t(lang, "msg_not_in_chat"), mainMenu(lang), false)
		return
	}

	partnerID = u.PartnerID
	wasCoffeeChat = u.IsCoffeeChat
	if p, ok := b.storage.Get(partnerID); ok {
		partnerCopy = p
		partnerWasCoffee = p.IsCoffeeChat
		b.storage.WithUser(partnerID, func(x *User) {
			x.PartnerID = 0
			x.State = StateIdle
			x.ChatEndsAt = time.Time{}
			x.IsCoffeeChat = false
		})
	}

	b.storage.WithUser(userID, func(x *User) {
		x.PartnerID = 0
		x.State = StateIdle
		x.ChatEndsAt = time.Time{}
		x.IsCoffeeChat = false
	})

	if notifyPartner && partnerID != 0 {
		b.sendText(partnerID, t(partnerCopy.Lang(), "msg_partner_left_chat"), mainMenu(partnerCopy.Lang()), false)
	}

	// Update chat counts and check achievements.
	now := time.Now()
	var newlyU, newlyP []Achievement
	b.storage.WithUser(userID, func(x *User) {
		if partnerCopy != nil {
			x.ChatCount++
		}
		newlyU = checkAchievements(x, partnerCopy, now, wasCoffeeChat)
	})
	if partnerCopy != nil {
		b.storage.WithUser(partnerID, func(x *User) {
			x.ChatCount++
			newlyP = checkAchievements(x, u, now, partnerWasCoffee)
		})
	}

	// Prompt both sides to rate (skip if no real partner).
	if partnerCopy != nil {
		b.storage.WithUser(userID, func(x *User) {
			x.PendingRatingFor = partnerID
			x.State = StateRatePartner
		})
		b.storage.WithUser(partnerID, func(x *User) {
			x.PendingRatingFor = userID
			x.State = StateRatePartner
		})

		b.sendText(chatID,
			tf(lang, "fmt_chat_with_ended", shortName(partnerCopy, lang)),
			withSkip(ratingKeyboard(partnerID)), false)
		b.sendText(partnerID,
			tf(partnerCopy.Lang(), "fmt_chat_with_ended", shortName(u, partnerCopy.Lang())),
			withSkip(ratingKeyboard(userID)), false)
	} else {
		b.sendText(chatID,
			t(lang, "msg_chat_ended_find_new"),
			mainMenu(lang), true)
	}

	// Notify about new achievements.
	pLang := partnerCopy.Lang()
	for _, a := range newlyU {
		b.sendText(userID,
			tf(lang, "msg_achievement_unlocked", a.Emoji, achTitle(lang, a.ID), achDesc(lang, a.ID)),
			mainMenu(lang), true)
	}
	for _, a := range newlyP {
		b.sendText(partnerID,
			tf(pLang, "msg_achievement_unlocked", a.Emoji, achTitle(pLang, a.ID), achDesc(pLang, a.ID)),
			mainMenu(pLang), true)
	}
}

// onContinueAction handles the post-coffee-chat "Keep chatting?" answer.
func (b *Bot) onContinueAction(cq *tgbotapi.CallbackQuery, partnerIDStr string) {
	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 64)
	if err != nil {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_id"))
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(UILangEn, "err_please_start"))
		return
	}
	lang := u.Lang()
	if u.PartnerID != partnerID || u.State != StateInChat {
		b.answerCallback(cq.ID, t(lang, "msg_chat_already_ended"))
		return
	}
	// Convert coffee chat into a regular chat.
	b.storage.WithUser(u.ID, func(x *User) {
		x.IsCoffeeChat = false
		x.ChatEndsAt = time.Time{}
	})
	if p, ok := b.storage.Get(partnerID); ok {
		b.storage.WithUser(partnerID, func(x *User) {
			x.IsCoffeeChat = false
			x.ChatEndsAt = time.Time{}
		})
		_ = p
	}
	b.answerCallback(cq.ID, t(lang, "msg_continuing"))
	b.sendText(cq.Message.Chat.ID, t(lang, "msg_continuing_no_timer"), chatKeyboard(lang), false)
	pLang := b.partnerLang(partnerID)
	b.sendText(partnerID, t(pLang, "msg_continuing_no_timer"), chatKeyboard(pLang), false)
}

// partnerLang fetches a user's Lang() safely.
func (b *Bot) partnerLang(id int64) string {
	if u, ok := b.storage.Get(id); ok {
		return u.Lang()
	}
	return UILangEn
}

// onRateAction handles a star-rating tap or "skip" from the rating prompt.
func (b *Bot) onRateAction(cq *tgbotapi.CallbackQuery, parts []string) {
	// parts = ["rate", "<partnerID>", "<stars>"]  OR  ["rate", "skip"]
	if parts[1] == "skip" {
		b.storage.WithUser(cq.From.ID, func(x *User) {
			x.PendingRatingFor = 0
			x.State = StateIdle
		})
		u, _ := b.storage.Get(cq.From.ID)
		lang := UILangEn
		if u != nil {
			lang = u.Lang()
		}
		b.answerCallback(cq.ID, t(lang, "msg_skipped"))
		b.sendText(cq.Message.Chat.ID, t(lang, "msg_no_problem_menu"), mainMenu(lang), false)
		return
	}
	if len(parts) < 3 {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_rating"))
		return
	}
	partnerID, err1 := strconv.ParseInt(parts[1], 10, 64)
	stars, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || stars < 1 || stars > 5 {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_rating"))
		return
	}

	u, ok := b.storage.Get(cq.From.ID)
	if !ok || u.PendingRatingFor != partnerID {
		b.answerCallback(cq.ID, t(UILangEn, "err_already_rated"))
		return
	}
	lang := u.Lang()

	// The rating is stored on the partner.
	var newAch []Achievement
	b.storage.WithUser(partnerID, func(x *User) {
		if x.RatedBy == nil {
			x.RatedBy = make(map[int64]bool)
		}
		if x.RatedBy[cq.From.ID] {
			return
		}
		x.RatedBy[cq.From.ID] = true
		x.RatingSum += stars
		x.RatingCount++
		newAch = checkAchievements(x, u, time.Now(), false)
	})

	b.storage.WithUser(cq.From.ID, func(x *User) {
		x.PendingRatingFor = 0
		x.State = StateIdle
	})

	b.answerCallback(cq.ID, tf(lang, "fmt_rated_stars", stars))
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		tf(lang, "fmt_you_rated", stars))
	b.send(edit)
	b.sendText(cq.Message.Chat.ID, t(lang, "msg_thanks_feedback"), mainMenu(lang), false)

	pLang := b.partnerLang(partnerID)
	for _, a := range newAch {
		b.sendText(partnerID,
			tf(pLang, "msg_achievement_unlocked", a.Emoji, achTitle(pLang, a.ID), achDesc(pLang, a.ID)),
			mainMenu(pLang), true)
	}
}

// ---------------------------------------------------------------------------
// Block / Report
// ---------------------------------------------------------------------------

func (b *Bot) onMenuReport(msg *tgbotapi.Message) {
	u, ok := b.storage.Get(msg.From.ID)
	if !ok || u.State != StateInChat || u.PartnerID == 0 {
		return
	}
	lang := u.Lang()
	partner, ok := b.storage.Get(u.PartnerID)
	if !ok {
		return
	}
	b.sendText(msg.Chat.ID,
		tf(lang, "fmt_report_confirm", shortName(partner, lang)),
		confirmKeyboard(
			fmt.Sprintf("report:confirm:%d", u.PartnerID),
			"endchat",
			lang,
		), false)
}

func (b *Bot) onMenuBlock(msg *tgbotapi.Message) {
	u, ok := b.storage.Get(msg.From.ID)
	if !ok || u.State != StateInChat || u.PartnerID == 0 {
		return
	}
	lang := u.Lang()
	partner, ok := b.storage.Get(u.PartnerID)
	if !ok {
		return
	}
	b.sendText(msg.Chat.ID,
		tf(lang, "fmt_block_confirm", shortName(partner, lang)),
		confirmKeyboard(
			fmt.Sprintf("block:confirm:%d", u.PartnerID),
			"endchat",
			lang,
		), false)
}

func (b *Bot) handleConfirm(cq *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 3 {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_request"))
		return
	}
	action, targetStr := parts[0], parts[2]
	target, err := strconv.ParseInt(targetStr, 10, 64)
	if err != nil {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_id"))
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(UILangEn, "err_please_start"))
		return
	}
	lang := u.Lang()

	switch action {
	case "report":
		count := b.storage.IncrementReport(target)
		b.answerCallback(cq.ID, t(lang, "msg_reported"))
		edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
			t(lang, "msg_reported"))
		b.send(edit)
		if count >= ReportThreshold {
			b.storage.WithUser(target, func(x *User) {
				x.SuspendedUntil = time.Now().Add(SuspendDuration)
			})
			b.sendText(target,
				t(b.partnerLang(target), "msg_suspended"),
				mainMenu(b.partnerLang(target)), false)
		}
		// End the chat for both sides.
		b.endChat(u.ID, cq.Message.Chat.ID, true)

	case "block":
		b.storage.WithUser(u.ID, func(x *User) {
			if x.Blocks == nil {
				x.Blocks = make(map[int64]bool)
			}
			x.Blocks[target] = true
		})
		b.answerCallback(cq.ID, t(lang, "msg_blocked"))
		edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
			t(lang, "msg_blocked"))
		b.send(edit)
		b.endChat(u.ID, cq.Message.Chat.ID, true)

	default:
		b.answerCallback(cq.ID, t(lang, "err_unknown_action"))
	}
}

// ---------------------------------------------------------------------------
// Profile
// ---------------------------------------------------------------------------

func (b *Bot) showProfile(chatID int64, u *User) {
	text := formatProfile(u)
	b.sendText(chatID, text, profileKeyboard(u.Lang()), true)
}

func formatProfile(u *User) string {
	lang := u.Lang()
	name := shortName(u, lang)
	alias := u.Alias
	if alias == "" {
		alias = t(lang, "p_not_set")
	}
	bio := u.Bio
	if bio == "" {
		bio = t(lang, "p_not_set")
	}
	interests := t(lang, "p_none")
	if len(u.Interests) > 0 {
		interests = strings.Join(u.SortedInterests(), ", ")
	}
	langLabel := u.Language
	if langLabel == "" {
		langLabel = t(lang, "p_not_set_lang")
	} else if name, ok := SupportedLanguages[langLabel]; ok {
		langLabel = name
	}
	wake := t(lang, "p_not_set_wake")
	if u.Timezone != "" {
		wake = tf(lang, "fmt_p_wake_detail", u.WakeFrom, u.WakeTo, u.Timezone)
	}
	rating := t(lang, "p_no_ratings")
	if u.RatingCount > 0 {
		rating = tf(lang, "fmt_p_rating_detail", u.AverageRating(), u.RatingCount)
	}
	notify := t(lang, "p_off")
	if u.NotifyWhenNearby {
		notify = tf(lang, "fmt_p_on", u.NotifyRadius)
	}

	return fmt.Sprintf(
		"%s\n\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s",
		t(lang, "p_your_profile"),
		t(lang, "p_name"), name,
		t(lang, "p_alias"), alias,
		t(lang, "p_bio"), bio,
		t(lang, "p_interests"), interests,
		t(lang, "p_language"), langLabel,
		t(lang, "p_wake_hours"), wake,
		t(lang, "p_rating"), rating,
		t(lang, "p_notifications"), notify)
}

func (b *Bot) onProfileAction(cq *tgbotapi.CallbackQuery, action string) {
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(UILangEn, "err_please_start"))
		return
	}
	lang := u.Lang()

	switch action {
	case "alias":
		b.storage.WithUser(cq.From.ID, func(x *User) { x.State = StateNeedAlias })
		b.answerCallback(cq.ID, "")
		b.sendText(cq.Message.Chat.ID,
			t(lang, "msg_alias_prompt"), nil, false)

	case "bio":
		b.storage.WithUser(cq.From.ID, func(x *User) { x.State = StateNeedBio })
		b.answerCallback(cq.ID, "")
		b.sendText(cq.Message.Chat.ID,
			t(lang, "msg_bio_prompt"), nil, false)

	case "interests":
		u, _ := b.storage.Get(cq.From.ID)
		b.storage.WithUser(cq.From.ID, func(x *User) { x.State = StateNeedInterests })
		b.answerCallback(cq.ID, "")
		page := 0
		if u != nil {
			page = len(u.Interests) / 12
		}
		b.sendText(cq.Message.Chat.ID,
			t(lang, "msg_interests_prompt"),
			interestsKeyboard(interestSet(u.Interests), page), true)

	case "photo":
		b.storage.WithUser(cq.From.ID, func(x *User) { x.State = StateNeedPhoto })
		b.answerCallback(cq.ID, "")
		b.sendText(cq.Message.Chat.ID,
			t(lang, "msg_photo_prompt"), nil, false)

	default:
		b.answerCallback(cq.ID, t(lang, "err_unknown_profile_action"))
	}
}

func (b *Bot) onInterestAction(cq *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 2 {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_request"))
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(UILangEn, "err_please_start"))
		return
	}
	lang := u.Lang()

	selected := interestSet(u.Interests)
	switch parts[1] {
	case "toggle":
		if len(parts) < 3 {
			b.answerCallback(cq.ID, t(lang, "err_bad_tag"))
			return
		}
		tag := parts[2]
		if !validInterest(tag) {
			b.answerCallback(cq.ID, t(lang, "err_bad_tag"))
			return
		}
		if selected[tag] {
			delete(selected, tag)
		} else {
			selected[tag] = true
		}
		var newList []string
		for _, t := range AllInterests {
			if selected[t] {
				newList = append(newList, t)
			}
		}
		b.storage.WithUser(u.ID, func(x *User) { x.Interests = newList })
		b.answerCallback(cq.ID, "")
		edit := tgbotapi.NewEditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID,
			interestsKeyboard(selected, 0))
		b.send(edit)

	case "page":
		page, err := strconv.Atoi(parts[2])
		if err != nil {
			b.answerCallback(cq.ID, t(lang, "err_bad_page"))
			return
		}
		b.answerCallback(cq.ID, "")
		edit := tgbotapi.NewEditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID,
			interestsKeyboard(selected, page))
		b.send(edit)

	case "done":
		b.storage.WithUser(u.ID, func(x *User) { x.State = StateIdle })
		b.answerCallback(cq.ID, t(lang, "msg_saved"))
		b.sendText(cq.Message.Chat.ID, t(lang, "msg_interests_saved"), mainMenu(lang), false)

	case "clear":
		b.storage.WithUser(u.ID, func(x *User) { x.Interests = nil })
		b.answerCallback(cq.ID, t(lang, "msg_saved"))
		edit := tgbotapi.NewEditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID,
			interestsKeyboard(map[string]bool{}, 0))
		b.send(edit)

	default:
		b.answerCallback(cq.ID, t(lang, "err_unknown_interest_action"))
	}
}

// handlePhotoInput handles StateNeedPhoto.
func (b *Bot) handlePhotoInput(msg *tgbotapi.Message, u *User) {
	lang := u.Lang()
	if len(msg.Photo) == 0 {
		b.sendText(msg.Chat.ID, t(lang, "msg_send_photo_or_skip"), nil, false)
		return
	}
	best := msg.Photo[len(msg.Photo)-1]
	b.storage.WithUser(u.ID, func(x *User) {
		x.PhotoID = best.FileID
		x.State = StateIdle
	})
	b.sendText(msg.Chat.ID, t(lang, "msg_photo_updated"), mainMenu(lang), false)
}

// ---------------------------------------------------------------------------
// Language
// ---------------------------------------------------------------------------

func (b *Bot) onLanguagePick(cq *tgbotapi.CallbackQuery, code string) {
	if !isValidUILang(code) {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_language"))
		return
	}

	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(code, "err_please_start"))
		return
	}

	isInitial := u.State == StateNeedUILang

	b.storage.WithUser(cq.From.ID, func(x *User) {
		x.UILang = code
		x.Language = code
		if isInitial {
			x.State = StateNeedLocation
		} else {
			x.State = StateIdle
		}
	})

	b.answerCallback(cq.ID, t(code, "msg_saved"))
	name := UILangNames[code]

	if isInitial {
		// Show the intro + location request in the chosen language.
		edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
			tf(code, "fmt_lang_set", name))
		edit.ParseMode = "Markdown"
		b.send(edit)
		b.sendText(cq.Message.Chat.ID,
			tf(code, "msg_start_intro", u.FirstName),
			requestLocationKeyboard(code), true)
		return
	}

	// Language changed by an already-registered user.
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		tf(code, "fmt_lang_set", name))
	edit.ParseMode = "Markdown"
	b.send(edit)
	b.sendText(cq.Message.Chat.ID,
		t(code, "msg_lang_translation_info"),
		mainMenu(code), false)
}

// ---------------------------------------------------------------------------
// Wake hours
// ---------------------------------------------------------------------------

func (b *Bot) startWakeHoursFlow(chatID int64, u *User) {
	lang := u.Lang()
	if u.Timezone == "" {
		b.storage.WithUser(u.ID, func(x *User) { x.State = StateNeedTimezone })
		b.sendText(chatID,
			t(lang, "msg_pick_timezone"), timezoneKeyboard(), false)
		return
	}
	b.storage.WithUser(u.ID, func(x *User) { x.State = StateNeedWakeFrom })
	b.sendText(chatID,
		tf(lang, "fmt_timezone_set", u.Timezone), nil, true)
}

func (b *Bot) onTimezonePick(cq *tgbotapi.CallbackQuery, tz string) {
	if _, err := time.LoadLocation(tz); err != nil {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_timezone"))
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(UILangEn, "err_please_start"))
		return
	}
	lang := u.Lang()
	b.storage.WithUser(cq.From.ID, func(x *User) {
		x.Timezone = tz
		x.State = StateNeedWakeFrom
	})
	b.answerCallback(cq.ID, t(lang, "msg_saved"))
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		tf(lang, "fmt_timezone_set", tz))
	edit.ParseMode = "Markdown"
	b.send(edit)
	b.sendText(cq.Message.Chat.ID,
		t(lang, "msg_send_wake_from"), nil, false)
}

func (b *Bot) onWakePick(cq *tgbotapi.CallbackQuery, scope, val string) {
	h, err := strconv.Atoi(val)
	if err != nil || h < 0 || h > 23 {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_hour"))
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(UILangEn, "err_please_start"))
		return
	}
	lang := u.Lang()
	switch scope {
	case "from":
		b.storage.WithUser(cq.From.ID, func(x *User) {
			x.WakeFrom = h
			x.State = StateNeedWakeTo
		})
		b.answerCallback(cq.ID, fmt.Sprintf("Wake at %d", h))
		b.sendText(cq.Message.Chat.ID,
			tf(lang, "fmt_wake_from_set", h),
			nil, true)
	case "to":
		var wakeFrom int
		b.storage.WithUser(cq.From.ID, func(x *User) {
			wakeFrom = x.WakeFrom
			x.WakeTo = h
			x.State = StateIdle
		})
		b.answerCallback(cq.ID, fmt.Sprintf("Sleep at %d", h))
		b.sendText(cq.Message.Chat.ID,
			tf(lang, "fmt_wake_hours_set", wakeFrom, h),
			mainMenu(lang), true)
	default:
		b.answerCallback(cq.ID, t(lang, "err_unknown"))
	}
}

// ---------------------------------------------------------------------------
// Notifications
// ---------------------------------------------------------------------------

func (b *Bot) toggleNotifications(u *User, chatID int64) {
	lang := u.Lang()
	if u.Latitude == 0 && u.Longitude == 0 {
		b.sendText(chatID, t(lang, "msg_share_location_for_notify"), requestLocationKeyboard(lang), false)
		return
	}
	if u.NotifyWhenNearby {
		b.storage.WithUser(u.ID, func(x *User) {
			x.NotifyWhenNearby = false
		})
		b.sendText(chatID, t(lang, "msg_notifications_disabled"), mainMenu(lang), false)
		return
	}
	b.storage.WithUser(u.ID, func(x *User) {
		x.State = StateNeedNotifyRadius
	})
	b.sendText(chatID,
		t(lang, "msg_notify_prompt"),
		notifyRadiusKeyboard(), true)
}

func (b *Bot) onNotifyRadiusPick(cq *tgbotapi.CallbackQuery, raw string) {
	r, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		b.answerCallback(cq.ID, t(UILangEn, "err_bad_radius"))
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, t(UILangEn, "err_please_start"))
		return
	}
	lang := u.Lang()
	b.storage.WithUser(cq.From.ID, func(x *User) {
		x.NotifyWhenNearby = true
		x.NotifyRadius = r
		x.State = StateIdle
		x.LastNotifiedAt = time.Time{}
	})
	b.answerCallback(cq.ID, fmt.Sprintf("%.0f km", r))
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		tf(lang, "fmt_notify_set", r))
	b.send(edit)
	b.sendText(cq.Message.Chat.ID,
		t(lang, "msg_notify_turn_off"), mainMenu(lang), false)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func shortName(u *User, lang string) string {
	if u == nil {
		return t(lang, "label_stranger")
	}
	if u.Alias != "" {
		return u.Alias
	}
	if u.FirstName != "" {
		return u.FirstName
	}
	if u.Username != "" {
		return "@" + u.Username
	}
	return t(lang, "label_stranger")
}

func prettyGender(g, lang string) string {
	switch g {
	case GenderMale:
		return t(lang, "label_male")
	case GenderFemale:
		return t(lang, "label_female")
	case GenderAny:
		return t(lang, "label_both")
	default:
		return t(lang, "label_unknown_gender")
	}
}

func genderEmoji(g string) string {
	switch g {
	case GenderMale:
		return "♂"
	case GenderFemale:
		return "♀"
	default:
		return "•"
	}
}

func formatDistance(km float64) string {
	if km < 1 {
		return fmt.Sprintf("%.0f m", km*1000)
	}
	if km < 10 {
		return fmt.Sprintf("%.1f km", km)
	}
	return fmt.Sprintf("%.0f km", km)
}

func formatMatchLabel(c *User, dist float64, lang string) string {
	name := shortName(c, lang)
	parts := []string{
		name,
		formatDistance(dist),
		genderEmoji(c.Gender),
	}
	if len(c.Interests) > 0 {
		ints := c.SortedInterests()
		if len(ints) > 3 {
			ints = ints[:3]
		}
		parts = append(parts, strings.Join(ints, ", "))
	}
	if c.AverageRating() > 0 {
		parts = append(parts, fmt.Sprintf("%.1f⭐", c.AverageRating()))
	}
	return strings.Join(parts, " · ")
}

func interestSet(s []string) map[string]bool {
	out := make(map[string]bool, len(s))
	for _, tag := range s {
		out[tag] = true
	}
	return out
}

func validInterest(t string) bool {
	for _, x := range AllInterests {
		if x == t {
			return true
		}
	}
	return false
}

// withSkip appends a Skip button to the rating keyboard.
func withSkip(m tgbotapi.InlineKeyboardMarkup) tgbotapi.InlineKeyboardMarkup {
	rows := m.InlineKeyboard
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⏭ Skip rating", "rate:skip"),
	))
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// sendText is a small helper that fills ReplyMarkup, ParseMode and sends.
func (b *Bot) sendText(chatID int64, text string, markup interface{}, markdown bool) {
	m := tgbotapi.NewMessage(chatID, text)
	if markup != nil {
		m.ReplyMarkup = markup
	}
	if markdown {
		m.ParseMode = "Markdown"
	}
	b.send(m)
}
