// Copyright (c) 2013 The Go Authors. All rights reserved.
// Copyright (c) 2013 memfs Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"os"
	"time"
)

type memFileInfo struct {
	content []byte
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	path    string
	fs      *memFileSystem
}

func (fi *memFileInfo) Name() string {
	return fi.name
}

func (fi *memFileInfo) Size() int64 {
	return fi.size
}

func (fi *memFileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi *memFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *memFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi *memFileInfo) Sys() interface{} {
	return nil
}

type fileInfoSorter struct {
	infos []os.FileInfo
}

func (s *fileInfoSorter) Len() int {
	return len(s.infos)
}

func (s *fileInfoSorter) Swap(i, j int) {
	s.infos[i], s.infos[j] = s.infos[j], s.infos[i]
}

func (s *fileInfoSorter) Less(i, j int) bool {
	return s.infos[i].Name() < s.infos[j].Name()
}
