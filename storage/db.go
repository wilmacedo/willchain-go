package storage

import (
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	dbPath = "./tmp/blocks"
	dbFile = "./tmp/blocks/CURRENT"
)

func Open() (*leveldb.DB, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Exists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
