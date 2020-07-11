package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/pkg/profile"
	"github.com/sebnyberg/wikirel"
)

func main() {
	// Print elapsed time
	defer func(start time.Time) {
		log.Println("Elapsed time: ", time.Now().Sub(start))
	}(time.Now())

	// Profile CPU usage
	defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()

	r, err := wikirel.NewPageReaderFromFile("tmp/regular-part1.xml.bz2")
	if err != nil {
		log.Fatalln(err)
	}

	var p = new(wikirel.Page)
	count := 0
	for ; ; count++ {
		if err := r.ReadInto(p); err != nil {
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
