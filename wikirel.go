package wikirel

import (
	"fmt"
	"io"
	"time"

	"github.com/sebnyberg/wikirel/protoutil"
	"github.com/sebnyberg/wikirel/wikixml"
)

type PageStore interface {
	Get(id int32) (Page, error)
	Set(id int32, p *Page) error
}

type PageBlockReader interface {
	Next() ([]Page, error)
	io.Closer
}

type PageBlockWriter interface {
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

func NewPageFromXML(xml *wikixml.Page) Page {
	revisions := make([]*Revision, len(xml.Revisions))
	for i, p := range xml.Revisions {
		ts := protoutil.MustParseTSFromString(time.RFC3339, p.Timestamp)
		revisions[i] = &Revision{
			Id:   int32(p.ID),
			Ts:   ts,
			Text: p.Text,
		}
	}
	p := Page{
		Id:        xml.ID,
		Title:     xml.Title,
		Namespace: xml.Namespace,
		Revisions: revisions,
	}
	if xml.Redirect != nil {
		p.RedirectTitle = xml.Redirect.Title
	}
	return p
}
