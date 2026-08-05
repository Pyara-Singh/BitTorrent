package torrent

import (
	"bytes"
	"testing"
)

func TestMessageSerialize(t *testing.T) {
	msg := &Message{
		ID:      MsgHave,
		Payload: []byte{1, 2, 3, 4},
	}

	serialized := msg.Serialize()

	// Length = 4 bytes length prefix + 1 byte ID + 4 bytes payload = 9 bytes total
	if len(serialized) != 9 {
		t.Fatalf("Expected serialized length 9, got %d", len(serialized))
	}

	// First 4 bytes should be length of ID+Payload (which is 1 + 4 = 5)
	// Big-endian 5 is [0, 0, 0, 5]
	expectedPrefix := []byte{0, 0, 0, 5}
	if !bytes.Equal(serialized[0:4], expectedPrefix) {
		t.Errorf("Expected length prefix %v, got %v", expectedPrefix, serialized[0:4])
	}

	// 5th byte should be Message ID (MsgHave = 4)
	if serialized[4] != MsgHave {
		t.Errorf("Expected Message ID %d, got %d", MsgHave, serialized[4])
	}

	// Remaining bytes should be payload
	if !bytes.Equal(serialized[5:], []byte{1, 2, 3, 4}) {
		t.Errorf("Expected payload [1, 2, 3, 4], got %v", serialized[5:])
	}
}

func TestMessageSerializeKeepAlive(t *testing.T) {
	var msg *Message = nil
	serialized := msg.Serialize()

	if len(serialized) != 4 {
		t.Fatalf("Expected Keep-Alive length 4, got %d", len(serialized))
	}

	expected := []byte{0, 0, 0, 0}
	if !bytes.Equal(serialized, expected) {
		t.Errorf("Expected Keep-Alive to be %v, got %v", expected, serialized)
	}
}

func TestReadMessage(t *testing.T) {
	// Simulate incoming stream: [0, 0, 0, 5] (length) + [4] (ID) + [1, 2, 3, 4] (payload)
	mockStream := bytes.NewReader([]byte{0, 0, 0, 5, 4, 1, 2, 3, 4})

	msg, err := ReadMessage(mockStream)
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if msg.ID != MsgHave {
		t.Errorf("Expected ID %d, got %d", MsgHave, msg.ID)
	}

	if !bytes.Equal(msg.Payload, []byte{1, 2, 3, 4}) {
		t.Errorf("Expected Payload [1, 2, 3, 4], got %v", msg.Payload)
	}
}

func TestReadMessageSecurityLimit(t *testing.T) {
	// Simulate a malicious peer sending a length prefix of 20 MB (exceeds 16MB limit)
	// 20 MB = 20,971,520 bytes = 0x01400000 -> [1, 64, 0, 0]
	mockStream := bytes.NewReader([]byte{1, 64, 0, 0})

	_, err := ReadMessage(mockStream)
	if err == nil {
		t.Errorf("Expected ReadMessage to fail due to exceeding MaxPayloadSize, but it succeeded")
	}
}
