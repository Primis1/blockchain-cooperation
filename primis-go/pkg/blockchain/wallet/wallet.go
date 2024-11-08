package wallet

import (
	"blockchain/pkg/utils"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

// 															wallet
// NOTE Private key - unique and secure field, which is identifier for user auth

// NOTE Public key - not secret but also kinda unique identifier for user interactions
// NOTE it created from Private key
// 															wallet

// --
// NOTE Address - created from public key

type Wallet struct {
	// NOTE elliptical curve assigning algorithm

	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NOTE generate key for wallet
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	//NOTE outputs will be equal to 256 bytes
	curve := elliptic.P256()

	// NOTE not deterministic random value generator
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	utils.HandleErr(err)

	// NOTE remember the struct. We simply combine X/Y intercepts into pubKey
	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pub
}


