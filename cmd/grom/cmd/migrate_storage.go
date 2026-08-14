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

Copied: users, profiles, equipment, workouts, follows, federation followers and
inbox, workout likes and comments (local, federated cache, outbound activity ids),
personal access tokens, and speed/heart-rate charts (converted between file JSON
blobs and bbolt binary buckets).

Not copied: password-reset tokens (short-lived; in-flight reset links become
invalid). Blob files (tracks, photos, avatars, keys) under storage.location are
shared and not duplicated.

Legacy plain-text Like activity ids (without object_id) are reconstructed via
federated inbox and, when federation.domain is set, local workout object URLs.

After a successful migration, set storage.driver in your config to the target
driver and restart the server.

Stop the server before running this command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrateFrom == "" || migrateTo == "" {
			return fmt.Errorf("--from and --to are required")
		}
		config.GetConfig(configPath)
		result, err := migrate.Run(migrate.Options{
			From:             config.StorageDriver(migrateFrom),
			To:               config.StorageDriver(migrateTo),
			Config:           config.Cfg.Storage,
			FederationDomain: config.Cfg.Federation.Domain,
			DryRun:           migrateDryRun,
			Verify:           migrateVerify,
			Force:            migrateForce,
		})
		if err != nil {
			return fmt.Errorf("migrate-storage failed: %w", err)
		}
		fmt.Printf("Migrated metadata: users=%d profiles=%d equipment=%d follows=%d workouts=%d fed_followers=%d fed_authors=%d fed_inbox=%d local_likes=%d fed_likes=%d like_activities=%d local_comments=%d fed_comments=%d comment_activities=%d local_speed_charts=%d local_hr_charts=%d fed_speed_charts=%d fed_hr_charts=%d pats=%d\n",
			result.Users, result.Profiles, result.Equipment, result.Follows, result.Workouts,
			result.FedFollowers, result.FedAuthors, result.FedInboxWorkouts,
			result.LocalLikes, result.FedLikes, result.LikeActivities,
			result.LocalComments, result.FedComments, result.CommentActivities,
			result.LocalSpeedCharts, result.LocalHeartRateCharts,
			result.FedSpeedCharts, result.FedHeartRateCharts,
			result.PersonalAccessTokens)
		fmt.Println("Note: password-reset tokens are not copied; in-flight reset links become invalid.")
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
