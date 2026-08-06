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

const peerHandshakeTimeout = 5 * time.Second

type PeerConn struct {
	Conn       net.Conn
	Choked     bool
	Interested bool
	Peer       models.Peer
	InfoHash   [20]byte
	PeerID     [20]byte
}

func Connect(peer models.Peer, infoHash [20]byte, peerID [20]byte) (*PeerConn, error) {
	addr := net.JoinHostPort(peer.IP, strconv.Itoa(int(peer.Port)))
	conn, err := net.DialTimeout("tcp", addr, peerHandshakeTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer %s: %w", addr, err)
	}

	if err := conn.SetDeadline(time.Now().Add(peerHandshakeTimeout)); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to set handshake deadline for %s: %w", addr, err)
	}
	defer func() { _ = conn.SetDeadline(time.Time{}) }()

	handshake := NewHandshake(infoHash, peerID)
	if err := writeFull(conn, handshake.Serialize()); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to send handshake to %s: %w", addr, err)
	}

	replyBuf := make([]byte, 68)
	if _, err = io.ReadFull(conn, replyBuf); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to read handshake from %s: %w", addr, err)
	}

	reply, err := Deserialize(replyBuf)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("invalid handshake from %s: %w", addr, err)
	}
	if !bytes.Equal(reply.InfoHash[:], infoHash[:]) {
		_ = conn.Close()
		return nil, fmt.Errorf("info hash mismatch from %s: expected %x got %x", addr, infoHash, reply.InfoHash)
	}

	return &PeerConn{
		Conn:       conn,
		Choked:     true,
		Interested: false,
		Peer:       peer,
		InfoHash:   infoHash,
		PeerID:     reply.PeerID,
	}, nil
}
