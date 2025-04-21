// pkg/crypto/encryption.go
package crypto

import (
	"crypto/rand"
	"github.com/cloudflare/bn256"
	"math/big"
)

func EncryptVote(pubKey *bn256.G1, vote int) ([]*bn256.G1, error) {
	r, _ := rand.Int(rand.Reader, bn256.Order)
	return []*bn256.G1{
		new(bn256.G1).ScalarBaseMult(r),
		new(bn256.G1).ScalarMult(pubKey, r).Add(
			new(bn256.G1).ScalarBaseMult(big.NewInt(int64(vote)))),
	}, nil
}
