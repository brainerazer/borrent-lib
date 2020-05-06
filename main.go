package main

import (
	"fmt"
	"math/rand"
	"os"
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

	conn, err := borrentlib.DialPeerTCP(responce.Peers[0])
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	hs, err := borrentlib.PeerHandshake(conn, tf.InfoHash[:], peerID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", hs)
	fmt.Printf("%s, %s, %v\n", hs.Str, hs.PeerID, hs.InfoHash)
	// fmt.Println(tf.Info.PiecesHashes)
}
