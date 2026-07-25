package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	apiV1 "github.com/solargate/grom/api/v1"
	"github.com/solargate/grom/internal/config"
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
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: ./config.yaml)")
}
