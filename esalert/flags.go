package main

import (
	"strings"
	"flag"
	"os"
	"fmt"
	"github.com/tehmoon/errors"
)

type Flags struct {
	Server string
	Index string
	Dir string
	SleepFor int
	Exec string
	Owners []string
}

func parseFlags() (*Flags) {
	flags := &Flags{
		Owners: make([]string, 0),
	}

	var owners string

	flag.StringVar(&flags.Server, "server", "http://localhost:9200", "Specify elasticsearch server to query")
	flag.StringVar(&flags.Index, "index", "", "Specify the elasticsearch index to query")
	flag.StringVar(&flags.Dir, "dir", "", "Directory where the .json files are")
	flag.StringVar(&flags.Exec, "exec", "", "Execute a command when alerting")
	flag.IntVar(&flags.SleepFor, "sleep-for", 60, "Sleep for in seconds after all queries have been ran")
	flag.StringVar(&owners, "owners", "", "List of default owners separated by \",\" to notify")

	flag.Parse()

	if flags.Index == "" {
		fmt.Fprintf(os.Stderr, "Flag %q is required\n", "-index")
		flag.Usage()
		os.Exit(2)
	}

	if flags.Dir == "" {
		fmt.Fprintf(os.Stderr, "Flag %q is required\n", "-dir")
		flag.Usage()
		os.Exit(2)
	}

	if flags.SleepFor < 0 {
		fmt.Fprintf(os.Stderr, "Flag %q cannot be lower than 1", "sleep for")
		flag.Usage()
		os.Exit(2)
	}

	if owners != "" {
		flags.Owners = strings.Split(owners, ",")
	}

	err := isDir(flags.Dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Fail to asert flag %q", "-dir").Error())
		flag.Usage()
		os.Exit(2)
	}

	return flags
}

func init() {
	flag.Usage = func () {
		fmt.Fprintf(os.Stderr, "Usage of %s: <-server=Url> <-index=Index> <-dir=Directory>\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func isDir(p string) (error) {
	fm, err := os.Stat(p)
	if err != nil {
		return err
	}

	if ! fm.IsDir() {
		return errors.Errorf("Path %q is not a directory", p)
	}

	return nil
}
