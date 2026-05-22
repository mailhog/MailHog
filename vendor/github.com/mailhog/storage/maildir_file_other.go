//go:build !linux
// +build !linux

package storage

import "os"

type cachedMaildirDataFile struct {
	*os.File
}

func (file *cachedMaildirDataFile) Sync() error {
	if err := file.File.Sync(); err != nil {
		return err
	}
	dropFileCache(file.File)
	return nil
}

func (file *cachedMaildirDataFile) WriteString(value string) (int, error) {
	return file.File.WriteString(value)
}

func openMaildirDataFile(path string) (maildirDataFile, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return nil, err
	}
	return &cachedMaildirDataFile{File: file}, nil
}
