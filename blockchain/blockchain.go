package blockchain

import (
	"fmt"
	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./tmp/blocks"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}
// BlockChainIterator is for doing stuff and fucking around in blockchain data
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// InitBlockChain initialize the blockchain creates genesis to fire up !
func InitBlockChain() *BlockChain {
	var lastHash []byte

	//connect to database
	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	db, err := badger.Open(opts)
	Handle(err)

	//check for blockchain database.
	//if there is one. get last hash for exploring
	// if not create genesis block and add it to data base
	// ls is for last hash key for db
	//badger db works with keys and values !
	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No blockchain found")
			genesis := CreateGenesisBlock()
			fmt.Println("Genesis Created")

			err = txn.Set(genesis.Hash, genesis.Serialize())
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			lastHash, err = item.Value()
			return err
		}
	})

	Handle(err)
	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

// AddBlockToBlockchain adds a block to blockchain with CreateBlock
//gets last hash from db
//creates a new block with new data and last hash
func (chain *BlockChain) AddBlockToBlockchain(data string) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()

		return err
	})
	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	Handle(err)
}

//Iterator we need this to explore in blockchain
// all data are stored in database. so we need to
//  get database and last hash to fill the BlockChainIterator struct
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
	}
	return iter
}

//Next with BlockChainIterator and getting db and last hash
//it will get the next blocks one by one
func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		encodedBlock, err := item.Value()
		block = Deserialize(encodedBlock)
		return err
	})
	Handle(err)
	// changes current hash to last hash to get the next block and goes on
	iter.CurrentHash = block.PrevHash
	return block
}
