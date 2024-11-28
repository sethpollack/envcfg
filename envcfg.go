package envcfg

import (
	"reflect"

	"github.com/sethpollack/envcfg/internal/decoder"
	"github.com/sethpollack/envcfg/internal/loader"
	"github.com/sethpollack/envcfg/internal/matcher"
	"github.com/sethpollack/envcfg/internal/parser"
	"github.com/sethpollack/envcfg/internal/walker"
)

// WithTagName sets a custom struct tag name to override the default "env" tag.
func WithTagName(tag string) any {
	return walker.WithTagName(tag)
}

// WithDelimiterTag sets the struct tag name used for the delimiter.
// The default tag name is "delim".
func WithDelimiterTag(tag string) any {
	return walker.WithDelimiterTag(tag)
}

// WithDelimiter sets the delimiter used to separate slice/map elements
// in environment variable values. The default delimiter is ",".
func WithDelimiter(delim string) any {
	return walker.WithDelimiter(delim)
}

// WithSeparatorTag sets the struct tag name used for the separator.
// The default tag name is "sep".
func WithSeparatorTag(tag string) any {
	return walker.WithSeparatorTag(tag)
}

// WithSeparator sets the separator used for key-value pairs in map environment
// variable values. The default separator is ":".
func WithSeparator(sep string) any {
	return walker.WithSeparator(sep)
}

// WithDecodeUnsetTag sets the struct tag name used for decoding unset environment variables.
// The default tag name is "decodeunset".
func WithDecodeUnsetTag(tag string) any {
	return walker.WithDecodeUnsetTag(tag)
}

// WithDecodeUnset enables decoding unset environment variables.
// By default, unset environment variables are not decoded.
func WithDecodeUnset() any {
	return walker.WithDecodeUnset()
}

// WithInitTag sets the struct tag name used for initialization mode.
// The default tag name is "init".
func WithInitTag(tag string) any {
	return walker.WithInitTag(tag)
}

// WithInitNever disables automatic initialization of maps, slices, and pointers
// By default they are initialized only when a matching
// environment variable is found.
func WithInitNever() any {
	return walker.WithInitNever()
}

// WithInitAlways enables automatic initialization of maps, slices, and pointers
// regardless of whether matching environment variables are found.
// By default they are initialized only when a matching
// environment variable is found.
func WithInitAlways() any {
	return walker.WithInitAlways()
}

// WithDefaultTag sets the struct tag name used for default values.
// The default tag name is "default".
func WithDefaultTag(tag string) any {
	return matcher.WithDefaultTag(tag)
}

// WithExpandTag sets the struct tag name used for environment variable expansion.
// The default tag name is "expand".
func WithExpandTag(tag string) any {
	return matcher.WithExpandTag(tag)
}

// WithFileTag sets the struct tag name used for file paths.
// The default tag name is "file".
func WithFileTag(tag string) any {
	return matcher.WithFileTag(tag)
}

// WithNotEmptyTag sets the struct tag name used for validating that values are not empty.
// The default tag name is "notempty".
func WithNotEmptyTag(tag string) any {
	return matcher.WithNotEmptyTag(tag)
}

// WithDisableFallback enforces strict matching using the "env" tag.
// By default, it will try the field name, snake case field name, and all struct tags until a match is found.
func WithDisableFallback() any {
	return matcher.WithDisableFallback()
}

// WithDecoder registers a custom decoder function for a specific interface.
func WithDecoder(iface any, f func(v any, value string) error) any {
	return decoder.WithDecoder(iface, f)
}

// WithTypeParser registers a custom parser function for a specific type.
// This allows extending the parser to support additional types beyond
// the built-in supported types.
func WithTypeParser(t reflect.Type, f func(value string) (any, error)) any {
	return parser.WithTypeParser(t, f)
}

// WithTypeParsers registers multiple custom parser functions for specific types.
// This allows extending the parser to support additional types beyond
// the built-in supported types.
// This is a convenience function for registering multiple type parsers at once.
func WithTypeParsers(parsers map[reflect.Type]func(value string) (any, error)) any {
	return parser.WithTypeParsers(parsers)
}

// WithKindParser registers a custom parser function for a specific reflect.Kind.
// This allows extending the parser to support additional kinds beyond
// the built-in supported kinds.
func WithKindParser(k reflect.Kind, f func(value string) (any, error)) any {
	return parser.WithKindParser(k, f)
}

// WithKindParsers registers multiple custom parser functions for specific reflect.Kinds.
// This allows extending the parser to support additional kinds beyond
// the built-in supported kinds.
// This is a convenience function for registering multiple kind parsers at once.
func WithKindParsers(parsers map[reflect.Kind]func(value string) (any, error)) any {
	return parser.WithKindParsers(parsers)
}

// WithSource adds a source to the loader.
func WithSource(source loader.Source) any {
	return loader.WithSource(source)
}

// WithSources adds multiple sources to the loader.
// This is a convenience function for adding multiple sources at once.
func WithSources(sources ...loader.Source) any {
	return loader.WithSources(sources...)
}

// WithFilter registers a custom filter function for environment variables.
// The filter function is used to determine which environment variables should be used.
func WithFilter(filter func(string) bool) any {
	return loader.WithFilter(filter)
}

// WithTransform registers a custom transformation function for environment variables.
// The transformation function is used to modify environment variable keys before they are applied.
func WithTransform(transform func(string) string) any {
	return loader.WithTransform(transform)
}

// WithPrefix filters environment variables by prefix and strips the prefix
// before matching. For example, with prefix "APP_", the environment variable
// "APP_PORT=8080" would be matched as "PORT=8080".
func WithPrefix(prefix string) any {
	return loader.WithPrefix(prefix)
}

// WithSuffix filters environment variables by suffix and strips the suffix
// during matching. For example, with suffix "_TEST", the environment variable
// "PORT_TEST=8080" would be matched as "PORT=8080".
func WithSuffix(suffix string) any {
	return loader.WithSuffix(suffix)
}

// WithHasPrefix filters environment variables by prefix but preserves the prefix
// during matching. For example, with prefix "APP_", the environment variable
// "APP_PORT=8080" would be matched as "APP_PORT=8080".
func WithHasPrefix(prefix string) any {
	return loader.WithHasPrefix(prefix)
}

// WithHasSuffix filters environment variables by suffix but preserves the suffix
// during matching. For example, with suffix "_TEST", the environment variable
// "PORT_TEST=8080" would be matched as "PORT_TEST=8080".
func WithHasSuffix(suffix string) any {
	return loader.WithHasSuffix(suffix)
}

// WithHasMatch filters environment variables using a regular expression pattern.
func WithHasMatch(pattern string) any {
	return loader.WithHasMatch(pattern)
}

// WithTrimPrefix removes the specified prefix from environment variable names
// before matching. Unlike WithPrefix, it does not filter variables.
func WithTrimPrefix(prefix string) any {
	return loader.WithTrimPrefix(prefix)
}

// WithTrimSuffix removes the specified suffix from environment variable names
// before matching. Unlike WithHasSuffix, it does not filter variables.
func WithTrimSuffix(suffix string) any {
	return loader.WithTrimSuffix(suffix)
}

// WithEnvVarsSourceSource uses the provided map of environment variables instead of reading
// from the OS environment.
func WithEnvVarsSource(envs map[string]string) any {
	return loader.WithEnvVarsSource(envs)
}

// WithOsEnvSource adds OS environment variables as a source.
func WithOsEnvSource() any {
	return loader.WithOsEnvSource()
}

// WithFileSource adds environment variables from a file as a source.
// The file should contain environment variables in KEY=VALUE format.
func WithFileSource(path string) any {
	return loader.WithFileSource(path)
}

// WithDefaults adds default values as a fallback source when no other
// sources provide a value. This can be used as an alternative to
// setting default values via struct tags.
func WithDefaults(envs map[string]string) any {
	return loader.WithDefaults(envs)
}

// Parse processes the provided configuration struct using environment variables
// and the specified options. It traverses the struct fields and applies the
// environment configuration according to the defined rules and options.
func Parse(cfg any, opts ...any) error {
	w := walker.New()

	if err := w.Build(opts...); err != nil {
		return err
	}

	return w.Walk(cfg)
}

// MustParse is like Parse but panics if an error occurs during parsing.
func MustParse(cfg any, opts ...any) {
	if err := Parse(cfg, opts...); err != nil {
		panic(err)
	}
}

// ParseAs is a generic version of Parse that creates and returns a new instance
// of the specified type T with the environment configuration applied.
func ParseAs[T any](opts ...any) (T, error) {
	var t T
	err := Parse(&t, opts...)
	return t, err
}

// MustParseAs is like ParseAs but panics if an error occurs during parsing.
func MustParseAs[T any](opts ...any) T {
	t, err := ParseAs[T](opts...)
	if err != nil {
		panic(err)
	}
	return t
}
