package wikipedia

import (
	"fmt"
	"io"
)

// PageStore stores individual pages.
type PageStore interface {
	Get(id int32) (Page, error)
	Set(id int32, p *Page) error
}

// PageBlockReader reads blocks of pages in Protobuf format.
type PageBlockReader interface {
	// Next returns the next block of pages.
	// If there are no more pages, io.EOF is returned.
	Next() ([]Page, error)
	io.Closer
}

// PageBlockWriter writes blocks of pages in Protobuf format.
type PageBlockWriter interface {
	// Write writes a block of pages.
	Write([]Page) error
	io.Closer
}

func Transfer(from PageBlockReader, to PageBlockWriter) error {
	i := 0
	n := 0
	for {
		i++
		if i%100 == 0 {
			fmt.Printf("\riter: %v, count: %v", i, n)
		}
		p, err := from.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		n += len(p)
		if err := to.Write(p); err != nil {
			return err
		}
	}
}
