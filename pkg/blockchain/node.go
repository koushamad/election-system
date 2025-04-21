package blockchain

import (
	"sync"
)

type Node struct {
	Chain           *Chain
	Peers           []string
	mu              sync.Mutex
	TransactionPool []*Transaction
}

func NewNode() *Node {
	return &Node{
		Chain: NewChain(),
		Peers: make([]string, 0),
	}
}

func (n *Node) AddPeer(peer string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Peers = append(n.Peers, peer)
}

func (n *Node) CreateBlock() *Block {
	n.mu.Lock()
	defer n.mu.Unlock()

	block := &Block{
		Index:        len(n.Chain.Blocks),
		Transactions: n.TransactionPool,
	}

	if len(n.Chain.Blocks) > 0 {
		block.PrevHash = n.Chain.Blocks[len(n.Chain.Blocks)-1].Hash
	}
	block.Hash = block.CalculateHash()
	n.Chain.Blocks = append(n.Chain.Blocks, block)
	n.TransactionPool = []*Transaction{}
	return block
}
