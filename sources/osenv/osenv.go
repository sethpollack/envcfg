package osenv

import (
	"os"

	"github.com/sethpollack/envcfg/internal/loader"
	"github.com/sethpollack/envcfg/sources"
)

var _ loader.Source = (*source)(nil)

type source struct{}

func New() *source {
	return &source{}
}

func (s *source) Load() (map[string]string, error) {
	return sources.ToMap(os.Environ()), nil
}
