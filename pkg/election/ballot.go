package election

import (
	"election-system/crypto"
	"github.com/cloudflare/bn256"
	"github.com/gtank/merlin"
)

type Ballot struct {
	Ciphertext []*bn256.G1
	ZKProof    []byte
	VoterID    string
}

func NewBallot(ciphertext []*bn256.G1, proof []byte, voterID string) *Ballot {
	return &Ballot{
		Ciphertext: ciphertext,
		ZKProof:    proof,
		VoterID:    voterID,
	}
}

func (b *Ballot) Validate() bool {
	transcript := merlin.NewTranscript("vote_proof")
	return crypto.VerifyZKProof(transcript, b.Ciphertext, b.ZKProof)
}
