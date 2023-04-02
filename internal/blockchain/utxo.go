package blockchain

import (
	"encoding/hex"

	"github.com/boltdb/bolt"
	"github.com/yanglinshu/glock/internal/block"
	"github.com/yanglinshu/glock/internal/transaction"
)

// utxoBucket is the name of the bucket used to store the UTXO set
const utxoBucket = "chainstate"

// UTXOSet represents a set of UTXOs
type UTXOSet struct {
	Blockchain *Blockchain
}

// Reindex rebuilds the UTXO set when the blockchain is updated
func (u *UTXOSet) Reindex() error {
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket)
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			return err
		}

		_, err = tx.CreateBucket(bucketName)
		return err
	})
	if err != nil {
		return err
	}

	UTXO, err := u.Blockchain.FindUTXO()
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}

			sl, err := outs.Serialize()
			if err != nil {
				return err
			}

			err = b.Put(key, sl)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (u *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0

	db := u.Blockchain.db
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs, err := transaction.DeserializeOutputs(v)
			if err != nil {
				return err
			}

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})
	if err != nil {
		return 0, nil, err
	}

	return accumulated, unspentOutputs, nil
}

// FindUTXO finds and returns all unspent transaction outputs
func (u *UTXOSet) FindUTXO(pubKeyHash []byte) ([]transaction.TXOutput, error) {
	var UTXOs []transaction.TXOutput

	db := u.Blockchain.db
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs, err := transaction.DeserializeOutputs(v)
			if err != nil {
				return err
			}

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return UTXOs, nil
}

// CountTransactions returns the number of transactions in the UTXO set
func (u UTXOSet) CountTransactions() (int, error) {
	db := u.Blockchain.db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return counter, nil
}

// Update updates the UTXO set with transactions from the Block
func (u *UTXOSet) Update(block *block.Block) error {
	db := u.Blockchain.db
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					updatedOuts := transaction.TXOutputs{}
					outsBytes := b.Get(in.Txid)
					outs, err := transaction.DeserializeOutputs(outsBytes)
					if err != nil {
						return err
					}

					for outIdx, out := range outs.Outputs {
						if outIdx != in.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(in.Txid)
						if err != nil {
							return err
						}
					} else {
						sl, err := updatedOuts.Serialize()
						if err != nil {
							return err
						}

						err = b.Put(in.Txid, sl)
						if err != nil {
							return err
						}
					}
				}
			}

			newOutputs := transaction.TXOutputs{}
			newOutputs.Outputs = append(newOutputs.Outputs, tx.Vout...)

			sl, err := newOutputs.Serialize()
			if err != nil {
				return err
			}

			err = b.Put(tx.ID, sl)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return err
}
