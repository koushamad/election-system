// pkg/crypto/keys.go
package crypto

import (
	"crypto/rand"
	"github.com/cloudflare/bn256"
	"math/big"
)

type KeyPair struct {
	PrivateKey *big.Int
	PublicKey  *bn256.G1
}

func GenerateKeys() *KeyPair {
	priv, _ := rand.Int(rand.Reader, bn256.Order)
	pub := new(bn256.G1).ScalarBaseMult(priv)
	return &KeyPair{priv, pub}
}
