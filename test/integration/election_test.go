package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cloudflare/bn256"
	"github.com/koushamad/election-system/pkg/blockchain"
	"github.com/koushamad/election-system/pkg/election"
	"github.com/koushamad/election-system/test/utils"
	"math/big"
	"testing"
	"time"
)

// Existing tests remain unchanged...

func TestMultiNodeConsensus(t *testing.T) {
	// Create multiple nodes to simulate a network
	node1 := utils.SetupTestNode()
	node2 := utils.SetupTestNode()
	node3 := utils.SetupTestNode()

	// Create test election
	election, _ := utils.CreateTestElection("Presidential Election 2025", []string{"Alice", "Bob", "Charlie"})
	tx, err := utils.CreateElectionTransaction(election)
	if err != nil {
		t.Fatalf("Failed to create election transaction: %v", err)
	}

	// Add transaction to node1 and create block
	node1.TransactionPool = append(node1.TransactionPool, tx)
	block := node1.CreateBlock()

	// Verify block was created on node1
	if len(node1.Chain.Blocks) != 2 { // Genesis + new block
		t.Errorf("Expected 2 blocks in node1 chain, got %d", len(node1.Chain.Blocks))
	}

	// Simulate block propagation to other nodes
	err = node2.AddBlock(block)
	if err != nil {
		t.Errorf("Failed to add block to node2: %v", err)
	}

	err = node3.AddBlock(block)
	if err != nil {
		t.Errorf("Failed to add block to node3: %v", err)
	}

	// Verify all nodes have the same blockchain state
	if len(node2.Chain.Blocks) != 2 || len(node3.Chain.Blocks) != 2 {
		t.Errorf("Chain length mismatch across nodes: node1=%d, node2=%d, node3=%d",
			len(node1.Chain.Blocks), len(node2.Chain.Blocks), len(node3.Chain.Blocks))
	}

	// Verify block hash consistency across nodes
	if !bytes.Equal(node1.Chain.Blocks[1].Hash, node2.Chain.Blocks[1].Hash) ||
		!bytes.Equal(node1.Chain.Blocks[1].Hash, node3.Chain.Blocks[1].Hash) {
		t.Error("Block hash inconsistency across nodes")
	}
}

func TestBlockchainPersistence(t *testing.T) {
	// Create a node and add some blocks
	node := utils.SetupTestNode()

	// Create and add multiple transactions
	for i := 0; i < 3; i++ {
		election, _ := utils.CreateTestElection(
			fmt.Sprintf("Election %d", i),
			[]string{"Candidate A", "Candidate B"},
		)
		tx, _ := utils.CreateElectionTransaction(election)
		node.TransactionPool = append(node.TransactionPool, tx)
		node.CreateBlock()
	}

	// Serialize the blockchain to JSON (simulating persistence)
	chainData, err := json.Marshal(node.Chain)
	if err != nil {
		t.Fatalf("Failed to serialize blockchain: %v", err)
	}

	// Deserialize into a new chain (simulating loading from storage)
	var loadedChain blockchain.Chain
	err = json.Unmarshal(chainData, &loadedChain)
	if err != nil {
		t.Fatalf("Failed to deserialize blockchain: %v", err)
	}

	// Verify the loaded chain has the correct number of blocks
	if len(loadedChain.Blocks) != 4 { // Genesis + 3 blocks
		t.Errorf("Expected 4 blocks in loaded chain, got %d", len(loadedChain.Blocks))
	}

	// Verify block integrity in the loaded chain
	for i := 1; i < len(loadedChain.Blocks); i++ {
		block := loadedChain.Blocks[i]
		prevBlock := loadedChain.Blocks[i-1]

		// Verify block index
		if block.Index != prevBlock.Index+1 {
			t.Errorf("Block index mismatch at position %d", i)
		}

		// Verify previous hash
		if !bytes.Equal(block.PrevHash, prevBlock.Hash) {
			t.Errorf("Previous hash mismatch at block %d", i)
		}

		// Verify block hash is correct
		calculatedHash := block.CalculateHash()
		if !bytes.Equal(calculatedHash, block.Hash) {
			t.Errorf("Block hash mismatch at position %d", i)
		}
	}
}

func TestLedgerConsistency(t *testing.T) {
	// Create a node
	node := utils.SetupTestNode()

	// Create test election
	election, _ := utils.CreateTestElection("Local Election 2025", []string{"Alice", "Bob"})
	electionTx, _ := utils.CreateElectionTransaction(election)

	// Add election transaction and create block
	node.TransactionPool = append(node.TransactionPool, electionTx)
	node.CreateBlock()

	// Create and add votes
	for i := 0; i < 5; i++ {
		candidate := "Alice"
		if i%2 == 0 {
			candidate = "Bob"
		}

		ballot, _ := utils.CreateTestVote(election, candidate)
		voteTx, _ := utils.CreateVoteTransaction(election.ID, ballot)

		node.TransactionPool = append(node.TransactionPool, voteTx)
	}

	// Create block with votes
	node.CreateBlock()

	// Verify ledger contains all transactions
	var electionFound, votesFound bool
	var voteCount int

	// Check each block for transactions
	for _, block := range node.Chain.Blocks {
		for _, tx := range block.Transactions {
			if tx.Type == blockchain.TxCreateElection {
				// Verify election transaction
				var electionData election.Election
				json.Unmarshal(tx.Payload, &electionData)
				if electionData.ID == election.ID {
					electionFound = true
				}
			} else if tx.Type == blockchain.TxCastVote {
				// Verify vote transaction
				var voteData struct {
					ElectionID string          `json:"election_id"`
					Ballot     election.Ballot `json:"ballot"`
				}
				json.Unmarshal(tx.Payload, &voteData)
				if voteData.ElectionID == election.ID {
					voteCount++
				}
			}
		}
	}

	votesFound = voteCount == 5

	if !electionFound {
		t.Error("Election transaction not found in blockchain")
	}

	if !votesFound {
		t.Errorf("Expected 5 vote transactions, found %d", voteCount)
	}
}

func TestForkResolution(t *testing.T) {
	// Create two nodes with identical genesis blocks
	node1 := utils.SetupTestNode()
	node2 := utils.SetupTestNode()

	// Create different transactions for each node
	election1, _ := utils.CreateTestElection("Election Node1", []string{"Alice", "Bob"})
	tx1, _ := utils.CreateElectionTransaction(election1)

	election2, _ := utils.CreateTestElection("Election Node2", []string{"Charlie", "Dave"})
	tx2, _ := utils.CreateElectionTransaction(election2)

	// Add transactions and create blocks independently
	node1.TransactionPool = append(node1.TransactionPool, tx1)
	block1 := node1.CreateBlock()

	node2.TransactionPool = append(node2.TransactionPool, tx2)
	block2 := node2.CreateBlock()

	// Verify both nodes have different blocks at position 1
	if bytes.Equal(node1.Chain.Blocks[1].Hash, node2.Chain.Blocks[1].Hash) {
		t.Error("Expected different blocks, but hashes match")
	}

	// Create more blocks on node1 to make its chain longer
	for i := 0; i < 3; i++ {
		election, _ := utils.CreateTestElection(fmt.Sprintf("Extra Election %d", i), []string{"X", "Y"})
		tx, _ := utils.CreateElectionTransaction(election)
		node1.TransactionPool = append(node1.TransactionPool, tx)
		node1.CreateBlock()
	}

	// Verify node1 has a longer chain
	if len(node1.Chain.Blocks) <= len(node2.Chain.Blocks) {
		t.Errorf("Expected node1 to have longer chain: node1=%d, node2=%d",
			len(node1.Chain.Blocks), len(node2.Chain.Blocks))
	}

	// Simulate node2 receiving node1's chain
	originalNode2BlockCount := len(node2.Chain.Blocks)
	node2.ReplaceChain(node1.Chain)

	// Verify node2 adopted the longer chain
	if len(node2.Chain.Blocks) != len(node1.Chain.Blocks) {
		t.Errorf("Chain length mismatch after fork resolution: node1=%d, node2=%d",
			len(node1.Chain.Blocks), len(node2.Chain.Blocks))
	}

	// Verify node2's original block is no longer in its chain
	var originalBlockFound bool
	for _, block := range node2.Chain.Blocks {
		if bytes.Equal(block.Hash, block2.Hash) {
			originalBlockFound = true
			break
		}
	}

	if originalBlockFound {
		t.Error("Node2's original block should have been replaced")
	}

	t.Logf("Successfully resolved fork: node2 replaced its %d blocks with node1's %d blocks",
		originalNode2BlockCount, len(node1.Chain.Blocks))
}

func TestTransactionPoolConsistency(t *testing.T) {
	// Create a node
	node := utils.SetupTestNode()

	// Create multiple transactions
	txs := make([]*blockchain.Transaction, 0)
	for i := 0; i < 5; i++ {
		election, _ := utils.CreateTestElection(
			fmt.Sprintf("Election %d", i),
			[]string{"Candidate A", "Candidate B"},
		)
		tx, _ := utils.CreateElectionTransaction(election)
		txs = append(txs, tx)
	}

	// Add transactions to pool
	for _, tx := range txs {
		node.TransactionPool = append(node.TransactionPool, tx)
	}

	// Verify pool contains all transactions
	if len(node.TransactionPool) != 5 {
		t.Errorf("Expected 5 transactions in pool, got %d", len(node.TransactionPool))
	}

	// Create a block (should include transactions from pool)
	block := node.CreateBlock()

	// Verify block contains all transactions
	if len(block.Transactions) != 5 {
		t.Errorf("Expected 5 transactions in block, got %d", len(block.Transactions))
	}

	// Verify transaction pool is cleared
	if len(node.TransactionPool) != 0 {
		t.Errorf("Expected empty transaction pool after block creation, got %d transactions",
			len(node.TransactionPool))
	}

	// Add a duplicate transaction (that's already in a block)
	err := node.AddTransaction(txs[0])
	if err == nil {
		t.Error("Expected error when adding duplicate transaction, got nil")
	}
}

func TestTimestampOrdering(t *testing.T) {
	// Create a node
	node := utils.SetupTestNode()

	// Create multiple blocks with controlled timestamps
	var timestamps []int64

	for i := 0; i < 5; i++ {
		election, _ := utils.CreateTestElection(
			fmt.Sprintf("Election %d", i),
			[]string{"Candidate A", "Candidate B"},
		)
		tx, _ := utils.CreateElectionTransaction(election)
		node.TransactionPool = append(node.TransactionPool, tx)

		block := node.CreateBlock()
		timestamps = append(timestamps, block.Timestamp)

		// Ensure some time passes between blocks
		time.Sleep(100 * time.Millisecond)
	}

	// Verify timestamps are monotonically increasing
	for i := 1; i < len(timestamps); i++ {
		if timestamps[i] <= timestamps[i-1] {
			t.Errorf("Block timestamps not increasing: block[%d]=%d, block[%d]=%d",
				i-1, timestamps[i-1], i, timestamps[i])
		}
	}

	// Verify block timestamps match their index in the chain
	for i := 1; i < len(node.Chain.Blocks); i++ {
		block := node.Chain.Blocks[i]
		if block.Index != i {
			t.Errorf("Block index mismatch: expected %d, got %d", i, block.Index)
		}
	}
}

func TestLedgerDataRetrieval(t *testing.T) {
	// Create a node and populate with election and votes
	node := utils.SetupTestNode()

	// Create test election
	election, _ := utils.CreateTestElection("Presidential Election 2025", []string{"Alice", "Bob", "Charlie"})
	electionTx, _ := utils.CreateElectionTransaction(election)

	// Add election transaction and create block
	node.TransactionPool = append(node.TransactionPool, electionTx)
	node.CreateBlock()

	// Create votes with specific distribution: 3 for Alice, 2 for Bob, 1 for Charlie
	votes := map[string]int{
		"Alice":   3,
		"Bob":     2,
		"Charlie": 1,
	}

	for candidate, count := range votes {
		for i := 0; i < count; i++ {
			ballot, _ := utils.CreateTestVote(election, candidate)
			voteTx, _ := utils.CreateVoteTransaction(election.ID, ballot)
			node.TransactionPool = append(node.TransactionPool, voteTx)
		}
	}

	// Create block with votes
	node.CreateBlock()

	// Simulate election result tallying by scanning the blockchain
	results := make(map[string]int)

	// Find the election first
	var electionData election.Election
	for _, block := range node.Chain.Blocks {
		for _, tx := range block.Transactions {
			if tx.Type == blockchain.TxCreateElection {
				err := json.Unmarshal(tx.Payload, &electionData)
				if err == nil && electionData.ID == election.ID {
					// Found the election, now count votes
					for _, block := range node.Chain.Blocks {
						for _, tx := range block.Transactions {
							if tx.Type == blockchain.TxCastVote {
								var voteData struct {
									ElectionID string           `json:"election_id"`
									Ballot     *election.Ballot `json:"ballot"`
								}

								if err := json.Unmarshal(tx.Payload, &voteData); err == nil {
									if voteData.ElectionID == election.ID {
										// In a real implementation, we would decrypt the vote
										// For this test, we'll use a mock decryption
										candidateIndex := mockDecryptVote(voteData.Ballot)
										if candidateIndex >= 0 && candidateIndex < len(electionData.Candidates) {
											candidateName := electionData.Candidates[candidateIndex].Name
											results[candidateName]++
										}
									}
								}
							}
						}
					}
					break
				}
			}
		}
	}

	// Verify vote counts match expected distribution
	for candidate, expectedCount := range votes {
		actualCount := results[candidate]
		if actualCount != expectedCount {
			t.Errorf("Vote count mismatch for %s: expected %d, got %d",
				candidate, expectedCount, actualCount)
		}
	}

	t.Logf("Successfully retrieved and tallied votes from blockchain: %v", results)
}

// Mock function to simulate vote decryption
// In a real implementation, this would use homomorphic decryption
func mockDecryptVote(ballot *election.Ballot) int {
	// For testing purposes, we'll determine the candidate based on a hash of the ballot data
	hash := ballot.ZKProof[0] % 3 // Use first byte of proof to determine candidate (0, 1, or 2)
	return int(hash)
}

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
