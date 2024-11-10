package blockchain

import (
	"blockchain/pkg/utils"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

type (
	BlockChain struct {
		LastHash []byte
		Database *badger.DB
	}

	BlockChainIterator struct {
		CurrentHash []byte
		Database    *badger.DB
	}
)

var (
	dbPath      = os.Getenv("dbPath")
	genesisData = os.Getenv("genesisData")
)


func ContinueBlockChain(address string) *BlockChain {
	if !DBexists() {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	utils.HandleErr(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		utils.HandleErr(err)
		err = item.Value(func(val []byte) error {
			lastHash = val

			return nil
		})

		return err
	})
	utils.HandleErr(err)

	chain := BlockChain{lastHash, db}

	return &chain
}

func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if utils.DirExist(dbPath) {
		info.Info("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	utils.HandleErr(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		genesis := CreateGenesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.serialize())
		utils.HandleErr(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err

	})

	utils.HandleErr(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func DBexists() bool {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false
	}

	return true
}

func (r *BlockChain) SaveBlock(block *Block) error {
	return r.Database.Update(func(txn *badger.Txn) error {
		// NOTE if passed block is empty
		if block == nil {
			utils.HandleErr("can't save nil block")
		}

		serializedBlock := block.serialize()

		err := txn.Set(block.Hash, serializedBlock)
		utils.HandleErr(err)

		return nil
	})
}

func (r *BlockChain) GetBlockByHash(hash []byte) *Block {
	var block *Block

	// NOTE reading access to the block
	err := r.Database.View(func(t *badger.Txn) error {
		// NOTE actual look up by hash
		item, err := t.Get(hash)

		// NOTE bunch of checks
		if err == badger.ErrKeyNotFound {
			utils.HandleErr("block is not found")
		}
		// NOTE copy value for return in immutable way
		blockData, err := item.ValueCopy(nil)
		utils.HandleErr(err)

		block = deserialize(blockData)
		return nil
	})

	utils.HandleErr(err)
	return block
}

func (r *BlockChain) GetLastHash() ([]byte, error) {
	var lhash []byte

	// NOTE the same thing like in GBBH, buy
	// NOTE retrieving "lh" hash value
	err := r.Database.View(func(t *badger.Txn) error {
		item, err := t.Get([]byte("lh"))
		utils.HandleErr("failed to get last hash")

		lhash = item.KeyCopy(nil)
		return err
	})

	return lhash, err
}

func (r *BlockChain) SaveLastHash(hash []byte) error {
	return r.Database.Update(func(t *badger.Txn) error {
		// NOTE badger is key-value DB, so we re-assign the last
		// NOTE "lh" hash to the argument; all done in bytes - bla bla bla
		err := t.Set([]byte("lh"), hash)
		utils.HandleErr(err)

		return nil
	})
}

func (r *BlockChain) FindUniqueTransaction(address string) ([]Transaction, error) {
	var unspentT []Transaction

	spentMap := make(map[string][]int)
	// NOTE we get the hash, i.e the unique parameter
	// NOTE of the block
	lastHash, err := r.GetLastHash()
	utils.HandleErr(err)

	currentHash := lastHash
	for len(currentHash) > 0 {
		// NOTE we retrieve the block by its hash
		block := r.GetBlockByHash(currentHash)

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			// NOTE label to terminate loop
		Output:
			// NOTE we iterate over list of transaction of user
			// NOTE and then store 'em in the hashmap
			for outIndex, out := range tx.Output {
				if spentMap[txID] != nil {
					for _, spentOut := range spentMap[txID] {
						if spentOut == outIndex {
							continue Output
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspentT = append(unspentT, *tx)
				}
			}

			// NOTE after we append found transaction
			// NOTE from hashmap to slice

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.ID)
						spentMap[inTxID] = append(spentMap[inTxID], in.Out)
					}
				}
			}
		}

		currentHash = block.PrevHash
	}

	return unspentT, nil
}

func (r *BlockChain) FindUnspentTransactionsOutputs(address string) []TXO {
	var UTXs []TXO

	// NOTE look up for transaction by hash
	unspentTransactions, err := r.FindUniqueTransaction(address)
	utils.HandleErr(err)

	for _, t := range unspentTransactions {
		for _, out := range t.Output {

			if out.CanBeUnlocked(address) {
				UTXs = append(UTXs, out)
			}
		}
	}

	return UTXs
}

func (r *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	accumulated := 0
	unspnedable := make(map[string][]int)

	unspentT, err := r.FindUniqueTransaction(address)
	utils.HandleErr(err)

	// label
Work:
	for _, t := range unspentT {
		txID := hex.EncodeToString(t.ID)

		for outIndex, out := range t.Output {
			// NOTE compare the does amount valid for transfer or not
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				unspnedable[txID] = append(unspnedable[txID], outIndex)

				// NOTE if available amount is smaller, then kick him out!
				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspnedable
}

func (s *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	err := s.Database.View(func(t *badger.Txn) error {
		item, err := t.Get([]byte("lh"))
		utils.HandleErr(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})

		return err
	})

	utils.HandleErr(err)

	newBlock := CreateBlock(transactions, lastHash)

	err = s.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.serialize())
		utils.HandleErr(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		s.LastHash = newBlock.Hash
		return err
	})

	utils.HandleErr(err)
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		utils.HandleErr(err)
		var encodedBlock []byte
		err = item.Value(func(val []byte) error {
			encodedBlock = val
			return nil
		})
		block = deserialize(encodedBlock)

		return err
	})
	utils.HandleErr(err)

	iter.CurrentHash = block.PrevHash

	return block
}

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Output {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTxs
}
