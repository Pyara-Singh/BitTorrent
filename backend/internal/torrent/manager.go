package torrent


import (
	"sync"
	"torrent-backend/internal/models"
)
// If both modify the map simultaneously, Go may crash
type TorrentManager struct {
	torrents map[string]models.Torrent
	mu       sync.RWMutex
}

// Constructor
func NewTorrentManager() *TorrentManager {
	return &TorrentManager{
		torrents: make(map[string]models.Torrent),
	}
}

// Add a torrent
func (tm *TorrentManager) AddTorrent(torrent models.Torrent) {

	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.torrents[torrent.ID] = torrent
}

// Get one torrent by ID
func (tm *TorrentManager) GetTorrent(id string) (models.Torrent, bool) {

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	torrent, exists := tm.torrents[id]
	return torrent, exists
}
// Get all torrents
func (tm *TorrentManager) GetAllTorrents() []models.Torrent {

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var torrents []models.Torrent

	for _, torrent := range tm.torrents {
		torrents = append(torrents, torrent)
	}

	return torrents
}

// Remove torrent
func (tm *TorrentManager) RemoveTorrent(id string) {

	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.torrents, id)
}