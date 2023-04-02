package cli

import (
	"fmt"

	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/errors"
	"github.com/yanglinshu/glock/internal/server"
	"github.com/yanglinshu/glock/internal/transaction"
)

// sendTransaction sends coins from one address to another
func sendTransaction(from, to string, amount int, nodeID string, mineNow bool) error {
	if !transaction.ValidateAddress(from) {
		return errors.ErrInvalidAddress
	}

	if !transaction.ValidateAddress(to) {
		return errors.ErrInvalidAddress
	}

	bc, err := blockchain.NewBlockchain(nodeID)
	if err != nil {
		return err
	}
	defer bc.CloseDB()

	UTXOSet := blockchain.UTXOSet{Blockchain: bc}

	wallets, err := transaction.NewWallets(nodeID)
	if err != nil {
		return err
	}

	wallet := wallets.GetWallet(from)

	tx, err := blockchain.NewUTXOTransaction(&wallet, to, amount, &UTXOSet)
	if err != nil {
		return err
	}

	if mineNow {
		cbTx, err := transaction.NewCoinbaseTX(from, "")
		if err != nil {
			return err
		}

		txs := []*transaction.Transaction{cbTx, tx}

		newBlock, err := bc.MineBlock(txs)
		if err != nil {
			return err
		}

		UTXOSet.Update(newBlock)
	} else {
		server.SendTransaction(tx)
	}

	fmt.Println("Success!")
	return nil
}
