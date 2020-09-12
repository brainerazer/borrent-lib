package borrentlib

import (
	"crypto/sha1"
	"os"
)

type ChunkPersister interface {
	PersistChunk(idx uint32, offset uint32, data []byte) error
	ReadChunkHash(idx uint32) ([]byte, error)
}

type SparseFileDiskChunkPersister struct {
	fileName  string
	file      *os.File
	chunkSize uint32
}

func InitSparseFileDiskChunkPersister(fileName string, size uint32, chunkSize uint32) (p *SparseFileDiskChunkPersister, err error) {
	p = &SparseFileDiskChunkPersister{
		fileName:  fileName,
		file:      nil,
		chunkSize: chunkSize,
	}
	f, err := os.Create(p.fileName)
	if err != nil {
		return nil, err
	}

	p.file = f
	return
}

func (p *SparseFileDiskChunkPersister) PersistChunk(idx uint32, offset uint32, data []byte) error {
	_, err := p.file.WriteAt(data, int64(idx*p.chunkSize+offset))
	return err
}

func (p *SparseFileDiskChunkPersister) ReadChunkHash(idx uint32) ([]byte, error) {
	var buf = make([]byte, p.chunkSize)
	_, err := p.file.ReadAt(buf, int64(idx*p.chunkSize))

	hash := sha1.Sum(buf)
	return hash[:], err
}
