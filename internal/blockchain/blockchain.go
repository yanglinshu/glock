// Package blockchain implements the essential data structures and functions for a blockchain.

package blockchain

import (
	"time"
)

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
