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

package memfs

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
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

type memFile struct {
	info      *memFileInfo
	offset    int64
	dirOffset int
}

func (f *memFile) Close() error {
	return nil
}

func (f *memFile) Stat() (os.FileInfo, error) {
	return f.info, nil
}

func (f *memFile) Readdir(count int) ([]os.FileInfo, error) {
	infos := []os.FileInfo{}
	prefix := f.info.path
	skip := 0
	for path, info := range f.info.fs.cache {
		if strings.HasPrefix(path, prefix) && path != prefix {
			if (skip < f.dirOffset) {
				skip++
				continue
			}
			infos = append(infos, info)
			if len(infos) == count {
				break
			}
		}
	}
	f.dirOffset += len(infos)
	return infos, nil
}

func (f *memFile) Read(p []byte) (n int, err error) {
	if len(f.info.content)-int(f.offset) > len(p) {
		n = len(p)
	} else {
		n = len(f.info.content) - int(f.offset)
	}
	p = f.info.content[f.offset:n]
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
		offset += int64(len(f.info.content))
	}
	if offset < 0 || int(offset) > len(f.info.content) {
		return 0, errOffset
	}
	f.offset = offset
	return f.offset, nil
}

type memFileSystem struct {
	root  string
	cache map[string]*memFileInfo
	lock  *sync.RWMutex
}

func (fs *memFileSystem) Open(name string) (http.File, error) {
	name = filepath.Join(fs.root, name)
	name = path.Clean(name)
	fs.lock.RLock()
	fi, ok := fs.cache[name]
	fs.lock.RUnlock()
	if !ok {
		return nil, errors.New("file/dir not found")
	}
	return &memFile{
		info:   fi,
	}, nil
}

func (fs *memFileSystem) refreshCache(path string, info os.FileInfo) (err error) {
	// Create memory fileinfo and read contents.
	fi := &memFileInfo{
		name:    info.Name(),
		size:    info.Size(),
		mode:    info.Mode(),
		modTime: info.ModTime(),
		isDir:   info.IsDir(),
		path:    path,
		fs:      fs,
	}

	// Fill content of the file from disk.
	if !fi.isDir {
		fi.content, err = ioutil.ReadFile(path)
		if err != nil {
			return
		}
	}

	// Update cache and return file.
	fs.lock.Lock()
	fs.cache[path] = fi
	fs.lock.Unlock()
	return
}

func New(root string) (http.FileSystem, error) {
	root = path.Clean(root)
	memFS := &memFileSystem{
		root:  root,
		cache: map[string]*memFileInfo{},
		lock:  &sync.RWMutex{},
	}

	// Walk the whole directory and cache everything.
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		err = memFS.refreshCache(path, info)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return memFS, nil
}
