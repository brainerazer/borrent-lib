package borrentlib

import (
	"fmt"
	"net"
	"sync"
	"time"
)

//
type PeerWorkerJob struct {
	ChunkIndex  uint32
	ChunkOffset uint32
	ChunkLength uint32
}

//
type PeerWorkerResult struct {
	Job    PeerWorkerJob
	Status bool
	Chunk  []byte
}

//
type pendingJob struct {
	job        PeerWorkerJob
	expiration time.Time
}

//
type PeerState struct {
	PeerConnectionInfo
	peerID      []byte
	conn        net.Conn
	batchSize   int
	pendingJobs []pendingJob
	mu          sync.Mutex
}

//
func PeerWorker(infoHash []byte, ownPeerID string, peerInfo PeerInfoExt, jobs chan PeerWorkerJob, results chan<- PeerWorkerResult) {
	conn, err := DialPeerTCP(peerInfo)
	if err != nil {
		return
	}

	hs, err := PeerHandshake(conn, infoHash, ownPeerID)
	if err != nil {
		return
	}
	fmt.Printf("Connected to the peer with ID: %s\n", hs.PeerID)

	peerState := PeerState{}
	peerState.peerID = hs.PeerID
	peerState.conn = conn
	peerState.PeerConnectionInfo = NewPeerConnectionInfo()
	peerState.batchSize = 100

	// lets unchoke the peer
	err = WriteMessage(conn, Interested{})
	if err != nil {
		return
	}

	go peerConnReader(&peerState, results)
	go peerConnWriter(&peerState, jobs)
	go retryPendingJobs(&peerState, jobs)
}

func retryPendingJobs(peerState *PeerState, jobs chan PeerWorkerJob) {
	for true {
		peerState.mu.Lock()
		cleanedUpPending := []pendingJob{}
		for _, job := range peerState.pendingJobs {
			if job.expiration.Before(time.Now()) {
				jobs <- job.job
			} else {
				cleanedUpPending = append(cleanedUpPending, job)
			}
		}
		peerState.pendingJobs = cleanedUpPending
		peerState.mu.Unlock()

		time.Sleep(200 * time.Millisecond)
	}
}

func peerConnReader(peerState *PeerState, results chan<- PeerWorkerResult) {
	for true {
		msg, err := ReadMessage(peerState.conn)
		if err != nil {
			return
		}

		switch v := msg.(type) {
		case Unchoke:
			peerState.mu.Lock()
			peerState.PeerChoking = 0
			peerState.mu.Unlock()
		case Choke:
			peerState.mu.Lock()
			peerState.PeerChoking = 1
			peerState.mu.Unlock()
		case Interested:
			peerState.mu.Lock()
			peerState.PeerInterested = 1
			peerState.mu.Unlock()
		case notInterested:
			peerState.mu.Lock()
			peerState.PeerInterested = 0
			peerState.mu.Unlock()

		case Piece:
			// fmt.Printf("begin: %d, idx: %d, block: %v..., peer: %s\n", v.Begin, v.Index, v.Block[:5], peerState.peerID)
			results <- PeerWorkerResult{
				Job: PeerWorkerJob{
					ChunkIndex:  v.Index,
					ChunkOffset: v.Begin,
					ChunkLength: uint32(len(v.Block)),
				},
				Status: true,
				Chunk:  v.Block,
			}
			peerState.mu.Lock()
			cleanedUpPending := []pendingJob{}
			for _, job := range peerState.pendingJobs {
				if job.job.ChunkIndex != v.Index || job.job.ChunkOffset != v.Begin {
					cleanedUpPending = append(cleanedUpPending, job)
				}
			}
			peerState.pendingJobs = cleanedUpPending
			peerState.mu.Unlock()
		}
	}
}

func peerConnWriter(peerState *PeerState, jobs chan PeerWorkerJob) {
	for true {
		peerState.mu.Lock()
		isChoked := peerState.PeerChoking
		pendingJobsSize := len(peerState.pendingJobs)
		batchSize := peerState.batchSize
		peerState.mu.Unlock()

		if isChoked == 1 {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if batchSize == pendingJobsSize {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		for i := 0; i < batchSize-pendingJobsSize; i++ {
			job := <-jobs
			err := WriteMessage(peerState.conn, Request{
				Index:  job.ChunkIndex,
				Begin:  job.ChunkOffset,
				Length: job.ChunkLength,
			})

			if err != nil {
				jobs <- job
				return
			}
			peerState.mu.Lock()
			peerState.pendingJobs = append(peerState.pendingJobs, pendingJob{
				job:        job,
				expiration: time.Now().Add(30 * time.Second),
			})
			peerState.mu.Unlock()
		}

	}
}
