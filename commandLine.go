package main

import "fmt"

func (cli *CLI)PrintBlockChain(){
	cli.bc.Printchain()
	fmt.Println("区块链打印完成")
}

func(cli *CLI)PrintBlockChainR(){
	bc := cli.bc

	//反向打印是调用迭代器完成的
	it := NewIterator(bc)
	for {
		block := it.Next()
		for _,tx := range block.Transactions{
			//交易实现了String()方法
			fmt.Println(tx)
		}
		if len(block.PrevHash) == 0{
			fmt.Println("反向打印区块完成")
			break
		}
	}
}


func (cli *CLI)GetBalance(address string) {
	if !IsValidAddress(address){
		fmt.Println("无效地址")
		return
	}
	pubKeyHash := GetPubKeyByAddress(address)
	utxos := cli.bc.FindUTXOs(pubKeyHash)
	total := 0.0
	for _,utxo := range utxos{
		total += utxo.Value
	}
	fmt.Println("余额为 ：",total)
}

func (cli *CLI)Send (from,to string ,amount float64,miner,data string){
	if !IsValidAddress(from)|| !IsValidAddress(to) ||!IsValidAddress(miner){
		fmt.Println("存在无效地址")
	}

	//挖矿交易
	coinbase := NewCoinbaseTx(miner,data)
	tx := NewTransaction(from,to,amount,cli.bc)
	if tx == nil{
		fmt.Println("创建交易失败")
	}

	cli.bc.AddBlock([]*Transaction{coinbase,tx})
	fmt.Println("转账成功")
}


func (cli *CLI)NewWallet(){
	ws := NewWallets()
	address := ws.CreateWallet()
	fmt.Println("新地址 ：",address)
}

func (cli *CLI)ListAddresses(){
	ws := NewWallets()
	addresses := ws.ListAllAddresses()
	for _,address := range addresses{
		fmt.Println(address)
	}
}