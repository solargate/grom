package main

import (
	"flag"
	"fmt"
	"os"

	apiV1 "github.com/solargate/travka/api/v1"
	"github.com/solargate/travka/internal/config"
	"github.com/solargate/travka/internal/tlsutil"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "gencerts" {
		runGenCerts(os.Args[2:])
		return
	}

	fmt.Println("Travka 0.0.1")

	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	config.GetConfig(*configPath)

	apiV1.RunRouter()
}

func runGenCerts(args []string) {
	fs := flag.NewFlagSet("gencerts", flag.ExitOnError)
	ip := fs.String("ip", "", "Server IP for certificate SAN")
	domain := fs.String("domain", "", "Server domain for certificate SAN")
	out := fs.String("out", "tls", "Output directory")
	_ = fs.Parse(args)

	if err := tlsutil.GenerateCerts(tlsutil.GenOptions{
		IP:     *ip,
		Domain: *domain,
		OutDir: *out,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "gencerts failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Certificates written to %s\n", *out)
}
