package cli

import (
	"fmt"

	"github.com/yanglinshu/glock/internal/errors"
	"github.com/yanglinshu/glock/internal/server"
	"github.com/yanglinshu/glock/internal/transaction"
)

// startNode creates a new node
func startNode(minerAddress, nodeID string) error {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if transaction.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			return errors.ErrInvalidAddress
		}
	}

	err := server.StartServer(nodeID, minerAddress)
	if err != nil {
		return err
	}

	fmt.Println("Shutting down node ", nodeID)
	return nil
}
