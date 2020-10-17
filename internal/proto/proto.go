package proto

import (
	"bufio"
	"io"
	"os"

	"github.com/DataDog/zstd"
	"github.com/sebnyberg/protoio"
	"github.com/sebnyberg/wikipedia"
)

type writer struct {
	protow *protoio.Writer
	close  func() error
}

// NewPageWriter returns a writer that puts pages into
// a file in zstd-compressed length-delimited protobuf format.
//
// If the provided path already exists, an error is returned.
//
// The length-delimited format weaves serialized Protobuf messages with
// their length as a prefix, which allows for reading messages one by one
// from the file, acting like an append-only log of serialized messages.
func NewPageWriter(path string) (wikipedia.PageWriter, error) {
	f, err := os.OpenFile(path, os.O_EXCL|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	buf := bufio.NewWriter(f)
	zs := zstd.NewWriter(buf)
	protow := protoio.NewWriter(zs)

	close := func() error {
		if err := zs.Close(); err != nil {
			return err
		}
		if err := buf.Flush(); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		return nil
	}

	return &writer{
		protow: protow,
		close:  close,
	}, nil
}

func (w *writer) Close() error {
	return w.close()
}

func (w *writer) Write(page *wikipedia.Page) error {
	if err := w.protow.WriteMsg(page); err != nil {
		return err
	}

	return nil
}

type reader struct {
	r         *protoio.Reader
	close     func() error
	blocksize int
}

// NewProtoBlockReader returns a reader that retrieves blocks of pages
// from the provided path.
//
// If a file does not exist at the provided path, an error is returned.
//
// The file is expected to be zstd-compressed append-only (length-delimited)
// log of serialized Protobuf messages. To write messages in this format,
// use the NewProtoBlockWriter() constructor.
func NewProtoBlockReader(path string) (wikipedia.PageReader, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	buf := bufio.NewReader(f)
	zs := zstd.NewReader(buf)
	r := protoio.NewReader(zs)

	close := func() error {
		if err := zs.Close(); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		return nil
	}

	return &reader{
		r:     r,
		close: close,
	}, nil
}

func (r *reader) Close() error {
	return r.close()
}

func (r *reader) Next() (*wikipedia.Page, error) {
	var p wikipedia.Page
	if err := r.r.ReadMsg(&p); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}
	return &p, nil
}
