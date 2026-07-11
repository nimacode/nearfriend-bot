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
		var expired []int64
		seen := make(map[int64]bool)
		b.storage.Range(func(u *User) {
			if u.ChatEndsAt.IsZero() || !now.After(u.ChatEndsAt) {
				return
			}
			if seen[u.ID] {
				return
			}
			seen[u.ID] = true
			if u.PartnerID != 0 {
				seen[u.PartnerID] = true
			}
			expired = append(expired, u.ID)
		})
		for _, uid := range expired {
			b.endCoffeeChat(uid)
		}
	}
}

// endCoffeeChat handles the timer-expiry branch: notify both sides and
// offer them a "continue?" prompt while keeping them in the chat.
func (b *Bot) endCoffeeChat(userID int64) {
	var partnerID int64

	b.storage.WithUser(userID, func(u *User) {
		u.ChatEndsAt = time.Time{}
		u.IsCoffeeChat = false
		if u.PartnerID == 0 || u.State != StateInChat {
			return
		}
		partnerID = u.PartnerID
	})
	if partnerID == 0 {
		return
	}

	b.storage.WithUser(partnerID, func(p *User) {
		p.ChatEndsAt = time.Time{}
		p.IsCoffeeChat = false
	})

	userLang := b.partnerLang(userID)
	partnerLang := b.partnerLang(partnerID)

	b.sendText(userID,
		t(userLang, "msg_coffee_time_up"),
		continueKeyboard(partnerID, userLang), true)
	b.sendText(partnerID,
		t(partnerLang, "msg_coffee_time_up"),
		continueKeyboard(userID, partnerLang), true)
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
	type pendingNotification struct {
		id   int64
		name string
		dist float64
		lang string
	}
	var pending []pendingNotification

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
		switch u.State {
		case StateInChat, StateBrowsing, StateNeedRadius,
			StateNeedSearchGender, StateNeedLocation, StateNeedGender,
			StateNeedUILang:
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

		best := candidates[0]
		bestDist := DistanceKm(u.Latitude, u.Longitude, best.Latitude, best.Longitude)
		for _, c := range candidates[1:] {
			d := DistanceKm(u.Latitude, u.Longitude, c.Latitude, c.Longitude)
			if d < bestDist {
				best, bestDist = c, d
			}
		}

		pending = append(pending, pendingNotification{
			id:   u.ID,
			name: shortName(best, u.Lang()),
			dist: bestDist,
			lang: u.Lang(),
		})
	})

	for _, p := range pending {
		b.sendText(p.id,
			tf(p.lang, "fmt_notify_someone_nearby", p.name, formatDistance(p.dist)),
			nil, true)
		b.storage.WithUser(p.id, func(x *User) {
			x.LastNotifiedAt = now
		})
	}
	if len(pending) > 0 {
		log.Printf("[notify] pinged %d users", len(pending))
	}
}
