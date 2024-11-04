package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

// Block represents a block in the blockchain
type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

// BlockFactory handles the creation of different types of blocks
type BlockFactory struct {
	powFactory *ProofOfWorkFactory
}

// NewBlockFactory creates a new instance of BlockFactory
func NewBlockFactory() *BlockFactory {
	return &BlockFactory{
		powFactory: &ProofOfWorkFactory{},
	}
}

// BlockConfig contains configuration for block creation
type BlockConfig struct {
	Transactions []*Transaction
	PrevHash     []byte
	Difficulty   int
}

// CreateBlock creates a new regular block with provided configuration
func (f *BlockFactory) CreateBlock(config BlockConfig) *Block {
	block := &Block{
		Hash:         []byte{},
		Transactions: config.Transactions,
		PrevHash:     config.PrevHash,
		Nonce:        0,
	}

	// Create and run proof of work
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// CreateGenesisBlock creates a genesis block with a coinbase transaction
func (f *BlockFactory) CreateGenesisBlock(coinbaseTx *Transaction) *Block {
	return f.CreateBlock(BlockConfig{
		Transactions: []*Transaction{coinbaseTx},
		PrevHash:     []byte{},
	})
}

// ProofOfWorkFactory handles creation of proof of work instances
type ProofOfWorkFactory struct{}

// HashTransactions creates a hash of all transactions in the block
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// Serialize converts a block into bytes
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	if err := encoder.Encode(b); err != nil {
		log.Panic("Failed to serialize block:", err)
	}

	return res.Bytes()
}

// DeserializeBlock converts bytes back into a Block
func DeserializeBlock(data []byte) (*Block, error) {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&block); err != nil {
		return nil, err
	}

	return &block, nil
}

// Example usage:
/*
func main() {
    factory := NewBlockFactory()
    
    // Create genesis block
    coinbaseTx := CoinbaseTx("Genesis", "First Transaction")
    genesisBlock := factory.CreateGenesisBlock(coinbaseTx)
    
    // Create regular block
    transactions := []*Transaction{...}
    regularBlock := factory.CreateBlock(BlockConfig{
        Transactions: transactions,
        PrevHash:    genesisBlock.Hash,
    })
}
*/