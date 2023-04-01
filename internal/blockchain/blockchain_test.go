package blockchain

import (
	"testing"
)

func TestAddBlock(t *testing.T) {
	bc := NewBlockchain()
	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	if len(bc.blocks) != 3 {
		t.Errorf("Expected length of 3, got %d", len(bc.blocks))
	} else {
		t.Log("Blockchain length is 3")
	}

	if string(bc.blocks[1].Data) != "Send 1 BTC to Ivan" {
		t.Errorf("Expected 'Send 1 BTC to Ivan', got %s", bc.blocks[1].Data)
	} else {
		t.Log("Blockchain data is 'Send 1 BTC to Ivan'")
	}

	if string(bc.blocks[2].Data) != "Send 2 more BTC to Ivan" {
		t.Errorf("Expected 'Send 2 more BTC to Ivan', got %s", bc.blocks[2].Data)
	} else {
		t.Log("Blockchain data is 'Send 2 more BTC to Ivan'")
	}

	if string(bc.blocks[1].PrevBlockHash) != string(bc.blocks[0].Hash) {
		t.Errorf("Expected %x, got %x", bc.blocks[1].Hash, bc.blocks[2].PrevBlockHash)
	} else {
		t.Log("Blockchain prev block hash is correct")
	}
}
