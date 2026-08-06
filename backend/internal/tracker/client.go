package tracker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"torrent-backend/internal/bencode"
	"torrent-backend/internal/models"
)

// GetPeers contacts the HTTP tracker and returns compact IPv4 peers.
func GetPeers(meta *models.TorrentMeta, peerID string, port int) ([]models.Peer, error) {
	trackerURL, err := buildAnnounceURL(meta, peerID, port)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(trackerURL)
	if err != nil {
		return nil, fmt.Errorf("tracker request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tracker returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tracker response: %w", err)
	}

	decodedData, err := bencode.NewDecoder(body).Decode()
	if err != nil {
		return nil, fmt.Errorf("failed to decode tracker response: %w", err)
	}

	respDict, ok := decodedData.(map[string]any)
	if !ok {
		return nil, errors.New("tracker response must be a dictionary")
	}

	if failure, ok := respDict["failure reason"].(string); ok {
		return nil, fmt.Errorf("tracker failed: %s", failure)
	}

	peersStr, ok := respDict["peers"].(string)
	if !ok {
		return nil, errors.New("missing or invalid compact peers")
	}

	peers, err := parseCompactPeers([]byte(peersStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse compact peers: %w", err)
	}
	return peers, nil
}

func buildAnnounceURL(meta *models.TorrentMeta, peerID string, port int) (string, error) {
	if meta == nil {
		return "", errors.New("torrent metadata is nil")
	}
	base, err := url.Parse(meta.Announce)
	if err != nil {
		return "", fmt.Errorf("invalid announce URL: %w", err)
	}
	if base.Scheme != "http" && base.Scheme != "https" {
		return "", fmt.Errorf("unsupported tracker scheme %q", base.Scheme)
	}

	query := base.Query()
	query.Set("peer_id", peerID)
	query.Set("port", strconv.Itoa(port))
	query.Set("uploaded", "0")
	query.Set("downloaded", "0")
	query.Set("left", strconv.FormatInt(meta.Length, 10))
	query.Set("compact", "1")
	query.Set("event", "started")

	encoded := query.Encode()
	if encoded == "" {
		base.RawQuery = "info_hash=" + urlEncodeHash(meta.InfoHash)
	} else {
		base.RawQuery = "info_hash=" + urlEncodeHash(meta.InfoHash) + "&" + encoded
	}
	return base.String(), nil
}

func urlEncodeHash(hash [20]byte) string {
	var builder strings.Builder
	builder.Grow(60)
	for _, b := range hash {
		fmt.Fprintf(&builder, "%%%02X", b)
	}
	return builder.String()
}

func parseCompactPeers(peersBin []byte) ([]models.Peer, error) {
	const peerSize = 6
	if len(peersBin)%peerSize != 0 {
		return nil, errors.New("compact peer list length must be divisible by 6")
	}

	peers := make([]models.Peer, 0, len(peersBin)/peerSize)
	for i := 0; i < len(peersBin); i += peerSize {
		peers = append(peers, models.Peer{
			IP:   net.IP(peersBin[i : i+4]).String(),
			Port: binary.BigEndian.Uint16(peersBin[i+4 : i+6]),
		})
	}
	return peers, nil
}
