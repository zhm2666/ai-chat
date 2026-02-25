package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

// 互斥锁
var mu sync.Mutex

var sum uint

var packageList *sync.Map = new(sync.Map)

const TaskNum = 5

type task struct {
	id       uint32
	callback chan uint
}

var ChanTasks []chan task = make([]chan task, TaskNum)
var r *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

var logger *log.Logger

func initLog() {
	f, _ := os.Create("./lottery_demo.log")
	logger = log.New(f, "", log.Ldate|log.Lmicroseconds)
}

func SetRedPack(id, money, num int) {
	moneyTotal := int(money * 100)

}

func main() {
	var r *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

}
