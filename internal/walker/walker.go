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

type Value struct {
	reflect.Value
	IsSet     bool
	IsDefault bool
	Path      []tag.TagMap
}

type InitMode int

const (
	InitVars InitMode = iota
	InitAny
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
		InitMode:       InitVars,

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

	return w.walkStruct(&Value{
		Value: elem,
		Path:  []tag.TagMap{},
	})
}

func (w *Walker) visit(v *Value) error {
	if isNilPtr(v) {
		initMode := w.initMode(v.Path)

		tmp := &Value{
			Value: reflect.New(v.Type().Elem()).Elem(),
			Path:  v.Path,
		}

		if initMode == InitNever {
			return nil
		}

		err := w.visit(tmp)
		if err != nil {
			return err
		}

		if tmp.Kind() == reflect.Struct {
			if initMode == InitVars && !tmp.IsSet {
				return nil
			}
		}

		// never init empty pointers unless init mode is always
		if initMode != InitAlways && (!tmp.IsSet && !tmp.IsDefault) {
			return nil
		}

		newPtr := reflect.New(v.Type().Elem())
		newPtr.Elem().Set(tmp.Value)
		v.Set(newPtr)

		v.IsSet = tmp.IsSet
		v.IsDefault = tmp.IsDefault

		return nil
	}

	if isPtr(v) {
		v.Value = v.Value.Elem()
	}

	value, isSet, isDefault, err := w.Matcher.GetValue(v.Path)
	if err != nil {
		return err
	}

	if w.hasParserOrSetter(v) {
		if (!isSet && !isDefault) && !w.decodeUnset(v.Path) {
			return nil
		}

		return w.parse(v, value, isDefault)
	}

	if value != "" {
		switch v.Kind() {
		case reflect.Slice:
			return w.walkDelimitedSlice(v, value, isDefault)
		case reflect.Map:
			return w.walkDelimitedMap(v, value, isDefault)
		}
	}

	switch v.Kind() {
	case reflect.Struct:
		return w.walkStruct(v)
	case reflect.Slice:
		return w.walkSlice(v)
	case reflect.Map:
		return w.walkMap(v)
	}

	return nil
}

func (w *Walker) walkStruct(v *Value) error {
	rt := v.Type()
	// Iterate over each field in the struct.
	for i := 0; i < rt.NumField(); i++ {
		rf := v.Field(i)

		if !rf.CanSet() {
			continue // Skip unexported fields that cannot be set.
		}

		fieldPath := append(v.Path, tag.ParseTags(rt.Field(i)))

		if w.ignore(fieldPath) {
			continue
		}

		child := &Value{Value: rf, Path: fieldPath}

		err := w.visit(child)
		if err != nil {
			return err
		}

		if child.IsSet {
			v.IsSet = true
			v.IsDefault = false
		} else if child.IsDefault && !v.IsSet {
			v.IsDefault = true
		}
	}

	return nil
}

func (w *Walker) walkDelimitedSlice(v *Value, value string, isDefault bool) error {
	delim := w.delimiter(v.Path)

	elemType := v.Type().Elem()

	for _, part := range strings.Split(value, delim) {
		elemValue := &Value{
			Value: reflect.New(elemType).Elem(),
			Path:  v.Path,
		}

		if err := w.parse(elemValue, part, isDefault); err != nil {
			return err
		}

		appendSlice(v, elemValue)
	}

	return nil
}

func (w *Walker) walkSlice(v *Value) error {
	for i := 0; ; i++ {
		elemPath := append(v.Path, tag.TagMap{
			FieldName: fmt.Sprintf("%d", i),
			Tags: map[string]tag.Tag{
				w.TagName: {Value: fmt.Sprintf("%d", i)},
			},
		})

		if !w.Matcher.HasPrefix(elemPath) {
			return nil
		}

		elemValue := &Value{
			Value: reflect.New(v.Type().Elem()).Elem(),
			Path:  elemPath,
		}

		err := w.visit(elemValue)
		if err != nil {
			return err
		}

		appendSlice(v, elemValue)
	}
}

func (w *Walker) walkDelimitedMap(v *Value, value string, isDefault bool) error {
	mapType := v.Type()
	elemType := mapType.Elem()
	keyType := mapType.Key()

	delim := w.delimiter(v.Path)
	sep := w.separator(v.Path)

	parts := strings.Split(value, delim)
	if len(parts) == 0 {
		return nil
	}

	for _, part := range parts {
		kv := strings.SplitN(part, sep, 2)
		if len(kv) != 2 {
			return fmt.Errorf("expected key and value to be separated by %q, got %q", sep, part)
		}

		keyValue := &Value{
			Value: reflect.New(keyType).Elem(),
			Path:  v.Path,
		}

		if err := w.parse(keyValue, kv[0], isDefault); err != nil {
			return err
		}

		elemValue := &Value{
			Value: reflect.New(elemType).Elem(),
			Path:  v.Path,
		}

		if err := w.parse(elemValue, kv[1], isDefault); err != nil {
			return err
		}

		setMapIndex(v, keyValue, elemValue)
	}

	return nil
}

func (w *Walker) walkMap(v *Value) error {
	keyType := v.Type().Key()
	elemType := v.Type().Elem()

	keys := w.Matcher.GetMapKeys(v.Path)
	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		newKey := &Value{
			Value: reflect.New(keyType).Elem(),
			Path:  v.Path,
		}

		if err := w.parse(newKey, key, false); err != nil {
			return err
		}

		valuePath := append(v.Path, tag.TagMap{
			FieldName: key,
			Tags:      map[string]tag.Tag{w.TagName: {Value: key}},
		})

		newValue := &Value{
			Value: reflect.New(elemType).Elem(),
			Path:  valuePath,
		}

		if err := w.visit(newValue); err != nil {
			return err
		}

		setMapIndex(v, newKey, newValue)
	}

	return nil
}

func (w *Walker) hasParserOrSetter(v *Value) bool {
	if dec := w.Decoder.ToDecoder(reflect.New(v.Type()).Elem()); dec != nil {
		return true
	}

	if isPtr(v) {
		return w.Parser.HasParser(v.Type().Elem())
	}

	return w.Parser.HasParser(v.Type())
}

func (w *Walker) parse(v *Value, value string, isDefault bool) error {
	if dec := w.Decoder.ToDecoder(v.Value); dec != nil {
		if err := dec.Decode(value); err != nil {
			return err
		}

		if isDefault {
			v.IsDefault = true
		} else {
			v.IsSet = true
		}

		return nil
	}

	nv := v.Value
	typ := v.Type()

	if isPtr(v) {
		typ = typ.Elem()
		nv = nv.Elem()
	}

	if newValue, found, err := w.Parser.ParseType(typ, value); found {
		if err != nil {
			return err
		}

		if newValue != nil {
			nv.Set(reflect.ValueOf(newValue))
			if isDefault {
				v.IsDefault = true
			} else {
				v.IsSet = true
			}
			return nil
		}

		return nil
	}

	if newValue, found, err := w.Parser.ParseKind(typ.Kind(), value); found {
		if err != nil {
			return err
		}

		if newValue != nil {
			nv.Set(reflect.ValueOf(newValue).Convert(typ))
			if isDefault {
				v.IsDefault = true
			} else {
				v.IsSet = true
			}
			return nil
		}

		return nil
	}

	return nil
}

func (w *Walker) initTag(path []tag.TagMap) string {
	current := path[len(path)-1]

	if tag, ok := current.Tags[w.InitTag]; ok {
		return tag.Value
	}

	if tagName, ok := current.Tags[w.TagName]; ok {
		if tv, ok := tagName.Options[w.InitTag]; ok {
			return tv
		}
	}

	return ""
}

func (w *Walker) initMode(path []tag.TagMap) InitMode {
	switch w.initTag(path) {
	case "always":
		return InitAlways
	case "never":
		return InitNever
	case "vars":
		return InitVars
	case "any":
		return InitAny
	default:
		return w.InitMode
	}
}

func (w *Walker) ignore(path []tag.TagMap) bool {
	current := path[len(path)-1]

	if _, ok := current.Tags[w.IgnoreTag]; ok {
		return true
	}

	if tagName, ok := current.Tags[w.TagName]; ok {
		if tagName.Value == "-" {
			return true
		}

		if _, ok := tagName.Options[w.IgnoreTag]; ok {
			return true
		}
	}

	return false
}

func (w *Walker) decodeUnset(path []tag.TagMap) bool {
	current := path[len(path)-1]

	if _, ok := current.Tags[w.DecodeUnsetTag]; ok {
		return true
	}

	if tagName, ok := current.Tags[w.TagName]; ok {
		if _, ok := tagName.Options[w.DecodeUnsetTag]; ok {
			return true
		}
	}

	return w.DecodeUnset
}

func (w *Walker) delimiter(path []tag.TagMap) string {
	current := path[len(path)-1]

	if d, ok := current.Tags[w.DelimTag]; ok {
		return d.Value
	}

	if tagName, ok := current.Tags[w.TagName]; ok {
		if delim, ok := tagName.Options[w.DelimTag]; ok {
			return delim
		}
	}

	return w.DefaultDelim
}

func (w *Walker) separator(path []tag.TagMap) string {
	current := path[len(path)-1]

	if s, ok := current.Tags[w.SepTag]; ok {
		return s.Value
	}

	if tagName, ok := current.Tags[w.TagName]; ok {
		if sep, ok := tagName.Options[w.SepTag]; ok {
			return sep
		}
	}

	return w.DefaultSep
}

func isPtr(v *Value) bool {
	return v.Value.Kind() == reflect.Ptr
}

func isNilPtr(v *Value) bool {
	return isPtr(v) && v.Value.IsNil()
}

func appendSlice(v, e *Value) {
	if !e.IsSet && !e.IsDefault {
		return
	}

	if v.IsNil() {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}

	v.Set(reflect.Append(v.Value, e.Value))

	if e.IsSet {
		v.IsSet = true
		v.IsDefault = false
	} else if e.IsDefault && !v.IsSet {
		v.IsDefault = true
	}
}

func setMapIndex(v, k, e *Value) {
	if !e.IsSet && !e.IsDefault {
		return
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	v.SetMapIndex(k.Value, e.Value)

	if e.IsSet {
		v.IsSet = true
		v.IsDefault = false
	} else if e.IsDefault && !v.IsSet {
		v.IsDefault = true
	}
}
