package swarm

import (
	"net/url"
	"strings"
)

func ParseLabels(q string) (url.Values, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, nil
	}

	return url.ParseQuery(q)
}

func HasString(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
