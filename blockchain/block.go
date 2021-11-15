package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

// CreateBlock This function creates a new block using data and CreateHash() function
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash, 0}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// CreateGenesisBlock create genesis block
// the genesis block is the first block of blockchain and doesn't have PrevHash. so we must create it manually

func CreateGenesisBlock() *Block {
	return CreateBlock("This is Genesis Block", []byte{})
}

func (b *Block) Serialize() []byte  {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)
	return res.Bytes()
}

func Deserialize(data []byte) *Block  {
	var b Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&b)
	Handle(err)
	return &b
}

func Handle(err error) {
	if err != nil {
		log.Panic()
	}
}