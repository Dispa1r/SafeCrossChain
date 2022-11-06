package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"math/rand"
	"testing"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
}

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	//fmt.Println(string(res.Payload))
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

func checkInvoke(t *testing.T, stub *shim.MockStub, args [][]byte) peer.Response {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.FailNow()
	}
	return res
}

func Test_addData(t *testing.T) {
	MerkleTreeChainTest := new(merkleTreeChain)
	stub := shim.NewMockStub("test", MerkleTreeChainTest)

	// parameters here means the height of merkle tree chain
	checkInit(t, stub, [][]byte{[]byte("4")})

	dataNum := 10000
	m := make(map[string]string)
	var tmpKey []byte
	var tmpValue []byte

	start := time.Now()
	for i := 0; i < dataNum; i++ {
		tmpKey = randBytes(4)
		tmpValue = randBytes(4)
		checkInvoke(t, stub, [][]byte{[]byte("addData1"), tmpKey, tmpValue})
		m[string(tmpKey)] = string(tmpValue)
	}
	elapsed := time.Since(start)
	fmt.Println("插入法1执行完成耗时：", elapsed)

	start = time.Now()
	for i := 0; i < dataNum; i++ {
		tmpKey = randBytes(4)
		tmpValue = randBytes(4)
		checkInvoke(t, stub, [][]byte{[]byte("addData2"), tmpKey, tmpValue})
		m[string(tmpKey)] = string(tmpValue)
	}
	elapsed = time.Since(start)
	fmt.Println("插入法2执行完成耗时：", elapsed)

	start = time.Now()
	for i := 0; i < dataNum; i++ {
		tmpKey = randBytes(4)
		tmpValue = randBytes(4)
		checkInvoke(t, stub, [][]byte{[]byte("updateData1"), tmpKey, []byte("changeData1")})
		m[string(tmpKey)] = string(tmpValue)
	}
	elapsed = time.Since(start)
	fmt.Println("更新法2执行完成耗时：", elapsed)

	start = time.Now()
	for i := 0; i < dataNum; i++ {
		tmpKey = randBytes(4)
		tmpValue = randBytes(4)
		checkInvoke(t, stub, [][]byte{[]byte("updateData2"), tmpKey, []byte("changeData2")})
		m[string(tmpKey)] = string(tmpValue)
	}
	elapsed = time.Since(start)
	fmt.Println("更新法2执行完成耗时：", elapsed)

	start = time.Now()
	for i := 0; i < dataNum; i++ {
		tmpKey = randBytes(4)
		tmpValue = randBytes(4)
		checkInvoke(t, stub, [][]byte{[]byte("verifyData1"), tmpKey})
		m[string(tmpKey)] = string(tmpValue)
	}
	elapsed = time.Since(start)
	fmt.Println("验证法1执行完成耗时：", elapsed)

	start = time.Now()
	for i := 0; i < dataNum; i++ {
		tmpKey = randBytes(4)
		tmpValue = randBytes(4)
		checkInvoke(t, stub, [][]byte{[]byte("verifyData2"), tmpKey})
		m[string(tmpKey)] = string(tmpValue)
	}
	elapsed = time.Since(start)
	fmt.Println("验证法2执行完成耗时：", elapsed)

	// compare the storage with MerkleTree and MerkleChain
	checkInvoke(t, stub, [][]byte{[]byte("compareStorage"), tmpKey})

	//lockedNum := 0
	//var counter int
	//for k := range m {
	//	if counter >= lockedNum {
	//		break
	//	}
	//	counter++
	//	checkInvoke(t, stub, [][]byte{[]byte("lockData"), []byte(k), []byte("lockedHash")})
	//}
	//
	//updateNum := 0
	//var counter1 int
	//for k := range m {
	//	if counter1 >= updateNum {
	//		break
	//	}
	//	counter1++
	//	checkInvoke(t, stub, [][]byte{[]byte("updateData"), []byte(k), []byte("updateData")})
	//}
	//
	//var timeSum int64
	//for i := 0; i < 10000; i++ {
	//	start := time.Now()
	//	for k, _ := range m {
	//		checkInvoke(t, stub, [][]byte{[]byte("getData"), []byte(k)})
	//	}
	//	elapsed := time.Since(start)
	//	timeSum += int64(elapsed)
	//	//fmt.Println("第", i, "轮该函数执行完成耗时：", elapsed)
	//}
	//fmt.Println("10000次平均耗时：", timeSum/10000)

}

// 0  : 3734134
// 20 : 3744248
// 40 : 3742785
// 60 ：3854321
// 80 : 3926885
// 100: 3967888

// 0  : 3967888
// 20 ： 4042471
// 40 : 4088033
// 60 : 4280645
// 80 ： 4362651
// 100： 5436364
