package blockchain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMerkleTree(t *testing.T) {
	data := [][]byte{
		[]byte("node1"),
		[]byte("node2"),
		[]byte("node3"),
		[]byte("node4"),
		[]byte("node5"),
		[]byte("node6"),
		[]byte("node7"),
	}

	md1 := NewMerkleNode(nil, nil, data[0])
	md2 := NewMerkleNode(nil, nil, data[1])
	md3 := NewMerkleNode(nil, nil, data[2])
	md4 := NewMerkleNode(nil, nil, data[3])
	md5 := NewMerkleNode(nil, nil, data[4])
	md6 := NewMerkleNode(nil, nil, data[5])
	md7 := NewMerkleNode(nil, nil, data[6])
	md8 := NewMerkleNode(nil, nil, data[6])

	// 2 - tier

	md9 := NewMerkleNode(md1, md2, nil)
	md10 := NewMerkleNode(md3, md4, nil)
	md11 := NewMerkleNode(md5, md6, nil)
	md12 := NewMerkleNode(md7, md8, nil)

	// 3 - tier

	md13 := NewMerkleNode(md9, md10, nil)
	md14 := NewMerkleNode(md11, md12, nil)

	// root
	mn15 := NewMerkleNode(md13, md14, nil)

	root := fmt.Sprintf("%x", mn15.Data)
	tree := NewMerkleTree(data)

	assert.Equal(t, root, fmt.Sprintf("%x", tree.RootNode.Data), "Merkle node root has is equal")
}
