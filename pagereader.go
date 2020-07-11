package wikirel

import (
	"compress/bzip2"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

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

// PageReader reads Pages from a Wikipedia database download.
type PageReader interface {
	// Read returns one page from the download.
	// The last page will return io.EOF. Subsequent calls
	// return a nil page with io.EOF as the error.
	Read() (*Page, error)

	// ReadInto reads the content next page into the provided page.
	ReadInto(*Page) error
}

type pageReader struct {
	dec           *xml.Decoder
	headerSkipped bool
}

var ErrParseFailed = errors.New("parse failed")

// NewPageReader returns a new page parser.
func NewPageReader(r io.Reader) PageReader {
	return &pageReader{
		dec: xml.NewDecoder(r),
	}
}

var ErrInvalidFile = errors.New("invalid file")

// NewPageReaderFromFile creates a new page reader from file.
// If the file path does not end with .bz2, an error is returned.
func NewPageReaderFromFile(filename string) (PageReader, error) {
	if ext := path.Ext(filename); ext != ".bz2" {
		return nil, fmt.Errorf(
			"%w: file must be in bzip2 format, was: %v",
			ErrInvalidFile, ext,
		)
	}

	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	bz := bzip2.NewReader(f)

	return NewPageReader(bz), nil
}

func (r *pageReader) Read() (*Page, error) {
	var page = new(Page)
	if err := r.ReadInto(page); err != nil {
		return nil, err
	}
	return page, nil
}

func (r *pageReader) ReadInto(page *Page) error {
	if !r.headerSkipped {
		// Skip <mediawiki> tag
		if _, err := r.dec.Token(); err != nil {
			return fmt.Errorf("%w: could not parse mediawiki tag, err: %v", ErrParseFailed, err)
		}

		// Skip <siteinfo> tag
		si := struct{}{}
		if err := r.dec.Decode(&si); err != nil {
			return fmt.Errorf("%w: could not parse siteinfo tag, err: %v", ErrParseFailed, err)
		}

		r.headerSkipped = true
	}

	if err := r.dec.Decode(page); err != nil {
		if err == io.EOF {
			return io.EOF
		}
		return fmt.Errorf("%w: could not parse page, err: %v", ErrParseFailed, err)
	}

	return nil
}
