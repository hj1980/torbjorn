package pool

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/hj1980/torbjorn/chunker"
	"github.com/hj1980/torbjorn/tor"
)

const BasePort = 20000

type In struct {
	Req   *http.Request
	Chunk *chunker.Chunk
}
type Out struct {
	Buf   []byte
	Chunk *chunker.Chunk
}

type pool struct {
	wgWorkers sync.WaitGroup
}

func NewPool(workerCount int, requests chan *In) (chan *Out, error) {
	p := &pool{}

	responses := make(chan *Out, workerCount)

	workerQueue := make(chan *tor.WorkItem)

	// TODO: cleanup created workers if something fails
	for i := 0; i < workerCount; i++ {

		p.wgWorkers.Add(1)
		go func(id int) {

			// TODO: this won't return an error as we are in a goroutine
			workerComplete, err := tor.NewWorker(BasePort+id, workerQueue)
			if err != nil {
				log.Fatal(err)
			}
			// if err != nil {
			// 	return nil, err
			// }

			for wi := range workerComplete {

				if wi.Err != nil {
					fmt.Printf("worker %d returned an erroring workitem: %+v\n", id, wi.Err)

					time.Sleep(time.Second)

					wi.Err = nil
					workerQueue <- wi
					continue
				}

				// Return response back to main in some channel
				fmt.Printf("worker %d returned a response\n", id)
				responses <- &Out{
					Chunk: wi.Chunk,
					Buf:   wi.Buf,
				}

			}

			fmt.Printf("worker %d has closed the workerComplete channel\n", id)

			// This is where things should tidy up
			p.wgWorkers.Done()

		}(i)

	}

	go func() {
		for in := range requests {
			workerQueue <- &tor.WorkItem{
				Chunk: in.Chunk,
				Req:   in.Req,
			}
		}
	}()

	return responses, nil
}
