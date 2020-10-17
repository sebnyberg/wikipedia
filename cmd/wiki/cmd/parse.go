package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/pkg/profile"
	"github.com/sebnyberg/wikipedia"
	"github.com/sebnyberg/wikipedia/bdg"
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

	var reader wikipedia.PageBlockReader
	var err error
	switch c.String("infmt") {
	case "proto":
		reader, err = wikipedia.NewProtoBlockReader(pagefile, 100)
		if err != nil {
			return err
		}
	case "xml":
		idxfile := c.String("idxfile")
		if len(idxfile) == 0 {
			return errors.New("idxfile is required")
		}
		reader, err = wikipedia.GetXMLPageReader(idxfile, pagefile)
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

	var writer wikipedia.PageBlockWriter
	switch outfmt {
	case "badger":
		writer, err = bdg.NewPageWriter(outpath)
		if err != nil {
			return fmt.Errorf("failed to create badger writer, err: %w", err)
		}
	case "proto":
		writer, err = wikipedia.NewProtoBlockWriter(outpath)
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

	return wikipedia.Transfer(reader, writer)
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
