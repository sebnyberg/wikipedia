package main

import (
	"compress/bzip2"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/sebnyberg/wikirel"
)

func main() {
	// Print elapsed time
	defer func(start time.Time) {
		log.Println("Elapsed time: ", time.Now().Sub(start))
	}(time.Now())

	// // Profile CPU usage
	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()

	f, err := os.OpenFile("tmp/regular-part1.xml.bz2", os.O_RDONLY, 0644)
	check(err)
	bz := bzip2.NewReader(f)

	r := wikirel.NewPageReader(bz)
	if err != nil {
		log.Fatalln(err)
	}

	var p = new(wikirel.Page)
	count := 0
	for ; ; count++ {
		if err := r.Read(p); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Unexpected err: %v", err)
		}
		if count%10 == 0 {
			fmt.Printf("Read: %v\r", count)
		}
	}

	log.Printf("Done! Read %v files\n", count)
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
