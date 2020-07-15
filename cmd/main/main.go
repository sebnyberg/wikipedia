package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	_ "net/http/pprof"

	_ "github.com/mkevac/debugcharts"
	"github.com/pkg/profile"
	"github.com/sebnyberg/protoio"
	"github.com/sebnyberg/wikirel"
)

type Mapper struct {
	idMtx     sync.RWMutex
	idToTitle map[int32]string
	titleMtx  sync.RWMutex
	titleToID map[string]int32
}

func main() {
	// fmt.Println("writing proto files")
	// writeProto()
	fmt.Println("reading proto files")
	readProto()
}

func readProto() {
	r, err := os.OpenFile("tmp/pages.proto.ld", os.O_RDONLY, 0644)
	check(err)
	defer func() {
		check(r.Close())
	}()
	protor := protoio.NewReader(bufio.NewReader(r))

	defer func(start time.Time) {
		fmt.Println("elapsed: ", time.Now().Sub(start))
	}(time.Now())

	i := 0
	m := new(wikirel.Page)
	for {
		i++
		if err := protor.ReadMsg(m); err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
		if i%100000 == 0 {
			fmt.Printf("%v\r", i)
		}
	}
	fmt.Println("done!")
}

func writeProto() {
	idxfile := "tmp/multistream-index.txt.bz2"
	pagesfile := "tmp/multistream.xml.bz2"
	// idxfile := "tmp/enwiki-20200620-pages-articles-multistream-index1.txt-p1p30303.bz2"
	// pagesfile := "tmp/enwiki-20200620-pages-articles-multistream1.xml-p1p30303.bz2"

	r, err := wikirel.ReadMultiStream(context.Background(), idxfile, pagesfile, 16)
	check(err)

	defer profile.Start(profile.ProfilePath("."), profile.CPUProfile).Stop()

	defer func(start time.Time) {
		fmt.Println("elapsed: ", time.Now().Sub(start))
	}(time.Now())

	outf, err := os.OpenFile("tmp/pages.proto.ld", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	check(err)
	defer func() {
		check(outf.Close())
	}()
	protow := protoio.NewWriter(bufio.NewWriter(outf))

	i := 0
	ntotal := 0
	for {
		i++
		pages, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			check(err)
			break
		}
		for _, page := range pages {
			check(protow.WriteMsg(&wikirel.Page{
				Id:    page.ID,
				Text:  page.Text,
				Title: page.Title,
			}))
		}
		ntotal += len(pages)
		fmt.Printf("%v\r", ntotal)
	}
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
