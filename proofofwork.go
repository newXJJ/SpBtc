package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
)

type ProofOfWork struct{
	block *Block
	target *big.Int
}


func NewProofOfWork(block *Block,difficulty ... string )*ProofOfWork{
	pow := ProofOfWork{
		block:  block,
		target: nil,
	}

	//指定难度值
	targetStr := "0000200000000000000000000000000000000000000000000000000000000000"
	if len(difficulty) > 1 && difficulty[0] != ""{
		targetStr = difficulty[0]
	}
	tmpInt := big.Int{}
	tmpInt.SetString(targetStr,16)
	pow.target = &tmpInt
	return &pow
}

func (pow *ProofOfWork) Run() ([]byte, uint64) {

	var nonce uint64
	block := pow.block
	var hash [32]byte

	fmt.Println("开始挖矿...")
	for {
		tmp := [][]byte{
			Uint64ToByte(block.Version),
			block.PrevHash,
			block.MerkelRoot,
			Uint64ToByte(block.TimeStamp),
			Uint64ToByte(block.Difficulty),
			Uint64ToByte(nonce),
		}

		//将二维的切片数组链接起来，返回一个一维的切片
		blockInfo := bytes.Join(tmp, []byte{})

		//2. 做哈希运算
		//func Sum256(data []byte) [Size]byte {
		hash = sha256.Sum256(blockInfo)
		//3. 与pow中的target进行比较
		tmpInt := big.Int{}
		//将我们得到hash数组转换成一个big.int
		tmpInt.SetBytes(hash[:])
		if tmpInt.Cmp(pow.target) == -1 {
			fmt.Printf("挖矿成功！hash : %x, nonce : %d\n", hash, nonce)
			//break
			return hash[:], nonce
		} else {
			nonce++
		}

	}

	//return []byte("HelloWorld"), 10
}

func Uint64ToByte(num uint64)[]byte{
	var buffer bytes.Buffer
	err := binary.Write(&buffer,binary.BigEndian,num)
	if err != nil{
		log.Panic(err)
	}
	return buffer.Bytes()
}