// NOTE we create a unique identifier - private key
// NOTE then public key, which will be used for signatures
// NOTE then address, but it is not part of wallet object, rather its method

package wallet

import (
	"blockchain/pkg/logging"
	"blockchain/pkg/sha"
	"blockchain/pkg/utils"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"golang.org/x/crypto/ripemd160"
)

// 															wallet
// NOTE Private key - unique and secure field, which is identifier for user auth

// NOTE Public key - not secret but also unique identifier, which allow to the check the signature
// NOTE it created from Private key
// 															wallet
// NOTE Address - shortened version of pubkey, made for transferring funds

var info = logging.Info
var errMsg = logging.Error

const (
	checksumLength = 4
	// 0x80 = 0 in HEX

	version = byte(0x80)
)

type Wallet struct {
	// NOTE elliptical curve assigning algorithm

	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NOTE To create address:
// NOTE		sha256()
// NOTE		ripemd160(also a hashing algorithm)
// NOTE		public Key hash
// NOTE		1. sha256()
// NOTE		2. sha256()
// NOTE		3. take first 4 bytes
//
// NOTE		base 58 <- checksum() + version + public key hash

func (w Wallet) Address() []byte {
	// NOTE		sha256()
	pubHash := PublicKey(w.PublicKey)

	// NOTE version
	versionHash := append([]byte{version}, pubHash...)

	// NOTE checksum()
	checksum := checkSum(versionHash)

	fullHash := append(versionHash, checksum...)

	// NOTE		base 58
	address := utils.Base58Encode(fullHash)

	info.Info(address)

	return address
}

// NOTE generate key for wallet
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	//NOTE outputs will be equal to 256 bytes
	curve := elliptic.P256()

	// NOTE not deterministic random value generator
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	utils.HandleErr(err)

	// NOTE remember the struct. We simply combine X/Y intercepts into pubKey
	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pub
}

func PublicKey(pubkey []byte) []byte {
	sha1 := sha.ComputeHash(pubkey)

	hasher := ripemd160.New()
	_, err := hasher.Write(sha1[:])
	utils.HandleErr(err)

	// NOTE Sum() concatenates two values, but we want to launch
	// NOTE the ripemd160 algorithm we pass nil value
	publicRipMD := hasher.Sum(nil)

	return publicRipMD
}

func checkSum(payload []byte) []byte {
	sha1 := sha.ComputeHash(payload)
	sha2 := sha.ComputeHash(sha1[:])

	// return first 4 bytes
	return sha2[:checksumLength]
}

func ValidateAddress(address string) bool {
	pubKeyHash := utils.Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]
	targetChecksum := checkSum(append([]byte{version}, pubKeyHash...))

	return bytes.Equal(actualChecksum, targetChecksum)
}

func makeWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}
