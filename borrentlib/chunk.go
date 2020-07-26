package borrentlib

import (
	"os"
)

type ChunkPersister interface {
	PersistChunk(idx int64, data []byte) error
	ReadChunkHash(idx int64) ([]byte, error)
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

func (p *DenseFileDiskChunkPersister) PersistChunk(idx int64, data []byte) error {
	_, err := p.file.WriteAt(data, idx*p.chunkSize)
	return err
}

func (p *DenseFileDiskChunkPersister) ReadChunkHash(idx int64) ([]byte, error) {
	var buf []byte
	_, err := p.file.ReadAt(buf, idx*p.chunkSize)
	return buf, err
}
