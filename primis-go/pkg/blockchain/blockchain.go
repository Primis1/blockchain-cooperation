package blockchain

import (
	"blockchain/pkg/sha"
	"bytes"
	// "crypto/sha256"
)

// NOTE During writing i yet again found myself, that blockchain is emphasize the "linked list like"
// NOTE abstractions. Linked-List of Peer's Databases - sounds wonderful

type BlockChain struct {
	// NOTE not really get why we use pointer
	// NOTE to the array, when arrays
	// NOTE sent-by-reference by default
	Blocks []*Block
}

type Block struct {
	Hash         []byte
	Data         []byte
	PreviousHash []byte
}

// FacadeType acts as a mediator, providing a simplified interface to blockchain functionality.
type FacadeType struct {
	Chain *BlockChain
}

// Global instance of FacadeType for easy access
var Facade = &FacadeType{}

// DeriveHash computes the hash of the block based on its data and previous hash.
func (b *Block) DeriveHash() {
	// NOTE Hash the created block, we hash Data as well as Previous Hash, and then
	// NOTE Fill the Block.Hash with new generated one
	info := bytes.Join([][]byte{b.Data, b.PreviousHash}, []byte{})
	// my SHA-256 - yay!
	hash := sha.ComputeHash(info)
	b.Hash = hash[:]
}

// CreateBlock generates a new block with provided data and previous hash.
func (f *FacadeType) CreateBlock(data string, prevHash []byte) *Block {
	// NOTE We generate the block/node for blockchain
	// NOTE then we hash the block
	block := &Block{[]byte{}, []byte(data), prevHash}
	block.DeriveHash()
	return block
}

// AddBlock appends a new block to the blockchain.
func (f *FacadeType) AddBlock(data string) {
	prevBlock := f.Chain.Blocks[len(f.Chain.Blocks)-1]
	newBlock := f.CreateBlock(data, prevBlock.Hash)
	f.Chain.Blocks = append(f.Chain.Blocks, newBlock)
}

// GenesisBlock creates the initial block in the blockchain.
func (f *FacadeType) GenesisBlock() *Block {
	return f.CreateBlock("Genesis", []byte{})
}

// InitBlockChain initializes the blockchain with the genesis block.
func (f *FacadeType) InitBlockChain() *BlockChain {
	f.Chain = &BlockChain{Blocks: []*Block{f.GenesisBlock()}}
	return f.Chain
}
