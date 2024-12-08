package dotenv

import (
	"os"
	"strings"

	"github.com/sethpollack/envcfg/internal/loader"
	"github.com/sethpollack/envcfg/sources"
)

var _ loader.Source = (*source)(nil)

type source struct {
	path string
}

func New(path string) *source {
	return &source{
		path: path,
	}
}

func (s *source) Load() (map[string]string, error) {
	bytes, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	return sources.ToMap(strings.Split(string(bytes), "\n")), nil
}
