package api

import (
	"net/http"
	"strings"
)

// RegisterRoutes wires the HTTP API without external router dependencies.
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("BitTorrent Backend"))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	mux.HandleFunc("/torrent/all", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		handler.GetAllTorrents(w, r)
	})

	mux.HandleFunc("/torrent/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		handler.AddTorrent(w, r)
	})

	mux.HandleFunc("/torrent/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/torrent/")
		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}

		r = r.WithContext(withTorrentID(r.Context(), id))
		switch r.Method {
		case http.MethodGet:
			handler.GetTorrent(w, r)
		case http.MethodPut:
			handler.UpdateTorrent(w, r)
		case http.MethodDelete:
			handler.DeleteTorrent(w, r)
		default:
			methodNotAllowed(w)
		}
	})
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
