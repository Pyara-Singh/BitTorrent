package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler) {

	// Home
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Torrent APIs
	r.Get("/torrent/all", handler.GetAllTorrents)

	r.Post("/torrent/add", handler.AddTorrent)
	r.Get("/torrent/{id}", handler.GetTorrent)
	r.Put("/torrent/{id}", handler.UpdateTorrent)
	r.Delete("/torrent/{id}", handler.DeleteTorrent)
}
