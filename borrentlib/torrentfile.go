package borrentlib

import (
	"bytes"
	"crypto/sha1"
	"io"

	"github.com/jackpal/bencode-go"
)

// TorrentFile - Torrent file parent structure
type TorrentFile struct {
	Announce string
	Info     TorrentFileInfo
	InfoHash [20]byte
}

// TorrentFileInfo - Member of info map in the torrent file
type TorrentFileInfo struct {
	Name         string `bencode:"name"`
	PieceLength  uint64 `bencode:"piece length"`
	PiecesHashes string `bencode:"pieces"`
	Length       uint64 `bencode:"length"`
	// Files        []TorrentFileInfoFile
}

// TorrentFileInfoFile - Member of the files list in torrent file info map
type TorrentFileInfoFile struct {
	length uint64
	path   []string
}

// DecodeTorrentFile - decode .torrent file into go structs
func DecodeTorrentFile(r io.Reader) (result TorrentFile, err error) {
	err = bencode.Unmarshal(r, &result)
	var b bytes.Buffer
	bencode.Marshal(&b, result.Info)
	result.InfoHash = sha1.Sum(b.Bytes())
	return result, err
}
