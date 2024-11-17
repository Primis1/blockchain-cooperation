// NOTE the bigger our blockchain become, the more inefficient adding new blocks/transactions become
// NOTE optimization - iterate only over specific component in the block, such as unspent transactions

package blockchain

import (
	"blockchain/pkg/utils"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

type (
	Blockchain struct {
		LastHash []byte
		Database *badger.DB
	}

	BlockchainIterator struct {
		CurrentHash []byte
		Database    *badger.DB
	}
)

var (
	dbPath      = os.Getenv("dbPath")
	genesisData = os.Getenv("genesisData")
)

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		// NOTE if we reach the last block
		if len(block.PrevHash) == 0 {
			break
		}

		// NOTE if we found the transaction within a block, which id
		// NOTE matches the settled ID - win-win
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}
	}

	return Transaction{}, nil
}

func (b *Blockchain) SignTransaction(t *Transaction, privateKey ecdsa.PrivateKey) {
	prevTs := make(map[string]Transaction)

	for _, in := range t.Inputs {
		prevT, err := b.FindTransaction(in.ID)
		utils.HandleErr(err)

		prevTs[hex.EncodeToString(prevT.ID)] = prevT
	}

	t.Sign(privateKey, prevTs)
}
func (b *Blockchain) VerifyTransaction(t *Transaction) bool {
	if t.IsCoinbase() {
		return true
	}

	prevTs := make(map[string]Transaction)

	// NOTE collect all transactions into hash-table
	for _, in := range t.Inputs {
		prevT, err := b.FindTransaction(in.ID)
		utils.HandleErr(err)

		prevTs[hex.EncodeToString(prevT.ID)] = prevT
	}

	// NOTE send hash-table for verification
	return t.Verify(prevTs)
}

func ContinueBlockchain(address string) *Blockchain {
	if !utils.DirExist(dbPath) {
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

	chain := Blockchain{lastHash, db}

	return &chain
}

func InitBlockchain(address string) *Blockchain {
	var lastHash []byte

	if utils.DirExist(dbPath) {
		info.Info("Blockchain already exists")
		runtime.Goexit()
	} 
info.Info(dbPath)
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

	blockchain := Blockchain{lastHash, db}
	return &blockchain
}

func (r *Blockchain) SaveBlock(block *Block) error {
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

func (r *Blockchain) GetBlockByHash(hash []byte) *Block {
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

		block = deserializeBlock(blockData)
		return nil
	})

	utils.HandleErr(err)
	return block
}

func (r *Blockchain) GetLastHash() ([]byte, error) {
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

func (r *Blockchain) SaveLastHash(hash []byte) error {
	return r.Database.Update(func(t *badger.Txn) error {
		// NOTE badger is key-value DB, so we re-assign the last
		// NOTE "lh" hash to the argument; all done in bytes - bla bla bla
		err := t.Set([]byte("lh"), hash)
		utils.HandleErr(err)

		return nil
	})
}

func (r *Blockchain) FindUniqueTransaction(address []byte) ([]Transaction, error) {
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
				if out.IsLockedWithKey(address) {
					unspentT = append(unspentT, *tx)
				}
			}

			// NOTE after we append found transaction
			// NOTE from hashmap to slice

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.UserKey(address) {
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

func (s *Blockchain) AddBlock(transactions []*Transaction) *Block {
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
	return newBlock
}

func (chain *Blockchain) Iterator() *BlockchainIterator {
	iter := &BlockchainIterator{chain.LastHash, chain.Database}

	return iter
}

func (iter *BlockchainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		utils.HandleErr(err)
		var encodedBlock []byte
		err = item.Value(func(val []byte) error {
			encodedBlock = val
			return nil
		})
		block = deserializeBlock(encodedBlock)

		return err
	})
	utils.HandleErr(err)

	iter.CurrentHash = block.PrevHash

	return block
}

// TODO we should make a method which will iterate over blockchain transactions
// TODO and find all unspent outputs from these transactions

func (chain *Blockchain) FindUnspentTransactionsOutputs() map[string]TXOs {
	var UTXO = make(map[string]TXOs)

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

				// NOTE we put collected outs'
				outs := UTXO[txID]
				outs.Outs = append(outs.Outs, out)
				UTXO[txID] = outs // NOTE put it back into map

			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					// NOTE translate into string
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)

				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}
