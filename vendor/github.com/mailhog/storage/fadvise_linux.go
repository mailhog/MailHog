//go:build linux
// +build linux

package storage

import (
	"os"
	"syscall"
)

const posixFadvDontNeed = 4

func dropFileCache(file *os.File) {
	syscall.Syscall6(syscall.SYS_FADVISE64, file.Fd(), 0, 0, posixFadvDontNeed, 0, 0)
}
