package main

import (
	"fmt"
	"log"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/Eric-GreenComb/parse-eth-balance/common"
	"github.com/Eric-GreenComb/parse-eth-balance/config"
	"github.com/Eric-GreenComb/parse-eth-balance/parser"
	"github.com/Eric-GreenComb/parse-eth-balance/persist"
)

func main() {

	session, err := mgo.Dial(config.MongoDB.Host)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	mongo := persist.Mongo{}
	mongo.SetCollection(session.DB(config.MongoDB.DB).C(config.MongoDB.Block), session.DB(config.MongoDB.DB).C(config.MongoDB.Token))

	_startNum := uint64(config.Ethereum.BlockNum)

	fmt.Println("start block num : ", _startNum)

	sync := make(chan int, 1)
	go mongo.Sync(_startNum, parser.GetLatestValidBlockNumber(), sync)

	// 周期同步
	for {
		select {
		case <-sync:
			log.Println("syncing task is completed.")
			time.Sleep(time.Duration(config.Server.Timer) * time.Second) // TODO: using event listen
		}
	}
}

func parseTx() {
	fmt.Println(parser.GetLatestValidBlockNumber())

	block := common.Block{}

	resp, err := parser.Call(config.Ethereum.Host, "eth_getBlockByNumber", []interface{}{config.Ethereum.BlockNum, true})
	if err != nil {
		log.Fatal(err)
	}

	if err := parser.MapToObject(resp.Result, &block); err != nil {
		log.Fatalln(err)
	}

	for _, _tx := range block.TXs {

		if _tx.To == config.Ethereum.TokenAddress {
			_addr, _value, err := parser.ParseTokenTransfer(_tx.Input)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			if _addr != config.Ethereum.ToAddress {
				continue
			}

			fmt.Println(_tx.From, _addr, _value)
		}

	}
}
