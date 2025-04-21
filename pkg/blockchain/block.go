// pkg/blockchain/block.go
package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

type Block struct {
	Index        int
	Timestamp    string
	Transactions []Transaction
	PrevHash     []byte
	Hash         []byte
	Nonce        int
}

func (b *Block) CalculateHash() []byte {
	data, _ := json.Marshal(b.Transactions)
	blockData := append(data, b.PrevHash...)
	hash := sha256.Sum256(blockData)
	return hash[:]
}
