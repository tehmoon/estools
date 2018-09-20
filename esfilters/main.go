package main

import (
	"github.com/tehmoon/errors"
	"./lib/esfilters"
	"os"
	"fmt"
)

func main() {
	flags, err := parseFlags()
	if err != nil {
		usage()
		os.Exit(2)
	}

	config, err := esfilters.ImportConfigFromFile(flags.ConfigFile)
	if err != nil {
		if ok := ErrAssertSyscallErrno(err, 0x02); ok {
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(2)
		}

		config = esfilters.NewConfig()
	}

	module, err := parseModule(flags.Module, config)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing module %s", flags.Module).Error())
		os.Exit(2)
	}

	err = module.Configure(flags.Command, flags.Rest)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error configuring module %s", flags.Module).Error())
		os.Exit(2)
	}

	err = module.Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, errors.Wrapf(err, "Error executing command %s %s", flags.Command, flags.Module).Error())
		os.Exit(2)
	}

	err = config.ExportConfigToFile(flags.ConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, errors.Wrapf(err, "Error saving config file").Error())
		os.Exit(2)
	}
}
