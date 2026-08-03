package parser

import (
	"encoding/hex"
	"testing"
)

func TestParseMagnet(t *testing.T) {
	// 1. Test standard Hex magnet link (40 characters)
	hexLink := "magnet:?xt=urn:btih:dafc8c076ca2f3ed376eeae7c76a0d6be2415c45&dn=ubuntu&tr=http://tracker1.com/announce&tr=udp://tracker2.com:1337/announce"
	magnet, err := ParseMagnet(hexLink)
	if err != nil {
		t.Fatalf("ParseMagnet failed on hex link: %v", err)
	}

	expectedHash := "dafc8c076ca2f3ed376eeae7c76a0d6be2415c45"
	actualHash := hex.EncodeToString(magnet.InfoHash[:])
	if actualHash != expectedHash {
		t.Errorf("Expected InfoHash %s, got %s", expectedHash, actualHash)
	}

	if magnet.Name != "ubuntu" {
		t.Errorf("Expected Name 'ubuntu', got '%s'", magnet.Name)
	}

	if len(magnet.Trackers) != 2 || magnet.Trackers[0] != "http://tracker1.com/announce" || magnet.Trackers[1] != "udp://tracker2.com:1337/announce" {
		t.Errorf("Expected 2 trackers, got: %v", magnet.Trackers)
	}

	// 2. Test Base32 magnet link (32 characters)
	// dafc8c076ca2f3ed376eeae7c76a0d6be2415c45 in standard RFC4648 Base32 is "3l6iyb3mulz62n3o5lt4o2qnnprecxcf"
	b32Link := "magnet:?xt=urn:btih:3l6iyb3mulz62n3o5lt4o2qnnprecxcf&dn=ubuntu-base32"
	magnetB32, err := ParseMagnet(b32Link)
	if err != nil {
		t.Fatalf("ParseMagnet failed on base32 link: %v", err)
	}

	actualB32Hash := hex.EncodeToString(magnetB32.InfoHash[:])
	if actualB32Hash != expectedHash {
		t.Errorf("Expected base32 decoded InfoHash %s, got %s", expectedHash, actualB32Hash)
	}

	if magnetB32.Name != "ubuntu-base32" {
		t.Errorf("Expected Name 'ubuntu-base32', got '%s'", magnetB32.Name)
	}

	// 3. Test Invalid magnet links
	invalidLinks := []string{
		"http://example.com",                                             // Invalid scheme
		"magnet:?dn=ubuntu",                                              // Missing xt
		"magnet:?xt=urn:sha1:dafc8c076ca2f3ed376eeae7c76a0d6be2415c45",   // Invalid protocol (sha1 instead of btih)
		"magnet:?xt=urn:btih:dafc8c076ca2f3ed376eeae7c76a",               // Hash too short
	}

	for _, link := range invalidLinks {
		_, err := ParseMagnet(link)
		if err == nil {
			t.Errorf("Expected error parsing invalid magnet link, but it passed: %s", link)
		}
	}
}
