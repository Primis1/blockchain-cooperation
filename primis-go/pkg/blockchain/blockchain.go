package blockchain

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
	Nonce        int
}

// FacadeType acts as a mediator, providing a simplified interface to blockchain functionality.
type FacadeType struct {
	chain *BlockChain
}

// Global instance of FacadeType for easy access
var Facade = &FacadeType{}

func (f *FacadeType) GetBlocks() []*Block {
	return f.chain.Blocks
}

// CreateBlock generates a new block with provided data and previous hash.
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

// AddBlock appends a new block to the blockchain.
func (f *FacadeType) AddBlock(data string) {
	prevBlock := f.chain.Blocks[len(f.chain.Blocks)-1]
	newBlock := f.CreateBlock(data, prevBlock.Hash)
	f.chain.Blocks = append(f.chain.Blocks, newBlock)
}

// GenesisBlock creates the initial block in the blockchain.
func (f *FacadeType) GenesisBlock() *Block {
	return f.CreateBlock("Genesis", []byte{})
}

// InitBlockChain initializes the blockchain with the genesis block.
func (f *FacadeType) InitBlockChain() *FacadeType {
	f.chain = &BlockChain{Blocks: []*Block{f.GenesisBlock()}}
	return f
}
