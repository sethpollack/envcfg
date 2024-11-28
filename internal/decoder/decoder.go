package decoder

import (
	"encoding"
	"flag"
	"reflect"
)

type Decode interface {
	Decode(value string) error
}

type wrapper struct {
	decoder func(value string) error
}

func (u wrapper) Decode(value string) error {
	return u.decoder(value)
}

type DecodeBuilderFunc func(v any, value string) error

type Option func(*Decoder)

func WithDecoder(iface any, f DecodeBuilderFunc) Option {
	return func(r *Decoder) {
		r.decoders[iface] = f
	}
}

type Decoder struct {
	decoders map[any]DecodeBuilderFunc
}

func New() *Decoder {
	return &Decoder{
		decoders: make(map[any]DecodeBuilderFunc),
	}
}

func (r *Decoder) Build(opts ...any) error {
	for _, opt := range opts {
		if v, ok := opt.(Option); ok {
			v(r)
		}
	}
	return nil
}

func (r *Decoder) ToDecoder(rv reflect.Value) Decode {
	if !rv.IsValid() || !rv.CanInterface() {
		return nil
	}

	// Handle both value and pointer types
	var v interface{}
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		v = rv.Interface()
	} else {
		if rv.CanAddr() {
			v = rv.Addr().Interface()
		} else {
			v = rv.Interface()
		}
	}

	return r.toDecoder(v)
}

func (r *Decoder) toDecoder(v any) Decode {
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case Decode:
		return &wrapper{func(value string) error {
			return v.Decode(value)
		}}
	case flag.Value:
		return &wrapper{func(value string) error {
			return v.Set(value)
		}}
	case encoding.TextUnmarshaler:
		return &wrapper{func(value string) error {
			return v.UnmarshalText([]byte(value))
		}}
	case encoding.BinaryUnmarshaler:
		return &wrapper{func(value string) error {
			return v.UnmarshalBinary([]byte(value))
		}}
	}

	// Check custom decoders
	for iface, f := range r.decoders {
		if reflect.TypeOf(v).Implements(reflect.TypeOf(iface).Elem()) {
			return &wrapper{func(value string) error {
				return f(v, value)
			}}
		}
	}

	return nil
}
