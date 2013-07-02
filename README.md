memfs
=====

Implementation of http.FileSystem where the files stay in memory.<br>
It uses [<b>fsnotify</b>](https://github.com/howeyc/fsnotify) to keep the cache updated.

Example:
<code><pre>
github.com/ranveerkunal/memfs/example $ go build memfs_code.go
github.com/ranveerkunal/memfs/example $ ./memfs_code
</pre></code>

[http://localhost:9999/memfs](http://localhost:9999/memfs)

```go
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ranveerkunal/memfs"
)

func main() {
	path := flag.String("path", "./", "")
	addr := flag.String("addr", ":9999", "")
	verbose := flag.Bool("verbose", true, "")
	flag.Parse()

	fs, err := memfs.New(*path)
	if err != nil {
		log.Fatalf("Failed to create memfs: %s err: %v", *path, err)
	}

	if (*verbose) {
		log.Printf("logging to stderr ...")
		memfs.SetLogger(memfs.Verbose)
	}

	http.Handle("/memfs/", http.StripPrefix("/memfs/", http.FileServer(fs)))

	log.Printf("path: %s addr:%s", *path, *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
```
<pre>
Benchmark on mac: darwin 64
~/gocode/src/github.com/ranveerkunal/memfs % go test memfs_test.go -bench=. -cpu=4 -parallel=4
temp dir: /tmp/memfs731592791
writing big file
ready to benchmark ...
testing: warning: no tests to run
PASS
BenchmarkNonExistentMemFS-4       500000              2779 ns/op
BenchmarkNonExistentDiskFS-4      200000              7513 ns/op
BenchmarkSmallFileMemFS-4          10000            129763 ns/op
BenchmarkSmallFileDiskFS-4         10000            135561 ns/op
BenchmarkBigFileMemFS-4               20          88782691 ns/op
BenchmarkBigFileDiskFS-4              20          86461808 ns/op
ok      command-line-arguments  20.301s
</pre>

<pre>
Benchmark on linux:
~/gocode/src/github.com/ranveerkunal/memfs % go test memfs_test.go -bench=. -cpu=4 -parallel=4
temp dir: /tmp/memfs016684973
writing small file
writing big file
ready to benchmark ...
testing: warning: no tests to run
PASS
BenchmarkNonExistentMemFS-4        50000             43828 ns/op
BenchmarkNonExistentDiskFS-4       50000             37428 ns/op
BenchmarkSmallFileMemFS-4           1000           1763882 ns/op
BenchmarkSmallFileDiskFS-4          1000           1507493 ns/op
BenchmarkBigFileMemFS-4                1        1550468000 ns/op
BenchmarkBigFileDiskFS-4               1        1261333000 ns/op
ok      command-line-arguments  63.824s
</pre>
