package main

import (
	"fmt"
	"log"
	"runtime"
)

func BtcLog(v interface{}){
	_,file,line,_ := runtime.Caller(1)
	s := fmt.Sprintf("file :%s,line %d",file,line)
	log.Panic(v,s)
}


