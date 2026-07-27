package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	apiV1 "github.com/solargate/grom/api/v1"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/logging"
	"github.com/solargate/grom/internal/version"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:     "grom",
	Short:   "Self-hosted workout tracker with optional ActivityPub federation",
	Long:    `Grom is a self-hosted workout tracker. Run without a subcommand to start the HTTP server.`,
	Version: version.Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Grom %s\n", version.Version)
		config.GetConfig(configPath)
		if err := logging.SetupDefault(config.Cfg.Logging); err != nil {
			return err
		}
		slog.Info("starting server",
			"tls_mode", config.Cfg.Server.TLS.Mode,
			"federation_enabled", config.Cfg.Federation.Enabled,
			"storage_driver", storageDriverLabel(config.Cfg.Storage.Driver),
			"log_level", config.Cfg.Logging.Level,
			"log_format", config.Cfg.Logging.Format,
		)
		apiV1.RunRouter()
		return nil
	},
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate("Grom {{.Version}}\n")
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: ./config.yaml)")
}

func storageDriverLabel(driver config.StorageDriver) string {
	if driver == "" {
		return string(config.StorageDriverFile)
	}
	return string(driver)
}
