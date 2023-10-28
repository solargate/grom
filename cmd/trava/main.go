package main

import (
	"fmt"

	apiV1 "github.com/solargate/trava/api/v1"
	"github.com/solargate/trava/config"
)

func main() {
	fmt.Println("Trava 0.0.1")

	config.GetConfig()

	apiV1.RunRouter()
}
