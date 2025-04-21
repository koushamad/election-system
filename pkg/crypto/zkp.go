package crypto

import (
	"bytes"
	"math/big"

	"github.com/cloudflare/bn256"
	"github.com/gtank/merlin"
)

func GenerateVoteProof(ciphertext []*bn256.G1, r *big.Int, vote int) []byte {
	transcript := merlin.NewTranscript("vote_proof")
	transcript.AppendMessage([]byte("commitment"), ciphertext[0].Marshal())
	transcript.AppendMessage([]byte("ciphertext"), ciphertext[1].Marshal())

	// Use the challenge in the proof calculation or remove it if not needed
	// The following line is a simplified way to use the challenge
	challenge := transcript.ExtractBytes([]byte("challenge"), 32)

	// Create a proof that combines the random value r with the challenge
	// This is a simplified implementation - a real ZKP would be more complex
	challengeInt := new(big.Int).SetBytes(challenge)
	proofInt := new(big.Int).Add(r, challengeInt)

	// Return the proof as a marshaled point
	proof := new(bn256.G1).ScalarBaseMult(proofInt)
	return proof.Marshal()
}

func VerifyZKProof(transcript *merlin.Transcript, ciphertext []*bn256.G1, proof []byte) bool {
	transcript.AppendMessage([]byte("commitment"), ciphertext[0].Marshal())
	transcript.AppendMessage([]byte("ciphertext"), ciphertext[1].Marshal())

	challenge := transcript.ExtractBytes([]byte("challenge"), 32)

	// Recreate the expected proof point
	expectedInt := new(big.Int).SetBytes(challenge)
	expected := new(bn256.G1).ScalarBaseMult(expectedInt)

	// Unmarshal the provided proof
	actual, err := new(bn256.G1).Unmarshal(proof)
	if err != nil {
		return false
	}

	// Compare the expected and actual proofs
	// In a real implementation, this would verify the mathematical relationship
	return bytes.Equal(expected.Marshal(), actual)
}
