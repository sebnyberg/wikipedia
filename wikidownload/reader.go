package wikidownload

import (
	"context"
	"runtime"

	"github.com/sebnyberg/wikipedia"
)

type reader struct {
	r     *MultiStreamReader
	err   error
	block []Page
}

func GetPageReader(indexfile string, pagefile string) (wikipedia.PageReader, error) {
	nworker := runtime.NumCPU()
	r, err := NewMultiStreamReader(context.Background(), indexfile, pagefile, nworker)
	if err != nil {
		return nil, err
	}
	return &reader{r: r}, nil
}

func (r *reader) Close() error {
	return nil
}

func (r *reader) Next() (*wikipedia.Page, error) {
	xmlBlock, err := r.r.Next()
	if err != nil {
		return nil, err
	}

}
