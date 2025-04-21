// pkg/network/p2p.go
package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/koushamad/election-system/pkg/blockchain"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type P2PNetwork struct {
	NodeAddress string
	KnownPeers  map[string]bool
	mu          sync.RWMutex
	node        *blockchain.Node
}

func NewP2PNetwork(nodeAddr string, node *blockchain.Node) *P2PNetwork {
	return &P2PNetwork{
		NodeAddress: nodeAddr,
		KnownPeers:  make(map[string]bool),
		node:        node,
	}
}

func (p *P2PNetwork) AddPeer(peerAddr string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peerAddr != p.NodeAddress && !p.KnownPeers[peerAddr] {
		p.KnownPeers[peerAddr] = true
		go p.SyncWithPeer(peerAddr)
	}
}

func (p *P2PNetwork) BroadcastBlock(block *blockchain.Block) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for peer := range p.KnownPeers {
		go func(peerAddr string) {
			blockData, _ := json.Marshal(block)
			http.Post(fmt.Sprintf("http://%s/blocks", peerAddr),
				"application/json", bytes.NewBuffer(blockData))
		}(peer)
	}
}

func (p *P2PNetwork) BroadcastTransaction(tx *blockchain.Transaction) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for peer := range p.KnownPeers {
		go func(peerAddr string) {
			txData, _ := json.Marshal(tx)
			http.Post(fmt.Sprintf("http://%s/transactions", peerAddr),
				"application/json", bytes.NewBuffer(txData))
		}(peer)
	}
}

func (p *P2PNetwork) SyncWithPeer(peerAddr string) {
	// Get peer's blockchain
	resp, err := http.Get(fmt.Sprintf("http://%s/chain", peerAddr))
	if err != nil {
		fmt.Printf("Failed to sync with peer %s: %v\n", peerAddr, err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var peerChain blockchain.Chain
	json.Unmarshal(body, &peerChain)

	// Compare with our chain and resolve conflicts
	if len(peerChain.Blocks) > len(p.node.Chain.Blocks) {
		// Verify the peer's chain
		if p.node.VerifyChain(&peerChain) {
			p.node.ReplaceChain(&peerChain)
			fmt.Printf("Replaced chain with peer %s\n", peerAddr)
		}
	}

	// Get peer's known peers
	resp, err = http.Get(fmt.Sprintf("http://%s/peers", peerAddr))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	var peers []string
	json.Unmarshal(body, &peers)

	// Add new peers
	for _, peer := range peers {
		p.AddPeer(peer)
	}
}

func (p *P2PNetwork) StartSyncLoop() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			<-ticker.C
			p.mu.RLock()
			peers := make([]string, 0, len(p.KnownPeers))
			for peer := range p.KnownPeers {
				peers = append(peers, peer)
			}
			p.mu.RUnlock()

			for _, peer := range peers {
				go p.SyncWithPeer(peer)
			}
		}
	}()
}
