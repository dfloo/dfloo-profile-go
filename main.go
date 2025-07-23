package main

import (
	"log"
	"net/http"

	"github.com/dfloo/dfloo-profile-go/router"
)

func main() {
	router := router.New()

	log.Print("Server listening on http://localhost:8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
