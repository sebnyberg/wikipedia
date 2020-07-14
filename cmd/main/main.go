package main

import (
	"compress/bzip2"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	_ "net/http/pprof"

	_ "github.com/mkevac/debugcharts"

	"github.com/sebnyberg/wikirel"
)

type Mapper struct {
	idMtx     sync.RWMutex
	idToTitle map[int32]string
	titleMtx  sync.RWMutex
	titleToID map[string]int32
}

func main() {
	readPages()
	// readIndicesToMap()
}

func readPages() {
	// Print elapsed time
	defer func(start time.Time) {
		log.Println("Elapsed time: ", time.Now().Sub(start))
	}(time.Now())

	f, err := os.OpenFile("tmp/multistream-index.txt.bz2", os.O_RDONLY, 0644)
	check(err)
	bz := bzip2.NewReader(f)
	indexReader := wikirel.NewPageIndexBlockReader(bz)

	ctx, cancel := context.WithCancel(context.Background())

	// Generate indices
	indices := make(chan *wikirel.PageIndexBlock, 1000)
	go func() {
		defer close(indices)
		for {
			offset, count, err := indexReader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				cancel()
			}
			select {
			case indices <- indexBlock:
			case <-ctx.Done():
				fmt.Println("received cancel on index channel, exiting...")
				return
			}
		}
	}()

	nworker := 8
	var wg sync.WaitGroup
	pages := make(chan []wikirel.Page, 1000)
	wg.Add(nworker)
	for i := 0; i < nworker; i++ {
		go func() {
			defer wg.Done()

			// Consume indices
			f, err = os.OpenFile("tmp/multistream.xml.bz2", os.O_RDONLY, 0644)
			check(err)
			pageReader := wikirel.NewMultiPageReader(f)

			var err error
			p := make([]wikirel.Page, 1000)
			for idx := range indices {
				p, err = pageReader.ReadPagesFromOffset(idx.Offset, idx.Count, p)
				if err != nil {
					fmt.Println("encountered error when reading pages...", err)
					cancel()
					return
				}
				select {
				case pages <- p:
				case <-ctx.Done():
					fmt.Println("reading of pages skipped due to context cancel")
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(pages)
	}()

	ntotal := 0
	iter := 0
	for pagechunk := range pages {
		iter++
		ntotal += len(pagechunk)

		if iter%100 == 0 {
			fmt.Println(ntotal)
			fmt.Println(pagechunk[0].Title)
		}
	}

	fmt.Println("done!")
}

func readIndicesToMap() {
	// Print elapsed time
	defer func(start time.Time) {
		log.Println("Elapsed time: ", time.Now().Sub(start))
	}(time.Now())

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	// defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()

	f, err := os.OpenFile("tmp/multistream-index.txt.bz2", os.O_RDONLY, 0644)
	check(err)
	bz := bzip2.NewReader(f)
	r := wikirel.NewPageIndexReader(bz)
	if err != nil {
		log.Fatalln(err)
	}

	m := Mapper{}
	m.idToTitle = map[int32]string{}
	m.titleToID = map[string]int32{}

	count := 0
	var p = new(wikirel.PageIndex)
	for ; ; count++ {
		if err := r.Read(p); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Unexpected err: %v", err)
		}

		m.idToTitle[p.ID] = p.Title
		m.titleToID[p.Title] = p.ID

		if count%10000 == 0 {
			fmt.Printf("Read: %v\r", count)
		}
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("total alloc during execution: %d\n", ms.TotalAlloc)
	fmt.Printf("heap alloc (after free): %d\n", ms.HeapAlloc)
	fmt.Printf("heap sys: %d\n", ms.HeapSys)
	fmt.Printf("total memory from OS: %d\n", ms.Sys)
	fmt.Printf("heap objects: %d\n", ms.HeapObjects)

	log.Printf("Done! Read %v page indices\n", count)
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
