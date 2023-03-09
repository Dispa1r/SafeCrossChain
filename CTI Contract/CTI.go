package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"time"
)

type CTI struct {
	CTIHash   [20]byte
	IP        string
	FileHash  string
	TimeStamp int64
	Score     int
	URL       string
	data      []byte
}
type CTIContract struct {
}

func (t *CTIContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	args := stub.GetStringArgs()
	if len(args) != 0 {
		shim.Error("invalid parameter numbers")
	}
	return shim.Success([]byte("success to init the contract"))
}

func (t *CTIContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	fn, args := stub.GetFunctionAndParameters()

	if fn == "AggregateData" {
		return t.AggregateData(stub, args)
	}
	return shim.Error("Invoke fn error")
}

func (t *CTIContract) AggregateData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var Score int
	var IPList string
	var FileHash string
	var URL string
	var data []byte
	for i := range args {
		var tmpCTI CTI
		err := json.Unmarshal([]byte(args[i]), &tmpCTI)
		if err != nil {
			return shim.Error("fail to unmarshal Data")
		}
		Score += tmpCTI.Score
		IPList += tmpCTI.IP
		IPList += ";"
		FileHash += tmpCTI.FileHash
		FileHash += ";"
		URL += tmpCTI.URL
		URL += ";"
		data = append(data, tmpCTI.data...)
		data = append(data, ';')

	}
	now := time.Now().Unix()

	finalCTI := CTI{
		IP:        IPList,
		FileHash:  FileHash,
		TimeStamp: now,
		Score:     Score / len(args),
		data:      data,
		URL:       URL,
	}
	jsBytes, _ := json.Marshal(finalCTI)
	CTIHash := sha1.Sum(jsBytes)
	finalCTI.CTIHash = CTIHash

	jsBytesFinal, _ := json.Marshal(finalCTI)
	stub.PutState(string(now), jsBytesFinal)
	return shim.Success([]byte("success to aggregate data"))
}

func main() {
	err := shim.Start(new(CTIContract))
	if err != nil {
		fmt.Println("start error")
	}
}
