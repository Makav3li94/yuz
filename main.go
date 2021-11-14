package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

type BlockChain struct {
	blocks []*Block
}

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

// CreateHash This function creates hash for current block.
func (b *Block) CreateHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}

// CreateBlock This function creates a new block using data and CreateHash() function
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash}
	block.CreateHash()
	return block
}


// AddBlockToBlockchain adds a block to blockchain with  CreateBlock()

func (chain *BlockChain) AddBlockToBlockchain(data string) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
	chain.blocks = append(chain.blocks, new)
}


// CreateGenesisBlock create genesis block
// the genesis block is the first block of blockchain and doesn't have PrevHash. so we must create it manually

func CreateGenesisBlock() *Block {
	return CreateBlock("This is Genesis Block", []byte{})
}

// initBlockChain initialize the blockchain creates genesis to fire up !
func initBlockChain() *BlockChain {
	return &BlockChain{[]*Block{CreateGenesisBlock()}}
}

func main() {
	chain := initBlockChain()

	chain.AddBlockToBlockchain("First")
	chain.AddBlockToBlockchain("Second")
	chain.AddBlockToBlockchain("Third")

	for _, block := range chain.blocks {
		fmt.Printf("Prev Hash %x \n", block.PrevHash)
		fmt.Printf("Data in Block %s \n", block.Data)
		fmt.Printf("Current Hash %x \n", block.Hash)
	}
}
