package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/cbergoon/merkletree"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"log"
	"math"
	"strconv"
)

var merkleTreeHeight int
var merkleMaxNodes int
var merkleMap map[string]int
var totalTreeNum int

type nodeValue struct {
	ledgerData
	isRoot   bool
	rootHash []byte
}

type resData struct {
	data     string
	proof    [][]byte
	location []int64
}

type ledgerData struct {
	key   string
	value string
}

var cacheData1 []merkletree.Content

var cacheData2 []merkletree.Content

var IMerkleTreeChain []*merkletree.MerkleTree

var globalMerkleTree *merkletree.MerkleTree

type merkleTreeChain struct {
}

// CalculateHash hashes the values of a TestContent
func (t nodeValue) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.ledgerData.key + t.ledgerData.value)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// Equals tests for equality of two Contents
func (t nodeValue) Equals(other merkletree.Content) (bool, error) {
	return t.ledgerData.key == other.(nodeValue).ledgerData.key && t.ledgerData.value == other.(nodeValue).ledgerData.value, nil
}

func (t *merkleTreeChain) Init(stub shim.ChaincodeStubInterface) peer.Response {
	args := stub.GetStringArgs()
	if len(args) != 1 {
		shim.Error("invalid parameter numbers")
	}
	height, _ := strconv.Atoi(args[0])
	merkleTreeHeight = height
	merkleMaxNodes = int(math.Pow(2, float64(merkleTreeHeight)))
	cacheData1 = []merkletree.Content{}
	cacheData2 = []merkletree.Content{}
	merkleMap = make(map[string]int)

	return shim.Success([]byte("success to init the contract"))
}

func (t *merkleTreeChain) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	fn, args := stub.GetFunctionAndParameters()

	if fn == "addData1" {
		return t.addData1(stub, args)
	} else if fn == "addData2" {
		return t.addData2(stub, args)
	} else if fn == "getData" {
		return t.getData(stub, args)
	} else if fn == "verifyData1" {
		return t.verifyData1(stub, args)
	} else if fn == "verifyData2" {
		return t.verifyData2(stub, args)
	} else if fn == "compareStorage" {
		return t.compareStorage(stub, args)
	} else if fn == "updateData1" {
		return t.updateData1(stub, args)
	} else if fn == "updateData2" {
		return t.updateData2(stub, args)
	}
	return shim.Error("Invoke fn error")
}

func (t *merkleTreeChain) getData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("invalid parameter numbers")
	}
	data, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("no such data")
	}
	targetTree := IMerkleTreeChain[merkleMap[args[0]]]
	merklePath, merkleLocation, err1 := targetTree.GetMerklePath(nodeValue{
		ledgerData: ledgerData{key: args[0]},
	})
	if err1 != nil {
		return shim.Error("could not find such data in this merkle tree")
	}
	resData := resData{
		data:     string(data),
		proof:    merklePath,
		location: merkleLocation,
	}
	bs, _ := json.Marshal(resData)

	return shim.Success(bs)
}

func checkUpdateMerkleNode(key, value string) {
	cacheData1 = append(cacheData1, nodeValue{ledgerData: ledgerData{key: key, value: value}, isRoot: false, rootHash: nil})
	if len(cacheData1) == merkleMaxNodes {

		for j := range cacheData1 {
			merkleMap[cacheData1[j].(nodeValue).ledgerData.key] = totalTreeNum
		}

		tree, err2 := merkletree.NewTree(cacheData1)
		IMerkleTreeChain = append(IMerkleTreeChain, tree)
		if err2 != nil {
			log.Fatal(err2)
		}
		cacheData1 = []merkletree.Content{}
		cacheData1 = append(cacheData1, nodeValue{isRoot: true, rootHash: tree.Root.Hash})
		totalTreeNum++
	}
}

func (t *merkleTreeChain) addData1(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args)%2 != 0 {
		return shim.Error("invalid parameter numbers")
	}

	for i := 0; i < len(args); i += 2 {
		err := stub.PutState(args[i], []byte(args[i+1]))
		if err != nil {
			return shim.Error("fail to put state")
		}

		cacheData2 = append(cacheData1, nodeValue{ledgerData: ledgerData{key: args[i], value: args[i+1]}})
		tmpTree, err1 := merkletree.NewTree(cacheData2)
		globalMerkleTree = tmpTree
		if err1 != nil {
			log.Fatal(err1)
		}

	}
	return shim.Success([]byte("success to add data"))
}

func (t *merkleTreeChain) addData2(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args)%2 != 0 {
		return shim.Error("invalid parameter numbers")
	}

	for i := 0; i < len(args); i += 2 {
		err := stub.PutState(args[i], []byte(args[i+1]))
		if err != nil {
			return shim.Error("fail to put state")
		}
		checkUpdateMerkleNode(args[i], args[i+1])

	}
	return shim.Success([]byte("success to add data"))
}

func (t *merkleTreeChain) updateData1(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("invalid parameter numbers")
	}
	_, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("no such data")
	}
	checkUpdateMerkleNode(args[0], args[1])
	return shim.Success([]byte("success to update data and merkleTree"))

}

func (t *merkleTreeChain) updateData2(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("invalid parameter numbers")
	}
	_, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("no such data")
	}
	for i := range cacheData2 {
		if cacheData2[i].(nodeValue).ledgerData.key == args[0] {
			cacheData2 = append(cacheData2[:i], cacheData2[i+1:]...)
			cacheData2 = append(cacheData2, nodeValue{ledgerData: ledgerData{key: args[0], value: args[1]}})
			globalMerkleTree.RebuildTree()
		}
	}
	return shim.Success([]byte("success to update data and merkleTree"))

}

func (t *merkleTreeChain) verifyData1(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("invalid parameter numbers")
	}
	targetTree := IMerkleTreeChain[merkleMap[args[0]]]
	res, err := targetTree.VerifyContent(nodeValue{ledgerData: ledgerData{key: args[0]}})
	if err != nil {
		return shim.Error("internal error")
	}
	if res {
		return shim.Success([]byte("success to verify data"))
	} else {
		return shim.Error("fail to verify data")
	}

}

func (t *merkleTreeChain) verifyData2(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("invalid parameter numbers")
	}
	res, err := globalMerkleTree.VerifyContent(nodeValue{ledgerData: ledgerData{key: args[0]}})
	if err != nil {
		return shim.Error("internal error")
	}
	if res {
		return shim.Success([]byte("success to verify data"))
	} else {
		return shim.Error("fail to verify data")
	}
}

func (t *merkleTreeChain) compareStorage(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	bs1, err := json.Marshal(IMerkleTreeChain)
	if err != nil {
		log.Fatal("fail to marshal the data")
	}
	bs2, err1 := json.Marshal(globalMerkleTree)
	if err1 != nil {
		log.Fatal("fail to marshal the data")
	}
	fmt.Println("merkle chain storage: ", len(bs1))
	fmt.Println("merkle tree storage: ", len(bs2))
	return shim.Success([]byte("finish to calc the storage"))

}

func main() {
	err := shim.Start(new(merkleTreeChain))
	if err != nil {
		fmt.Println("start error")
	}
}
