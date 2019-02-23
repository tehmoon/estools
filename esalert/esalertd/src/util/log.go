package util

import (
	"log"
	"os"
	"runtime"
	"fmt"
	"strings"
	"path/filepath"
)

var (
	logger = log.New(os.Stderr, "", log.LstdFlags | log.Lmicroseconds)
	absPath string
)

func init() {
	_, fn, _, _ := runtime.Caller(0)
	fn = filepath.Clean(fn)
	dirs := strings.Split(fn, string(filepath.Separator))

	absPath = strings.Join(dirs[:len(dirs) - 2], string(filepath.Separator))
}

func Printf(format string, v ...interface{}) {
	fn, line := getCaller()
	format = fmt.Sprintf("%s:%d: %s", fn, line, format)

	logger.Printf(format, v...)
}

func Println(v ...interface{}) {
	fn, line := getCaller()
	format := fmt.Sprintf("%s:%d:", fn, line)

	args := append([]interface{}{format,}, v...)
	logger.Println(args...)
}

func Fatal(v ...interface{}) {
	fn, line := getCaller()
	format := fmt.Sprintf("%s:%d:", fn, line)

	args := append([]interface{}{format,}, v...)
	logger.Fatal(args...)
}

// Get the caller of the caller and output
// the relative path to absPath global
// Also call clean on it
func getCaller() (fn string, line int) {
	_, fn, line, _ = runtime.Caller(2)
	fn = filepath.Clean(fn)
	fn, _ = filepath.Rel(absPath, fn)

	return fn, line
}
