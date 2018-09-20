package main

import (
	"./lib/esfilters"
	"github.com/tehmoon/errors"
)

type Module interface {
	Configure(action string, rest []string) (error)
	Do() (error)
}

var (
	ErrModuleNotFound error = errors.New("module not found")
	ErrModuleNotConfigured error = errors.New("module not configured, call Configure() first.")
	ErrModuleAlreadyConfigured error = errors.New("module is already configured, call Do() insteadt.")
)

func parseModule(module string, config *esfilters.Config) (Module, error) {
	var m Module

	switch module {
		case "filter":
			filter := NewModuleFilter(config)
			m = filter
		default:
			return nil, ErrModuleNotFound
	}

	return m, nil
}
