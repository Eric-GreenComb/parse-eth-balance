package persist

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/Eric-GreenComb/parse-eth-balance/common"
	"github.com/Eric-GreenComb/parse-eth-balance/config"
	"github.com/Eric-GreenComb/parse-eth-balance/parser"
)

// Mongo Mongo
type Mongo struct {
	Block *mgo.Collection
	Token *mgo.Collection
}

// SetCollection SetCollection
func (m *Mongo) SetCollection(block *mgo.Collection, token *mgo.Collection) *Mongo {
	m.Block = block
	m.Token = token
	return m
}

// InsertBlockInfo InsertBlockInfo
func (m *Mongo) InsertBlockInfo(block interface{}) error {
	if err := m.Block.Insert(block); err != nil {
		return err
	}
	return nil
}

// GetSyncedBlockCount GetSyncedBlockCount
func (m *Mongo) GetSyncedBlockCount() uint64 {
	result := common.MBlock{}
	m.Block.Find(bson.M{}).Sort("-number").Limit(1).One(&result)
	return uint64(result.Number)
}

// InsertTokenTransfer InsertTokenTransfer
func (m *Mongo) InsertTokenTransfer(tokenTransfer interface{}) error {
	if err := m.Token.Insert(tokenTransfer); err != nil {
		return err
	}
	return nil
}

// HasTokenTransfer HasTokenTransfer
func (m *Mongo) HasTokenTransfer(blockNum int64, hash, address string) error {
	result := common.MTransaction{}
	err := m.Token.Find(bson.M{"from": address, "hash": hash}).Limit(1).One(&result)
	return err
}

// Sync Sync
func (m *Mongo) Sync(syncedNumber, latestBlock uint64, c chan int) {
	block := common.Block{}
	if syncedNumber > 0 {
		// 从下一个块开始同步
		syncedNumber++
	}

	var notinFile *os.File
	notin := "./notin.log"
	if checkFileIsExist(notin) { //如果文件存在
		notinFile, _ = os.OpenFile(notin, os.O_APPEND, 0666) //打开文件
		fmt.Println("文件存在")
	} else {
		notinFile, _ = os.Create(notin) //创建文件
		fmt.Println("文件不存在")
	}
	defer notinFile.Close()

	var inFile *os.File
	in := "./in.log"
	if checkFileIsExist(in) { //如果文件存在
		inFile, _ = os.OpenFile(in, os.O_APPEND, 0666) //打开文件
		fmt.Println("文件存在")
	} else {
		inFile, _ = os.Create(in) //创建文件
		fmt.Println("文件不存在")
	}
	defer inFile.Close()

	for i := syncedNumber; i <= latestBlock; i++ {

		number := fmt.Sprintf("0x%s", strconv.FormatUint(uint64(i), 16))
		resp, err := parser.Call(config.Ethereum.Host, "eth_getBlockByNumber", []interface{}{number, true})
		if err != nil {
			log.Fatal(err)
		}

		if err := parser.MapToObject(resp.Result, &block); err != nil {
			log.Fatalln(err)
		}

		fmt.Println("block : ", i, block.Number, len(block.TXs))

		for _, _tx := range block.TXs {

			if strings.ToLower(_tx.To) == config.Ethereum.TokenAddress {

				_addr, _value, err := parser.ParseTokenTransfer(_tx.Input)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}

				if strings.ToLower(_addr) != config.Ethereum.ToAddress {
					continue
				}
				fmt.Println(">>>>>>>> _addr == config.Ethereum.ToAddress")

				_blockNum, err := strconv.ParseInt(_tx.BlockNumber, 10, 64)
				err = m.HasTokenTransfer(_blockNum, _tx.Hash, _tx.From)
				if err != nil {
					fmt.Println("========= not in : ", _tx.From)

					notinLog := log.New(notinFile, "[NotIn]", log.Ldate)
					_msg := _tx.Hash + " " + _tx.From + " " + _addr + " " + _value
					notinLog.Println(_msg)
				} else {
					fmt.Println("--------- in : ", _tx.From)

					inLog := log.New(inFile, "[In]", log.Ldate)
					_msg := _tx.Hash + " " + _tx.From + " " + _addr + " " + _value
					inLog.Println(_msg)
				}
			}
		}
	}

	c <- 1
}

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
