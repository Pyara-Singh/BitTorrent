package torrent

import "torrent-backend/internal/models"

type Repository interface {
	AddTorrent(models.Torrent)
	GetTorrent(string) (models.Torrent, bool)
	GetAllTorrents() []models.Torrent
	RemoveTorrent(string)
}