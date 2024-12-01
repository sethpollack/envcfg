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

type InitMode int

const (
	InitValues InitMode = iota
	InitAlways
	InitNever
)

type Walker struct {
	TagName        string
	DelimTag       string
	DefaultDelim   string
	SepTag         string
	DefaultSep     string
	InitTag        string
	InitMode       InitMode
	IgnoreTag      string
	DecodeUnsetTag string
	DecodeUnset    bool

	Parser  *parser.Parser
	Matcher *matcher.Matcher
	Decoder *decoder.Decoder
}

func New() *Walker {
	return &Walker{
		TagName:        "env",
		DelimTag:       "delim",
		DefaultDelim:   ",",
		SepTag:         "sep",
		DefaultSep:     ":",
		InitTag:        "init",
		IgnoreTag:      "ignore",
		DecodeUnsetTag: "decodeunset",
		InitMode:       InitValues,

		Parser:  parser.New(),
		Matcher: matcher.New(),
		Decoder: decoder.New(),
	}
}

func (w *Walker) Walk(v any) error {
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

func (w *Walker) walkStruct(rv reflect.Value, path []tag.TagMap) error {
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

func (w *Walker) walkStructField(rv reflect.Value, path []tag.TagMap) error {
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

func (w *Walker) walkNestedStructField(rv reflect.Value, path []tag.TagMap) error {
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

func (w *Walker) walkSlice(rv reflect.Value, path []tag.TagMap) error {
	rv.Set(reflect.MakeSlice(rv.Type(), 0, 0))

	value, found, err := w.Matcher.GetValue(path)
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
				w.TagName: {Value: fmt.Sprintf("%d", i)},
			},
		})

		if !w.Matcher.HasPrefix(elemPath) {
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

func (w *Walker) walkMap(rv reflect.Value, path []tag.TagMap) error {
	current := path[len(path)-1]

	mapType := rv.Type()

	if rv.IsNil() {
		rv.Set(reflect.MakeMap(mapType))
	}

	value, _, err := w.Matcher.GetValue(path)
	if err != nil {
		return err
	}

	if value != "" {
		return w.parseDelimitedMap(rv, value, current.Tags)
	}

	return w.parseMap(rv, path)
}

func (w *Walker) parseStructField(rv reflect.Value, path []tag.TagMap) error {
	current := path[len(path)-1]

	value, found, err := w.Matcher.GetValue(path)
	if err != nil {
		return err
	}

	decodeUnset := w.parseDecodeUnsetOption(current.Tags)

	if !found && !decodeUnset {
		return nil
	}

	return w.parseField(rv, rv.Type(), value)
}

func (w *Walker) setIfNeeded(temp, rv reflect.Value, tags map[string]tag.Tag) error {
	initMode := w.parseInitMode(tags)

	if initMode == InitNever {
		return nil
	}

	if initMode == InitValues && isZero(temp) {
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

func (w *Walker) parseField(rv reflect.Value, typ reflect.Type, value string) error {
	if dec := w.Decoder.ToDecoder(rv); dec != nil {
		return dec.Decode(value)
	}

	nrv := rv
	if isPtr(rv) {
		typ = typ.Elem()
		nrv = rv.Elem()
	}

	if newValue, found, err := w.Parser.ParseType(typ, value); found {
		if err != nil {
			return err
		}

		if newValue != nil {
			nrv.Set(reflect.ValueOf(newValue))
			return nil
		}

		return nil
	}

	if newValue, found, err := w.Parser.ParseKind(typ.Kind(), value); found {
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

func (w *Walker) parseDelimitedSlice(rv reflect.Value, value string, path []tag.TagMap) error {
	current := path[len(path)-1]

	delim := w.parseDelimiter(current.Tags)

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

func (w *Walker) parseDelimitedMap(rv reflect.Value, value string, tags map[string]tag.Tag) error {
	mapType := rv.Type()
	elemType := mapType.Elem()
	keyType := mapType.Key()

	delim := w.parseDelimiter(tags)
	sep := w.parseSeparator(tags)

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

func (w *Walker) parseMap(rv reflect.Value, path []tag.TagMap) error {
	keyType := rv.Type().Key()
	elemType := rv.Type().Elem()

	keys := w.Matcher.GetMapKeys(path)

	for _, key := range keys {
		newKey := reflect.New(keyType).Elem()
		if err := w.parseField(newKey, keyType, key); err != nil {
			return err
		}

		valuePath := append(path, tag.TagMap{
			FieldName: key,
			Tags:      map[string]tag.Tag{w.TagName: {Value: key}},
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

func (w *Walker) hasParserOrSetter(rv reflect.Value, typ reflect.Type) bool {
	if dec := w.Decoder.ToDecoder(reflect.New(typ).Elem()); dec != nil {
		return true
	}

	if isPtr(rv) {
		return w.Parser.HasParser(typ.Elem())
	}

	return w.Parser.HasParser(typ)
}

func (w *Walker) parseInitMode(tags map[string]tag.Tag) InitMode {
	initMode := w.InitMode

	if tag, ok := tags[w.InitTag]; ok {
		switch tag.Value {
		case "always":
			initMode = InitAlways
		case "never":
			initMode = InitNever
		case "values":
			initMode = InitValues
		}
	}

	if tagName, ok := tags[w.TagName]; ok {
		switch tagName.Options[w.InitTag] {
		case "always":
			initMode = InitAlways
		case "never":
			initMode = InitNever
		case "values":
			initMode = InitValues
		}
	}

	return initMode
}

func (w *Walker) parseIgnoreOption(tags map[string]tag.Tag) bool {
	if _, ok := tags[w.IgnoreTag]; ok {
		return true
	}

	if tagName, ok := tags[w.TagName]; ok {
		if tagName.Value == "-" {
			return true
		}

		if _, ok := tagName.Options[w.IgnoreTag]; ok {
			return true
		}
	}

	return false
}

func (w *Walker) parseDecodeUnsetOption(tags map[string]tag.Tag) bool {
	if _, ok := tags[w.DecodeUnsetTag]; ok {
		return true
	}

	if tagName, ok := tags[w.TagName]; ok {
		if _, ok := tagName.Options[w.DecodeUnsetTag]; ok {
			return true
		}
	}

	return w.DecodeUnset
}

func (w *Walker) parseDelimiter(tags map[string]tag.Tag) string {
	if d, ok := tags[w.DelimTag]; ok {
		return d.Value
	}

	if tagName, ok := tags[w.TagName]; ok {
		if delim, ok := tagName.Options[w.DelimTag]; ok {
			return delim
		}
	}

	return w.DefaultDelim
}

func (w *Walker) parseSeparator(tags map[string]tag.Tag) string {
	if s, ok := tags[w.SepTag]; ok {
		return s.Value
	}

	if tagName, ok := tags[w.TagName]; ok {
		if sep, ok := tagName.Options[w.SepTag]; ok {
			return sep
		}
	}

	return w.DefaultSep
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
