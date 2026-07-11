package bot

import (
	"fmt"
	"time"
)

// checkAchievements looks at the user's stats and grants any new
// achievements. Returns the list of achievements unlocked by this call
// so the caller can send notifications.
func checkAchievements(u *User, partner *User, now time.Time, completedCoffee bool) []Achievement {
	if u.Achievements == nil {
		u.Achievements = make(map[string]bool)
	}

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

	if u.ChatCount >= 1 {
		grant(AchievementFirstChat)
	}
	if u.ChatCount >= 10 {
		grant(AchievementChatterbox)
	}
	if len(u.CitiesChat) >= 3 {
		grant(AchievementMultiCity)
	}
	if len(u.CitiesChat) >= 5 {
		grant(AchievementExplorer)
	}
	if hour >= 0 && hour < 5 {
		grant(AchievementNightOwl)
	}
	if hour < 8 {
		grant(AchievementEarlyBird)
	}
	if u.RatingCount > 0 && u.RatingSum == 5*u.RatingCount {
		grant(AchievementFiveStar)
	}
	if u.RatingCount >= 5 && u.AverageRating() >= 4.5 {
		grant(AchievementWellLiked)
	}
	if len(u.LangPartners) >= 3 {
		grant(AchievementPolyglot)
	}
	if completedCoffee {
		grant(AchievementCoffeeLover)
	}

	return newlyUnlocked
}

// achievementsText returns the formatted achievements message in the
// user's UI language.
func achievementsText(u *User) string {
	lang := u.Lang()
	out := t(lang, "msg_achievements_title")
	unlocked, locked := 0, 0
	for _, a := range AllAchievements {
		if u.Achievements[a.ID] {
			out += "✅ " + a.Emoji + " " + achTitle(lang, a.ID) + " — " + achDesc(lang, a.ID) + "\n"
			unlocked++
		} else {
			out += "🔒 " + achTitle(lang, a.ID) + "\n"
			locked++
		}
	}
	out += fmt.Sprintf("\n"+t(lang, "fmt_achievements_count"), unlocked, locked)
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
