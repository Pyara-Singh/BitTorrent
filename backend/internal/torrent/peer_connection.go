package torrent

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
	"torrent-backend/internal/models"
)

// PeerConn represents a persistent TCP connection to a peer in the swarm.
type PeerConn struct {
	Conn       net.Conn
	Choked     bool // True if the peer has choked us (default: true)
	Interested bool // True if we are interested in what the peer has (default: false)
	Peer       models.Peer
	InfoHash   [20]byte
	PeerID     [20]byte
}

// Connect establishes a TCP connection, performs the 68-byte BitTorrent handshake,
// and returns an active PeerConn if the handshake is successful and validated.
func Connect(peer models.Peer, infoHash [20]byte, peerID [20]byte) (*PeerConn, error) {
	// 1. Build host:port address
	addr := net.JoinHostPort(peer.IP, strconv.Itoa(int(peer.Port)))

	// 2. Open TCP connection with a 5-second timeout
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer %s: %w", addr, err)
	}

	// Set connection deadline for the handshake phase
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Reset deadline after handshake

	// 3. Send our handshake
	handshake := NewHandshake(infoHash, peerID)
	_, err = conn.Write(handshake.Serialize())
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send handshake to %s: %w", addr, err)
	}

	// 4. Read the peer's handshake response (exactly 68 bytes)
	replyBuf := make([]byte, 68)
	_, err = io.ReadFull(conn, replyBuf)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read handshake from %s: %w", addr, err)
	}

	// 5. Parse the peer's handshake
	reply, err := Deserialize(replyBuf)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("invalid handshake from %s: %w", addr, err)
	}

	// 6. Validate that the peer's InfoHash matches ours
	if !bytes.Equal(reply.InfoHash[:], infoHash[:]) {
		conn.Close()
		return nil, fmt.Errorf("info hash mismatch from %s (expected %x, got %x)", addr, infoHash, reply.InfoHash)
	}

	// 7. Handshake successful, return active PeerConn
	return &PeerConn{
		Conn:       conn,
		Choked:     true,       // BitTorrent spec states peer starts choked
		Interested: false,      // We start not interested
		Peer:       peer,
		InfoHash:   infoHash,
		PeerID:     reply.PeerID, // Store the peer's actual client ID
	}, nil
}
