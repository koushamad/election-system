package crypto

import (
	"crypto/rand"
	"math/big"

	"github.com/cloudflare/bn256"
	"github.com/gtank/merlin"
)

func GenerateVoteProof(ciphertext []*bn256.G1, r *big.Int, vote int) []byte {
	transcript := merlin.NewTranscript("vote_proof")
	transcript.AppendMessage([]byte("commitment"), ciphertext[0].Marshal())
	transcript.AppendMessage([]byte("ciphertext"), ciphertext[1].Marshal())

	challenge := transcript.ExtractBytes([]byte("challenge"), 32)
	proof := new(bn256.G1).ScalarBaseMult(r)
	return proof.Marshal()
}

func VerifyVoteProof(ciphertext []*bn256.G1, proof []byte) bool {
	transcript := merlin.NewTranscript("vote_proof")
	transcript.AppendMessage([]byte("commitment"), ciphertext[0].Marshal())
	transcript.AppendMessage([]byte("ciphertext"), ciphertext[1].Marshal())

	challenge := transcript.ExtractBytes([]byte("challenge"), 32)
	expected := new(bn256.G1).ScalarBaseMult(new(big.Int).SetBytes(challenge))
	actual := new(bn256.G1).Unmarshal(proof)
	return expected.String() == actual.String()
}
