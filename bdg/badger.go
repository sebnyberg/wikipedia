package bdg

import (
	"context"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/sebnyberg/wikirel"
	"github.com/sebnyberg/wikirel/tbyte"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

type pageEntry struct {
	key   []byte
	value []byte
}

type pageWriter struct {
	db    *badger.DB
	pageC chan []wikirel.Page
	g     *errgroup.Group
	ctx   context.Context
	i     int
	mtx   sync.RWMutex
}

func NewPageWriter(outpath string) (wikirel.PageBlockWriter, error) {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	db, err := badger.Open(badger.DefaultOptions(outpath))
	if err != nil {
		return nil, err
	}
	db.DropAll()
	g, ctx := errgroup.WithContext(context.Background())

	w := &pageWriter{
		db:    db,
		pageC: make(chan []wikirel.Page, 1000),
		g:     g,
		ctx:   ctx,
	}

	return w, nil
}

func (w *pageWriter) Close() error {
	close(w.pageC)
	if err := w.g.Wait(); err != nil {
		return err
	}
	return w.db.Close()
}

func (w *pageWriter) Write(pageblock []wikirel.Page) error {
	w.g.Go(func() error {
		return w.db.Update(func(txn *badger.Txn) error {
			for _, p := range pageblock {
				b, err := proto.Marshal(&p)
				if err != nil {
					return err
				}
				if err := txn.Set(tbyte.Int32ToBytes(p.Id), b); err != nil {
					return err
				}
			}
			return nil
		})
	})
	return nil
}
