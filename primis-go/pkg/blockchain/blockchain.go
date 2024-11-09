package blockchain

import (
	"blockchain/pkg/utils"
	"encoding/hex"
	"os"

	"github.com/dgraph-io/badger"
)

type BlockchainRepository interface {
	SaveBlock(block *Block) error
	GetBlockByHash(hash []byte) *Block
	GetLastHash() ([]byte, error)
	SaveLastHash(hash []byte) error
	FindUnspentTransactionsOutputs(address string) []TXO
	FindUniqueTransaction(address string) ([]Transaction, error)
	FindSpendableOutputs(address string, amount int) (int, map[string][]int)
}

type BadgerBlockchainRepository struct {
	db *badger.DB
}

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

var (
	dbPath = os.Getenv("dbPath")
)

func NewBlockchainRepository() *BadgerBlockchainRepository {

	if !DBexists() {
		os.MkdirAll(dbPath, os.ModePerm)
	}

	// NOTE we should set define the "type",
	// NOTE over which our DB will created
	opts := badger.DefaultOptions(dbPath)
	// NOTE we put these parameters, and retrieve DB object
	db, err := badger.Open(opts)

	utils.HandleErr(err)

	return &BadgerBlockchainRepository{db: db}
}

func DBexists() bool {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false
	}

	return true
}

func (r *BadgerBlockchainRepository) SaveBlock(block *Block) error {
	return r.db.Update(func(txn *badger.Txn) error {
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

func (r *BadgerBlockchainRepository) GetBlockByHash(hash []byte) *Block {
	var block *Block

	// NOTE reading access to the block
	err := r.db.View(func(t *badger.Txn) error {
		// NOTE actual look up by hash
		item, err := t.Get(hash)

		// NOTE bunch of checks
		if err == badger.ErrKeyNotFound {
			utils.HandleErr("block is not found")
		}
		utils.HandleErr(err)

		// NOTE copy value for return in immutable way
		blockData, err := item.ValueCopy(nil)
		utils.HandleErr(err)

		block = deserialize(blockData)
		return nil
	})

	utils.HandleErr(err)
	return block
}

func (r *BadgerBlockchainRepository) GetLastHash() ([]byte, error) {
	var lhash []byte

	// NOTE the same thing like in GBBH, buy
	// NOTE retrieving "lh" hash value
	err := r.db.View(func(t *badger.Txn) error {
		item, err := t.Get([]byte("lh"))
		utils.HandleErr("failed to get last hash")

		lhash = item.KeyCopy(nil)
		return err
	})

	return lhash, err
}

func (r *BadgerBlockchainRepository) SaveLastHash(hash []byte) error {
	return r.db.Update(func(t *badger.Txn) error {
		// NOTE badger is key-value DB, so we re-assign the last
		// NOTE "lh" hash to the argument; all done in bytes - bla bla bla
		err := t.Set([]byte("lh"), hash)
		utils.HandleErr(err)

		return nil
	})
}

func (r *BadgerBlockchainRepository) FindUniqueTransaction(address string) ([]Transaction, error) {
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

func (r *BadgerBlockchainRepository) FindUnspentTransactionsOutputs(address string) []TXO {
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

func (r *BadgerBlockchainRepository) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspnedable := make(map[string][]int)
	accumulated := 0

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

type BlockchainService struct {
	Repo    BlockchainRepository
	Factory *BlockFactory // Add Factory here
}

func NewBlockchainService(repo BlockchainRepository) *BlockchainService {
	return &BlockchainService{
		Repo:    repo,
		Factory: newBlockFactory(),
	}
}

func (s *BlockchainService) AddBlock(transactions []*Transaction) error {
	lastHash, err := s.Repo.GetLastHash()
	if err != nil {
		return err
	}

	// Use Factory to create block
	newBlock := s.Factory.CreateBlock(BlockConfig{
		Transaction: transactions,
		PrevHash:    lastHash,
	})

	if err := s.Repo.SaveBlock(newBlock); err != nil {
		return err
	}

	return s.Repo.SaveLastHash(newBlock.Hash)
}
