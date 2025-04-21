// pkg/crypto/zkp.go
package crypto

import (
	"github.com/cloudflare/bn256"
	"github.com/gtank/merlin"
	"math/big"
)

func GenerateVoteProof(ciphertext []*bn256.G1, r *big.Int, vote int) []byte {
	transcript := merlin.NewTranscript("vote_proof")
	transcript.AppendMessage([]byte("commitment"), ciphertext[0].Marshal())
	transcript.AppendMessage([]byte("ciphertext"), ciphertext[1].Marshal())

	challenge := transcript.ExtractBytes([]byte("challenge"), 32)

	// Use the challenge in the proof calculation
	// This is a simplified proof for demonstration
	// In a real implementation, this would be more complex
	proof := new(bn256.G1).ScalarBaseMult(new(big.Int).SetBytes(challenge))
	return proof.Marshal()
}

func VerifyVoteProof(ciphertext []*bn256.G1, proof []byte) bool {
	transcript := merlin.NewTranscript("vote_proof")
	transcript.AppendMessage([]byte("commitment"), ciphertext[0].Marshal())
	transcript.AppendMessage([]byte("ciphertext"), ciphertext[1].Marshal())

	challenge := transcript.ExtractBytes([]byte("challenge"), 32)
	expected := new(bn256.G1).ScalarBaseMult(new(big.Int).SetBytes(challenge))

	// Fix the Unmarshal call to handle both return values
	actual, err := new(bn256.G1).Unmarshal(proof)
	if err != nil {
		return false
	}

	return expected.String() == string(actual)
}
