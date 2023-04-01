package blockchain

import "github.com/boltdb/bolt"

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte   // Current hash of the block
	db          *bolt.DB // Database
}

// Iterator returns a BlockchainIterator from the tip of the chain
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}

// Next returns the next block starting from the tip
func (i *BlockchainIterator) Next() (*Block, error) {
	var block *Block

	// Read the block from the database
	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)

		var err error = nil
		block, err = DeserializeBlock(encodedBlock)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	i.currentHash = block.PrevBlockHash

	return block, nil
}
