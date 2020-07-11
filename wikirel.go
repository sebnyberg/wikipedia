package wikirel

import "errors"

var ErrNotImplemented = errors.New("not implemented")

type Redirect struct {
	Title string `xml:"title,attr"`
}

type Page struct {
	Title     string    `xml:"title"`
	Namespace uint64    `xml:"ns"`
	ID        uint64    `xml:"id"`
	Redirect  *Redirect `xml:"redirect"`
	Text      string    `xml:"revision>text"`
}
