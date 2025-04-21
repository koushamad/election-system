// pkg/blockchain/node.go
package blockchain

import (
	"bytes"
	"errors"
	"sync"
)

type Node struct {
	Chain           *Chain
	Peers           []string
	mu              sync.RWMutex
	TransactionPool []*Transaction
	Address         string // Node's blockchain address for validation
	IsValidator     bool   // Whether this node is a validator
}

func NewNode() *Node {
	return &Node{
		Chain:           NewChain(),
		Peers:           make([]string, 0),
		TransactionPool: make([]*Transaction, 0),
	}
}

func (n *Node) AddTransaction(tx *Transaction) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Verify transaction
	if !tx.Validate() {
		return errors.New("invalid transaction")
	}

	// Check for duplicates
	for _, t := range n.TransactionPool {
		if bytes.Equal(t.Hash, tx.Hash) {
			return errors.New("transaction already exists in pool")
		}
	}

	n.TransactionPool = append(n.TransactionPool, tx)

	// If we have enough transactions and we're a validator, create a block
	if len(n.TransactionPool) >= 5 && n.IsValidator {
		go n.CreateBlock()
	}

	return nil
}

func (n *Node) CreateBlock() *Block {
	n.mu.Lock()
	defer n.mu.Unlock()

	if len(n.TransactionPool) == 0 {
		return nil
	}

	prevBlock := n.Chain.Blocks[len(n.Chain.Blocks)-1]
	newBlock := NewBlock(
		prevBlock.Index+1,
		n.TransactionPool,
		prevBlock.Hash,
		n.Address,
	)

	n.Chain.AddBlock(newBlock)
	n.TransactionPool = []*Transaction{}

	return newBlock
}
