package election

import (
	"encoding/json"
	"time"

	"github.com/cloudflare/bn256"
)

type Election struct {
	ID         string
	Name       string
	Candidates []Candidate
	StartTime  time.Time
	EndTime    time.Time
	PublicKey  *bn256.G1
}

type Candidate struct {
	ID   string
	Name string
}

func (e *Election) Serialize() ([]byte, error) {
	return json.Marshal(e)
}
