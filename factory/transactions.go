package factory

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/wilmacedo/willchain-go/core"
	wData "github.com/wilmacedo/willchain-go/data"
	"github.com/wilmacedo/willchain-go/utils"
	"github.com/wilmacedo/willchain-go/wallet"
)

type Transaction struct {
	ID       []byte
	Requests []TXRequest
	Results  []TXResult
}

type TXRequest struct {
	ID        []byte
	Out       int
	Signature []byte
	PubKey    []byte
}

type TXResult struct {
	Value      int
	PubKeyHash []byte
}

type TXResults struct {
	Results []TXResults
}

func (tx *Transaction) CalculateHash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Requests) == 1 && len(tx.Requests[0].ID) == 0 && tx.Requests[0].Out == -1
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var requests []TXRequest
	var results []TXResult

	for _, req := range tx.Requests {
		requests = append(requests, TXRequest{
			ID:        req.ID,
			Out:       req.Out,
			Signature: nil,
			PubKey:    nil,
		})
	}

	for _, res := range tx.Results {
		results = append(results, TXResult{
			Value:      res.Value,
			PubKeyHash: nil,
		})
	}

	txCopy := Transaction{
		ID:       tx.ID,
		Requests: requests,
		Results:  results,
	}

	return txCopy
}

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTxs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, req := range tx.Requests {
		if prevTxs[hex.EncodeToString(req.ID)].ID == nil {
			core.Handle(core.ErrNilPreviousTransactions)
		}
	}

	txCopy := tx.TrimmedCopy()

	for reqId, req := range txCopy.Requests {
		prevTx := prevTxs[hex.EncodeToString(req.ID)]
		txCopy.Requests[reqId].Signature = nil
		txCopy.Requests[reqId].PubKey = prevTx.Results[req.Out].PubKeyHash
		txCopy.ID = txCopy.CalculateHash()
		txCopy.Requests[reqId].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		core.Handle(err)

		signature := append(r.Bytes(), s.Bytes()...)

		tx.Requests[reqId].Signature = signature
	}
}

func (req *TXRequest) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(req.PubKey)

	return bytes.Equal(lockingHash, pubKeyHash)
}

func (res *TXResult) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-wallet.ChecksumLength]

	res.PubKeyHash = pubKeyHash
}

func (res *TXResult) IsLockWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(res.PubKeyHash, pubKeyHash)
}

func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, req := range tx.Requests {
		if prevTxs[hex.EncodeToString(req.ID)].ID == nil {
			core.Handle(core.ErrNilPreviousTransactions)
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for reqId, req := range tx.Requests {
		prevTx := prevTxs[hex.EncodeToString(req.ID)]
		txCopy.Requests[reqId].Signature = nil
		txCopy.Requests[reqId].PubKey = prevTx.Results[req.Out].PubKeyHash
		txCopy.ID = txCopy.CalculateHash()
		txCopy.Requests[reqId].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		signLen := len(req.Signature)

		r.SetBytes(req.Signature[:(signLen / 2)])
		s.SetBytes(req.Signature[(signLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(req.PubKey)

		x.SetBytes(req.PubKey[:(keyLen / 2)])
		y.SetBytes(req.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{
			Curve: curve,
			X:     &x,
			Y:     &y,
		}

		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	return true
}

func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("	Transaction %x:", tx.ID))

	for i, req := range tx.Requests {
		lines = append(lines, fmt.Sprintf("		Request %d:", i))
		lines = append(lines, fmt.Sprintf("			TXID: %x", req.ID))
		lines = append(lines, fmt.Sprintf("			Out: %d", req.Out))
		lines = append(lines, fmt.Sprintf("			Signature: %x", req.Signature))
		lines = append(lines, fmt.Sprintf("			PubKey: %x", req.PubKey))
	}

	for i, res := range tx.Results {
		lines = append(lines, fmt.Sprintf("		Result %d:", i))
		lines = append(lines, fmt.Sprintf("			Value: %d", res.Value))
		lines = append(lines, fmt.Sprintf("			Script: %x", res.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

func (tx *Transaction) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
	core.Handle(err)

	return result.Bytes()
}

func (ress TXResults) Serialize() []byte {
	var buffer bytes.Buffer

	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(ress)
	core.Handle(err)

	return buffer.Bytes()
}

func DeserializeResults(data []byte) TXResults {
	var results TXResults
	decoder := gob.NewDecoder(bytes.NewBuffer(data))

	err := decoder.Decode(&results)
	core.Handle(err)

	return results
}

func CoinbaseTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 24)
		_, err := rand.Read(randData)
		core.Handle(err)

		data = fmt.Sprintf("%x", randData)
	}

	txReq := TXRequest{
		ID:        []byte{},
		Out:       -1,
		Signature: nil,
		PubKey:    []byte(data),
	}

	txResp := NewTXResult(wData.INITIAL_GENESIS_REWARD, to)

	tx := &Transaction{
		ID:       nil,
		Requests: []TXRequest{txReq},
		Results:  []TXResult{*txResp},
	}
	tx.ID = tx.CalculateHash()

	return tx
}

func NewTransaction(from, to string, amount int, chain *Blockchain) *Transaction {
	var requests []TXRequest
	var results []TXResult

	wallets, err := wallet.CreateWallets()
	core.Handle(err)

	w := wallets.GetWallet(from)
	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)

	acc, validResults := chain.FindSpendableResults(pubKeyHash, amount)

	if acc < amount {
		core.Handle(core.ErrEnoughFunds)
	}

	for txhash, res := range validResults {
		txHash, err := hex.DecodeString(txhash)
		core.Handle(err)

		for _, rs := range res {
			request := TXRequest{
				ID:        txHash,
				Out:       rs,
				Signature: nil,
				PubKey:    w.PublicKey,
			}
			requests = append(requests, request)
		}
	}

	results = append(results, *NewTXResult(amount, to))

	if acc > amount {
		results = append(results, *NewTXResult(acc-amount, to))
	}

	tx := Transaction{
		ID:       nil,
		Requests: requests,
		Results:  results,
	}
	tx.ID = tx.CalculateHash()
	chain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

func NewTXResult(value int, address string) *TXResult {
	tx := &TXResult{
		Value:      value,
		PubKeyHash: nil,
	}

	tx.Lock([]byte(address))

	return tx
}
