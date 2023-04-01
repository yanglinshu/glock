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
	block  *Block
	target *big.Int
}

// NewProofOfWork creates a new ProofOfWork with the upper bound of the hash of a block.
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	p := &ProofOfWork{b, target}

	return p
}

// prepareData returns the data to be hashed. The data is the concatenation of the fields of the
// block and the nonce.
func (p *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			p.block.PrevBlockHash,
			p.block.Data,
			[]byte(fmt.Sprintf("%x", p.block.Timestamp)),
			[]byte(fmt.Sprintf("%x", targetBits)),
			[]byte(fmt.Sprintf("%x", nonce)),
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
