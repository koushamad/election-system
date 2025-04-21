// pkg/crypto/encryption.go
package crypto

import (
	"crypto/rand"
	"github.com/cloudflare/bn256"
	"math/big"
)

func EncryptVote(pubKey *bn256.G1, vote int) ([]*bn256.G1, error) {
	r, _ := rand.Int(rand.Reader, bn256.Order)

	// Create the first part of the ciphertext: g^r
	c1 := new(bn256.G1).ScalarBaseMult(r)

	// Create the second part: pubKey^r * g^vote
	votePoint := new(bn256.G1).ScalarBaseMult(big.NewInt(int64(vote)))
	temp := new(bn256.G1).ScalarMult(pubKey, r)
	c2 := new(bn256.G1).Add(temp, votePoint)

	return []*bn256.G1{c1, c2}, nil
}
