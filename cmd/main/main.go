package main

import (
	"log"
	"sync"

	_ "net/http/pprof"

	_ "github.com/mkevac/debugcharts"
)

type Mapper struct {
	idMtx     sync.RWMutex
	idToTitle map[int32]string
	titleMtx  sync.RWMutex
	titleToID map[string]int32
}

func main() {
	readPages()
}

func readPages() {
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
