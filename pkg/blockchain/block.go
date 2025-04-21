// pkg/blockchain/block.go
package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

type Block struct {
	Index        int            `json:"index"`
	Timestamp    int64          `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
	PrevHash     []byte         `json:"prev_hash"`
	Hash         []byte         `json:"hash"`
	Nonce        int            `json:"nonce"`
	Validator    string         `json:"validator"` // Address of the validator who created this block
}

func NewBlock(index int, transactions []*Transaction, prevHash []byte, validator string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prevHash,
		Validator:    validator,
	}

	block.Hash = block.CalculateHash()
	return block
}

func (b *Block) CalculateHash() []byte {
	blockData := struct {
		Index        int
		Timestamp    int64
		Transactions [][32]byte // Using transaction hashes for efficiency
		PrevHash     []byte
		Validator    string
		Nonce        int
	}{
		Index:     b.Index,
		Timestamp: b.Timestamp,
		PrevHash:  b.PrevHash,
		Validator: b.Validator,
		Nonce:     b.Nonce,
	}

	// Extract transaction hashes
	txHashes := make([][32]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		copy(txHashes[i][:], tx.Hash)
	}
	blockData.Transactions = txHashes

	data, _ := json.Marshal(blockData)
	hash := sha256.Sum256(data)
	return hash[:]
}
