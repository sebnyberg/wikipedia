package cmd

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/pkg/profile"
	"github.com/sebnyberg/wikirel"
	"github.com/sebnyberg/wikirel/bdg"
	"github.com/urfave/cli/v2"
)

func Parse() *cli.Command {
	return &cli.Command{
		Name:        "parse",
		Description: "parse the wikipedia dataset to protobuf or badger format",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "infmt",
				Usage:   "input `FORMAT`, can be either 'proto' or 'xml'. If XML is used, idxfile and pagefile must be set.",
				Aliases: []string{"i"},
				Value:   "tmp/multistream.xml.bz2",
			},
			&cli.StringFlag{
				Name:    "pagefile",
				Usage:   "input `FILE` to parse pages from. If XML is used, idxfile must be provided as well.",
				Aliases: []string{"f"},
			},
			&cli.StringFlag{
				Name:  "idxfile",
				Usage: "`FILE` to parse indices from. Only used when parsing from XML.",
			},
			&cli.StringFlag{
				Name:     "outfmt",
				Usage:    "output `FORMAT`, can be either 'badger', or 'proto'",
				Aliases:  []string{"o"},
				Required: true,
			},
			&cli.StringFlag{
				Name:  "outpath",
				Usage: "output `PATH`. For proto, use a file, for badger, use a directory",
			},
		},
		Action: func(c *cli.Context) error {
			return parseAction(c)
		},
	}
}

func parseAction(c *cli.Context) error {
	pagefile := c.String("pagefile")
	if len(pagefile) == 0 {
		return errors.New("pagefile is required")
	}

	var reader wikirel.PageBlockReader
	var err error
	switch c.String("infmt") {
	case "proto":
		reader, err = wikirel.NewProtoBlockReader(pagefile, 100)
		if err != nil {
			return err
		}
	case "xml":
		idxfile := c.String("idxfile")
		if len(idxfile) == 0 {
			return errors.New("idxfile is required")
		}
		reader, err = wikirel.GetXMLPageReader(idxfile, pagefile)
		if err != nil {
			return err
		}
	default:
		fmt.Println("output must be of type 'proto' or 'xml'")
	}

	// Output sink
	outfmt := c.String("outfmt")
	outpath := c.String("outpath")
	if len(outpath) == 0 {
		return errors.New("outpath is required")
	}

	var writer wikirel.PageBlockWriter
	switch outfmt {
	case "badger":
		writer, err = bdg.NewPageWriter(outpath)
		if err != nil {
			return fmt.Errorf("failed to create badger writer, err: %w", err)
		}
	case "proto":
		writer, err = wikirel.NewProtoBlockWriter(outpath)
		if err != nil {
			return fmt.Errorf("failed to create proto writer, err: %w", err)
		}
	default:
		fmt.Println("output must be of type 'badger' or 'proto'")
	}
	defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()

	defer func() {
		check(reader.Close())
		check(writer.Close())
	}()

	return wikirel.Transfer(reader, writer)
}

func writeXMLToProto(idxfile string, pagesfile string, protofile string) {
	defer func(start time.Time) {
		fmt.Println("elapsed: ", time.Now().Sub(start))
	}(time.Now())

	// f, err := os.OpenFile(protofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	// check(err)
	// buf := bufio.NewWriter(f)
	// zs := zstd.NewWriter(buf)
	// protow := protoio.NewWriter(zs)

	// defer func() {
	// 	check(zs.Close())
	// 	check(buf.Flush())
	// 	check(f.Close())
	// }()

	// i := 0
	// ntotal := 0
	// for {
	// 	i++
	// 	pages, err := r.Next()
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		check(err)
	// 		break
	// 	}
	// 	for _, page := range pages {
	// 		revisions := make([]*wikirel.Revision, len(page.Revisions))
	// 		for i, p := range page.Revisions {
	// 			t, err := time.Parse(time.RFC3339, p.Timestamp)
	// 			check(err)
	// 			ts, err := ptypes.TimestampProto(t)
	// 			check(err)
	// 			revisions[i] = &wikirel.Revision{
	// 				Id:   int32(p.ID),
	// 				Ts:   ts,
	// 				Text: p.Text,
	// 			}
	// 		}

	// 		p := &wikirel.Page{
	// 			Id:        page.ID,
	// 			Title:     page.Title,
	// 			Namespace: page.Namespace,
	// 			Revisions: revisions,
	// 		}
	// 		if page.Redirect != nil {
	// 			p.Title = page.Redirect.Title
	// 		}

	// 		check(protow.WriteMsg(p))
	// 	}
	// 	ntotal += len(pages)
	// 	fmt.Printf("%v\r", ntotal)
	// }
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
