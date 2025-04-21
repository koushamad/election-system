// pkg/blockchain/chain.go
package blockchain

import (
	"errors"
	"time"
)

type Chain struct {
	Blocks []*Block
}

func NewChain() *Chain {
	return &Chain{
		Blocks: []*Block{GenesisBlock()},
	}
}

func GenesisBlock() *Block {
	return &Block{
		Index:     0,
		Timestamp: time.Now().Unix(), // Changed from String() to Unix() to get int64
		Hash:      []byte("genesis-hash"),
	}
}

// Add the missing AddBlock method
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
	// Add transaction validation logic here
	return nil
}
