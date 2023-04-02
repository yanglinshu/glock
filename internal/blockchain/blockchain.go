// Package blockchain implements the essential data structures and functions for a blockchain.

package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
	"github.com/yanglinshu/glock/internal/block"
	"github.com/yanglinshu/glock/internal/errors"
	"github.com/yanglinshu/glock/internal/transaction"
)

const dbFile = "blockchain.db" // Name of the database file
const blocksBucket = "blocks"  // Name of the bucket in the database
// genesisCoinbaseData is the data in the coinbase transaction of the genesis block.
// See https://blockchain.info/tx/4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b?show_adv=true
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchain represents a blockchain. It contains the tip hash to the last block in the chain and
// a pointer to the boltDB database.
type Blockchain struct {
	tip []byte   // Tip hash to the last block in the chain
	db  *bolt.DB // Pointer to the boltDB database
}

// dbExists checks if the database file exists.
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockchain creates a new blockchain from boltDB. If the blockchain does not exist, it creates
// a genesis block and adds it to the database.
func NewBlockchain() (*Blockchain, error) {
	if !dbExists() {
		return nil, errors.ErrDBDoesNotExist
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		return nil, err
	}

	bc := Blockchain{tip, db}

	return &bc, nil
}

// createBlockchain creates a new blockchain database. It also creates a genesis block and adds it
// to the database.
func CreateBlockchain(address string) (*Blockchain, error) {
	if dbExists() {
		return nil, errors.ErrDBExists
	}

	var tip []byte

	cbtx, err := transaction.NewCoinbaseTX(address, genesisCoinbaseData)
	if err != nil {
		return nil, err
	}

	genesis := block.NewGenesisBlock(cbtx)

	// Open the database
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	// Create a bucket if it does not exist
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}

		sb, err := genesis.Serialize()
		if err != nil {
			return err
		}

		err = b.Put(genesis.Hash, sb)
		if err != nil {
			return err
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			return err
		}

		tip = genesis.Hash
		return nil
	})
	if err != nil {
		return nil, err
	}

	bc := Blockchain{tip, db}

	fmt.Printf("%x\n", tip)

	return &bc, nil
}

// FindUTXO finds and returns all unspent transaction outputs.
func (bc *Blockchain) FindUTXO() (map[string]transaction.TXOutputs, error) {
	UTXOs := make(map[string]transaction.TXOutputs)
	spentTXO := make(map[string][]int)
	bci := bc.Iterator()

	// Iterate over the blockchain
	for {
		block, err := bci.Next()
		if err != nil {
			return nil, err
		}

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// If the output has already been spent, skip it
				if spentTXO[txID] != nil {
					for _, spentOut := range spentTXO[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXOs[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXOs[txID] = outs
			}

			// If the transaction is not a coinbase transaction, add the inputs to the spentTXOs map
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXO[inTxID] = append(spentTXO[inTxID], in.Vout)
				}
			}
		}

		// If the genesis block has been reached, break out of the loop
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXOs, nil
}

// MineBlock mines a new block with the provided transactions. It adds the block to the blockchain
// and updates the database. Verify the transactions happens before the block is mined.
func (bc *Blockchain) MineBlock(transactions []*transaction.Transaction) (*block.Block, error) {
	var lastHash []byte

	// Verify the transactions
	for _, tx := range transactions {
		if ok, err := bc.VerifyTransaction(tx); err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.ErrInvalidTransaction
		}
	}

	// Get the last block's hash
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		return nil, err
	}

	newBlock := block.NewBlock(transactions, lastHash)

	// Write the new block to the database
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		sb, err := newBlock.Serialize()
		if err != nil {
			return err
		}

		err = b.Put(newBlock.Hash, sb)
		if err != nil {
			return err
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			return err
		}

		bc.tip = newBlock.Hash
		return nil
	})
	if err != nil {
		return nil, err
	}

	fmt.Printf("%x\n", newBlock.Hash)

	return newBlock, nil
}

func (bc *Blockchain) FindTransaction(ID []byte) (transaction.Transaction, error) {
	bci := bc.Iterator()

	// Iterate over the blockchain
	for {
		block, err := bci.Next()
		if err != nil {
			return transaction.Transaction{}, err
		}

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return transaction.Transaction{}, errors.ErrTransactionNotFound
}

// SignTransaction signs inputs of a Transaction.
func (bc *Blockchain) SignTransaction(tx *transaction.Transaction, privKey ecdsa.PrivateKey) error {
	prevTXs := make(map[string]transaction.Transaction)

	// Iterate over the transaction inputs
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return err
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
	return nil
}

// VerifyTransaction verifies transaction inputs.
func (bc *Blockchain) VerifyTransaction(tx *transaction.Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	prevTXs := make(map[string]transaction.Transaction)

	// Iterate over the transaction inputs
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return false, err
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs), nil
}

// NewUTXOTransaction creates a new transaction. Signing is done here.
func NewUTXOTransaction(from, to string, amount int, UTXOSet *UTXOSet) (*transaction.Transaction, error) {
	var inputs []transaction.TXInput
	var outputs []transaction.TXOutput

	wallets, err := transaction.NewWallets()
	if err != nil {
		return nil, err
	}

	wallet := wallets.GetWallet(from)
	pubKeyHash, err := transaction.HashPubKey(wallet.PublicKey)
	if err != nil {
		return nil, err
	}

	acc, validOutputs, err := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)
	if err != nil {
		return nil, err
	}

	if acc < amount {
		return nil, errors.ErrNotEnoughFunds
	}

	// Build a list of inputs
	for txID, outs := range validOutputs {
		txID, err := hex.DecodeString(txID)
		if err != nil {
			return nil, err
		}

		for _, out := range outs {
			input := transaction.TXInput{Txid: txID, Vout: out, Signature: nil, PublicKey: wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, *transaction.NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *transaction.NewTXOutput(acc-amount, from)) // a change
	}

	tx := transaction.Transaction{ID: nil, Vin: inputs, Vout: outputs}
	tx.ID = tx.Hash()
	UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx, nil
}

// Close closes the database connection in the blockchain.
func (bc *Blockchain) CloseDB() {
	bc.db.Close()
}
