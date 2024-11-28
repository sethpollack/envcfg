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

type initMode int

const (
	initAlways initMode = iota
	initNever
	initValues
)

type walker struct {
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

	parser  *parser.Parser
	matcher *matcher.Matcher
	decoder *decoder.Decoder
}

type Option func(*walker)

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
		decoder:        decoder.New(),
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

	if err := w.matcher.Build(
		// Pass the tag name as an option to the matcher
		append(opts, matcher.WithTagName(w.tagName))...); err != nil {
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

	return w.walkStruct(elem, []tag.TagMap{})
}

func (w *walker) walkStruct(rv reflect.Value, path []tag.TagMap) error {
	rt := rv.Type()
	// Iterate over each field in the struct.
	for i := 0; i < rt.NumField(); i++ {
		rf := rv.Field(i)

		if !rf.CanSet() {
			continue // Skip unexported fields that cannot be set.
		}

		fieldTags := tag.ParseTags(rt.Field(i))
		fieldPath := append(path, fieldTags)

		if err := w.walkStructField(rf, fieldPath); err != nil {
			return err
		}
	}

	return nil
}

func (w *walker) walkStructField(rv reflect.Value, path []tag.TagMap) error {
	current := path[len(path)-1]

	// If the field is ignored, skip it.
	if w.parseIgnoreOption(current.Tags) {
		return nil
	}

	var temp = reflect.New(rv.Type()).Elem()
	if isNilPtr(rv) {
		temp = reflect.New(rv.Type().Elem()).Elem()
	} else if isPtr(rv) {
		temp.Set(rv.Elem())
	}

	if w.hasParserOrSetter(temp, temp.Type()) {
		if err := w.parseStructField(temp, path); err != nil {
			return err
		}
	} else {
		if err := w.walkNestedStructField(temp, path); err != nil {
			return err
		}
	}

	return w.setIfNeeded(temp, rv, current.Tags)
}

func (w *walker) walkNestedStructField(rv reflect.Value, path []tag.TagMap) error {
	switch rv.Kind() {
	case reflect.Struct:
		return w.walkStruct(rv, path)
	case reflect.Slice:
		return w.walkSlice(rv, path)
	case reflect.Map:
		return w.walkMap(rv, path)
	}

	return nil
}

func (w *walker) walkSlice(rv reflect.Value, path []tag.TagMap) error {
	rv.Set(reflect.MakeSlice(rv.Type(), 0, 0))

	value, found, err := w.matcher.GetValue(path)
	if err != nil {
		return err
	}

	if found && value != "" {
		return w.parseDelimitedSlice(rv, value, path)
	}

	for i := 0; ; i++ {
		elem := reflect.New(rv.Type().Elem()).Elem()

		elemPath := append(path, tag.TagMap{
			FieldName: fmt.Sprintf("%d", i),
			Tags: map[string]tag.Tag{
				w.tagName: {Value: fmt.Sprintf("%d", i)},
			},
		})

		if !w.matcher.HasPrefix(elemPath) {
			return nil
		}

		if w.hasParserOrSetter(elem, elem.Type()) {
			if err := w.parseStructField(elem, elemPath); err != nil {
				return err
			}
		} else {
			if err := w.walkStructField(elem, elemPath); err != nil {
				return err
			}
		}

		rv.Set(reflect.Append(rv, elem))
	}
}

func (w *walker) walkMap(rv reflect.Value, path []tag.TagMap) error {
	current := path[len(path)-1]

	mapType := rv.Type()

	if rv.IsNil() {
		rv.Set(reflect.MakeMap(mapType))
	}

	value, _, err := w.matcher.GetValue(path)
	if err != nil {
		return err
	}

	if value != "" {
		return w.parseDelimitedMap(rv, value, current.Tags)
	}

	return w.parseMap(rv, path)
}

func (w *walker) parseStructField(rv reflect.Value, path []tag.TagMap) error {
	current := path[len(path)-1]

	value, found, err := w.matcher.GetValue(path)
	if err != nil {
		return err
	}

	decodeUnset := w.parseDecodeUnsetOption(current.Tags)

	if !found && !decodeUnset {
		return nil
	}

	return w.parseField(rv, rv.Type(), value)
}

func (w *walker) setIfNeeded(temp, rv reflect.Value, tags map[string]tag.Tag) error {
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
	if dec := w.decoder.ToDecoder(rv); dec != nil {
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

func (w *walker) parseDelimitedSlice(rv reflect.Value, value string, path []tag.TagMap) error {
	current := path[len(path)-1]

	delim := w.defaultDelim
	if d, ok := current.Tags[w.delimTag]; ok {
		delim = d.Value
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

func (w *walker) parseDelimitedMap(rv reflect.Value, value string, tags map[string]tag.Tag) error {
	mapType := rv.Type()
	elemType := mapType.Elem()
	keyType := mapType.Key()

	delim := w.defaultDelim
	if d, ok := tags[w.delimTag]; ok {
		delim = d.Value
	}

	sep := w.defaultSep
	if s, ok := tags[w.sepTag]; ok {
		sep = s.Value
	}

	for _, part := range strings.Split(value, delim) {
		kv := strings.SplitN(part, sep, 2)
		if len(kv) != 2 {
			return fmt.Errorf("expected key and value to be separated by %q, got %q", sep, part)
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

func (w *walker) parseMap(rv reflect.Value, path []tag.TagMap) error {
	keyType := rv.Type().Key()
	elemType := rv.Type().Elem()

	keys := w.matcher.GetMapKeys(path)

	for _, key := range keys {
		newKey := reflect.New(keyType).Elem()
		if err := w.parseField(newKey, keyType, key); err != nil {
			return err
		}

		valuePath := append(path, tag.TagMap{
			FieldName: key,
			Tags:      map[string]tag.Tag{w.tagName: {Value: key}},
		})

		newValue := reflect.New(elemType).Elem()
		if err := w.walkStructField(newValue, valuePath); err != nil {
			return err
		}

		if isZero(newValue) {
			continue
		}

		rv.SetMapIndex(newKey, newValue)
	}

	return nil
}

func (w *walker) hasParserOrSetter(rv reflect.Value, typ reflect.Type) bool {
	if dec := w.decoder.ToDecoder(reflect.New(typ).Elem()); dec != nil {
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
