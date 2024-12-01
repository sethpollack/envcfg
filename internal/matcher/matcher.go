package matcher

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/sethpollack/envcfg/internal/tag"
)

type Matcher struct {
	// tags
	TagName     string
	DefaultTag  string
	ExpandTag   string
	FileTag     string
	NotEmptyTag string
	RequiredTag string
	// default options
	Expand          bool
	Required        bool
	NotEmpty        bool
	DisableFallback bool

	EnvVars map[string]string
}

func New() *Matcher {
	return &Matcher{
		TagName:     "env",
		DefaultTag:  "default",
		ExpandTag:   "expand",
		FileTag:     "file",
		NotEmptyTag: "notempty",
		RequiredTag: "required",
		EnvVars:     map[string]string{},
	}
}

func (m *Matcher) GetValue(path []tag.TagMap) (string, bool, error) {
	opts := m.parseOptions(path[len(path)-1])

	foundMatch, foundKey, foundValue := m.getValue("", path)

	if !foundMatch {
		if _, ok := opts[m.RequiredTag]; ok {
			return "", false, fmt.Errorf("required field %s not found", fieldPath(path))
		}

		if _, ok := opts[m.DefaultTag]; ok {
			if _, ok := opts[m.ExpandTag]; ok {
				return m.expandValue(opts[m.DefaultTag]), true, nil
			}
			return opts[m.DefaultTag], true, nil
		}

		return "", false, nil
	}

	if _, ok := opts[m.NotEmptyTag]; ok && foundValue == "" {
		return "", true, fmt.Errorf("environment variable %s is empty", foundKey)
	}

	if _, ok := opts[m.FileTag]; ok {
		bytes, err := os.ReadFile(foundValue)
		if err != nil {
			return "", true, err
		}

		if _, ok := opts[m.ExpandTag]; ok {
			return m.expandValue(string(bytes)), true, nil
		}

		return string(bytes), true, nil
	}

	if _, ok := opts[m.ExpandTag]; ok {
		return m.expandValue(foundValue), true, nil
	}

	return foundValue, true, nil
}

func (m *Matcher) HasPrefix(path []tag.TagMap) bool {
	return m.hasPrefix("", path)
}

func (m *Matcher) GetMapKeys(path []tag.TagMap) []string {
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

func (m *Matcher) getPrimitiveMapKeys(path []tag.TagMap) []string {
	uniqueKeys := make(map[string]struct{})

	for key := range m.EnvVars {
		if found, prefix := m.toPrefix(key, "", path); found {
			if key := parseMapKey(key, prefix, ""); key != "" {
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

func (m *Matcher) getSliceMapKeys(path []tag.TagMap) []string {
	uniqueKeys := make(map[string]struct{})

	for i := 0; ; i++ {
		found := false
		for key := range m.EnvVars {
			if ok, prefix := m.toPrefix(key, "", path); ok {
				if mapKey := parseMapKey(key, prefix, strconv.Itoa(i)); mapKey != "" {
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
func (m *Matcher) getStructMapKeys(path []tag.TagMap) []string {
	uniqueKeys := make(map[string]struct{})

	for envVarName := range m.EnvVars {
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

func (m *Matcher) findLongestMatchingKey(key, prefix string, path []tag.TagMap) string {
	bestKey := ""
	longestMatch := 0

	current := path[len(path)-1]

	for i := 0; i < current.Type.Elem().NumField(); i++ {
		field := current.Type.Elem().Field(i)

		parsedTags := tag.ParseTags(field)

		if tag, ok := parsedTags.Tags[m.TagName]; ok {
			if mapKey := parseMapKey(key, prefix, strings.ToUpper(tag.Value)); mapKey != "" {
				if len(tag.Value) > longestMatch {
					longestMatch = len(tag.Value)
					bestKey = mapKey
				}
			}
		}

		for tagName, tag := range parsedTags.Tags {
			if tag.Value == "" || m.isKnownTag(tagName) || m.DisableFallback {
				continue
			}

			if mapKey := parseMapKey(key, prefix, strings.ToUpper(tag.Value)); mapKey != "" {
				if len(tag.Value) > longestMatch {
					longestMatch = len(tag.Value)
					bestKey = mapKey
				}
			}
		}
	}

	return bestKey
}

func (m *Matcher) getValue(prefix string, path []tag.TagMap) (bool, string, string) {
	if len(path) == 0 {
		envVarName := strings.ToUpper(prefix)

		if value, ok := m.EnvVars[envVarName]; ok {
			return true, envVarName, value
		}

		return false, "", ""
	}

	current, rest := path[0], path[1:]

	if tag, ok := current.Tags[m.TagName]; ok {
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
		if tag.Value == "" || m.isKnownTag(tagName) || m.DisableFallback {
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

func (m *Matcher) hasPrefix(prefix string, path []tag.TagMap) bool {
	if len(path) == 0 {
		envVarName := strings.ToUpper(prefix)

		for env := range m.EnvVars {
			if strings.HasPrefix(env, envVarName) {
				return true
			}
		}

		return false
	}

	current, rest := path[0], path[1:]

	if tag, ok := current.Tags[m.TagName]; ok {
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

func (m *Matcher) toPrefix(key, prefix string, path []tag.TagMap) (bool, string) {
	if len(path) == 0 {
		envVarPrefix := strings.ToUpper(prefix)
		if strings.HasPrefix(key, envVarPrefix) {
			return true, envVarPrefix
		}

		return false, ""
	}

	current, rest := path[0], path[1:]

	if tag, ok := current.Tags[m.TagName]; ok {
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

func (m *Matcher) expandValue(value string) string {
	return os.Expand(value, func(s string) string { return m.EnvVars[s] })
}

func (m *Matcher) parseOptions(tm tag.TagMap) map[string]string {
	opts := map[string]string{}

	if m.Expand {
		opts[m.ExpandTag] = "true"
	}

	if m.Required {
		opts[m.RequiredTag] = "true"
	}

	if m.NotEmpty {
		opts[m.NotEmptyTag] = "true"
	}

	if tag, ok := tm.Tags[m.RequiredTag]; ok {
		opts[m.RequiredTag] = tag.Value
	}

	if tag, ok := tm.Tags[m.DefaultTag]; ok {
		opts[m.DefaultTag] = tag.Value
	}

	if tag, ok := tm.Tags[m.ExpandTag]; ok {
		opts[m.ExpandTag] = tag.Value
	}

	if tag, ok := tm.Tags[m.NotEmptyTag]; ok {
		opts[m.NotEmptyTag] = tag.Value
	}

	if tag, ok := tm.Tags[m.FileTag]; ok {
		opts[m.FileTag] = tag.Value
	}

	// then check for env tag options
	if tagName, ok := tm.Tags[m.TagName]; ok {
		if value, ok := tagName.Options[m.DefaultTag]; ok {
			opts[m.DefaultTag] = value
		}

		if value, ok := tagName.Options[m.RequiredTag]; ok {
			opts[m.RequiredTag] = value
		}

		if value, ok := tagName.Options[m.ExpandTag]; ok {
			opts[m.ExpandTag] = value
		}

		if value, ok := tagName.Options[m.NotEmptyTag]; ok {
			opts[m.NotEmptyTag] = value
		}

		if value, ok := tagName.Options[m.FileTag]; ok {
			opts[m.FileTag] = value
		}
	}

	return opts
}

func (m *Matcher) isKnownTag(tagName string) bool {
	tags := map[string]bool{
		m.TagName:     true,
		m.RequiredTag: true,
		m.DefaultTag:  true,
		m.ExpandTag:   true,
		m.NotEmptyTag: true,
		m.FileTag:     true,
	}

	_, ok := tags[tagName]
	return ok
}

func parseMapKey(key, prefix, suffix string) string {
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

func fieldPath(path []tag.TagMap) string {
	prefix := path[0].FieldName

	for _, tm := range path[1:] {
		prefix += fmt.Sprintf(".%s", tm.FieldName)
	}

	return prefix
}
