package storage

// PieceWriter persists verified torrent pieces.
type PieceWriter interface {
	WritePiece(index int, data []byte) error
}
