//go:build linux
// +build linux

package storage

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

const directIOAlignment = 4096
const directIOBufferSize = 1024 * 1024

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

type directMaildirDataFile struct {
	file    *os.File
	rawBuf  []byte
	buf     []byte
	offset  int
	logical int64
}

func openMaildirDataFile(path string) (maildirDataFile, error) {
	fd, err := syscall.Open(path, syscall.O_CREAT|syscall.O_TRUNC|syscall.O_WRONLY|syscall.O_DIRECT, 0660)
	if err == nil {
		return newDirectMaildirDataFile(os.NewFile(uintptr(fd), path)), nil
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return nil, err
	}
	return &cachedMaildirDataFile{File: file}, nil
}

func newDirectMaildirDataFile(file *os.File) *directMaildirDataFile {
	raw := make([]byte, directIOBufferSize+directIOAlignment)
	start := uintptr(unsafe.Pointer(&raw[0]))
	shift := int((directIOAlignment - (start % directIOAlignment)) % directIOAlignment)
	return &directMaildirDataFile{
		file:   file,
		rawBuf: raw,
		buf:    raw[shift : shift+directIOBufferSize],
	}
}

func (file *directMaildirDataFile) Write(p []byte) (int, error) {
	total := len(p)
	for len(p) > 0 {
		n := copy(file.buf[file.offset:], p)
		file.offset += n
		file.logical += int64(n)
		p = p[n:]

		if file.offset == len(file.buf) {
			if err := file.writeAligned(file.buf); err != nil {
				return total - len(p), err
			}
			file.offset = 0
		}
	}
	return total, nil
}

func (file *directMaildirDataFile) WriteString(value string) (int, error) {
	total := len(value)
	for len(value) > 0 {
		n := copy(file.buf[file.offset:], value)
		file.offset += n
		file.logical += int64(n)
		value = value[n:]

		if file.offset == len(file.buf) {
			if err := file.writeAligned(file.buf); err != nil {
				return total - len(value), err
			}
			file.offset = 0
		}
	}
	return total, nil
}

func (file *directMaildirDataFile) Sync() error {
	return file.file.Sync()
}

func (file *directMaildirDataFile) Close() error {
	if file.offset > 0 {
		for i := file.offset; i < len(file.buf); i++ {
			file.buf[i] = 0
		}
		if err := file.writeAligned(file.buf); err != nil {
			file.file.Close()
			return err
		}
	}

	if err := file.file.Truncate(file.logical); err != nil {
		file.file.Close()
		return err
	}
	if _, err := file.file.Seek(file.logical, io.SeekStart); err != nil {
		file.file.Close()
		return err
	}
	if err := file.file.Sync(); err != nil {
		file.file.Close()
		return err
	}
	return file.file.Close()
}

func (file *directMaildirDataFile) writeAligned(data []byte) error {
	written := 0
	for written < len(data) {
		n, err := file.file.Write(data[written:])
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}
