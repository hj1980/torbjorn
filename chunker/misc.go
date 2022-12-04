package chunker

type ChunkerOption interface {
	apply(*Chunker)
}

type withChunkSizeChunkerOption struct {
	s int64
}

func (o *withChunkSizeChunkerOption) apply(c *Chunker) {
	c.ChunkSize = o.s
}

func WithChunkSize(size int64) ChunkerOption {
	return &withChunkSizeChunkerOption{
		s: size,
	}
}
