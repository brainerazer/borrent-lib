package main

import (
	"os"
	"math/rand"
	"fmt"
	"time"
	"github.com/brainerazer/borrent-lib/borrentlib"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	path := "borrentlib/testdata/ubuntu-20.04-desktop-amd64.iso.torrent"
	bytes, err := os.Open(path)
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

	fmt.Println(peerID)
	fmt.Println(responce)
}
