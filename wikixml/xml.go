package wikixml

import (
	"errors"
	"time"

	"github.com/sebnyberg/wikirel"
	"github.com/sebnyberg/wikirel/protoutil"
)

var ErrNotImplemented = errors.New("not implemented")

type Redirect struct {
	Title string `xml:"title,attr"`
}

type Revision struct {
	ID        uint32 `xml:"id"`
	Timestamp string `xml:"timestamp"`
	Text      string `xml:"text"`
}

type Page struct {
	ID        int32      `xml:"id"`
	Title     string     `xml:"title"`
	Namespace uint32     `xml:"ns"`
	Redirect  *Redirect  `xml:"redirect"`
	Revisions []Revision `xml:"revision"`
}

func NewFullPage(xml *Page) *wikirel.FullPage {
	revisions := make([]*wikirel.Revision, len(xml.Revisions))
	for i, p := range xml.Revisions {
		ts := protoutil.MustParseTSFromString(time.RFC3339, p.Timestamp)
		revisions[i] = &wikirel.Revision{
			Id:   int32(p.ID),
			Ts:   ts,
			Text: p.Text,
		}
	}
	p := &wikirel.FullPage{
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
