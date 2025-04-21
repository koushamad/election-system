package integration

import (
	"github.com/cloudflare/bn256"
	"github.com/koushamad/election-system/pkg/blockchain"
	"github.com/koushamad/election-system/pkg/election"
	"github.com/koushamad/election-system/test/utils"
	"math/big"
	"testing"
	"time"
)

func TestElectionCreation(t *testing.T) {
	node := utils.SetupTestNode()

	// Create test election
	electionData, keys := utils.CreateTestElection(
		"Presidential Election 2025",
		[]string{"Alice", "Bob", "Charlie"},
	)

	// Verify election properties
	if electionData.Name != "Presidential Election 2025" {
		t.Errorf("Expected name 'Presidential Election 2025', got '%s'", electionData.Name)
	}

	if len(electionData.Candidates) != 3 {
		t.Errorf("Expected 3 candidates, got %d", len(electionData.Candidates))
	}

	// Verify election dates (should be in the future as of April 21, 2025)
	now := time.Date(2025, 4, 21, 12, 0, 0, 0, time.UTC)
	if !electionData.StartTime.After(now) {
		t.Errorf("Election start time should be in the future")
	}

	if !electionData.EndTime.After(electionData.StartTime) {
		t.Errorf("Election end time should be after start time")
	}

	// Create and add election transaction
	tx, err := utils.CreateElectionTransaction(electionData)
	if err != nil {
		t.Fatalf("Failed to create election transaction: %v", err)
	}

	node.TransactionPool = append(node.TransactionPool, tx)
	block := node.CreateBlock()

	// Verify transaction was included in block
	if len(block.Transactions) != 1 {
		t.Errorf("Expected 1 transaction in block, got %d", len(block.Transactions))
	}

	if block.Transactions[0].Type != blockchain.TxCreateElection {
		t.Errorf("Expected transaction type '%s', got '%s'",
			blockchain.TxCreateElection, block.Transactions[0].Type)
	}
}

func TestVoteCasting(t *testing.T) {
	node := utils.SetupTestNode()

	// Create test election
	electionData, _ := utils.CreateTestElection(
		"Local Election 2025",
		[]string{"Alice", "Bob", "Charlie"},
	)

	// Create and add election transaction
	electionTx, _ := utils.CreateElectionTransaction(electionData)
	node.TransactionPool = append(node.TransactionPool, electionTx)
	node.CreateBlock()

	// Create votes for different candidates
	candidates := []string{"Alice", "Bob", "Charlie"}
	for _, candidate := range candidates {
		// Create ballot
		ballot, err := utils.CreateTestVote(electionData, candidate)
		if err != nil {
			t.Fatalf("Failed to create vote for %s: %v", candidate, err)
		}

		// Verify ballot
		if !ballot.Validate() {
			t.Errorf("Ballot for %s failed validation", candidate)
		}

		// Create vote transaction
		voteTx, err := utils.CreateVoteTransaction(electionData.ID, ballot)
		if err != nil {
			t.Fatalf("Failed to create vote transaction: %v", err)
		}

		// Add to transaction pool
		node.TransactionPool = append(node.TransactionPool, voteTx)
	}

	// Create block with votes
	block := node.CreateBlock()

	// Verify all votes were included
	if len(block.Transactions) != 3 {
		t.Errorf("Expected 3 transactions in block, got %d", len(block.Transactions))
	}

	// Verify all transactions are vote transactions
	for _, tx := range block.Transactions {
		if tx.Type != blockchain.TxCastVote {
			t.Errorf("Expected transaction type '%s', got '%s'",
				blockchain.TxCastVote, tx.Type)
		}
	}
}

func TestZeroKnowledgeProofs(t *testing.T) {
	// Create test election
	electionData, _ := utils.CreateTestElection(
		"ZKP Test Election",
		[]string{"Alice", "Bob"},
	)

	// Create valid vote
	ballot, err := utils.CreateTestVote(electionData, "Alice")
	if err != nil {
		t.Fatalf("Failed to create vote: %v", err)
	}

	// Verify valid ballot passes validation
	if !ballot.Validate() {
		t.Error("Valid ballot failed verification")
	}

	// Create invalid proof
	invalidBallot := &election.Ballot{
		Ciphertext: ballot.Ciphertext,
		ZKProof:    make([]byte, 64), // Random invalid proof
		VoterID:    ballot.VoterID,
	}

	// Verify invalid ballot fails validation
	if invalidBallot.Validate() {
		t.Error("Invalid ballot passed verification")
	}

	// Test tampering with ciphertext
	tamperedBallot := &election.Ballot{
		Ciphertext: []*bn256.G1{
			new(bn256.G1).ScalarBaseMult(big.NewInt(999)),
			ballot.Ciphertext[1],
		},
		ZKProof: ballot.ZKProof,
		VoterID: ballot.VoterID,
	}

	// Verify tampered ballot fails validation
	if tamperedBallot.Validate() {
		t.Error("Tampered ballot passed verification")
	}
}
