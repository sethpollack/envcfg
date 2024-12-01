package sources

import (
	"strings"
)

func ToMap(env []string) map[string]string {
	m := make(map[string]string)
	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			m[pair[0]] = pair[1]
		}
	}

	return m
}
