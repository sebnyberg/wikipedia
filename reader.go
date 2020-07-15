package wikirel

import (
	"bufio"
	"compress/bzip2"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
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
func (r *PageReader) Read(p *XMLPage) error {
	// Skip <mediawiki> and <siteinfo> tag once per document
	if !r.headerSkipped {
		// Skip <mediawiki> tag
		if _, err := r.dec.Token(); err != nil {
			return fmt.Errorf("%w: could not parse mediawiki tag, err: %v", ErrFailedToParse, err)
		}

		// Skip <siteinfo> tag
		si := struct{}{}
		if err := r.dec.Decode(&si); err != nil {
			return fmt.Errorf("%w: could not parse siteinfo tag, err: %v", ErrFailedToParse, err)
		}

		r.headerSkipped = true
	}

	if err := r.dec.Decode(p); err != nil {
		if err == io.EOF {
			return io.EOF
		}
		return fmt.Errorf("%w: could not parse page, err: %v", ErrFailedToParse, err)
	}

	return nil
}

// ReadPagesFromOffset puts the next chunk of pages into the provided slice.
// If the slice cannot fit into the provided pages slice, a new slice will be created.
func ReadPagesFromOffset(r io.ReadSeeker, offset int64, count int) ([]XMLPage, error) {
	pages := make([]XMLPage, count)

	if _, err := r.Seek(offset, 0); err != nil {
		return nil, fmt.Errorf("%w: failed to seek to offset, err: %v", ErrFailedToParse, err)
	}
	bz := bzip2.NewReader(r)
	dec := xml.NewDecoder(bz)

	// Decode pages until end of chunk
	for i := 0; i < count; i++ {
		if err := dec.Decode(&pages[i]); err != nil {
			return nil, fmt.Errorf("%w: failed to parse page, err: %v", ErrFailedToParse, err)
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
// The reader is expected to read plaintext XML from the multi-stream
// index file. To use this with the bzipped Wikipedia download,
// use it like so:
//
//	f, _ := os.Open('path/to/multistream-index.xml.bz2)
//	bz := bzip2.NewReader(f)
//	r := NewMultiStreamIndexReader(bz)
//
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
		return fmt.Errorf("offset %w, err: %v", ErrFailedToParse, err)
	}
	parsedInt, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return fmt.Errorf("ID %w, err: %v", ErrFailedToParse, err)
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

type MultiStreamResult struct {
	Pages []XMLPage
	Err   error
}

type MultiStreamReader struct {
	idxfile  string
	pagefile string

	indices chan MultiStreamIndex
	pages   chan []XMLPage

	ctx    context.Context
	cancel context.CancelFunc

	errOnce sync.Once
	err     error
}

// NewMultiReader creates a reader that returns pages from the multi-stream download.
// Both files should be provided in bzip2 format.
func ReadMultiStream(
	ctx context.Context,
	idxfile string,
	pagefile string,
	nworkers int,
) (*MultiStreamReader, error) {

	r := new(MultiStreamReader)

	r.idxfile = idxfile
	r.pagefile = pagefile
	r.ctx, r.cancel = context.WithCancel(ctx)

	r.indices = make(chan MultiStreamIndex, 1000)

	// Read indices and put them on the indices channel
	go r.indexWorker()

	// There are <100 pages per block, so this channel will buffer 100k pages total
	r.pages = make(chan []XMLPage, 1000)

	var wg sync.WaitGroup
	wg.Add(nworkers)
	for i := 0; i < nworkers; i++ {
		go r.pageWorker(&wg)
	}
	go func() {
		wg.Wait()
		close(r.pages)
		r.done(io.EOF)
	}()

	return r, nil
}

func (r *MultiStreamReader) done(err error) {
	r.errOnce.Do(func() {
		r.err = err
		r.cancel()
	})
}

func (r *MultiStreamReader) indexWorker() {
	defer close(r.indices)

	f, err := os.OpenFile(r.idxfile, os.O_RDONLY, 0644)
	defer f.Close()
	if err != nil {
		r.done(fmt.Errorf("%w: failed to open index file, err: %v", ErrInvalidFile, err))
		return
	}

	bz := bzip2.NewReader(f)
	indexrd := NewMultiStreamIndexReader(bz)

	for {
		idx, err := indexrd.ReadIndex()
		if err != nil {
			if err == io.EOF {
				break
			}
			r.done(fmt.Errorf("%w index file, err: %v", ErrFailedToParse, err))
		}

		select {
		case r.indices <- *idx:
		case <-r.ctx.Done():
			r.done(r.ctx.Err())
			break
		}
	}
}

func (r *MultiStreamReader) pageWorker(wg *sync.WaitGroup) {
	defer wg.Done()

	f, err := os.OpenFile(r.pagefile, os.O_RDONLY, 0644)
	if err != nil {
		r.done(fmt.Errorf("%w: failed to open pages file, err: %v", ErrInvalidFile, err))
	}
	defer f.Close()

	for idx := range r.indices {
		var pages []XMLPage
		pages, err := ReadPagesFromOffset(f, idx.Offset, idx.PageCount)
		if err != nil {
			r.done(fmt.Errorf("unexpected error when reading multi-stream pages, err: %v", err))
		}

		select {
		case r.pages <- pages:
		case <-r.ctx.Done():
			r.done(r.ctx.Err())
			break
		}
	}
}

func (r *MultiStreamReader) Next() ([]XMLPage, error) {
	if r.err != nil {
		return nil, r.err
	}
	return <-r.pages, nil
}
