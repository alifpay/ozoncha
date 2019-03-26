package db

import (
	"github.com/dgraph-io/badger"
)

var bCon *badger.DB

//Connect opens with DefaultOptions the Badger database
func Connect(path string) (err error) {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	bCon, err = badger.Open(opts)
	return
}

//IsPanic ErrKeyNotFound then panic
func IsPanic(err error) bool {
	if err == badger.ErrKeyNotFound {
		return false
	}
	return true
}

//Close closes the Badger database
func Close() {
	bCon.Close()
}

//Get value by key
func Get(key string) (rtn []byte, err error) {
	err = bCon.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		rtn, err = item.ValueCopy(nil)
		return err
	})
	return
}

//Set save value
func Set(k, v []byte) (err error) {
	err = bCon.Update(func(txn *badger.Txn) error {
		return txn.Set(k, v)
	})
	return
}

//Del delete from db
func Del(key string) (err error) {
	err = bCon.View(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	return
}

//GetAllBy Prefix scans
func GetAllBy(prefix string) (str []string, err error) {
	bCon.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(prefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			//k := item.Key()
			err = item.Value(func(v []byte) error {
				//fmt.Printf("key=%s, value=%s\n", k, v)
				str = append(str, string(v))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}
