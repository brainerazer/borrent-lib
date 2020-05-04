package borrentlib

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
)

// AnnounceMyself - generate a random peerId & perform an announce get request to the tracker.
// Returns generated peerId
func AnnounceMyself(torrentFile TorrentFile) (peerID string, responce string, err error) {
	peerID = generatePeerID()
	announceURL, err := buildAnnounceURL(torrentFile, peerID)
	if err != nil {
		return "", "", err
	}

	resp, err := http.Get(announceURL)
	if err != nil {
		return "", "", err
	}

	responceB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	resp.Body.Close()

	responce = string(responceB)
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
	params.Add("downloaded", "0")
	params.Add("left", strconv.FormatUint(torr.Info.Length, 10))
	base.RawQuery = params.Encode()

	return base.String(), nil
}
