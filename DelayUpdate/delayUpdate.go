package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"time"
)
import cuckoo "github.com/seiflotfy/cuckoofilter"

type Msg struct {
	key        string
	value      []byte
	timeStamp  int64
	isConsumed bool
}

type IDelayUpdate struct {
	UpdateQueue []Msg
	LCF         *cuckoo.Filter
	UCF         *cuckoo.Filter
	HashLock    map[string]string
}

var delayUpdate IDelayUpdate

type DelayUpdate struct {
}

func (t *DelayUpdate) Init(stub shim.ChaincodeStubInterface) peer.Response {
	m := make(map[string]string)
	delayUpdate = IDelayUpdate{
		UpdateQueue: []Msg{},
		LCF:         cuckoo.NewFilter(1000),
		UCF:         cuckoo.NewFilter(1000),
		HashLock:    m,
	}
	args := stub.GetStringArgs()
	if len(args) != 0 {
		shim.Error("invalid parameter numbers")
	}
	return shim.Success([]byte("success to init the contract"))
}

func (t *DelayUpdate) lockData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("invalid parameter numbers")
	}
	_, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("locked data must exist")
	}
	delayUpdate.HashLock[args[0]] = args[1]
	delayUpdate.LCF.InsertUnique([]byte(args[0]))
	return shim.Success([]byte("success to init the contract"))
}

/*
  1. verify the hash data
  2. consume the msg
  3. remove key from the LCF and UCF
*/
func (t *DelayUpdate) unlockData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("invalid parameter numbers")
	}
	_, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("locked data must exist")
	}
	bs := sha1.Sum([]byte(args[1]))
	s := fmt.Sprintf("%x", bs)
	if s == delayUpdate.HashLock[args[0]] {
		delete(delayUpdate.HashLock, args[0])
		latestData := getLastData(args[0])
		// delay update the data
		stub.PutState(args[0], latestData)
		for i := range delayUpdate.UpdateQueue {
			if delayUpdate.UpdateQueue[i].key == args[0] {
				// consume the queue
				delayUpdate.UpdateQueue = append(delayUpdate.UpdateQueue[:i], delayUpdate.UpdateQueue[i+1:]...)
			}
		}
		delayUpdate.UCF.Delete([]byte(args[0]))
		delayUpdate.LCF.Delete([]byte(args[0]))
	} else {
		return shim.Error("invalid hash key")
	}
	return shim.Success([]byte("success to init the contract"))
}

func (t *DelayUpdate) addData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args)%2 != 0 {
		return shim.Error("invalid parameter numbers")
	}
	for i := 0; i < len(args); i += 2 {
		err := stub.PutState(args[i], []byte(args[i+1]))
		if err != nil {
			return shim.Error("fail to put state")
		}
	}
	return shim.Success([]byte("success to add data"))
}

func getLastData(key string) []byte {
	for i := len(delayUpdate.UpdateQueue) - 1; i >= 0; i-- {
		if delayUpdate.UpdateQueue[i].key == key {
			return delayUpdate.UpdateQueue[i].value
		}
	}
	return nil
}

/*
 1. query LCF
 2. query UCF
*/
func (t *DelayUpdate) getData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("invalid parameter numbers")
	}
	data, _ := stub.GetState(args[0])
	if delayUpdate.LCF.Lookup([]byte(args[0])) {
		if delayUpdate.UCF.Lookup([]byte(args[0])) {
			res := getLastData(args[0])
			if res != nil {
				return shim.Success(res)
			}
		} else {
			return shim.Success(data)
		}
	} else {
		return shim.Success(data)
	}
	return shim.Success(data)
}

func (t *DelayUpdate) updateData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("invalid parameter numbers")
	}
	//fmt.Println(delayUpdate)
	if delayUpdate.LCF.Lookup([]byte(args[0])) {
		msg := Msg{
			key:        args[0],
			value:      []byte(args[1]),
			timeStamp:  time.Now().Unix(),
			isConsumed: false,
		}
		delayUpdate.UpdateQueue = append(delayUpdate.UpdateQueue, msg)
		if !delayUpdate.UCF.Lookup([]byte(args[0])) {
			delayUpdate.UCF.InsertUnique([]byte(args[0]))
		}
	} else {
		err := stub.PutState(args[0], []byte(args[1]))
		if err != nil {
			return shim.Success([]byte("fail to update the data"))
		}
	}
	return shim.Success([]byte("success to update the data"))
}

func (t *DelayUpdate) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	fn, args := stub.GetFunctionAndParameters()

	if fn == "lockData" {
		return t.lockData(stub, args)
	} else if fn == "addData" {
		return t.addData(stub, args)
	} else if fn == "getData" {
		return t.getData(stub, args)
	} else if fn == "updateData" {
		return t.updateData(stub, args)
	} else if fn == "unlockData" {
		return t.unlockData(stub, args)
	}
	return shim.Error("Invoke fn error")
}

func main() {
	err := shim.Start(new(DelayUpdate))
	if err != nil {
		fmt.Println("start error")
	}
}
