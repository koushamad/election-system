// pkg/blockchain/node.go
package blockchain

import (
	"errors"
	"sync"
	_ "time"
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
		if string(t.Hash) == string(tx.Hash) {
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

	// In a real implementation, we would sign the block here
	// newBlock.Sign(privateKey)

	n.Chain.AddBlock(newBlock)
	n.TransactionPool = []*Transaction{}

	return newBlock
}

func (n *Node) AddBlock(block *Block) error {
	// Verify block
	if !n.VerifyBlock(block) {
		return errors.New("invalid block")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	// Check if block already exists
	for _, b := range n.Chain.Blocks {
		if string(b.Hash) == string(block.Hash) {
			return errors.New("block already exists")
		}
	}

	n.Chain.Blocks = append(n.Chain.Blocks, block)

	// Remove transactions that are now in the block
	var newPool []*Transaction
	for _, tx := range n.TransactionPool {
		found := false
		for _, btx := range block.Transactions {
			if string(tx.Hash) == string(btx.Hash) {
				found = true
				break
			}
		}
		if !found {
			newPool = append(newPool, tx)
		}
	}
	n.TransactionPool = newPool

	return nil
}

func (n *Node) VerifyBlock(block *Block) bool {
	// Verify block hash
	if string(block.CalculateHash()) != string(block.Hash) {
		return false
	}

	// Verify block index and previous hash
	prevBlock := n.Chain.Blocks[len(n.Chain.Blocks)-1]
	if block.Index != prevBlock.Index+1 || string(block.PrevHash) != string(prevBlock.Hash) {
		return false
	}

	// Verify block signature (in a real implementation)
	// if !block.VerifySignature() {
	//     return false
	// }

	// Verify all transactions in the block
	for _, tx := range block.Transactions {
		if !tx.Validate() {
			return false
		}
	}

	return true
}

func (n *Node) VerifyChain(chain *Chain) bool {
	if len(chain.Blocks) <= 1 {
		return true // Only genesis block
	}

	for i := 1; i < len(chain.Blocks); i++ {
		block := chain.Blocks[i]
		prevBlock := chain.Blocks[i-1]

		// Verify block hash
		if string(block.CalculateHash()) != string(block.Hash) {
			return false
		}

		// Verify block links to previous block
		if block.Index != prevBlock.Index+1 || string(block.PrevHash) != string(prevBlock.Hash) {
			return false
		}

		// Verify all transactions
		for _, tx := range block.Transactions {
			if !tx.Validate() {
				return false
			}
		}
	}

	return true
}

func (n *Node) ReplaceChain(chain *Chain) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Chain = chain

	// Rebuild transaction pool
	// Remove transactions that are now in the blockchain
	var newPool []*Transaction
	for _, tx := range n.TransactionPool {
		found := false
		for _, block := range chain.Blocks {
			for _, btx := range block.Transactions {
				if string(tx.Hash) == string(btx.Hash) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			newPool = append(newPool, tx)
		}
	}
	n.TransactionPool = newPool
}
