package blockchain

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func FuzzPrepareData(f *testing.F) {
	f.Fuzz(func(t *testing.T, nonce int) {
		p := NewProofOfWork(NewBlock("Genesis Block", []byte{}))
		t.Logf("nonce: %d", nonce)
		data := p.prepareData(nonce)

		if !strings.Contains(string(data), "Genesis Block") {
			t.Errorf("Expected 'Genesis Block', got %s", data)
		} else {
			t.Log("Data contains 'Genesis Block'")
		}

		if !bytes.Contains(data, []byte(fmt.Sprintf("%x", nonce))) {
			t.Errorf("Expected %x, got %s", nonce, data)
		} else {
			t.Log("Data contains nonce")
		}
	})
}

func TestRun(t *testing.T) {
	p := NewProofOfWork(NewBlock("Genesis Block", []byte{}))
	_, hash := p.Run()

	if bytes.Compare(hash, bytes.Repeat([]byte{0}, targetBits)) <= 0 {
		t.Errorf("Expected %s, got %s", strings.Repeat("0", targetBits), hash)
	} else {
		t.Log("Hash starts with targetBits")
	}
}
