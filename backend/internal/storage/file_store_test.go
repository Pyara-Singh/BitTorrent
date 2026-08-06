package storage

import (
	"os"
	"path/filepath"
	"testing"

	"torrent-backend/internal/models"
)

func TestFileStoreWritePieceAcrossFiles(t *testing.T) {
	root := t.TempDir()
	meta := &models.TorrentMeta{
		PieceLength: 4,
		Files: []models.FileInfo{
			{Path: []string{"a.txt"}, Length: 3},
			{Path: []string{"nested", "b.txt"}, Length: 5},
		},
	}

	store, err := NewFileStore(root, meta)
	if err != nil {
		t.Fatalf("NewFileStore() failed: %v", err)
	}
	if err := store.WritePiece(0, []byte("abcd")); err != nil {
		t.Fatalf("WritePiece() failed: %v", err)
	}

	a, err := os.ReadFile(filepath.Join(root, "a.txt"))
	if err != nil {
		t.Fatalf("failed to read a.txt: %v", err)
	}
	if string(a) != "abc" {
		t.Fatalf("a.txt = %q, want abc", a)
	}

	b, err := os.ReadFile(filepath.Join(root, "nested", "b.txt"))
	if err != nil {
		t.Fatalf("failed to read b.txt: %v", err)
	}
	if string(b) != "d" {
		t.Fatalf("b.txt = %q, want d", b)
	}
}
