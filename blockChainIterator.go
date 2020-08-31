package main

import (
	"log"
	"myBtc/lib/bolt"
)

type BlockChainIterator struct {
	db *bolt.DB
	currentHashPointer []byte
}

func NewIterator(bc *BlockChain)*BlockChainIterator{
	return &BlockChainIterator{
		db:                 bc.db,
		currentHashPointer: bc.tail,
	}
}

//迭代器

func (it *BlockChainIterator)Next() *Block{
	var block Block
	it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil{
			log.Panic("区块为空")
		}
		blockTmp := bucket.Get(it.currentHashPointer)
		block = Deserialize(blockTmp)
		it.currentHashPointer = block.PrevHash
		return nil
	})
	return &block
}

