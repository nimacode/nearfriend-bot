package bot

import (
	"math"
	"testing"
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
			name: "Tehran to Karaj ~32km",
			lat1: 35.6892, lon1: 51.3890, // Tehran
			lat2: 35.8400, lon2: 50.9391, // Karaj
			want: 42, tolerance: 5, // generous — actual ~42km
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

	// Upsert overwrites.
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

func TestStorageNearby(t *testing.T) {
	s := NewStorage()

	// Self at lat 10, lon 10.
	me := &User{ID: 1, Latitude: 10, Longitude: 10, Gender: GenderMale, LookingFor: GenderAny}
	s.Upsert(me)

	// 50 km away — should be inside 100 km radius.
	near := &User{ID: 2, Latitude: 10.5, Longitude: 10.5, Gender: GenderFemale, LookingFor: GenderAny}
	s.Upsert(near)

	// 1000 km away — outside 100 km radius.
	far := &User{ID: 3, Latitude: 20, Longitude: 20, Gender: GenderFemale, LookingFor: GenderAny}
	s.Upsert(far)

	// Not registered (no gender) — should be excluded.
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