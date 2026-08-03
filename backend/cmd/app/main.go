package main

import (
	"fmt"
	"net/http"
	"torrent-backend/internal/api"
	"torrent-backend/internal/torrent"

	"github.com/go-chi/chi/v5"
)

func main() {

	fmt.Println("BitTorrent Backend is starting...")

	// Create Torrent Manager
	torrentManager := torrent.NewTorrentManager()

	// Create API Handler
	handler := api.NewHandler(torrentManager)

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	http.HandleFunc("/torrent/all", handler.GetAllTorrents)
	r := chi.NewRouter()

	api.RegisterRoutes(r, handler)

	http.ListenAndServe(":8080", r)
}
