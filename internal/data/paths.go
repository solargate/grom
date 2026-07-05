package data

import (
	"fmt"
	"os"
	"path/filepath"
)

func ResolveDataDir(location string) (string, error) {
	if filepath.IsAbs(location) {
		return location, nil
	}

	exec, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	exec, err = filepath.EvalSymlinks(exec)
	if err != nil {
		return "", fmt.Errorf("resolve executable symlinks: %w", err)
	}

	return filepath.Join(filepath.Dir(exec), location), nil
}
