package torrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
)

// PieceWork manages the memory buffer and download progress of a single piece.
type PieceWork struct {
	Index      int
	Length     int
	Hash       [20]byte
	Buffer     []byte
	Downloaded int
}

// NewPieceWork initializes a memory buffer for a piece of the specified length.
func NewPieceWork(index int, length int, hash [20]byte) *PieceWork {
	return &PieceWork{
		Index:  index,
		Length: length,
		Hash:   hash,
		Buffer: make([]byte, length),
	}
}

// WriteBlock copies a downloaded block of data into the correct offset within the piece buffer.
func (pw *PieceWork) WriteBlock(begin int, block []byte) error {
	if begin < 0 {
		return errors.New("invalid block offset")
	}

	if len(block) == 0 {
		return errors.New("empty block")
	}

	if begin+len(block) > pw.Length {
		return fmt.Errorf(
			"block offset %d with length %d exceeds piece length %d",
			begin,
			len(block),
			pw.Length,
		)
	}

	copy(pw.Buffer[begin:], block)
	pw.Downloaded += len(block)

	return nil
}

// Verify calculates the SHA-1 hash of the assembled piece and compares it to the expected hash.
func (pw *PieceWork) Verify() error {
	sum := sha1.Sum(pw.Buffer)
	if !bytes.Equal(sum[:], pw.Hash[:]) {
		return errors.New("piece hash verification failed (data corruption)")
	}
	return nil
}
