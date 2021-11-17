package blockchain

import (
	"bytes"
	"github.com/Makav3li94/yuz/wallet"
)

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

//TxInput refrence to out puts.
// is for transaction hash
//out is index of specific output
type TxInput struct {
	ID        []byte
	Out       int
	Signature []byte
	PubKey    []byte
}



func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutPut(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}
