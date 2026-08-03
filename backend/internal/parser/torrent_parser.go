package parser

import (
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
	}

	return meta, nil
}
