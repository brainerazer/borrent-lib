package borrentlib

import (
	"fmt"
)

func checkAndPersistJobs(tf TorrentFile, persister ChunkPersister, jobs chan<- PeerWorkerJob, results <-chan PeerWorkerResult) {
	for res := range results {
		if !res.Status {
			// return the job back
			jobs <- res.Job
		}

		// TODO: hash, then save
		persister.PersistChunk(res.Job.ChunkIndex, res.Job.ChunkOffset, res.Chunk)
		// hash, _ := persister.ReadChunkHash(res.Job.ChunkIndex)
		// if !bytes.Equal(hash, tf.FileInfo.PiecesHashes[res.Job.ChunkIndex]) {
		// 	fmt.Printf("Different hashes on chunk %d\n", res.Job.ChunkIndex)
		// }
	}
}

func sendJobs(tf TorrentFile, jobs chan<- PeerWorkerJob) {
	var transferBlockSize uint64 = 0x4000

	for i := 0; i < len(tf.FileInfo.PiecesHashes)-1; i++ {
		for j := uint64(0); j < tf.FileInfo.PieceLength; j += transferBlockSize {
			jobs <- PeerWorkerJob{
				ChunkIndex:  uint32(i),
				ChunkOffset: uint32(j),
				ChunkLength: uint32(transferBlockSize),
			}
		}
	}

	// Last cunk may be smaller
	for j := uint64(0); j < tf.FileInfo.Length%tf.FileInfo.PieceLength; j += transferBlockSize {
		jobs <- PeerWorkerJob{
			ChunkIndex:  uint32(len(tf.FileInfo.PiecesHashes) - 1),
			ChunkOffset: uint32(j),
			ChunkLength: uint32(transferBlockSize),
		}
	}
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

	jobs := make(chan PeerWorkerJob)
	results := make(chan PeerWorkerResult)

	for i := 0; i < peerCount; i++ {
		go PeerWorker(tf.InfoHash, peerID, response.Peers[i], jobs, results)
	}

	go sendJobs(tf, jobs)
	go checkAndPersistJobs(tf, persister, jobs, results)

	select {}
}
