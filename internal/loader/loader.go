package loader

import (
	"os"
	"regexp"
	"strings"
)

type Source interface {
	Load() map[string]string
	Build(opts ...any) error
}

type Option func(*env)

// WithSource adds a custom source to load environment variables from.
// The source must implement the Source interface with Load() and Build() methods.
// Multiple sources can be added and will be loaded in the order they were added.
func WithSource(source Source) Option {
	return func(e *env) {
		e.sources = append(e.sources, source)
	}
}

// WithEnvVarsSource adds a source that loads environment variables from the provided map.
// This is useful for testing or providing a fixed set of environment variables.
// The provided map will be used as-is without any modifications.
func WithEnvVarsSource(envs map[string]string) Option {
	return func(e *env) {
		e.sources = append(e.sources, &envSource{envs})
	}
}

// WithOsEnvSource adds a source that loads environment variables from the operating system environment.
// This source uses os.Environ() to get the current process environment variables.
// This is the default source if no other sources are specified.
func WithOsEnvSource() Option {
	return func(e *env) {
		e.sources = append(e.sources, &osSource{})
	}
}

// WithFileSource adds a source that loads environment variables from a file at the given path.
// The file should contain environment variables in KEY=VALUE format, one per line.
// If the file cannot be read, the source will return an empty map.
func WithFileSource(path string) Option {
	return func(e *env) {
		e.sources = append(e.sources, &fileSource{path})
	}
}

// WithDefaults adds default values for environment variables.
// The provided map will be used as fallback values when a key is not found in any source.
// If a key exists in both a source and defaults, the source value takes precedence.
func WithDefaults(envs map[string]string) Option {
	return func(e *env) {
		e.defaults = envs
	}
}

// WithFilter adds a filter function that determines which environment variables to include.
// The filter function takes a key string and returns true if the key should be included,
// false if it should be excluded. Multiple filters can be added and a key will be included
// if any filter returns true for it. If no filters are added, all keys are included.
func WithFilter(filter func(string) bool) Option {
	return func(e *env) {
		e.filters = append(e.filters, filter)
	}
}

// WithTransform adds a transform function that modifies environment variable keys.
// The transform function takes a key string and returns a new key string.
// Multiple transforms can be added and they are applied in the order they were added.
func WithTransform(transform func(string) string) Option {
	return func(e *env) {
		e.transforms = append(e.transforms, transform)
	}
}

// WithHasPrefix adds a filter that only includes environment variables whose keys start with the given prefix.
// For example, WithHasPrefix("APP_") would include "APP_NAME" but exclude "NAME".
// Multiple filters can be added and a key will be included if any filter matches.
func WithHasPrefix(prefix string) Option {
	return func(e *env) {
		e.filters = append(e.filters, func(key string) bool {
			return strings.HasPrefix(key, prefix)
		})
	}
}

// WithHasSuffix adds a filter that only includes environment variables whose keys end with the given suffix.
// For example, WithHasSuffix("_TEST") would include "APP_TEST" but exclude "APP_TEST_2".
// Multiple filters can be added and a key will be included if any filter matches.
func WithHasSuffix(suffix string) Option {
	return func(e *env) {
		e.filters = append(e.filters, func(key string) bool {
			return strings.HasSuffix(key, suffix)
		})
	}
}

// WithHasMatch adds a filter that only includes environment variables whose keys match the given regular expression pattern.
// For example, WithHasMatch(`TEST_\d+`) would include "TEST_123" but exclude "TEST_ABC".
// Multiple filters can be added and a key will be included if any filter matches.
// The pattern must be a valid regular expression - if invalid, it will panic.
func WithHasMatch(pattern string) Option {
	regex := regexp.MustCompile(pattern)
	return func(e *env) {
		e.filters = append(e.filters, func(key string) bool {
			return regex.MatchString(key)
		})
	}
}

// WithTrimPrefix adds a transform that removes the given prefix from environment variable keys.
// For example, WithTrimPrefix("APP_") would transform "APP_NAME" to "NAME".
// Multiple transforms can be added and they are applied in the order they were added.
// If the key does not start with the prefix, it is returned unchanged.
func WithTrimPrefix(prefix string) Option {
	return func(e *env) {
		e.transforms = append(e.transforms, func(key string) string {
			return strings.TrimPrefix(key, prefix)
		})
	}
}

// WithTrimSuffix adds a transform that removes the given suffix from environment variable keys.
// For example, WithTrimSuffix("_TEST") would transform "APP_TEST" to "APP".
// Multiple transforms can be added and they are applied in the order they were added.
// If the key does not end with the suffix, it is returned unchanged.
func WithTrimSuffix(suffix string) Option {
	return func(e *env) {
		e.transforms = append(e.transforms, func(key string) string {
			return strings.TrimSuffix(key, suffix)
		})
	}
}

// WithPrefix adds a filter that only includes environment variables whose keys start with the given prefix,
// and a transform that removes that prefix from the keys.
// For example, WithPrefix("APP_") would transform "APP_NAME" to "NAME" and exclude "OTHER_NAME".
// Multiple filters and transforms can be added and they are applied in the order they were added.
func WithPrefix(prefix string) Option {
	return func(e *env) {
		e.filters = append(e.filters, func(key string) bool {
			return strings.HasPrefix(key, prefix)
		})
		e.transforms = append(e.transforms, func(key string) string {
			return strings.TrimPrefix(key, prefix)
		})
	}
}

// WithSuffix adds a filter that only includes environment variables whose keys end with the given suffix,
// and a transform that removes that suffix from the keys.
// For example, WithSuffix("_TEST") would transform "APP_TEST" to "APP" and exclude "APP_TEST_2".
// Multiple filters and transforms can be added and they are applied in the order they were added.
func WithSuffix(suffix string) Option {
	return func(e *env) {
		e.filters = append(e.filters, func(key string) bool {
			return strings.HasSuffix(key, suffix)
		})
		e.transforms = append(e.transforms, func(key string) string {
			return strings.TrimSuffix(key, suffix)
		})
	}
}

type env struct {
	sources    []Source
	defaults   map[string]string
	filters    []func(string) bool
	transforms []func(string) string
}

func New() *env {
	return &env{}
}

func (e *env) Build(opts ...any) error {
	for _, opt := range opts {
		if v, ok := opt.(Option); ok {
			v(e)
		}
	}

	if len(e.sources) == 0 {
		e.sources = append(e.sources, &osSource{})
	}

	for _, s := range e.sources {
		if err := s.Build(opts...); err != nil {
			return err
		}
	}

	return nil
}

func (e *env) Load() map[string]string {
	envs := make(map[string]string)

	for k, v := range e.defaults {
		envs[k] = v
	}

	for _, s := range e.sources {
		for k, v := range s.Load() {
			if e.matches(k) {
				k = e.transform(k)
				envs[k] = v
			}
		}
	}

	return envs
}

func (e *env) matches(key string) bool {
	if len(e.filters) == 0 {
		return true
	}

	for _, f := range e.filters {
		if f(key) {
			return true
		}
	}

	return false
}

func (e *env) transform(key string) string {
	for _, t := range e.transforms {
		key = t(key)
	}

	return key
}

type osSource struct{}

func (o *osSource) Load() map[string]string {
	return toMap(os.Environ())
}

func (o *osSource) Build(opts ...any) error {
	return nil
}

type envSource struct {
	envs map[string]string
}

func (e *envSource) Load() map[string]string {
	return e.envs
}

func (e *envSource) Build(opts ...any) error {
	return nil
}

type fileSource struct {
	path string
}

func (f *fileSource) Load() map[string]string {
	bytes, err := os.ReadFile(f.path)
	if err != nil {
		return nil
	}

	return toMap(strings.Split(string(bytes), "\n"))
}

func (f *fileSource) Build(opts ...any) error {
	return nil
}

func toMap(env []string) map[string]string {
	m := make(map[string]string)
	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			m[pair[0]] = pair[1]
		}
	}

	return m
}
