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
		b.sendText(msg.Chat.ID, "Send /start to begin.", nil, false)
		return
	}

	// In chat: first check for chat-keyboard buttons (which would
	// otherwise be relayed as plain text to the partner).
	if u.State == StateInChat {
		switch strings.TrimSpace(msg.Text) {
		case "❌ End chat":
			b.endChat(u.ID, msg.Chat.ID, true)
			return
		case "🚫 Block":
			b.onMenuBlock(msg)
			return
		case "🚩 Report":
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
		b.sendText(msg.Chat.ID, "Send /start to begin.", nil, false)
		return
	}

	// While in chat only a few commands are meaningful.
	if u.State == StateInChat {
		switch cmd {
		case "end", "stop", "cancel":
			b.endChat(msg.From.ID, msg.Chat.ID, true)
		default:
			b.sendText(msg.Chat.ID,
				"You're in a chat — send messages to your partner, or /end to disconnect.",
				chatKeyboard, false)
		}
		return
	}

	switch cmd {
	case "start":
		b.cmdStart(msg)
	case "menu":
		b.sendText(msg.Chat.ID, "Main menu 👇", mainMenu, false)
	case "end", "stop", "cancel":
		b.sendText(msg.Chat.ID, "You're not in a chat right now.", mainMenu, false)
	case "profile":
		b.showProfile(msg.Chat.ID, u)
	case "achievements":
		b.sendText(msg.Chat.ID, achievementsText(u), nil, true)
	case "language":
		b.sendText(msg.Chat.ID, "Pick your language:", languageKeyboard(), false)
	case "hours":
		b.startWakeHoursFlow(msg.Chat.ID, u)
	case "notify":
		b.toggleNotifications(u, msg.Chat.ID)
	case "skip":
		b.handleSkip(msg.Chat.ID, u)
	default:
		b.sendText(msg.Chat.ID, "Unknown command. Try /start or /menu.", nil, false)
	}
}

func (b *Bot) handleCallback(cq *tgbotapi.CallbackQuery) {
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
			b.answerCallback(cq.ID, "Bad user id")
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
		b.answerCallback(cq.ID, "Unknown action")
	}
}

// ---------------------------------------------------------------------------
// /start and registration flow
// ---------------------------------------------------------------------------

func (b *Bot) cmdStart(msg *tgbotapi.Message) {
	uid := msg.From.ID

	existing, ok := b.storage.Get(uid)
	if ok && existing.Gender != "" && existing.Latitude != 0 {
		b.sendText(msg.Chat.ID, "Welcome back! 👋", mainMenu, false)
		return
	}

	u := &User{
		ID:        uid,
		FirstName: msg.From.FirstName,
		Username:  msg.From.UserName,
		State:     StateNeedLocation,
	}
	if ok {
		u.State = StateNeedLocation
		u.Gender = existing.Gender
		u.LookingFor = existing.LookingFor
	}
	b.storage.Upsert(u)

	hello := fmt.Sprintf(
		"Hey %s! 👋\n\n"+
			"I'm *NearFriend* — I help you meet people nearby for a chat.\n\n"+
			"1️⃣ Share your location with the button below.\n"+
			"2️⃣ Tell me your gender.\n"+
			"3️⃣ Tap *Find nearby friends* whenever you want to match.\n\n"+
			"_Tip: type /skip at any prompt to skip._",
		msg.From.FirstName,
	)
	b.sendText(msg.Chat.ID, hello, requestLocationKeyboard, true)
}

func (b *Bot) onLocation(msg *tgbotapi.Message) {
	u, ok := b.storage.Get(msg.From.ID)
	if !ok {
		u = &User{ID: msg.From.ID, FirstName: msg.From.FirstName, Username: msg.From.UserName}
	}
	u.Latitude = msg.Location.Latitude
	u.Longitude = msg.Location.Longitude
	u.LocationAt = time.Now()

	if u.State == StateNeedLocation {
		u.State = StateIdle
		b.storage.Upsert(u)

		b.sendText(msg.Chat.ID,
			"📍 Got your location! Now tell me your gender:",
			genderKeyboard("gender:self"), false)
		b.sendText(msg.Chat.ID, "Main menu 👇", mainMenu, false)
		return
	}

	b.storage.Upsert(u)
	if msg.Location.LivePeriod > 0 {
		b.sendText(msg.Chat.ID,
			"📍 Live location updated! (This only affects your profile — share live location with a chat partner from inside a chat.)",
			mainMenu, false)
		return
	}
	b.sendText(msg.Chat.ID, "📍 Location updated!", mainMenu, false)
}

func (b *Bot) promptLocation(chatID int64) {
	b.sendText(chatID, "Tap the button to share your current location:",
		requestLocationKeyboard, false)
}

// ---------------------------------------------------------------------------
// Stateful text / photo input
// ---------------------------------------------------------------------------

func (b *Bot) handleStatefulInput(msg *tgbotapi.Message, u *User) {
	// Menu buttons work from any state — let users jump around.
	switch strings.TrimSpace(msg.Text) {
	case "👤 My profile":
		b.showProfile(msg.Chat.ID, u)
		return
	case "🚻 Set my gender":
		b.sendText(msg.Chat.ID, "What's your gender?", genderKeyboard("gender:self"), false)
		return
	case "📍 Update my location":
		b.promptLocation(msg.Chat.ID)
		return
	case "🔍 Find nearby friends":
		b.startSearch(u, msg.Chat.ID, false)
		return
	case "☕ Coffee chat (15 min)":
		b.startSearch(u, msg.Chat.ID, true)
		return
	case "🌐 Language":
		b.sendText(msg.Chat.ID, "Pick your language:", languageKeyboard(), false)
		return
	case "⏰ Wake hours":
		b.startWakeHoursFlow(msg.Chat.ID, u)
		return
	case "🔔 Notify me":
		b.toggleNotifications(u, msg.Chat.ID)
		return
	case "🏆 Achievements":
		b.sendText(msg.Chat.ID, achievementsText(u), nil, true)
		return
	case "❌ End chat":
		b.endChat(u.ID, msg.Chat.ID, true)
		return
	}

	switch u.State {
	case StateNeedGender:
		b.sendText(msg.Chat.ID,
			"Please pick your gender with the buttons below 👇", nil, false)
		b.sendText(msg.Chat.ID, "What's your gender?", genderKeyboard("gender:self"), false)

	case StateNeedSearchGender:
		b.sendText(msg.Chat.ID, "Pick who you want to chat with 👇", nil, false)

	case StateNeedRadius:
		b.sendText(msg.Chat.ID, "Pick a search radius 👇", nil, false)

	case StateNeedAlias:
		alias := strings.TrimSpace(msg.Text)
		if alias == "" {
			b.sendText(msg.Chat.ID, "Alias can't be empty. Try again or /skip.", nil, false)
			return
		}
		if len(alias) > 32 {
			alias = alias[:32]
		}
		b.storage.WithUser(u.ID, func(x *User) {
			x.Alias = alias
			x.State = StateIdle
		})
		b.sendText(msg.Chat.ID, fmt.Sprintf("✅ Alias set to *%s*", alias), mainMenu, true)

	case StateNeedBio:
		bio := strings.TrimSpace(msg.Text)
		if len(bio) > 200 {
			bio = bio[:200]
		}
		b.storage.WithUser(u.ID, func(x *User) {
			x.Bio = bio
			x.State = StateIdle
		})
		b.sendText(msg.Chat.ID, "✅ Bio saved.", mainMenu, false)

	case StateNeedInterests:
		b.sendText(msg.Chat.ID,
			"Tap the tags below to toggle them, then press *Done*.",
			interestsKeyboard(interestSet(u.Interests), 0), true)

	case StateNeedPhoto:
		b.handlePhotoInput(msg, u)

	case StateNeedLanguage:
		b.sendText(msg.Chat.ID, "Pick your language:", languageKeyboard(), false)

	case StateNeedTimezone:
		b.sendText(msg.Chat.ID, "Pick your timezone:", timezoneKeyboard(), false)

	case StateNeedWakeFrom:
		h, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || h < 0 || h > 23 {
			b.sendText(msg.Chat.ID, "Send a number 0-23, or /skip.", nil, false)
			return
		}
		b.storage.WithUser(u.ID, func(x *User) {
			x.WakeFrom = h
			x.State = StateNeedWakeTo
		})
		b.sendText(msg.Chat.ID, fmt.Sprintf("Wake from *%d:00*. Now send the hour you usually go to sleep (0-23):", h), nil, true)

	case StateNeedWakeTo:
		h, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || h < 0 || h > 23 {
			b.sendText(msg.Chat.ID, "Send a number 0-23, or /skip.", nil, false)
			return
		}
		b.storage.WithUser(u.ID, func(x *User) {
			x.WakeTo = h
			x.State = StateIdle
		})
		b.sendText(msg.Chat.ID,
			fmt.Sprintf("✅ Wake hours: *%02d:00 - %02d:00*", u.WakeFrom, h),
			mainMenu, true)

	case StateNeedNotifyRadius:
		b.sendText(msg.Chat.ID, "Pick a notification radius 👇", notifyRadiusKeyboard(), false)

	case StateRatePartner:
		b.sendText(msg.Chat.ID, "Tap a star to rate your chat 👇",
			ratingKeyboard(u.PendingRatingFor), false)

	default:
		b.sendText(msg.Chat.ID, "Use the menu below 👇", mainMenu, false)
	}
}

func (b *Bot) handleSkip(chatID int64, u *User) {
	switch u.State {
	case StateNeedAlias, StateNeedBio, StateNeedInterests,
		StateNeedLanguage, StateNeedTimezone, StateNeedWakeFrom, StateNeedWakeTo:
		b.storage.WithUser(u.ID, func(x *User) {
			x.State = StateIdle
		})
		b.sendText(chatID, "Skipped.", mainMenu, false)
	case StateRatePartner:
		b.storage.WithUser(u.ID, func(x *User) {
			x.PendingRatingFor = 0
			x.State = StateIdle
		})
		b.sendText(chatID, "Skipped rating.", mainMenu, false)
	default:
		b.sendText(chatID, "Nothing to skip.", mainMenu, false)
	}
}

// ---------------------------------------------------------------------------
// Gender
// ---------------------------------------------------------------------------

func (b *Bot) onGenderPick(cq *tgbotapi.CallbackQuery, scope, value string) {
	if value != GenderMale && value != GenderFemale && value != GenderAny {
		b.answerCallback(cq.ID, "Bad value")
		return
	}

	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, "Please /start first")
		return
	}

	switch scope {
	case "self":
		b.storage.WithUser(u.ID, func(x *User) {
			x.Gender = value
		})
		b.answerCallback(cq.ID, "Saved!")

		edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
			fmt.Sprintf("✅ Your gender is set to *%s*.", prettyGender(value)))
		edit.ParseMode = "Markdown"
		b.send(edit)
		b.sendText(cq.Message.Chat.ID,
			"You're all set! Tap *Find nearby friends* when you're ready.",
			mainMenu, true)

	case "search":
		b.storage.WithUser(u.ID, func(x *User) {
			x.LookingFor = value
			x.State = StateNeedRadius
		})
		b.answerCallback(cq.ID, "Got it")

		edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
			fmt.Sprintf("Looking for: *%s*. Now pick a radius 👇", prettyGender(value)))
		edit.ParseMode = "Markdown"
		b.send(edit)
		b.sendText(cq.Message.Chat.ID, "Search within…", radiusKeyboard(), false)

	default:
		b.answerCallback(cq.ID, "Unknown gender scope")
	}
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func (b *Bot) startSearch(u *User, chatID int64, coffee bool) {
	if u.Latitude == 0 && u.Longitude == 0 {
		b.sendText(chatID, "Please share your location first.",
			requestLocationKeyboard, false)
		return
	}
	if u.Gender == "" {
		b.sendText(chatID, "Set your gender first:",
			genderKeyboard("gender:self"), false)
		return
	}

	b.storage.WithUser(u.ID, func(x *User) {
		x.State = StateNeedSearchGender
		x.IsCoffeeSearch = coffee
	})

	prompt := "Who do you want to chat with?"
	if coffee {
		prompt = "☕ Coffee chat (15 min) — who do you want to chat with?"
	}
	b.sendText(chatID, prompt, genderKeyboard("gender:search"), false)
}

func (b *Bot) onRadiusPick(cq *tgbotapi.CallbackQuery, raw string) {
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, "Please /start first")
		return
	}
	if u.State != StateNeedRadius {
		b.answerCallback(cq.ID, "Not waiting for radius")
		return
	}

	radius, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		b.answerCallback(cq.ID, "Bad radius")
		return
	}

	candidates := b.filterMatches(u, radius)
	coffee := u.IsCoffeeSearch

	b.storage.WithUser(u.ID, func(x *User) {
		x.State = StateBrowsing
	})
	b.answerCallback(cq.ID, fmt.Sprintf("%.0f km", radius))

	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		fmt.Sprintf("Searching within %.0f km…", radius))
	b.send(edit)

	if len(candidates) == 0 {
		hint := "😕 No nearby matches. Try a bigger radius, or update your preferences."
		if coffee {
			hint = "😕 No nearby matches for a coffee chat. Try a bigger radius."
		}
		b.sendText(cq.Message.Chat.ID, hint, mainMenu, false)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, c := range candidates {
		dist := DistanceKm(u.Latitude, u.Longitude, c.Latitude, c.Longitude)
		label := formatMatchLabel(c, dist)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label,
				"nearby:"+strconv.FormatInt(c.ID, 10)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("↩️ Cancel", "endchat"),
	))

	title := fmt.Sprintf("Found *%d* nearby %s. Pick one 👇",
		len(candidates), prettyGender(u.LookingFor))
	if coffee {
		title = fmt.Sprintf("☕ Found *%d* nearby %s for a coffee chat. Pick one 👇",
			len(candidates), prettyGender(u.LookingFor))
	}
	b.sendText(cq.Message.Chat.ID, title, tgbotapi.NewInlineKeyboardMarkup(rows...), true)
}

func (b *Bot) onNearbyPick(cq *tgbotapi.CallbackQuery, targetID int64) {
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, "Please /start first")
		return
	}
	if u.State != StateBrowsing {
		b.answerCallback(cq.ID, "Pick a fresh search first")
		return
	}
	target, ok := b.storage.Get(targetID)
	if !ok {
		b.answerCallback(cq.ID, "User unavailable")
		return
	}
	if target.State == StateInChat {
		b.answerCallback(cq.ID, "Already chatting")
		return
	}
	if !target.IsAwake() {
		b.answerCallback(cq.ID, "They're sleeping right now 😴")
		return
	}

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

	b.answerCallback(cq.ID, "Connected!")

	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		fmt.Sprintf("✅ You're now chatting with %s. Say hi!", shortName(target)))
	b.send(edit)

	b.sendText(targetID,
		fmt.Sprintf("💬 %s wants to chat with you. Say hi!", shortName(u)),
		chatKeyboard, false)

	// Suggest sharing live location.
	b.sendText(u.ID,
		"📍 _Tip: share your live location with your match!_\n"+
			"_Tap the 📎 next to the text field → Location → Share My Live Location._",
		chatKeyboard, false)
	b.sendText(targetID,
		"📍 _Tip: share your live location with your match!_\n"+
			"_Tap the 📎 next to the text field → Location → Share My Live Location._",
		chatKeyboard, false)

	// Icebreaker
	q := Icebreakers[time.Now().UnixNano()%int64(len(Icebreakers))]
	b.sendText(u.ID, "💡 *Icebreaker:* "+q, nil, true)
	b.sendText(targetID, "💡 *Icebreaker:* "+q, nil, true)
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
	partner, ok := b.storage.Get(u.PartnerID)
	if !ok {
		b.sendText(msg.Chat.ID, "Your partner left. Chat ended.", mainMenu, false)
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
		b.sendText(msg.Chat.ID, "⚠️ Could not deliver your message.", nil, false)
	}
}

func (b *Bot) onEndChatButton(cq *tgbotapi.CallbackQuery) {
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, "")
		return
	}
	if u.State == StateInChat {
		b.endChat(cq.From.ID, cq.Message.Chat.ID, false)
		b.answerCallback(cq.ID, "Cancelled")
		return
	}
	// Search-list cancel button: just reset state.
	b.storage.WithUser(u.ID, func(x *User) {
		x.State = StateIdle
		x.IsCoffeeSearch = false
	})
	b.answerCallback(cq.ID, "Cancelled")
	b.sendText(cq.Message.Chat.ID, "Cancelled.", mainMenu, false)
}

// onEndHereButton ends the chat (with partner notification) from the
// post-coffee "Continue?" prompt.
func (b *Bot) onEndHereButton(cq *tgbotapi.CallbackQuery, partnerIDStr string) {
	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 64)
	if err != nil {
		b.answerCallback(cq.ID, "Bad id")
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok || u.PartnerID != partnerID || u.State != StateInChat {
		b.answerCallback(cq.ID, "Chat already ended")
		return
	}
	b.answerCallback(cq.ID, "Chat ended")
	b.endChat(cq.From.ID, cq.Message.Chat.ID, true)
}

func (b *Bot) endChat(userID, chatID int64, notifyPartner bool) {
	var (
		partnerID     int64
		partnerCopy   *User
		wasCoffeeChat bool
		partnerLang   string
	)
	u, ok := b.storage.Get(userID)
	if !ok {
		return
	}
	if u.State != StateInChat || u.PartnerID == 0 {
		b.sendText(chatID, "You're not in a chat right now.", mainMenu, false)
		return
	}

	partnerID = u.PartnerID
	wasCoffeeChat = u.IsCoffeeChat
	if p, ok := b.storage.Get(partnerID); ok {
		partnerCopy = p
		partnerLang = p.Language
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
		b.sendText(partnerID, "💔 Your chat partner has left the conversation.", mainMenu, false)
	}

	// Update chat counts and check achievements.
	now := time.Now()
	var newlyU, newlyP []Achievement
	b.storage.WithUser(userID, func(x *User) {
		if partnerCopy != nil {
			x.ChatCount++
		}
		newlyU = checkAchievements(x, partnerCopy, now)
	})
	if partnerCopy != nil {
		b.storage.WithUser(partnerID, func(x *User) {
			x.ChatCount++
			newlyP = checkAchievements(x, u, now)
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
			fmt.Sprintf("Chat with %s ended. How was it? Tap a star:",
				shortName(partnerCopy)),
			withSkip(ratingKeyboard(partnerID)), false)
		b.sendText(partnerID,
			fmt.Sprintf("Chat with %s ended. How was it? Tap a star:",
				shortName(u)),
			withSkip(ratingKeyboard(userID)), false)
	} else {
		b.sendText(chatID,
			"Chat ended. Tap *Find nearby friends* to start a new one.",
			mainMenu, true)
	}

	_ = wasCoffeeChat
	_ = partnerLang

	// Notify about new achievements.
	for _, a := range newlyU {
		b.sendText(userID,
			"🎉 Achievement unlocked!\n"+a.Emoji+" *"+a.Title+"* — "+a.Description,
			mainMenu, true)
	}
	for _, a := range newlyP {
		b.sendText(partnerID,
			"🎉 Achievement unlocked!\n"+a.Emoji+" *"+a.Title+"* — "+a.Description,
			mainMenu, true)
	}
}

// onContinueAction handles the post-coffee-chat "Keep chatting?" answer.
func (b *Bot) onContinueAction(cq *tgbotapi.CallbackQuery, partnerIDStr string) {
	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 64)
	if err != nil {
		b.answerCallback(cq.ID, "Bad id")
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, "Please /start first")
		return
	}
	if u.PartnerID != partnerID || u.State != StateInChat {
		b.answerCallback(cq.ID, "Chat already ended")
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
	b.answerCallback(cq.ID, "Continuing!")
	b.sendText(cq.Message.Chat.ID, "✅ Continuing the chat — no more timer.", chatKeyboard, false)
	b.sendText(partnerID, "✅ Continuing the chat — no more timer.", chatKeyboard, false)
}

// onRateAction handles a star-rating tap or "skip" from the rating prompt.
func (b *Bot) onRateAction(cq *tgbotapi.CallbackQuery, parts []string) {
	// parts = ["rate", "<partnerID>", "<stars>"]  OR  ["rate", "skip"]
	if parts[1] == "skip" {
		b.storage.WithUser(cq.From.ID, func(x *User) {
			x.PendingRatingFor = 0
			x.State = StateIdle
		})
		b.answerCallback(cq.ID, "Skipped")
		b.sendText(cq.Message.Chat.ID, "No problem — back to the main menu.", mainMenu, false)
		return
	}
	if len(parts) < 3 {
		b.answerCallback(cq.ID, "Bad rating")
		return
	}
	partnerID, err1 := strconv.ParseInt(parts[1], 10, 64)
	stars, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || stars < 1 || stars > 5 {
		b.answerCallback(cq.ID, "Bad rating")
		return
	}

	u, ok := b.storage.Get(cq.From.ID)
	if !ok || u.PendingRatingFor != partnerID {
		b.answerCallback(cq.ID, "Already rated or chat ended")
		return
	}

	// Record on the rater (RatedBy is for stats; not strictly needed here).
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
		newAch = checkAchievements(x, u, time.Now())
	})

	b.storage.WithUser(cq.From.ID, func(x *User) {
		x.PendingRatingFor = 0
		x.State = StateIdle
	})

	b.answerCallback(cq.ID, fmt.Sprintf("Rated %d ⭐", stars))
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		fmt.Sprintf("✅ You rated your partner %d ⭐", stars))
	b.send(edit)
	b.sendText(cq.Message.Chat.ID, "Thanks for the feedback!", mainMenu, false)

	for _, a := range newAch {
		b.sendText(partnerID,
			"🎉 Achievement unlocked!\n"+a.Emoji+" *"+a.Title+"* — "+a.Description,
			mainMenu, true)
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
	partner, ok := b.storage.Get(u.PartnerID)
	if !ok {
		return
	}
	b.sendText(msg.Chat.ID,
		fmt.Sprintf("🚩 Report %s for bad behavior?",
			shortName(partner)),
		confirmKeyboard(
			fmt.Sprintf("report:confirm:%d", u.PartnerID),
			"endchat",
		), false)
}

func (b *Bot) onMenuBlock(msg *tgbotapi.Message) {
	u, ok := b.storage.Get(msg.From.ID)
	if !ok || u.State != StateInChat || u.PartnerID == 0 {
		return
	}
	partner, ok := b.storage.Get(u.PartnerID)
	if !ok {
		return
	}
	b.sendText(msg.Chat.ID,
		fmt.Sprintf("🚫 Block %s? You won't see them in future searches.",
			shortName(partner)),
		confirmKeyboard(
			fmt.Sprintf("block:confirm:%d", u.PartnerID),
			"endchat",
		), false)
}

func (b *Bot) handleConfirm(cq *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 3 {
		b.answerCallback(cq.ID, "Bad request")
		return
	}
	action, targetStr := parts[0], parts[2]
	target, err := strconv.ParseInt(targetStr, 10, 64)
	if err != nil {
		b.answerCallback(cq.ID, "Bad id")
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, "Please /start first")
		return
	}

	switch action {
	case "report":
		count := b.storage.IncrementReport(target)
		b.answerCallback(cq.ID, "Reported")
		edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
			"🚩 Thanks — we'll review it.")
		b.send(edit)
		if count >= ReportThreshold {
			b.storage.WithUser(target, func(x *User) {
				x.SuspendedUntil = time.Now().Add(SuspendDuration)
			})
			b.sendText(target,
				"⏸ You've been temporarily suspended for 24h due to multiple reports. "+
					"You can still browse, but you won't appear in others' results.",
				mainMenu, false)
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
		b.answerCallback(cq.ID, "Blocked")
		edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
			"🚫 User blocked.")
		b.send(edit)
		b.endChat(u.ID, cq.Message.Chat.ID, true)

	default:
		b.answerCallback(cq.ID, "Unknown action")
	}
}

// ---------------------------------------------------------------------------
// Profile
// ---------------------------------------------------------------------------

func (b *Bot) showProfile(chatID int64, u *User) {
	text := formatProfile(u)
	b.sendText(chatID, text, profileKeyboard(), true)
}

func formatProfile(u *User) string {
	name := shortName(u)
	alias := u.Alias
	if alias == "" {
		alias = "_(not set)_"
	}
	bio := u.Bio
	if bio == "" {
		bio = "_(not set)_"
	}
	interests := "_(none)_"
	if len(u.Interests) > 0 {
		interests = strings.Join(u.SortedInterests(), ", ")
	}
	lang := u.Language
	if lang == "" {
		lang = "_(not set — chat will not be translated)_"
	} else if name, ok := SupportedLanguages[lang]; ok {
		lang = name
	}
	wake := "_(not set — always shown)_"
	if u.Timezone != "" {
		wake = fmt.Sprintf("`%02d:00 - %02d:00` (%s)", u.WakeFrom, u.WakeTo, u.Timezone)
	}
	rating := "_(no ratings yet)_"
	if u.RatingCount > 0 {
		rating = fmt.Sprintf("%.1f ⭐ (%d reviews)", u.AverageRating(), u.RatingCount)
	}
	notify := "off"
	if u.NotifyWhenNearby {
		notify = fmt.Sprintf("on, within %.0f km", u.NotifyRadius)
	}

	return fmt.Sprintf(
		"👤 *Your profile*\n\n"+
			"📛 *Name:* %s\n"+
			"🎭 *Alias:* %s\n"+
			"📝 *Bio:* %s\n"+
			"🏷️ *Interests:* %s\n"+
			"🌐 *Language:* %s\n"+
			"⏰ *Wake hours:* %s\n"+
			"⭐ *Rating:* %s\n"+
			"🔔 *Notifications:* %s",
		name, alias, bio, interests, lang, wake, rating, notify)
}

func (b *Bot) onProfileAction(cq *tgbotapi.CallbackQuery, action string) {
	switch action {
	case "alias":
		b.storage.WithUser(cq.From.ID, func(x *User) { x.State = StateNeedAlias })
		b.answerCallback(cq.ID, "")
		b.sendText(cq.Message.Chat.ID,
			"📛 Send me your alias (max 32 chars), or /skip:", nil, false)

	case "bio":
		b.storage.WithUser(cq.From.ID, func(x *User) { x.State = StateNeedBio })
		b.answerCallback(cq.ID, "")
		b.sendText(cq.Message.Chat.ID,
			"📝 Send me a short bio (max 200 chars), or /skip:", nil, false)

	case "interests":
		u, _ := b.storage.Get(cq.From.ID)
		b.storage.WithUser(cq.From.ID, func(x *User) { x.State = StateNeedInterests })
		b.answerCallback(cq.ID, "")
		page := 0
		if u != nil {
			page = len(u.Interests) / 12
		}
		b.sendText(cq.Message.Chat.ID,
			"🏷️ Tap tags to toggle, then press *Done*:",
			interestsKeyboard(interestSet(u.Interests), page), true)

	case "photo":
		b.storage.WithUser(cq.From.ID, func(x *User) { x.State = StateNeedPhoto })
		b.answerCallback(cq.ID, "")
		b.sendText(cq.Message.Chat.ID,
			"🖼️ Send me a profile photo, or /skip:", nil, false)

	default:
		b.answerCallback(cq.ID, "Unknown profile action")
	}
}

func (b *Bot) onInterestAction(cq *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) < 2 {
		b.answerCallback(cq.ID, "Bad request")
		return
	}
	u, ok := b.storage.Get(cq.From.ID)
	if !ok {
		b.answerCallback(cq.ID, "Please /start first")
		return
	}

	selected := interestSet(u.Interests)
	switch parts[1] {
	case "toggle":
		if len(parts) < 3 {
			b.answerCallback(cq.ID, "Bad tag")
			return
		}
		tag := parts[2]
		if !validInterest(tag) {
			b.answerCallback(cq.ID, "Bad tag")
			return
		}
		if selected[tag] {
			delete(selected, tag)
		} else {
			selected[tag] = true
		}
		var newList []string
		// Preserve stable order from AllInterests.
		for _, t := range AllInterests {
			if selected[t] {
				newList = append(newList, t)
			}
		}
		b.storage.WithUser(u.ID, func(x *User) { x.Interests = newList })
		b.answerCallback(cq.ID, "")
		// Refresh the message.
		edit := tgbotapi.NewEditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID,
			interestsKeyboard(selected, 0))
		b.send(edit)

	case "page":
		page, err := strconv.Atoi(parts[2])
		if err != nil {
			b.answerCallback(cq.ID, "Bad page")
			return
		}
		b.answerCallback(cq.ID, "")
		edit := tgbotapi.NewEditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID,
			interestsKeyboard(selected, page))
		b.send(edit)

	case "done":
		b.storage.WithUser(u.ID, func(x *User) { x.State = StateIdle })
		b.answerCallback(cq.ID, "Saved!")
		b.sendText(cq.Message.Chat.ID, "✅ Interests saved.", mainMenu, false)

	case "clear":
		b.storage.WithUser(u.ID, func(x *User) { x.Interests = nil })
		b.answerCallback(cq.ID, "Cleared")
		edit := tgbotapi.NewEditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID,
			interestsKeyboard(map[string]bool{}, 0))
		b.send(edit)

	default:
		b.answerCallback(cq.ID, "Unknown interest action")
	}
}

// handleStatefulInput handles StateNeedPhoto.
func (b *Bot) handlePhotoInput(msg *tgbotapi.Message, u *User) {
	if len(msg.Photo) == 0 {
		b.sendText(msg.Chat.ID, "Please send a photo, or /skip.", nil, false)
		return
	}
	best := msg.Photo[len(msg.Photo)-1]
	b.storage.WithUser(u.ID, func(x *User) {
		x.PhotoID = best.FileID
		x.State = StateIdle
	})
	b.sendText(msg.Chat.ID, "✅ Profile photo updated.", mainMenu, false)
}

// ---------------------------------------------------------------------------
// Language
// ---------------------------------------------------------------------------

func (b *Bot) onLanguagePick(cq *tgbotapi.CallbackQuery, code string) {
	if _, ok := SupportedLanguages[code]; !ok {
		b.answerCallback(cq.ID, "Bad language")
		return
	}
	b.storage.WithUser(cq.From.ID, func(x *User) {
		x.Language = code
		x.State = StateIdle
	})
	b.answerCallback(cq.ID, "Saved!")
	name := SupportedLanguages[code]
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		fmt.Sprintf("🌐 Language set to *%s*.", name))
	edit.ParseMode = "Markdown"
	b.send(edit)
	b.sendText(cq.Message.Chat.ID,
		"From now on, text chats with someone speaking a different language will be auto-translated.",
		mainMenu, false)
}

// ---------------------------------------------------------------------------
// Wake hours
// ---------------------------------------------------------------------------

func (b *Bot) startWakeHoursFlow(chatID int64, u *User) {
	if u.Timezone == "" {
		b.storage.WithUser(u.ID, func(x *User) { x.State = StateNeedTimezone })
		b.sendText(chatID,
			"🌍 First, pick your timezone:", timezoneKeyboard(), false)
		return
	}
	b.storage.WithUser(u.ID, func(x *User) { x.State = StateNeedWakeFrom })
	b.sendText(chatID,
		fmt.Sprintf("🌍 Timezone: *%s*\n\n"+
			"Send the hour you usually wake up (0-23), or /skip to clear wake hours:",
			u.Timezone), nil, true)
}

func (b *Bot) onTimezonePick(cq *tgbotapi.CallbackQuery, tz string) {
	if _, err := time.LoadLocation(tz); err != nil {
		b.answerCallback(cq.ID, "Bad timezone")
		return
	}
	b.storage.WithUser(cq.From.ID, func(x *User) {
		x.Timezone = tz
		x.State = StateNeedWakeFrom
	})
	b.answerCallback(cq.ID, "Saved")
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		fmt.Sprintf("🌍 Timezone: *%s*", tz))
	edit.ParseMode = "Markdown"
	b.send(edit)
	b.sendText(cq.Message.Chat.ID,
		"Now send the hour you usually wake up (0-23):", nil, false)
}

func (b *Bot) onWakePick(cq *tgbotapi.CallbackQuery, scope, val string) {
	h, err := strconv.Atoi(val)
	if err != nil || h < 0 || h > 23 {
		b.answerCallback(cq.ID, "Bad hour")
		return
	}
	switch scope {
	case "from":
		b.storage.WithUser(cq.From.ID, func(x *User) {
			x.WakeFrom = h
			x.State = StateNeedWakeTo
		})
		b.answerCallback(cq.ID, fmt.Sprintf("Wake at %d", h))
		b.sendText(cq.Message.Chat.ID,
			fmt.Sprintf("Wake from *%02d:00*. Now send the hour you go to sleep (0-23):", h),
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
			fmt.Sprintf("✅ Wake hours: *%02d:00 - %02d:00*", wakeFrom, h),
			mainMenu, true)
	default:
		b.answerCallback(cq.ID, "Unknown")
	}
}

// ---------------------------------------------------------------------------
// Notifications
// ---------------------------------------------------------------------------

func (b *Bot) toggleNotifications(u *User, chatID int64) {
	if u.Latitude == 0 && u.Longitude == 0 {
		b.sendText(chatID, "Share your location first.", requestLocationKeyboard, false)
		return
	}
	if u.NotifyWhenNearby {
		b.storage.WithUser(u.ID, func(x *User) {
			x.NotifyWhenNearby = false
		})
		b.sendText(chatID, "🔕 Notifications disabled.", mainMenu, false)
		return
	}
	b.storage.WithUser(u.ID, func(x *User) {
		x.State = StateNeedNotifyRadius
	})
	b.sendText(chatID,
		"🔔 *Get notified when someone new joins nearby.*\nPick a radius:",
		notifyRadiusKeyboard(), true)
}

func (b *Bot) onNotifyRadiusPick(cq *tgbotapi.CallbackQuery, raw string) {
	r, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		b.answerCallback(cq.ID, "Bad radius")
		return
	}
	b.storage.WithUser(cq.From.ID, func(x *User) {
		x.NotifyWhenNearby = true
		x.NotifyRadius = r
		x.State = StateIdle
		x.LastNotifiedAt = time.Time{}
	})
	b.answerCallback(cq.ID, fmt.Sprintf("%.0f km", r))
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		fmt.Sprintf("🔔 You'll be notified when someone joins within %.0f km.", r))
	b.send(edit)
	b.sendText(cq.Message.Chat.ID,
		"You can turn this off any time from the menu.", mainMenu, false)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func shortName(u *User) string {
	if u == nil {
		return "Stranger"
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
	return "Stranger"
}

func prettyGender(g string) string {
	switch g {
	case GenderMale:
		return "Male"
	case GenderFemale:
		return "Female"
	case GenderAny:
		return "Both / Any"
	default:
		return "Unknown"
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

func formatMatchLabel(c *User, dist float64) string {
	name := shortName(c)
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
	for _, t := range s {
		out[t] = true
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
