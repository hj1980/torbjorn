package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/hj1980/torbjorn/chunker"
	"github.com/hj1980/torbjorn/pool"
)

func main() {

	// Change these.
	url := "http://haystak5njsmn2hqkewecpaxetahtwhsbsa64jom2k22z5afxhnpxfid.onion/needle-haystack-big.jpg"
	l := int64(73266)

	fileName, err := getFilename(url)
	if err != nil {
		log.Panic(err)
	}

	fi, err := os.Stat(fileName)
	if err != nil && !os.IsNotExist(err) {
		log.Panic(err)
	}
	if fi != nil {
		log.Panicf("%s already exists", fileName)
	}

	done := make(chan bool)

	var wgRequests sync.WaitGroup

	requests := make(chan *pool.In)

	// Adjust as needed
	responses, err := pool.NewPool(4, requests)
	if err != nil {
		log.Panic(err)
	}

	// Adjust as needed
	c, err := chunker.NewChunker(fileName, l, chunker.WithChunkSize(16*1024))
	if err != nil {
		log.Fatal(err)
	}

	pending, err := c.GetPendingChunks()
	if err != nil {
		log.Panic(err)
	}

	go func() {

		for _, chunk := range pending {
			start, end := c.GetRange(chunk)
			fmt.Printf("Pending: %d (%d-%d)\n", chunk.Id, start, end)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Panic(err)
			}

			req.Header.Add("Accept", "application/octet-stream")

			r := fmt.Sprintf("bytes=%d-%d", start, end)
			req.Header.Add("Range", r)

			wgRequests.Add(1)
			requests <- &pool.In{
				Chunk: chunk,
				Req:   req,
			}

		}
		done <- true
	}()

	go func() {
		for out := range responses {
			fmt.Printf("main response %p\n", out.Chunk)
			c.WriteChunk(out.Chunk, out.Buf)
			wgRequests.Done()
		}
	}()

	<-done

	wgRequests.Wait()
	fmt.Printf("main finished\n")
}

func getFilename(location string) (string, error) {
	addr, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	return path.Base(addr.Path), nil
}
