// Copyright (c) 2013 The Go Authors. All rights reserved.
// Copyright (c) 2013 memfs Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package memfs creates a watched in memory filesystem.
package memfs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/howeyc/fsnotify"
)

var (
	// Silent is a logger that throws away the log messages to /dev/null.
	Silent = log.New(ioutil.Discard, "memfs: ", log.Lshortfile|log.Ldate|log.Ltime)

	// Verbose is a logger that prints the log messages to Stderr.
	Verbose = log.New(os.Stderr, "memfs: ", log.Lshortfile|log.Ldate|log.Ltime)

	logger *log.Logger
)

// Set logger sets the logger to be used by this package.
func SetLogger(l *log.Logger) {
	logger = l
}

func init() {
	SetLogger(Silent)
}

type memFileSystem struct {
	root    string
	cache   map[string]*memFileInfo
	lock    *sync.RWMutex
	watcher *fsnotify.Watcher
}

func (fs *memFileSystem) Open(name string) (http.File, error) {
	name = filepath.Join(fs.root, name)

	fs.lock.RLock()
	fi, ok := fs.cache[name]
	fs.lock.RUnlock()
	if !ok {
		return nil, errors.New("file/dir not found")
	}
	return &memFile{
		fi: fi,
	}, nil
}

func (fs *memFileSystem) refreshCache(path string, info os.FileInfo) (err error) {
	// Delete the file if fi is nil.
	if info == nil {
		fs.lock.Lock()
		delete(fs.cache, path)
		fs.lock.Unlock()
		return
	}

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

	// Update cache and return.
	fs.lock.Lock()
	fs.cache[path] = fi
	fs.lock.Unlock()
	return
}

func (fs *memFileSystem) walk() error {
	// Walk the whole directory and cache everything.
	err := filepath.Walk(fs.root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk: %s with err: %v", fs.root, err)
		}
		err = fs.refreshCache(path, fi)
		if err != nil {
			return fmt.Errorf("failed to walk: %s with err: %v", fs.root, err)
		}
		if fi.IsDir() {
			err = fs.watcher.Watch(path)
			if err != nil {
				return fmt.Errorf("failed to add watch: %s err: %v", path, err)
			}
		}
		return nil
	})
	return err
}

func (fs *memFileSystem) reloadFile(name string) os.FileInfo {
	fi, err := os.Stat(name)
	if err != nil {
		logger.Printf("failed to stat: %s with err: %v", name, err)
		return nil
	}
	err = fs.refreshCache(name, fi)
	if err != nil {
		logger.Printf("failed to reload: %s with err: %v", name, err)
		return nil
	}
	return fi
}

func (fs *memFileSystem) deleteFile(name string) os.FileInfo {
	fs.lock.RLock()
	fi, ok := fs.cache[name]
	fs.lock.RUnlock()
	if !ok {
		return nil
	}
	err := fs.refreshCache(name, nil)
	if err != nil {
		logger.Printf("failed to delete: %s with err: %v", name, err)
	}
	return fi
}

func (fs *memFileSystem) watcherCallback() {
	for {
		select {
		case e := <-fs.watcher.Event:
			if e.IsCreate() {
				fi := fs.reloadFile(e.Name)
				if fi != nil && fi.IsDir() {
					err := fs.watcher.Watch(e.Name)
					if err != nil {
						logger.Printf("failed to add watch: %s err: %v", e.Name, err)
					}
				}
				fs.reloadFile(path.Dir(e.Name))
			}
			if e.IsModify() {
				fs.reloadFile(e.Name)
			}
			if e.IsDelete() || e.IsRename() {
				fi := fs.deleteFile(e.Name)
				if fi != nil && fi.IsDir() {
					err := fs.watcher.RemoveWatch(e.Name)
					if err != nil {
						logger.Printf("failed to remove watch: %s err: %v", e.Name, err)
					}
				}
				fs.reloadFile(path.Dir(e.Name))
			}
		case err := <-fs.watcher.Error:
			logger.Printf("watcher error: %v", err)
		}
	}
}

// New creates a new in memory filesystem at root.
func New(root string) (http.FileSystem, error) {
	root = path.Clean(root)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %v", err)
	}

	memFS := &memFileSystem{
		root:    root,
		cache:   map[string]*memFileInfo{},
		lock:    &sync.RWMutex{},
		watcher: watcher,
	}

	// Set watcher callback.
	go memFS.watcherCallback()

	// Cache all the files and directory.
	err = memFS.walk()
	if err != nil {
		return nil, err
	}

	return memFS, nil
}
