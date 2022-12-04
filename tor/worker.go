package tor

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"syscall"
	"time"

	"github.com/hj1980/torbjorn/chunker"
	"golang.org/x/net/proxy"
)

type WorkItem struct {
	Chunk *chunker.Chunk
	Req   *http.Request
	Buf   []byte
	Err   error
}

func NewWorker(port int, in chan *WorkItem) (chan *WorkItem, error) {

	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", port), nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	client := &http.Client{
		Transport: transport,
	}

	cmd := exec.Command("sudo", "tor", "--socksport", fmt.Sprintf("%d", port), "--datadir", fmt.Sprintf("/var/lib/tor/%d", port))
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	go cmd.Wait()

	time.Sleep(time.Second)

	out := make(chan *WorkItem)

	go func() {
		for wi := range in {
			fmt.Printf("worker %d %p in\n", port, wi.Req)

			res, err := client.Do(wi.Req)
			if err != nil {
				fmt.Printf("worker %d %p out error\n", port, wi.Req)
				wi.Err = err
				out <- wi
				continue
			}

			if wi.Req.Method == "GET" && res.StatusCode != 206 {
				fmt.Printf("worker %d %p out error (non-206)\n", port, wi.Req)
				wi.Err = fmt.Errorf("http code %d returned", res.StatusCode)
				out <- wi
				continue
			}

			buf, err := io.ReadAll(res.Body)
			if err != nil {
				wi.Err = fmt.Errorf("error during readall %s", err)
				out <- wi
				continue
			}
			wi.Buf = buf
			out <- wi

		}

		fmt.Printf("worker %d observed in closed\n", port)
		err := cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			fmt.Printf("Error during process term: %s\n", err)
		}

	}()

	return out, nil
}
