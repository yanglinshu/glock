package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/yanglinshu/glock/internal/util"
	"golang.org/x/crypto/ripemd160"
)

// version is the current version of the wallet
const version = byte(0x00)

// walletFileFormat is the format of the wallet file
const walletFileFormat = "wallet_%s.dat"

// addressChecksumLen is the length of the checksum in the address
const addressChecksumLen = 4

// Wallet stores a private and public key
type Wallet struct {
	PrivateKey ecdsa.PrivateKey // Private key
	PublicKey  []byte           // Public key
}

// NewWallet creates and returns a Wallet
func NewWallet() (*Wallet, error) {
	private, public, err := newKeyPair()
	if err != nil {
		return nil, err
	}

	wallet := Wallet{private, public}

	return &wallet, nil
}

// newKeyPair generates a private and public key
func newKeyPair() (ecdsa.PrivateKey, []byte, error) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return ecdsa.PrivateKey{}, nil, err
	}

	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey, nil
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	pubKeyHash := util.Base58Decode([]byte(address))

	if len(pubKeyHash)-addressChecksumLen < 0 {
		return false
	}

	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]

	if len(pubKeyHash)-addressChecksumLen < 1 {
		return false
	}

	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Equal(actualChecksum, targetChecksum)
}

// GetAddress returns wallet address: a hash of the public key. Address contains the version, the
// public key hash, and a checksum.
func (w Wallet) GetAddress() ([]byte, error) {
	pubKeyHash, err := HashPubKey(w.PublicKey)
	if err != nil {
		return nil, err
	}

	versionedHash := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	addr := util.Base58Encode(fullHash)

	return addr, nil
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) ([]byte, error) {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		return nil, err
	}

	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160, nil
}

// checksum returns a checksum for a public key hash
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

// Wallets stores a collection of wallets.
type Wallets struct {
	Wallets map[string]*Wallet // Wallets
}

// NewWallets creates a new wallet
func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile(nodeID)

	return &wallets, err
}

// CreateWallet creates a new wallet
func (ws *Wallets) CreateWallet() (string, error) {
	wallet, err := NewWallet()
	if err != nil {
		return "", err
	}

	address, err := wallet.GetAddress()
	if err != nil {
		return "", err
	}

	ws.Wallets[string(address)] = wallet

	return string(address), nil
}

// GetAddresses returns all addresses from the collection of wallets
func (ws Wallets) GetAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a wallet by its address
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFromFile loads wallets from file
func (ws *Wallets) LoadFromFile(nodeID string) error {
	walletFile := fmt.Sprintf(walletFileFormat, nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveToFile saves wallets to file
func (ws Wallets) SaveToFile(nodeID string) error {
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		return err
	}

	walletFile := fmt.Sprintf(walletFileFormat, nodeID)
	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
