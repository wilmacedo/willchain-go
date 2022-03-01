package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"github.com/wilmacedo/willchain-go/core"
	"github.com/wilmacedo/willchain-go/utils"
	"golang.org/x/crypto/ripemd160"
)

const (
	ChecksumLength = 4
	version        = byte(0x00) // hex representation of zero number
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (wallet Wallet) Address() []byte {
	pubHash := PublicKeyHash(wallet.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := utils.Base58Encode(fullHash)

	return address
}

func ValidateAddress(address string) bool {
	pubKeyHash := utils.Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-ChecksumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-ChecksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Equal(actualChecksum, targetChecksum)
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	core.Handle(err)

	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, public
}

func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := &Wallet{
		PublicKey:  public,
		PrivateKey: private,
	}

	return wallet
}

func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	core.Handle(err)

	publicRipMD := hasher.Sum(nil)

	return publicRipMD
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:ChecksumLength]
}
