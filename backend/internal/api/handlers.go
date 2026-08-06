package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"torrent-backend/internal/models"
	"torrent-backend/internal/torrent"
)

type torrentIDKey struct{}

type Handler struct {
	TorrentManager *torrent.TorrentManager
}

func NewHandler(tm *torrent.TorrentManager) *Handler {
	return &Handler{TorrentManager: tm}
}

func (h *Handler) GetAllTorrents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.TorrentManager.GetAllTorrents())
}

func (h *Handler) AddTorrent(w http.ResponseWriter, r *http.Request) {
	var newTorrent models.Torrent
	if err := json.NewDecoder(r.Body).Decode(&newTorrent); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := validateTorrent(newTorrent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, exists := h.TorrentManager.GetTorrent(newTorrent.ID); exists {
		http.Error(w, "torrent already exists", http.StatusConflict)
		return
	}

	h.TorrentManager.AddTorrent(newTorrent)
	writeJSON(w, http.StatusCreated, newTorrent)
}

func (h *Handler) GetTorrent(w http.ResponseWriter, r *http.Request) {
	id := torrentID(r)
	torrent, exists := h.TorrentManager.GetTorrent(id)
	if !exists {
		http.Error(w, "torrent not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, torrent)
}

func (h *Handler) UpdateTorrent(w http.ResponseWriter, r *http.Request) {
	id := torrentID(r)
	if _, exists := h.TorrentManager.GetTorrent(id); !exists {
		http.Error(w, "torrent not found", http.StatusNotFound)
		return
	}

	var updatedTorrent models.Torrent
	if err := json.NewDecoder(r.Body).Decode(&updatedTorrent); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	updatedTorrent.ID = id
	if err := validateTorrent(updatedTorrent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.TorrentManager.AddTorrent(updatedTorrent)
	writeJSON(w, http.StatusOK, updatedTorrent)
}

func (h *Handler) DeleteTorrent(w http.ResponseWriter, r *http.Request) {
	id := torrentID(r)
	if _, exists := h.TorrentManager.GetTorrent(id); !exists {
		http.Error(w, "torrent not found", http.StatusNotFound)
		return
	}

	h.TorrentManager.RemoveTorrent(id)
	w.WriteHeader(http.StatusNoContent)
}

func validateTorrent(t models.Torrent) error {
	if t.ID == "" {
		return errors.New("torrent ID is required")
	}
	if t.Name == "" {
		return errors.New("torrent name is required")
	}
	if !strings.HasPrefix(t.MagnetLink, "magnet:") {
		return errors.New("invalid magnet link")
	}
	if t.Size <= 0 {
		return errors.New("invalid torrent size")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func withTorrentID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, torrentIDKey{}, id)
}

func torrentID(r *http.Request) string {
	id, _ := r.Context().Value(torrentIDKey{}).(string)
	return id
}
