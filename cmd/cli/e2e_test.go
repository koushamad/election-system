package main

import (
	"election-system/blockchain"
	"election-system/crypto"
	"election-system/election"
	"encoding/json"
	"github.com/cloudflare/bn256"
	"github.com/gtank/merlin"
	"math/big"
	"testing"
)

func TestFullElectionFlow(t *testing.T) {
	// Initialize blockchain
	blockchain.ClearTransactionPool()

	// Create election
	adminKeys := crypto.GenerateKeys()
	election := struct {
		ID         string
		Candidates []string
	}{
		ID:         "2024-election",
		Candidates: []string{"Candidate A", "Candidate B"},
	}

	data, _ := json.Marshal(election)
	tx := blockchain.Transaction{
		ID:   "tx1",
		Type: blockchain.TxCreateElection,
		Data: data,
	}
	if err := blockchain.AddTransaction(tx); err != nil {
		t.Fatalf("Failed to create election: %v", err)
	}

	// Generate voter keys
	voterKeys := crypto.GenerateKeys()

	// Test valid vote
	t.Run("Valid vote submission", func(t *testing.T) {
		ciphertext, _ := crypto.EncryptVote(adminKeys.PublicKey, 1)
		transcript := merlin.NewTranscript("vote_proof")
		r, _ := rand.Int(rand.Reader, bn256.Order)
		proof := crypto.GenerateZKProof(transcript, ciphertext, r, 1)

		ballot := election.NewBallot(ciphertext, proof, "voter1")
		if !ballot.Validate() {
			t.Error("Valid ballot failed verification")
		}

		ballotData, _ := json.Marshal(ballot)
		tx := blockchain.Transaction{
			ID:   "tx2",
			Type: blockchain.TxCastVote,
			Data: ballotData,
		}
		if err := blockchain.AddTransaction(tx); err != nil {
			t.Errorf("Failed to cast valid vote: %v", err)
		}
	})

	// Test invalid ZKP
	t.Run("Invalid zero-knowledge proof", func(t *testing.T) {
		ciphertext, _ := crypto.EncryptVote(adminKeys.PublicKey, 1)
		invalidProof := make([]byte, 64) // Random invalid proof

		ballot := election.NewBallot(ciphertext, invalidProof, "voter2")
		if ballot.Validate() {
			t.Error("Invalid proof was accepted")
		}
	})

	// Test double voting
	t.Run("Double voting prevention", func(t *testing.T) {
		ciphertext, _ := crypto.EncryptVote(adminKeys.PublicKey, 1)
		transcript := merlin.NewTranscript("vote_proof")
		r, _ := rand.Int(rand.Reader, bn256.Order)
		proof := crypto.GenerateZKProof(transcript, ciphertext, r, 1)

		ballotData, _ := json.Marshal(election.NewBallot(ciphertext, proof, "voter3"))
		tx := blockchain.Transaction{
			ID:   "tx3",
			Type: blockchain.TxCastVote,
			Data: ballotData,
		}
		blockchain.AddTransaction(tx)

		// Try to vote again with same ID
		if err := blockchain.AddTransaction(tx); err == nil {
			t.Error("Duplicate transaction was accepted")
		}
	})

	// Test tallying
	t.Run("Vote tallying", func(t *testing.T) {
		tallyResult := struct {
			Winner string `json:"winner"`
		}{
			Winner: "Candidate A",
		}

		data, _ := json.Marshal(tallyResult)
		tx := blockchain.Transaction{
			ID:   "tx4",
			Type: blockchain.TxTallyVotes,
			Data: data,
		}
		if err := blockchain.AddTransaction(tx); err != nil {
			t.Errorf("Tally transaction failed: %v", err)
		}
	})

	// Final blockchain validation
	t.Run("Blockchain integrity", func(t *testing.T) {
		pool := blockchain.GetTransactionPool()
		if len(pool) != 3 { // 1 create + 2 valid votes + 1 tally (but 1 vote was invalid)
			t.Errorf("Unexpected transaction count: %d", len(pool))
		}
	})
}
