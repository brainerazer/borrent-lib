package borrentlib

import (
	"math/rand"
	"testing"
)

func Test_generatePeerID(t *testing.T) {
	rand.Seed(42) // for reproducibility
	if got := generatePeerID(); len(got) != 20 {
		t.Errorf(
			"generatePeerID() should return 20 bytes strings, actual string len: %d, string: %s",
			len(got), got,
		)
	}
}

// Note: breaks on URL query params reorder, need a better test
func Test_buildAnnounceURL(t *testing.T) {
	type args struct {
		torr   TorrentFile
		peerID string
	}
	tests := []struct {
		name            string
		args            args
		wantAnnounceURL string
		wantErr         bool
	}{
		{
			"Builds a good announce URL",
			args{
				TorrentFile{
					AnnounceURL: "http://testtracker.com/announce",
					InfoHash:    []byte("aaaaaaaaaaaaaaaaaaaa"),
					FileInfo: DataFileInfo{
						Length: 80,
					},
				},
				"abcfghtyujskfteyrwgd",
			},
			"http://testtracker.com/announce?downloaded=0&info_hash=aaaaaaaaaaaaaaaaaaaa&left=80" +
				"&peer_id=abcfghtyujskfteyrwgd&port=6881&uploaded=0",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAnnounceURL, err := buildAnnounceURL(tt.args.torr, tt.args.peerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildAnnounceURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAnnounceURL != tt.wantAnnounceURL {
				t.Errorf("buildAnnounceURL() = %v, want %v", gotAnnounceURL, tt.wantAnnounceURL)
			}
		})
	}
}
