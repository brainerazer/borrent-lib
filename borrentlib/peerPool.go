package borrentlib

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

//
type PeerPoolEntry struct {
	peerInfo PeerInfoExt
	ConnInfo PeerConnectionInfo
	Conn     net.Conn
	isUsed   bool
}

//
type PeerPool struct {
	infoHash []byte
	myPeerID string
	peers    []PeerPoolEntry
	mu       sync.Mutex
}

//
func (p *PeerPoolEntry) Return() {
	p.isUsed = false
}

//
func InitPeerPool(infoHash []byte, peerID string) (p PeerPool) {
	return PeerPool{
		infoHash: infoHash,
		myPeerID: peerID,
	}
}

//
func (p *PeerPool) ConnectPeer(peerInfo PeerInfoExt) error {
	conn, err := DialPeerTCP(peerInfo)
	if err != nil {
		return err
	}

	hs, err := PeerHandshake(conn, p.infoHash, p.myPeerID)
	if err != nil {
		return err
	}
	fmt.Printf("Connected to the peer with ID: %s\n", hs.PeerID)

	entry := PeerPoolEntry{
		peerInfo: peerInfo,
		ConnInfo: NewPeerConnectionInfo(),
		Conn:     conn,
	}

	p.mu.Lock()
	p.peers = append(p.peers, entry)
	p.mu.Unlock()

	return nil
}

//
func (p *PeerPool) GetPeer() (*PeerPoolEntry, error) {
	p.mu.Lock()
	for idx := range p.peers {
		if !p.peers[idx].isUsed {
			p.peers[idx].isUsed = true
			p.mu.Unlock()
			return &p.peers[idx], nil
		}
	}
	p.mu.Unlock()
	return nil, errors.New("Cannot get a free peer")
}

//
func (p *PeerPool) Close() {
	p.mu.Lock()
	for _, entry := range p.peers {
		entry.Conn.Close()
	}
	p.peers = []PeerPoolEntry{}
	p.mu.Unlock()
}
