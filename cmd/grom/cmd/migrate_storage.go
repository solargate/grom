package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/storage/migrate"
)

var (
	migrateFrom   string
	migrateTo     string
	migrateDryRun bool
	migrateVerify bool
	migrateForce  bool
)

var migrateStorageCmd = &cobra.Command{
	Use:   "migrate-storage",
	Short: "Copy storage metadata between file and bbolt drivers (blobs stay on disk)",
	Long: `Migrate application metadata between storage drivers.

Blob files (tracks, photos, avatars, keys) under storage.location are shared and
not copied. Speed and heart-rate charts are converted between file JSON blobs and
bbolt binary buckets. After a successful migration, set storage.driver in your
config to the target driver and restart the server.

Stop the server before running this command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrateFrom == "" || migrateTo == "" {
			return fmt.Errorf("--from and --to are required")
		}
		config.GetConfig(configPath)
		result, err := migrate.Run(migrate.Options{
			From:   config.StorageDriver(migrateFrom),
			To:     config.StorageDriver(migrateTo),
			Config: config.Cfg.Storage,
			DryRun: migrateDryRun,
			Verify: migrateVerify,
			Force:  migrateForce,
		})
		if err != nil {
			return fmt.Errorf("migrate-storage failed: %w", err)
		}
		fmt.Printf("Migrated metadata: users=%d equipment=%d follows=%d workouts=%d fed_followers=%d fed_authors=%d fed_inbox=%d\n",
			result.Users, result.Equipment, result.Follows, result.Workouts,
			result.FedFollowers, result.FedAuthors, result.FedInboxWorkouts)
		if migrateDryRun {
			fmt.Println("(dry-run: no changes written)")
		} else {
			fmt.Printf("Set storage.driver: %s in your config, then restart the server.\n", migrateTo)
		}
		return nil
	},
	SilenceUsage: true,
}

func init() {
	migrateStorageCmd.Flags().StringVar(&migrateFrom, "from", "", "Source driver: file | bbolt")
	migrateStorageCmd.Flags().StringVar(&migrateTo, "to", "", "Target driver: file | bbolt")
	migrateStorageCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Count source records without writing")
	migrateStorageCmd.Flags().BoolVar(&migrateVerify, "verify", false, "Compare source and target counts after migration")
	migrateStorageCmd.Flags().BoolVar(&migrateForce, "force", false, "Overwrite existing bbolt database when targeting bbolt")
	_ = migrateStorageCmd.MarkFlagRequired("from")
	_ = migrateStorageCmd.MarkFlagRequired("to")
	rootCmd.AddCommand(migrateStorageCmd)
}
