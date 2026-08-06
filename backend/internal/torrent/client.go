package torrent

import (
	"net"
	"time"

	"torrent-backend/internal/models"
)

// Client manages a single peer connection.
type Client struct {
	conn net.Conn

	Peer     models.Peer
	InfoHash [20]byte
	PeerID   [20]byte

	Choked     bool // Peer is choking us.
	Interested bool // We have sent Interested.
}

// NewClient creates a client from an established peer connection.
func NewClient(peerConn *PeerConn) *Client {
	return &Client{
		conn:       peerConn.Conn,
		Peer:       peerConn.Peer,
		InfoHash:   peerConn.InfoHash,
		PeerID:     peerConn.PeerID,
		Choked:     peerConn.Choked,
		Interested: peerConn.Interested,
	}
}

// Close closes the peer connection.
func (c *Client) Close() error {

	return c.conn.Close()
}

// SetDeadline sets a deadline for all network operations.
func (c *Client) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// Read reads the next message from the peer.
func (c *Client) Read() (*Message, error) {
	return ReadMessage(c.conn)
}

// SendInterested tells the peer we want pieces.
func (c *Client) SendInterested() error {
	msg := &Message{ID: MsgInterested}

	_, err := c.conn.Write(msg.Serialize())
	if err != nil {
		return err
	}

	c.Interested = true
	return nil
}

// SendRequest requests a block from a piece.
func (c *Client) SendRequest(index, begin, length int) error {
	msg := FormatRequest(index, begin, length)

	_, err := c.conn.Write(msg.Serialize())
	return err
}
