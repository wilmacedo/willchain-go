package factory

import (
	"bytes"
	"crypto/ecdsa"
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

func (chain *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction

	spentTXRes := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txHash := hex.EncodeToString(tx.ID)

		Result:
			for resHashx, res := range tx.Results {
				if spentTXRes[txHash] != nil {
					for _, spentRes := range spentTXRes[txHash] {
						if spentRes == resHashx {
							continue Result
						}
					}
				}

				if res.IsLockWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, req := range tx.Requests {
					if req.UsesKey(pubKeyHash) {
						reqHash := hex.EncodeToString(req.ID)

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

func (chain *Blockchain) FindResTX(pubKeyHash []byte) []TXResult {
	var resTxs []TXResult
	unspentTxs := chain.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTxs {
		for _, res := range tx.Results {
			if res.IsLockWithKey(pubKeyHash) {
				resTxs = append(resTxs, res)
			}
		}
	}

	return resTxs
}

func (chain *Blockchain) FindSpendableResults(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentRes := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(pubKeyHash)
	accumlated := 0

Work:
	for _, tx := range unspentTxs {
		txHash := hex.EncodeToString(tx.ID)

		for resHashx, res := range tx.Results {
			if res.IsLockWithKey(pubKeyHash) && accumlated < amount {
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

func (chain *Blockchain) FindTransaction(ID []byte) (*Transaction, error) {
	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return tx, nil
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return nil, core.ErrNilTransaction
}

func (chain *Blockchain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, req := range tx.Requests {
		prevTX, err := chain.FindTransaction(req.ID)
		core.Handle(err)

		prevTXs[hex.EncodeToString(prevTX.ID)] = *prevTX
	}

	tx.Sign(privateKey, prevTXs)
}

func (chain *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, req := range tx.Requests {
		prevTX, err := chain.FindTransaction(req.ID)
		core.Handle(err)

		prevTXs[hex.EncodeToString(prevTX.ID)] = *prevTX
	}

	return tx.Verify(prevTXs)
}
