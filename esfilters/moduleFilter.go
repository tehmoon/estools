package main

import (
	"./lib/esfilters"
	"github.com/tehmoon/errors"
	"flag"
	"text/tabwriter"
	"fmt"
	"os"
)

var (
	ErrModuleFilterCommandNotFound error
	ErrModuleFilterFlagMissing error
)

type FilterModule struct {
	filters *esfilters.QueryFilters
	command string
	options interface{}
	configured bool
}

type FilterModuleOptionsCommandResolve struct {
	Query string
}

type FilterModuleOptionsCommandDelete struct {
	Name string
}

type FilterModuleOptionsCommandAdd struct {
	Query string
	Name string
}

func (m *FilterModule) configureDelete(set *flag.FlagSet, rest []string) (error) {
	if m.configured {
		return ErrModuleAlreadyConfigured
	}

	options := &FilterModuleOptionsCommandDelete{}
	m.options = options

	set.StringVar(&options.Name, "name", "", "Filter to delete")

	set.Parse(rest)

	if options.Name == "" {
		return errors.Wrapf(ErrModuleFilterFlagMissing, "Flag -name is missing")
	}

	return nil
}

func (m *FilterModule) configureResolve(set *flag.FlagSet, rest []string) (error) {
	if m.configured {
		return ErrModuleAlreadyConfigured
	}

	options := &FilterModuleOptionsCommandResolve{}
	m.options = options

	set.StringVar(&options.Query, "query", "", "Query to add")

	set.Parse(rest)

	if options.Query == "" {
		return errors.Wrapf(ErrModuleFilterFlagMissing, "Flag -query is missing")
	}

	return nil
}

func (m *FilterModule) configureAdd(set *flag.FlagSet, rest []string) (error) {
	if m.configured {
		return ErrModuleAlreadyConfigured
	}

	options := &FilterModuleOptionsCommandAdd{}
	m.options = options

	set.StringVar(&options.Query, "query", "", "Query to add")
	set.StringVar(&options.Name, "name", "", "Name of the query")

	set.Parse(rest)

	if options.Query == "" {
		return errors.Wrapf(ErrModuleFilterFlagMissing, "Flag -query is missing")
	}

	if options.Name == "" {
		return errors.Wrapf(ErrModuleFilterFlagMissing, "Flag -name is missing")
	}

	return nil
}

func (m FilterModule) doDelete() (error) {
	options, ok := m.options.(*FilterModuleOptionsCommandDelete)
	if ! ok {
		return errors.New("Error type assertion")
	}

	return m.filters.Delete(options.Name)
}

func (m FilterModule) doList() (error) {
	filters := m.filters.List()

	writer := tabwriter.NewWriter(os.Stdout, 0, 1, 1, ' ', 0)

	fmt.Fprintln(writer, "Name\tQuery")
	fmt.Fprintln(writer, "\t")

	for _, filter := range filters {
		fmt.Fprintf(writer, "%s\t%s\n", filter.Name, filter.Query)
	}

	writer.Flush()

	return nil
}

func (m FilterModule) doResolve() (error) {
	options, ok := m.options.(*FilterModuleOptionsCommandResolve)
	if ! ok {
		return errors.New("Error type assertion")
	}

	query, err := m.filters.Resolve(options.Query)
	if err != nil {
		return err
	}

	fmt.Println(query)

	return nil
}

func (m FilterModule) doAdd() (error) {
	options, ok := m.options.(*FilterModuleOptionsCommandAdd)
	if ! ok {
		return errors.New("Error type assertion")
	}

	return m.filters.Add(options.Name, options.Query)
}

func (m *FilterModule) Configure(command string, rest []string) (error) {
	set := flag.NewFlagSet(fmt.Sprintf("%s filter", command), flag.ExitOnError)

	var err error
	switch command {
		case "add":
			err = m.configureAdd(set, rest)
		case "delete":
			err = m.configureDelete(set, rest)
		case "list":
		case "resolve":
			err = m.configureResolve(set, rest)
		default:
			return errors.Wrapf(ErrModuleFilterCommandNotFound, "command %s not found", command)
	}

	if err != nil {
		return err
	}

	m.configured = true
	m.command = command

	return nil
}

func (m FilterModule) Do() (error) {
	if ! m.configured {
		return ErrModuleNotConfigured
	}

	switch m.command {
		case "add":
			return m.doAdd()
		case "resolve":
			return m.doResolve()
		case "list":
			return m.doList()
		case "delete":
			return m.doDelete()
	}

	return nil
}

func NewModuleFilter(config *esfilters.Config) (*FilterModule) {
	return &FilterModule{
		filters: config.Filters,
	}
}
