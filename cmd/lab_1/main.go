package main

import (
	"log"

	"lab_1/internal/api"
)

func main() {
	log.Println("application start")
	api.StartServer()
	log.Println("application terminated")
}
