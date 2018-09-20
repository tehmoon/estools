package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/tehmoon/errors"
)

var (
	ErrFlagsMissing error
	ErrFlagsModuleMissing error = errors.New("Module is missing")
	ErrFlagsCommandMissing error = errors.New("Command is missing")
)

type Flags struct {
	ConfigFile string
	Command string
	Module string
	Rest []string
}

func init() {
	flag.Usage = usage
}

func parseFlags() (*Flags, error) {
	flags := &Flags{}

	flag.StringVar(&flags.ConfigFile, "c", "", "Config file to use")

	flag.Parse()

	if flags.ConfigFile == "" {
		return nil, errors.Wrapf(ErrFlagsMissing, "Flag -c is missing")
	}

	args := flag.Args()
	switch len(args){
		case 0:
			return nil, ErrFlagsModuleMissing
		case 1:
			return nil, ErrFlagsCommandMissing
	}

	flags.Command = args[0]
	flags.Module = args[1]
	flags.Rest = args[2:]

	return flags, nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <Options> <Command> <Module> [-h | --help]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()

	fmt.Fprintln(os.Stderr, "")

	fmt.Fprintf(os.Stderr, "Command:\n")
	for _, command := range []string{"resolve", "add", "delete", "list"} {
		fmt.Fprintf(os.Stderr, "  %s\n", command)
	}

	fmt.Fprintln(os.Stderr, "")

	fmt.Fprintf(os.Stderr, "Module:\n")
	for _, module := range []string{"filter"} {
		fmt.Fprintf(os.Stderr, "  %s\n", module)
	}
}
