package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/DataDog/zstd"
	"github.com/golang/protobuf/ptypes"
	"github.com/sebnyberg/protoio"
	"github.com/sebnyberg/wikirel"
	"github.com/sebnyberg/wikirel/wikixml"
	"github.com/urfave/cli/v2"
)

func Parse() *cli.Command {
	return &cli.Command{
		Name:        "parse",
		Description: "parse the wikipedia dataset to some other format",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "pagefile",
				Usage: "`FILE` to parse articles from - must be a multi-stream pages",
				Value: "tmp/multistream.xml.bz2",
			},
			&cli.StringFlag{
				Name:  "idxfile",
				Usage: "`FILE` to parse indices from",
				Value: "tmp/multistream-index.txt.bz2",
			},
			&cli.StringFlag{
				Name:     "output",
				Usage:    "output format, can be either 'badger', or 'proto'",
				Aliases:  []string{"o"},
				Required: true,
			},
			&cli.StringFlag{
				Name:  "outpath",
				Usage: "output path. For proto, use a file, for badger, use a directory",
			},
			&cli.IntFlag{
				Name:  "nworker",
				Usage: "`N` workers to use when parsing pages",
				Value: 8,
			},
		},
		Action: func(c *cli.Context) error {
			return parseAction(c)
		},
	}
}

func parseAction(c *cli.Context) error {
	idxfile := c.String("idxfile")
	if len(idxfile) == 0 {
		return errors.New("idxfile is required")
	}
	pagefile := c.String("pagefile")
	if len(pagefile) == 0 {
		return errors.New("pagefile is required")
	}
	out := c.String("output")
	outpath := c.String("outpath")
	if len(outpath) == 0 {
		return errors.New("outpath is required")
	}

	switch out {
	case "badger":
		if path.Dir(outpath) != outpath {
			return errors.New("when using 'badger', output path must point to a directory")
		}
	case "proto":
		if path.Dir(outpath) == outpath {
			return errors.New("when using 'proto', output path must point to a file")
		}
	default:
		fmt.Println("output must be of type 'badger' or 'proto'")
	}

	return nil
}

// readXML reads wikipedia pages from the provided XML files, parses
// the XML pages into Protobuf pages, and puts them on pageC.
func readXML(
	ctx context.Context,
	idxfile string,
	pagesfile string,
	nworker int,
	pageC chan *wikirel.FullPage,
) error {
	defer close(pageC)
	r, err := wikixml.ReadMultiStream(context.Background(), idxfile, pagesfile, nworker)
	if err != nil {
		return err
	}
	for {
		pages, err := r.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		for _, page := range pages {
			select {
			case pageC <- wikirel.NewFullPageFromXML(&page):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func writeXMLToProto(idxfile string, pagesfile string, protofile string) {
	defer func(start time.Time) {
		fmt.Println("elapsed: ", time.Now().Sub(start))
	}(time.Now())

	f, err := os.OpenFile(protofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	check(err)
	buf := bufio.NewWriter(f)
	zs := zstd.NewWriter(buf)
	protow := protoio.NewWriter(zs)

	defer func() {
		check(zs.Close())
		check(buf.Flush())
		check(f.Close())
	}()

	i := 0
	ntotal := 0
	for {
		i++
		pages, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			check(err)
			break
		}
		for _, page := range pages {
			revisions := make([]*wikirel.Revision, len(page.Revisions))
			for i, p := range page.Revisions {
				t, err := time.Parse(time.RFC3339, p.Timestamp)
				check(err)
				ts, err := ptypes.TimestampProto(t)
				check(err)
				revisions[i] = &wikirel.Revision{
					Id:   int32(p.ID),
					Ts:   ts,
					Text: p.Text,
				}
			}

			p := &wikirel.FullPage{
				Id:        page.ID,
				Title:     page.Title,
				Namespace: page.Namespace,
				Revisions: revisions,
			}
			if page.Redirect != nil {
				p.Title = page.Redirect.Title
			}

			check(protow.WriteMsg(p))
		}
		ntotal += len(pages)
		fmt.Printf("%v\r", ntotal)
	}
	fmt.Println(i)
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
