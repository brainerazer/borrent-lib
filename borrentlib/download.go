package borrentlib

import (
	"bytes"
	"fmt"
	"sync"
)

type downloadStatus struct {
	chunksRequested    int
	chunksPersisted    int
	allChunksRequested bool
	chunks             map[int]*torrentChunk
	mu                 sync.Mutex
}

type torrentChunk struct {
	blocksTotal    int
	blocksReceived int
}

const transferBlockSize uint64 = 0x4000

func checkAndPersistJobs(tf TorrentFile, persister ChunkPersister, jobs chan PeerWorkerJob, results chan PeerWorkerResult, signal chan struct{}, status *downloadStatus) {
	for true {
		select {
		case res := <-results:
			if !res.Status {
				// return the job back
				jobs <- res.Job
			}

			// TODO: hash, then save
			persister.PersistChunk(res.Job.ChunkIndex, res.Job.ChunkOffset, res.Chunk)
			status.mu.Lock()
			status.chunks[int(res.Job.ChunkIndex)].blocksReceived++
			received := status.chunks[int(res.Job.ChunkIndex)].blocksReceived
			total := status.chunks[int(res.Job.ChunkIndex)].blocksTotal
			status.mu.Unlock()

			if received == total {
				hash, _ := persister.ReadChunkHash(res.Job.ChunkIndex)
				if !bytes.Equal(hash, tf.FileInfo.PiecesHashes[res.Job.ChunkIndex]) {
					fmt.Printf("Different hashes on chunk %d\n", res.Job.ChunkIndex)
					go sendJobsForChunk(tf, res.Job.ChunkIndex, jobs, status)
				} else {
					fmt.Printf("Hash check succesful on chunk %d\n", res.Job.ChunkIndex)
					status.mu.Lock()
					status.chunksPersisted++
					status.mu.Unlock()
					fmt.Printf("Persisted: %d, Requested: %d\n", status.chunksPersisted, status.chunksRequested)
				}
			}

		default:
			status.mu.Lock()
			if status.allChunksRequested && status.chunksPersisted == status.chunksRequested {
				close(signal) // persister is single, so daijioubu
				status.mu.Unlock()
				return
			}
			status.mu.Unlock()
		}
	}
}

func sendJobsForChunk(tf TorrentFile, chunkIndex uint32, jobs chan<- PeerWorkerJob, status *downloadStatus) {
	chunkSize := tf.FileInfo.PieceLength
	if chunkIndex == uint32(len(tf.FileInfo.PiecesHashes))-1 {
		// Last chunk may be smaller
		chunkSize := tf.FileInfo.Length % tf.FileInfo.PieceLength
		if chunkSize == 0 {
			chunkSize = tf.FileInfo.PieceLength
		}
	}

	status.mu.Lock()
	status.chunks[int(chunkIndex)] = new(torrentChunk)
	status.mu.Unlock()

	for j := uint64(0); j < chunkSize; j += transferBlockSize {
		jobs <- PeerWorkerJob{
			ChunkIndex:  uint32(chunkIndex),
			ChunkOffset: uint32(j),
			ChunkLength: uint32(transferBlockSize),
		}
		status.mu.Lock()
		status.chunks[int(chunkIndex)].blocksTotal++
		status.mu.Unlock()
	}

}

func sendJobs(tf TorrentFile, jobs chan<- PeerWorkerJob, status *downloadStatus) {
	for i := 0; i < len(tf.FileInfo.PiecesHashes); i++ {
		sendJobsForChunk(tf, uint32(i), jobs, status)
		status.mu.Lock()
		status.chunksRequested++
		status.mu.Unlock()
	}

	status.mu.Lock()
	status.allChunksRequested = true
	status.mu.Unlock()
}

//
func DownloadTorrentFile(tf TorrentFile, peerCount int) {
	persister, err := InitSparseFileDiskChunkPersister(tf.FileInfo.Name, uint32(tf.FileInfo.Length), uint32(tf.FileInfo.PieceLength))
	if err != nil {
		return
	}

	peerID, response, err := AnnounceMyself(tf)
	if err != nil {
		return
	}

	fmt.Println(peerID)
	fmt.Println(response)

	jobs := make(chan PeerWorkerJob, 1000)
	results := make(chan PeerWorkerResult)
	signal := make(chan struct{})

	for i := 0; i < peerCount; i++ {
		go PeerWorker(tf.InfoHash, peerID, response.Peers[i], jobs, results)
	}

	status := downloadStatus{}
	status.chunks = make(map[int]*torrentChunk)

	go checkAndPersistJobs(tf, persister, jobs, results, signal, &status)
	go sendJobs(tf, jobs, &status)

	<-signal
	close(jobs)
	persister.Close()
}
