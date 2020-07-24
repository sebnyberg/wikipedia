package wikipedia

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/sebnyberg/wikipedia/wikixml"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type xmlReader struct {
	r   *wikixml.MultiStreamReader
	err error
}

// GetXMLPageReader returns a reader that retrieves blocks of pages from the
// provided files. The provided index and pagefile should be in the Wikipedia
// multi-stream download format.
func GetXMLPageReader(indexfile string, pagefile string) (PageBlockReader, error) {
	nworker := runtime.NumCPU()
	r, err := wikixml.NewMultiStreamReader(context.Background(), indexfile, pagefile, nworker)
	if err != nil {
		return nil, err
	}
	return &xmlReader{r: r}, nil
}

func (r *xmlReader) Close() error {
	return nil
}

func (r *xmlReader) Next() ([]Page, error) {
	xmlblock, err := r.r.Next()
	if err != nil {
		return nil, err
	}

	protoblock := make([]Page, len(xmlblock))
	for idx := range xmlblock {
		protoblock[idx] = NewPageFromXML(&xmlblock[idx])
	}
	return protoblock, nil
}

// NewPageFromXML parses an XML page into Protobuf format.
func NewPageFromXML(xml *wikixml.Page) Page {
	revisions := make([]*Revision, len(xml.Revisions))
	for i, p := range xml.Revisions {
		t, err := time.Parse(time.RFC3339, p.Timestamp)
		if err != nil {
			log.Fatalln(err)
		}
		revisions[i] = &Revision{
			Id:   int32(p.ID),
			Ts:   timestamppb.New(t),
			Text: p.Text,
		}
	}
	p := Page{
		Id:        xml.ID,
		Title:     xml.Title,
		Namespace: xml.Namespace,
		Revisions: revisions,
	}
	if xml.Redirect != nil {
		p.RedirectTitle = xml.Redirect.Title
	}
	return p
}
