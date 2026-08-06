package models

type Magnet struct {
	InfoHash [20]byte // The unique 20-byte identity of the torrent
	Name     string   // The display name of the torrent (optional)
	Trackers []string // Redundant matchmaker servers to find peers
}
