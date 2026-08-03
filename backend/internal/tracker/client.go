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
	"time"

	"torrent-backend/internal/bencode"
	"torrent-backend/internal/models"
)

// GetPeers contacts the tracker and returns a list of discovered peers.
func GetPeers(meta *models.TorrentMeta, peerID string, port int) ([]models.Peer, error) {
	// 1. Manually URL-encode the 20-byte binary InfoHash
	encodedHash := urlEncodeHash(meta.InfoHash)

	// 2. Build the query parameters
	base, err := url.Parse(meta.Announce)
	if err != nil {
		return nil, fmt.Errorf("invalid announce URL: %w", err)
	}

	params := url.Values{}
	params.Set("peer_id", peerID)
	params.Set("port", strconv.Itoa(port))
	params.Set("uploaded", "0")
	params.Set("downloaded", "0")
	params.Set("left", strconv.FormatInt(meta.Length, 10))
	params.Set("compact", "1")
	params.Set("event", "started")

	// Append the query parameters to announce URL, appending our raw info_hash
	trackerURL := fmt.Sprintf("%s?info_hash=%s&%s", base.String(), encodedHash, params.Encode())

	// 3. Send the HTTP GET request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(trackerURL)
	if err != nil {
		return nil, fmt.Errorf("tracker request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tracker returned status %d", resp.StatusCode)
	}

	// 4. Decode the Bencoded response
	// Read full response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tracker response: %w", err)
	}

	decoder := bencode.NewDecoder(body)

	decodedData, err := decoder.Decode()
	if err != nil {
		return nil, fmt.Errorf("failed to decode tracker response: %w", err)
	}

	respDict, ok := decodedData.(map[string]any)
	if !ok {
		return nil, errors.New("tracker response must be a dictionary")
	}

	// Check if the tracker returned a failure reason
	if failure, ok := respDict["failure reason"].(string); ok {
		return nil, fmt.Errorf("tracker failed: %s", failure)
	}

	// Extract compact peers string
	peersStr, ok := respDict["peers"].(string)
	if !ok {
		return nil, errors.New("missing or invalid peers in tracker response")
	}

	// 5. Parse the compact peer list
	return parseCompactPeers([]byte(peersStr))
}

// urlEncodeHash converts a 20-byte InfoHash into a raw URL-escaped string (e.g. %da%fc...)
func urlEncodeHash(hash [20]byte) string {
	result := ""

	for _, b := range hash {
		result += fmt.Sprintf("%%%02x", b)
	}

	return result
}

// parseCompactPeers splits the compact peers binary data into individual Peer structs
func parseCompactPeers(peersBin []byte) ([]models.Peer, error) {
	const peerSize = 6 // 4 bytes IP + 2 bytes Port

	if len(peersBin)%peerSize != 0 {
		return nil, errors.New("invalid compact peers binary data length")
	}

	numPeers := len(peersBin) / peerSize
	peers := make([]models.Peer, 0, numPeers)

	for i := 0; i < len(peersBin); i += peerSize {
		ip := net.IP(peersBin[i : i+4]).String()
		port := binary.BigEndian.Uint16(peersBin[i+4 : i+6])

		peers = append(peers, models.Peer{
			IP:   ip,
			Port: port,
		})
	}

	return peers, nil
}
