package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
)

type TransactionType string

const (
	TxCreateElection TransactionType = "create_election"
	TxCastVote       TransactionType = "cast_vote"
)

type Transaction struct {
	Type    TransactionType `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Hash    []byte          `json:"hash"`
}

func (t *Transaction) CalculateHash() []byte {
	data, _ := json.Marshal(t)
	hash := sha256.Sum256(data)
	return hash[:]
}

func (t *Transaction) Validate() bool {
	// Add validation logic based on transaction type
	return true
}
