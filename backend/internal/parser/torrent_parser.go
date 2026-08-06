package parser

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"math"
	"os"
	"torrent-backend/internal/bencode"
	"torrent-backend/internal/models"
)

func ParseTorrentFile(path string) (*models.TorrentMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read torrent file: %w", err)
	}

	return ParseTorrent(data)
}

func ParseTorrent(data []byte) (*models.TorrentMeta, error) {
	rawInfo, err := ExtractInfoBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract info dictionary: %w", err)
	}

	infoHash := sha1.Sum(rawInfo)
	decodedData, err := bencode.NewDecoder(data).Decode()
	if err != nil {
		return nil, fmt.Errorf("failed to decode torrent file: %w", err)
	}

	torrentDict, ok := decodedData.(map[string]any)
	if !ok {
		return nil, errors.New("torrent root must be a dictionary")
	}

	announce, ok := torrentDict["announce"].(string)
	if !ok || announce == "" {
		return nil, errors.New("missing announce URL")
	}

	infoDict, ok := torrentDict["info"].(map[string]any)
	if !ok {
		return nil, errors.New("missing info dictionary")
	}

	name, err := requiredString(infoDict, "name")
	if err != nil {
		return nil, err
	}

	pieceLength, err := requiredPositiveInt(infoDict, "piece length")
	if err != nil {
		return nil, err
	}

	piecesString, err := requiredString(infoDict, "pieces")
	if err != nil {
		return nil, err
	}
	pieces := []byte(piecesString)
	if len(pieces) == 0 || len(pieces)%sha1.Size != 0 {
		return nil, fmt.Errorf("pieces field length %d is not a positive multiple of %d", len(pieces), sha1.Size)
	}

	files, totalLength, err := parseFiles(infoDict, name)
	if err != nil {
		return nil, err
	}

	return &models.TorrentMeta{
		Announce:     announce,
		AnnounceList: parseAnnounceList(torrentDict),
		Name:         name,
		Length:       totalLength,
		PieceLength:  pieceLength,
		Pieces:       pieces,
		Files:        files,
		InfoHash:     infoHash,
	}, nil
}

func requiredString(dict map[string]any, key string) (string, error) {
	value, ok := dict[key].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("missing %s", key)
	}
	return value, nil
}

func requiredPositiveInt(dict map[string]any, key string) (int, error) {
	value, ok := dict[key].(int)
	if !ok || value <= 0 {
		return 0, fmt.Errorf("missing or invalid %s", key)
	}
	return value, nil
}

func parseFiles(infoDict map[string]any, name string) ([]models.FileInfo, int64, error) {
	if length, ok := infoDict["length"].(int); ok {
		if length <= 0 {
			return nil, 0, errors.New("file length must be positive")
		}
		file := models.FileInfo{Path: []string{name}, Length: int64(length)}
		return []models.FileInfo{file}, int64(length), nil
	}

	rawFiles, ok := infoDict["files"].([]any)
	if !ok || len(rawFiles) == 0 {
		return nil, 0, errors.New("missing file length or files list")
	}

	files := make([]models.FileInfo, 0, len(rawFiles))
	var total int64
	for i, rawFile := range rawFiles {
		fileDict, ok := rawFile.(map[string]any)
		if !ok {
			return nil, 0, fmt.Errorf("file entry %d must be a dictionary", i)
		}

		length, ok := fileDict["length"].(int)
		if !ok || length < 0 {
			return nil, 0, fmt.Errorf("file entry %d has invalid length", i)
		}

		path, err := parseFilePath(fileDict, i)
		if err != nil {
			return nil, 0, err
		}

		if int64(length) > math.MaxInt64-total {
			return nil, 0, errors.New("torrent total length overflows int64")
		}
		total += int64(length)
		files = append(files, models.FileInfo{Path: path, Length: int64(length)})
	}

	if total <= 0 {
		return nil, 0, errors.New("torrent payload length must be positive")
	}
	return files, total, nil
}

func parseFilePath(fileDict map[string]any, index int) ([]string, error) {
	rawPath, ok := fileDict["path"].([]any)
	if !ok || len(rawPath) == 0 {
		return nil, fmt.Errorf("file entry %d has missing path", index)
	}

	path := make([]string, 0, len(rawPath))
	for j, rawPart := range rawPath {
		part, ok := rawPart.(string)
		if !ok || part == "" {
			return nil, fmt.Errorf("file entry %d has invalid path element %d", index, j)
		}
		path = append(path, part)
	}
	return path, nil
}

func parseAnnounceList(torrentDict map[string]any) [][]string {
	rawList, ok := torrentDict["announce-list"].([]any)
	if !ok {
		return nil
	}

	announceList := make([][]string, 0, len(rawList))
	for _, rawTier := range rawList {
		rawTrackers, ok := rawTier.([]any)
		if !ok {
			continue
		}

		tier := make([]string, 0, len(rawTrackers))
		for _, rawTracker := range rawTrackers {
			tracker, ok := rawTracker.(string)
			if ok && tracker != "" {
				tier = append(tier, tracker)
			}
		}
		if len(tier) > 0 {
			announceList = append(announceList, tier)
		}
	}
	return announceList
}

// ExtractInfoBytes returns the exact raw bencoded value of the top-level info key.
func ExtractInfoBytes(data []byte) ([]byte, error) {
	if len(data) == 0 || data[0] != 'd' {
		return nil, errors.New("torrent root must be a bencoded dictionary")
	}

	pos := 1
	for {
		if pos >= len(data) {
			return nil, errors.New("unterminated top-level dictionary")
		}
		if data[pos] == 'e' {
			return nil, errors.New("info dictionary not found")
		}

		key, next, err := scanString(data, pos)
		if err != nil {
			return nil, fmt.Errorf("invalid top-level key: %w", err)
		}
		pos = next

		valueStart := pos
		valueEnd, err := skip(data, pos)
		if err != nil {
			return nil, fmt.Errorf("invalid value for top-level key %q: %w", key, err)
		}
		if key == "info" {
			return data[valueStart:valueEnd], nil
		}
		pos = valueEnd
	}
}

func scanString(data []byte, pos int) (string, int, error) {
	if pos >= len(data) || data[pos] < '0' || data[pos] > '9' {
		return "", pos, errors.New("expected string")
	}

	length := 0
	start := pos
	for pos < len(data) && data[pos] != ':' {
		if data[pos] < '0' || data[pos] > '9' {
			return "", pos, errors.New("invalid string length")
		}
		if length > (math.MaxInt-int(data[pos]-'0'))/10 {
			return "", pos, errors.New("string length overflows int")
		}
		length = length*10 + int(data[pos]-'0')
		pos++
	}
	if pos >= len(data) {
		return "", pos, errors.New("unterminated string length")
	}
	if data[start] == '0' && pos-start > 1 {
		return "", pos, errors.New("string length has leading zero")
	}

	pos++
	if pos+length > len(data) {
		return "", pos, errors.New("string length exceeds input")
	}
	return string(data[pos : pos+length]), pos + length, nil
}

func skip(data []byte, pos int) (int, error) {
	if pos >= len(data) {
		return pos, errors.New("unexpected end of input")
	}

	switch data[pos] {
	case 'i':
		return skipInteger(data, pos)
	case 'l':
		return skipList(data, pos)
	case 'd':
		return skipDictionary(data, pos)
	default:
		_, next, err := scanString(data, pos)
		return next, err
	}
}

func skipInteger(data []byte, pos int) (int, error) {
	pos++
	start := pos
	for pos < len(data) && data[pos] != 'e' {
		if data[pos] != '-' && (data[pos] < '0' || data[pos] > '9') {
			return pos, errors.New("invalid integer digit")
		}
		pos++
	}
	if pos >= len(data) {
		return pos, errors.New("unterminated integer")
	}
	if start == pos {
		return pos, errors.New("empty integer")
	}
	return pos + 1, nil
}

func skipList(data []byte, pos int) (int, error) {
	pos++
	for {
		if pos >= len(data) {
			return pos, errors.New("unterminated list")
		}
		if data[pos] == 'e' {
			return pos + 1, nil
		}
		next, err := skip(data, pos)
		if err != nil {
			return pos, err
		}
		pos = next
	}
}

func skipDictionary(data []byte, pos int) (int, error) {
	pos++
	for {
		if pos >= len(data) {
			return pos, errors.New("unterminated dictionary")
		}
		if data[pos] == 'e' {
			return pos + 1, nil
		}

		next, err := skip(data, pos)
		if err != nil {
			return pos, err
		}
		pos = next

		next, err = skip(data, pos)
		if err != nil {
			return pos, err
		}
		pos = next
	}
}
