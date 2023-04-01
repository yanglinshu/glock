package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

// targetBits is the number of leading zero bits required in the hash of a block.
const targetBits = 24

// ProofOfWork represents a proof-of-work.
type ProofOfWork struct {
	block  *Block   // block is the block to be mined
	target *big.Int // target is the upper bound of the hash of a block
}

// NewProofOfWork creates a new ProofOfWork with the upper bound of the hash of a block.
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	p := &ProofOfWork{b, target}

	return p
}

// IntToHex converts an integer to a hexadecimal byte array.
func IntToHex(n int64) []byte {
	return []byte(fmt.Sprintf("%x", n))
}

// prepareData returns the data to be hashed. The data is the concatenation of the fields of the
// block and the nonce.
func (p *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			p.block.PrevBlockHash,
			p.block.HashTransactions(),
			IntToHex(p.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// maxNonce is the maximum number of times the hash of the block is calculated.
const maxNonce = math.MaxInt64

// Run performs a proof-of-work.
func (p *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0 // nonce is the number of times the hash of the block is calculated

	// Calculate the hash of the block until the hash is less than the upper bound.
	for nonce < maxNonce {
		data := p.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(p.target) == -1 {
			break
		} else {
			nonce++
		}
	}

	return nonce, hash[:]

}

// Validate validates a proof-of-work.
func (p *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := p.prepareData(p.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(p.target) == -1

	return isValid
}
