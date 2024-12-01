package mapenv

import "github.com/sethpollack/envcfg/internal/loader"

var _ loader.Source = (*source)(nil)

type source struct {
	env map[string]string
}

func New(env map[string]string) *source {
	return &source{
		env: env,
	}
}

func (s *source) Load() (map[string]string, error) {
	return s.env, nil
}
