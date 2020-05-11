package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
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

	pInfo := borrentlib.NewPeerConnectionInfo()
	mu := sync.Mutex{}
	go read(conn, &pInfo, &mu)
	go write(conn, &pInfo, &mu)

	select {}
	// fmt.Println(tf.Info.PiecesHashes)
}

func read(conn net.Conn, pInfo *borrentlib.PeerConnectionInfo, mu *sync.Mutex) {
	for true {
		msg, err := borrentlib.ReadMessage(conn)
		if err != nil {
			panic(err)
		}
		piece, ok := msg.(borrentlib.Piece)
		if ok {
			fmt.Printf("begin: %d, idx: %d, block: %v...\n", piece.Begin, piece.Index, piece.Block[:5])
		} else {
			fmt.Printf("%#v\n", msg)
		}
		_, ok = msg.(borrentlib.Unchoke)
		if ok {
			mu.Lock()
			pInfo.PeerChoking = 0
			mu.Unlock()
		}
	}
}

func write(conn net.Conn, pInfo *borrentlib.PeerConnectionInfo, mu *sync.Mutex) {
	err := borrentlib.WriteMessage(conn, borrentlib.Interested{})
	if err != nil {
		panic(err)
	}
	for true {
		time.Sleep(1 * time.Second)
		mu.Lock()
		isChoking := pInfo.PeerChoking
		mu.Unlock()
		if isChoking == 0 {
			err := borrentlib.WriteMessage(conn, borrentlib.Request{Index: 0x0, Begin: 0x0, Length: 0x4000})
			if err != nil {
				panic(err)
			}
			break
		}
	}
}
