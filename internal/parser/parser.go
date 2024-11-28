package parser

import (
	"reflect"
	"strconv"
	"time"
)

type ParserFunc func(value string) (any, error)

type Option func(*Parser)

func WithTypeParser(t reflect.Type, parserFunc ParserFunc) Option {
	return func(r *Parser) {
		r.typeParsers[t] = parserFunc
	}
}

func WithKindParser(k reflect.Kind, parserFunc ParserFunc) Option {
	return func(p *Parser) {
		p.kindParsers[k] = parserFunc
	}
}

type Parser struct {
	kindParsers map[reflect.Kind]ParserFunc
	typeParsers map[reflect.Type]ParserFunc
}

func New() *Parser {
	return &Parser{
		kindParsers: kindParsers(),
		typeParsers: typeParsers(),
	}
}

func (p *Parser) Build(opts ...any) error {
	for _, opt := range opts {
		if v, ok := opt.(Option); ok {
			v(p)
		}
	}

	return nil
}

func (p *Parser) ParseType(rt reflect.Type, value string) (any, bool, error) {
	parser, ok := p.typeParsers[rt]
	if !ok {
		return nil, false, nil
	}

	newValue, err := parser(value)
	if err != nil {
		return nil, true, err
	}

	return newValue, true, nil
}

func (p *Parser) ParseKind(k reflect.Kind, value string) (any, bool, error) {
	parser, ok := p.kindParsers[k]
	if !ok {
		return nil, false, nil
	}

	newValue, err := parser(value)
	if err != nil {
		return nil, true, err
	}

	return newValue, true, nil
}

func (p *Parser) HasParser(rt reflect.Type) bool {
	return p.typeParsers[rt] != nil || p.kindParsers[rt.Kind()] != nil
}

func typeParsers() map[reflect.Type]ParserFunc {
	return map[reflect.Type]ParserFunc{
		reflect.TypeOf(time.Nanosecond): func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			return time.ParseDuration(value)
		},
	}
}

func kindParsers() map[reflect.Kind]ParserFunc {
	return map[reflect.Kind]ParserFunc{
		reflect.String: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			return value, nil
		},
		reflect.Int: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			i, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}

			return int(i), nil
		},
		reflect.Int8: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			i, err := strconv.ParseInt(value, 10, 8)
			if err != nil {
				return nil, err
			}

			return int8(i), nil
		},
		reflect.Int16: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			i, err := strconv.ParseInt(value, 10, 16)
			if err != nil {
				return nil, err
			}

			return int16(i), nil
		},
		reflect.Int32: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			i, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return nil, err
			}

			return int32(i), nil
		},
		reflect.Int64: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}

			return int64(i), nil
		},
		reflect.Uint: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			i, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}

			return uint(i), nil
		},
		reflect.Uint8: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			i, err := strconv.ParseUint(value, 10, 8)
			if err != nil {
				return nil, err
			}

			return uint8(i), nil
		},
		reflect.Uint16: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			i, err := strconv.ParseUint(value, 10, 16)
			if err != nil {
				return nil, err
			}

			return uint16(i), nil
		},
		reflect.Uint32: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			i, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return nil, err
			}

			return uint32(i), nil
		},
		reflect.Uint64: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			i, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}

			return uint64(i), nil
		},
		reflect.Float32: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			f, err := strconv.ParseFloat(value, 32)
			if err != nil {
				return nil, err
			}

			return float32(f), nil
		},
		reflect.Float64: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, err
			}

			return float64(f), nil
		},
		reflect.Bool: func(value string) (any, error) {
			if value == "" {
				return nil, nil
			}

			b, err := strconv.ParseBool(value)
			if err != nil {
				return nil, err
			}

			return b, nil
		},
	}
}
