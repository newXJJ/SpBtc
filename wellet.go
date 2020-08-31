package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
	"myBtc/lib/base58"
	"golang.org/x/crypto/ripemd160"
)

//定义钱包


type Wallet struct {
	Private *ecdsa.PrivateKey //私钥，椭圆曲线
	PubKey []byte //公钥，不是原始公钥，是通过原始公钥得到的字符数组
}

//创建钱包
func NewWallet()*Wallet{
	curve := elliptic.P256()
	privateKey,err := ecdsa.GenerateKey(curve,rand.Reader)
	if err != nil{
		log.Panic("生成私钥失败")
	}
	pubkeyOrigin := privateKey.PublicKey
	//拼接公钥
	pubkey := append(pubkeyOrigin.X.Bytes(),pubkeyOrigin.Y.Bytes()...)
	return &Wallet{
		Private: privateKey,
		PubKey:  pubkey,
	}
}

//生成地址，地址是根据公钥得到的
func (w *Wallet) NewAddress()string{
	//地址是有版本号+公钥hash+4字节校验组成
	pub := w.PubKey
	pubHash := HashPubkey(pub)
	version := byte(00)
	payload := append([]byte{version},pubHash...)
	checkCode := CheckSum(payload)
	payload = append(payload,checkCode...)
	address := base58.Encode(payload)
	return address
}

func HashPubkey(data[]byte) []byte{
	hash := sha256.Sum256(data)
	//编码
	rip160hasher := ripemd160.New()
	_,err := rip160hasher.Write(hash[:])
	if err != nil{
		log.Panic(err)
	}

	ripHash := rip160hasher.Sum(nil)
	return ripHash
}

func CheckSum (data []byte) []byte{
	//两次hash取前4个字节
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	return h2[:4]
}

//验证地址有效
func IsValidAddress(address string)bool{
	//解码
	 addressByte := base58.Decode(address)
	 if len(addressByte) < 4{
	 	return  false
	 }

	 payload := addressByte[:len(addressByte)-4]
	 checksum1 := addressByte[len(addressByte)-4:]
	 checksum2 := CheckSum(payload)
	 return bytes.Equal(checksum1,checksum2)

}

