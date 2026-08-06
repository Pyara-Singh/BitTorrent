package torrent

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"testing"

	"torrent-backend/internal/models"
)

const (
	testHandshakeLength = 68
)

var (
	testInfoHash = [20]byte{
		1, 2, 3, 4, 5,
		6, 7, 8, 9, 10,
		11, 12, 13, 14, 15,
		16, 17, 18, 19, 20,
	}

	testPeerID = [20]byte{
		7, 7, 7, 7, 7,
		7, 7, 7, 7, 7,
		7, 7, 7, 7, 7,
		7, 7, 7, 7, 7,
	}
)

// startMockPeer starts a local peer that serves a single piece.
func startMockPeer(t *testing.T, piece []byte) (string, int, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock peer: %v", err)
	}

	go runMockPeer(t, listener, piece)

	host, portStr, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		listener.Close()
		t.Fatalf("failed to parse listener address: %v", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		listener.Close()
		t.Fatalf("failed to parse listener port: %v", err)
	}

	return host, port, func() {
		listener.Close()
	}
}

// runMockPeer accepts one connection and behaves like a minimal BitTorrent peer.
func runMockPeer(t *testing.T, listener net.Listener, piece []byte) {
	t.Helper()

	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	if err := handleHandshake(conn); err != nil {
		return
	}

	if err := handleInterested(conn); err != nil {
		return
	}

	if err := sendUnchoke(conn); err != nil {
		return
	}

	if err := servePieceRequests(conn, piece); err != nil && err != io.EOF {
		return
	}
}

// Reads the client's handshake and sends one back.
func handleHandshake(conn net.Conn) error {

	buffer := make([]byte, testHandshakeLength)

	if _, err := io.ReadFull(conn, buffer); err != nil {
		return err
	}

	reply := NewHandshake(testInfoHash, testPeerID)

	_, err := conn.Write(reply.Serialize())
	return err
}

// Waits for the Interested message.
func handleInterested(conn net.Conn) error {

	msg, err := ReadMessage(conn)
	if err != nil {
		return err
	}

	if msg == nil {
		return io.ErrUnexpectedEOF
	}

	if msg.ID != MsgInterested {
		return io.ErrUnexpectedEOF
	}

	return nil
}

// Sends an Unchoke message.
func sendUnchoke(conn net.Conn) error {

	msg := &Message{
		ID: MsgUnchoke,
	}

	_, err := conn.Write(msg.Serialize())
	return err
}

// Responds to every Request message with the requested block.
func servePieceRequests(conn net.Conn, piece []byte) error {

	for {

		msg, err := ReadMessage(conn)
		if err != nil {
			return err
		}

		if msg == nil {
			continue
		}

		if msg.ID != MsgRequest {
			continue
		}

		index := binary.BigEndian.Uint32(msg.Payload[0:4])
		begin := binary.BigEndian.Uint32(msg.Payload[4:8])
		length := binary.BigEndian.Uint32(msg.Payload[8:12])

		block := piece[begin : begin+length]

		payload := make([]byte, 8+len(block))

		binary.BigEndian.PutUint32(payload[0:4], index)
		binary.BigEndian.PutUint32(payload[4:8], begin)

		copy(payload[8:], block)

		response := &Message{
			ID:      MsgPiece,
			Payload: payload,
		}

		if _, err := conn.Write(response.Serialize()); err != nil {
			return err
		}
	}
}

func TestAttemptDownloadPiece(t *testing.T) {

	// Mock piece returned by the peer.
	mockPiece := []byte(
		"This is a mock piece of data that we want to download. It is exactly 70 bytes long!",
	)

	expectedHash := sha1.Sum(mockPiece)

	// Start a local BitTorrent peer.
	host, port, cleanup := startMockPeer(t, mockPiece)
	defer cleanup()

	peer := models.Peer{
		IP:   host,
		Port: uint16(port),
	}

	clientPeerID := [20]byte{1}

	// Establish the peer connection.
	peerConn, err := Connect(peer, testInfoHash, clientPeerID)
	if err != nil {
		t.Fatalf("failed to connect to mock peer: %v", err)
	}

	client := NewClient(peerConn)
	defer client.Close()

	// Allocate memory for the piece.
	work := NewPieceWork(
		0,
		len(mockPiece),
		expectedHash,
	)

	// Download the complete piece.
	data, err := AttemptDownloadPiece(client, work)
	if err != nil {
		t.Fatalf("piece download failed: %v", err)
	}

	// Verify downloaded bytes.
	if len(data) != len(mockPiece) {
		t.Fatalf(
			"unexpected piece length: got %d want %d",
			len(data),
			len(mockPiece),
		)
	}

	if !bytes.Equal(data, mockPiece) {
		t.Fatal("downloaded piece does not match expected data")
	}

	// Verify the piece state.
	if work.Downloaded != work.Length {
		t.Fatalf(
			"piece download incomplete: downloaded %d of %d bytes",
			work.Downloaded,
			work.Length,
		)
	}
}
