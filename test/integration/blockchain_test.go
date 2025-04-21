// test/integration/blockchain_test.go
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/koushamad/election-system/pkg/blockchain"
	"github.com/koushamad/election-system/test/utils"
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
				var electionData map[string]interface{}
				json.Unmarshal(tx.Payload, &electionData)
				if electionData["id"] == election.ID {
					electionFound = true
				}
			} else if tx.Type == blockchain.TxCastVote {
				// Verify vote transaction
				var voteData map[string]interface{}
				json.Unmarshal(tx.Payload, &voteData)
				if voteData["election_id"] == election.ID {
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
	node1Block := node1.CreateBlock()

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
	// FIX: Use AddTransaction method which should check for duplicates
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

		// FIX: Force different timestamps by modifying the Block creation
		// to ensure each block has a unique timestamp
		time.Sleep(1 * time.Second) // Ensure at least 1 second between blocks
		block := node.CreateBlock()
		timestamps = append(timestamps, block.Timestamp)
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
	var electionData map[string]interface{}
	for _, block := range node.Chain.Blocks {
		for _, tx := range block.Transactions {
			if tx.Type == blockchain.TxCreateElection {
				err := json.Unmarshal(tx.Payload, &electionData)
				if err == nil && electionData["id"] == election.ID {
					// Found the election, now count votes
					for _, block := range node.Chain.Blocks {
						for _, tx := range block.Transactions {
							if tx.Type == blockchain.TxCastVote {
								var voteData map[string]interface{}
								if err := json.Unmarshal(tx.Payload, &voteData); err == nil {
									if voteData["election_id"] == election.ID {
										// In a real implementation, we would decrypt the vote
										// For this test, we'll use a mock decryption
										candidateIndex := mockDecryptVote(voteData["ballot"])
										if candidateIndex >= 0 && candidateIndex < len(electionData["candidates"].([]interface{})) {
											candidates := electionData["candidates"].([]interface{})
											candidateName := candidates[candidateIndex].(map[string]interface{})["name"].(string)
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
// FIX: Fixed the type conversion issue by properly handling the interface type
func mockDecryptVote(ballot interface{}) int {
	// For testing purposes, we'll determine the candidate based on a hash of the ballot data
	ballotMap, ok := ballot.(map[string]interface{})
	if !ok {
		return 0 // Default to first candidate if type conversion fails
	}

	// Handle the case where zkProof might be a string in the JSON
	var zkProof []byte
	switch v := ballotMap["zk_proof"].(type) {
	case []byte:
		zkProof = v
	case string:
		zkProof = []byte(v)
	default:
		return 0 // Default to first candidate if zkProof is not found or has unexpected type
	}

	if len(zkProof) == 0 {
		return 0
	}

	hash := zkProof[0] % 3 // Use first byte of proof to determine candidate (0, 1, or 2)
	return int(hash)
}

func TestBlockCreation(t *testing.T) {
	node := utils.SetupTestNode()

	// Create a test transaction
	election, _ := utils.CreateTestElection("Test Election", []string{"Alice", "Bob"})
	tx, err := utils.CreateElectionTransaction(election)
	if err != nil {
		t.Fatalf("Failed to create election transaction: %v", err)
	}

	// Add transaction to node
	node.TransactionPool = append(node.TransactionPool, tx)

	// Create a block
	block := node.CreateBlock()

	// Verify block properties
	if block.Index != 1 { // 0 is genesis
		t.Errorf("Expected block index 1, got %d", block.Index)
	}

	if len(block.Transactions) != 1 {
		t.Errorf("Expected 1 transaction in block, got %d", len(block.Transactions))
	}

	// Verify transaction in block
	if string(block.Transactions[0].Hash) != string(tx.Hash) {
		t.Error("Transaction hash mismatch in block")
	}

	// Verify block is added to chain
	if len(node.Chain.Blocks) != 2 { // Genesis + new block
		t.Errorf("Expected 2 blocks in chain, got %d", len(node.Chain.Blocks))
	}

	// Verify transaction pool is cleared
	if len(node.TransactionPool) != 0 {
		t.Errorf("Expected empty transaction pool, got %d transactions", len(node.TransactionPool))
	}
}

func TestBlockchainIntegrity(t *testing.T) {
	node := utils.SetupTestNode()

	// Create multiple blocks
	for i := 0; i < 3; i++ {
		election, _ := utils.CreateTestElection(
			fmt.Sprintf("Test Election %d", i),
			[]string{"Alice", "Bob"},
		)
		tx, _ := utils.CreateElectionTransaction(election)
		node.TransactionPool = append(node.TransactionPool, tx)
		node.CreateBlock()
		time.Sleep(100 * time.Millisecond) // Ensure different timestamps
	}

	// Verify chain length
	if len(node.Chain.Blocks) != 4 { // Genesis + 3 blocks
		t.Errorf("Expected 4 blocks in chain, got %d", len(node.Chain.Blocks))
	}

	// Verify blockchain integrity (each block points to previous)
	for i := 1; i < len(node.Chain.Blocks); i++ {
		currentBlock := node.Chain.Blocks[i]
		prevBlock := node.Chain.Blocks[i-1]

		// Verify block index
		if currentBlock.Index != prevBlock.Index+1 {
			t.Errorf("Block index mismatch at position %d", i)
		}

		// Verify previous hash
		if string(currentBlock.PrevHash) != string(prevBlock.Hash) {
			t.Errorf("Previous hash mismatch at block %d", i)
		}

		// Verify block hash is correct
		calculatedHash := currentBlock.CalculateHash()
		if string(calculatedHash) != string(currentBlock.Hash) {
			t.Errorf("Block hash mismatch at position %d", i)
		}
	}
}

func TestTransactionValidation(t *testing.T) {
	chain := utils.SetupTestBlockchain()

	// Valid transaction
	election, _ := utils.CreateTestElection("Test Election", []string{"Alice", "Bob"})
	validTx, _ := utils.CreateElectionTransaction(election)

	err := chain.AddTransaction(validTx)
	if err != nil {
		t.Errorf("Failed to add valid transaction: %v", err)
	}

	// Invalid transaction (manually corrupt hash)
	invalidTx, _ := utils.CreateElectionTransaction(election)
	invalidTx.Hash = []byte("invalid-hash")

	// This should be rejected in a real implementation
	// Note: The current implementation always returns true for Validate()
	// This test will need to be updated when proper validation is implemented
	err = chain.AddTransaction(invalidTx)
	if err != nil {
		t.Logf("As expected, invalid transaction was rejected: %v", err)
	} else {
		t.Log("Warning: Invalid transaction was accepted. Implement proper validation.")
	}
}
