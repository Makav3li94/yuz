package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	dbPath      = "./tmp/blocks_%s"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

//DBExists checks if there is a blockchain ? or we should create genesis
func DBExists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}
	return true
}

//ContinueBlockChain if Database exists. we retrieve last hash and put it in database instance
func ContinueBlockChain(nodeId string) *BlockChain {
	// if there is no blockchain exit !
	path := fmt.Sprintf(dbPath, nodeId)
	if DBExists(path) == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	db, err := openDB(path, opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()

		return err
	})
	Handle(err)

	chain := BlockChain{lastHash, db}

	return &chain
}

//InitBlockChain 1-if there is not a blockchain create one with given address
//2-after thad reward for mining genesis
//3- change last hash value to genesis block last hash ( genesis block is the recent block that mined here)
func InitBlockChain(address, nodeId string) *BlockChain {
	path:= fmt.Sprintf(dbPath,nodeId)
	if DBExists(path) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}
	var lastHash []byte
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	db, err := openDB(path, opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		genesis := CreateGenesisBlock(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err

	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(block *Block) {
	err := chain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		Handle(err)

		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.Value()

		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			Handle(err)
			chain.LastHash = block.Hash
		}

		return nil
	})
	Handle(err)
}

func (chain *BlockChain) GetBestHeight() int {
	var lastBlock Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.Value()

		lastBlock = *Deserialize(lastBlockData)

		return nil
	})
	Handle(err)

	return lastBlock.Height
}

//GetBlock based on block hash
func (chain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block is not found")
		} else {
			blockData, _ := item.Value()

			block = *Deserialize(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := chain.Iterator()

	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}


// MineBlock adds a block to blockchain with CreateBlock
//gets last hash from db
//creates a new block with new data and last hash
func (chain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("Invalid Transaction")
		}
	}

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.Value()

		lastBlock := Deserialize(lastBlockData)

		lastHeight = lastBlock.Height

		return err
	})
	Handle(err)

	newBlock := CreateBlock(transactions, lastHash, lastHeight+1)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)

	return newBlock
}

//FindUTXO uses FindUnspentTransactions to get unspent transactions
//and separate them by outputs
func (chain *BlockChain) FindUTXO() map[string]TxOutPuts {
	UTXO := make(map[string]TxOutPuts)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.OutPuts = append(outs.OutPuts, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}

//FindTransaction find transaction by transaction id
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

//SignTransaction sign transaction by sending them to Sign
// we should send a map of transaction
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

//VerifyTransaction sign transaction by sending them to Verify
// we should send a map of transaction
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}
