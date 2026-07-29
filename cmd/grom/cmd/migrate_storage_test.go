package cmd_test

import (
	"testing"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/storage/migrate"
)

func TestMigrateRunRejectsSameDrivers(t *testing.T) {
	_, err := migrate.Run(migrate.Options{
		From:   config.StorageDriverFile,
		To:     config.StorageDriverFile,
		Config: config.StorageConfig{ResolvedLocation: t.TempDir()},
	})
	if err == nil {
		t.Fatal("expected error when from == to")
	}
}

func TestMigrateRunRejectsUnsupportedDriver(t *testing.T) {
	_, err := migrate.Run(migrate.Options{
		From:   config.StorageDriverPostgres,
		To:     config.StorageDriverFile,
		Config: config.StorageConfig{ResolvedLocation: t.TempDir()},
	})
	if err == nil {
		t.Fatal("expected unsupported driver error")
	}
}
