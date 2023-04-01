// Package blockchain implements the essential data structures and functions for a blockchain.

package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db" // Name of the database file
const blocksBucket = "blocks"  // Name of the bucket in the database

// Block represents a block in the blockchain. It contains the timestamp of creation, the data to
// be stored, the hash of the previous block and the hash of the current block.
type Block struct {
	Timestamp     int64  // Time of creation of the block
	Data          []byte // Data to be stored in the block
	PrevBlockHash []byte // Hash of the previous block
	Hash          []byte // Hash of the current block
	Nonce         int    // Nonce is the number of times the hash of the block is calculated
}

// NewBlock creates and returns a pointer to a Block.
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// Serialize serializes the block into a byte slice using the Gob encoding.
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)

	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a byte slice into a block using the Gob encoding.
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)

	if err != nil {
		log.Panic(err)
	}

	return &block
}

// Blockchain represents a blockchain. It contains the tip hash to the last block in the chain and
// a pointer to the boltDB database.
type Blockchain struct {
	tip []byte   // Tip hash to the last block in the chain
	db  *bolt.DB // Pointer to the boltDB database
}

// AddBlock adds a block to the blockchain.
func (bc *Blockchain) AddBlock(data string) error {
	var lastHash []byte

	// Read the last hash from the database
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		return err
	}

	newBlock := NewBlock(data, lastHash)

	// Write the new block to the database
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
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
		return err
	}

	return nil
}

// NewGenesisBlock creates and returns a pointer to a genesis block.
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// NewBlockchain creates a new blockchain from boltDB. If the blockchain does not exist, it creates
// a genesis block and adds it to the database.
func NewBlockchain() (*Blockchain, error) {
	var tip []byte

	// Open the database
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	// Create a bucket if it does not exist
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		// If the bucket does not exist, create a genesis block and add it to the database
		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				return err
			}

			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				return err
			}

			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				return err
			}

			tip = genesis.Hash
		} else { // If the bucket exists, get the tip hash
			tip = b.Get([]byte("l"))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bc := Blockchain{tip, db}

	return &bc, nil
}

// Close closes the database connection in the blockchain.
func (bc *Blockchain) Close() {
	bc.db.Close()
}
