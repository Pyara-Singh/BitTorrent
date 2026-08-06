package models

// FileInfo describes one file inside a torrent payload.
type FileInfo struct {
	Path   []string
	Length int64
}

type TorrentMeta struct {
	Announce     string
	AnnounceList [][]string
	Name         string
	Length       int64
	PieceLength  int
	Pieces       []byte
	Files        []FileInfo
	InfoHash     [20]byte
}
