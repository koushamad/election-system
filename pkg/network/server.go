// pkg/network/server.go
package network

import (
	"election-system/pkg/blockchain"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	Node   *blockchain.Node
	Port   int
	P2PNet *P2PNetwork
}

func NewServer(node *blockchain.Node, port int) *Server {
	server := &Server{
		Node: node,
		Port: port,
	}

	nodeAddr := fmt.Sprintf("localhost:%d", port)
	server.P2PNet = NewP2PNetwork(nodeAddr, node)

	return server
}

func (s *Server) Start() error {
	// Chain endpoints
	http.HandleFunc("/chain", s.handleGetChain)
	http.HandleFunc("/blocks", s.handleBlocks)
	http.HandleFunc("/transactions", s.handleTransactions)

	// P2P endpoints
	http.HandleFunc("/peers", s.handlePeers)
	http.HandleFunc("/addPeer", s.handleAddPeer)

	// Start P2P sync
	s.P2PNet.StartSyncLoop()

	// Load initial peers from config
	s.loadInitialPeers()

	fmt.Printf("Server running on port %d\n", s.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
}

func (s *Server) handleGetChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.Node.Chain)
}

func (s *Server) handleBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var block blockchain.Block
		if err := json.NewDecoder(r.Body).Decode(&block); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Verify and add block
		if err := s.Node.AddBlock(&block); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var tx blockchain.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Verify and add transaction
		if err := s.Node.AddTransaction(&tx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Broadcast to peers
		s.P2PNet.BroadcastTransaction(&tx)

		w.WriteHeader(http.StatusCreated)
	} else if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.Node.TransactionPool)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePeers(w http.ResponseWriter, r *http.Request) {
	s.P2PNet.mu.RLock()
	defer s.P2PNet.mu.RUnlock()

	peers := make([]string, 0, len(s.P2PNet.KnownPeers))
	for peer := range s.P2PNet.KnownPeers {
		peers = append(peers, peer)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

func (s *Server) handleAddPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Peer string `json:"peer"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.P2PNet.AddPeer(req.Peer)
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) loadInitialPeers() {
	// This would normally load from config.yaml
	// For simplicity, we'll hardcode some peers
	initialPeers := []string{
		"localhost:5001",
		"localhost:5002",
	}

	for _, peer := range initialPeers {
		if peer != s.P2PNet.NodeAddress {
			s.P2PNet.AddPeer(peer)
		}
	}
}
