package borrentlib

import (
	"errors"
	"net"
)

// PeerConnectionInfo ...
type PeerConnectionInfo struct {
	AmChoking      int
	AmInterested   int
	PeerChoking    int
	PeerInterested int
}

// NewPeerConnectionInfo ...
func NewPeerConnectionInfo() PeerConnectionInfo {
	return PeerConnectionInfo{1, 0, 1, 0}
}

// DialPeerTCP ...
func DialPeerTCP(peerInfo PeerInfoExt) (net.Conn, error) {
	clientIP := net.ParseIP(peerInfo.IP)
	if clientIP == nil {
		return nil, errors.New("IP parsing error")
	}

	tcpAddr := net.TCPAddr{IP: clientIP, Port: peerInfo.Port}
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// PeerHandshake ...
func PeerHandshake(conn net.Conn, infoHash []byte, myPeerID string) (handshake, error) {
	message := createHandshakeMessage(infoHash, myPeerID)

	err := writeHandshake(conn, &message)
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
