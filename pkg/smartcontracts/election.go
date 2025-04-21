// pkg/smartcontracts/election.go
package smartcontracts

import (
	"election-system/pkg/blockchain"
	"election-system/pkg/election"
)

type ElectionContract struct {
	Chain *blockchain.Chain
}

func (ec *ElectionContract) CreateElection(e *election.Election) error {
	tx := blockchain.Transaction{
		Type: "create_election",
		Data: e,
	}
	return ec.Chain.AddTransaction(tx)
}
