package parser

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"torrent-backend/internal/models"
)

// ParseMagnet parses a magnet link string and returns a models.Magnet struct.
func ParseMagnet(uri string) (*models.Magnet, error) {
	// Parse the string as a URL
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	// Validate the custom URI scheme
	if u.Scheme != "magnet" {
		return nil, errors.New("invalid scheme, must be magnet:")
	}

	// Extract the query parameters
	queryParams := u.Query()

	// Extract the exact topic (xt)
	xt := queryParams.Get("xt")
	if xt == "" {
		return nil, errors.New("missing xt parameter")
	}

	// xt must begin with the BitTorrent info hash prefix
	if !strings.HasPrefix(xt, "urn:btih:") {
		return nil, errors.New("invalid xt parameter, must start with urn:btih:")
	}

	hashStr := strings.TrimPrefix(xt, "urn:btih:")
	var infoHash [20]byte

	// An info hash in a magnet link can be encoded in Hex (40 chars) or Base32 (32 chars)
	switch len(hashStr) {
	case 40:
		hashBytes, err := hex.DecodeString(hashStr)
		if err != nil {
			return nil, errors.New("invalid hex info hash")
		}
		copy(infoHash[:], hashBytes)

	case 32:
		// Base32 encoding (e.g. standard A-Z, 2-7, case-insensitive, no padding)
		hashBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(hashStr))
		if err != nil {
			return nil, errors.New("invalid base32 info hash")
		}
		copy(infoHash[:], hashBytes)

	default:
		return nil, errors.New("invalid info hash length, must be 40 (hex) or 32 (base32) characters")
	}

	// Extract the optional display name (dn)
	name := queryParams.Get("dn")

	// Extract the tracker URLs (tr) - can be multiple tr parameters
	trackers := queryParams["tr"]

	return &models.Magnet{
		InfoHash: infoHash,
		Name:     name,
		Trackers: trackers,
	}, nil
}
