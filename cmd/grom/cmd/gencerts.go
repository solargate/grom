package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/solargate/grom/internal/tlsutil"
)

var (
	gencertsIP     string
	gencertsDomain string
	gencertsOut    string
)

var gencertsCmd = &cobra.Command{
	Use:   "gencerts",
	Short: "Generate a self-signed TLS certificate and CA for local development",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := tlsutil.GenerateCerts(tlsutil.GenOptions{
			IP:     gencertsIP,
			Domain: gencertsDomain,
			OutDir: gencertsOut,
		}); err != nil {
			return fmt.Errorf("gencerts failed: %w", err)
		}
		fmt.Printf("Certificates written to %s\n", gencertsOut)
		return nil
	},
	SilenceUsage: true,
}

func init() {
	gencertsCmd.Flags().StringVarP(&gencertsIP, "ip", "i", "", "Server IP for certificate SAN")
	gencertsCmd.Flags().StringVarP(&gencertsDomain, "domain", "d", "", "Server domain for certificate SAN")
	gencertsCmd.Flags().StringVarP(&gencertsOut, "out", "o", "tls", "Output directory")
	rootCmd.AddCommand(gencertsCmd)
}
