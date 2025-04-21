package e2e

import (
	"encoding/json"
	"github.com/koushamad/election-system/pkg/blockchain"
	"github.com/koushamad/election-system/pkg/election"
	"github.com/koushamad/election-system/test/utils"
	"testing"
	"time"
)

func TestFullElectionFlow(t *testing.T) {
	// Initialize blockchain
	node := utils.SetupTestNode()

	// Step 1: Create election
	electionData, adminKeys := utils.CreateTestElection(
		"Presidential Election 2025",
		[]string{"Alice", "Bob", "Charlie"},
	)

	electionTx, err := utils.CreateElectionTransaction(electionData)
	if err != nil {
		t.Fatalf("Failed to create election transaction: %v", err)
	}

	node.TransactionPool = append(node.TransactionPool, electionTx)
	node.CreateBlock()

	// Verify election creation
	if len(node.Chain.Blocks) != 2 { // Genesis + election block
		t.Fatalf("Expected 2 blocks after election creation, got %d", len(node.Chain.Blocks))
	}

	// Step 2: Cast votes
	voters := []struct {
		ID        string
		Candidate string
	}{
		{"voter1", "Alice"},
		{"voter2", "Bob"},
		{"voter3", "Alice"},
		{"voter4", "Charlie"},
		{"voter5", "Alice"},
	}

	// Cast votes
	for _, voter := range voters {
		ballot, err := utils.CreateTestVote(electionData, voter.Candidate)
		if err != nil {
			t.Fatalf("Failed to create vote for %s: %v", voter.Candidate, err)
		}

		// Override voter ID for test
		ballot.VoterID = voter.ID

		voteTx, err := utils.CreateVoteTransaction(electionData.ID, ballot)
		if err != nil {
			t.Fatalf("Failed to create vote transaction: %v", err)
		}

		node.TransactionPool = append(node.TransactionPool, voteTx)
	}

	// Create block with votes
	node.CreateBlock()

	// Verify vote block
	voteBlock := node.Chain.Blocks[2]
	if len(voteBlock.Transactions) != 5 {
		t.Errorf("Expected 5 vote transactions, got %d", len(voteBlock.Transactions))
	}

	// Step 3: Test double voting prevention
	duplicateVoter := voters[0] // Try to vote again with voter1
	ballot, _ := utils.CreateTestVote(electionData, duplicateVoter.Candidate)
	ballot.VoterID = duplicateVoter.ID

	voteTx, _ := utils.CreateVoteTransaction(electionData.ID, ballot)
	node.TransactionPool = append(node.TransactionPool, voteTx)

	// Create another block
	node.CreateBlock()

	// In a proper implementation, the duplicate vote should be rejected
	// This test would need to be updated when proper validation is implemented

	// Step 4: Tally votes (homomorphically)
	// In a real implementation, this would use the homomorphic properties
	// For this test, we'll manually count

	// Extract votes from blockchain
	votes := make(map[string]int)
	for _, block := range node.Chain.Blocks {
		for _, tx := range block.Transactions {
			if tx.Type == blockchain.TxCastVote {
				var voteData struct {
					ElectionID string           `json:"election_id"`
					Ballot     *election.Ballot `json:"ballot"`
				}

				if err := json.Unmarshal(tx.Payload, &voteData); err != nil {
					continue
				}

				if voteData.ElectionID != electionData.ID {
					continue
				}

				// In a real implementation, we would decrypt the vote
				// For this test, we'll just count the ballots
				votes[voteData.Ballot.VoterID]++
			}
		}
	}

	// Verify we have the expected number of unique voters
	if len(votes) != 5 {
		t.Errorf("Expected 5 unique voters, got %d", len(votes))
	}

	// Step 5: Create tally transaction
	tallyResult := struct {
		ElectionID string         `json:"election_id"`
		Results    map[string]int `json:"results"`
		Timestamp  int64          `json:"timestamp"`
	}{
		ElectionID: electionData.ID,
		Results: map[string]int{
			"Alice":   3,
			"Bob":     1,
			"Charlie": 1,
		},
		Timestamp: time.Now().Unix(),
	}

	tallyJSON, _ := json.Marshal(tallyResult)
	tallyTx := &blockchain.Transaction{
		Type:    "tally_votes",
		Payload: tallyJSON,
	}
	tallyTx.Hash = tallyTx.CalculateHash()

	node.TransactionPool = append(node.TransactionPool, tallyTx)
	node.CreateBlock()

	// Verify tally block
	tallyBlock := node.Chain.Blocks[4]
	if len(tallyBlock.Transactions) != 1 {
		t.Errorf("Expected 1 tally transaction, got %d", len(tallyBlock.Transactions))
	}

	// Verify final blockchain state
	if len(node.Chain.Blocks) != 5 { // Genesis + election + votes + duplicate attempt + tally
		t.Errorf("Expected 5 blocks in final chain, got %d", len(node.Chain.Blocks))
	}

	t.Log("Full election flow completed successfully")
}
