package workouts_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/solargate/grom/internal/workouts"
)

func TestCreateValidation(t *testing.T) {
	svc := newTestService(t.TempDir())
	validStart := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		workout *workouts.Workout
		wantErr error
		wantMsg string
	}{
		{name: "nil", workout: nil, wantErr: workouts.ErrInvalidWorkout},
		{
			name:    "empty name",
			workout: &workouts.Workout{Name: "  ", SportType: "Run", StartDate: validStart},
			wantErr: workouts.ErrInvalidWorkout,
			wantMsg: "name is required",
		},
		{
			name:    "invalid sport",
			workout: &workouts.Workout{Name: "Run", SportType: "Nope", StartDate: validStart},
			wantErr: workouts.ErrInvalidSportType,
		},
		{
			name: "negative duration",
			workout: &workouts.Workout{
				Name: "Run", SportType: "Run", StartDate: validStart, DurationSeconds: -1,
			},
			wantErr: workouts.ErrInvalidWorkout,
			wantMsg: "duration",
		},
		{
			name: "negative distance",
			workout: &workouts.Workout{
				Name: "Run", SportType: "Run", StartDate: validStart, Distance: -1,
			},
			wantErr: workouts.ErrInvalidWorkout,
			wantMsg: "distance",
		},
		{
			name: "zero start date",
			workout: &workouts.Workout{
				Name: "Run", SportType: "Run",
			},
			wantErr: workouts.ErrInvalidWorkout,
			wantMsg: "start_date",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create("athlete", tc.workout)
			if err == nil {
				t.Fatal("expected error")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if tc.wantMsg != "" && !strings.Contains(err.Error(), tc.wantMsg) {
				t.Fatalf("err = %q, want substring %q", err.Error(), tc.wantMsg)
			}
		})
	}
}

func TestCreateValidBoundary(t *testing.T) {
	svc := newTestService(t.TempDir())
	created, err := svc.Create("athlete", &workouts.Workout{
		Name:            "Zero metrics",
		SportType:       "Run",
		StartDate:       time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 0,
		Distance:        0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected id")
	}
}
