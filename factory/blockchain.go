package factory

import (
	"encoding/hex"
	"fmt"
	"runtime"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/wilmacedo/willchain-go/core"
	"github.com/wilmacedo/willchain-go/storage"
)

const (
	genesisData = "First transaction from genesis"
)

type Blockchain struct {
	LastHash []byte
	Database *leveldb.DB
}

type Iterator struct {
	CurrentHash []byte
	Database    *leveldb.DB
}

func InitBlockchain(address string) *Blockchain {
	var lastHash []byte

	if storage.Exists() {
		fmt.Printf("Blockchain already exists")
		runtime.Goexit()
	}

	db, err := storage.Open()
	core.Handle(err)

	if _, err := db.Get([]byte("lh"), nil); err == leveldb.ErrNotFound {
		coinbaseTx := CoinbaseTX(address, genesisData)

		genesis := Genesis(coinbaseTx)
		fmt.Println("Genesis created")

		err := db.Put(genesis.Hash, genesis.Serialize(), nil)
		core.Handle(err)

		err = db.Put([]byte("lh"), genesis.Hash, nil)
		core.Handle(err)

		lastHash = genesis.Hash
	} else {
		lastHash, err = db.Get([]byte("lh"), nil)
		core.Handle(err)
	}

	chain := &Blockchain{
		LastHash: lastHash,
		Database: db,
	}

	return chain
}

func (chain *Blockchain) AddBlock(transactions []*Transaction) {
	lastHash, err := chain.Database.Get([]byte("lh"), nil)
	core.Handle(err)

	newBlock := CreateBlock(transactions, lastHash)

	err = chain.Database.Put(newBlock.Hash, newBlock.Serialize(), nil)
	core.Handle(err)

	err = chain.Database.Put([]byte("lh"), newBlock.Hash, nil)
	core.Handle(err)

	chain.LastHash = newBlock.Hash
}

func ContinueBlockchain(address string) *Blockchain {
	if !storage.Exists() {
		fmt.Println("No existing blockchain found, need to be created!")
		runtime.Goexit()
	}

	var lastHash []byte

	db, err := storage.Open()
	core.Handle(err)

	lastHash, err = db.Get([]byte("lh"), nil)
	core.Handle(err)

	chain := &Blockchain{
		LastHash: lastHash,
		Database: db,
	}

	return chain
}

func (chain *Blockchain) Iterator() *Iterator {
	iter := &Iterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
	}

	return iter
}

func (iter *Iterator) Next() *Block {
	var block *Block

	encodedBlock, err := iter.Database.Get(iter.CurrentHash, nil)
	core.Handle(err)

	block = Deserialize(encodedBlock)

	iter.CurrentHash = block.PreviousHash

	return block
}

func (chain *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction

	spentTXRes := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txHash := hex.EncodeToString(tx.Hash)

		Result:
			for resHashx, res := range tx.Results {
				if spentTXRes[txHash] != nil {
					for _, spentRes := range spentTXRes[txHash] {
						if spentRes == resHashx {
							continue Result
						}
					}
				}

				if res.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, req := range tx.Requests {
					if req.CanUnlock(address) {
						reqHash := hex.EncodeToString(req.Hash)

						spentTXRes[reqHash] = append(spentTXRes[reqHash], req.Out)
					}
				}
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return unspentTxs
}

func (chain *Blockchain) FindResTX(address string) []TXResult {
	var resTxs []TXResult
	unspentTxs := chain.FindUnspentTransactions(address)

	for _, tx := range unspentTxs {
		for _, res := range tx.Results {
			if res.CanBeUnlocked(address) {
				resTxs = append(resTxs, res)
			}
		}
	}

	return resTxs
}

func (chain *Blockchain) FindSpendableResults(address string, amount int) (int, map[string][]int) {
	unspentRes := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(address)
	accumlated := 0

Work:
	for _, tx := range unspentTxs {
		txHash := hex.EncodeToString(tx.Hash)

		for resHashx, res := range tx.Results {
			if res.CanBeUnlocked(address) && accumlated < amount {
				accumlated += res.Value
				unspentRes[txHash] = append(unspentRes[txHash], resHashx)

				if accumlated >= amount {
					break Work
				}
			}
		}
	}

	return accumlated, unspentRes
}
