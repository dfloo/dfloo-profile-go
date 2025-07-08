package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Resume struct {
	ResumeID string `json:"resumeId"`
	Name     string `json:"name"`
}

func getResume(w http.ResponseWriter, r *http.Request) {
	log.Println("getting resume...")
	w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-allow-Headers", "Content-Type Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Resume{
		ResumeID: "1",
		Name:     "Mock Resume",
	})
}

func main() {
	server := http.NewServeMux()
	server.HandleFunc("GET /resume", getResume)

	log.Println("Starting server...")
	err := http.ListenAndServe(":8080", server)
	log.Fatal(err)
}
