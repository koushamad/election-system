package utils

import (
	"encoding/json"
	"fmt"
	"github.com/koushamad/election-system/pkg/blockchain"
	"github.com/koushamad/election-system/pkg/crypto"
	"github.com/koushamad/election-system/pkg/election"
	"math/big"
	"time"

	"github.com/cloudflare/bn256"
)

// SetupTestBlockchain creates a blockchain with genesis block for testing
func SetupTestBlockchain() *blockchain.Chain {
	return blockchain.NewChain()
}

// SetupTestNode creates a node with initialized blockchain for testing
func SetupTestNode() *blockchain.Node {
	return blockchain.NewNode()
}

// CreateTestElection creates an election for testing purposes
func CreateTestElection(name string, candidates []string) (*election.Election, *crypto.KeyPair) {
	// Generate election keys
	electionKeys := crypto.GenerateKeys()

	// Create candidate objects
	electionCandidates := make([]election.Candidate, len(candidates))
	for i, name := range candidates {
		electionCandidates[i] = election.Candidate{
			ID:   fmt.Sprintf("candidate-%d", i+1),
			Name: name,
		}
	}

	// Create election with future dates (relative to April 2025)
	now := time.Date(2025, 4, 21, 12, 0, 0, 0, time.UTC)
	startTime := now.Add(24 * time.Hour)
	endTime := startTime.Add(48 * time.Hour)

	return &election.Election{
		ID:         fmt.Sprintf("election-%x", time.Now().Unix()),
		Name:       name,
		Candidates: electionCandidates,
		StartTime:  startTime,
		EndTime:    endTime,
		PublicKey:  electionKeys.PublicKey,
	}, electionKeys
}

// CreateTestVote creates a test vote for a specific candidate
func CreateTestVote(electionData *election.Election, candidateName string) (*election.Ballot, error) {
	// Find candidate index
	candidateIndex := 0
	found := false
	for i, candidate := range electionData.Candidates {
		if candidate.Name == candidateName {
			candidateIndex = i + 1 // 1-based index for voting
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("candidate '%s' not found", candidateName)
	}

	// Generate voter keys
	voterKeys := crypto.GenerateKeys()

	// Create encrypted vote
	ciphertext, err := crypto.EncryptVote(electionData.PublicKey, candidateIndex)
	if err != nil {
		return nil, err
	}

	// Generate random value for encryption (normally this would be saved)
	r, _ := rand.Int(rand.Reader, bn256.Order)

	// Generate zero-knowledge proof
	proof := crypto.GenerateVoteProof(ciphertext, r, candidateIndex)

	// Create ballot
	ballot := election.NewBallot(
		ciphertext,
		proof,
		fmt.Sprintf("%x", voterKeys.PublicKey.Marshal()[:8]),
	)

	return ballot, nil
}

// CreateElectionTransaction creates a transaction for election creation
func CreateElectionTransaction(electionData *election.Election) (*blockchain.Transaction, error) {
	electionJSON, err := json.Marshal(electionData)
	if err != nil {
		return nil, err
	}

	tx := &blockchain.Transaction{
		Type:    blockchain.TxCreateElection,
		Payload: electionJSON,
	}

	tx.Hash = tx.CalculateHash()
	return tx, nil
}

// CreateVoteTransaction creates a transaction for vote casting
func CreateVoteTransaction(electionID string, ballot *election.Ballot) (*blockchain.Transaction, error) {
	voteData := struct {
		ElectionID string           `json:"election_id"`
		Ballot     *election.Ballot `json:"ballot"`
	}{
		ElectionID: electionID,
		Ballot:     ballot,
	}

	voteJSON, err := json.Marshal(voteData)
	if err != nil {
		return nil, err
	}

	tx := &blockchain.Transaction{
		Type:    blockchain.TxCastVote,
		Payload: voteJSON,
	}

	tx.Hash = tx.CalculateHash()
	return tx, nil
}
