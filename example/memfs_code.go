// Copyright (c) 2013 The Go Authors. All rights reserved.
// Copyright (c) 2013 memfs Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
