package bencode

import "testing"

func TestDecoderValidValues(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "negative integer", data: []byte("i-42e")},
		{name: "empty string", data: []byte("0:")},
		{name: "list", data: []byte("li1e3:twoe")},
		{name: "dictionary", data: []byte("d3:fooi1e3:bar4:spame")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewDecoder(tt.data).Decode(); err != nil {
				t.Fatalf("Decode() failed: %v", err)
			}
		})
	}
}

func TestDecoderMalformedValuesDoNotPanic(t *testing.T) {
	tests := [][]byte{
		{},
		[]byte("i12"),
		[]byte("ie"),
		[]byte("i-0e"),
		[]byte("i03e"),
		[]byte("3:ab"),
		[]byte("03:abc"),
		[]byte("l4:spam"),
		[]byte("d3:keyi1e"),
		[]byte("d3:keye"),
	}

	for _, data := range tests {
		t.Run(string(data), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Decode() panicked: %v", r)
				}
			}()

			if _, err := NewDecoder(data).Decode(); err == nil {
				t.Fatalf("expected error for malformed value %q", data)
			}
		})
	}
}
