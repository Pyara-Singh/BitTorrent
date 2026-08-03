package torrent

import (
	"bytes"
	"io"
	"net"
	"strconv"
	"testing"
	"torrent-backend/internal/models"
)

func TestConnectHandshakeSuccess(t *testing.T) {
	infoHash := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	clientPeerID := [20]byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	mockPeerID := [20]byte{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}

	// 1. Start a local mock TCP server on a random available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock TCP listener: %v", err)
	}
	defer listener.Close()

	// Parse out host and port
	host, portStr, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portStr)

	// 2. Accept and handle the connection in the background (goroutine)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read client's handshake
		buf := make([]byte, 68)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Errorf("Mock server failed to read client handshake: %v", err)
			return
		}

		// Validate client's info hash in the background
		clientHandshake, err := Deserialize(buf)
		if err != nil || !bytes.Equal(clientHandshake.InfoHash[:], infoHash[:]) {
			t.Errorf("Mock server received invalid handshake or mismatching InfoHash")
			return
		}

		// Respond with mock peer's handshake
		responseHandshake := NewHandshake(infoHash, mockPeerID)
		conn.Write(responseHandshake.Serialize())
	}()

	// 3. Connect our client to the local TCP server
	peer := models.Peer{
		IP:   host,
		Port: uint16(port),
	}

	peerConn, err := Connect(peer, infoHash, clientPeerID)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer peerConn.Conn.Close()

	// 4. Validate peer connection state
	if !peerConn.Choked {
		t.Errorf("Expected peer to start Choked=true")
	}

	if peerConn.Interested {
		t.Errorf("Expected peer to start Interested=false")
	}

	if !bytes.Equal(peerConn.PeerID[:], mockPeerID[:]) {
		t.Errorf("Expected PeerID to match mock peer, got %v", peerConn.PeerID)
	}
}

func TestConnectHandshakeMismatch(t *testing.T) {
	infoHash1 := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	infoHash2 := [20]byte{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}
	peerID := [20]byte{0}

	// 1. Start a local mock TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock TCP listener: %v", err)
	}
	defer listener.Close()

	host, portStr, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.Atoi(portStr)

	// 2. Accept and respond with a mismatching InfoHash
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read client's handshake
		buf := make([]byte, 68)
		io.ReadFull(conn, buf)

		// Respond with a DIFFERENT info hash (infoHash2)
		responseHandshake := NewHandshake(infoHash2, peerID)
		conn.Write(responseHandshake.Serialize())
	}()

	// 3. Connect client expecting infoHash1
	peer := models.Peer{
		IP:   host,
		Port: uint16(port),
	}

	_, err = Connect(peer, infoHash1, peerID)
	if err == nil {
		t.Fatalf("Expected Connect to fail due to info hash mismatch, but it succeeded")
	}
}
