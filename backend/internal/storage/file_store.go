package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"torrent-backend/internal/models"
)

// FileStore maps torrent piece offsets onto one or more files on disk.
type FileStore struct {
	root        string
	pieceLength int64
	files       []fileEntry
	totalLength int64
}

type fileEntry struct {
	path   string
	length int64
	begin  int64
	end    int64
}

func NewFileStore(root string, meta *models.TorrentMeta) (*FileStore, error) {
	if meta == nil {
		return nil, errors.New("torrent metadata is nil")
	}
	if meta.PieceLength <= 0 {
		return nil, errors.New("piece length must be positive")
	}
	if len(meta.Files) == 0 {
		return nil, errors.New("torrent has no files")
	}

	entries := make([]fileEntry, 0, len(meta.Files))
	var offset int64
	for _, file := range meta.Files {
		if file.Length < 0 {
			return nil, fmt.Errorf("file %v has negative length", file.Path)
		}
		if len(file.Path) == 0 {
			return nil, errors.New("file path is empty")
		}

		path := filepath.Join(append([]string{root}, file.Path...)...)
		entries = append(entries, fileEntry{
			path:   path,
			length: file.Length,
			begin:  offset,
			end:    offset + file.Length,
		})
		offset += file.Length
	}

	return &FileStore{
		root:        root,
		pieceLength: int64(meta.PieceLength),
		files:       entries,
		totalLength: offset,
	}, nil
}

func (s *FileStore) WritePiece(index int, data []byte) error {
	if s == nil {
		return errors.New("file store is nil")
	}
	if index < 0 {
		return errors.New("piece index cannot be negative")
	}
	if len(data) == 0 {
		return errors.New("piece data is empty")
	}

	pieceBegin := int64(index) * s.pieceLength
	pieceEnd := pieceBegin + int64(len(data))
	if pieceBegin >= s.totalLength || pieceEnd > s.totalLength {
		return fmt.Errorf("piece %d range [%d,%d) exceeds payload length %d", index, pieceBegin, pieceEnd, s.totalLength)
	}

	for _, file := range s.files {
		begin := maxInt64(pieceBegin, file.begin)
		end := minInt64(pieceEnd, file.end)
		if begin >= end {
			continue
		}

		chunk := data[begin-pieceBegin : end-pieceBegin]
		if err := writeAt(file.path, begin-file.begin, chunk); err != nil {
			return fmt.Errorf("failed to write piece %d to %s: %w", index, file.path, err)
		}
	}
	return nil
}

func writeAt(path string, offset int64, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	written := 0
	for written < len(data) {
		n, err := file.WriteAt(data[written:], offset+int64(written))
		written += n
		if err != nil {
			return fmt.Errorf("short write after %d of %d bytes: %w", written, len(data), err)
		}
		if n == 0 {
			return errors.New("short write")
		}
	}
	return nil
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
