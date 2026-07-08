package bot

import (
	"math"
	"testing"
	"time"
)

func TestDistanceKm(t *testing.T) {
	tests := []struct {
		name                   string
		lat1, lon1, lat2, lon2 float64
		want                   float64
		tolerance              float64
	}{
		{
			name: "same point is zero",
			lat1: 35.6892, lon1: 51.3890, // Tehran
			lat2: 35.6892, lon2: 51.3890,
			want: 0, tolerance: 0.001,
		},
		{
			name: "Tehran to Karaj ~42km",
			lat1: 35.6892, lon1: 51.3890,
			lat2: 35.8400, lon2: 50.9391,
			want: 42, tolerance: 5,
		},
		{
			name: "Tehran to Isfahan ~340km",
			lat1: 35.6892, lon1: 51.3890,
			lat2: 32.6539, lon2: 51.6660,
			want: 339, tolerance: 15,
		},
		{
			name: "antipodes ~half Earth circumference",
			lat1: 0, lon1: 0,
			lat2: 0, lon2: 180,
			want: 20015, tolerance: 50,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DistanceKm(tc.lat1, tc.lon1, tc.lat2, tc.lon2)
			if math.Abs(got-tc.want) > tc.tolerance {
				t.Errorf("DistanceKm = %.2f km, want %.2f ± %.2f", got, tc.want, tc.tolerance)
			}
		})
	}
}

func TestStorageUpsertAndGet(t *testing.T) {
	s := NewStorage()
	u := &User{ID: 42, FirstName: "Test", Latitude: 1, Longitude: 2, Gender: GenderMale}
	s.Upsert(u)

	got, ok := s.Get(42)
	if !ok {
		t.Fatal("expected to find user 42")
	}
	if got.FirstName != "Test" {
		t.Errorf("got FirstName=%q, want %q", got.FirstName, "Test")
	}

	s.Upsert(&User{ID: 42, FirstName: "Updated"})
	got, _ = s.Get(42)
	if got.FirstName != "Updated" {
		t.Errorf("after upsert FirstName=%q, want %q", got.FirstName, "Updated")
	}

	s.Delete(42)
	if _, ok := s.Get(42); ok {
		t.Error("expected user 42 to be deleted")
	}
}

func TestStorageWithUser(t *testing.T) {
	s := NewStorage()
	s.Upsert(&User{ID: 1, Gender: GenderMale, ChatCount: 0})

	ok := s.WithUser(1, func(u *User) {
		u.ChatCount = 5
		u.State = StateInChat
	})
	if !ok {
		t.Fatal("WithUser should return true for existing user")
	}

	got, _ := s.Get(1)
	if got.ChatCount != 5 {
		t.Errorf("got ChatCount=%d, want 5", got.ChatCount)
	}
	if got.State != StateInChat {
		t.Errorf("got State=%d, want %d", got.State, StateInChat)
	}

	if s.WithUser(999, func(*User) {}) {
		t.Error("WithUser should return false for missing user")
	}
}

func TestStorageNearby(t *testing.T) {
	s := NewStorage()

	me := &User{ID: 1, Latitude: 10, Longitude: 10, Gender: GenderMale, LookingFor: GenderAny}
	s.Upsert(me)

	near := &User{ID: 2, Latitude: 10.5, Longitude: 10.5, Gender: GenderFemale, LookingFor: GenderAny}
	s.Upsert(near)

	far := &User{ID: 3, Latitude: 20, Longitude: 20, Gender: GenderFemale, LookingFor: GenderAny}
	s.Upsert(far)

	incomplete := &User{ID: 4, Latitude: 10.1, Longitude: 10.1}
	s.Upsert(incomplete)

	got := s.Nearby(1, 10, 10, 100)
	if len(got) != 1 {
		t.Fatalf("expected 1 nearby user, got %d", len(got))
	}
	if got[0].ID != 2 {
		t.Errorf("got user %d, want 2", got[0].ID)
	}
}

func TestStorageNearbyExcludesSuspended(t *testing.T) {
	s := NewStorage()
	s.Upsert(&User{ID: 1, Latitude: 10, Longitude: 10, Gender: GenderMale, LookingFor: GenderAny})
	target := &User{
		ID: 2, Latitude: 10.1, Longitude: 10.1,
		Gender: GenderFemale, LookingFor: GenderAny,
		SuspendedUntil: time.Now().Add(time.Hour),
	}
	s.Upsert(target)

	got := s.Nearby(1, 10, 10, 100)
	if len(got) != 0 {
		t.Errorf("suspended user should be excluded, got %d", len(got))
	}
}

func TestStorageNearbyExcludesLowRating(t *testing.T) {
	s := NewStorage()
	s.Upsert(&User{ID: 1, Latitude: 10, Longitude: 10, Gender: GenderMale, LookingFor: GenderAny})
	// Average rating 1.0 over 5 reviews — well below threshold.
	target := &User{
		ID: 2, Latitude: 10.1, Longitude: 10.1,
		Gender: GenderFemale, LookingFor: GenderAny,
		RatingSum: 5, RatingCount: 5,
	}
	s.Upsert(target)

	got := s.Nearby(1, 10, 10, 100)
	if len(got) != 0 {
		t.Errorf("low-rated user should be excluded, got %d", len(got))
	}
}

func TestStorageNearbyExcludesBlocked(t *testing.T) {
	s := NewStorage()
	me := &User{ID: 1, Latitude: 10, Longitude: 10, Gender: GenderMale, LookingFor: GenderAny}
	target := &User{ID: 2, Latitude: 10.1, Longitude: 10.1, Gender: GenderFemale, LookingFor: GenderAny}
	s.Upsert(me)
	s.Upsert(target)
	me.Blocks = map[int64]bool{2: true}

	got := s.Nearby(1, 10, 10, 100)
	if len(got) != 0 {
		t.Errorf("blocked user should be excluded, got %d", len(got))
	}
}

func TestStorageNearbyExcludesBusy(t *testing.T) {
	s := NewStorage()
	s.Upsert(&User{ID: 1, Latitude: 10, Longitude: 10, Gender: GenderMale, LookingFor: GenderAny})
	target := &User{
		ID: 2, Latitude: 10.1, Longitude: 10.1,
		Gender: GenderFemale, LookingFor: GenderAny, State: StateInChat,
	}
	s.Upsert(target)

	got := s.Nearby(1, 10, 10, 100)
	if len(got) != 0 {
		t.Errorf("busy user should be excluded, got %d", len(got))
	}
}

func TestGenderMatches(t *testing.T) {
	if !genderMatches(GenderAny, GenderMale) {
		t.Error("GenderAny should match Male")
	}
	if !genderMatches(GenderMale, GenderMale) {
		t.Error("Male filter should match Male target")
	}
	if genderMatches(GenderMale, GenderFemale) {
		t.Error("Male filter should NOT match Female target")
	}
	if !genderMatches("", GenderFemale) {
		t.Error("empty filter should match anyone")
	}
}

func TestHourInRange(t *testing.T) {
	tests := []struct {
		name           string
		hour, from, to int
		want           bool
	}{
		{"mid morning, 9-23", 10, 9, 23, true},
		{"early, 9-23", 8, 9, 23, false},
		{"late, 9-23", 23, 9, 23, false},
		{"wrap 22-3, 22", 22, 22, 3, true},
		{"wrap 22-3, 2", 2, 22, 3, true},
		{"wrap 22-3, 4", 4, 22, 3, false},
		{"wrap 22-3, 12", 12, 22, 3, false},
		{"from==to always", 5, 9, 9, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := hourInRange(tc.hour, tc.from, tc.to); got != tc.want {
				t.Errorf("hourInRange(%d, %d, %d) = %v, want %v",
					tc.hour, tc.from, tc.to, got, tc.want)
			}
		})
	}
}

func TestUserIsAwakeNoTimezone(t *testing.T) {
	u := &User{} // no timezone, no wake hours
	if !u.IsAwake() {
		t.Error("user with no timezone should be treated as awake")
	}
}

func TestUserIsAwakeValidTimezone(t *testing.T) {
	u := &User{Timezone: "UTC", WakeFrom: 9, WakeTo: 23}
	// We can't assert a specific outcome without knowing the test
	// runner's wall clock — but it should be either true or false,
	// not panic.
	_ = u.IsAwake()
}

func TestUserIsAwakeInvalidTimezone(t *testing.T) {
	u := &User{Timezone: "Not/A/Zone", WakeFrom: 9, WakeTo: 23}
	if !u.IsAwake() {
		t.Error("invalid timezone should be treated as awake (fail-open)")
	}
}

func TestAverageRating(t *testing.T) {
	u := &User{RatingSum: 14, RatingCount: 4}
	if got := u.AverageRating(); math.Abs(got-3.5) > 0.001 {
		t.Errorf("AverageRating = %f, want 3.5", got)
	}
	if (&User{}).AverageRating() != 0 {
		t.Error("unrated user should have 0 average")
	}
}

func TestCityKey(t *testing.T) {
	a := &User{Latitude: 35.71, Longitude: 51.41}
	b := &User{Latitude: 35.69, Longitude: 51.39}
	// Both should bucket to 35.5, 51.5 (rounded to 0.5).
	if a.CityKey() != b.CityKey() {
		t.Errorf("nearby users should share a city key: %s vs %s",
			a.CityKey(), b.CityKey())
	}

	c := &User{Latitude: 32.6, Longitude: 51.7}
	if a.CityKey() == c.CityKey() {
		t.Error("distant users should have different city keys")
	}

	if (&User{Latitude: 0, Longitude: 0}).CityKey() != "" {
		t.Error("(0,0) should return empty key")
	}
}

func TestIncrementReport(t *testing.T) {
	s := NewStorage()
	if got := s.IncrementReport(7); got != 1 {
		t.Errorf("first report = %d, want 1", got)
	}
	if got := s.IncrementReport(7); got != 2 {
		t.Errorf("second report = %d, want 2", got)
	}
	s.ResetReports(7)
	if got := s.IncrementReport(7); got != 1 {
		t.Errorf("after reset = %d, want 1", got)
	}
}

func TestAchievementFirstChat(t *testing.T) {
	u := &User{ChatCount: 1, Achievements: map[string]bool{}}
	now := time.Now()
	granted := checkAchievements(u, nil, now)
	if !u.HasAchievement(AchievementFirstChat) {
		t.Error("ChatCount=1 should grant FirstChat")
	}
	found := false
	for _, a := range granted {
		if a.ID == AchievementFirstChat {
			found = true
		}
	}
	if !found {
		t.Error("granted list should include FirstChat")
	}
}

func TestAchievementNoDuplicates(t *testing.T) {
	u := &User{ChatCount: 5, Achievements: map[string]bool{}}
	now := time.Now()
	first := checkAchievements(u, nil, now)
	second := checkAchievements(u, nil, now)
	if len(second) != 0 {
		t.Errorf("second pass should grant no new achievements, got %d", len(second))
	}
	_ = first
}

func TestAchievementFiveStar(t *testing.T) {
	u := &User{
		RatingSum: 5, RatingCount: 1,
		Achievements: map[string]bool{},
	}
	now := time.Now()
	granted := checkAchievements(u, nil, now)
	found := false
	for _, a := range granted {
		if a.ID == AchievementFiveStar {
			found = true
		}
	}
	if !found {
		t.Error("RatingSum=5, RatingCount=1 should grant FiveStar")
	}
}

func TestAchievementWellLiked(t *testing.T) {
	u := &User{
		RatingSum: 23, RatingCount: 5, // avg 4.6
		Achievements: map[string]bool{},
	}
	now := time.Now()
	granted := checkAchievements(u, nil, now)
	found := false
	for _, a := range granted {
		if a.ID == AchievementWellLiked {
			found = true
		}
	}
	if !found {
		t.Error("avg 4.6 with 5 reviews should grant WellLiked")
	}
}

func TestAchievementCities(t *testing.T) {
	u := &User{
		ChatCount: 1,
		CitiesChat: map[string]int64{
			"a": 1, "b": 2, "c": 3,
		},
		Achievements: map[string]bool{},
	}
	now := time.Now()
	granted := checkAchievements(u, nil, now)
	found := false
	for _, a := range granted {
		if a.ID == AchievementMultiCity {
			found = true
		}
	}
	if !found {
		t.Error("3 distinct cities should grant MultiCity")
	}
}

func TestFormatDistance(t *testing.T) {
	cases := []struct {
		km   float64
		want string
	}{
		{0.5, "500 m"},
		{1.5, "1.5 km"},
		{12, "12 km"},
		{150, "150 km"},
	}
	for _, c := range cases {
		if got := formatDistance(c.km); got != c.want {
			t.Errorf("formatDistance(%f) = %q, want %q", c.km, got, c.want)
		}
	}
}

func TestFormatMatchLabel(t *testing.T) {
	c := &User{FirstName: "Ali", Interests: []string{"code", "books"}, Gender: GenderMale}
	got := formatMatchLabel(c, 5)
	if got == "" {
		t.Error("formatMatchLabel should return non-empty")
	}
	// No rating yet — emoji should be ♂.
	if !contains(got, "♂") {
		t.Error("expected male emoji in label")
	}
}

func TestInterestSetAndValid(t *testing.T) {
	s := []string{"code", "books"}
	set := interestSet(s)
	if !set["code"] || !set["books"] || set["travel"] {
		t.Error("set should contain code, books and not travel")
	}
	if !validInterest("code") {
		t.Error("code should be a valid interest")
	}
	if validInterest("not-a-tag") {
		t.Error("not-a-tag should not be a valid interest")
	}
}

func TestSortedInterests(t *testing.T) {
	u := &User{Interests: []string{"travel", "books", "code"}}
	s := u.SortedInterests()
	if s[0] != "books" || s[1] != "code" || s[2] != "travel" {
		t.Errorf("expected alphabetical order, got %v", s)
	}
}

func TestTranslateClientNoEndpoint(t *testing.T) {
	c := NewTranslateClient("", "")
	_, err := c.Translate(nil, "hi", "en", "fa")
	if err != ErrNoClient {
		t.Errorf("expected ErrNoClient, got %v", err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
