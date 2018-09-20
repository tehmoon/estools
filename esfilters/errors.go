package main

import (
	"github.com/tehmoon/errors"
	"os"
	"syscall"
)

func ErrAssertSyscallErrno(err error, i uintptr) (bool) {
	if e, ok := err.(*errors.Error); ok {
		if e, ok := e.Root().(*os.PathError); ok {
			if e, ok := e.Err.(syscall.Errno); ok {
				if uintptr(e) == i {
					return true
				}
			}
		}
	}

	return false
}
