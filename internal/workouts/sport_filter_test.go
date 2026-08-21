package workouts

import "testing"

func TestMatchSportType(t *testing.T) {
	if !MatchSportType("Run", nil) {
		t.Fatal("nil allow should match")
	}
	allow := map[string]struct{}{"Run": {}, "Ride": {}}
	if !MatchSportType("Run", allow) {
		t.Fatal("Run should match")
	}
	if MatchSportType("Walk", allow) {
		t.Fatal("Walk should not match")
	}
	if MatchSportType("Run", map[string]struct{}{}) {
		t.Fatal("empty allow should match nothing")
	}
}
