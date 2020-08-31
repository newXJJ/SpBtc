package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

//定义区块

type Block struct {
	Version uint64 //版本号
	PrevHash []byte //前hash
	MerkelRoot []byte //梅克尔根，验证数据完整性
	TimeStamp uint64 //时间戳
	Difficulty uint64 //挖矿难度值
	Nonce uint64 //基于难度值的目标随机值
	Hash []byte //当前区块hash
	Transactions []*Transaction //交易数据
}


func NewBlock(txs []*Transaction, PrevBlockHash []byte) *Block{
	block := &Block{
		00,
		PrevBlockHash,
		[]byte{},
		uint64(time.Now().Unix()),
		0,
		0,
		[]byte{},
		txs,
	}

	block.MerkelRoot = block.MakeMerkelRoot()

	pow := NewProofOfWork(block)
	hash,nonce := pow.Run()

	block.Hash = hash
	block.Nonce = nonce
	return  block
}

//序列化
func (block *Block)Serialize()[]byte{
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(&block)
	if err != nil{
		log.Panic("编码出错")
	}
	return buffer.Bytes()
}

//反序列化
func Deserialize(data []byte)Block{
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var block Block
	err  := decoder.Decode(&block)
	if err != nil{
		//fmt.Println(data)
		log.Panic(err)
		BtcLog("解码出错")
	}
	return block
}

//模拟生成梅克尔根
func (block *Block) MakeMerkelRoot()[]byte{
	var info []byte
	for _,tx := range block.Transactions{
		info = append(info,tx.TXID...)
	}
	hash := sha256.Sum256(info)
	return hash[:]
}