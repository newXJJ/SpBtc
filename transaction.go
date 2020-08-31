package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const reward   = 50

//交易结构
type Transaction struct {
	TXID []byte //交易ID
	TXInputs []TXInput //交易输入
	TXOutputs []TXOutput //交易输出
}

//交易输入
type TXInput struct{
	TXid []byte //引用的交易ID
	Index int64 //交易索引号，就是交易数组中的第几个
	Signature []byte //数字签名，有r,s组成
	Pubkey []byte //公钥
}

//交易输出
type TXOutput struct{
	Value float64 //交易金额
	PubKeyHash []byte //收款方的公钥hash

}

//模拟锁定脚本
func (ot *TXOutput) Lock(address string) {
	//从地址得到公钥hash
	ot.PubKeyHash = GetPubKeyByAddress(address)
}

//创建输出
func NewTXOutput(value float64,address string) *TXOutput{
	output := TXOutput{
		Value:      value,
		PubKeyHash: nil,
	}
	output.Lock(address)
	return &output
}

//设置交易ID
func (tx *Transaction) SetHash(){
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil{
		log.Panic("err")
	}
	data := buffer.Bytes()
	hash := sha256.Sum256(data)
	tx.TXID = hash[:]
}



//创建挖矿交易
func NewCoinbaseTx(address string, data string) *Transaction{
	//挖矿交易只有一个输入，没有交易ID，index设为-1
	input := TXInput{
		TXid:      nil,
		Index:     -1,
		Signature: nil,
		Pubkey:    []byte(data),
	}
	output := NewTXOutput(reward,address)
	//
	tx := &Transaction{
		[]byte{},
		[]TXInput{input},
		[]TXOutput{*output},
	}
	tx.SetHash()
	return tx
}

//创建普通交易
func NewTransaction(from,to string ,amount float64, bc *BlockChain) *Transaction{
	//打开钱包，找到自己的私钥公钥
	ws := NewWallets()
	wallet := ws.WalletsMap[from]
	if wallet == nil{
		fmt.Println("钱包中没有这个地址，地址无效")
		return nil
	}
	pub := wallet.PubKey
	priv := wallet.Private

	//交易里面传递的是公钥hash
	pubKeyHash := HashPubkey(pub)

	//找到合适的UTXO集合
	utxos,resValue := bc.FindNeedUTXOs(pubKeyHash,amount)
	if resValue < amount{
		fmt.Println("余额不足")
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput

	//构造交易输入
	for id ,indexArray := range utxos {
		for _,i  := range indexArray{
			input := TXInput{
				[]byte(id),int64(i),nil,pub,
			}
			inputs = append(inputs,input)
		}
	}
	//构造输出
	output := NewTXOutput(amount,to)
	outputs = append(outputs,*output)

	//找零
	if resValue > amount{
		output = NewTXOutput(resValue - amount,from)
		outputs = append(outputs,*output)
	}

	tx := Transaction{[]byte{},inputs,outputs}
	tx.SetHash()
	bc.SignTransaction(&tx,priv)
	return  &tx

}


//判断是否是挖矿交易
func (tx *Transaction) IsCoinbase()bool{
	//挖矿交易只有一个输入，交易ID为空，index为-1
	if len(tx.TXInputs)  ==1 && len(tx.TXInputs[0].TXid)  == 0 && tx.TXInputs[0].Index == -1{
		return true
	}
	return false
}

//签名
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey,prevTXs map[string]Transaction){
	if tx.IsCoinbase(){
		return
	}
	//创建交易副本
	txCopy := tx.TrimmedCopy()
	//
	for i,input := range txCopy.TXInputs{
		prevTX := prevTXs[string(input.TXid)]
		if len(prevTX.TXID) == 0{
			log.Panic("无效交易")
		}
		txCopy.TXInputs[i].Pubkey = prevTX.TXOutputs[input.Index].PubKeyHash
		txCopy.SetHash()

		txCopy.SetHash()
		txCopy.TXInputs[i].Pubkey = nil
		signDataHash := txCopy.TXID
		r,s,err := ecdsa.Sign(rand.Reader,privateKey,signDataHash)
		if err != nil{
			log.Panic(err)
		}
		signature := append(r.Bytes(),s.Bytes()...)
		tx.TXInputs[i].Signature = signature
	}
}

func (tx *Transaction) TrimmedCopy()Transaction{
	var inputs []TXInput
	var outputs []TXOutput
	for _,input := range tx.TXInputs{
		inputs = append (inputs,TXInput{input.TXid,input.Index,nil,nil})
	}
	for _,output := range tx.TXOutputs{
		outputs = append(outputs,output)
	}
	return Transaction{tx.TXID,inputs,outputs}
}


//验证
func (tx *Transaction)Verify(prevTXs map[string]Transaction)bool{
	if tx.IsCoinbase(){
		return true
	}

	txCopy := tx.TrimmedCopy()
	for i ,input := range tx.TXInputs{
		prevTx := prevTXs[string(input.TXid)]
		if len(prevTx.TXID) == 0{
			log.Panic("无效交易")
		}
		txCopy.TXInputs[i].Pubkey = prevTx.TXOutputs[input.Index].PubKeyHash
		txCopy.SetHash()
		dataHash := txCopy.TXID
		signature := input.Signature //用来反推r，s
		pubKey := input.Pubkey//这里的pubkey不是原始的pubkey，需要拼接
		//
		r,s := big.Int{},big.Int{}
		r.SetBytes(signature[:len(signature)/2])
		s.SetBytes(signature[len(signature)/2:])
		//
		x,y := big.Int{},big.Int{}
		x.SetBytes(pubKey[:len(pubKey)/2])
		y.SetBytes(pubKey[len(pubKey)/2:])
		//
		pubkeyOrigin := ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     &x,
			Y:     &y,
		}
		if !ecdsa.Verify(&pubkeyOrigin,dataHash,&r,&s){
			return false
		}
	}
	return  true
}


func (tx *Transaction)String()string{
	var lines []string
	lines  = append(lines,fmt.Sprintf("----Transaction %x:",tx.TXID))

	for i,input := range tx.TXInputs{
		lines = append(lines,fmt.Sprintf("\t\tInput      %d: \n",i))
		lines = append(lines,fmt.Sprintf("\t\tTXID:      %x\n",input.TXid))
		lines = append(lines,fmt.Sprintf("\t\tOut :      %d\n",input.Index))
		lines = append(lines,fmt.Sprintf("\t\tSignature: %x\n",input.Signature))
		lines = append(lines,fmt.Sprintf("\t\tPubKey:    %x\n",input.Pubkey))
	}

	for i,output:= range tx.TXOutputs{
		lines = append(lines,fmt.Sprintf("\t\tOutput %d\n",i))
		lines = append(lines,fmt.Sprintf("\t\tValut  %f\n",output.Value))
		lines = append(lines,fmt.Sprintf("\t\tScript: %x\n",output.PubKeyHash))
	}
	return strings.Join(lines," ")
}

