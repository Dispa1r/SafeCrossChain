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

func Test_DelayUpdate(t *testing.T) {
	DelayUpdateTest := new(DelayUpdate)
	stub := shim.NewMockStub("test", DelayUpdateTest)

	checkInit(t, stub, [][]byte{})
	dataNum := 10000
	m := make(map[string]string)
	var tmpKey []byte
	var tmpValue []byte
	for i := 0; i < dataNum; i++ {
		tmpKey = randBytes(4)
		tmpValue = randBytes(4)
		checkInvoke(t, stub, [][]byte{[]byte("addData"), tmpKey, tmpValue})
		m[string(tmpKey)] = string(tmpValue)
	}
	lockedNum := 10000
	var counter int
	for k := range m {
		if counter >= lockedNum {
			break
		}
		counter++
		checkInvoke(t, stub, [][]byte{[]byte("lockData"), []byte(k), []byte("lockedHash")})
	}

	updateNum := 2000
	var counter1 int
	for k := range m {
		if counter1 >= updateNum {
			break
		}
		counter1++
		checkInvoke(t, stub, [][]byte{[]byte("updateData"), []byte(k), []byte("updateData")})
	}

	var timeSum int64
	for i := 0; i < 10000; i++ {
		start := time.Now()
		for k, _ := range m {
			checkInvoke(t, stub, [][]byte{[]byte("getData"), []byte(k)})
		}
		elapsed := time.Since(start)
		timeSum += int64(elapsed)
		//fmt.Println("第", i, "轮该函数执行完成耗时：", elapsed)
	}
	fmt.Println("10000次平均耗时：", timeSum/10000)

}
