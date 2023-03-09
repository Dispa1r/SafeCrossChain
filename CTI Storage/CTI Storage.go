package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
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

type CTIStorage struct {
}

func (t *CTIStorage) Init(stub shim.ChaincodeStubInterface) peer.Response {
	args := stub.GetStringArgs()
	if len(args) != 0 {
		shim.Error("invalid parameter numbers")
	}
	return shim.Success([]byte("success to init the contract"))
}

func (t *CTIStorage) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	fn, args := stub.GetFunctionAndParameters()

	if fn == "addData" {
		return t.addData(stub, args)
	} else if fn == "getData" {
		return t.getData(stub, args)
	}
	return shim.Error("Invoke fn error")
}

func (t *CTIStorage) getData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("invalid parameters")
	}
	queryType := args[0]
	queryTarget := args[1]

	queryString := fmt.Sprintf("{\"selector\":{\"%s\":\"%s\"}}", queryType, queryTarget)
	qis, err := stub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error("query error:" + err.Error())
	}

	defer qis.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for qis.HasNext() {
		queryResponse, err := qis.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())

}

func (t *CTIStorage) addData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	for i := range args {
		var tmpCTI CTI
		err := json.Unmarshal([]byte(args[i]), &tmpCTI)
		if err != nil {
			return shim.Error("fail to unmarshal Data")
		}
		jsbyte, _ := json.Marshal(tmpCTI)
		var tmpByte []byte
		for _, j := range tmpCTI.CTIHash {
			tmpByte = append(tmpByte, j)
		}
		stub.PutState(string(tmpByte), jsbyte)
	}
	return shim.Success([]byte("success to storage CTI info"))

}

func main() {
	err := shim.Start(new(CTIStorage))
	if err != nil {
		fmt.Println("start error")
	}
}
