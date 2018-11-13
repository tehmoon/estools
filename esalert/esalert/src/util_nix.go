// +build !windows

package main

import (
	"os"
	"syscall"
	"github.com/tehmoon/errors"
)

func validateBin(bin string) (error) {
	stat, err := os.Lstat(bin)
	if err != nil {
		return errors.Wrap(err, "Error calling lstat()")
	}

	stat_t, ok := stat.Sys().(*syscall.Stat_t)
	if ! ok {
		return errors.Errorf("Error assessing Sys() interface of type %t\n", stat.Sys())
	}

	if stat_t.Uid != 0 {
		return errors.Errorf("Error validating uid of owner, should have 0 got %d\n", stat_t.Uid)
	}

	if stat_t.Gid != 0 && (stat_t.Mode & 0040 != 0040) {
		return errors.Errorf("Error validating gid of owner, if gid is not 0, write should not be permited")
	}

	return nil
}
