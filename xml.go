package wikirel

import "errors"

var ErrNotImplemented = errors.New("not implemented")

type XMLRedirect struct {
	Title string `xml:"title,attr"`
}

type XMLPage struct {
	Title     string       `xml:"title"`
	Namespace uint64       `xml:"ns"`
	ID        int32        `xml:"id"`
	Redirect  *XMLRedirect `xml:"redirect"`
	Text      string       `xml:"revision>text"`
}
