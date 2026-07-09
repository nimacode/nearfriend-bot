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
			// Process each pair only once: when the first side is
			// collected, mark its partner as already handled so we
			// don't send the "time's up" prompt twice.
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
//
// Each lock acquisition is done as a separate, non-nested call.
// Calling Get/Upsert/WithUser from inside another WithUser callback
// would self-deadlock because sync.RWMutex is not re-entrant.
func (b *Bot) endCoffeeChat(userID int64) {
	var partnerID int64

	b.storage.WithUser(userID, func(u *User) {
		// Always clear our own timer.
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

	// Clear the partner's timer in its own critical section.
	b.storage.WithUser(partnerID, func(p *User) {
		p.ChatEndsAt = time.Time{}
		p.IsCoffeeChat = false
	})

	b.sendText(userID,
		"⏰ *Coffee time's up!*\nWould you like to keep chatting?",
		continueKeyboard(partnerID, true), true)
	b.sendText(partnerID,
		"⏰ *Coffee time's up!*\nWould you like to keep chatting?",
		continueKeyboard(userID, false), true)
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
	}
	var pending []pendingNotification

	// Range holds only the read lock, so we must not mutate users here.
	// Collect the notifications first, then update state afterwards.
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

		pending = append(pending, pendingNotification{
			id:   u.ID,
			name: shortName(best),
			dist: bestDist,
		})
	})

	for _, p := range pending {
		b.sendText(p.id,
			"🔔 *Someone new is nearby!*\n"+
				"👤 "+p.name+" · "+
				formatDistance(p.dist)+" away\n\n"+
				"Tap *Find nearby friends* to connect!",
			nil, true)
		// Stamp the cooldown under the write lock.
		b.storage.WithUser(p.id, func(x *User) {
			x.LastNotifiedAt = now
		})
	}
	if len(pending) > 0 {
		log.Printf("[notify] pinged %d users", len(pending))
	}
}
