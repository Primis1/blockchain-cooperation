import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
)

// Repository interface defines all data access operations
type BlockchainRepository interface {
	SaveBlock(block *Block) error
	GetBlockByHash(hash []byte) (*Block, error)
	GetLastHash() ([]byte, error)
	SaveLastHash(hash []byte) error
	FindUnspentTransactions(address string) ([]Transaction, error)
	FindUTXO(address string) ([]TxOutput, error)
	FindSpendableOutputs(address string, amount int) (int, map[string][]int, error)
	Close() error
}

// Concrete repository implementation using BadgerDB
type BadgerBlockchainRepository struct {
	db *badger.DB
}

// Repository constructor
func NewBlockchainRepository(dbPath string) (*BadgerBlockchainRepository, error) {
	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &BadgerBlockchainRepository{
		db: db,
	}, nil
}

// SaveBlock persists a block to the database
func (r *BadgerBlockchainRepository) SaveBlock(block *Block) error {
	return r.db.Update(func(txn *badger.Txn) error {
		if block == nil {
			return errors.New("cannot save nil block")
		}

		serializedBlock := block.serialize()
		if err := txn.Set(block.Hash, serializedBlock); err != nil {
			return fmt.Errorf("failed to save block: %w", err)
		}

		return nil
	})
}

// GetBlockByHash retrieves a block by its hash
func (r *BadgerBlockchainRepository) GetBlockByHash(hash []byte) (*Block, error) {
	var block *Block

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(hash)
		if err == badger.ErrKeyNotFound {
			return fmt.Errorf("block not found")
		}
		if err != nil {
			return fmt.Errorf("failed to get block: %w", err)
		}

		blockData, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("failed to read block data: %w", err)
		}

		block = Deserialize(blockData)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return block, nil
}

// GetLastHash retrieves the last hash from the database
func (r *BadgerBlockchainRepository) GetLastHash() ([]byte, error) {
	var lastHash []byte

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return fmt.Errorf("failed to get last hash: %w", err)
		}

		lastHash, err = item.ValueCopy(nil)
		return err
	})

	return lastHash, err
}

// SaveLastHash saves the last hash to the database
func (r *BadgerBlockchainRepository) SaveLastHash(hash []byte) error {
	return r.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("lh"), hash)
		if err != nil {
			return fmt.Errorf("failed to save last hash: %w", err)
		}
		return nil
	})
}

// FindUnspentTransactions finds all unspent transactions for an address
func (r *BadgerBlockchainRepository) FindUnspentTransactions(address string) ([]Transaction, error) {
	var unspentTxs []Transaction
	spentTXOs := make(map[string][]int)

	lastHash, err := r.GetLastHash()
	if err != nil {
		return nil, fmt.Errorf("failed to get last hash: %w", err)
	}

	currentHash := lastHash

	for len(currentHash) > 0 {
		block, err := r.GetBlockByHash(currentHash)
		if err != nil {
			return nil, fmt.Errorf("failed to get block: %w", err)
		}

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
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

		currentHash = block.PrevHash
	}

	return unspentTxs, nil
}

// FindUTXO finds all unspent transaction outputs for an address
func (r *BadgerBlockchainRepository) FindUTXO(address string) ([]TxOutput, error) {
	var UTXOs []TxOutput

	unspentTransactions, err := r.FindUnspentTransactions(address)
	if err != nil {
		return nil, err
	}

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs, nil
}

// FindSpendableOutputs finds spendable outputs for an address up to an amount
func (r *BadgerBlockchainRepository) FindSpendableOutputs(address string, amount int) (int, map[string][]int, error) {
	unspentOuts := make(map[string][]int)
	accumulated := 0

	unspentTxs, err := r.FindUnspentTransactions(address)
	if err != nil {
		return 0, nil, err
	}

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts, nil
}

// Close closes the database connection
func (r *BadgerBlockchainRepository) Close() error {
	return r.db.Close()
}

// BlockchainService uses the repository
type BlockchainService struct {
	repo BlockchainRepository
}

// NewBlockchainService creates a new blockchain service
func NewBlockchainService(repo BlockchainRepository) *BlockchainService {
	return &BlockchainService{
		repo: repo,
	}
}

// AddBlock adds a new block to the blockchain
func (s *BlockchainService) AddBlock(transactions []*Transaction) error {
	lastHash, err := s.repo.GetLastHash()
	if err != nil {
		return fmt.Errorf("failed to get last hash: %w", err)
	}

	newBlock := CreateBlock(transactions, lastHash)

	if err := s.repo.SaveBlock(newBlock); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	if err := s.repo.SaveLastHash(newBlock.Hash); err != nil {
		return fmt.Errorf("failed to save last hash: %w", err)
	}

	return nil
}
