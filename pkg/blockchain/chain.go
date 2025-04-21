package blockchain

import (
	"crypto/sha256"
	"encoding/json"
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
		Timestamp: time.Now().String(),
		Hash:      []byte("genesis-hash"),
	}
}

func (c *Chain) AddTransaction(tx *Transaction) error {
	if !tx.Validate() {
		return errors.New("invalid transaction")
	}
	// Add transaction validation logic here
	return nil
}
