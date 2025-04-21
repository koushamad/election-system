package crypto

import (
	"bytes" // Replace crypto/rand with bytes for comparison
	"math/big"

	"github.com/cloudflare/bn256"
	"github.com/gtank/merlin"
)

func GenerateVoteProof(ciphertext []*bn256.G1, r *big.Int, vote int) []byte {
	transcript := merlin.NewTranscript("vote_proof")
	transcript.AppendMessage([]byte("commitment"), ciphertext[0].Marshal())
	transcript.AppendMessage([]byte("ciphertext"), ciphertext[1].Marshal())

	// Use the challenge in the proof calculation
	challenge := transcript.ExtractBytes([]byte("challenge"), 32)

	// This is a simplified proof for demonstration
	// In a real implementation, this would be more complex
	proof := new(bn256.G1).ScalarBaseMult(r)
	return proof.Marshal()
}

func VerifyZKProof(transcript *merlin.Transcript, ciphertext []*bn256.G1, proof []byte) bool {
	transcript.AppendMessage([]byte("commitment"), ciphertext[0].Marshal())
	transcript.AppendMessage([]byte("ciphertext"), ciphertext[1].Marshal())

	challenge := transcript.ExtractBytes([]byte("challenge"), 32)
	expected := new(bn256.G1).ScalarBaseMult(new(big.Int).SetBytes(challenge))

	// Fix the Unmarshal call to handle both return values
	actual, err := new(bn256.G1).Unmarshal(proof)
	if err != nil {
		return false
	}

	// Replace String() with a proper comparison
	return bytes.Equal(expected.Marshal(), actual.Marshal())
}
