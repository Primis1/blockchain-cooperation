package blockchain

import "github.com/dgraph-io/badger"

// to achieve functionality similar to table DB, in badger
// we can assign prefix to the field
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

func (u *UnspentTransactionSET) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := u.Blockchain.Database.Update(func(txn *badger.Txn) error {
			for _, v := range keysForDelete {
				if err := txn.Delete(v); err == nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return nil
		}

		return nil
	}
	collectSize := 100000
	u.Blockchain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					info.Info(err)
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				info.Info(err)
			}
		}
		return nil
	})

}
