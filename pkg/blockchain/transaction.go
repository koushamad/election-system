// pkg/blockchain/transaction.go
package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type TransactionType string

const (
	TxCreateElection TransactionType = "create_election"
	TxCastVote       TransactionType = "cast_vote"
	TxTallyVotes     TransactionType = "tally_votes"
)

type Transaction struct {
	ID        string          `json:"id"`
	Type      TransactionType `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Hash      []byte          `json:"hash"`
	Timestamp int64           `json:"timestamp"`
	Signature []byte          `json:"signature"`
	PublicKey []byte          `json:"public_key"` // Sender's public key
}

func NewTransaction(txType TransactionType, payload interface{}) (*Transaction, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	tx := &Transaction{
		ID:        GenerateUUID(),
		Type:      txType,
		Payload:   payloadBytes,
		Timestamp: time.Now().Unix(),
	}

	tx.Hash = tx.CalculateHash()
	return tx, nil
}

func (t *Transaction) CalculateHash() []byte {
	txData := struct {
		ID        string
		Type      TransactionType
		Payload   json.RawMessage
		Timestamp int64
		PublicKey []byte
	}{
		ID:        t.ID,
		Type:      t.Type,
		Payload:   t.Payload,
		Timestamp: t.Timestamp,
		PublicKey: t.PublicKey,
	}

	data, _ := json.Marshal(txData)
	hash := sha256.Sum256(data)
	return hash[:]
}

func (t *Transaction) Validate() bool {
	// Basic validation
	if t.ID == "" || len(t.Hash) == 0 {
		return false
	}

	// For testing purposes, detect manually corrupted hash
	if string(t.Hash) == "invalid-hash" {
		return false
	}

	// Verify hash matches the calculated hash
	calculatedHash := t.CalculateHash()
	return bytes.Equal(calculatedHash, t.Hash)
}

// Helper function to generate a UUID
func GenerateUUID() string {
	// Simple implementation for demonstration purposes
	hash := sha256.Sum256([]byte(time.Now().String()))
	return fmt.Sprintf("%x", hash[:8])
}
