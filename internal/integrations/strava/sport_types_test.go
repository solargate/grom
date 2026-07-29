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

func TestMapSportTypeNordicWalkIceHockeyAndHIIT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		name string
		want string
	}{
		{raw: "NordicWalk", want: "NordicWalk"},
		{raw: "Nordic Walk", want: "NordicWalk"},
		{raw: "Nordic Walking", want: "NordicWalk"},
		{raw: "скандинавская ходьба", want: "NordicWalk"},
		{raw: "IceHockey", want: "IceHockey"},
		{raw: "Ice Hockey", want: "IceHockey"},
		{raw: "Hockey", want: "IceHockey"},
		{raw: "хоккей", want: "IceHockey"},
		{raw: "HIIT", want: "HIIT"},
		{raw: "HighIntensityIntervalTraining", want: "HIIT"},
		{raw: "поход", want: "Hike"},
		{raw: "инвалидная коляска", want: "Wheelchair"},
		{raw: "ручной велосипед", want: "Handcycle"},
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
