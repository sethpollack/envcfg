package matcher

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/sethpollack/envcfg/internal/loader"
	"github.com/sethpollack/envcfg/internal/tag"
)

type Loader interface {
	Load() map[string]string
	Build(opts ...any) error
}

type Option func(*matcher)

func WithTagName(tagName string) Option {
	return func(m *matcher) {
		m.tagName = tagName
	}
}

func WithDefaultTag(tag string) Option {
	return func(m *matcher) {
		m.defaultTag = tag
	}
}

func WithExpandTag(tag string) Option {
	return func(m *matcher) {
		m.expandTag = tag
	}
}

func WithFileTag(tag string) Option {
	return func(m *matcher) {
		m.fileTag = tag
	}
}

func WithNotEmptyTag(tag string) Option {
	return func(m *matcher) {
		m.notEmptyTag = tag
	}
}

func WithRequiredTag(tag string) Option {
	return func(m *matcher) {
		m.requiredTag = tag
	}
}

func WithLoader(loader Loader) Option {
	return func(m *matcher) {
		m.loader = loader
	}
}

func WithDisableFallback() Option {
	return func(m *matcher) {
		m.disableFallback = true
	}
}

type matcher struct {
	tagName         string
	defaultTag      string
	expandTag       string
	fileTag         string
	notEmptyTag     string
	requiredTag     string
	disableFallback bool
	loader          Loader
	envVars         map[string]string
}

func New() *matcher {
	return &matcher{
		tagName:     "env",
		defaultTag:  "default",
		expandTag:   "expand",
		fileTag:     "file",
		notEmptyTag: "notempty",
		requiredTag: "required",
		loader:      loader.New(),
	}
}

func (m *matcher) Build(opts ...any) error {
	for _, opt := range opts {
		if v, ok := opt.(Option); ok {
			v(m)
		}
	}

	if err := m.loader.Build(opts...); err != nil {
		return err
	}

	m.envVars = m.loader.Load()

	return nil
}

func (m *matcher) GetValue(rsf reflect.StructField, prefixes []string) (string, bool, error) {
	parsedTags := tag.ParseTags(rsf)
	opts := m.parseOptions(parsedTags)

	var foundMatch bool
	var foundKey string
	var foundValue string

	for tagName, tag := range parsedTags {
		if m.disableFallback && m.tagName != tagName {
			continue
		}

		found, key, value := m.getValue(tag.Value, prefixes)
		if found {
			foundMatch = true
			foundKey = key
			foundValue = value
		}
	}

	if !foundMatch {
		if _, ok := opts[m.requiredTag]; ok {
			return "", false, fmt.Errorf("required field %s not found", rsf.Name)
		}

		if _, ok := opts[m.defaultTag]; ok {
			if _, ok := opts[m.expandTag]; ok {
				return m.expandValue(opts[m.defaultTag]), true, nil
			}
			return opts[m.defaultTag], true, nil
		}

		return "", false, nil
	}

	if _, ok := opts[m.notEmptyTag]; ok && foundValue == "" {
		return "", true, fmt.Errorf("environment variable %s is empty", foundKey)
	}

	if _, ok := opts[m.fileTag]; ok {
		bytes, err := os.ReadFile(foundValue)
		if err != nil {
			return "", true, err
		}

		if _, ok := opts[m.expandTag]; ok {
			return m.expandValue(string(bytes)), true, nil
		}

		return string(bytes), true, nil
	}

	if _, ok := opts[m.expandTag]; ok {
		return m.expandValue(foundValue), true, nil
	}

	return foundValue, true, nil
}

func (m *matcher) GetPrefix(rsf reflect.StructField, prefixes []string) string {
	parsedTags := tag.ParseTags(rsf)

	for tagName, tag := range parsedTags {
		if tag.Value == "" {
			continue
		}

		if m.disableFallback && m.tagName != tagName {
			continue
		}

		envVarName := toEnvVarName(prefixes, tag.Value)
		for key := range m.envVars {
			if strings.HasPrefix(key, envVarName) {
				return strings.ToUpper(tag.Value)
			}
		}
	}

	return strings.ToUpper(rsf.Name)
}

func (m *matcher) GetMapKeys(rsf reflect.StructField, prefixes []string) []string {
	rte := rsf.Type.Elem()
	// if its a map of primitives
	if rte.Kind() != reflect.Struct {
		return m.getPrimitiveMapKeys(rsf, prefixes)
	}

	return m.getStructMapKeys(rsf, prefixes)
}

func (m *matcher) getPrimitiveMapKeys(rsf reflect.StructField, prefixes []string) []string {
	uniqueKeys := make(map[string]struct{})

	for envVarName := range m.envVars {
		if key := m.getMapKey("", envVarName, prefixes); key != "" {
			uniqueKeys[key] = struct{}{}
		}
	}

	keys := make([]string, 0, len(uniqueKeys))
	for key := range uniqueKeys {
		keys = append(keys, key)
	}

	return keys
}

func (m *matcher) getStructMapKeys(rsf reflect.StructField, prefixes []string) []string {
	rte := rsf.Type.Elem()

	prefix := m.GetPrefix(rsf, prefixes)
	uniqueKeys := make(map[string]struct{})

	for envVarName := range m.envVars {
		if key := m.findLongestMatchingKey(rte, envVarName, append(prefixes, prefix)); key != "" {
			uniqueKeys[key] = struct{}{}
		}
	}

	keys := make([]string, 0, len(uniqueKeys))
	for key := range uniqueKeys {
		keys = append(keys, key)
	}

	return keys
}

func (m *matcher) findLongestMatchingKey(rte reflect.Type, envVarName string, prefixes []string) string {
	bestKey := ""
	longestMatch := 0

	for i := 0; i < rte.NumField(); i++ {
		field := rte.Field(i)

		parsedTags := tag.ParseTags(field)
		for tagName, tag := range parsedTags {
			if m.disableFallback && m.tagName != tagName {
				continue
			}

			mapKey := m.getMapKey(tag.Value, envVarName, prefixes)
			if mapKey != "" {
				if len(tag.Value) > longestMatch {
					longestMatch = len(tag.Value)
					bestKey = mapKey
				}
			}
		}
	}

	return bestKey
}

func (m *matcher) getValue(fieldName string, prefixes []string) (bool, string, string) {
	fieldName = strings.ToUpper(fieldName)

	envVarName := toEnvVarName(prefixes, fieldName)
	if value, ok := m.envVars[envVarName]; ok {
		return true, envVarName, value
	}

	return false, "", ""
}

func (m *matcher) getMapKey(fieldName, envVarName string, prefixes []string) string {
	fieldName = strings.ToUpper(fieldName)

	prefix := strings.ToUpper(strings.Join(prefixes, "_"))

	if !strings.HasPrefix(envVarName, prefix) ||
		!strings.HasSuffix(envVarName, fieldName) {
		return ""
	}

	return strings.ToLower(
		strings.TrimSuffix(
			strings.TrimPrefix(envVarName, fmt.Sprintf("%s_", prefix)),
			fmt.Sprintf("_%s", fieldName),
		),
	)
}

func (m *matcher) expandValue(value string) string {
	return os.Expand(value, func(s string) string { return m.envVars[s] })
}

func (m *matcher) parseOptions(tags map[string]tag.Tag) map[string]string {
	opts := map[string]string{}

	// first check for first class tags

	if tag, ok := tags[m.requiredTag]; ok {
		opts[m.requiredTag] = tag.Value
	}

	if tag, ok := tags[m.defaultTag]; ok {
		opts[m.defaultTag] = tag.Value
	}

	if tag, ok := tags[m.expandTag]; ok {
		opts[m.expandTag] = tag.Value
	}

	if tag, ok := tags[m.notEmptyTag]; ok {
		opts[m.notEmptyTag] = tag.Value
	}

	if tag, ok := tags[m.fileTag]; ok {
		opts[m.fileTag] = tag.Value
	}

	// then check for env tag options
	if tagName, ok := tags[m.tagName]; ok {
		if value, ok := tagName.Options[m.defaultTag]; ok {
			opts[m.defaultTag] = value
		}

		if value, ok := tagName.Options[m.requiredTag]; ok {
			opts[m.requiredTag] = value
		}

		if value, ok := tagName.Options[m.expandTag]; ok {
			opts[m.expandTag] = value
		}

		if value, ok := tagName.Options[m.notEmptyTag]; ok {
			opts[m.notEmptyTag] = value
		}

		if value, ok := tagName.Options[m.fileTag]; ok {
			opts[m.fileTag] = value
		}
	}

	return opts
}

func toEnvVarName(prefixes []string, tag string) string {
	if len(prefixes) == 0 {
		return strings.ToUpper(tag)
	}

	prefix := strings.Join(prefixes, "_")
	if prefix == "" {
		return strings.ToUpper(tag)
	}

	return strings.ToUpper(
		fmt.Sprintf("%s_%s", prefix, tag),
	)
}
