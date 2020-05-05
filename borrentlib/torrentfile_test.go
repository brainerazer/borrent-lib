package borrentlib

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func helperLoadFile(t *testing.T, name string) io.Reader {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

// Testdata created with the help of https://chocobo1.github.io/bencode_online/
func Test_DecodeTorrentFile(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		wantResult TorrentFile
		wantErr    bool
	}{
		{
			"Ubuntu", "ubuntu-20.04-desktop-amd64.iso.torrent",
			TorrentFile{
				AnnounceURL: "https://torrent.ubuntu.com/announce",
				FileInfo: DataFileInfo{
					Name:        "ubuntu-20.04-desktop-amd64.iso",
					Length:      2715254784,
					PieceLength: 1048576,
				},
				InfoHash: []byte{0x9f, 0xc2, 0x0b, 0x9e, 0x98, 0xea, 0x98, 0xb4, 0xa3, 0x5e, 0x62, 0x23, 0x04, 0x1a, 0x5e, 0xf9,
					0x4e, 0xa2, 0x78, 0x09},
			},
			false,
		},
		{
			"Arch", "archlinux-2020.05.01-x86_64.iso.torrent",
			TorrentFile{
				AnnounceURL: "http://tracker.archlinux.org:6969/announce",
				FileInfo: DataFileInfo{
					Name:        "archlinux-2020.05.01-x86_64.iso",
					Length:      683671552,
					PieceLength: 524288,
				},
				InfoHash: []byte{0xf9, 0x5c, 0x37, 0x1d, 0x56, 0x09, 0xd1, 0x5f, 0x66, 0x15, 0x13, 0x9b, 0xe8, 0x4e, 0xdb, 0xb5,
					0xb9, 0x4a, 0x79, 0xbc},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			torrentFile := helperLoadFile(t, tt.filename)

			gotResult, err := DecodeTorrentFile(torrentFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeTorrentFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Ignore PiecesHashes key for now - it's very big
			opt := cmp.FilterPath(
				func(p cmp.Path) bool {
					return p.String() == "FileInfo.PiecesHashes"
				},
				cmp.Ignore(),
			)
			if diff := cmp.Diff(tt.wantResult, gotResult, opt); diff != "" {
				t.Errorf("DecodeTorrentFile()  mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
