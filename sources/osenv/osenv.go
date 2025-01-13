package osenv

import (
	"os"

	"github.com/sethpollack/envcfg/internal/loader"
	"github.com/sethpollack/envcfg/sources"
)

var _ loader.Source = (*source)(nil)

type source struct {
	keys []string
}

func New(keys ...string) *source {
	return &source{keys: keys}
}

func (s *source) Load() (map[string]string, error) {
	if len(s.keys) == 0 {
		return sources.ToMap(os.Environ()), nil
	}

	envs := make(map[string]string)

	for _, key := range s.keys {
		envs[key] = os.Getenv(key)
	}

	return envs, nil
}
