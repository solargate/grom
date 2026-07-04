package main

import (
	"flag"
	"fmt"

	apiV1 "github.com/solargate/travka/api/v1"
	"github.com/solargate/travka/config"
)

func main() {
	fmt.Println("Travka 0.0.1")

	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	config.GetConfig(*configPath)

	apiV1.RunRouter()
}
