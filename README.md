# Torbjorn
A chunked file downloader using multiple TOR instances

### How it works
- A bunch of TOR worker instances are created, each as a SOCKS5 proxy
- The target size is specified and broken into chunks - each of which can be retrieved independently (HTTP byte ranges)
- Chunks are queued and requested from the workers and written into the .part file. Restarting is possible as the part file contains state
- When completed "main finished", there will be a <filename>.part file. Use the truncate command and rename as needed.

```
go run cmd/main.go
truncate -s <size> target.part
mv target.part target
```

Sometimes this leaves TOR processes running in the background. A cheap and easy way to fix this:
```
killall sudo
```

Have fun :)
