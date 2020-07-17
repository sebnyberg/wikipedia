package main

import (
	"fmt"
	"log"
	"os"

	"net/http"
	_ "net/http/pprof"

	_ "github.com/mkevac/debugcharts"
	"github.com/sebnyberg/wikirel/cmd/wiki/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	if err := NewApp().Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

const pkgName string = "wikirel"

func NewApp() *cli.App {
	app := &cli.App{
		Name:     pkgName,
		HelpName: pkgName,
		Usage:    "wiki relation commands",
		Commands: []*cli.Command{
			cmd.Parse(),
		},
	}

	return app
}

// type Mapper struct {
// 	idMtx     sync.RWMutex
// 	idToTitle map[int32]string
// 	titleMtx  sync.RWMutex
// 	titleToID map[string]int32
// }

// func main() {
// 	idxfile := "tmp/multistream-index.txt.bz2"
// 	cmd.WriteXMLToProto(idxfile)

// 	// Open the Badger database located in the /tmp/badger directory.
// 	// It will be created if it doesn't exist.
// 	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	db.DropAll()
// 	defer db.Close()

// 	writeProtoToDB(db)

// 	// err = db.Update(func(txn *badger.Txn) error {
// 	// 	err := txn.Set([]byte("greeting"), []byte("hello"))
// 	// 	return err
// 	// })
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func writeProtoToDB(db *badger.DB) {
// 	f, err := os.OpenFile("tmp/pages.proto.ld.zs", os.O_RDONLY, 0644)
// 	check(err)
// 	buf := bufio.NewReader(f)
// 	zs := zstd.NewReader(buf)
// 	r := protoio.NewReader(zs)
// 	defer func() {
// 		check(zs.Close())
// 		check(f.Close())
// 	}()

// 	type Entry struct {
// 		Key   []byte
// 		Value []byte
// 	}

// 	g, ctx := errgroup.WithContext(context.Background())

// 	pageC := make(chan Entry, 1000)

// 	g.Go(func() error {
// 		defer close(pageC)
// 		for {
// 			var p wikirel.FullPage
// 			if err := r.ReadMsg(&p); err != nil {
// 				if err == io.EOF {
// 					break
// 				}
// 				log.Fatalln(err)
// 			}
// 			b, err := proto.Marshal(&p)
// 			if err != nil {
// 				return err
// 			}
// 			select {
// 			case pageC <- Entry{tbyte.Int32ToBytes(p.Id), b}:
// 			case <-ctx.Done():
// 				return nil
// 			}
// 		}
// 		return nil
// 	})

// 	nworkers := 50
// 	doneC := make(chan struct{}, 1000)
// 	for i := 0; i < nworkers; i++ {
// 		g.Go(func() error {
// 			for e := range pageC {
// 				err := db.Update(func(txn *badger.Txn) error {
// 					return txn.Set(e.Key, e.Value)
// 				})
// 				if err != nil {
// 					return err
// 				}
// 				doneC <- struct{}{}
// 			}
// 			return nil
// 		})
// 	}

// 	go func() {
// 		i := 0
// 		for range doneC {
// 			i++
// 			if i%50 == 0 {
// 				fmt.Printf("\r%v", i)
// 			}
// 		}
// 		fmt.Println("done! inserted", i, "messages")
// 	}()

// 	if err := g.Wait(); err != nil {
// 		log.Fatalln(err)
// 		log.Fatalln(err)
// 		log.Fatalln(err)
// 	}
// }

// // err = db.View(func(txn *badger.Txn) error {
// // 	item, err := txn.Get([]byte("greeting"))
// // 	if err != nil {
// // 		return err
// // 	}
// // 	err = item.Value(func(val []byte) error {
// // 		fmt.Println(string(val))
// // 		return nil
// // 	})
// // 	if err != nil {
// // 		return err
// // 	}
// // 	return nil
// // })

// // func readProto() {
// // 	f, err := os.OpenFile("tmp/pages.proto.ld.zs", os.O_RDONLY, 0644)
// // 	check(err)
// // 	buf := bufio.NewReader(f)
// // 	z, err := zstd.NewReader(buf)
// // 	check(err)
// // 	protor := protoio.NewReader(z)

// // 	defer func() {
// // 		check(f.Close())
// // 	}()

// // 	defer func(start time.Time) {
// // 		fmt.Println("elapsed: ", time.Now().Sub(start))
// // 	}(time.Now())

// // 	i := 0
// // 	m := new(wikirel.Page)
// // 	for {
// // 		i++
// // 		if err := protor.ReadMsg(m); err != nil {
// // 			if err != io.EOF {
// // 				fmt.Println(err)
// // 			}
// // 			break
// // 		}
// // 		if i%100000 == 0 {
// // 			fmt.Printf("%v\r", i)
// // 		}
// // 	}
// // 	fmt.Println("done!")
// // }

// func check(err error) {
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// }
