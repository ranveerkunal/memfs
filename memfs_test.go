// Copyright (c) 2013 The Go Authors. All rights reserved.
// Copyright (c) 2013 memfs Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs_test

import (
	"testing"

	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"time"

	"github.com/ranveerkunal/memfs"
)

func setUp() string {
	name, err := ioutil.TempDir("/tmp", "memfs")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("temp dir: %s\n", name)
	return name
}

func createBigFile(name string) string {
	bigFile := name + "/big.txt"
	fmt.Printf("writing big file\n")
	cmd := exec.Command("dd", "if=/dev/urandom", "of=" + bigFile, "bs=1048576", "count=100")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return bigFile
}

func createSmallFile(name string) string {
	smallFile := name + "/small.txt"
	fmt.Printf("writing small file\n")
	cmd := exec.Command("dd", "if=/dev/urandom", "of=" + smallFile, "bs=64", "count=10")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return smallFile
}

func startServer(fs http.FileSystem, prefix, addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	handler := http.NewServeMux()
	handler.Handle(prefix, http.StripPrefix(prefix, http.FileServer(fs)))

	err = http.Serve(ln, handler)
	if err != nil {
		log.Fatal(err)
	}
}

var (
	memFS  http.FileSystem
	diskFS http.FileSystem
	err    error
)

func init() {
	name := setUp()
	createSmallFile(name)
	createBigFile(name)

	memFS, err = memfs.New(name)
	if err != nil {
		log.Fatal(err)
	}
	go startServer(memFS, "/memfs/", ":6666")

	diskFS = http.Dir(name)
	go startServer(diskFS, "/diskfs/", ":7777")

	time.Sleep(5 * time.Second)
	fmt.Printf("ready to benchmark ...\n")
}

func BenchmarkNonExistentMemFS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		memFS.Open("./non_existent.%d")
	}
}

func BenchmarkNonExistentDiskFS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		diskFS.Open("./non_existent.%d")
	}
}

func BenchmarkSmallFileMemFS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res, err := http.Get("http://localhost:6666/memfs/small.txt")
		if err != nil {
			b.Fatal(err)
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSmallFileDiskFS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res, err := http.Get("http://localhost:7777/diskfs/small.txt")
		if err != nil {
			b.Fatal(err)
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBigFileMemFS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res, err := http.Get("http://localhost:6666/memfs/big.txt")
		if err != nil {
			b.Fatal(err)
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBigFileDiskFS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res, err := http.Get("http://localhost:7777/diskfs/big.txt")
		if err != nil {
			b.Fatal(err)
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			b.Fatal(err)
		}
	}
}
