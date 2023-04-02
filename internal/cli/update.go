package cli

import (
	"fmt"

	"github.com/yanglinshu/glock/internal/blockchain"
)

// updateUTXO rebuilds the UTXO set
func updateUTXO(nodeID string) error {
	bc, err := blockchain.NewBlockchain(nodeID)
	if err != nil {
		return err
	}
	defer bc.CloseDB()

	UTXOSet := blockchain.UTXOSet{Blockchain: bc}
	UTXOSet.Reindex()

	count, err := UTXOSet.CountTransactions()
	if err != nil {
		return err
	}

	fmt.Printf("Done! There are now %d transactions in the UTXO set.\n", count)
	return nil
}
