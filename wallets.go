package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"myBtc/lib/base58"
	"os"
)

const walletFile  =  "wallet.dat"

//定义钱包

type Wallets struct{
	WalletsMap map[string]*Wallet //key为地址
}

func NewWallets() *Wallets{
	var ws Wallets
	ws.WalletsMap  = make(map[string]*Wallet)
	ws.loadFile()
	return &ws
}

func (ws *Wallets) CreateWallet()string{
	wallet := NewWallet()
	address := wallet.NewAddress()
	ws.WalletsMap[address] = wallet
	ws.saveToFile()
	return address
}

//保存钱包
func (ws *Wallets)saveToFile(){
	var buffer bytes.Buffer
	//钱包中包含接口，需要先注册接口才能序列化
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&buffer)
	err  := encoder.Encode(ws)
	if err != nil{
		log.Panic(err)
	}

	ioutil.WriteFile(walletFile,buffer.Bytes(),0600)

}

//读取钱包
func (ws *Wallets)loadFile(){
	_,err := os.Stat(walletFile)
	if os.IsNotExist(err){
		return
	}

	//
	content ,err := ioutil.ReadFile(walletFile)
	if err != nil{
		return
	}
	//解码
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(content))
	var wsLocal Wallets
	err  = decoder.Decode(&wsLocal)
	if err != nil{
		log.Panic(err)
	}

	ws.WalletsMap = wsLocal.WalletsMap
	return
}

//获取钱包所有地址

func (ws *Wallets) ListAllAddresses()[]string{
	addresses := make([]string,0)
	for ad := range ws.WalletsMap{
		addresses = append(addresses,ad)
	}
	return  addresses
}

//根据地址得到公钥hash
func GetPubKeyByAddress (address string)[]byte{
	addressByte := base58.Decode(address)
	return addressByte[1:len(addressByte)-4]
}



