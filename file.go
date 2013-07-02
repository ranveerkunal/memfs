// Copyright (c) 2013 The Go Authors. All rights reserved.
// Copyright (c) 2013 memfs Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package memfs creates a watched in memory filesystem.
package memfs

import (
	"errors"
	"io"
	"os"
	"sort"
	"strings"
)

type memFile struct {
	fi        *memFileInfo
	offset    int64
	dirOffset int
}

func (f *memFile) Close() error {
	return nil
}

func (f *memFile) Stat() (os.FileInfo, error) {
	return f.fi, nil
}

func (f *memFile) Readdir(count int) ([]os.FileInfo, error) {
	infos := []os.FileInfo{}
	prefix := f.fi.path
	skip := 0
	for path, fi := range f.fi.fs.cache {
		if strings.HasPrefix(path, prefix) || prefix == "." {
			if len(path) == len(prefix) {
				continue // Do not return the current directory.
			}
			if strings.LastIndex(path, "/") > len(prefix) {
				continue // Do not return files from sub directories.
			}

			if skip < f.dirOffset {
				skip++
				continue
			}
			infos = append(infos, fi)
			if len(infos) == count {
				break
			}
		}
	}
	f.dirOffset += len(infos)

	// Sort the infos by name and return.
	fis := &fileInfoSorter {
		infos : infos,
	}
	sort.Sort(fis)
	return infos, nil
}

func (f *memFile) Read(p []byte) (n int, err error) {
	if len(f.fi.content)-int(f.offset) >= len(p) {
		n = len(p)
	} else {
		n = len(f.fi.content) - int(f.offset)
		err = io.EOF
	}
	copy(p, f.fi.content[f.offset:f.offset+int64(n)])
	f.offset += int64(n)
	return
}

var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

func (f *memFile) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	default:
		return 0, errWhence
	case os.SEEK_SET:
	case os.SEEK_CUR:
		offset += f.offset
	case os.SEEK_END:
		offset += int64(len(f.fi.content))
	}
	if offset < 0 || int(offset) > len(f.fi.content) {
		return 0, errOffset
	}
	f.offset = offset
	return f.offset, nil
}
