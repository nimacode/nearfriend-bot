package bot

import (
	"math"
	"sync"
	"time"
)

// State represents where the user is in the conversation flow.
type State int

const (
	StateIdle             State = iota // registered, waiting for a menu choice
	StateNeedLocation                  // just /start'd, waiting for location share
	StateNeedGender                    // location received, waiting for self gender
	StateNeedSearchGender              // in "find nearby" flow, waiting for who they want
	StateNeedRadius                    // in "find nearby" flow, waiting for distance
	StateBrowsing                      // looking at the inline list of nearby users
	StateInChat                        // paired with another user, relaying messages
)

// Gender values used for self-declared gender and "looking for" filter.
const (
	GenderMale   = "male"
	GenderFemale = "female"
	GenderAny    = "both"
)

// User holds everything we know about a Telegram user.
type User struct {
	ID         int64     // Telegram user ID (primary key in storage)
	FirstName  string    // Telegram first name
	Username   string    // Telegram username (may be empty)
	Latitude   float64   // last shared latitude
	Longitude  float64   // last shared longitude
	LocationAt time.Time // when the location was last shared

	Gender     string // self-declared: GenderMale / GenderFemale / other
	LookingFor string // gender filter: GenderMale / GenderFemale / GenderAny

	State State // current conversation state

	// Active chat partner (0 if not chatting).
	PartnerID  int64
	SearchFrom string // context: which search brought them to browsing
}

// Storage is a tiny in-memory key/value store. Thread-safe.
// For production swap with SQLite/Postgres/Redis — the interface is tiny.
type Storage struct {
	mu    sync.RWMutex
	users map[int64]*User
}

func NewStorage() *Storage {
	return &Storage{users: make(map[int64]*User)}
}

// Get returns a copy or the live pointer. We return the pointer so handlers
// can mutate fields, but Storage guards all access with its mutex.
func (s *Storage) Get(id int64) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
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
}

// All returns a snapshot of every registered user (used for matching).
func (s *Storage) All() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, u)
	}
	return out
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

// Nearby filters All() to users within radiusKm of (lat, lon).
// Self is excluded. Only users with a known gender AND known LookingFor
// (i.e. fully registered) are eligible.
func (s *Storage) Nearby(selfID int64, lat, lon, radiusKm float64) []*User {
	all := s.All()
	out := make([]*User, 0, len(all))
	for _, u := range all {
		if u.ID == selfID {
			continue
		}
		if u.Gender == "" || u.LookingFor == "" {
			continue
		}
		if u.State == StateInChat {
			continue // busy
		}
		if DistanceKm(lat, lon, u.Latitude, u.Longitude) <= radiusKm {
			out = append(out, u)
		}
	}
	return out
}
