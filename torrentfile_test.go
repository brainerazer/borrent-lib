package borrentlib

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
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
func Test_decodeTorrentFile(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		wantResult TorrentFile
		wantErr    bool
	}{
		{
			"Ubuntu", "ubuntu-20.04-desktop-amd64.iso.torrent",
			TorrentFile{
				Announce: "https://torrent.ubuntu.com/announce",
				Info: TorrentFileInfo{
					Name:        "ubuntu-20.04-desktop-amd64.iso",
					Length:      2715254784,
					PieceLength: 1048576,
				},
			},
			false,
		},
		{
			"Arch", "archlinux-2020.05.01-x86_64.iso.torrent",
			TorrentFile{
				Announce: "http://tracker.archlinux.org:6969/announce",
				Info: TorrentFileInfo{
					Name:        "archlinux-2020.05.01-x86_64.iso",
					Length:      683671552,
					PieceLength: 524288,
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			torrentFile := helperLoadFile(t, tt.filename)

			gotResult, err := decodeTorrentFile(torrentFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeTorrentFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("decodeTorrentFile() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
