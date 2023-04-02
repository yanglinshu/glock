package transaction

import "bytes"

// TXInput represents a transaction input. It contains the ID of the transaction that contains the
// output, the index of the output in the transaction, and the signature of the input. The signature
// is used to verify that the owner of the output is the one spending it.
type TXInput struct {
	Txid      []byte // Txid is the ID of the transaction that contains the output
	Vout      int    // Vout is the index of the output in the transaction
	Signature []byte // Signature is the signature of the input
	PublicKey []byte // PublicKey is the public key of the owner of the output
}

// UsesKey checks whether the address is the owner of the output.
func (in *TXInput) UsesKey(pubKeyHash []byte) (bool, error) {
	lockingHash, err := HashPubKey(in.PublicKey)
	if err != nil {
		return false, err
	}

	return bytes.Equal(lockingHash, pubKeyHash), nil
}
