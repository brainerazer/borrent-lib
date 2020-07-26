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

	persister, err := borrentlib.InitDenseFileDiskChunkPersister(tf.FileInfo.Name, tf.FileInfo.Length, tf.FileInfo.PieceLength)

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
	go read(conn, &pInfo, persister, &mu)
	go write(conn, &pInfo, tf.FileInfo.Length/tf.FileInfo.PieceLength, &mu)

	select {}
	// fmt.Println(tf.Info.PiecesHashes)
}

func read(conn net.Conn, pInfo *borrentlib.PeerConnectionInfo, persister borrentlib.ChunkPersister, mu *sync.Mutex) {
	for true {
		msg, err := borrentlib.ReadMessage(conn)
		if err != nil {
			panic(err)
		}
		mu.Lock()
		pInfo.AmInterested = 0
		mu.Unlock()
		piece, ok := msg.(borrentlib.Piece)
		if ok {
			fmt.Printf("begin: %d, idx: %d, block: %v...\n", piece.Begin, piece.Index, piece.Block[:5])
			persister.PersistChunk(int64(piece.Index), piece.Block)
		} else {
			fmt.Printf("%#v\n", msg)
		}
		_, ok = msg.(borrentlib.Unchoke)
		if ok {
			mu.Lock()
			pInfo.PeerChoking = 0
			mu.Unlock()
		}
		mu.Lock()
		pInfo.AmInterested = 1
		mu.Unlock()
	}
}

func write(conn net.Conn, pInfo *borrentlib.PeerConnectionInfo, chunks uint64, mu *sync.Mutex) {
	err := borrentlib.WriteMessage(conn, borrentlib.Interested{})
	if err != nil {
		panic(err)
	}
	for i := uint64(0); i < chunks; {
		time.Sleep(1 * time.Second)
		mu.Lock()
		isChoking := pInfo.PeerChoking
		amInterested := pInfo.AmInterested
		mu.Unlock()
		if isChoking == 0 && amInterested == 1 {
			err := borrentlib.WriteMessage(conn, borrentlib.Request{Index: uint32(i), Begin: 0x0, Length: 0x4000})
			i++
			if err != nil {
				panic(err)
			}
		}
	}
}
