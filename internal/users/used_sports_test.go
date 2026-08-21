package users

import (
	"reflect"
	"testing"
)

func TestMoveSportToFront(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		list  []string
		sport string
		want  []string
	}{
		{name: "empty sport", list: []string{"Run"}, sport: "", want: []string{"Run"}},
		{name: "prepend new", list: []string{"Run", "Ride"}, sport: "Walk", want: []string{"Walk", "Run", "Ride"}},
		{name: "move existing", list: []string{"Walk", "Run", "Ride"}, sport: "Run", want: []string{"Run", "Walk", "Ride"}},
		{name: "first already", list: []string{"Run", "Walk"}, sport: "Run", want: []string{"Run", "Walk"}},
		{name: "nil list", list: nil, sport: "Run", want: []string{"Run"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MoveSportToFront(tc.list, tc.sport)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestPruneUsedSports(t *testing.T) {
	t.Parallel()

	remaining := map[string]struct{}{"Run": {}, "Walk": {}}
	got := PruneUsedSports([]string{"Ride", "Run", "Hike", "Walk"}, remaining)
	want := []string{"Run", "Walk"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}

	if PruneUsedSports([]string{"Ride"}, remaining) != nil {
		t.Fatalf("expected nil when nothing remains")
	}
	if PruneUsedSports(nil, remaining) != nil {
		t.Fatalf("expected nil for empty list")
	}
}
