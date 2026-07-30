package api

import (
	"encoding/json"
	"net/http"
	"torrent-backend/internal/models"
	"torrent-backend/internal/torrent"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	TorrentManager *torrent.TorrentManager
}


func NewHandler(tm *torrent.TorrentManager) *Handler {
	return &Handler{
		TorrentManager: tm,
	}
}


func (h *Handler) GetAllTorrents(w http.ResponseWriter, r *http.Request) {

	
	torrents := h.TorrentManager.GetAllTorrents()


	w.Header().Set("Content-Type", "application/json")

	
	json.NewEncoder(w).Encode(torrents)
}
// this is getting all the Req from the user

// Handles POST /torrent/add
func (h *Handler) AddTorrent(w http.ResponseWriter, r *http.Request) {

	// Create a Torrent object
	var newTorrent models.Torrent

	// Read JSON from the request body into newTorrent
	err := json.NewDecoder(r.Body).Decode(&newTorrent)

	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	// if id == null -> empty ID
	if newTorrent.ID == "" {
		http.Error(w, "Torrent ID is required", http.StatusBadRequest)
		return
	}
	// duplicate torrent 
	_, exists := h.TorrentManager.GetTorrent(newTorrent.ID)

	if exists {
		http.Error(w, "Torrent already exists", http.StatusConflict)
		return
	}
	// if name == null
	if newTorrent.Name == "" {
		http.Error(w, "Torrent name is required", http.StatusBadRequest)
		return
	}
	// invalid magnet link
	if len(newTorrent.MagnetLink) < 8 || newTorrent.MagnetLink[:8] != "magnet:" {
		http.Error(w, "Invalid magnet link", http.StatusBadRequest)
		return
	}
	// memory size of the torrent
	if newTorrent.Size <= 0 {
		http.Error(w, "Invalid torrent size", http.StatusBadRequest)
		return
	}
	// Add the torrent
	h.TorrentManager.AddTorrent(newTorrent)

	// Send success response
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Torrent added successfully"))
}
func (h *Handler) GetTorrent(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	torrent, exists := h.TorrentManager.GetTorrent(id)

	if !exists {
		http.Error(w, "Torrent not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(torrent)
}
func (h *Handler) UpdateTorrent(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	var updatedTorrent models.Torrent

	err := json.NewDecoder(r.Body).Decode(&updatedTorrent)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, exists := h.TorrentManager.GetTorrent(id)
	if !exists {
		http.Error(w, "Torrent not found", http.StatusNotFound)
		return
	}

	updatedTorrent.ID = id

	h.TorrentManager.AddTorrent(updatedTorrent)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTorrent)
}
func (h *Handler) DeleteTorrent(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	_, exists := h.TorrentManager.GetTorrent(id)
	if !exists {
		http.Error(w, "Torrent not found", http.StatusNotFound)
		return
	}

	h.TorrentManager.RemoveTorrent(id)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Torrent deleted successfully"))
}