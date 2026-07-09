package bot

import (
	"fmt"
	"time"
)

// checkAchievements looks at the user's stats and grants any new
// achievements. Returns the list of achievements unlocked by this call
// so the caller can send notifications.
//
// completedCoffee reports whether the chat that just ended (or the
// event being evaluated) was a completed coffee chat — used for the
// "Coffee Lover" badge, since u.IsCoffeeChat is cleared by the caller
// before we run.
func checkAchievements(u *User, partner *User, now time.Time, completedCoffee bool) []Achievement {
	if u.Achievements == nil {
		u.Achievements = make(map[string]bool)
	}

	// Record this conversation's data FIRST so the current partner's
	// city/language counts toward the thresholds checked below (otherwise
	// Globetrotter/Explorer/Polyglot would unlock one event too late).
	if partner != nil {
		if u.CitiesChat == nil {
			u.CitiesChat = make(map[string]int64)
		}
		key := partner.CityKey()
		if key != "" {
			u.CitiesChat[key] = partner.ID
		}
		if partner.Language != "" {
			if u.LangPartners == nil {
				u.LangPartners = make(map[string]bool)
			}
			u.LangPartners[partner.Language] = true
		}
	}

	var newlyUnlocked []Achievement

	grant := func(id string) {
		if u.Achievements[id] {
			return
		}
		for _, a := range AllAchievements {
			if a.ID == id {
				u.Achievements[id] = true
				newlyUnlocked = append(newlyUnlocked, a)
				return
			}
		}
	}

	hour := localHour(u, now)

	// First chat
	if u.ChatCount >= 1 {
		grant(AchievementFirstChat)
	}

	// Chatterbox: 10 chats
	if u.ChatCount >= 10 {
		grant(AchievementChatterbox)
	}

	// Globetrotter: 3+ distinct cities
	if len(u.CitiesChat) >= 3 {
		grant(AchievementMultiCity)
	}

	// Explorer: 5+ distinct cities
	if len(u.CitiesChat) >= 5 {
		grant(AchievementExplorer)
	}

	// Night owl: chatted between 0 and 5 (local time)
	if hour >= 0 && hour < 5 {
		grant(AchievementNightOwl)
	}

	// Early bird: chatted before 8 (local time)
	if hour < 8 {
		grant(AchievementEarlyBird)
	}

	// Five stars: every rating received is a 5/5
	if u.RatingCount > 0 && u.RatingSum == 5*u.RatingCount {
		grant(AchievementFiveStar)
	}

	// Well liked: average >= 4.5 with 5+ reviews
	if u.RatingCount >= 5 && u.AverageRating() >= 4.5 {
		grant(AchievementWellLiked)
	}

	// Polyglot: chatted with 3+ distinct languages
	if len(u.LangPartners) >= 3 {
		grant(AchievementPolyglot)
	}

	// Coffee lover: completed a coffee chat
	if completedCoffee {
		grant(AchievementCoffeeLover)
	}

	return newlyUnlocked
}

// formatAchievementLine returns a human-readable line for a single
// achievement: "🥇 First Chat — You had your first conversation!".
func formatAchievementLine(a Achievement, unlocked bool) string {
	prefix := "🔒"
	if unlocked {
		prefix = a.Emoji
	}
	return fmt.Sprintf("%s %s — %s", prefix, a.Title, a.Description)
}

// achievementsText returns the formatted "🏆 Achievements" message.
func achievementsText(u *User) string {
	out := "🏆 *Your achievements*\n\n"
	unlocked, locked := 0, 0
	for _, a := range AllAchievements {
		if u.Achievements[a.ID] {
			out += "✅ " + formatAchievementLine(a, true) + "\n"
			unlocked++
		} else {
			out += "🔒 " + a.Title + "\n"
			locked++
		}
	}
	out += fmt.Sprintf("\n_%d unlocked · %d to go_", unlocked, locked)
	return out
}

// localHour returns the current hour in the user's local timezone (or
// the server's local time if the user hasn't set a timezone).
func localHour(u *User, now time.Time) int {
	if u != nil && u.Timezone != "" {
		if loc, err := time.LoadLocation(u.Timezone); err == nil {
			return now.In(loc).Hour()
		}
	}
	return now.Hour()
}
