package server

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/util"
)

// nodeVersion is the current version of the node
const nodeVersion = 1

// Version is the version of the node
type Version struct {
	Version    int    // version of the node
	BestHeight int    // the best height of the blockchain
	AddrFrom   string // the address of the node
}

// handleVersion handles the version command
func handleVersion(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload Version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	myBestHeight, err := bc.GetBestHeight()
	if err != nil {
		return err
	}

	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}

	return nil
}

// handleConnection handles the connection
func sendVersion(addr string, bc *blockchain.Blockchain) error {
	bestHeight, err := bc.GetBestHeight()
	if err != nil {
		return err
	}

	payload, err := util.GobEncode(Version{nodeVersion, bestHeight, nodeAddress})
	if err != nil {
		return err
	}

	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
	return nil
}

type Addr struct {
	AddrList []string
}

// sendAddr sends the address
func sendAddr(addr string) error {
	nodes := Addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload, err := util.GobEncode(nodes)
	if err != nil {
		return err
	}

	request := append(commandToBytes("addr"), payload...)

	err = sendData(addr, request)
	if err != nil {
		return err
	}

	return nil
}

// handleAddr handles the address
func handleAddr(request []byte) error {
	var buff bytes.Buffer
	var payload Addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	log.Printf("There are %d known nodes now", len(knownNodes))
	return nil
}
