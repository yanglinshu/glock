package blockchain

import (
	"testing"
)

func TestAddBlock(t *testing.T) {
	bc, err := NewBlockchain()
	if err != nil {
		t.Errorf("Error creating blockchain: %s", err)
	}
	defer bc.db.Close()

	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	bci := bc.Iterator()
	chain_len := 0

	for {
		block, err := bci.Next()
		if err != nil {
			t.Errorf("Error iterating blockchain: %s", err)
		}

		chain_len++

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	if chain_len != 3 {
		t.Errorf("Blockchain length is %d, expected 3", chain_len)
	}
}
