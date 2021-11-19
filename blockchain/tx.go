package blockchain

import (
	"bytes"
	"encoding/gob"
	"github.com/Makav3li94/yuz/wallet"
)

//TxOutput Value is amount
//PubKeyHash is for locking it for an address
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

//TxOutPuts we want to only check unspent transaction for doing stuff
//for optimization
type TxOutPuts struct {
	OutPuts []TxOutput
}

//TxInput refrence to out puts.
// is for transaction hash
//out is index of specific output
// Tx input give permissions to create transactions with Signature and PubKey
type TxInput struct {
	ID        []byte
	Out       int
	Signature []byte
	PubKey    []byte
}

//UsesKey compares PubKey in output to PubKeyHash in input
// it`s for input
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

//Lock Passing PubKeyHash to output by decoding address
// that
func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

//IsLockedWithKey checks if PubKeyHash exists
// a simple validation
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

//NewTXOutPut generates output with Lock func and locks it ***
func NewTXOutPut(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

//SerializeOutput for TxOutPuts
func (outs TxOutPuts) SerializeOutput() []byte {
	var buffer bytes.Buffer
	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(outs)
	Handle(err)
	return buffer.Bytes()
}

//DeserializeOutput for TxOutPuts
func DeserializeOutput(data []byte) TxOutPuts {
	var outputs TxOutPuts
	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&outputs)
	Handle(err)
	return outputs
}
