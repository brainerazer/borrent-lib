package main

import (
	"flag"
	"fmt"
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
	peerID, responce, err := borrentlib.AnnounceMyself(tf)
	if err != nil {
		panic(err)
	}

	// fmt.Println("#v", tf)

	fmt.Println(peerID)
	fmt.Println(responce)

	persister, err := borrentlib.InitDenseFileDiskChunkPersister(tf.FileInfo.Name, tf.FileInfo.Length, tf.FileInfo.PieceLength)

	peerPool := borrentlib.InitPeerPool(tf.InfoHash, peerID)

	for _, peer := range responce.Peers[:10] {
		peerPool.ConnectPeer(peer)

	}

	jobs := make(chan job, 20)
	done := make(chan bool)

	for w := 1; w <= 10; w++ {
		go worker(jobs, done)
	}

	for i := 0; i < len(tf.FileInfo.PiecesHashes); i++ {
		jobs <- job{
			&peerPool,
			persister,
			tf,
			uint64(i),
		}
	}

	close(jobs)

	for w := 1; w <= 10; w++ {
		<-done
		fmt.Printf("Done %d\n", w)
	}

	peerPool.Close()
}

type job struct {
	peerPool    *borrentlib.PeerPool
	persister   borrentlib.ChunkPersister
	torrentFile borrentlib.TorrentFile
	chunkID     uint64
}

func worker(jobs <-chan job, done chan<- bool) {
	var err error
	for j := range jobs {
		fmt.Println("job started")

		var peer *borrentlib.PeerPoolEntry
		for true {
			peer, err = j.peerPool.GetPeer()
			if err != nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			break
		}
		fmt.Println("got peer")

		err := borrentlib.DownloadChunk(peer, j.persister, j.torrentFile.FileInfo.PieceLength, uint64(j.chunkID), j.torrentFile.FileInfo.PiecesHashes[j.chunkID])
		if err != nil {
			panic(err)
		}

		peer.Return()
	}
	done <- true
}
