package bot

import (
	"log"
	"time"
)

// startWorkers spins up the background loops the bot needs:
//   - chatTimerLoop: ends coffee chats when their timer expires
//   - notifyLoop: pings users who opted in to "notify me when someone
//     nearby joins"
//
// The loops stop when the bot's stop channel is closed.
func (b *Bot) startWorkers() {
	go b.chatTimerLoop()
	go b.notifyLoop()
}

// chatTimerLoop wakes up every 15 seconds and ends any coffee chat whose
// timer has expired.
func (b *Bot) chatTimerLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		expired := []int64{}
		b.storage.Range(func(u *User) {
			if !u.ChatEndsAt.IsZero() && now.After(u.ChatEndsAt) {
				expired = append(expired, u.ID)
			}
		})
		for _, uid := range expired {
			b.endCoffeeChat(uid)
		}
	}
}

// endCoffeeChat handles the timer-expiry branch: notify both sides and
// return them to idle (but preserve the chat if they both agree to
// continue — handled by handleCommand "continuechat" callback).
func (b *Bot) endCoffeeChat(userID int64) {
	b.storage.WithUser(userID, func(u *User) {
		if u.PartnerID == 0 || u.State != StateInChat {
			u.ChatEndsAt = time.Time{}
			u.IsCoffeeChat = false
			return
		}

		// Reset timer but keep both in chat; offer a "continue?" prompt.
		partnerID := u.PartnerID
		u.ChatEndsAt = time.Time{}
		u.IsCoffeeChat = false

		b.sendText(u.ID,
			"⏰ *Coffee time's up!*\nWould you like to keep chatting?",
			continueKeyboard(partnerID, true), true)

		if p, ok := b.storage.Get(partnerID); ok {
			p.ChatEndsAt = time.Time{}
			p.IsCoffeeChat = false
			b.storage.Upsert(p)
			b.sendText(partnerID,
				"⏰ *Coffee time's up!*\nWould you like to keep chatting?",
				continueKeyboard(u.ID, false), true)
		}
	})
}

// notifyLoop runs every 5 minutes and pings users who opted in to
// "notify me when a new person joins nearby".
func (b *Bot) notifyLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		b.checkAndNotify()
	}
}

func (b *Bot) checkAndNotify() {
	now := time.Now()
	notified := 0
	b.storage.Range(func(u *User) {
		if !u.NotifyWhenNearby {
			return
		}
		if now.Before(u.SuspendedUntil) {
			return
		}
		if u.Latitude == 0 && u.Longitude == 0 {
			return
		}
		if now.Sub(u.LastNotifiedAt) < 30*time.Minute {
			return
		}
		// Skip users who are mid-flow.
		switch u.State {
		case StateInChat, StateBrowsing, StateNeedRadius,
			StateNeedSearchGender, StateNeedLocation, StateNeedGender:
			return
		}

		radius := u.NotifyRadius
		if radius <= 0 {
			radius = 50
		}
		candidates := b.filterMatches(u, radius)
		if len(candidates) == 0 {
			return
		}

		// Pick the closest one.
		best := candidates[0]
		bestDist := DistanceKm(u.Latitude, u.Longitude, best.Latitude, best.Longitude)
		for _, c := range candidates[1:] {
			d := DistanceKm(u.Latitude, u.Longitude, c.Latitude, c.Longitude)
			if d < bestDist {
				best, bestDist = c, d
			}
		}

		b.sendText(u.ID,
			"🔔 *Someone new is nearby!*\n"+
				"👤 "+shortName(best)+" · "+
				formatDistance(bestDist)+" away\n\n"+
				"Tap *Find nearby friends* to connect!",
			nil, true)
		u.LastNotifiedAt = now
		notified++
	})
	if notified > 0 {
		log.Printf("[notify] pinged %d users", notified)
	}
}
