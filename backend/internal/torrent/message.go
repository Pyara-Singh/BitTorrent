package torrent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

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

const MaxPayloadSize = 16 * 1024 * 1024

type Message struct {
	ID      byte
	Payload []byte
}

type BlockRequest struct {
	Index  int
	Begin  int
	Length int
}

type PieceBlock struct {
	Index int
	Begin int
	Block []byte
}

func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	length := uint32(1 + len(m.Payload))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = m.ID
	copy(buf[5:], m.Payload)
	return buf
}

func ReadMessage(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lengthBuf); err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return nil, nil
	}
	if length > MaxPayloadSize {
		return nil, fmt.Errorf("message length %d exceeds maximum %d", length, MaxPayloadSize)
	}

	messageBuf := make([]byte, length)
	if _, err := io.ReadFull(r, messageBuf); err != nil {
		return nil, fmt.Errorf("failed to read message body: %w", err)
	}

	return &Message{ID: messageBuf[0], Payload: messageBuf[1:]}, nil
}

func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

func FormatCancel(index, begin, length int) *Message {
	msg := FormatRequest(index, begin, length)
	msg.ID = MsgCancel
	return msg
}

func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}

func FormatPiece(index, begin int, block []byte) *Message {
	payload := make([]byte, 8+len(block))
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	copy(payload[8:], block)
	return &Message{ID: MsgPiece, Payload: payload}
}

func ParseHave(msg *Message) (int, error) {
	if msg == nil || msg.ID != MsgHave {
		return 0, errors.New("expected have message")
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("have payload length must be 4, got %d", len(msg.Payload))
	}
	return int(binary.BigEndian.Uint32(msg.Payload)), nil
}

func ParseRequest(msg *Message) (BlockRequest, error) {
	if msg == nil || msg.ID != MsgRequest {
		return BlockRequest{}, errors.New("expected request message")
	}
	return parseBlockRequestPayload(msg.Payload)
}

func ParseCancel(msg *Message) (BlockRequest, error) {
	if msg == nil || msg.ID != MsgCancel {
		return BlockRequest{}, errors.New("expected cancel message")
	}
	return parseBlockRequestPayload(msg.Payload)
}

func ParsePiece(msg *Message) (PieceBlock, error) {
	if msg == nil || msg.ID != MsgPiece {
		return PieceBlock{}, errors.New("expected piece message")
	}
	if len(msg.Payload) < 8 {
		return PieceBlock{}, fmt.Errorf("piece payload length must be at least 8, got %d", len(msg.Payload))
	}
	return PieceBlock{
		Index: int(binary.BigEndian.Uint32(msg.Payload[0:4])),
		Begin: int(binary.BigEndian.Uint32(msg.Payload[4:8])),
		Block: msg.Payload[8:],
	}, nil
}

func ParseBitfield(msg *Message, pieceCount int) (*Bitfield, error) {
	if msg == nil || msg.ID != MsgBitfield {
		return nil, errors.New("expected bitfield message")
	}
	if pieceCount < 0 {
		return nil, errors.New("piece count cannot be negative")
	}
	neededBytes := (pieceCount + 7) / 8
	if len(msg.Payload) < neededBytes {
		return nil, fmt.Errorf("bitfield too short: got %d need %d", len(msg.Payload), neededBytes)
	}
	return &Bitfield{bytes: append([]byte(nil), msg.Payload...), pieceCount: pieceCount}, nil
}

func parseBlockRequestPayload(payload []byte) (BlockRequest, error) {
	if len(payload) != 12 {
		return BlockRequest{}, fmt.Errorf("request payload length must be 12, got %d", len(payload))
	}
	request := BlockRequest{
		Index:  int(binary.BigEndian.Uint32(payload[0:4])),
		Begin:  int(binary.BigEndian.Uint32(payload[4:8])),
		Length: int(binary.BigEndian.Uint32(payload[8:12])),
	}
	if request.Length <= 0 {
		return BlockRequest{}, errors.New("request length must be positive")
	}
	return request, nil
}

// Bitfield stores which pieces a peer claims to have.
type Bitfield struct {
	bytes      []byte
	pieceCount int
}

func (b *Bitfield) HasPiece(index int) bool {
	if b == nil || index < 0 || index >= b.pieceCount {
		return false
	}
	byteIndex := index / 8
	bitOffset := uint(index % 8)
	return b.bytes[byteIndex]&(1<<(7-bitOffset)) != 0
}

func (b *Bitfield) Len() int {
	if b == nil {
		return 0
	}
	return b.pieceCount
}
