package wikirel

import (
	"bufio"
	"compress/bzip2"
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

// NewPageReader returns a new page reader reading from r.
//
// The provided reader is expected to read plaintext XML from
// the non-multi-stream Wikipedia database download.
//
func NewPageReader(r io.Reader) *PageReader {
	return &PageReader{
		dec: xml.NewDecoder(r),
	}
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

// MultiStreamReader reads blocks of articles from the multistream export.
// Each time a new MultiStreamIndex is read, the provided reader will seek
// to the start of the block, and read all articles from the block.
type MultiStreamReader struct {
	r io.ReadSeeker
}

// ReadPagesFromOffset puts the next chunk of pages into the provided slice.
// If the slice cannot fit into the provided pages slice, a new slice will be created.
func (r *MultiStreamReader) ReadPagesFromOffset(offset int64, count int) ([]Page, error) {
	pages := make([]Page, count)

	if _, err := r.r.Seek(offset, 0); err != nil {
		return nil, fmt.Errorf("%w: failed to seek to offset, err: %v", ErrParseFailed, err)
	}
	bz := bzip2.NewReader(r.r)
	dec := xml.NewDecoder(bz)

	// Decode pages until end of chunk
	for i := 0; i < count; i++ {
		if err := dec.Decode(&pages[i]); err != nil {
			return nil, fmt.Errorf("%w: failed to parse page, err: %v", ErrParseFailed, err)
		}
	}

	return pages, nil
}

// MultiStreamIndexReader reads blocks of indices from the multistream index.
// The multi-stream index file contains lists of indices,
// where up to 100 articles (a block) share the same byte offset in the pages file.
type MultiStreamIndexReader struct {
	scanner    *bufio.Scanner
	prevoffset int64
	npages     int
}

// MultiStreamIndex contains the offset of the first article,
// and the number of articles in a block of articles.
type MultiStreamIndex struct {
	// Offset is the byte offset of the first articles in the block.
	Offset int64

	// PageCount is the number of articles in the block.
	PageCount int
}

// MultiStreamIndexRow is one row from the multi-stream index file.
type MultiStreamIndexRow struct {
	// Offset denotes the number of bytes from the start of the articles file
	// to where the index block begins.
	Offset int64

	// ID of the page
	ID int32

	// Title of the page
	Title string
}

// NewMultiStreamIndexReader returns a reader that returns index blocks
// from the provided file.
//
// The reader is expected to read plaintext from the multi-stream index file.
func NewMultiStreamIndexReader(r io.Reader) *MultiStreamIndexReader {
	return &MultiStreamIndexReader{
		scanner: bufio.NewScanner(r),
	}
}

// NextRow returns the next row from the multi-stream index file.
func (r *MultiStreamIndexReader) ReadRow(row *MultiStreamIndexRow) error {
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
	row.Offset, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("offset %w, err: %v", ErrParseFailed, err)
	}
	parsedInt, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return fmt.Errorf("ID %w, err: %v", ErrParseFailed, err)
	}
	row.ID = int32(parsedInt)
	row.Title = parts[2]

	return nil
}

// ReadIndex returns the offset and count of pages in the next index block
// If there are no more blocks, an offset and count of zero, and an
// error of io.EOF is returned.
func (r *MultiStreamIndexReader) ReadIndex() (*MultiStreamIndex, error) {
	for r.scanner.Scan() {
		curOffset, err := parseOffset(r.scanner.Text())
		if err != nil {
			return nil, err
		}

		if r.prevoffset == 0 && curOffset > 0 {
			r.prevoffset = curOffset
		}

		if curOffset < r.prevoffset {
			return nil, ErrInvalidOffset
		}

		if curOffset == r.prevoffset {
			r.npages++
			continue
		}

		// Set current offset & reset counter
		defer func() {
			r.prevoffset = curOffset
			r.npages = 1
		}()

		return &MultiStreamIndex{r.prevoffset, r.npages}, nil
	}

	// Return an error if the scanner stopped unexpectedly
	// Err() returns nil if we are at io.EOF
	if err := r.scanner.Err(); err != nil {
		return nil, err
	}

	if r.npages == 0 {
		return nil, io.EOF
	}

	// Reset counter to trigger nil response on next call
	defer func() { r.npages = 0 }()

	return &MultiStreamIndex{r.prevoffset, r.npages}, nil
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
