package network

import (
	"election-system/pkg/blockchain"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	Node *blockchain.Node
	Port int
}

func NewServer(node *blockchain.Node, port int) *Server {
	return &Server{Node: node, Port: port}
}

func (s *Server) Start() error {
	http.HandleFunc("/chain", s.handleGetChain)
	http.HandleFunc("/transactions", s.handleTransactions)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
}

func (s *Server) handleGetChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.Node.Chain)
}

func (s *Server) handleTransactions(w http.ResponseWriter, r *http.Request) {
	var tx blockchain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.Node.Chain.AddTransaction(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Transaction added successfully")
}
