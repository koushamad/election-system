// pkg/blockchain/transaction.go
package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
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

func (t *Transaction) Sign(privateKey []byte) error {
	// Implementation of transaction signing
	// This would use the crypto package for actual signing
	return nil
}

func (t *Transaction) VerifySignature() bool {
	// Verify that the transaction was signed by the owner of the public key
	return true
}

func (t *Transaction) Validate() bool {
	// Basic validation
	if t.ID == "" || len(t.Hash) == 0 || t.Timestamp == 0 {
		return false
	}

	// Verify hash
	calculatedHash := t.CalculateHash()
	if string(calculatedHash) != string(t.Hash) {
		return false
	}

	// Verify signature
	if !t.VerifySignature() {
		return false
	}

	// Type-specific validation
	switch t.Type {
	case TxCreateElection:
		return t.validateCreateElection()
	case TxCastVote:
		return t.validateCastVote()
	case TxTallyVotes:
		return t.validateTallyVotes()
	default:
		return false
	}
}

func (t *Transaction) validateCreateElection() bool {
	var election struct {
		ID         string   `json:"id"`
		Name       string   `json:"name"`
		Candidates []string `json:"candidates"`
		StartTime  int64    `json:"start_time"`
		EndTime    int64    `json:"end_time"`
	}

	if err := json.Unmarshal(t.Payload, &election); err != nil {
		return false
	}

	// Basic validation
	if election.ID == "" || election.Name == "" || len(election.Candidates) < 2 {
		return false
	}

	// Time validation
	now := time.Now().Unix()
	if election.StartTime <= now || election.EndTime <= election.StartTime {
		return false
	}

	return true
}

func (t *Transaction) validateCastVote() bool {
	var vote struct {
		ElectionID string `json:"election_id"`
		Ballot     struct {
			Ciphertext [][]byte `json:"ciphertext"`
			ZKProof    []byte   `json:"zk_proof"`
		} `json:"ballot"`
	}

	if err := json.Unmarshal(t.Payload, &vote); err != nil {
		return false
	}

	// Basic validation
	if vote.ElectionID == "" || len(vote.Ballot.Ciphertext) != 2 || len(vote.Ballot.ZKProof) == 0 {
		return false
	}

	// In a real implementation, we would verify the ZK proof here

	return true
}

func (t *Transaction) validateTallyVotes() bool {
	var tally struct {
		ElectionID string         `json:"election_id"`
		Results    map[string]int `json:"results"`
		Proof      []byte         `json:"proof"`
	}

	if err := json.Unmarshal(t.Payload, &tally); err != nil {
		return false
	}

	// Basic validation
	if tally.ElectionID == "" || len(tally.Results) == 0 || len(tally.Proof) == 0 {
		return false
	}

	// In a real implementation, we would verify the tally proof here

	return true
}

// Helper function to generate a UUID
func GenerateUUID() string {
	// Simple implementation for demonstration purposes
	hash := sha256.Sum256([]byte(time.Now().String()))
	return fmt.Sprintf("%x", hash[:8])
}
