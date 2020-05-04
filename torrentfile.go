package borrentlib

import (
	"io"

	"github.com/jackpal/bencode-go"
)

// TorrentFile - Torrent file parent structure
type TorrentFile struct {
	Announce string
	Info     TorrentFileInfo
}

// TorrentFileInfo - Member of info map in the torrent file
type TorrentFileInfo struct {
	Name         string
	PieceLength  uint64   `piece length`
	PiecesHashes []string `pieces`
	Length       uint64
	Files        []TorrentFileInfoFile
}

// TorrentFileInfoFile - Member of the files list in torrent file info map
type TorrentFileInfoFile struct {
	length uint64
	path   []string
}

func DecodeTorrentFile(r io.Reader) (result TorrentFile, err error) {
	err = bencode.Unmarshal(r, &result)
	return result, err
}
