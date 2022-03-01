package utils

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

func DecodeAddress(address string) []byte {
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	return pubKeyHash
}
