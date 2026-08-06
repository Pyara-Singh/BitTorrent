package torrent

import (
	"crypto/sha1"
	"testing"
)

func TestPieceWorkDoesNotDoubleCountDuplicateBlocks(t *testing.T) {
	piece := []byte("abcdefghijklmnopqrstuvwxyz")
	hash := sha1.Sum(piece)
	work := NewPieceWork(0, len(piece), hash)

	if err := work.WriteBlock(0, piece[:10]); err != nil {
		t.Fatalf("WriteBlock() failed: %v", err)
	}
	if err := work.WriteBlock(0, piece[:10]); err != nil {
		t.Fatalf("duplicate WriteBlock() failed: %v", err)
	}
	if work.Downloaded != 10 {
		t.Fatalf("Downloaded = %d, want 10", work.Downloaded)
	}

	if err := work.WriteBlock(5, piece[5:15]); err != nil {
		t.Fatalf("overlapping WriteBlock() failed: %v", err)
	}
	if work.Downloaded != 15 {
		t.Fatalf("Downloaded = %d, want 15", work.Downloaded)
	}
}
