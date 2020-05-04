package borrentlib

import (
	"bytes"
	"crypto/sha1"
	"io"

	"github.com/jackpal/bencode-go"
)

// TorrentFile - Torrent file parent structure
type TorrentFile struct {
	Announce string          `bencode:"announce"`
	Info     TorrentFileInfo `bencode:"info"`
	InfoHash [20]byte        // Calculated, not read
}

// TorrentFileInfo - Member of info map in the torrent file.
// Only for single-file torrents right now
type TorrentFileInfo struct {
	Name         string `bencode:"name"`
	PieceLength  uint64 `bencode:"piece length"`
	PiecesHashes string `bencode:"pieces"`
	Length       uint64 `bencode:"length"`
}

// DecodeTorrentFile - decode .torrent file into go structs
func DecodeTorrentFile(r io.Reader) (result TorrentFile, err error) {
	err = bencode.Unmarshal(r, &result)
	if err != nil {
		return
	}
	var b bytes.Buffer
	bencode.Marshal(&b, result.Info)
	result.InfoHash = sha1.Sum(b.Bytes())
	return
}
