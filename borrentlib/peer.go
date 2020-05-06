package borrentlib

import (
	"errors"
	"net"
)

// PeerConnectionInfo ...
type PeerConnectionInfo struct {
	amChoking      int
	amInterested   int
	peerChoking    int
	peerInterested int
}

// NewPeerConnectionInfo ...
func NewPeerConnectionInfo() PeerConnectionInfo {
	return PeerConnectionInfo{1, 0, 1, 0}
}

// PeerHandshake ...
func PeerHandshake(infoHash []byte, myPeerID string, peerInfo PeerInfoExt) (handshake, error) {

	clientIP := net.ParseIP(peerInfo.IP)
	if clientIP == nil {
		return handshake{}, errors.New("IP parsing error")
	}

	tcpAddr := net.TCPAddr{IP: clientIP, Port: peerInfo.Port}
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err != nil {
		return handshake{}, err
	}
	defer conn.Close()

	message := createHandshakeMessage(infoHash, myPeerID)

	err = writeHandshake(conn, &message)
	if err != nil {
		return handshake{}, err
	}

	reply, err := readHandshake(conn)
	if err != nil {
		return handshake{}, err
	}

	return reply, nil
}

func createHandshakeMessage(infoHash []byte, peerID string) handshake {
	return handshake{
		// BitTorent protocol v1.0
		StrLength: 19,
		Str:       []byte("BitTorrent protocol"),
		Reserved:  make([]byte, 8), // 8 zeroes
		InfoHash:  infoHash,
		PeerID:    []byte(peerID),
	}
}
