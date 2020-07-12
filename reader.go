package wikirel

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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

type MultiPageReader struct {
	r io.ReadSeeker
}

// NewMultiPageReader returns a new page reader reading from r.
//
// The provided reader is expected to read plaintext XML from
// the multi-stream Wikipedia database download.
//
// To read from a bzip2 file, open the file like so:
//
// 	f, err := os.OpenFile("path/to/file.xml.bzip2", os.O_RDONLY, 0644)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	bz := bzip2.NewReader(f)
// 	r := wikirel.NewMultiPageReader(bz)
//
func NewMultiPageReader(r io.ReadSeeker) *MultiPageReader {
	return &MultiPageReader{
		r: r,
	}
}

// ReadChunkFromOffset puts the next chunk of pages into the provided slice.
// If the slice cannot fit into the provided pages slice, a new slice will be created.
func (r *MultiPageReader) ReadChunkFromOffset(offset int64, count int, pages []Page) ([]Page, error) {
	// Create new slice if necessary
	if cap(pages) < count {
		pages = make([]Page, count, count)
	}
	pages = pages[:count]

	if _, err := r.r.Seek(offset, 0); err != nil {
		return nil, fmt.Errorf("%w: failed to seek to offset, err: %v", ErrParseFailed, err)
	}
	dec := xml.NewDecoder(r.r)

	// Decode pages until end of chunk
	for i := 0; i < count; i++ {
		dec.Decode(&pages[i])
	}

	return pages, nil
}

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
	Offset int64

	// Count is the number of articles in the index block.
	Count int
}

type PageIndexBlockReader struct {
	scanner    *bufio.Scanner
	prevOffset int64
	pageCount  int
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

// parseOffset parses one row from the index summary file
// Each row is on the format: "offset:articleID:articleName", e.g.
// "10:592:Andorra"
func parseOffset(s string) (int64, error) {
	for idx, ch := range s {
		if ch == ':' {
			offset, err := strconv.ParseInt(s[:idx], 10, 64)
			if err != nil {
				return 0, ErrInvalidOffset
			}
			return offset, nil
		}
	}

	return 0, ErrBadRecord
}

// PageIndex points to a block of articles in the multi-stream articles file.
type PageIndex struct {
	// Offset denotes the number of bytes from the start of the articles file
	// to where the index block begins.
	Offset int64

	// ID of the page
	ID int64

	// Title of the page
	Title string
}

type PageIndexReader struct {
	scanner *bufio.Scanner
}

// PageIndexReader returns a reader that reads indices from the file.
//
// The provided reader is expected to read from a multi-index file.
//
// When using the multi-stream index file to read from the multi-stream
// articles file - use the PageIndexBlockReader.
//
// To use the bzipped Wikipedia extract, use this function like so:
//
//	f, err := os.OpenFile("path/to/some-multi-stream-index.txt.bz2", os.O_RDONLY, 0644)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	bz := bzip2.NewReader(f)
//	r := wikirel.NewPageIndexBlockReader(bz)
//
func NewPageIndexReader(r io.Reader) *PageIndexReader {
	return &PageIndexReader{
		scanner: bufio.NewScanner(r),
	}
}

// Read returns the next index from the file.
// If there are no more indices, io.EOF is returned.
func (r *PageIndexReader) Read(index *PageIndex) error {
	if !r.scanner.Scan() {
		// Err() returns nil if EOF
		if err := r.scanner.Err(); err != nil {
			return err
		}
		return io.EOF
	}

	parts := strings.SplitN(r.scanner.Text(), ":", 3)
	if len(parts) != 3 {
		return ErrBadRecord
	}
	var err error
	index.Offset, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("offset %w, err: %v", ErrParseFailed, err)
	}
	index.ID, err = strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return fmt.Errorf("ID %w, err: %v", ErrParseFailed, err)
	}
	index.Title = parts[2]

	return nil
}

// parseIndexRow parses one row from the index summary file
// Each row is on the format: "offset:articleID:articleName", e.g.
// "10:592:Andorra"
func parseIndexRow(s string) (*PageIndex, error) {

	return nil, ErrBadRecord
}
