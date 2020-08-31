package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"myBtc/lib/bolt"
)

type BlockChain struct{
	db *bolt.DB

	tail []byte //最后一个区块哈希
}

const blockChainDb  = "blockChain.db"
const blockBucket  = "blockBucket"

func NewBlockChain(add string) *BlockChain{
	var lastHs []byte


	db,err := bolt.Open(blockChainDb,0600,nil)
	if err != nil{
		log.Panic("open db fail")
	}

	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil{//没有数据就新建
			//
			fmt.Println("生成创世块")
			bucket,err  = tx.CreateBucket([]byte(blockBucket))
			if err != nil{
				log.Panic("create bucket fail")
			}
			//创世块
			genesisBlock := GenesisBlock(add)
			//保存创世块，key为hash，值为字节流
			bucket.Put(genesisBlock.Hash,genesisBlock.Serialize())
			bucket.Put([]byte("LastHashKey"),genesisBlock.Hash)
			lastHs = genesisBlock.Hash
		}else{
			lastHs = bucket.Get([]byte("LastHashKey"))
		}
		return nil
	})
	return &BlockChain{db,lastHs}
}

//创世块
func GenesisBlock(add string) *Block {
	coinbase := NewCoinbaseTx(add,"hello world")
	return NewBlock([]*Transaction{coinbase},[]byte{})
}

//添加区块
func (bc *BlockChain)AddBlock(txs []*Transaction){

	//fmt.Println("Transactions",len(txs))
	for _,tx := range txs{
		//fmt.Println("Transaction ",i)
		if !bc.VerifyTransaction(tx){
			fmt.Printf("发现无效交易:%v",tx)
			return
		}
	}
	db := bc.db
	lastHash := bc.tail
	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Println("bc.tail , ",hex.EncodeToString(lastHash))
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil{
			log.Panic("空bucket")
		}
		block := NewBlock(txs,lastHash)
		bucket.Put(block.Hash,block.Serialize())
		bucket.Put([]byte("LastHashKey"),block.Hash)

		bc.tail = block.Hash
		return nil
	})
}

func (bc *BlockChain)Printchain(){
	blockHeight := 0
	bc.db.View(func(tx *bolt.Tx) error {
		b:= tx.Bucket([]byte(blockBucket))
		b.ForEach(func(k, v []byte) error {
			if bytes.Equal(k,[]byte("LastHashKey")){
				return nil
			}

			block := Deserialize(v)
			fmt.Printf("------------区块高度： %d--------------\n",blockHeight)
			blockHeight++
			fmt.Printf("版本号: %d\n", block.Version)
			fmt.Printf("前区块哈希值: %x\n", block.PrevHash)
			fmt.Printf("梅克尔根: %x\n", block.MerkelRoot)
			fmt.Printf("时间戳: %d\n", block.TimeStamp)
			fmt.Printf("难度值(随便写的）: %d\n", block.Difficulty)
			fmt.Printf("随机数 : %d\n", block.Nonce)
			fmt.Printf("当前区块哈希值: %x\n", block.Hash)
			fmt.Printf("区块数据 :%s\n", block.Transactions[0].TXInputs[0].Pubkey)
			return nil

		})
		return nil
	})
}

//找到指定地址拥有的utxo
func (bc *BlockChain) FindUTXOs(pubKeyHash []byte)[]TXOutput{
	var UTXO []TXOutput
	txs := bc.FindUTXOTransactions(pubKeyHash)
	for _,tx := range txs{
		for _,output := range tx.TXOutputs{
			if bytes.Equal(pubKeyHash,output.PubKeyHash){
				UTXO = append(UTXO,output)
			}
		}
	}
	return UTXO
}

//找出合适的utxo
func (bc *BlockChain)FindNeedUTXOs(senderPubKeyHash []byte,amount float64) (map[string][]uint64,float64){
	utxos := make(map[string][]uint64)
	var calc float64
	txs := bc.FindUTXOTransactions(senderPubKeyHash)
	if amount <= 0{
		return utxos,calc
	}

	for _,tx := range txs {
		for i,output := range tx.TXOutputs{
			if bytes.Equal(senderPubKeyHash,output.PubKeyHash){
				//找到的钱累加
				if calc < amount{
					utxos[string(tx.TXID)] = append(utxos[string(tx.TXID)],uint64(i))
					calc += output.Value
					//z找到足够的钱
					if calc >= amount{
						fmt.Printf("找到合适金额 ：%f",calc)
						return utxos,calc
					}
				}
			}

		}
	}
	//走到这表示金额不够
	fmt.Printf("金额不足，当前金额：%f,需要金额：%f",calc,amount)
	return utxos,calc
}


func (bc *BlockChain)FindUTXOTransactions(senderPubKeyHash []byte)[]*Transaction{
	var txs []*Transaction

	//已经花了的输出
	spentOutput := make(map[string][]int64)
	it := NewIterator(bc)

	for{
		//遍历区块
		block  := it.Next()
		//遍历交易
		for _,tx := range block.Transactions{
		OUTPUT:
			//遍历一笔交易输出
			for i ,output := range tx.TXOutputs{
				//过滤消耗过多的utxo
				if spentOutput[string(tx.TXID)] != nil{
					for _,j := range spentOutput[string(tx.TXID)]{
						if int64(i) == j{
							continue OUTPUT
						}
					}
				}
				if bytes.Equal(output.PubKeyHash,senderPubKeyHash) {
					txs = append(txs, tx)
				}
				//}else{
				//	fmt.Println((output.PubKeyHash),"\t",(senderPubKeyHash))
				//	log.Panic("******不应该走到这")
				//}

			}
			if !tx.IsCoinbase(){
				//遍历输入，找到消耗过得utxo
				for _,input := range tx.TXInputs{
					pubKeyHash := HashPubkey(input.Pubkey)
					if bytes.Equal(pubKeyHash,senderPubKeyHash){
						spentOutput[string(input.TXid)] = append(spentOutput[string(input.TXid)],input.Index)
					}
				}
			}
		}
		if len(block.PrevHash) == 0{
			//区块遍历完成
			break
		}

	}
	return txs
}

func (bc *BlockChain)VerifyTransaction(tx *Transaction)bool{
	if tx.IsCoinbase(){
		return true
	}
	prevTXs := make(map[string]Transaction)
	for _,input := range tx.TXInputs{
		tx,err := bc.FindTransactionByTxid(input.TXid)
		if err != nil{
			log.Panic(err)
		}
		prevTXs[string(input.TXid)] = tx
	}

	return tx.Verify(prevTXs)
}

func (bc *BlockChain)FindTransactionByTxid(id []byte)(Transaction,error){

	it := NewIterator(bc)
	for{
		block := it.Next()
		for _,tx := range block.Transactions{
			if bytes.Equal(tx.TXID,id){
				return *tx,nil
			}
		}
		if len(block.PrevHash) == 0{
			//遍历区块结束
		}
	}
	return Transaction{},errors.New("无效交易ID")

}

//交易签名
func (bc *BlockChain) SignTransaction(tx *Transaction,privateKey *ecdsa.PrivateKey){
	prevTXs := make(map[string]Transaction)

	for _,input := range tx.TXInputs{
		tx,err := bc.FindTransactionByTxid(input.TXid)
		if err != nil{
			log.Panic(err)
		}
		prevTXs[string(input.TXid)] = tx

	}
	tx.Sign(privateKey,prevTXs)

}
