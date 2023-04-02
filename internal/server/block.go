package server

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"log"

	"github.com/yanglinshu/glock/internal/block"
	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/errors"
	"github.com/yanglinshu/glock/internal/transaction"
	"github.com/yanglinshu/glock/internal/util"
)

// Block is the block command
type Block struct {
	AddrFrom string // the address of the node
	Block    []byte // the block data
}

// sendBlock sends the block to the known nodes
func sendBlock(addr string, b *block.Block) error {
	sl, err := b.Serialize()
	if err != nil {
		return err
	}

	payload, err := util.GobEncode(Block{nodeAddress, sl})
	if err != nil {
		return err
	}

	request := append(commandToBytes("block"), payload...)

	err = sendData(addr, request)
	if err != nil {
		return err
	}

	return nil
}

// handleBlock handles the block command
func handleBlock(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload Block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	blockData := payload.Block
	bl, err := block.DeserializeBlock(blockData)
	if err != nil {
		return err
	}

	log.Printf("Received a new block!")
	bc.AddBlock(bl)

	log.Printf("Added block %x", bl.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := blockchain.UTXOSet{Blockchain: bc}
		UTXOSet.Reindex()
	}

	return nil
}

// Tx is the transaction
type Tx struct {
	AddrFrom    string
	Transaction []byte
}

// sendTx sends the transaction to the known nodes
func sendTx(addr string, tx *transaction.Transaction) error {
	sl, err := tx.Serialize()
	if err != nil {
		return err
	}

	data := Tx{nodeAddress, sl}
	payload, err := util.GobEncode(data)
	if err != nil {
		return err
	}
	request := append(commandToBytes("tx"), payload...)

	sendData(addr, request)
	return nil
}

// handleTx handles the tx command
func handleTx(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload Tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	txData := payload.Transaction
	tx, err := transaction.DeserializeTransaction(txData)
	if err != nil {
		return err
	}

	// Save the transaction to the mempool
	txID := hex.EncodeToString(tx.ID)
	mempool[txID] = tx

	if nodeAddress == knownNodes[0] { // If this is the coordinator node
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*transaction.Transaction
			for id := range mempool {
				tx := mempool[id]
				if ok, err := bc.VerifyTransaction(&tx); err != nil {
					return err
				} else if ok {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				log.Println("All transactions are invalid. Waiting for new transactions")
				return nil
			}

			cbTx, err := transaction.NewCoinbaseTX(miningAddress, "")
			if err != nil {
				return err
			}
			txs = append(txs, cbTx)

			// Create a new block containing the transactions
			newBlock, err := bc.MineBlock(txs)
			if err != nil {
				return err
			}

			UTXOSet := blockchain.UTXOSet{Blockchain: bc}
			UTXOSet.Reindex()

			log.Printf("New block is mined: %x", newBlock.Hash)

			// Clear the mempool
			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			// Broadcast the new block to all the nodes
			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
	return nil
}

// Inv shows other nodes what blocks or transactions it has
type Inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

// sendInv sends the inventory to the known nodes
func sendInv(addr, kind string, items [][]byte) error {
	inventory := Inv{nodeAddress, kind, items}
	payload, err := util.GobEncode(inventory)
	if err != nil {
		return err
	}

	request := append(commandToBytes("inv"), payload...)
	sendData(addr, request)
	return nil
}

// handleInv handles the inv command
func handleInv(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload Inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	log.Printf("Received inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if !bytes.Equal(b, blockHash) {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.Items[0])
		if mempool[txID].ID == nil {
			sendGetData(payload.AddrFrom, "tx", payload.Items[0])
		}
	}

	return nil
}

// GetBlocks is the getblocks command
type GetBlocks struct {
	AddrFrom string // the address of the node
}

// requestBlocks requests the blocks from the known nodes
func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}

// sendGetBlocks sends the getblocks command to the given address
func sendGetBlocks(addr string) error {
	payload, err := util.GobEncode(GetBlocks{nodeAddress})
	if err != nil {
		return err
	}

	request := append(commandToBytes("getblocks"), payload...)
	sendData(addr, request)
	return nil
}

// handleGetBlocks handles the getblocks command
func handleGetBlocks(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload GetBlocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	blocks, err := bc.GetBlockHashes()
	if err != nil {
		return err
	}
	sendInv(payload.AddrFrom, "block", blocks)

	return nil
}

// GetData is a message that requests data from another node
type GetData struct {
	AddrFrom string // the address of the node that sent the message
	Type     string // the type of data requested
	ID       []byte // the ID of the data requested
}

// sendGetData sends a GetData message to the given address
func sendGetData(addr, kind string, id []byte) error {
	payload, err := util.GobEncode(GetData{nodeAddress, kind, id})
	if err != nil {
		return err
	}

	request := append(commandToBytes("getdata"), payload...)
	sendData(addr, request)
	return nil
}

// handleGetData handles a GetData message
func handleGetData(request []byte, bc *blockchain.Blockchain) error {
	var buff bytes.Buffer
	var payload GetData

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	if payload.Type == "block" { // if the data requested is a block
		block, err := bc.GetBlock(payload.ID)
		if err != nil {
			return err
		}
		sendBlock(payload.AddrFrom, block)
	} else if payload.Type == "tx" { // if the data requested is a transaction
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]
		sendTx(payload.AddrFrom, &tx)
	} else {
		return errors.ErrUnknownGetDataType
	}

	return nil
}
