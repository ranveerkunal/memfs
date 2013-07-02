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
~/ranveerkunal/memfs % go test memfs_test.go -bench=. -cpu=4 -parallel=4
temp dir: /tmp/memfs406771321
writing small file
writing big file
ready to benchmark ...
testing: warning: no tests to run
PASS
BenchmarkNonExistentMemFS-4      5000000               700 ns/op
BenchmarkNonExistentDiskFS-4      500000              3996 ns/op
BenchmarkSmallFileMemFS-4          10000            111634 ns/op
BenchmarkSmallFileDiskFS-4         10000            128475 ns/op
BenchmarkBigFileMemFS-4               20          83455262 ns/op
BenchmarkBigFileDiskFS-4              20          96320175 ns/op
ok      command-line-arguments  26.610s
</pre>

<pre>
Benchmark on linux:
~/ranveerkunal/memfs % go test memfs_test.go -bench=. -cpu=4 -parallel=4
temp dir: /tmp/memfs945391658
writing small file
writing big file
ready to benchmark ...
testing: warning: no tests to run
PASS
BenchmarkNonExistentMemFS-4       500000              7918 ns/op
BenchmarkNonExistentDiskFS-4      100000             16862 ns/op
BenchmarkSmallFileMemFS-4           2000            811382 ns/op
BenchmarkSmallFileDiskFS-4          2000            836498 ns/op
BenchmarkBigFileMemFS-4                5        4629695200 ns/op
BenchmarkBigFileDiskFS-4               1       11134569000 ns/op
ok      command-line-arguments  109.350s
</pre>
