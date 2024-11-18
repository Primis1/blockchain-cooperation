package blockchain

import (
	"blockchain/pkg/utils"
	"bytes"
	"encoding/hex"

	"github.com/dgraph-io/badger"
)

// NOTE to achieve functionality similar to table DB, in badger
// NOTE we can assign prefix to the field for data separation
var (
	utxoPrefix   = []byte("utxo-")
	prefixLength = len(utxoPrefix)
)

type (
	// gain access to the database
	UnspentTransactionSET struct {
		Blockchain *Blockchain
	}
)

func (u *UnspentTransactionSET) Reindex() {
	db := u.Blockchain.Database

	u.DeleteUnspent(utxoPrefix)

	// NOTE get all unspent transactions from the particular block
	UTXO := u.Blockchain.FindUnspentTransactionsOutputs()

	err := db.Update(func(txn *badger.Txn) error {
		for txId, outs := range UTXO {
			key, err := hex.DecodeString(txId)
			utils.HandleErr(err)

			key = append(utxoPrefix, key...)

			// PUSH it into database
			err = txn.Set(key, outs.SerializeOuts())
			utils.HandleErr(err)
		}

		return nil
	})

	utils.HandleErr(err)
}

// Add some transaction into block
func (u *UnspentTransactionSET) Update(block *Block) {
	db := u.Blockchain.Database

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					updateOutputs := TXOs{}
					// NOTE we put prefix on front of outputs of transaction
					// NOTE then we put
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID) // NOTE get value attached to index
					utils.HandleErr(err)
					var v []byte
					err = item.Value(func(val []byte) error {
						v = val
						return nil
					})

					utils.HandleErr(err)
					// deserialize bytes into TXOs
					outs := DeserializeOuts(v)

					// TODO take all the outputs which are not attached to the input we are iterating through
					for outIdx, out := range outs.Outs {
						if outIdx != in.Out {
							// NOTE Inputs contain the reference to Output value that created an input
							// NOTE Also input contains an INDEX to old transaction
							// NOTE Unspent output is not attached to the current input, but spent one does
							updateOutputs.Outs = append(updateOutputs.Outs, out)
						}
					}

					if len(updateOutputs.Outs) == 0 {
						if err := txn.Delete(inID); err != nil {
							errMsg.Error(err)
						}
					} else {
						// NOTE convert transactions into bytes
						if err := txn.Set(inID, updateOutputs.SerializeOuts()); err != nil {
							errMsg.Error(err)
						}
					}
				}
			}

			newOutputs := TXOs{}
			newOutputs.Outs = append(newOutputs.Outs, tx.Output...)

			txID := append(utxoPrefix, tx.ID...)

			if err := txn.Set(txID, newOutputs.SerializeOuts()); err != nil {
				errMsg.Error(err)
			}
		}

		return nil
	})

	utils.HandleErr(err)
}

func (u UnspentTransactionSET) FindUnspentTransactions(pubHash []byte) []TXO {
	var UTXOs []TXO

	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)

		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			var v []byte
			err := item.Value(func(val []byte) error {
				v = val
				return nil
			})
			utils.HandleErr(err)

			outs := DeserializeOuts(v)

			for _, out := range outs.Outs {
				if out.IsLockedWithKey(pubHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	utils.HandleErr(err)

	return nil
}

// TODO count how many unspent outputs are there in block
func (u *UnspentTransactionSET) CountUnspentOuts() int {
	db := u.Blockchain.Database
	counter := 0

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)

		defer it.Close()

		// Seek implemented by stack btw
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}

		return nil
	})

	utils.HandleErr(err)

	return counter
}

// TODO Task - iterate over transactions, but
// run through the database and delete all prefixed keys
func (u *UnspentTransactionSET) DeleteUnspent(prefix []byte) {
	// NOTE create a closure with modification function in
	deleteClosure := func(keyToDelete [][]byte) error {
		// NOTE enable read/write transaction
		if err := u.Blockchain.Database.Update(func(txn *badger.Txn) error {

			// iterate over list of keys, and invoke write transaction
			for _, v := range keyToDelete {
				if err := txn.Delete(v); err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	}

	collectSize := 10000 // optimal amount of keys for deletion per one function call

	u.Blockchain.Database.View(func(txn *badger.Txn) error {

		opts := badger.DefaultIteratorOptions // NOTE modifying default query parameters of badger
		opts.PrefetchValues = false           // NOTE retrieving a key without reeding the data

		it := txn.NewIterator(opts)

		defer it.Close()

		// add capacity of length of utx prefix in bytes
		keyForDelete := make([][]byte, 0, collectSize)
		keyCollectedForDelete := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			// NOTE look up for those which we want to delete, and collect 'em into slice
			key := it.Item().KeyCopy(nil)
			keyForDelete = append(keyForDelete, key)

			// NOTE indicate that the quantity to delete
			keyCollectedForDelete++

			// delete each key with function utility-closure
			// by remembering lexical environment we do not
			// iterate from the start each time
			if keyCollectedForDelete == collectSize {
				if err := deleteClosure(keyForDelete); err != nil {
					errMsg.Error(err)
				}
				keyForDelete = make([][]byte, 0, collectSize)
			}
		}

		// if we have left keys for delete - delete
		if keyCollectedForDelete > 0 {
			if err := deleteClosure(keyForDelete); err != nil {
				errMsg.Error(err)
			}
		}

		return nil
	})

}

func (u UnspentTransactionSET) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()
			var v []byte
			err := item.Value(func(val []byte) error {
				v = val
				return nil
			})
			utils.HandleErr(err)
			k = bytes.TrimPrefix(k, utxoPrefix)
			txID := hex.EncodeToString(k)
			outs := DeserializeOuts(v)

			for outIdx, out := range outs.Outs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	})
	utils.HandleErr(err)
	return accumulated, unspentOuts
}
