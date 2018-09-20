package esfilters

import (
	"regexp"
	"strings"
	"fmt"
)

var (
	QueryRegexp = regexp.MustCompile(`%{([a-zA-Z0-9_-]+:[a-zA-Z0-9_-]+)}`)
	QueryRegexpError = regexp.MustCompile(`%{(error:[a-zA-Z0-9_-]+:[a-zA-Z0-9_-]+)}`)
	QueryRegexpName = regexp.MustCompilePOSIX(`^[[:alnum:]]+([_-]?[[:alnum:]])*$`)
)

func QueryRegexpFunc(qf map[string]*QueryFilter, dependsOn map[string][]string) (func (string) (string)) {
	return func(str string) (string) {
		part := QueryRegexp.FindStringSubmatch(str)
		split := strings.Split(part[1], ":")

		partType := split[0]
		partValue := split[1]

		if dependsOn != nil {
			values, found := dependsOn[partType];
			if ! found {
				if values == nil {
					values = make([]string, 0)
				}
			}

			found = false

			for _, value := range values {
				if value == partValue {
					found = true
					break
				}
			}

			if ! found {
				values = append(values, partValue)
			}

			dependsOn[partType] = values
		}

		switch partType {
			case "filter":
				if f, found := qf[partValue]; found {
					return f.ParsedQuery
				}

				return fmt.Sprintf(`%%{error:filter_not_found:%s}`, partValue)
			case "placeholder":
		}

		return fmt.Sprintf(`%%{error:type_not_found:%s}`, partType)
	}
}
