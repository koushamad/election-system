// pkg/blockchain/chain.go
package blockchain

import (
	"bytes"
	"errors"
	"time"
)

type Chain struct {
	Blocks              []*Block
	PendingTransactions []*Transaction
}

func NewChain() *Chain {
	return &Chain{
		Blocks:              []*Block{GenesisBlock()},
		PendingTransactions: []*Transaction{},
	}
}

func GenesisBlock() *Block {
	return &Block{
		Index:     0,
		Timestamp: time.Now().Unix(),
		Hash:      []byte("genesis-hash"),
	}
}

func (c *Chain) AddBlock(block *Block) {
	if len(c.Blocks) > 0 {
		block.PrevHash = c.Blocks[len(c.Blocks)-1].Hash
	}
	block.Hash = block.CalculateHash()
	c.Blocks = append(c.Blocks, block)
}

func (c *Chain) AddTransaction(tx *Transaction) error {
	if !tx.Validate() {
		return errors.New("invalid transaction")
	}

	// Check for duplicates
	for _, t := range c.PendingTransactions {
		if bytes.Equal(t.Hash, tx.Hash) {
			return errors.New("transaction already exists")
		}
	}

	// Add to pending transactions
	c.PendingTransactions = append(c.PendingTransactions, tx)
	return nil
}
