package chunker

import (
	"bytes"
	"encoding/binary"
	"log"
)

func (c *Chunker) readChunkSize() (int64, error) {
	offset := c.ContentLength

	var i int64

	buf := make([]byte, 8)
	_, err := c.PartFile.ReadAt(buf, offset)
	if err != nil {
		return i, err
	}
	// log.Printf("Read: %x", buf)

	r := bytes.NewReader(buf)
	err = binary.Read(r, binary.LittleEndian, &i)
	return i, err
}

func (c *Chunker) readChunkState(ch *Chunk) (byte, error) {
	offset := c.ContentLength + 8 + ch.Id

	buf := make([]byte, 1)
	// log.Printf("reading at %x", offset)
	_, err := c.PartFile.ReadAt(buf, offset)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (c *Chunker) WriteChunk(ch *Chunk, chunk []byte) error {
	offset := c.getChunkOffset(ch)

	c.Lock()
	defer c.Unlock()

	_, err := c.PartFile.WriteAt(chunk, offset)
	if err != nil {
		return err
	}
	log.Printf("wrote chunk %d (%d bytes at %x)", ch.Id, len(chunk), offset)
	return c.writeChunkState(ch, 1)
}

func (c *Chunker) writeChunkSize() error {
	offset := c.ContentLength

	c.Lock()
	defer c.Unlock()

	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, c.ChunkSize)
	if err != nil {
		return err
	}

	_, err = c.PartFile.WriteAt(buf.Bytes(), offset)
	return err
}

func (c *Chunker) writeChunkState(ch *Chunk, state byte) error {
	offset := c.ContentLength + 8 + ch.Id

	buf := make([]byte, 1)
	buf[0] = state
	// log.Printf("writing %x at %x", buf, offset)
	_, err := c.PartFile.WriteAt(buf, offset)
	return err
}
