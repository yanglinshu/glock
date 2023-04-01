// Package blockchain implements the essential data structures and functions for a blockchain.

package blockchain

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

// Block represents a block in the blockchain. It contains the timestamp of creation, the data to
// be stored, the hash of the previous block and the hash of the current block.
type Block struct {
	Timestamp     int64  // Time of creation of the block
	Data          []byte // Data to be stored in the block
	PrevBlockHash []byte // Hash of the previous block
	Hash          []byte // Hash of the current block
}

// SetHash calculates the hash of the block and stores it in the Hash field.
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

// NewBlock creates and returns a pointer to a Block.
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	block.SetHash()
	return block
}

// Blockchain represents a blockchain. It contains a slice of pointers to blocks.
type Blockchain struct {
	blocks []*Block // Slice of pointers to blocks
}

// AddBlock adds a block to the blockchain.
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

// NewGenesisBlock creates and returns a pointer to a genesis block.
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// NewBlockchain creates and returns a pointer to a Blockchain.
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}
