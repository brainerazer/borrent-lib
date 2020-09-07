package borrentlib

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"time"
)

//
func DownloadChunk(peer *PeerPoolEntry, persister ChunkPersister, chunkSize uint64, chunkID uint64, chunkHash []byte) (err error) {
	conn := peer.Conn
	var transferBlockSize uint64 = 0x4000

	for j := uint64(0); j < chunkSize/transferBlockSize; j++ {
		if peer.ConnInfo.PeerChoking == 1 {
			unchokePeer(conn)
			peer.ConnInfo.PeerChoking = 0
		}

		err := WriteMessage(conn, Request{
			Index:  uint32(chunkID),
			Begin:  uint32(j * transferBlockSize),
			Length: uint32(transferBlockSize),
		})

		if err != nil {
			return err
		}

		for true {
			msg, err := ReadMessage(conn)
			if err != nil {
				return err
			}
			piece, ok := msg.(Piece)
			if ok {
				fmt.Printf("begin: %d, idx: %d, block: %v..., peer: %s\n", piece.Begin, piece.Index, piece.Block[:5], peer.peerInfo.PeerID)
				persister.PersistChunk(int64(piece.Index), int64(piece.Begin), piece.Block)
				break
			} else {
				fmt.Printf("%#v\n", msg)
			}
		}
	}

	hash, _ := persister.ReadChunkHash(chunkID)
	if !bytes.Equal(hash, chunkHash) {
		fmt.Printf("Different hashes on chunk %d\n", chunkID)
		return errors.New("Different hashes")
	}

	return nil
}

//
func unchokePeer(conn net.Conn) error {
	err := WriteMessage(conn, Interested{})
	if err != nil {
		return err
	}

	for true {
		msg, err := ReadMessage(conn)
		if err != nil {
			return err
		}

		_, ok := msg.(Unchoke)
		if ok {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil

}
