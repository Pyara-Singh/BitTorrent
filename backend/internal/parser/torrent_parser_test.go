package parser

import (
	"encoding/hex"
	"testing"
)

func TestParseTorrentFile(t *testing.T) {
	// Parse the local test torrent file
	meta, err := ParseTorrentFile("../../testdata/ubuntu.torrent")
	if err != nil {
		t.Fatalf("Failed to parse torrent file: %v", err)
	}

	// Basic validations
	t.Logf("Parsed Name: %s", meta.Name)
	t.Logf("Tracker URL: %s", meta.Announce)
	t.Logf("Length: %d bytes", meta.Length)
	t.Logf("Piece Length: %d bytes", meta.PieceLength)
	t.Logf("Number of pieces: %d", len(meta.Pieces)/20)

	if meta.Name == "" {
		t.Errorf("Expected name to be set")
	}

	if meta.Announce == "" {
		t.Errorf("Expected tracker URL (announce) to be set")
	}

	if meta.Length <= 0 {
		t.Errorf("Expected valid file length, got %d", meta.Length)
	}

	// Verify the computed InfoHash is not zeroed out
	var zeroHash [20]byte
	if meta.InfoHash == zeroHash {
		t.Errorf("Expected InfoHash to be populated, got all zeros")
	}

	// Log computed InfoHash as hex
	hexHash := hex.EncodeToString(meta.InfoHash[:])
	t.Logf("Computed InfoHash (hex): %s", hexHash)
}
