package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// mainMenu is shown to any registered user as a persistent keyboard.
var mainMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("📍 Update my location"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🚻 Set my gender"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🔍 Find nearby friends"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("❌ End chat"),
	),
)

// chatKeyboard is a stripped menu shown while in a chat.
var chatKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("❌ End chat"),
	),
)

// requestLocationKeyboard asks the user to share their Telegram location.
var requestLocationKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.KeyboardButton{Text: "📎 Share my location", RequestLocation: true},
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

// sendText is a small helper that fills ReplyMarkup, ParseMode and sends.
// Centralising it avoids the chained-call boilerplate from tgbotapi.
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

// ---------------------------------------------------------------------------
// Update dispatch
// ---------------------------------------------------------------------------

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	// Commands always take priority.
	if msg.IsCommand() {
		b.handleCommand(msg)
		return
	}

	// A location share works regardless of state (always overwrites).
	if msg.Location != nil {
		b.onLocation(msg)
		return
	}

	u, ok := b.storage.Get(msg.From.ID)
	if !ok {
		b.sendText(msg.Chat.ID, "Send /start to begin.", nil, false)
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

	case StateInChat:
		b.relayToPartner(msg)

	case StateIdle:
		switch strings.TrimSpace(msg.Text) {
		case "📍 Update my location":
			b.promptLocation(msg.Chat.ID)
		case "🚻 Set my gender":
			b.sendText(msg.Chat.ID, "What's your gender?", genderKeyboard("gender:self"), false)
		case "🔍 Find nearby friends":
			b.startSearch(msg.From.ID, msg.Chat.ID)
		case "❌ End chat":
			b.endChat(msg.From.ID, msg.Chat.ID, true)
		default:
			b.sendText(msg.Chat.ID, "Use the menu below 👇", mainMenu, false)
		}

	default:
		b.sendText(msg.Chat.ID, "Use the menu below 👇", mainMenu, false)
	}
}

// handleCommand processes /commands.
func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		b.cmdStart(msg)
	case "end", "stop", "cancel":
		b.endChat(msg.From.ID, msg.Chat.ID, true)
	case "menu":
		b.sendText(msg.Chat.ID, "Main menu 👇", mainMenu, false)
	default:
		b.sendText(msg.Chat.ID, "Unknown command. Try /start or /menu.", nil, false)
	}
}

// handleCallback routes inline-keyboard taps.
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

	default:
		b.answerCallback(cq.ID, "Unknown action")
	}
}

// ---------------------------------------------------------------------------
// /start and registration flow
// ---------------------------------------------------------------------------

func (b *Bot) cmdStart(msg *tgbotapi.Message) {
	uid := msg.From.ID

	u := &User{
		ID:        uid,
		FirstName: msg.From.FirstName,
		Username:  msg.From.UserName,
		State:     StateNeedLocation,
	}
	b.storage.Upsert(u)

	hello := fmt.Sprintf(
		"Hey %s! 👋\n\n"+
			"I'm *NearFriend* — I help you meet people nearby for a chat.\n\n"+
			"1️⃣ Share your location with the button below.\n"+
			"2️⃣ Tell me your gender.\n"+
			"3️⃣ Tap *Find nearby friends* whenever you want to match.",
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
	b.sendText(msg.Chat.ID, "📍 Location updated!", mainMenu, false)
}

func (b *Bot) promptLocation(chatID int64) {
	b.sendText(chatID, "Tap the button to share your current location:",
		requestLocationKeyboard, false)
}

// ---------------------------------------------------------------------------
// Gender buttons (self & search)
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
		u.Gender = value
		b.storage.Upsert(u)
		b.answerCallback(cq.ID, "Saved!")

		edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
			fmt.Sprintf("✅ Your gender is set to *%s*.", prettyGender(value)))
		edit.ParseMode = "Markdown"
		b.send(edit)

		b.sendText(cq.Message.Chat.ID,
			"You're all set! Tap *Find nearby friends* when you're ready.",
			mainMenu, true)

	case "search":
		u.LookingFor = value
		u.State = StateNeedRadius
		b.storage.Upsert(u)
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
// Radius selection -> build candidate list
// ---------------------------------------------------------------------------

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
	u.State = StateBrowsing
	b.storage.Upsert(u)
	b.answerCallback(cq.ID, fmt.Sprintf("%.0f km", radius))

	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		fmt.Sprintf("Searching within %.0f km…", radius))
	b.send(edit)

	if len(candidates) == 0 {
		b.sendText(cq.Message.Chat.ID,
			"😕 No nearby matches. Try a bigger radius, or update your preferences.",
			mainMenu, false)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, c := range candidates {
		dist := DistanceKm(u.Latitude, u.Longitude, c.Latitude, c.Longitude)
		label := fmt.Sprintf("%s · %.1f km · %s",
			shortName(c), dist, genderEmoji(c.Gender))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label,
				"nearby:"+strconv.FormatInt(c.ID, 10)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("↩️ Cancel", "endchat"),
	))

	b.sendText(cq.Message.Chat.ID,
		fmt.Sprintf("Found *%d* nearby %s. Pick one 👇",
			len(candidates), prettyGender(u.LookingFor)),
		tgbotapi.NewInlineKeyboardMarkup(rows...), true)
}

// filterMatches returns nearby users that:
//   - are within radiusKm
//   - are not the requester
//   - have a gender the requester is looking for (u.LookingFor)
//   - are themselves looking for the requester's gender (mutual match)
//   - are not already in a chat
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
// Picking a nearby user -> connect both
// ---------------------------------------------------------------------------

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

	u.PartnerID = targetID
	u.State = StateInChat
	target.PartnerID = u.ID
	target.State = StateInChat
	b.storage.Upsert(u)
	b.storage.Upsert(target)

	b.answerCallback(cq.ID, "Connected!")

	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID,
		fmt.Sprintf("✅ You're now chatting with %s. Say hi!", shortName(target)))
	b.send(edit)

	b.sendText(targetID,
		fmt.Sprintf("💬 Someone nearby wants to chat with you. Say hi back to %s!",
			shortName(u)),
		chatKeyboard, false)
}

// ---------------------------------------------------------------------------
// Chat relay & disconnect
// ---------------------------------------------------------------------------

func (b *Bot) relayToPartner(msg *tgbotapi.Message) {
	u, ok := b.storage.Get(msg.From.ID)
	if !ok || u.PartnerID == 0 {
		return
	}
	partner, ok := b.storage.Get(u.PartnerID)
	if !ok {
		b.sendText(msg.Chat.ID, "Your partner left. Chat ended.", mainMenu, false)
		u.State = StateIdle
		u.PartnerID = 0
		b.storage.Upsert(u)
		return
	}

	// CopyMessage forwards text, photos, voice, stickers, etc. — anything.
	copyMsg := tgbotapi.NewCopyMessage(partner.ID, msg.Chat.ID, msg.MessageID)
	if _, err := b.api.Send(copyMsg); err != nil {
		log.Printf("[relay] copy failed: %v", err)
		b.sendText(msg.Chat.ID, "⚠️ Could not deliver your message.", nil, false)
	}
}

func (b *Bot) onEndChatButton(cq *tgbotapi.CallbackQuery) {
	b.endChat(cq.From.ID, cq.Message.Chat.ID, false)
	b.answerCallback(cq.ID, "Chat ended")
}

func (b *Bot) endChat(userID, chatID int64, notifyPartner bool) {
	u, ok := b.storage.Get(userID)
	if !ok {
		return
	}
	if u.State != StateInChat || u.PartnerID == 0 {
		b.sendText(chatID, "You're not in a chat right now.", mainMenu, false)
		return
	}

	partnerID := u.PartnerID
	if p, ok := b.storage.Get(partnerID); ok {
		p.PartnerID = 0
		p.State = StateIdle
		b.storage.Upsert(p)
		if notifyPartner {
			b.sendText(partnerID,
				"💔 Your chat partner has left the conversation.",
				mainMenu, false)
		}
	}

	u.PartnerID = 0
	u.State = StateIdle
	b.storage.Upsert(u)

	b.sendText(chatID,
		"Chat ended. Tap *Find nearby friends* to start a new one.",
		mainMenu, true)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (b *Bot) startSearch(userID int64, chatID int64) {
	u, ok := b.storage.Get(userID)
	if !ok {
		return
	}
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

	u.State = StateNeedSearchGender
	b.storage.Upsert(u)

	b.sendText(chatID, "Who do you want to chat with?",
		genderKeyboard("gender:search"), false)
}

func shortName(u *User) string {
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
