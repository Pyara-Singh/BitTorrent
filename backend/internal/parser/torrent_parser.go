package parser

import (
	"crypto/sha1"
	"errors"
	"os"
	"torrent-backend/internal/bencode"
	"torrent-backend/internal/models"
)

func ParseTorrentFile(path string) (*models.TorrentMeta, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// First pass: extract the raw bencoded 'info' dictionary bytes
	rawInfo, err := ExtractInfoBytes(data)
	if err != nil {
		return nil, err
	}

	// Compute the 20-byte SHA-1 hash of the raw info dictionary
	infoHash := sha1.Sum(rawInfo)

	// Second pass: decode the entire file structure normally
	decoder := bencode.NewDecoder(data)

	decodedData, err := decoder.Decode()
	if err != nil {
		return nil, err
	}
	torrentDict, ok := decodedData.(map[string]any)

	if !ok {
		return nil, errors.New("invalid torrent file")
	}
	// Announce URL is a required field in torrent files
	announce, ok := torrentDict["announce"].(string)
	if !ok {
		return nil, errors.New("missing announce URL")
	}

	infoDict, ok := torrentDict["info"].(map[string]any)
	if !ok {
		return nil, errors.New("missing info dictionary")
	}

	name, ok := infoDict["name"].(string)
	if !ok {
		return nil, errors.New("missing torrent name")
	}
	pieceLength, ok := infoDict["piece length"].(int)
	if !ok {
		return nil, errors.New("missing piece length")
	}
	length, ok := infoDict["length"].(int)
	if !ok {
		return nil, errors.New("missing file length")
	}

	piecesString, ok := infoDict["pieces"].(string)
	if !ok {
		return nil, errors.New("missing pieces")
	}

	pieces := []byte(piecesString)
	meta := &models.TorrentMeta{
		Announce:    announce,
		Name:        name,
		Length:      int64(length),
		PieceLength: pieceLength,
		Pieces:      pieces,
		InfoHash:    infoHash,
	}

	return meta, nil
}

// ExtractInfoBytes parses the top-level keys of a Bencoded torrent file dictionary,
// locates the "info" key, and returns the slice of raw bytes corresponding to its value.
func ExtractInfoBytes(data []byte) ([]byte, error) {
	if len(data) == 0 || data[0] != 'd' {
		return nil, errors.New("invalid bencode dictionary")
	}

	pos := 1 // Skip initial 'd' of top-level dictionary
	for pos < len(data) && data[pos] != 'e' {
		// Read Bencoded string key
		length := 0
		for pos < len(data) && data[pos] >= '0' && data[pos] <= '9' {
			length = length*10 + int(data[pos]-'0')
			pos++
		}
		if pos >= len(data) || data[pos] != ':' {
			return nil, errors.New("invalid key string length")
		}
		pos++ // Skip ':'

		// We extract the key string
		key := string(data[pos : pos+length])
		pos += length

		// The value starts right after the key string
		valueStart := pos
		var err error
		pos, err = skip(data, pos)
		if err != nil {
			return nil, err
		}
		valueEnd := pos

		// If this is the "info" key, return the raw Bencoded slice of its value
		if key == "info" {
			return data[valueStart:valueEnd], nil
		}
	}

	return nil, errors.New("info dictionary not found in torrent file")
}

// skip scans one complete Bencoded value starting at pos, returning the new position after the value.
// It is allocation-free and optimized for scanning raw byte boundaries.
func skip(data []byte, pos int) (int, error) {
	if pos >= len(data) {
		return pos, errors.New("unexpected end of input during skip")
	}

	switch data[pos] {
	case 'i':
		// Integer: 'i' <digits> 'e'
		pos++ // Skip 'i'
		for pos < len(data) && data[pos] != 'e' {
			pos++
		}
		if pos >= len(data) {
			return pos, errors.New("unterminated bencoded integer")
		}
		pos++ // Skip 'e'
		return pos, nil

	case 'l':
		// List: 'l' <elements> 'e'
		pos++ // Skip 'l'
		for pos < len(data) && data[pos] != 'e' {
			var err error
			pos, err = skip(data, pos)
			if err != nil {
				return pos, err
			}
		}
		if pos >= len(data) {
			return pos, errors.New("unterminated bencoded list")
		}
		pos++ // Skip 'e'
		return pos, nil

	case 'd':
		// Dictionary: 'd' <key-value pairs> 'e'
		pos++ // Skip 'd'
		for pos < len(data) && data[pos] != 'e' {
			var err error
			// Skip key (must be a Bencoded string)
			pos, err = skip(data, pos)
			if err != nil {
				return pos, err
			}
			// Skip value (any Bencoded type)
			pos, err = skip(data, pos)
			if err != nil {
				return pos, err
			}
		}
		if pos >= len(data) {
			return pos, errors.New("unterminated bencoded dictionary")
		}
		pos++ // Skip 'e'
		return pos, nil

	default:
		// String: <length> ':' <bytes>
		length := 0
		for pos < len(data) && data[pos] >= '0' && data[pos] <= '9' {
			length = length*10 + int(data[pos]-'0')
			pos++
		}
		if pos >= len(data) || data[pos] != ':' {
			return pos, errors.New("invalid string bencode format")
		}
		pos++ // Skip ':'
		pos += length
		if pos > len(data) {
			return pos, errors.New("string length out of bounds")
		}
		return pos, nil
	}
}
