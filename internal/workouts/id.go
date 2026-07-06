package workouts

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"
)

const (
	workoutIDLength   = 8
	workoutIDAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func newWorkoutID() (string, error) {
	b := make([]byte, workoutIDLength)
	alphabetSize := big.NewInt(int64(len(workoutIDAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, alphabetSize)
		if err != nil {
			return "", fmt.Errorf("generate workout id: %w", err)
		}
		b[i] = workoutIDAlphabet[n.Int64()]
	}
	return string(b), nil
}

func (s *Store) workoutIDExists(userDir, id string) bool {
	suffix := "-" + id
	entries, err := os.ReadDir(userDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
			return true
		}
	}
	return false
}

func (s *Store) allocateWorkoutID(userDir string) (string, error) {
	const maxAttempts = 10
	for range maxAttempts {
		id, err := newWorkoutID()
		if err != nil {
			return "", err
		}
		if !s.workoutIDExists(userDir, id) {
			return id, nil
		}
	}
	return "", fmt.Errorf("failed to allocate unique workout id after %d attempts", maxAttempts)
}
