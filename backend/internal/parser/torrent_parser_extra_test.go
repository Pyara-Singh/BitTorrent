package parser

import "testing"

func TestParseTorrentMultiFile(t *testing.T) {
	data := []byte("d8:announce14:http://tracker4:infod5:filesld6:lengthi5e4:pathl5:a.txteed6:lengthi6e4:pathl3:dir5:b.bineee4:name4:root12:piece lengthi4e6:pieces20:abcdefghijklmnopqrstee")

	meta, err := ParseTorrent(data)
	if err != nil {
		t.Fatalf("ParseTorrent() failed: %v", err)
	}
	if meta.Length != 11 {
		t.Fatalf("Length = %d, want 11", meta.Length)
	}
	if len(meta.Files) != 2 {
		t.Fatalf("Files = %d, want 2", len(meta.Files))
	}
	if got := meta.Files[1].Path[0]; got != "dir" {
		t.Fatalf("second file path[0] = %q, want dir", got)
	}
}

func TestExtractInfoBytesRejectsMalformedTopLevel(t *testing.T) {
	inputs := [][]byte{
		[]byte(""),
		[]byte("l4:spame"),
		[]byte("d4:info999:namee"),
	}

	for _, input := range inputs {
		if _, err := ExtractInfoBytes(input); err == nil {
			t.Fatalf("expected error for %q", input)
		}
	}
}
