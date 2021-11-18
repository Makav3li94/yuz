package blockchain

import "github.com/dgraph-io/badger"

// BlockChainIterator is for doing stuff  in blockchain data
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

//Iterator we need this to explore in blockchain
// all data are stored in database. so we need to
//  get database and last hash to fill the BlockChainIterator struct
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}

//Next with BlockChainIterator and getting db and last hash
//it will get the next blocks one by one
func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodedBlock, err := item.Value()
		block = Deserialize(encodedBlock)
		return err
	})
	Handle(err)
	// changes current hash to last hash to get the next block and goes on
	iter.CurrentHash = block.PrevHash
	return block
}
