package wallet

import (
	"github.com/mr-tron/base58"
	"github.com/wilmacedo/willchain-go/core"
)

func Base58Encode(data []byte) []byte {
	encode := base58.Encode(data)

	return []byte(encode)
}

func Base58Decode(data []byte) []byte {
	decode, err := base58.Decode(string(data[:]))
	core.Handle(err)

	return decode
}
