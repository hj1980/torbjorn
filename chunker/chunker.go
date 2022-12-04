package chunker

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type Chunk struct {
	Id int64
}

type Chunker struct {
	sync.Mutex
	Chunks        []*Chunk
	ChunkSize     int64
	ContentLength int64
	PartFile      *os.File
}

func (c *Chunker) GetRange(ch *Chunk) (start int64, end int64) {
	return c.getChunkOffset(ch), c.getChunkOffset(ch) + c.getChunkLength(ch) - 1
}

func (c *Chunker) generateChunks() {
	id := int64(0)
	for offset := int64(0); offset < c.ContentLength; offset += c.ChunkSize {
		chunk := &Chunk{
			Id: id,
		}
		c.Chunks = append(c.Chunks, chunk)
		id += 1
	}
}

func (c *Chunker) getChunkLength(ch *Chunk) int64 {
	length := c.ChunkSize
	offset := c.getChunkOffset(ch)
	if c.ContentLength < offset+c.ChunkSize {
		length = c.ContentLength - offset
	}
	return length
}

func (c *Chunker) getChunkOffset(ch *Chunk) int64 {
	return c.ChunkSize * ch.Id
}

func (c *Chunker) Close() error {
	err := c.PartFile.Sync()
	if err != nil {
		return err
	}
	return c.PartFile.Close()
}

func NewChunker(name string, contentLength int64, opts ...ChunkerOption) (*Chunker, error) {

	var err error

	c := &Chunker{
		ChunkSize:     65536,
		ContentLength: contentLength,
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	partName := fmt.Sprintf("%s.part", name)
	c.PartFile, err = os.OpenFile(partName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	size, err := c.readChunkSize()
	if err != nil && err != io.EOF {
		return nil, err
	}

	if size > 0 {
		c.ChunkSize = size
	}

	c.generateChunks()

	log.Printf("Chunker configured: %+v\n", c)

	totalLength := contentLength + int64(len(c.Chunks)) + 8

	// log.Printf("there are %d chunks", len(c.Chunks))
	// log.Printf("the content length is 0x%x bytes", contentLength)
	// log.Printf("the total length is %d bytes", totalLength)

	err = c.PartFile.Truncate(totalLength)
	if err != nil {
		return nil, err
	}

	err = c.writeChunkSize()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Chunker) GetPendingChunks() ([]*Chunk, error) {

	pending := []*Chunk{}
	for _, chunk := range c.Chunks {
		s, err := c.readChunkState(chunk)
		if err != nil {
			return nil, err
		}
		log.Printf("state for chunk %d is %x\n", chunk.Id, s)

		if s == 0 {
			pending = append(pending, chunk)
		}
	}

	return pending, nil
}
