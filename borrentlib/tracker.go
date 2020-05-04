package borrentlib

import (
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

type TrackerResponce struct {
	FailureReason string        `bencode:"failure reason"`
	Interval      int           `bencode:"interval"`
	Peers         []PeerInfoExt `bencode:"peers"`
}

type PeerInfoExt struct {
	PeerID string `bencode:"peer id"`
	Ip     string `bencode:"ip"`
	Port   int    `bencode:"port"`
}

// AnnounceMyself - generate a random peerId & perform an announce get request to the tracker.
// Returns generated peerId
func AnnounceMyself(torrentFile TorrentFile) (peerID string, responce TrackerResponce, err error) {
	peerID = generatePeerID()
	announceURL, err := buildAnnounceURL(torrentFile, peerID)
	if err != nil {
		return
	}

	resp, err := http.Get(announceURL)
	if err != nil {
		return
	}

	err = bencode.Unmarshal(resp.Body, &responce)
	if err != nil {
		return
	}
	resp.Body.Close()

	return
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generatePeerID() string {
	b := make([]rune, 20)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func buildAnnounceURL(torr TorrentFile, peerID string) (announceURL string, err error) {
	base, err := url.Parse(torr.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("info_hash", string(torr.InfoHash[:]))
	params.Add("peer_id", peerID)
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	// params.Add("compact", "1")
	params.Add("downloaded", "0")
	params.Add("left", strconv.FormatUint(torr.Info.Length, 10))
	base.RawQuery = params.Encode()

	return base.String(), nil
}
