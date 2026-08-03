package tracker

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"torrent-backend/internal/models"
)

func TestGetPeers(t *testing.T) {
	// 1. Prepare compact peer binary data (12 bytes for 2 peers)
	// Peer 1: 192.168.1.100:6881 -> bytes: [192, 168, 1, 100, 26, 225] (since 6881 in hex is 0x1ae1)
	// Peer 2: 127.0.0.1:8080     -> bytes: [127, 0, 0, 1, 31, 144]   (since 8080 in hex is 0x1f90)
	peersBin := []byte{
		192, 168, 1, 100, 0x1a, 0xe1,
		127, 0, 0, 1, 0x1f, 0x90,
	}

	// Construct Bencoded mock response:
	// d8:intervali900e5:peers12:<peersBin>e
	responsePrefix := "d8:intervali900e5:peers12:"
	responseSuffix := "e"
	mockResponseBody := append([]byte(responsePrefix), peersBin...)
	mockResponseBody = append(mockResponseBody, []byte(responseSuffix)...)

	// 2. Start local HTTP test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters sent by client
		infoHash := r.URL.Query().Get("info_hash")
		if infoHash == "" {
			t.Errorf("Expected info_hash parameter in request")
		}

		peerID := r.URL.Query().Get("peer_id")
		if peerID != "-AG0001-123456789012" {
			t.Errorf("Unexpected peer_id: %s", peerID)
		}

		compact := r.URL.Query().Get("compact")
		if compact != "1" {
			t.Errorf("Expected compact=1 query param")
		}

		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseBody)
	}))
	defer server.Close()

	// 3. Run client pointing to the local mock server
	mockMeta := &models.TorrentMeta{
		Announce:    server.URL,
		Name:        "mock-torrent",
		Length:      1024 * 1024,
		PieceLength: 262144,
		Pieces:      make([]byte, 20),
		InfoHash:    [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
	}

	peers, err := GetPeers(mockMeta, "-AG0001-123456789012", 6881)
	if err != nil {
		t.Fatalf("GetPeers failed: %v", err)
	}

	// 4. Assertions on parsed peers
	if len(peers) != 2 {
		t.Fatalf("Expected 2 peers, got %d", len(peers))
	}

	if peers[0].IP != "192.168.1.100" || peers[0].Port != 6881 {
		t.Errorf("Unexpected peer 0: %+v", peers[0])
	}

	if peers[1].IP != "127.0.0.1" || peers[1].Port != 8080 {
		t.Errorf("Unexpected peer 1: %+v", peers[1])
	}
}
