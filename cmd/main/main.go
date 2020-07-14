package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	_ "net/http/pprof"

	_ "github.com/mkevac/debugcharts"
	"github.com/pkg/profile"
	"github.com/sebnyberg/wikirel"
)

type Mapper struct {
	idMtx     sync.RWMutex
	idToTitle map[int32]string
	titleMtx  sync.RWMutex
	titleToID map[string]int32
}

func main() {
	idxfile := "tmp/multistream-index.txt.bz2"
	pagesfile := "tmp/multistream.xml.bz2"

	r, err := wikirel.ReadMultiStream(context.Background(), idxfile, pagesfile, 16)
	if err != nil {
		log.Fatalln("failed to start pages stream", err)
	}

	defer profile.Start(profile.ProfilePath("."), profile.CPUProfile).Stop()

	defer func(start time.Time) {
		fmt.Println("elapsed: ", time.Now().Sub(start))
	}(time.Now())

	i := 0
	ntotal := 0
	for {
		i++
		pages, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("unexpected error", err)
			break
		}
		ntotal += len(pages)
		if i%1000 == 0 {
			fmt.Println(ntotal)
		}
	}
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
