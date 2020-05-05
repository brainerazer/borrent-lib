package borrentlib

import (
	"bytes"
	"crypto/sha1"
	"io"

	"github.com/jackpal/bencode-go"
)

// TorrentFile is a torrent file descriptor struct
type TorrentFile struct {
	AnnounceURL string
	InfoHash    [20]byte     // InfoHash is a unique torrent ID, 20 bytes of SHA-1 hash
	FileInfo    DataFileInfo // Only single-file torrents are suppoted right now
}

// DataFileInfo is an information about a particular data file in a torrent
type DataFileInfo struct {
	Name         string
	Length       uint64
	PieceLength  uint64
	PiecesHashes [][20]byte // contains hashes of each piece in bytes array
}

// FileInfo - Member of info map in the torrent file.

// Auxiliary structs for bencode unmarshalling
type torrentFile struct {
	Announce string          `bencode:"announce"`
	Info     torrentFileInfo `bencode:"info"`
}

type torrentFileInfo struct {
	Name         string `bencode:"name"`
	PieceLength  uint64 `bencode:"piece length"`
	PiecesHashes string `bencode:"pieces"`
	Length       uint64 `bencode:"length"`
}

// DecodeTorrentFile - decode .torrent file into go structs
func DecodeTorrentFile(r io.Reader) (result TorrentFile, err error) {
	// Parse torrrent file
	tFile := torrentFile{}
	err = bencode.Unmarshal(r, &tFile)
	if err != nil {
		return
	}

	// Calculate infohash - a hash of bencoded part of an `info` field
	var b bytes.Buffer
	err = bencode.Marshal(&b, tFile.Info)
	if err != nil {
		return
	}
	infoHash := sha1.Sum(b.Bytes())

	// Splitting piecesHashes string into hashes for each piece.
	// SHA-1 hash size is 20 bytes
	chunkNum := len(tFile.Info.PiecesHashes) / 20
	chunks := make([][20]byte, chunkNum)
	for i := 0; i < chunkNum; i++ {
		copy(chunks[i][:], []byte(tFile.Info.PiecesHashes[i*20:i*20+20]))
	}

	return TorrentFile{
		AnnounceURL: tFile.Announce,
		InfoHash:    infoHash,
		FileInfo: DataFileInfo{
			Name:         tFile.Info.Name,
			Length:       tFile.Info.Length,
			PieceLength:  tFile.Info.PieceLength,
			PiecesHashes: chunks,
		},
	}, nil
}
