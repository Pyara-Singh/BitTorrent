package torrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"sort"
)

type byteRange struct {
	begin int
	end   int
}

// PieceWork owns the in-memory state for one piece download.
type PieceWork struct {
	Index      int
	Length     int
	Hash       [20]byte
	Buffer     []byte
	Downloaded int

	received []byteRange
}

func NewPieceWork(index int, length int, hash [20]byte) *PieceWork {
	return &PieceWork{
		Index:  index,
		Length: length,
		Hash:   hash,
		Buffer: make([]byte, length),
	}
}

func (pw *PieceWork) WriteBlock(begin int, block []byte) error {
	if pw == nil {
		return errors.New("piece work is nil")
	}
	if begin < 0 {
		return errors.New("invalid block offset")
	}
	if len(block) == 0 {
		return errors.New("empty block")
	}
	if begin+len(block) > pw.Length {
		return fmt.Errorf("block offset %d with length %d exceeds piece length %d", begin, len(block), pw.Length)
	}

	copy(pw.Buffer[begin:begin+len(block)], block)
	pw.addReceivedRange(begin, begin+len(block))
	return nil
}

func (pw *PieceWork) Complete() bool {
	return pw.Downloaded == pw.Length
}

func (pw *PieceWork) Verify() error {
	if pw == nil {
		return errors.New("piece work is nil")
	}
	if !pw.Complete() {
		return fmt.Errorf("piece incomplete: downloaded %d of %d bytes", pw.Downloaded, pw.Length)
	}

	sum := sha1.Sum(pw.Buffer)
	if !bytes.Equal(sum[:], pw.Hash[:]) {
		return errors.New("piece hash verification failed")
	}
	return nil
}

func (pw *PieceWork) addReceivedRange(begin, end int) {
	pw.received = append(pw.received, byteRange{begin: begin, end: end})
	sort.Slice(pw.received, func(i, j int) bool {
		return pw.received[i].begin < pw.received[j].begin
	})

	merged := pw.received[:0]
	for _, current := range pw.received {
		if len(merged) == 0 || current.begin > merged[len(merged)-1].end {
			merged = append(merged, current)
			continue
		}
		if current.end > merged[len(merged)-1].end {
			merged[len(merged)-1].end = current.end
		}
	}

	pw.received = merged
	pw.Downloaded = 0
	for _, received := range pw.received {
		pw.Downloaded += received.end - received.begin
	}
}
