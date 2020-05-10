package borrentlib

import (
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

// TrackerResponce ...
type TrackerResponce struct {
	FailureReason string        `bencode:"failure reason"`
	Interval      int           `bencode:"interval"`
	Complete      int           `bencode:"complete"`
	Incomplete    int           `bencode:"incomplete"`
	Peers         []PeerInfoExt `bencode:"peers"`
}

// PeerInfoExt ...
type PeerInfoExt struct {
	PeerID string `bencode:"peer id"`
	IP     string `bencode:"ip"`
	Port   int    `bencode:"port"`
}

// AnnounceMyself - generate a random peerId & perform an announce get Request to the tracker.
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
	defer resp.Body.Close()

	err = bencode.Unmarshal(resp.Body, &responce)
	if err != nil {
		return
	}
	if responce.FailureReason != "" {
		return peerID, responce, errors.New(responce.FailureReason)
	}

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
	base, err := url.Parse(torr.AnnounceURL)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("info_hash", string(torr.InfoHash[:]))
	params.Add("peer_id", peerID)
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", strconv.FormatUint(torr.FileInfo.Length, 10))
	base.RawQuery = params.Encode()

	return base.String(), nil
}
