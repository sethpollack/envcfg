package loader

import (
	"fmt"

	errs "github.com/sethpollack/envcfg/errors"
)

type Source interface {
	Load() (map[string]string, error)
}

type Loader struct {
	Sources    []Source
	Filters    []func(string) bool
	Transforms []func(string) string
}

func (l *Loader) Load() (map[string]string, error) {
	envs := make(map[string]string)

	for _, s := range l.Sources {
		loaded, err := s.Load()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrLoadEnv, err)
		}

		for k, v := range loaded {
			if l.matches(k) {
				k = l.transform(k)
				envs[k] = v
			}
		}
	}

	return envs, nil
}

func (l *Loader) matches(key string) bool {
	if len(l.Filters) == 0 {
		return true
	}

	for _, f := range l.Filters {
		if f(key) {
			return true
		}
	}

	return false
}

func (l *Loader) transform(key string) string {
	for _, t := range l.Transforms {
		key = t(key)
	}

	return key
}
