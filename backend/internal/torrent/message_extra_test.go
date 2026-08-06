package torrent

import "testing"

func TestTypedMessageParsers(t *testing.T) {
	have, err := ParseHave(FormatHave(7))
	if err != nil {
		t.Fatalf("ParseHave() failed: %v", err)
	}
	if have != 7 {
		t.Fatalf("ParseHave() = %d, want 7", have)
	}

	request, err := ParseRequest(FormatRequest(1, 16, 32))
	if err != nil {
		t.Fatalf("ParseRequest() failed: %v", err)
	}
	if request.Index != 1 || request.Begin != 16 || request.Length != 32 {
		t.Fatalf("unexpected request: %+v", request)
	}

	piece, err := ParsePiece(FormatPiece(2, 4, []byte("block")))
	if err != nil {
		t.Fatalf("ParsePiece() failed: %v", err)
	}
	if piece.Index != 2 || piece.Begin != 4 || string(piece.Block) != "block" {
		t.Fatalf("unexpected piece: %+v", piece)
	}
}

func TestBitfieldHasPiece(t *testing.T) {
	msg := &Message{ID: MsgBitfield, Payload: []byte{0b10100000}}
	bitfield, err := ParseBitfield(msg, 4)
	if err != nil {
		t.Fatalf("ParseBitfield() failed: %v", err)
	}
	if !bitfield.HasPiece(0) || bitfield.HasPiece(1) || !bitfield.HasPiece(2) || bitfield.HasPiece(3) {
		t.Fatal("bitfield returned unexpected piece availability")
	}
}
