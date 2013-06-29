// Copyright 2013 Ranveer Kunal
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ranveerkunal/memfs"
)

func main() {
	path := flag.String("path", "/tmp/test/", "")
	fs, err := memfs.New(*path)
	if err != nil {
		log.Fatalf("Failed to create memfs: %s err: %v", *path, err)
	}
	http.Handle("/test/", http.StripPrefix("/test/", http.FileServer(fs)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
