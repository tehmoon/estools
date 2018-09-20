package esfilters

import (
	"github.com/tehmoon/errors"
	"fmt"
	"strings"
)

type QueryFilter struct {
	Name string
	Query string
	ParsedQuery string
}

func parseQuery(qf map[string]*QueryFilter, q *QueryFilter) (map[string][]string, error) {
	if ok := QueryRegexpName.MatchString(q.Name); ! ok {
		return nil, errors.Errorf("Invalid name %s", q.Name)
	}

	dependsOn := make(map[string][]string)

	parsedQuery, err := resolveQuery(qf, q.Query, dependsOn)
	if err != nil {
		return nil, err
	}

	q.ParsedQuery = parsedQuery

	return dependsOn, nil
}

func resolveQuery(qf map[string]*QueryFilter, q string, dependsOn map[string][]string) (string, error) {
	query := fmt.Sprintf("(%s)", q)
	parsedQuery := QueryRegexp.ReplaceAllStringFunc(query, QueryRegexpFunc(qf, dependsOn))

	err := parsedQueryHasErrors(parsedQuery)
	if err != nil {
		return "", err
	}

	return parsedQuery, nil
}

var (
	ErrQueryFilterValueNotFound error
	ErrQueryFilterTypeNotFound error
)

func parsedQueryHasErrors(query string) (error) {
	matchedErrors := QueryRegexpError.FindAllStringSubmatch(query, -1)
	if len(matchedErrors) != 0 {
		split := strings.Split(matchedErrors[0][1], ":")

		partErr := split[1]
		partTarget := split[2]

		switch partErr {
			case "filter_not_found":
				return errors.Wrapf(ErrQueryFilterValueNotFound, "Filter value %s has not been found", partTarget)
			case "type_not_found":
				return errors.Wrapf(ErrQueryFilterTypeNotFound, "Filter type %s is not yet implemented", partTarget)
		}

		return errors.Errorf("unknown error %s: %s", partErr, partTarget)
	}

	return nil
}
