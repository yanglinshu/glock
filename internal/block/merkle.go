package block

import "crypto/sha256"

// MerkleTree is a struct that contains a pointer to the root node of the tree
type MerkleTree struct {
	RootNode *MerkleNode
}

// MerkleNode is a struct that contains a pointer to the left and right node of the tree
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// NewMerkleTree creates a new Merkle tree based on a sequence of data
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	// If the number of data is odd, duplicate the last data
	if len(data)&1 != 0 {
		data = append(data, data[len(data)-1])
	}

	// Create a leaf node for each piece of data and add it to the node list
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, node)
	}

	// Create a parent node for each two child nodes until there is only one node left
	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, node)
		}
		nodes = newLevel
	}

	// The root node is the only node left
	mTree := MerkleTree{&nodes[0]}
	return &mTree
}

// NewMerkleNode creates a new Merkle node based on the left and right child nodes
func NewMerkleNode(left, right *MerkleNode, data []byte) MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return mNode
}
