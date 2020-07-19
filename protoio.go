package wikirel

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/DataDog/zstd"
	"github.com/sebnyberg/protoio"
)

type protoWriter struct {
	protow *protoio.Writer
	close  func() error
}

func NewProtoBlockWriter(path string) (PageBlockWriter, error) {
	// Force users to recognize the somewhat unorthodox protobuf file format
	if !strings.HasSuffix(path, ".ld.zs") {
		return nil, errors.New("when using protobuf, path should end with .ld.zs")
	}

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

	return &protoWriter{
		protow: protow,
		close:  close,
	}, nil
}

func (w *protoWriter) Close() error {
	return w.close()
}

func (w *protoWriter) Write(pages []Page) error {
	for _, p := range pages {
		if err := w.protow.WriteMsg(&p); err != nil {
			return err
		}
	}
	return nil
}

type protoReader struct {
	r         *protoio.Reader
	close     func() error
	blocksize int
}

func NewProtoBlockReader(path string, blocksize int) (PageBlockReader, error) {
	// Force users to recognize the somewhat unorthodox protobuf file format
	if !strings.HasSuffix(path, ".ld.zs") {
		return nil, errors.New("when using protobuf, path should end with .ld.zs")
	}

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

	return &protoReader{
		r:         r,
		close:     close,
		blocksize: blocksize,
	}, nil
}

func (r *protoReader) Close() error {
	return r.close()
}

func (r *protoReader) Next() ([]Page, error) {
	block := make([]Page, 0, r.blocksize)
	for i := 0; i < r.blocksize; i++ {
		var m Page
		if err := r.r.ReadMsg(&m); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		block = append(block, m)
	}
	if len(block) == 0 {
		return nil, io.EOF
	}
	return block, nil
}
