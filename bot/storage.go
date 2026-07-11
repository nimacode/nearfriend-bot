package bot

import (
	"math"
	"sort"
	"strconv"
	"sync"
	"time"
)

// State represents where the user is in the conversation flow.
type State int

const (
	StateIdle             State = iota // registered, waiting for a menu choice
	StateNeedUILang                    // just /start'd, waiting for UI language pick
	StateNeedLocation                  // language chosen, waiting for location share
	StateNeedGender                    // location received, waiting for self gender
	StateNeedSearchGender              // in "find nearby" flow, waiting for who they want
	StateNeedRadius                    // in "find nearby" flow, waiting for distance
	StateBrowsing                      // looking at the inline list of nearby users
	StateInChat                        // paired with another user, relaying messages

	// Profile setup (entered via menu)
	StateNeedAlias
	StateNeedBio
	StateNeedInterests
	StateNeedPhoto

	// Settings
	StateNeedLanguage
	StateNeedTimezone
	StateNeedWakeFrom
	StateNeedWakeTo

	// Notification subscription
	StateNeedNotifyRadius

	// Rating prompt
	StateRatePartner
)

// Gender values used for self-declared gender and "looking for" filter.
const (
	GenderMale   = "male"
	GenderFemale = "female"
	GenderAny    = "both"
)

// CoffeeChatDuration is how long a coffee chat lasts.
const CoffeeChatDuration = 15 * time.Minute

// RatingThreshold is the minimum average rating a user needs to be shown
// in match results.
const RatingThreshold = 3.5

// MinRatingsForFilter is the minimum number of ratings a user must have
// before the rating filter applies. Below this we don't filter them out
// (every new user deserves a chance).
const MinRatingsForFilter = 3

// ReportThreshold is the number of reports before a user is suspended.
const ReportThreshold = 3

// SuspendDuration is how long a user is suspended after ReportThreshold.
const SuspendDuration = 24 * time.Hour

// Languages supported for chat translation. Map code -> display name.
var SupportedLanguages = map[string]string{
	"en": "English",
	"fa": "Persian (فارسی)",
	"tr": "Turkish (Türkçe)",
	"de": "German (Deutsch)",
	"fr": "French (Français)",
	"es": "Spanish (Español)",
	"ar": "Arabic (العربية)",
	"ru": "Russian (Русский)",
	"zh": "Chinese (中文)",
	"ja": "Japanese (日本語)",
	"ko": "Korean (한국어)",
	"it": "Italian (Italiano)",
	"pt": "Portuguese (Português)",
	"hi": "Hindi (हिन्दी)",
	"nl": "Dutch (Nederlands)",
}

// Interest tags users can pick. Curated to keep the UI compact.
var AllInterests = []string{
	"books", "movies", "music", "travel", "food", "sports", "gaming",
	"code", "art", "photography", "fitness", "nature", "history",
	"science", "philosophy", "humor", "fashion", "tech", "anime",
	"coffee", "cooking", "hiking",
}

// Common timezones shown to the user.
var CommonTimezones = []string{
	"Asia/Tehran",
	"Asia/Istanbul",
	"Asia/Dubai",
	"Asia/Kolkata",
	"Europe/London",
	"Europe/Berlin",
	"Europe/Moscow",
	"America/New_York",
	"America/Los_Angeles",
	"Asia/Tokyo",
	"Asia/Shanghai",
	"Australia/Sydney",
	"UTC",
}

// Achievement IDs.
const (
	AchievementFirstChat   = "first_chat"
	AchievementMultiCity   = "multi_city"
	AchievementNightOwl    = "night_owl"
	AchievementFiveStar    = "five_star"
	AchievementChatterbox  = "chatterbox"
	AchievementPolyglot    = "polyglot"
	AchievementEarlyBird   = "early_bird"
	AchievementExplorer    = "explorer"
	AchievementWellLiked   = "well_liked"
	AchievementCoffeeLover = "coffee_lover"
)

// Achievement is a single unlockable badge.
type Achievement struct {
	ID          string
	Title       string
	Description string
	Emoji       string
}

// AllAchievements is the catalog shown in /achievements.
var AllAchievements = []Achievement{
	{AchievementFirstChat, "First Chat", "You had your first conversation!", "🥇"},
	{AchievementMultiCity, "Globetrotter", "Chatted with people from 3+ cities", "🌍"},
	{AchievementNightOwl, "Night Owl", "Chatted between midnight and 5 AM", "🌙"},
	{AchievementFiveStar, "Five Stars", "Received a perfect 5/5 rating", "⭐"},
	{AchievementChatterbox, "Chatterbox", "Had 10 conversations", "💬"},
	{AchievementPolyglot, "Polyglot", "Chatted with 3+ different languages", "🗣️"},
	{AchievementEarlyBird, "Early Bird", "Chatted before 8 AM", "🐦"},
	{AchievementExplorer, "Explorer", "Chatted in 5+ different cities", "🗺️"},
	{AchievementWellLiked, "Well Liked", "Average rating >= 4.5 with 5+ reviews", "❤️"},
	{AchievementCoffeeLover, "Coffee Lover", "Completed a coffee chat", "☕"},
}

// Icebreaker questions sent at the start of every chat.
var Icebreakers = []string{
	"If you could travel anywhere right now, where would you go? ✈️",
	"What's the last book you read? 📚",
	"Coffee or tea? ☕",
	"What's your favorite way to spend a weekend?",
	"If you could have dinner with anyone (alive or dead), who would it be? 🍽️",
	"What's a skill you'd love to learn? 🎯",
	"What's the best concert or festival you've been to? 🎵",
	"Cats or dogs? 🐱🐶",
	"What's something you're proud of from the past year? 🌟",
	"If you won the lottery tomorrow, what's the first thing you'd do? 💰",
	"Three things you can't live without?",
	"Most embarrassing song on your playlist? 🎧",
	"Morning person or night owl?",
	"What's your hidden talent?",
	"Best meal you've ever had? 🍕",
}

// User holds everything we know about a Telegram user.
type User struct {
	ID         int64
	FirstName  string
	Username   string
	Latitude   float64
	Longitude  float64
	LocationAt time.Time

	Gender     string
	LookingFor string

	State State

	PartnerID  int64
	SearchFrom string

	// Profile
	Alias     string
	Bio       string
	Interests []string
	PhotoID   string

	// UILang is the bot interface language (en, ru, fa).
	UILang string

	// Language for chat translation
	Language string

	// Wake hours in user's local timezone
	Timezone string
	WakeFrom int
	WakeTo   int

	// Chat timer (for coffee chat)
	ChatEndsAt   time.Time
	IsCoffeeChat bool

	// Notifications
	NotifyWhenNearby bool
	NotifyRadius     float64
	LastNotifiedAt   time.Time

	// Ratings (received from partners)
	RatingSum   int
	RatingCount int
	RatedBy     map[int64]bool

	// Safety
	SuspendedUntil time.Time
	Blocks         map[int64]bool

	// Achievements
	Achievements map[string]bool

	// Stats
	ChatCount    int
	CitiesChat   map[string]int64 // city key (rough: lat/lon bucket) -> partnerID
	LangPartners map[string]bool  // partner languages we've chatted with

	// Per-search state
	IsCoffeeSearch   bool  // user picked ☕ for this search
	PendingRatingFor int64 // non-zero: bot is waiting on a rating for this partner
}

// CityKey returns a coarse identifier for the user's location (used for
// the "chatted in N cities" achievement). We bucket to 0.5° (~50 km).
func (u *User) CityKey() string {
	if u.Latitude == 0 && u.Longitude == 0 {
		return ""
	}
	lat := math.Round(u.Latitude*2) / 2
	lon := math.Round(u.Longitude*2) / 2
	return formatFloat(lat) + "," + formatFloat(lon)
}

func formatFloat(f float64) string {
	if f == math.Trunc(f) {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// Storage is a tiny in-memory key/value store. Thread-safe.
// For production swap with SQLite/Postgres/Redis — the interface is tiny.
type Storage struct {
	mu      sync.RWMutex
	users   map[int64]*User
	reports map[int64]int // userID -> active report count
}

func NewStorage() *Storage {
	return &Storage{
		users:   make(map[int64]*User),
		reports: make(map[int64]int),
	}
}

// Get returns a live pointer. Storage guards all access with its mutex,
// but compound updates on the returned *User should go through WithUser
// (or wrap them in a critical section externally) to stay consistent.
func (s *Storage) Get(id int64) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

// WithUser runs fn while holding the write lock. Use this when you need
// to read + mutate a user atomically.
func (s *Storage) WithUser(id int64, fn func(*User)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return false
	}
	fn(u)
	return true
}

// Upsert inserts or updates a user record.
func (s *Storage) Upsert(u *User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[u.ID] = u
}

// Delete removes a user (used for /reset or testing).
func (s *Storage) Delete(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
	delete(s.reports, id)
}

// All returns a snapshot of every registered user.
func (s *Storage) All() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, u)
	}
	return out
}

// Range iterates over every user with the read lock held. The fn may not
// mutate the user pointer (use WithUser for that).
func (s *Storage) Range(fn func(*User)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		fn(u)
	}
}

// Count returns the number of registered users.
func (s *Storage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}

// IncrementReport increases the report count for userID and returns the
// new count.
func (s *Storage) IncrementReport(userID int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports[userID]++
	return s.reports[userID]
}

// ResetReports clears the report count for userID.
func (s *Storage) ResetReports(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.reports, userID)
}

// DistanceKm returns the great-circle distance between two lat/lon pairs
// in kilometers using the Haversine formula.
func DistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	toRad := func(deg float64) float64 { return deg * math.Pi / 180 }

	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

// Nearby returns users that:
//   - are within radiusKm of (lat, lon)
//   - are not the requester
//   - have a gender AND LookingFor set (fully registered)
//   - are not in a chat
//   - are not suspended
//   - are not blocked by the requester
//   - have a good standing average rating (or fewer than MinRatingsForFilter)
//
// Mutually-consent gender filtering is applied by the caller (it needs
// the requester's profile to compare both directions).
func (s *Storage) Nearby(selfID int64, lat, lon, radiusKm float64) []*User {
	all := s.All()
	out := make([]*User, 0, len(all))
	me, _ := s.Get(selfID)
	for _, u := range all {
		if u.ID == selfID {
			continue
		}
		if u.Gender == "" || u.LookingFor == "" {
			continue
		}
		if u.State == StateInChat {
			continue
		}
		if time.Now().Before(u.SuspendedUntil) {
			continue
		}
		if me != nil && me.Blocks[u.ID] {
			continue
		}
		if u.RatingCount >= MinRatingsForFilter {
			avg := float64(u.RatingSum) / float64(u.RatingCount)
			if avg < RatingThreshold {
				continue
			}
		}
		if DistanceKm(lat, lon, u.Latitude, u.Longitude) <= radiusKm {
			out = append(out, u)
		}
	}
	return out
}

// IsAwake returns true if the user is currently in their wake window
// (in their declared timezone). If the timezone is unknown or unset,
// we assume the user is awake.
func (u *User) IsAwake() bool {
	if u.Timezone == "" || u.WakeFrom == u.WakeTo {
		return true
	}
	loc, err := time.LoadLocation(u.Timezone)
	if err != nil {
		return true
	}
	hour := time.Now().In(loc).Hour()
	return hourInRange(hour, u.WakeFrom, u.WakeTo)
}

// HourInRange reports whether hour falls within [from, to). It supports
// wrap-around (e.g., 22..3 means 22, 23, 0, 1, 2).
func hourInRange(hour, from, to int) bool {
	if from == to {
		return true
	}
	if from < to {
		return hour >= from && hour < to
	}
	return hour >= from || hour < to
}

// AverageRating returns the user's average rating, or 0 if unrated.
func (u *User) AverageRating() float64 {
	if u.RatingCount == 0 {
		return 0
	}
	return float64(u.RatingSum) / float64(u.RatingCount)
}

// HasAchievement reports whether the user has unlocked the given ID.
func (u *User) HasAchievement(id string) bool {
	return u != nil && u.Achievements[id]
}

// Lang returns the user's UI language, defaulting to English.
func (u *User) Lang() string {
	if u == nil || u.UILang == "" {
		return UILangEn
	}
	return u.UILang
}

// SortedInterests returns the user's interests sorted alphabetically.
func (u *User) SortedInterests() []string {
	out := make([]string, len(u.Interests))
	copy(out, u.Interests)
	sort.Strings(out)
	return out
}
