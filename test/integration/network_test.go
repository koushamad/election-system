package integration

import (
	"bytes"
	"encoding/json"
	_ "fmt"
	"github.com/koushamad/election-system/pkg/blockchain"
	"github.com/koushamad/election-system/pkg/network"
	"github.com/koushamad/election-system/test/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerChainEndpoint(t *testing.T) {
	// Setup node and server
	node := utils.SetupTestNode()
	server := network.NewServer(node, 0) // Port 0 for testing

	// Create a test transaction and block
	election, _ := utils.CreateTestElection("Test Election", []string{"Alice", "Bob"})
	tx, _ := utils.CreateElectionTransaction(election)
	node.TransactionPool = append(node.TransactionPool, tx)
	node.CreateBlock()

	// Create test HTTP request
	req, err := http.NewRequest("GET", "/chain", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleGetChain)

	// Serve HTTP request
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	var responseChain blockchain.Chain
	err = json.Unmarshal(rr.Body.Bytes(), &responseChain)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify chain length
	if len(responseChain.Blocks) != 2 { // Genesis + 1 block
		t.Errorf("Expected 2 blocks in chain, got %d", len(responseChain.Blocks))
	}

	// Verify transaction in block
	if len(responseChain.Blocks[1].Transactions) != 1 {
		t.Errorf("Expected 1 transaction in block, got %d",
			len(responseChain.Blocks[1].Transactions))
	}
}

func TestServerTransactionEndpoint(t *testing.T) {
	// Setup node and server
	node := utils.SetupTestNode()
	server := network.NewServer(node, 0) // Port 0 for testing

	// Create a test transaction
	election, _ := utils.CreateTestElection("Test Election", []string{"Alice", "Bob"})
	tx, _ := utils.CreateElectionTransaction(election)

	// Convert transaction to JSON
	txJSON, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	// Create test HTTP request
	req, err := http.NewRequest("POST", "/transactions", bytes.NewBuffer(txJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleTransactions)

	// Serve HTTP request
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Verify transaction was added to the chain
	// Note: In the current implementation, transactions are validated but not stored
	// This would need to be updated when proper transaction pooling is implemented
}
