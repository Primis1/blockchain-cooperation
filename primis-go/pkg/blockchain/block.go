package blockchain

import (
	"blockchain/pkg/utils"
	"bytes"
	"encoding/gob"
)


// NOTE CreateBlock generates a new block with provided data and previous hash.
func (f *FacadeType) CreateBlock(data string, prevHash []byte) *Block {
	// NOTE We generate the block/node for blockchain
	// NOTE then we hash the block
	block := &Block{[]byte{}, []byte(data), prevHash, 0}
	pow := NewProof(block)

	nonce, hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce

	return block
}

// NOTE GenesisBlock creates the initial block in the blockchain.
func (f *FacadeType) GenesisBlock() *Block {
	return f.CreateBlock("Genesis", []byte{})
}

// NOTE We should declare serializer for default GO DB
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)
	utils.HandleErr(err)
	return res.Bytes()
}

func (b *Block) Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)
	utils.HandleErr(err)
	return &block
}
