package wikidownload

import (
	"errors"
)

var ErrNotImplemented = errors.New("not implemented")

type Page struct {
	ID        int32      `xml:"id"`
	Title     string     `xml:"title"`
	Namespace uint32     `xml:"ns"`
	Redirect  *Redirect  `xml:"redirect"`
	Revisions []Revision `xml:"revision"`
}

type Redirect struct {
	Title string `xml:"title,attr"`
}

type Revision struct {
	ID        uint32 `xml:"id"`
	Timestamp string `xml:"timestamp"`
	Text      string `xml:"text"`
}
