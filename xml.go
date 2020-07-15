package wikirel

import "errors"

var ErrNotImplemented = errors.New("not implemented")

type XMLRedirect struct {
	Title string `xml:"title,attr"`
}

type XMLRevision struct {
	ID        uint32 `xml:"id"`
	Timestamp string `xml:"timestamp"`
	Text      string `xml:"text"`
}

type XMLPage struct {
	ID        int32         `xml:"id"`
	Title     string        `xml:"title"`
	Namespace uint32        `xml:"ns"`
	Redirect  *XMLRedirect  `xml:"redirect"`
	Revisions []XMLRevision `xml:"revision"`
}
