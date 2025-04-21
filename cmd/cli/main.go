// cmd/cli/main.go
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cloudflare/bn256"
	"github.com/koushamad/election-system/pkg/blockchain"
	"github.com/koushamad/election-system/pkg/crypto"
	"github.com/koushamad/election-system/pkg/election"
	"github.com/koushamad/election-system/pkg/network"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	// Command-line flags
	nodeCmd := flag.NewFlagSet("node", flag.ExitOnError)
	nodePort := nodeCmd.Int("port", 5000, "Port number for the node")
	nodeValidator := nodeCmd.Bool("validator", false, "Run as a validator node")

	createElectionCmd := flag.NewFlagSet("create-election", flag.ExitOnError)
	electionName := createElectionCmd.String("name", "", "Election name")
	candidatesStr := createElectionCmd.String("candidates", "", "Comma-separated list of candidates")
	startTime := createElectionCmd.String("start", "", "Start time (YYYY-MM-DD HH:MM)")
	endTime := createElectionCmd.String("end", "", "End time (YYYY-MM-DD HH:MM)")
	nodeAddr := createElectionCmd.String("node", "localhost:5000", "Node address to submit transaction")

	voteCmd := flag.NewFlagSet("vote", flag.ExitOnError)
	voteElectionID := voteCmd.String("election", "", "Election ID")
	voteCandidate := voteCmd.String("candidate", "", "Candidate name")
	voteNodeAddr := voteCmd.String("node", "localhost:5000", "Node address to submit vote")

	// Parse command
	if len(os.Args) < 2 {
		fmt.Println("Expected 'node', 'create-election', or 'vote' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "node":
		nodeCmd.Parse(os.Args[2:])
		startNode(*nodePort, *nodeValidator)
	case "create-election":
		createElectionCmd.Parse(os.Args[2:])
		if *electionName == "" || *candidatesStr == "" || *startTime == "" || *endTime == "" {
			fmt.Println("All flags are required: --name, --candidates, --start, --end")
			os.Exit(1)
		}
		createElection(*electionName, *candidatesStr, *startTime, *endTime, *nodeAddr)
	case "vote":
		voteCmd.Parse(os.Args[2:])
		if *voteElectionID == "" || *voteCandidate == "" {
			fmt.Println("All flags are required: --election, --candidate")
			os.Exit(1)
		}
		castVote(*voteElectionID, *voteCandidate, *voteNodeAddr)
	default:
		fmt.Println("Expected 'node', 'create-election', or 'vote' subcommands")
		os.Exit(1)
	}
}

func startNode(port int, isValidator bool) {
	// Generate node keys
	keyPair := crypto.GenerateKeys()

	// Initialize node
	node := blockchain.NewNode()
	node.IsValidator = isValidator
	node.Address = fmt.Sprintf("%x", keyPair.PublicKey.Marshal()[:8]) // Use first 8 bytes of public key as address

	// Start server
	server := network.NewServer(node, port)
	fmt.Printf("Node running on port %d (Validator: %v, Address: %s)\n",
		port, isValidator, node.Address)
	log.Fatal(server.Start())
}

func createElection(name, candidatesStr, startTimeStr, endTimeStr, nodeAddr string) {
	// Parse candidates
	candidates := strings.Split(candidatesStr, ",")
	if len(candidates) < 2 {
		fmt.Println("At least two candidates are required")
		os.Exit(1)
	}

	// Parse times
	startTime, err := time.Parse("2006-01-02 15:04", startTimeStr)
	if err != nil {
		fmt.Printf("Invalid start time format: %v\n", err)
		os.Exit(1)
	}

	endTime, err := time.Parse("2006-01-02 15:04", endTimeStr)
	if err != nil {
		fmt.Printf("Invalid end time format: %v\n", err)
		os.Exit(1)
	}

	// Generate election keys
	electionKeys := crypto.GenerateKeys()

	// Create election object
	electionID := fmt.Sprintf("election-%x", time.Now().Unix())
	electionCandidates := make([]election.Candidate, len(candidates))
	for i, name := range candidates {
		electionCandidates[i] = election.Candidate{
			ID:   fmt.Sprintf("candidate-%d", i+1),
			Name: name,
		}
	}

	newElection := election.Election{
		ID:         electionID,
		Name:       name,
		Candidates: electionCandidates,
		StartTime:  startTime,
		EndTime:    endTime,
		PublicKey:  electionKeys.PublicKey,
	}

	// Create transaction
	tx, err := blockchain.NewTransaction(blockchain.TxCreateElection, newElection)
	if err != nil {
		fmt.Printf("Failed to create transaction: %v\n", err)
		os.Exit(1)
	}

	// Sign transaction (in a real implementation)
	// tx.Sign(electionKeys.PrivateKey)

	// Submit to node
	txJSON, _ := json.Marshal(tx)
	resp, err := http.Post(fmt.Sprintf("http://%s/transactions", nodeAddr),
		"application/json", bytes.NewBuffer(txJSON))
	if err != nil {
		fmt.Printf("Failed to submit transaction: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Failed to create election: %s\n", body)
		os.Exit(1)
	}

	fmt.Printf("Election created successfully with ID: %s\n", electionID)
	fmt.Printf("Save your private key for tallying: %x\n", electionKeys.PrivateKey)
}

func castVote(electionID, candidateName, nodeAddr string) {
	// Generate voter keys
	voterKeys := crypto.GenerateKeys()

	// Get election details from the blockchain
	resp, err := http.Get(fmt.Sprintf("http://%s/elections/%s", nodeAddr, electionID))
	if err != nil {
		fmt.Printf("Failed to get election details: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Election not found")
		os.Exit(1)
	}

	var electionData election.Election
	json.NewDecoder(resp.Body).Decode(&electionData)

	// Find candidate ID
	var candidateID string
	for _, candidate := range electionData.Candidates {
		if candidate.Name == candidateName {
			candidateID = candidate.ID
			break
		}
	}

	if candidateID == "" {
		fmt.Printf("Candidate '%s' not found in election\n", candidateName)
		os.Exit(1)
	}

	// Create encrypted vote
	candidateIndex := 0
	for i, candidate := range electionData.Candidates {
		if candidate.ID == candidateID {
			candidateIndex = i + 1 // 1-based index for voting
			break
		}
	}

	ciphertext, err := crypto.EncryptVote(electionData.PublicKey, candidateIndex)
	if err != nil {
		fmt.Printf("Failed to encrypt vote: %v\n", err)
		os.Exit(1)
	}

	// Generate zero-knowledge proof
	r, _ := rand.Int(rand.Reader, bn256.Order) // Random value used in encryption
	proof := crypto.GenerateVoteProof(ciphertext, r, candidateIndex)

	// Create ballot
	ballot := election.Ballot{
		Ciphertext: ciphertext,
		ZKProof:    proof,
		VoterID:    fmt.Sprintf("%x", voterKeys.PublicKey.Marshal()[:8]), // Use first 8 bytes of public key as voter ID
	}

	// Create vote transaction
	voteData := struct {
		ElectionID string          `json:"election_id"`
		Ballot     election.Ballot `json:"ballot"`
	}{
		ElectionID: electionID,
		Ballot:     ballot,
	}

	tx, err := blockchain.NewTransaction(blockchain.TxCastVote, voteData)
	if err != nil {
		fmt.Printf("Failed to create transaction: %v\n", err)
		os.Exit(1)
	}

	// Sign transaction (in a real implementation)
	// tx.Sign(voterKeys.PrivateKey)

	// Submit to node
	txJSON, _ := json.Marshal(tx)
	resp, err = http.Post(fmt.Sprintf("http://%s/transactions", nodeAddr),
		"application/json", bytes.NewBuffer(txJSON))
	if err != nil {
		fmt.Printf("Failed to submit vote: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Failed to cast vote: %s\n", body)
		os.Exit(1)
	}

	fmt.Println("Vote cast successfully!")
	fmt.Printf("Your vote receipt: %x\n", ballot.ZKProof[:8]) // First 8 bytes as receipt
}
