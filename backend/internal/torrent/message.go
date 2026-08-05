package torrent

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Message ID constants representing the standard BitTorrent peer wire protocol messages.
const (
	MsgChoke         byte = 0
	MsgUnchoke       byte = 1
	MsgInterested    byte = 2
	MsgNotInterested byte = 3
	MsgHave          byte = 4
	MsgBitfield      byte = 5
	MsgRequest       byte = 6
	MsgPiece         byte = 7
	MsgCancel        byte = 8
)

// MaxPayloadSize sets a safe upper limit on incoming message sizes to prevent
// memory exhaustion attacks from malicious peers sending fake large length prefixes.
// 16MB is a generous upper bound (standard piece blocks are usually 16KB).
const MaxPayloadSize = 16 * 1024 * 1024

// Message represents a parsed peer wire protocol message.
type Message struct {
	ID      byte
	Payload []byte
}

// Serialize encodes a Message into a standard binary format to be sent over the wire.
// Format: <length prefix (4 bytes)><message ID (1 byte)><payload>
func (m *Message) Serialize() []byte {
	if m == nil {
		// A nil message represents a Keep-Alive message (length prefix 0)
		return make([]byte, 4)
	}

	// Total length = 1 byte (ID) + length of payload
	length := uint32(1 + len(m.Payload))

	// Create a buffer large enough to hold length prefix, ID, and payload
	buf := make([]byte, 4+length)

	// Write length prefix (4 bytes, Big-Endian)
	binary.BigEndian.PutUint32(buf[0:4], length)

	// Write Message ID (1 byte)
	buf[4] = m.ID

	// Write Payload
	copy(buf[5:], m.Payload)

	return buf
}

// ReadMessage safely reads a Message from a TCP connection stream.
func ReadMessage(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)

	// Read the 4-byte length prefix
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	// Keep-Alive message (length 0)
	if length == 0 {
		return nil, nil
	}

	// Security check to prevent memory exhaustion
	if length > MaxPayloadSize {
		return nil, fmt.Errorf("message length %d exceeds safe maximum %d", length, MaxPayloadSize)
	}

	// Allocate a buffer for the rest of the message (ID + Payload)
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read message body: %w", err)
	}

	// Parse ID and Payload
	return &Message{
		ID:      messageBuf[0],
		Payload: messageBuf[1:],
	}, nil
}

// FormatRequest creates a Request message to ask the peer for a block of data.
func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

// FormatHave creates a Have message to tell the peer we finished downloading a piece.
func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}

// FormatPiece creates a Piece message to deliver a block of data to the peer.
func FormatPiece(index, begin int, block []byte) *Message {
	payload := make([]byte, 8+len(block))
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	copy(payload[8:], block)
	return &Message{ID: MsgPiece, Payload: payload}
}
