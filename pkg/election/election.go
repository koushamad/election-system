// pkg/election/election.go
package election

import (
	"github.com/cloudflare/bn256"
	"time"
)

type Election struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Candidates []Candidate `json:"candidates"`
	StartTime  time.Time   `json:"start_time"`
	EndTime    time.Time   `json:"end_time"`
	PublicKey  *bn256.G1   `json:"public_key"`
}

type Candidate struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
