package main

import (
	"log"
	"net/http"
	"os"

	"remi-jadul/server/internal/auth"
	"remi-jadul/server/internal/transport"
)

func main() {
	store := auth.NewGuestStore()
	handler := transport.NewGuestHandler(store)
	mux := http.NewServeMux()
	mux.Handle("/api/v1/guests", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
