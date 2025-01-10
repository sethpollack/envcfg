package envcfg

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/sethpollack/envcfg/internal/decoder"
	"github.com/sethpollack/envcfg/internal/loader"
	"github.com/sethpollack/envcfg/internal/matcher"
	"github.com/sethpollack/envcfg/internal/parser"
	"github.com/sethpollack/envcfg/internal/walker"
	"github.com/sethpollack/envcfg/sources/dotenv"
	"github.com/sethpollack/envcfg/sources/mapenv"
	"github.com/sethpollack/envcfg/sources/osenv"
)

type Option func(*Options)

type Options struct {
	Walker  *walker.Walker
	Loader  *loader.Loader
	Decoder *decoder.Decoder
	Parser  *parser.Parser
	Matcher *matcher.Matcher
}

func Build(opts ...Option) (*Options, error) {
	o := &Options{
		Walker:  walker.New(),
		Decoder: decoder.New(),
		Loader:  loader.New(),
		Matcher: matcher.New(),
		Parser:  parser.New(),
	}

	for _, opt := range opts {
		opt(o)
	}

	if len(o.Loader.Sources) == 0 {
		o.Loader.Sources = []loader.Source{osenv.New()}
	}

	loaded, err := o.Loader.Load()
	if err != nil {
		return nil, err
	}

	o.Matcher.EnvVars = loaded
	o.Walker.Matcher = o.Matcher
	o.Walker.Decoder = o.Decoder
	o.Walker.Parser = o.Parser

	return o, nil
}

// WithTagName sets a custom struct tag name to override the default "env" tag.
func WithTagName(tag string) Option {
	return func(o *Options) {
		o.Walker.TagName = tag
		o.Matcher.TagName = tag
	}
}

// WithDelimiterTag sets the struct tag name used for the delimiter.
// The default tag name is "delim".
func WithDelimiterTag(tag string) Option {
	return func(o *Options) {
		o.Walker.DelimTag = tag
	}
}

// WithDelimiter sets the delimiter used to separate slice/map elements
// in environment variable values. The default delimiter is ",".
func WithDelimiter(delim string) Option {
	return func(o *Options) {
		o.Walker.DefaultDelim = delim
	}
}

// WithSeparatorTag sets the struct tag name used for the separator.
// The default tag name is "sep".
func WithSeparatorTag(tag string) Option {
	return func(o *Options) {
		o.Walker.SepTag = tag
	}
}

// WithSeparator sets the separator used for key-value pairs in map environment
// variable values. The default separator is ":".
func WithSeparator(sep string) Option {
	return func(o *Options) {
		o.Walker.DefaultSep = sep
	}
}

// WithDecodeUnsetTag sets the struct tag name used for decoding unset environment variables.
// The default tag name is "decodeunset".
func WithDecodeUnsetTag(tag string) Option {
	return func(o *Options) {
		o.Walker.DecodeUnsetTag = tag
	}
}

// WithDecodeUnset enables decoding unset environment variables.
// By default, unset environment variables are not decoded.
func WithDecodeUnset() Option {
	return func(o *Options) {
		o.Walker.DecodeUnset = true
	}
}

// WithInitTag sets the struct tag name used for initialization mode.
// The default tag name is "init".
func WithInitTag(tag string) Option {
	return func(o *Options) {
		o.Walker.InitTag = tag
	}
}

// WithInitAny enables automatic initialization nil pointers
// if environment variables are found or if default values are provided.
// By default they are initialized only when a matching
// environment variable is found.
func WithInitAny() Option {
	return func(o *Options) {
		o.Walker.InitMode = walker.InitAny
	}
}

// WithInitNever disables automatic initialization of nil pointers
// By default they are initialized only when a matching
// environment variable is found.
func WithInitNever() Option {
	return func(o *Options) {
		o.Walker.InitMode = walker.InitNever
	}
}

// WithInitAlways enables automatic initialization of nil pointers
// regardless of whether matching environment variables are found.
// By default they are initialized only when a matching
// environment variable is found.
func WithInitAlways() Option {
	return func(o *Options) {
		o.Walker.InitMode = walker.InitAlways
	}
}

// WithDefaultTag sets the struct tag name used for default values.
// The default tag name is "default".
func WithDefaultTag(tag string) Option {
	return func(o *Options) {
		o.Matcher.DefaultTag = tag
	}
}

// WithExpandTag sets the struct tag name used for environment variable expansion.
// The default tag name is "expand".
func WithExpandTag(tag string) Option {
	return func(o *Options) {
		o.Matcher.ExpandTag = tag
	}
}

// WithFileTag sets the struct tag name used for file paths.
// The default tag name is "file".
func WithFileTag(tag string) Option {
	return func(o *Options) {
		o.Matcher.FileTag = tag
	}
}

// WithNotEmptyTag sets the struct tag name used for validating that values are not empty.
// The default tag name is "notempty".
func WithNotEmptyTag(tag string) Option {
	return func(o *Options) {
		o.Matcher.NotEmptyTag = tag
	}
}

// WithNotEmpty is a global setting to validate that values are not empty.
// By default, empty values are not allowed.
func WithNotEmpty() Option {
	return func(o *Options) {
		o.Matcher.NotEmpty = true
	}
}

// WithExpand is a global setting to expand environment variables in values.
// By default, environment variables are not expanded.
func WithExpand() Option {
	return func(o *Options) {
		o.Matcher.Expand = true
	}
}

// WithRequiredTag sets the struct tag name used for required values.
// The default tag name is "required".
func WithRequiredTag(tag string) Option {
	return func(o *Options) {
		o.Matcher.RequiredTag = tag
	}
}

// WithRequired is a global setting to validate that values are required.
// By default, fields are not required.
func WithRequired() Option {
	return func(o *Options) {
		o.Matcher.Required = true
	}
}

// WithDisableFallback enforces strict matching using the "env" tag.
// By default, it will try the field name, snake case field name, and all struct tags until a match is found.
func WithDisableFallback() Option {
	return func(o *Options) {
		o.Matcher.DisableFallback = true
	}
}

// WithDecoder registers a custom decoder function for a specific interface.
func WithDecoder(iface any, f func(v any, value string) error) Option {
	return func(o *Options) {
		o.Decoder.Decoders[iface] = f
	}
}

// WithTypeParser registers a custom parser function for a specific type.
// This allows extending the parser to support additional types beyond
// the built-in supported types.
func WithTypeParser(t reflect.Type, f func(value string) (any, error)) Option {
	return func(o *Options) {
		o.Parser.TypeParsers[t] = f
	}
}

// WithTypeParsers registers multiple custom parser functions for specific types.
// This allows extending the parser to support additional types beyond
// the built-in supported types.
// This is a convenience function for registering multiple type parsers at once.
func WithTypeParsers(parsers map[reflect.Type]func(value string) (any, error)) Option {
	return func(o *Options) {
		for t, f := range parsers {
			o.Parser.TypeParsers[t] = f
		}
	}
}

// WithKindParser registers a custom parser function for a specific reflect.Kind.
// This allows extending the parser to support additional kinds beyond
// the built-in supported kinds.
func WithKindParser(k reflect.Kind, f func(value string) (any, error)) Option {
	return func(o *Options) {
		o.Parser.KindParsers[k] = f
	}
}

// WithKindParsers registers multiple custom parser functions for specific reflect.Kinds.
// This allows extending the parser to support additional kinds beyond
// the built-in supported kinds.
// This is a convenience function for registering multiple kind parsers at once.
func WithKindParsers(parsers map[reflect.Kind]func(value string) (any, error)) Option {
	return func(o *Options) {
		for k, f := range parsers {
			o.Parser.KindParsers[k] = f
		}
	}
}

type LoaderOption func(*loader.Loader)

func WithLoader(opts ...LoaderOption) Option {
	return func(o *Options) {
		l := loader.New()

		for _, opt := range opts {
			opt(l)
		}

		if len(l.Sources) == 0 {
			l.Sources = []loader.Source{osenv.New()}
		}

		o.Loader.Sources = append(o.Loader.Sources, l)
	}
}

// WithSource adds a source to the loader.
func WithSource(source loader.Source) LoaderOption {
	return func(l *loader.Loader) {
		l.Sources = append(l.Sources, source)
	}
}

// WithSources adds multiple sources to the loader.
// This is a convenience function for adding multiple sources at once.
func WithSources(sources ...loader.Source) LoaderOption {
	return func(l *loader.Loader) {
		l.Sources = append(l.Sources, sources...)
	}
}

// WithFilter registers a custom filter function for environment variables.
// The filter function is used to determine which environment variables should be used.
func WithFilter(filter func(string) bool) LoaderOption {
	return func(l *loader.Loader) {
		l.Filters = append(l.Filters, filter)
	}
}

// WithTransform registers a custom transformation function for environment variables.
// The transformation function is used to modify environment variable keys before they are applied.
func WithTransform(transform func(string) string) LoaderOption {
	return func(l *loader.Loader) {
		l.Transforms = append(l.Transforms, transform)
	}
}

// WithPrefix filters environment variables by prefix and strips the prefix
// before matching. For example, with prefix "APP_", the environment variable
// "APP_PORT=8080" would be matched as "PORT=8080".
func WithPrefix(prefix string) LoaderOption {
	return func(l *loader.Loader) {
		l.Filters = append(l.Filters, func(key string) bool {
			return strings.HasPrefix(key, prefix)
		})

		l.Transforms = append(l.Transforms, func(key string) string {
			return strings.TrimPrefix(key, prefix)
		})
	}
}

// WithSuffix filters environment variables by suffix and strips the suffix
// during matching. For example, with suffix "_TEST", the environment variable
// "PORT_TEST=8080" would be matched as "PORT=8080".
func WithSuffix(suffix string) LoaderOption {
	return func(l *loader.Loader) {
		l.Filters = append(l.Filters, func(key string) bool {
			return strings.HasSuffix(key, suffix)
		})

		l.Transforms = append(l.Transforms, func(key string) string {
			return strings.TrimSuffix(key, suffix)
		})
	}
}

// WithHasPrefix filters environment variables by prefix but preserves the prefix
// during matching. For example, with prefix "APP_", the environment variable
// "APP_PORT=8080" would be matched as "APP_PORT=8080".
func WithHasPrefix(prefix string) LoaderOption {
	return func(l *loader.Loader) {
		l.Filters = append(l.Filters, func(key string) bool {
			return strings.HasPrefix(key, prefix)
		})
	}
}

// WithHasSuffix filters environment variables by suffix but preserves the suffix
// during matching. For example, with suffix "_TEST", the environment variable
// "PORT_TEST=8080" would be matched as "PORT_TEST=8080".
func WithHasSuffix(suffix string) LoaderOption {
	return func(l *loader.Loader) {
		l.Filters = append(l.Filters, func(key string) bool {
			return strings.HasSuffix(key, suffix)
		})
	}
}

// WithHasMatch filters environment variables using a regular expression pattern.
func WithHasMatch(pattern *regexp.Regexp) LoaderOption {
	return func(l *loader.Loader) {
		l.Filters = append(l.Filters, func(key string) bool {
			return pattern.MatchString(key)
		})
	}
}

// WithTrimPrefix removes the specified prefix from environment variable names
// before matching. Unlike WithPrefix, it does not filter variables.
func WithTrimPrefix(prefix string) LoaderOption {
	return func(l *loader.Loader) {
		l.Transforms = append(l.Transforms, func(key string) string {
			return strings.TrimPrefix(key, prefix)
		})
	}
}

// WithTrimSuffix removes the specified suffix from environment variable names
// before matching. Unlike WithHasSuffix, it does not filter variables.
func WithTrimSuffix(suffix string) LoaderOption {
	return func(l *loader.Loader) {
		l.Transforms = append(l.Transforms, func(key string) string {
			return strings.TrimSuffix(key, suffix)
		})
	}
}

// WithMapEnvSource uses the provided map of environment variables instead of reading
// from the OS environment.
func WithMapEnvSource(envs map[string]string) LoaderOption {
	return func(l *loader.Loader) {
		l.Sources = append(l.Sources, mapenv.New(envs))
	}
}

// WithOSEnvSource adds OS environment variables as a source.
func WithOSEnvSource() LoaderOption {
	return func(l *loader.Loader) {
		l.Sources = append(l.Sources, osenv.New())
	}
}

// WithDotEnvSource adds environment variables from a file as a source.
// The file should contain environment variables in KEY=VALUE format.
func WithDotEnvSource(path string) LoaderOption {
	return func(l *loader.Loader) {
		l.Sources = append(l.Sources, dotenv.New(path))
	}
}

// Parse processes the provided configuration struct using environment variables
// and the specified options. It traverses the struct fields and applies the
// environment configuration according to the defined rules and options.
func Parse(cfg any, opts ...Option) error {
	b, err := Build(opts...)
	if err != nil {
		return err
	}

	return b.Walker.Walk(cfg)
}

// MustParse is like Parse but panics if an error occurs during parsing.
func MustParse(cfg any, opts ...Option) {
	if err := Parse(cfg, opts...); err != nil {
		panic(err)
	}
}

// ParseAs is a generic version of Parse that creates and returns a new instance
// of the specified type T with the environment configuration applied.
func ParseAs[T any](opts ...Option) (T, error) {
	var t T
	err := Parse(&t, opts...)
	return t, err
}

// MustParseAs is like ParseAs but panics if an error occurs during parsing.
func MustParseAs[T any](opts ...Option) T {
	t, err := ParseAs[T](opts...)
	if err != nil {
		panic(err)
	}
	return t
}
