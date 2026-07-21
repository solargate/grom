package strava

import "testing"

func TestMapSportTypePackraftAndKayaking(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		name string
		want string
	}{
		{raw: "Kayaking", want: "Kayaking"},
		{raw: "каякинг", want: "Kayaking"},
		{raw: "Packraft", want: "Packraft"},
		{raw: "packraft", want: "Packraft"},
		{raw: "пакрафт", want: "Packraft"},
		{raw: "Workout", name: "Morning packraft", want: "Packraft"},
		{raw: "Тренировка", name: "Пакрафт по реке", want: "Packraft"},
		{raw: "Kayaking", name: "packraft trip", want: "Kayaking"},
	}

	for _, tt := range tests {
		got, err := mapSportType(tt.raw, tt.name)
		if err != nil {
			t.Fatalf("mapSportType(%q, %q): %v", tt.raw, tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("mapSportType(%q, %q) = %q, want %q", tt.raw, tt.name, got, tt.want)
		}
	}
}
