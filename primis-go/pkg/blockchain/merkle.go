// NOTE Merkle tree is a binary tree, where leafs are data,
// NOTE and internal nodes are combinations of hashes
// NOTE for that data. It kinda goes from bottom to top.
// NOTE We can access all hashes of leafs, if we have root node

// NOTE "SPV" - simplified payment verification system
// NOTE because each block must contain copies of all other block
// NOTE we have "lite" bitcoin node, which instead downloading entire blockchain,
// NOTE lite node contains reference to single "full" block,
// NOTE which contains all hashes. So tons of lite node are references on
// NOTE a single block.
// *I use words "nodes" and "blocks" as synonyms to avoid tautology
// *look at image at image folder

package blockchain

import "blockchain/pkg/sha"

type (
	MerkleTree struct {
		RootNode *MerkleNode
	}

	MerkleNode struct {
		Left  *MerkleNode
		Right *MerkleNode
		Data  []byte
	}
)

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		hash := sha.ComputeHash(data)
		node.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha.ComputeHash(prevHashes)
		node.Data = hash[:]
	}

	node.Left = left
	node.Right = right

	return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	// Check the "evenness" of nodes. if not even,
	// node duplicate
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, info := range data {
		// NOTE first leafs represent transactions,
		// NOTE therefore does not contain any children
		node := NewMerkleNode(nil, nil, info)
		nodes = append(nodes, *node)
	}
	// NOTE fill up the tree
	for i := 0; i < len(data)/2; i++ {
		var level []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			// NOTE iteratively create nodes
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			// add 'em to the tree
			level = append(level, *node)
		}
		nodes = level
	}

	// pass the root the new tree, and return reference 
	return &MerkleTree{&nodes[0]}
}
