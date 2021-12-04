package blockchain

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

// 数据库路径。
const dbPath = "blockchain.db"
const blocksBucket = "blocks"

// 区块链结构。
type BlockChain struct {
	rear []byte   // 最后一个记录的哈希值。
	DB   *bolt.DB // 数据库连接。
}

// 创建区块链。
func CreateBlockChain() *BlockChain {
	var rear []byte

	// 打开数据库。
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	// 从数据库读取目前的区块链信息。
	err = db.Update(func(tx *bolt.Tx) error {
		var bucket *bolt.Bucket
		bucket = tx.Bucket([]byte(blocksBucket))
		// 如果已经有bucket，就直接读取。
		if bucket != nil {
			rear = bucket.Get([]byte("l"))
			return nil
		}
		// 否则，新建一个区块链的bucket。
		fmt.Println("No chain found. Creating a new one...")
		bucket, err = tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}
		// 录入创世块。
		genesis := CreateGenesisBlock()
		genesis.AddToBucket(bucket)
		// 记录创世块哈希值。
		rear = genesis.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &BlockChain{rear, db}
}

// 添加区块。
func (chain *BlockChain) AddBlock(data string) {
	// 从数据库获取最后一个区块的哈希值。
	var lastHash []byte
	err := chain.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		lastHash = bucket.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// 创建新区块。
	newBlock := CreateBlock(data, lastHash)
	err = chain.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		newBlock.AddToBucket(bucket)
		chain.rear = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// 区块链迭代器结构。
type BlockChainIterator struct {
	CurHash []byte   // 当前指向区块的哈希值。
	db      *bolt.DB // 数据库连接。
}

// 创建迭代器。
func (chain *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{chain.rear, chain.DB}
}

// 从尾部开始遍历区块链。
func (iter *BlockChainIterator) Next() *Block {
	// 获取迭代器当前指向的区块。
	var block *Block
	err := iter.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		blockSeq := bucket.Get(iter.CurHash)
		block = DeserializeBlock(blockSeq)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	// 迭代器移向前一个区块。
	iter.CurHash = block.PrevBlockHash
	return block
}
