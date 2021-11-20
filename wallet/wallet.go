package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)
//0x00 is 0
const (
	checkSumLength = 4
	version        = byte(0x00)
)

//Wallet ; ecdsa is for digital signature
// a complex algorithm :|
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}


//1-Take the public key and hash it twice with RIPEMD160(SHA256(PubKey)) hashing algorithms.
//2-Prepend the version of the address generation algorithm to the hash.
//3-Calculate the checksum by hashing the result of step 2 with SHA256(SHA256(payload)). The checksum is the first four bytes of the resulted hash.
//4-Append the checksum to the version+PubKeyHash combination.
//5-Encode the version+PubKeyHash+checksum combination with Base58.



//Address is last part of algorithm for creating address (4)
func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)
	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)
	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	return address
}


//NewKeyPair ;  creates public key and private key
//elliptic  uses and x y algorithm to generate key
//it uses an elliptic cure
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	//elliptic is 256 bytes
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pub
}

//MakeWallet sets wallet pub and private with MakeWallet
func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

//PublicKeyHash is a part of algorithm for creating address (1)
func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}
	// sum hashes with ripemd160
	// it takes a parameter to concatenate, we don`t want to, so we pass nil
	publicRipMD := hasher.Sum(nil)

	return publicRipMD
}

//Checksum is a part of algorithm for creating address (2); part 3 is in utils
//it generate is first four bytes of hash
// its for verifying transactions
func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checkSumLength]
}

// ValidateAddress reverse generating address algorithm to check checksum
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualCheckSum := pubKeyHash[len(pubKeyHash)-checkSumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checkSumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualCheckSum, targetChecksum) == 0
}