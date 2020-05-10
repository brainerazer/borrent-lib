package main

import (
	"fmt"
	"math/rand"
	"net"
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

	go read(conn)
	time.Sleep(5 * time.Second)
	err = borrentlib.WriteMessage(conn, borrentlib.Request{Index: 0x0, Begin: 0x0, Length: 0x4000})
	if err != nil {
		panic(err)
	}
	select {}
	// fmt.Println(tf.Info.PiecesHashes)
}

func read(conn net.Conn) {
	for true {
		msg, err := borrentlib.ReadMessage(conn)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v", msg)
	}
}
