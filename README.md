# gstat
HTTP Benchmarking tool written in Golang

## Usage
```
$ go build -o gstat src/main.go
$ ./gstat -e http://google.com/
Running 5s benchmark test @ http://google.com/
 2299 concurrent requests / 8 threads
 Latency statistics:
    Max Min Avg
   3346 175 617
```

