package main

import (
	"github.com/spf13/pflag"
	"os"
	"fmt"
	"github.com/tehmoon/errors"
	"path/filepath"
	"net/url"
)

type Flags struct {
	Url *url.URL
	Tags []string
	ActionsDir string
}

func parseFlags() (*Flags) {
	var (
		uri string
		flags = &Flags{}
		err error
	)

	pflag.StringVar(&uri, "url", "", "Url of server")
	pflag.StringArrayVar(&flags.Tags, "tags", make([]string, 0), "Tags associated with this client")
	pflag.StringVar(&flags.ActionsDir, "dir", "", "Path to actions scripts' directory")

	pflag.Parse()

	if flags.ActionsDir == "" {
		fmt.Fprintf(os.Stderr, "Flag %q is required\n", "-dir")
		pflag.Usage()
		os.Exit(2)
	}

	if uri == "" {
		fmt.Fprintf(os.Stderr, "Flag %q is required\n", "-url")
		pflag.Usage()
		os.Exit(2)
	}

	flags.Url, err = url.Parse(uri)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Fail to asset flag %q", "-url").Error())
		pflag.Usage()
		os.Exit(2)
	}

	if flags.Url.Scheme != "https" && flags.Url.Scheme != "http" {
		fmt.Fprintf(os.Stderr, "Scheme %q not supported. One of %q or %q\n", flags.Url.Scheme, "https", "http")
		os.Exit(2)
	}

	if len(flags.Tags) == 0 {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Flag %q needs to have at least one tag", "-tags").Error())
		pflag.Usage()
		os.Exit(2)
	}

	for _, tag := range flags.Tags {
		if tag == "" {
			fmt.Fprintln(os.Stderr, "One flag is empty")
			pflag.Usage()
			os.Exit(2)
		}
	}

	err = isDir(flags.ActionsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Fail to assert flag %q", "-dir").Error())
		pflag.Usage()
		os.Exit(2)
	}

	flags.ActionsDir, err = filepath.Abs(flags.ActionsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Fail to assert flag %q", "-dir").Error())
		pflag.Usage()
		os.Exit(2)
	}

	return flags
}

func init() {
	pflag.Usage = func () {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
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
