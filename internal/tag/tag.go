package tag

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type Tag struct {
	Name    string
	Value   string
	Options map[string]string
}

type TagMap struct {
	FieldName string
	Type      reflect.Type
	Tags      map[string]Tag
}

func ParseTags(rfs reflect.StructField) TagMap {
	rft := rfs.Tag

	tm := TagMap{
		FieldName: rfs.Name,
		Type:      rfs.Type,
		Tags:      map[string]Tag{},
	}
	// otherwise parse all tags

	for rft != "" {
		// Skip leading space.
		i := 0
		for i < len(rft) && rft[i] == ' ' {
			i++
		}
		rft = rft[i:]
		if rft == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		i = 0
		for i < len(rft) && rft[i] > ' ' && rft[i] != ':' && rft[i] != '"' && rft[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(rft) || rft[i] != ':' || rft[i+1] != '"' {
			break
		}
		name := string(rft[:i])
		rft = rft[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(rft) && rft[i] != '"' {
			if rft[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(rft) {
			break
		}
		qvalue := string(rft[:i+1])
		rft = rft[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err == nil {
			value, options := parseTag(value)

			tm.Tags[name] = Tag{
				Name:    name,
				Value:   value,
				Options: options,
			}
		}
	}

	tm.Tags["struct"] = Tag{
		Name:    "struct",
		Value:   rfs.Name,
		Options: map[string]string{},
	}

	tm.Tags["struct_snake"] = Tag{
		Name:    "struct_snake",
		Value:   toSnakeCase(rfs.Name),
		Options: map[string]string{},
	}

	return tm
}

func parseTag(tag string) (string, map[string]string) {
	parts := strings.Split(tag, ",")

	options := make(map[string]string)
	for _, part := range parts[1:] {
		key, value := parseTagOption(part)
		options[key] = value
	}

	return parts[0], options
}

func parseTagOption(option string) (string, string) {
	if !strings.Contains(option, "=") {
		return option, ""
	}

	parts := strings.SplitN(option, "=", 2)
	return parts[0], parts[1]
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			// Check if it's not the start of the string and the previous character is lowercase
			// Or if it's not the last character and the next character is lowercase
			if (i > 0 && unicode.IsLower(rune(s[i-1]))) || (i < len(s)-1 && unicode.IsLower(rune(s[i+1]))) {
				result.WriteRune('_')
			}
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}
