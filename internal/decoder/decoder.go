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

type Decoder struct {
	Decoders map[any]DecodeBuilderFunc
}

func New() *Decoder {
	return &Decoder{
		Decoders: make(map[any]DecodeBuilderFunc),
	}
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
	for iface, f := range r.Decoders {
		if reflect.TypeOf(v).Implements(reflect.TypeOf(iface).Elem()) {
			return &wrapper{func(value string) error {
				return f(v, value)
			}}
		}
	}

	return nil
}
