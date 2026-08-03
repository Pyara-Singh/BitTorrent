package torrent

import (
	"errors"
	"fmt"
)

// Handshake represents the 68-byte message exchanged when establishing a peer connection.
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

// NewHandshake creates a default BitTorrent handshake struct.
func NewHandshake(infoHash [20]byte, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// Serialize converts the Handshake struct into a raw 68-byte slice to be sent over TCP.
func (h *Handshake) Serialize() []byte {
	buf := make([]byte, 68)

	// First byte: Protocol string length (always 19)
	buf[0] = byte(len(h.Pstr))

	// Next 19 bytes: Protocol string name
	curr := 1
	curr += copy(buf[curr:], h.Pstr)

	// Next 8 bytes: Reserved bytes (default to 0)
	curr += 8

	// Next 20 bytes: InfoHash
	curr += copy(buf[curr:], h.InfoHash[:])

	// Next 20 bytes: PeerID
	copy(buf[curr:], h.PeerID[:])

	return buf
}

// Deserialize parses a raw 68-byte slice received from a peer back into a Handshake struct.
func Deserialize(data []byte) (*Handshake, error) {
	if len(data) < 68 {
		return nil, errors.New("handshake message must be at least 68 bytes")
	}

	pstrlen := int(data[0])
	if pstrlen != 19 {
		return nil, fmt.Errorf("invalid protocol length: %d", pstrlen)
	}

	pstr := string(data[1 : 1+pstrlen])
	if pstr != "BitTorrent protocol" {
		return nil, fmt.Errorf("invalid protocol identifier: %s", pstr)
	}

	var infoHash [20]byte
	copy(infoHash[:], data[28:48])

	var peerID [20]byte
	copy(peerID[:], data[48:68])

	return &Handshake{
		Pstr:     pstr,
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}
