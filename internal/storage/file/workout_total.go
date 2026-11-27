package file

import (
	"log/slog"

	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

func (b *Backend) GetProfileWithWorkoutTotal(userID string) (*users.Profile, error) {
	workoutCommitMu.Lock()
	defer workoutCommitMu.Unlock()
	b.users.mu.Lock()
	defer b.users.mu.Unlock()

	nickname, err := b.users.nicknameByIDLocked(userID)
	if err != nil {
		return nil, err
	}
	profile, err := b.users.loadProfile(nickname)
	if err != nil {
		return nil, err
	}
	if !profile.HasWorkoutStats() {
		if err := b.workoutRepo.rebuildWorkoutStats(nickname, &profile); err != nil {
			return nil, err
		}
		if err := b.users.saveProfile(nickname, profile); err != nil {
			slog.Warn("failed to save workout statistics", "user", nickname, "err", err)
		}
	}
	return &profile, nil
}

func (s *WorkoutsStore) withWorkoutChange(nickname string, before, after *workouts.Workout, mutate func() error) error {
	return s.withWorkoutStats(nickname, users.SameWorkoutSummary(before, after), func(profile *users.Profile) bool {
		return profile.ApplyWorkoutChange(before, after)
	}, mutate)
}

// withWorkoutStats runs under workoutCommitMu. Invalidation keeps interrupted
// changes recoverable from the committed workout metadata.
func (s *WorkoutsStore) withWorkoutStats(nickname string, unchanged bool, apply func(*users.Profile) bool, mutate func() error) error {
	if s.users == nil {
		return mutate()
	}
	s.users.mu.Lock()
	defer s.users.mu.Unlock()

	profile, err := s.users.loadProfile(nickname)
	if err != nil {
		return err
	}
	if unchanged && profile.HasWorkoutStats() {
		return mutate()
	}
	if profile.TotalWorkouts != nil || profile.Workouts != nil {
		invalidated := profile
		invalidated.TotalWorkouts = nil
		invalidated.Workouts = nil
		if err := s.users.saveProfile(nickname, invalidated); err != nil {
			return err
		}
	}

	mutationErr := mutate()
	if mutationErr != nil || !apply(&profile) {
		if err := s.rebuildWorkoutStats(nickname, &profile); err != nil {
			slog.Warn("failed to rebuild workout statistics", "user", nickname, "err", err)
			return mutationErr
		}
	}
	if err := s.users.saveProfile(nickname, profile); err != nil {
		slog.Warn("failed to save workout statistics", "user", nickname, "err", err)
	}
	return mutationErr
}

func (s *WorkoutsStore) rebuildWorkoutStats(nickname string, profile *users.Profile) error {
	items, err := s.listAll(nickname)
	if err != nil {
		return err
	}
	profile.RebuildWorkoutStats(items)
	return nil
}
