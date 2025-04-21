package main

import (
	"election-system/pkg/blockchain"
	"election-system/pkg/crypto"
	"election-system/pkg/election"
	"election-system/pkg/network"
	"flag"
	"fmt"
	"log"
)

func main() {
	nodeCmd := flag.NewFlagSet("node", flag.ExitOnError)
	port := nodeCmd.Int("port", 5000, "Port number for the node")

	createElectionCmd := flag.NewFlagSet("create-election", flag.ExitOnError)
	electionName := createElectionCmd.String("name", "", "Election name")
	candidates := createElectionCmd.String("candidates", "", "Comma-separated list of candidates")

	voteCmd := flag.NewFlagSet("vote", flag.ExitOnError)
	electionID := voteCmd.String("election", "", "Election ID")
	candidate := voteCmd.String("candidate", "", "Candidate name")

	switch flag.Arg(0) {
	case "node":
		nodeCmd.Parse(flag.Args()[1:])
		startNode(*port)
	case "create-election":
		createElectionCmd.Parse(flag.Args()[1:])
		createElection(*electionName, *candidates)
	case "vote":
		voteCmd.Parse(flag.Args()[1:])
		castVote(*electionID, *candidate)
	default:
		fmt.Println("Usage:")
		fmt.Println("  node --port <port>")
		fmt.Println("  create-election --name <name> --candidates <c1,c2>")
		fmt.Println("  vote --election <id> --candidate <name>")
	}
}

func startNode(port int) {
	node := blockchain.NewNode()
	server := network.NewServer(node, port)
	fmt.Printf("Node running on port %d\n", port)
	log.Fatal(server.Start())
}

func createElection(name, candidates string) {
	// Implementation for election creation
}

func castVote(electionID, candidate string) {
	// Implementation for voting
}
