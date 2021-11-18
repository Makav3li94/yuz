package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

//Block Transactions are actually block`s data
type Block struct {
	Timestamp    int64
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
	Height       int
}


//HashTransactions hashes all transactions(data) as one hash
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}
	tree := NewMerkleTree(txHashes)
	return tree.RootNode.Data
}

// CreateBlock This function creates a new block using data and CreateHash() function
func CreateBlock(txs []*Transaction, prevHash []byte,height int) *Block {
	block := &Block{time.Now().Unix(),[]byte{}, txs, prevHash, 0,height}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// CreateGenesisBlock create genesis block
// the genesis block is the first block of blockchain and doesn't have PrevHash. so we must create it manually

func CreateGenesisBlock(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{},0)
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)
	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var b Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&b)
	Handle(err)
	return &b
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
