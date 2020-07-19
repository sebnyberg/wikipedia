package wikirel

import (
	"context"
	"runtime"

	"github.com/sebnyberg/wikirel/wikixml"
)

type xmlReader struct {
	r   *wikixml.MultiStreamReader
	err error
}

func GetXMLPageReader(indexfile string, pagefile string) (PageBlockReader, error) {
	nworker := runtime.NumCPU()
	r, err := wikixml.NewMultiStreamReader(context.Background(), indexfile, pagefile, nworker)
	if err != nil {
		return nil, err
	}
	return &xmlReader{r: r}, nil
}

func (r *xmlReader) Close() error {
	return nil
}

func (r *xmlReader) Next() ([]Page, error) {
	xmlblock, err := r.r.Next()
	if err != nil {
		return nil, err
	}

	protoblock := make([]Page, len(xmlblock))
	for idx := range xmlblock {
		protoblock[idx] = NewPageFromXML(&xmlblock[idx])
	}
	return protoblock, nil
}
