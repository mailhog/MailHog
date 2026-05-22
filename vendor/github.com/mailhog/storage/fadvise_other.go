//go:build !linux
// +build !linux

package storage

import "os"

func dropFileCache(file *os.File) {
}
