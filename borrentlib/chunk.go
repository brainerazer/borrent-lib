package borrentlib

import (
	"crypto/sha1"
	"os"
)

type ChunkPersister interface {
	PersistChunk(idx int64, offset int64, data []byte) error
	ReadChunkHash(idx uint64) ([]byte, error)
}

type DenseFileDiskChunkPersister struct {
	fileName  string
	file      *os.File
	chunkSize int64
}

func InitDenseFileDiskChunkPersister(fileName string, size uint64, chunkSize uint64) (p *DenseFileDiskChunkPersister, err error) {
	p = &DenseFileDiskChunkPersister{
		fileName:  fileName,
		file:      nil,
		chunkSize: int64(chunkSize),
	}
	f, err := os.Create(p.fileName)
	if err != nil {
		return nil, err
	}

	if err = f.Truncate(p.chunkSize); err != nil {
		return nil, err
	}
	p.file = f
	return
}

func (p *DenseFileDiskChunkPersister) PersistChunk(idx int64, offset int64, data []byte) error {
	_, err := p.file.WriteAt(data, idx*p.chunkSize+offset)
	return err
}

func (p *DenseFileDiskChunkPersister) ReadChunkHash(idx uint64) ([]byte, error) {
	var buf = make([]byte, p.chunkSize)
	_, err := p.file.ReadAt(buf, int64(idx)*p.chunkSize)

	hash := sha1.Sum(buf)
	return hash[:], err
}
