package walker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkString(t *testing.T) {
	type Config struct {
		Simple    string
		StringPtr *string
	}

	str := "hello"

	t.Setenv("SIMPLE", str)
	t.Setenv("STRING_PTR", str)
	cfg := Config{}

	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{
		Simple:    str,
		StringPtr: &str,
	}, cfg)
}

func TestWalkSlice(t *testing.T) {
	type Config struct {
		Slice []string
	}

	t.Setenv("SLICE", "a,b,c")

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{
		Slice: []string{"a", "b", "c"},
	}, cfg)
}

func TestWalkIndexSlice(t *testing.T) {
	type Config struct {
		Slice []string
	}

	t.Setenv("SLICE_0", "a")
	t.Setenv("SLICE_1", "b")
	t.Setenv("SLICE_2", "c")

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{
		Slice: []string{"a", "b", "c"},
	}, cfg)
}

func TestWalkSliceOfStructs(t *testing.T) {
	type Inner struct {
		Value string
	}

	type Config struct {
		Slice []Inner
	}

	t.Setenv("SLICE_0_VALUE", "a")
	t.Setenv("SLICE_1_VALUE", "b")
	t.Setenv("SLICE_2_VALUE", "c")

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{
		Slice: []Inner{{Value: "a"}, {Value: "b"}, {Value: "c"}},
	}, cfg)
}

func TestWalkMap(t *testing.T) {
	type Config struct {
		MyMap map[string]string
	}

	t.Setenv("MY_MAP", "a:a,b:b,c:c")

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{
		MyMap: map[string]string{"a": "a", "b": "b", "c": "c"},
	}, cfg)
}

func TestWalkMapWithInvalidDelim(t *testing.T) {
	type Config struct {
		MyMap map[string]string
	}

	t.Setenv("MY_MAP", "a:a,b:b,c")

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.Error(t, err)
}

func TestWalkNamedMap(t *testing.T) {
	type Config struct {
		MyMap map[string]string
	}

	t.Setenv("MY_MAP_A", "a")
	t.Setenv("MY_MAP_B", "b")
	t.Setenv("MY_MAP_C", "c")

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{
		MyMap: map[string]string{"a": "a", "b": "b", "c": "c"},
	}, cfg)
}

func TestWalkMapOfStructs(t *testing.T) {
	type Inner struct {
		Key      string
		CamelKey string
		SnakeKey string
		TagKey   string `env:"CUSTOM_KEY"`
	}

	type Config struct {
		MyMap map[string]Inner
	}

	t.Setenv("MY_MAP_A_KEY", "a_key")
	t.Setenv("MY_MAP_A_CAMELKEY", "a_camel_key")
	t.Setenv("MY_MAP_A_SNAKE_KEY", "a_snake_key")
	t.Setenv("MY_MAP_A_CUSTOM_KEY", "a_custom_key")

	t.Setenv("MY_MAP_B_KEY", "b_key")
	t.Setenv("MY_MAP_B_CAMELKEY", "b_camel_key")
	t.Setenv("MY_MAP_B_SNAKE_KEY", "b_snake_key")
	t.Setenv("MY_MAP_B_CUSTOM_KEY", "b_custom_key")

	t.Setenv("MY_MAP_C_KEY", "c_key")
	t.Setenv("MY_MAP_C_CAMELKEY", "c_camel_key")
	t.Setenv("MY_MAP_C_SNAKE_KEY", "c_snake_key")
	t.Setenv("MY_MAP_C_CUSTOM_KEY", "c_custom_key")

	t.Setenv("MY_MAP_D_D_KEY", "d_key")
	t.Setenv("MY_MAP_D_D_CAMELKEY", "d_camel_key")
	t.Setenv("MY_MAP_D_D_SNAKE_KEY", "d_snake_key")
	t.Setenv("MY_MAP_D_D_CUSTOM_KEY", "d_custom_key")

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{
		MyMap: map[string]Inner{
			"a":   {Key: "a_key", CamelKey: "a_camel_key", SnakeKey: "a_snake_key", TagKey: "a_custom_key"},
			"b":   {Key: "b_key", CamelKey: "b_camel_key", SnakeKey: "b_snake_key", TagKey: "b_custom_key"},
			"c":   {Key: "c_key", CamelKey: "c_camel_key", SnakeKey: "c_snake_key", TagKey: "c_custom_key"},
			"d_d": {Key: "d_key", CamelKey: "d_camel_key", SnakeKey: "d_snake_key", TagKey: "d_custom_key"},
		},
	}, cfg)
}

func TestParseEmptyPtr(t *testing.T) {
	type Inner struct {
		String string
	}

	type Config struct {
		Config *Inner
		Empty  *Inner
	}

	t.Setenv("CONFIG_STRING", "hello")

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{
		Config: &Inner{String: "hello"},
		Empty:  nil,
	}, cfg)
}

func TestWalkNilPtr(t *testing.T) {
	type Config struct {
		Ptr *string
	}

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)
}

func TestWalkNilPtrWithInitMode(t *testing.T) {
	type Inner struct {
		Value string
	}

	type Config struct {
		Never  *Inner `env:",init=never" init:"never"`
		Always *Inner `env:",init=always" init:"always"`
		Values *Inner `env:",init=values" init:"values"`
	}

	t.Setenv("NEVER_VALUE", "never")
	t.Setenv("VALUES_VALUE", "values")

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{
		Never:  nil,
		Always: &Inner{},
		Values: &Inner{Value: "values"},
	}, cfg)
}

func TestWalkWithOptions(t *testing.T) {
	type Config struct {
		Value string
	}

	t.Setenv("VALUE", "value")

	cfg := Config{}
	w := New()

	if err := w.Build(
		WithTagName("env"),
		WithDelimiter(","),
		WithSeparator(":"),
		WithIgnoreTag("ignore"),
		WithInitNever(),
		WithInitAlways(),
	); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{
		Value: "value",
	}, cfg)
}

func TestParseIgnore(t *testing.T) {
	type Config struct {
		IgnoreTag    string `ignore:"true"`
		IgnoreOption string `env:",ignore"`
		IgnoreValue  string `env:"-"`
	}

	t.Setenv("IGNORE_TAG", "ignore_tag")
	t.Setenv("IGNORE_OPTION", "ignore_option")
	t.Setenv("IGNORE_VALUE", "ignore_value")

	cfg := Config{}
	w := New()

	if err := w.Build(); err != nil {
		assert.NoError(t, err)
	}

	err := w.Walk(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, Config{}, cfg)
}
