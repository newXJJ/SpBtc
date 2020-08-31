package main

import (
	"fmt"
	"os"
	"strconv"
)

//操作区块链


type CLI struct{
	bc *BlockChain
}

const Help  = `
	printChain				"打印区块链"
	printChainR				"反向打印"
	getBalance --address  ADDRESS  "获取余额"
	send FROM TO AMOUNT MINER DATA "FROM转AMOUNT给TO，MINER挖矿，写入数据DATA"
	newWallet "创建一个新的钱包"
	listAddresses "打印所有地址"`


func (cli *CLI)Run(){
	args := os.Args
	if len(args) < 2{
		fmt.Println(Help)
		return
	}
	//
	cmd := args[1]
	switch cmd {
	case "printChain":
		cli.PrintBlockChain()
	case "printChainR":
		cli.PrintBlockChainR()
	case "getBalance":
		if len(args) == 4 && args[2]== "--address"{
			fmt.Println("address : ",args[3])
			//fmt.Println("余额 : ",)
			cli.GetBalance(args[3])
		}

	case "send":
		fmt.Println("转账开始")
		if len(args) != 7{
			fmt.Println("参数错误")
			fmt.Println(Help)
			break
		}
		from,to,miner,data := args[2],args[3],args[5],args[6]
		amount,_ := strconv.ParseFloat(args[4],64)
		cli.Send(from,to,amount,miner,data)

	case "newWallet":
		fmt.Println("创建新的钱包")
		cli.NewWallet()
	case "listAddresses":
		fmt.Println("所有地址")
		cli.ListAddresses()
	case "Help":
		fmt.Println(Help)
	default:
		fmt.Println("命令错误")
		fmt.Println(Help)
	}
	return
}
