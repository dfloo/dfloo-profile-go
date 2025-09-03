package main

import (
	"log"
	"net/http"

	"github.com/dfloo/dfloo-profile-go/internal/router"
)

func main() {
	log.Print("Server listening on http://localhost:8080")
	err := http.ListenAndServe(":8080", router.New())
	if err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
