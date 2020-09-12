package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"

	"github.com/brainerazer/borrent-lib/borrentlib"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var torrentfile = flag.String("torrentfile", "", "torrent file to download (required)")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)

		defer pprof.StopCPUProfile()
	}

	if *torrentfile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())
	bytes, err := os.Open(*torrentfile)
	if err != nil {
		panic(err)
	}

	tf, err := borrentlib.DecodeTorrentFile(bytes)
	if err != nil {
		panic(err)
	}

	borrentlib.DownloadTorrentFile(tf, 10)
}
