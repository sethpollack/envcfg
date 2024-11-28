package matcher

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/sethpollack/envcfg/internal/loader"
	"github.com/sethpollack/envcfg/internal/tag"
)

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
	loader          *loader.Loader
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

func (m *matcher) GetValue(path []tag.TagMap) (string, bool, error) {
	opts := m.parseOptions(path[len(path)-1])

	foundMatch, foundKey, foundValue := m.getValue("", path)

	if !foundMatch {
		if _, ok := opts[m.requiredTag]; ok {
			return "", false, fmt.Errorf("required field %s not found", fieldPath(path))
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

func (m *matcher) HasPrefix(path []tag.TagMap) bool {
	return m.hasPrefix("", path)
}

func (m *matcher) GetMapKeys(path []tag.TagMap) []string {
	if len(path) == 0 {
		return []string{}
	}

	current := path[len(path)-1]

	switch current.Type.Elem().Kind() {
	case reflect.Struct:
		return m.getStructMapKeys(path)
	case reflect.Slice:
		return m.getSliceMapKeys(path)
	default:
		return m.getPrimitiveMapKeys(path)
	}
}

func (m *matcher) getPrimitiveMapKeys(path []tag.TagMap) []string {
	uniqueKeys := make(map[string]struct{})

	for key := range m.envVars {
		if found, prefix := m.toPrefix(key, "", path); found {
			if key := m.getMapKey(key, prefix, ""); key != "" {
				uniqueKeys[key] = struct{}{}
			}
		}
	}

	keys := make([]string, 0, len(uniqueKeys))
	for key := range uniqueKeys {
		keys = append(keys, key)
	}

	return keys
}

func (m *matcher) getMapKey(key, prefix, suffix string) string {
	if !strings.HasPrefix(key, prefix) {
		return ""
	}

	// Get the part after prefix, removing the leading underscore
	afterPrefix := strings.TrimPrefix(key, fmt.Sprintf("%s_", prefix))

	// First try exact suffix match
	if strings.HasSuffix(afterPrefix, suffix) {
		return strings.ToLower(strings.TrimSuffix(afterPrefix, "_"+suffix))
	}

	// If no exact match, look for suffix elsewhere in the string
	if idx := strings.Index(afterPrefix, "_"+suffix+"_"); idx >= 0 {
		return strings.ToLower(afterPrefix[:idx])
	}

	return ""
}

func (m *matcher) getSliceMapKeys(path []tag.TagMap) []string {
	uniqueKeys := make(map[string]struct{})

	for i := 0; ; i++ {
		found := false
		for key := range m.envVars {
			if ok, prefix := m.toPrefix(key, "", path); ok {
				if mapKey := m.getMapKey(key, prefix, strconv.Itoa(i)); mapKey != "" {
					uniqueKeys[mapKey] = struct{}{}
					found = true
				}
			}
		}
		if !found {
			break
		}
	}

	keys := make([]string, 0, len(uniqueKeys))
	for key := range uniqueKeys {
		keys = append(keys, key)
	}

	return keys
}
func (m *matcher) getStructMapKeys(path []tag.TagMap) []string {
	uniqueKeys := make(map[string]struct{})

	for envVarName := range m.envVars {
		if found, prefix := m.toPrefix(envVarName, "", path); found {
			if key := m.findLongestMatchingKey(envVarName, prefix, path); key != "" {
				uniqueKeys[key] = struct{}{}
			}
		}
	}

	keys := make([]string, 0, len(uniqueKeys))
	for key := range uniqueKeys {
		keys = append(keys, key)
	}

	return keys
}

func (m *matcher) findLongestMatchingKey(key, prefix string, path []tag.TagMap) string {
	bestKey := ""
	longestMatch := 0

	current := path[len(path)-1]

	for i := 0; i < current.Type.Elem().NumField(); i++ {
		field := current.Type.Elem().Field(i)

		parsedTags := tag.ParseTags(field)

		if tag, ok := parsedTags.Tags[m.tagName]; ok {
			if mapKey := m.getMapKey(key, prefix, strings.ToUpper(tag.Value)); mapKey != "" {
				if len(tag.Value) > longestMatch {
					longestMatch = len(tag.Value)
					bestKey = mapKey
				}
			}
		}

		for tagName, tag := range parsedTags.Tags {
			if tag.Value == "" || m.isKnownTag(tagName) || m.disableFallback {
				continue
			}

			if mapKey := m.getMapKey(key, prefix, strings.ToUpper(tag.Value)); mapKey != "" {
				if len(tag.Value) > longestMatch {
					longestMatch = len(tag.Value)
					bestKey = mapKey
				}
			}
		}
	}

	return bestKey
}

func (m *matcher) getValue(prefix string, path []tag.TagMap) (bool, string, string) {
	if len(path) == 0 {
		envVarName := strings.ToUpper(prefix)

		if value, ok := m.envVars[envVarName]; ok {
			return true, envVarName, value
		}

		return false, "", ""
	}

	current, rest := path[0], path[1:]

	if tag, ok := current.Tags[m.tagName]; ok {
		if prefix == "" {
			if found, envvar, value := m.getValue(tag.Value, rest); found {
				return found, envvar, value
			}
		} else {
			if found, envvar, value := m.getValue(fmt.Sprint(prefix, "_", tag.Value), rest); found {
				return found, envvar, value
			}
		}
	}

	for tagName, tag := range current.Tags {
		if tag.Value == "" || m.isKnownTag(tagName) || m.disableFallback {
			continue
		}

		if prefix == "" {
			if found, envvar, value := m.getValue(tag.Value, rest); found {
				return found, envvar, value
			}
		} else {
			if found, envvar, value := m.getValue(fmt.Sprint(prefix, "_", tag.Value), rest); found {
				return found, envvar, value
			}
		}
	}

	return false, "", ""
}

func (m *matcher) hasPrefix(prefix string, path []tag.TagMap) bool {
	if len(path) == 0 {
		envVarName := strings.ToUpper(prefix)

		for env := range m.envVars {
			if strings.HasPrefix(env, envVarName) {
				return true
			}
		}

		return false
	}

	current, rest := path[0], path[1:]

	if tag, ok := current.Tags[m.tagName]; ok {
		if prefix == "" {
			if found := m.hasPrefix(tag.Value, rest); found {
				return found
			}
		} else {
			if found := m.hasPrefix(fmt.Sprint(prefix, "_", tag.Value), rest); found {
				return found
			}
		}
	}

	for tagName, tag := range current.Tags {
		if tag.Value == "" || m.isKnownTag(tagName) {
			continue
		}

		if prefix == "" {
			if found := m.hasPrefix(tag.Value, rest); found {
				return found
			}
		} else {
			if found := m.hasPrefix(fmt.Sprint(prefix, "_", tag.Value), rest); found {
				return found
			}
		}
	}

	return false
}

func (m *matcher) toPrefix(key, prefix string, path []tag.TagMap) (bool, string) {
	if len(path) == 0 {
		envVarPrefix := strings.ToUpper(prefix)
		if strings.HasPrefix(key, envVarPrefix) {
			return true, envVarPrefix
		}

		return false, ""
	}

	current, rest := path[0], path[1:]

	if tag, ok := current.Tags[m.tagName]; ok {
		var newPrefix string
		if prefix == "" {
			newPrefix = tag.Value
		} else {
			newPrefix = fmt.Sprint(prefix, "_", tag.Value)
		}

		if found, match := m.toPrefix(key, newPrefix, rest); found {
			return found, match
		}
	}

	for tagName, tag := range current.Tags {
		if tag.Value == "" || m.isKnownTag(tagName) {
			continue
		}

		var newPrefix string
		if prefix == "" {
			newPrefix = tag.Value
		} else {
			newPrefix = fmt.Sprint(prefix, "_", tag.Value)
		}

		if found, match := m.toPrefix(key, newPrefix, rest); found {
			return found, match
		}
	}

	return false, ""
}

func (m *matcher) expandValue(value string) string {
	return os.Expand(value, func(s string) string { return m.envVars[s] })
}

func (m *matcher) parseOptions(tm tag.TagMap) map[string]string {
	opts := map[string]string{}

	// first check for first class tags

	if tag, ok := tm.Tags[m.requiredTag]; ok {
		opts[m.requiredTag] = tag.Value
	}

	if tag, ok := tm.Tags[m.defaultTag]; ok {
		opts[m.defaultTag] = tag.Value
	}

	if tag, ok := tm.Tags[m.expandTag]; ok {
		opts[m.expandTag] = tag.Value
	}

	if tag, ok := tm.Tags[m.notEmptyTag]; ok {
		opts[m.notEmptyTag] = tag.Value
	}

	if tag, ok := tm.Tags[m.fileTag]; ok {
		opts[m.fileTag] = tag.Value
	}

	// then check for env tag options
	if tagName, ok := tm.Tags[m.tagName]; ok {
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

func (m *matcher) isKnownTag(tagName string) bool {
	tags := map[string]bool{
		m.tagName:     true,
		m.requiredTag: true,
		m.defaultTag:  true,
		m.expandTag:   true,
		m.notEmptyTag: true,
		m.fileTag:     true,
	}

	_, ok := tags[tagName]
	return ok
}

func fieldPath(path []tag.TagMap) string {
	prefix := path[0].FieldName

	for _, tm := range path[1:] {
		prefix += fmt.Sprintf(".%s", tm.FieldName)
	}

	return prefix
}
