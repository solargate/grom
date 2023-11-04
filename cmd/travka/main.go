package main

import (
	"fmt"

	apiV1 "github.com/solargate/travka/api/v1"
	"github.com/solargate/travka/config"
)

func main() {
	fmt.Println("Travka 0.0.1")

	config.GetConfig()

	apiV1.RunRouter()
}
