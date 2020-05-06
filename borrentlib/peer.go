package borrentlib

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"gopkg.in/restruct.v1"
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
func PeerHandshake(infoHash []byte, myPeerID string, peerInfo PeerInfoExt) error {
	message := createHandshakeMessage(infoHash, myPeerID)

	clientIP := net.ParseIP(peerInfo.IP)
	if clientIP == nil {
		return errors.New("IP parsing error")
	}

	tcpAddr := net.TCPAddr{IP: clientIP, Port: peerInfo.Port}
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	encoded, err := restruct.Pack(binary.LittleEndian, &message)
	if err != nil {
		return err
	}

	_, err = conn.Write(encoded)
	if err != nil {
		return err
	}

	decoded := make([]byte, 68)

	_, err = io.ReadFull(conn, decoded)
	if err != nil {
		return err
	}

	var reply handshake
	err = restruct.Unpack(decoded, binary.LittleEndian, &reply)

	fmt.Printf("%+v\n", reply)
	fmt.Printf("%s, %s, %v\n", reply.Str, reply.PeerID, reply.InfoHash)

	return nil
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
