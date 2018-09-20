package esfilters

import (
	"github.com/tehmoon/errors"
	"encoding/json"
	"strings"
	"sync"
)

type QueryFilters struct {
	sync.RWMutex
	filters map[string]*QueryFilter
	dependencies map[string][]string
	exports map[string]string
}

func (qf QueryFilters) Resolve(query string) (string, error) {
	qf.Lock()
	defer qf.Unlock()

	return resolveQuery(qf.filters, query, nil)
}

func (qf QueryFilters) List() ([]*QueryFilter) {
	qf.Lock()
	defer qf.Unlock()

	filters := make([]*QueryFilter, 0)

	for _, f := range qf.filters {
		filter := &QueryFilter{}
		*filter = *f

		filters = append(filters, filter)
	}

	return filters
}

func (qf QueryFilters) Add(name, query string) (error) {
	qf.RLock()
	defer qf.RUnlock()

	if _, found := qf.filters[name]; found {
		return errors.Errorf("Query %s is already declared", name)
	}

	q := &QueryFilter{
		Name: name,
		Query: query,
	}

	dependsOn, err := parseQuery(qf.filters, q)
	if err != nil {
		return errors.Wrapf(err, "Error parsing query %s", q.Name)
	}

	if values, found := dependsOn["filter"]; found {
		for _, value := range values {
			dependencies, found := qf.dependencies[value]
			if ! found {
				dependencies = make([]string, 0)
			}

			found = false
			for _, dependency := range dependencies {
				if dependency == name {
					found = true
					break
				}
			}

			if ! found {
				qf.dependencies[value] = append(dependencies, name)
			}
		}
	}

	qf.filters[name] = q
	qf.exports[name] = query

	return nil
}

func (qf QueryFilters) Get(name string) (string, bool) {
	qf.Lock()
	defer qf.Unlock()

	if f, found := qf.filters[name]; found {
		return f.ParsedQuery, true
	}

	return "", false
}

func (qf QueryFilters) ExportConfig() ([]byte, error) {
	qf.RLock()
	defer qf.RUnlock()

	payload, err := json.MarshalIndent(qf.exports, "", "	")
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling filters to JSON")
	}

	return payload, nil
}

func (qf QueryFilters) Delete(name string) (error) {
	qf.RLock()
	defer qf.RUnlock()

	if _, found := qf.filters[name]; found {
		dependencies, _ := qf.dependencies[name]

		if dependencies != nil {
			return errors.Errorf("Filter %s has dependencies: %s", name, strings.Join(dependencies, ", "))
		}

		delete(qf.filters, name)
		delete(qf.exports, name)
		return nil
	}

	return errors.Errorf("Filter %s has not been found", name)
}

func (qf *QueryFilters) ImportConfig(payload []byte) (error) {
	qf.RLock()
	defer qf.RUnlock()

	exports := make(map[string]string)
	queryFilters := NewQueryFilters()

	err := json.Unmarshal(payload, &exports)
	if err != nil {
		return errors.Wrap(err, "Error unmarshaling filters from JSON")
	}

	var lastError error

	for {
		length := len(exports)
		if length == 0 {
			break
		}

		for name, query := range exports {
			err := queryFilters.Add(name, query)
			if err != nil {
				if err, ok := err.(*errors.Error); ok {
					if err.Root() == ErrQueryFilterValueNotFound {
						lastError = err
						continue
					}
				}

				return errors.Wrapf(err, "Error processing filter %s", name)
			}

			delete(exports, name)
		}

		if l := len(exports); l == length {
			return errors.Wrap(lastError, "No filter left to add due to a recursive error")
		}
	}

	qf.filters = queryFilters.filters
	qf.exports = queryFilters.exports
	qf.dependencies = queryFilters.dependencies

	return nil
}

func NewQueryFilters() (*QueryFilters) {
	return &QueryFilters{
		filters: make(map[string]*QueryFilter),
		exports: make(map[string]string),
		dependencies: make(map[string][]string),
	}
}
