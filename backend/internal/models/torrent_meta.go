package models

type TorrentMeta struct {
	Announce    string
	Name        string
	Length      int64
	PieceLength int
	Pieces      []byte
}
