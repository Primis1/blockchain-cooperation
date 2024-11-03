package blockchain

import (
	"blockchain/pkg/utils"

	"github.com/dgraph-io/badger"
)

// FacadeType acts as a mediator, providing a simplified interface to blockchain functionality.
type FacadeType struct {
	chain    *BlockChain
	iterator *BlockChainIterator
}

const (
	dbPath = "./tmp/blocks"
)

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// Global instance of FacadeType for easy access
var Facade = &FacadeType{}

// NOTE During writing i yet again found myself, that blockchain is emphasize the "linked list like"
// NOTE abstractions. Linked-List of Peer's Databases - sounds wonderful
type BlockChain struct {
	// NOTE not really get why we use pointer
	// NOTE to the array, when arrays
	// NOTE sent-by-reference by default
	LastHash []byte
	Database *badger.DB
}

type Block struct {
	Hash         []byte
	Data         []byte
	PreviousHash []byte
	Nonce        int
}

// AddBlock appends a new block to the blockchain.
func (f *FacadeType) AddBlock(data string) {
	var lastHash []byte

	err := f.chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		utils.HandleErr(err)

		err = item.Value(func(val []byte) error {
			lastHash = append(lastHash, val...)
			return nil
		})
		return err
	})
	utils.HandleErr(err)

	newBlock := f.CreateBlock(data, lastHash)

	err = f.chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())

		utils.HandleErr(err)

		err = txn.Set([]byte("lh"), newBlock.Hash)

		f.chain.LastHash = newBlock.Hash

		return err
	})

	utils.HandleErr(err)
}

func (f *FacadeType) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{f.chain.LastHash, f.chain.Database}

	f.iterator = iter

	return f.iterator
}

func (f *BlockChainIterator) Next() *Block {
	var block *Block

	err := f.Database.View(func(x *badger.Txn) error {
		item, err := x.Get(f.CurrentHash)

		utils.HandleErr(err)

		err = item.Value(func(t []byte) error {
			block = block.Deserialize(t)
			return nil
		})
		return err
	})
	utils.HandleErr(err)
	f.CurrentHash = block.PreviousHash

	return block
}

// InitBlockChain initializes the blockchain with the genesis block.
func (f *FacadeType) InitBlockChain() *BlockChain {
	var lastHash []byte

	// NOTE we specify where our file have to stored
	// NOTE DefaultOptions is a struct btw
	opts := badger.Options{}
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	utils.HandleErr(err)

	// NOTE Update() allows us to read/write to DB
	// NOTE View() only read from db
	err = db.Update(func(txn *badger.Txn) error {
		// NOTE 1. Check the existence of blockchain
		// NOTE 1. If present create a new BC instance in memory
		// NOTE 2. Get the last hash from the disk DB,
		// NOTE    push it into instance of created BC
		// NOTE 3.
		// NOTE 2. If not:
		// NOTE 1. Create a Genesis block
		// NOTE 2. Store that in DB
		// NOTE 3. Take hash from DB

		if _, err := txn.Get([]byte("ln")); err == badger.ErrKeyNotFound {
			info.Info("No existing blockchain found")
			genesis := f.GenesisBlock()
			info.Info("Genesis proved")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			utils.HandleErr(err)
			err = txn.Set([]byte("lh"), genesis.Hash)
			lastHash = genesis.Hash

			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			utils.HandleErr(err)
			err = item.Value(func(val []byte) error {
				lastHash = append([]byte{}, val...) // Copy the value to lastHash
				return nil
			})
			utils.HandleErr(err)

			return err
		}
	})

	utils.HandleErr(err)

	blockchain := BlockChain{lastHash, db}

	return &blockchain
}
