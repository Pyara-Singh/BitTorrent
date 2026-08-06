package main

import (
	"fmt"
	"net/http"
	"torrent-backend/internal/api"
	"torrent-backend/internal/torrent"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	fmt.Println("BitTorrent Backend is starting...")

	manager := torrent.NewTorrentManager()
	handler := api.NewHandler(manager)
	mux := http.NewServeMux()
	api.RegisterRoutes(mux, handler)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
}
