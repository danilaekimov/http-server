package main

import (
	"log"
	"net/http"

	"http-server/service"
)

func main() {
	mux := http.NewServeMux()

	srv := service.NewService()

	mux.Handle("/vote", http.HandlerFunc(srv.VoteHandler))
	mux.Handle("/stats", http.HandlerFunc(srv.StatsHandler))

	log.Println("Starting server on :8000...")
	err := http.ListenAndServe(":8000", mux)
	if err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}
