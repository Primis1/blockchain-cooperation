package blockchain

import (
	"blockchain/pkg/utils"

	"github.com/dgraph-io/badger"
)

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
		block = DeserializeBlock(encodedBlock)

		return err
	})
	utils.HandleErr(err)

	iter.CurrentHash = block.PrevHash

	return block
}
