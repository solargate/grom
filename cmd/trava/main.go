package main

import (
	"fmt"
	apiV1 "trava/api/v1"
)

func main() {
	fmt.Println("Trava 0.0.1")

	apiV1.RunRouter()
}
