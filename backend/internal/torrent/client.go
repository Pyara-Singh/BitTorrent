package torrent

import (
	"fmt"
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

	Choked     bool
	Interested bool
}

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

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *Client) Read() (*Message, error) {
	msg, err := ReadMessage(c.conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read peer message: %w", err)
	}
	return msg, nil
}

func (c *Client) SendInterested() error {
	if err := c.writeMessage(&Message{ID: MsgInterested}); err != nil {
		return err
	}
	c.Interested = true
	return nil
}

func (c *Client) SendRequest(index, begin, length int) error {
	return c.writeMessage(FormatRequest(index, begin, length))
}

func (c *Client) writeMessage(msg *Message) error {
	if err := writeFull(c.conn, msg.Serialize()); err != nil {
		return fmt.Errorf("failed to write peer message: %w", err)
	}
	return nil
}
