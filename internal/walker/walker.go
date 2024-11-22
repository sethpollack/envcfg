package walker

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sethpollack/envcfg/internal/decoder"
	"github.com/sethpollack/envcfg/internal/matcher"
	"github.com/sethpollack/envcfg/internal/parser"
	"github.com/sethpollack/envcfg/internal/tag"
)

// ErrInvalidMapFormat indicates that a map string value couldn't be parsed due to incorrect format
var ErrInvalidMapFormat = fmt.Errorf("invalid map format")

type Parser interface {
	ParseType(reflect.Type, string) (any, bool, error)
	ParseKind(reflect.Kind, string) (any, bool, error)
	HasParser(reflect.Type) bool
	Build(opts ...any) error
}

type Matcher interface {
	GetPrefix(reflect.StructField, []string) string
	GetValue(reflect.StructField, []string) (string, bool, error)
	GetMapKeys(reflect.StructField, []string) []string
	Build(opts ...any) error
}

type initMode int

const (
	initAlways initMode = iota
	initNever
	initValues
)

type walker struct {
	parser         Parser
	matcher        Matcher
	tagName        string
	delimTag       string
	defaultDelim   string
	sepTag         string
	defaultSep     string
	initTag        string
	initMode       initMode
	ignoreTag      string
	decodeUnsetTag string
	decodeUnset    bool
}

type Option func(*walker)

func WithParser(p Parser) Option {
	return func(w *walker) {
		w.parser = p
	}
}

func WithMatcher(m Matcher) Option {
	return func(w *walker) {
		w.matcher = m
	}
}

func WithTagName(tag string) Option {
	return func(w *walker) {
		w.tagName = tag
	}
}

func WithDelimiterTag(tag string) Option {
	return func(w *walker) {
		w.delimTag = tag
	}
}

func WithDelimiter(delim string) Option {
	return func(w *walker) {
		w.defaultDelim = delim
	}
}

func WithSeparatorTag(tag string) Option {
	return func(w *walker) {
		w.sepTag = tag
	}
}

func WithSeparator(sep string) Option {
	return func(w *walker) {
		w.defaultSep = sep
	}
}

func WithIgnoreTag(tag string) Option {
	return func(w *walker) {
		w.ignoreTag = tag
	}
}

func WithDecodeUnsetTag(tag string) Option {
	return func(w *walker) {
		w.decodeUnsetTag = tag
	}
}

func WithDecodeUnset() Option {
	return func(w *walker) {
		w.decodeUnset = true
	}
}

func WithInitNever() Option {
	return func(w *walker) {
		w.initMode = initNever
	}
}

func WithInitAlways() Option {
	return func(w *walker) {
		w.initMode = initAlways
	}
}

func New() *walker {
	return &walker{
		tagName:        "env",
		delimTag:       "delim",
		defaultDelim:   ",",
		sepTag:         "sep",
		defaultSep:     ":",
		initTag:        "init",
		ignoreTag:      "ignore",
		decodeUnsetTag: "decodeunset",
		initMode:       initValues,
		parser:         parser.New(),
		matcher:        matcher.New(),
	}
}

func (w *walker) Build(opts ...any) error {
	for _, opt := range opts {
		if v, ok := opt.(Option); ok {
			v(w)
		}
	}

	if err := w.parser.Build(opts...); err != nil {
		return err
	}

	if err := w.matcher.Build(opts...); err != nil {
		return err
	}

	return nil
}

func (w *walker) Walk(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer to a struct, got %T", v)
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got %T", v)
	}

	return w.walkStruct(elem)
}

func (w *walker) walkStruct(rv reflect.Value, pfx ...string) error {
	rt := rv.Type()
	// Iterate over each field in the struct.
	for i := 0; i < rt.NumField(); i++ {
		rf := rv.Field(i)

		if !rf.CanSet() {
			continue // Skip unexported fields that cannot be set.
		}

		if err := w.walkStructField(rf, rt.Field(i), pfx...); err != nil {
			return err
		}
	}

	return nil
}

func (w *walker) walkStructField(rv reflect.Value, rsf reflect.StructField, pfx ...string) error {
	tags := tag.ParseTags(rsf)

	// If the field is ignored, skip it.
	if w.parseIgnoreOption(tags) {
		return nil
	}

	var temp = reflect.New(rv.Type()).Elem()
	if isNilPtr(rv) {
		temp = reflect.New(rv.Type().Elem()).Elem()
	} else if isPtr(rv) {
		temp.Set(rv.Elem())
	}

	if w.hasParserOrSetter(temp, temp.Type()) {
		if err := w.parseStructField(temp, rsf, pfx...); err != nil {
			return err
		}
	} else {
		if err := w.walkNestedStructField(temp, rsf, pfx...); err != nil {
			return err
		}
	}

	return w.setIfNeeded(temp, rv, rsf)
}

func (w *walker) walkNestedStructField(rv reflect.Value, rsf reflect.StructField, pfx ...string) error {
	prefix := w.matcher.GetPrefix(rsf, pfx)

	switch rv.Kind() {
	case reflect.Struct:
		return w.walkStruct(rv, append(pfx, prefix)...)
	case reflect.Slice:
		return w.walkSlice(rv, rsf, prefix, pfx...)
	case reflect.Map:
		return w.walkMap(rv, rsf, prefix, pfx...)
	}

	return nil
}

func (w *walker) walkSlice(rv reflect.Value, rsf reflect.StructField, fieldName string, pfx ...string) error {
	rv.Set(reflect.MakeSlice(rv.Type(), 0, 0))

	value, _, err := w.matcher.GetValue(rsf, pfx)
	if err != nil {
		return err
	}

	if value != "" {
		return w.parseDelimitedSlice(rv, rsf, value)
	}

	for i := 0; ; i++ {
		elem := reflect.New(rv.Type().Elem()).Elem()

		prefix := fmt.Sprintf("%s_%d", fieldName, i)

		tmp := reflect.StructField{
			Name: prefix,
			Type: elem.Type(),
			Tag:  rsf.Tag,
		}

		if elem.Kind() != reflect.Struct {
			value, found, err := w.matcher.GetValue(tmp, pfx)
			if err != nil || !found {
				return err
			}

			if err := w.parseField(elem, elem.Type(), value); err != nil {
				return err
			}

			rv.Set(reflect.Append(rv, elem))

			continue
		}

		// if it has a parser or setter, parse it
		if w.hasParserOrSetter(elem, elem.Type()) {
			if err := w.parseStructField(elem, tmp, pfx...); err != nil {
				return err
			}

			if isZero(elem) {
				return nil
			}

			rv.Set(reflect.Append(rv, elem))

			continue
		}

		// If the element is a struct, walk it.
		if err := w.walkStruct(elem, append(pfx, prefix)...); err != nil {
			return err
		}

		if isZero(elem) {
			return nil
		}

		rv.Set(reflect.Append(rv, elem))

		continue
	}
}

func (w *walker) walkMap(rv reflect.Value, rsf reflect.StructField, fieldName string, pfx ...string) error {
	mapType := rv.Type()
	elemType := mapType.Elem()

	if rv.IsNil() {
		rv.Set(reflect.MakeMap(mapType))
	}

	value, _, err := w.matcher.GetValue(rsf, pfx)
	if err != nil {
		return err
	}

	if value != "" {
		return w.parseDelimitedMap(rv, rsf, value)
	}

	if elemType.Kind() == reflect.Struct {
		return w.parseMapOfStructs(rv, rsf, fieldName, pfx...)
	}

	return w.parseMap(rv, rsf, fieldName, pfx...)
}

func (w *walker) parseStructField(rv reflect.Value, rsf reflect.StructField, pfx ...string) error {
	value, found, err := w.matcher.GetValue(rsf, pfx)
	if err != nil {
		return err
	}

	decodeUnset := w.parseDecodeUnsetOption(tag.ParseTags(rsf))

	if !found && !decodeUnset {
		return nil
	}

	return w.parseField(rv, rv.Type(), value)
}

func (w *walker) setIfNeeded(temp, rv reflect.Value, rsf reflect.StructField) error {
	tags := tag.ParseTags(rsf)

	initMode := w.parseInitMode(tags)

	if initMode == initNever {
		return nil
	}

	if initMode == initValues && isZero(temp) {
		return nil
	}

	switch rv.Kind() {
	case reflect.Ptr:
		newPtr := reflect.New(rv.Type().Elem())
		newPtr.Elem().Set(temp)
		rv.Set(newPtr)
	default:
		rv.Set(temp)
	}

	return nil
}

func (w *walker) parseField(rv reflect.Value, typ reflect.Type, value string) error {
	if dec := decoder.FromReflectValue(rv); dec != nil {
		return dec.Decode(value)
	}

	nrv := rv
	if isPtr(rv) {
		typ = typ.Elem()
		nrv = rv.Elem()
	}

	if newValue, found, err := w.parser.ParseType(typ, value); found {
		if err != nil {
			return err
		}

		if newValue != nil {
			nrv.Set(reflect.ValueOf(newValue))
			return nil
		}

		return nil
	}

	if newValue, found, err := w.parser.ParseKind(typ.Kind(), value); found {
		if err != nil {
			return err
		}

		if newValue != nil {
			nrv.Set(reflect.ValueOf(newValue).Convert(typ))
			return nil
		}

		return nil
	}

	return nil
}

func (w *walker) parseDelimitedSlice(rv reflect.Value, rsf reflect.StructField, value string) error {
	delim := rsf.Tag.Get(w.delimTag)
	if delim == "" {
		delim = w.defaultDelim
	}

	elemType := rv.Type().Elem()

	for _, part := range strings.Split(value, delim) {
		elemValue := reflect.New(elemType).Elem()

		if err := w.parseField(elemValue, elemType, part); err != nil {
			return err
		}

		rv.Set(reflect.Append(rv, elemValue))
	}

	return nil
}

func (w *walker) parseDelimitedMap(rv reflect.Value, rsf reflect.StructField, value string) error {
	mapType := rv.Type()
	elemType := mapType.Elem()
	keyType := mapType.Key()

	delim := rsf.Tag.Get(w.delimTag)
	if delim == "" {
		delim = w.defaultDelim
	}

	sep := rsf.Tag.Get(w.sepTag)
	if sep == "" {
		sep = w.defaultSep
	}

	for _, part := range strings.Split(value, delim) {
		kv := strings.SplitN(part, sep, 2)
		if len(kv) != 2 {
			return fmt.Errorf("%w: expected key and value to be separated by %q, got %q", ErrInvalidMapFormat, sep, part)
		}

		keyValue := reflect.New(keyType).Elem()
		if err := w.parseField(keyValue, keyType, kv[0]); err != nil {
			return err
		}

		elemValue := reflect.New(elemType).Elem()
		if err := w.parseField(elemValue, elemType, kv[1]); err != nil {
			return err
		}

		rv.SetMapIndex(keyValue, elemValue)
	}

	return nil
}

func (w *walker) parseMapOfStructs(rv reflect.Value, rsf reflect.StructField, fieldName string, pfx ...string) error {
	keyType := rv.Type().Key()

	keys := w.matcher.GetMapKeys(rsf, pfx)
	pfx = append(pfx, fieldName)

	for _, key := range keys {
		elem := reflect.New(rv.Type().Elem()).Elem()
		pfx := append(pfx, key)

		if err := w.walkStruct(elem, pfx...); err != nil {
			return err
		}

		keyValue := reflect.New(keyType).Elem()
		if err := w.parseField(keyValue, keyType, key); err != nil {
			return err
		}

		rv.SetMapIndex(keyValue, elem)
	}

	return nil
}

func (w *walker) parseMap(rv reflect.Value, rsf reflect.StructField, fieldName string, pfx ...string) error {
	keyType := rv.Type().Key()
	elemType := rv.Type().Elem()

	pfx = append(pfx, fieldName)

	keys := w.matcher.GetMapKeys(rsf, pfx)

	for _, key := range keys {
		elem := reflect.New(rv.Type().Elem()).Elem()

		newKey := reflect.New(keyType).Elem()
		if err := w.parseField(newKey, keyType, key); err != nil {
			return err
		}

		tmp := reflect.StructField{
			Name: key,
			Type: elem.Type(),
			Tag:  rsf.Tag,
		}

		value, found, err := w.matcher.GetValue(tmp, pfx)
		if err != nil {
			return err
		}

		decodeUnset := w.parseDecodeUnsetOption(tag.ParseTags(rsf))

		if !found && !decodeUnset {
			continue
		}

		newValue := reflect.New(elemType).Elem()
		if err := w.parseField(newValue, elemType, value); err != nil {
			return err
		}

		rv.SetMapIndex(newKey, newValue)
	}

	return nil
}

func (w *walker) hasParserOrSetter(rv reflect.Value, typ reflect.Type) bool {
	if dec := decoder.FromReflectValue(reflect.New(typ).Elem()); dec != nil {
		return true
	}

	if isPtr(rv) {
		return w.parser.HasParser(typ.Elem())
	}

	return w.parser.HasParser(typ)
}

func (w *walker) parseInitMode(tags map[string]tag.Tag) initMode {
	initMode := w.initMode

	if tag, ok := tags[w.initTag]; ok {
		switch tag.Value {
		case "always":
			initMode = initAlways
		case "never":
			initMode = initNever
		case "values":
			initMode = initValues
		}
	}

	if tagName, ok := tags[w.tagName]; ok {
		switch tagName.Options[w.initTag] {
		case "always":
			initMode = initAlways
		case "never":
			initMode = initNever
		case "values":
			initMode = initValues
		}
	}

	return initMode
}

func (w *walker) parseIgnoreOption(tags map[string]tag.Tag) bool {
	if _, ok := tags[w.ignoreTag]; ok {
		return true
	}

	if tagName, ok := tags[w.tagName]; ok {
		if tagName.Value == "-" {
			return true
		}

		if _, ok := tagName.Options[w.ignoreTag]; ok {
			return true
		}
	}

	return false
}

func (w *walker) parseDecodeUnsetOption(tags map[string]tag.Tag) bool {
	if _, ok := tags[w.decodeUnsetTag]; ok {
		return true
	}

	if tagName, ok := tags[w.tagName]; ok {
		if _, ok := tagName.Options[w.decodeUnsetTag]; ok {
			return true
		}
	}

	return false
}

func isZero(rv reflect.Value) bool {
	isZero := rv.IsZero()

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Map {
		isZero = rv.Len() == 0
	}

	return isZero
}

func isPtr(rv reflect.Value) bool {
	return rv.Kind() == reflect.Ptr
}

func isNilPtr(rv reflect.Value) bool {
	return isPtr(rv) && rv.IsNil()
}
