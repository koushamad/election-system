// pkg/election/ballot.go
package election

import (
	"github.com/cloudflare/bn256"
	"github.com/gtank/merlin"
	"github.com/koushamad/election-system/pkg/crypto"
)

type Ballot struct {
	Ciphertext []*bn256.G1 `json:"ciphertext"`
	ZKProof    []byte      `json:"zk_proof"`
	VoterID    string      `json:"voter_id"`
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
