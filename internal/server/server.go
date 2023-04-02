package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/errors"
	"github.com/yanglinshu/glock/internal/transaction"
)

// protocol is the protocol used to communicate with other nodes
const protocol = "tcp"

// nodeAddress is the address of the current node
var nodeAddress string

// miningAddress is the address of the miner
var miningAddress string

// knownNodes is the list of known nodes
var knownNodes = []string{"localhost:5000"}

// blocksInTransit is the list of blocks that are being downloaded
var blocksInTransit = [][]byte{}

// mempool is the list of transactions that are waiting to be mined
var mempool = make(map[string]transaction.Transaction)

func StartServer(nodeID, minerAddress string) error {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		return err
	}
	defer ln.Close()

	bc, err := blockchain.NewBlockchain(nodeID)
	if err != nil {
		return err
	}

	// send version to known nodes to get the latest blockchain
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn, bc)
	}
}

// handleConnection handles the connection
func handleConnection(conn net.Conn, bc *blockchain.Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	command := bytesToCommand(request[:commandLength])
	log.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		err := handleAddr(request)
		if err != nil {
			log.Println(err)
		}
	case "block":
		err := handleBlock(request, bc)
		if err != nil {
			log.Println(err)
		}
	case "inv":
		err := handleInv(request, bc)
		if err != nil {
			log.Println(err)
		}
	case "getblocks":
		err := handleGetBlocks(request, bc)
		if err != nil {
			log.Println(err)
		}
	case "getdata":
		err := handleGetData(request, bc)
		if err != nil {
			log.Println(err)
		}
	case "tx":
		err := handleTx(request, bc)
		if err != nil {
			log.Println(err)
		}
	case "version":
		err := handleVersion(request, bc)
		if err != nil {
			log.Println(err)
		}
	default:
		log.Println(errors.ErrUnknownCommand)
	}

	conn.Close()
}

// nodeIsKnown checks if the node is known
func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}
	return false
}

// sendData sends data to a node
func sendData(addr string, data []byte) error {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes
		return err
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		return err
	}

	return nil
}

// sendTransaction sends a transaction to the network
func SendTransaction(tnx *transaction.Transaction) {
	sendTx(knownNodes[0], tnx)
}
