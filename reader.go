package wikirel

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// PageReader reads Wikipedia pages from an input stream.
type PageReader struct {
	dec           *xml.Decoder
	headerSkipped bool
}

var ErrParseFailed = errors.New("parse failed")

// NewPageReader returns a new page reader reading from r.
//
// The provided reader is expected to read plaintext XML from
// the non-multi-stream Wikipedia database download.
//
// To read from a bzip2 file, open the file like so:
//
// 	f, err := os.OpenFile("path/to/file.xml.bzip2", os.O_RDONLY, 0644)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	bz := bzip2.NewReader(f)
// 	r := wikirel.NewPageReader(bz)
//
func NewPageReader(r io.Reader) *PageReader {
	return &PageReader{
		dec: xml.NewDecoder(r),
	}
}

var ErrInvalidFile = errors.New("invalid file")

// Read returns the next page from the reader.
// If there are no more pages, io.EOF is returned.
func (r *PageReader) Read(p *Page) error {
	// Skip <mediawiki> and <siteinfo> tag once per document
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

	if err := r.dec.Decode(p); err != nil {
		if err == io.EOF {
			return io.EOF
		}
		return fmt.Errorf("%w: could not parse page, err: %v", ErrParseFailed, err)
	}

	return nil
}

// PageIndexBlock points to a block of articles in the multi-stream articles file.
type PageIndexBlock struct {
	// Offset denotes the number of bytes from the start of the articles file
	// to where the index block begins.
	Offset uint64

	// Count is the number of articles in the index block.
	Count int
}

type PageIndexBlockReader struct {
	scanner    *bufio.Scanner
	prevOffset uint64
	pageCount  int
	done       bool
}

// NewPageIndexBlockReader returns a reader that returns index blocks
// from the provided file.
//
// The reader is expected to read plaintext from the multi-stream index file.
//
// To use the bzipped Wikipedia extract, use this function like so:
//
//	f, err := os.OpenFile("path/to/file.txt.bz2", os.O_RDONLY, 0644)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	bz := bzip2.NewReader(f)
//	r := wikirel.NewPageIndexBlockReader(bz)
//
func NewPageIndexBlockReader(r io.Reader) *PageIndexBlockReader {
	return &PageIndexBlockReader{
		scanner: bufio.NewScanner(r),
	}
}

// Read returns the next index block from the reader.
// If there are no more blocks, io.EOF is returned.
func (r *PageIndexBlockReader) Read() (*PageIndexBlock, error) {
	for r.scanner.Scan() {
		curOffset, err := parseOffset(r.scanner.Text())
		if err != nil {
			return nil, err
		}

		if r.prevOffset == 0 && curOffset > 0 {
			r.prevOffset = curOffset
		}

		if curOffset < r.prevOffset {
			return nil, ErrInvalidOffset
		}

		if curOffset == r.prevOffset {
			r.pageCount++
			continue
		}

		// Offset has changed and scan did not return false
		// We are entering a new indexing block
		chunk := &PageIndexBlock{
			Offset: r.prevOffset,
			Count:  r.pageCount,
		}

		// Set current offset & reset counter
		r.prevOffset = curOffset
		r.pageCount = 1

		return chunk, nil
	}

	// Return an error if the scanner stopped unexpectedly
	// Err() returns nil if we are at io.EOF
	if err := r.scanner.Err(); err != nil {
		return nil, err
	}

	if r.pageCount == 0 {
		return nil, io.EOF
	}

	// Return remainder of last chunk
	lastIndexBlock := &PageIndexBlock{
		Offset: r.prevOffset,
		Count:  r.pageCount,
	}

	// Reset counter to trigger nil response on next call
	r.pageCount = 0

	return lastIndexBlock, io.EOF
}

var ErrBadRecord = errors.New("bad record")
var ErrInvalidOffset = errors.New("invalid offset")

// parseIndexRow parses one row from the index summary file
// Each row is on the format: "offset:articleID:articleName", e.g.
// "10:592:Andorra"
func parseOffset(s string) (uint64, error) {
	for idx, ch := range s {
		if ch == ':' {
			offset, err := strconv.ParseUint(s[:idx], 10, 64)
			if err != nil {
				return 0, ErrInvalidOffset
			}
			return offset, nil
		}
	}

	return 0, ErrBadRecord
}
