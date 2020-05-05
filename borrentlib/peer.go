package borrentlib

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// PeerConnectionInfo ...
type PeerConnectionInfo struct {
	amChoking      int
	amInterested   int
	peerChoking    int
	peerInterested int
}

type handshake struct {
	StrLength uint8
	Str       [19]byte
	Reserved  [8]byte
	InfoHash  [20]byte
	PeerID    [20]byte
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

	err = binary.Write(conn, binary.LittleEndian, &message)
	if err != nil {
		return err
	}

	// reply := make([]byte, 68)
	reply := handshake{}

	err = binary.Read(conn, binary.LittleEndian, &reply)
	if err != nil {
		return err
	}

	fmt.Println(peerInfo)
	fmt.Printf("%+v\n", reply)
	fmt.Printf("%s, %s, %v\n", reply.Str, reply.PeerID, reply.InfoHash)

	return nil
}

func createHandshakeMessage(infoHash []byte, peerID string) handshake {
	// Thanks for rejecting https://github.com/golang/go/issues/36890 !
	var strArr [19]byte
	copy(strArr[:], "BitTorrent protocol")
	var ihArr [20]byte
	copy(ihArr[:], infoHash)
	var peerIDArr [20]byte
	copy(peerIDArr[:], peerID)

	return handshake{
		// BitTorent protocol v1.0
		StrLength: 19,
		Str:       strArr,
		Reserved:  [8]byte{}, // 8 zeroes
		InfoHash:  ihArr,
		PeerID:    peerIDArr,
	}
}
