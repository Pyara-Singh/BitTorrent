package torrent

import (
	"bytes"
	"testing"
)

func TestHandshakeSerialization(t *testing.T) {
	infoHash := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	peerID := [20]byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	// 1. Create and Serialize
	h := NewHandshake(infoHash, peerID)
	buf := h.Serialize()

	if len(buf) != 68 {
		t.Fatalf("Expected serialized handshake length to be 68, got %d", len(buf))
	}

	// Validate protocol string prefix
	if buf[0] != 19 {
		t.Errorf("Expected first byte (protocol length) to be 19, got %d", buf[0])
	}

	pstr := string(buf[1:20])
	if pstr != "BitTorrent protocol" {
		t.Errorf("Expected protocol string 'BitTorrent protocol', got '%s'", pstr)
	}

	// 2. Deserialize and verify fields
	parsed, err := Deserialize(buf)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if parsed.Pstr != "BitTorrent protocol" {
		t.Errorf("Expected Pstr 'BitTorrent protocol', got '%s'", parsed.Pstr)
	}

	if !bytes.Equal(parsed.InfoHash[:], infoHash[:]) {
		t.Errorf("Decoded InfoHash does not match original")
	}

	if !bytes.Equal(parsed.PeerID[:], peerID[:]) {
		t.Errorf("Decoded PeerID does not match original")
	}

	// 3. Test Error boundaries
	// Too short data
	_, err = Deserialize([]byte{1, 2, 3})
	if err == nil {
		t.Errorf("Expected error when deserializing short data, got nil")
	}

	// Incorrect protocol length
	buf[0] = 18
	_, err = Deserialize(buf)
	if err == nil {
		t.Errorf("Expected error when protocol length is not 19, got nil")
	}

	// Restore and break protocol identifier
	buf[0] = 19
	buf[1] = 'A' // Corrupt "BitTorrent protocol" string
	_, err = Deserialize(buf)
	if err == nil {
		t.Errorf("Expected error when protocol identifier is modified, got nil")
	}
}
