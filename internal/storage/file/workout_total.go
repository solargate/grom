package file

import (
	"log/slog"

	"github.com/solargate/grom/internal/users"
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
	if profile.TotalWorkouts == nil {
		if err := b.workoutRepo.recountWorkouts(nickname, &profile); err != nil {
			return nil, err
		}
		if err := b.users.saveProfile(nickname, profile); err != nil {
			slog.Warn("failed to save workout total", "user", nickname, "err", err)
		}
	}
	return &profile, nil
}

// withWorkoutTotal runs under workoutCommitMu. A missing total on disk makes
// an interrupted update recoverable from the committed workout metadata.
func (s *WorkoutsStore) withWorkoutTotal(nickname string, delta int, mutate func() error) error {
	if s.users == nil || delta == 0 {
		return mutate()
	}
	s.users.mu.Lock()
	defer s.users.mu.Unlock()

	profile, err := s.users.loadProfile(nickname)
	if err != nil {
		return err
	}
	previous := profile.TotalWorkouts
	if previous != nil {
		profile.TotalWorkouts = nil
		if err := s.users.saveProfile(nickname, profile); err != nil {
			return err
		}
	}

	mutationErr := mutate()
	if mutationErr == nil && previous != nil && *previous+delta >= 0 {
		total := *previous + delta
		profile.TotalWorkouts = &total
	} else if err := s.recountWorkouts(nickname, &profile); err != nil {
		slog.Warn("failed to count workouts", "user", nickname, "err", err)
		return mutationErr
	}
	if err := s.users.saveProfile(nickname, profile); err != nil {
		slog.Warn("failed to save workout total", "user", nickname, "err", err)
	}
	return mutationErr
}

func (s *WorkoutsStore) recountWorkouts(nickname string, profile *users.Profile) error {
	items, err := s.listAll(nickname)
	if err != nil {
		return err
	}
	total := len(items)
	profile.TotalWorkouts = &total
	return nil
}
