package envcfg_test

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/sethpollack/envcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseString(t *testing.T) {
	type Config struct {
		CamelCase    string
		CamelCasePtr *string
		SnakeCase    string
		SnakeCasePtr *string
		Tagged       string  `env:"CUSTOM"`
		TaggedPtr    *string `env:"CUSTOM_PTR"`
	}

	t.Setenv("CAMELCASE", "camel case")
	t.Setenv("CAMELCASEPTR", "camel case ptr")
	t.Setenv("SNAKE_CASE", "snake case")
	t.Setenv("SNAKE_CASE_PTR", "snake case ptr")
	t.Setenv("CUSTOM", "custom")
	t.Setenv("CUSTOM_PTR", "custom ptr")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		CamelCase:    "camel case",
		CamelCasePtr: strPtr("camel case ptr"),
		SnakeCase:    "snake case",
		SnakeCasePtr: strPtr("snake case ptr"),
		Tagged:       "custom",
		TaggedPtr:    strPtr("custom ptr"),
	}, cfg)
}

func TestParseBool(t *testing.T) {
	type Config struct {
		CamelCase    bool
		CamelCasePtr *bool
		SnakeCase    bool
		SnakeCasePtr *bool
		Tagged       bool  `env:"CUSTOM"`
		TaggedPtr    *bool `env:"CUSTOM_PTR"`
	}

	trueVal := true

	t.Setenv("SIMPLE", "true")
	t.Setenv("SIMPLE_PTR", "true")
	t.Setenv("CAMEL_CASE", "true")
	t.Setenv("CAMEL_CASE_PTR", "true")
	t.Setenv("SNAKE_CASE", "true")
	t.Setenv("SNAKE_CASE_PTR", "true")
	t.Setenv("CUSTOM", "true")
	t.Setenv("CUSTOM_PTR", "true")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		CamelCase:    trueVal,
		CamelCasePtr: &trueVal,
		SnakeCase:    trueVal,
		SnakeCasePtr: &trueVal,
		Tagged:       trueVal,
		TaggedPtr:    &trueVal,
	}, cfg)
}

func TestParseInt(t *testing.T) {
	type Config struct {
		Int       int
		IntPtr    *int
		Int8      int8
		Int8Ptr   *int8
		Int16     int16
		Int16Ptr  *int16
		Int32     int32
		Int32Ptr  *int32
		Int64     int64
		Int64Ptr  *int64
		Uint      uint
		UintPtr   *uint
		Uint8     uint8
		Uint8Ptr  *uint8
		Uint16    uint16
		Uint16Ptr *uint16
		Uint32    uint32
		Uint32Ptr *uint32
		Uint64    uint64
		Uint64Ptr *uint64
	}

	i := 123
	i8 := int8(123)
	i16 := int16(123)
	i32 := int32(123)
	i64 := int64(123)
	u := uint(123)
	u8 := uint8(123)
	u16 := uint16(123)
	u32 := uint32(123)
	u64 := uint64(123)

	t.Setenv("INT", fmt.Sprintf("%d", i))
	t.Setenv("INT_PTR", fmt.Sprintf("%d", i))
	t.Setenv("INT8", fmt.Sprintf("%d", i8))
	t.Setenv("INT8_PTR", fmt.Sprintf("%d", i8))
	t.Setenv("INT16", fmt.Sprintf("%d", i16))
	t.Setenv("INT16_PTR", fmt.Sprintf("%d", i16))
	t.Setenv("INT32", fmt.Sprintf("%d", i32))
	t.Setenv("INT32_PTR", fmt.Sprintf("%d", i32))
	t.Setenv("INT64", fmt.Sprintf("%d", i64))
	t.Setenv("INT64_PTR", fmt.Sprintf("%d", i64))
	t.Setenv("UINT", fmt.Sprintf("%d", u))
	t.Setenv("UINT_PTR", fmt.Sprintf("%d", u))
	t.Setenv("UINT8", fmt.Sprintf("%d", u8))
	t.Setenv("UINT8_PTR", fmt.Sprintf("%d", u8))
	t.Setenv("UINT16", fmt.Sprintf("%d", u16))
	t.Setenv("UINT16_PTR", fmt.Sprintf("%d", u16))
	t.Setenv("UINT32", fmt.Sprintf("%d", u32))
	t.Setenv("UINT32_PTR", fmt.Sprintf("%d", u32))
	t.Setenv("UINT64", fmt.Sprintf("%d", u64))
	t.Setenv("UINT64_PTR", fmt.Sprintf("%d", u64))

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		Int:       123,
		IntPtr:    &i,
		Int8:      123,
		Int8Ptr:   &i8,
		Int16:     123,
		Int16Ptr:  &i16,
		Int32:     123,
		Int32Ptr:  &i32,
		Int64:     123,
		Int64Ptr:  &i64,
		Uint:      123,
		UintPtr:   &u,
		Uint8:     123,
		Uint8Ptr:  &u8,
		Uint16:    123,
		Uint16Ptr: &u16,
		Uint32:    123,
		Uint32Ptr: &u32,
		Uint64:    123,
		Uint64Ptr: &u64,
	}, cfg)
}

func TestParseUInt(t *testing.T) {
	type Config struct {
		Uint      uint
		UintPtr   *uint
		Uint8     uint8
		Uint8Ptr  *uint8
		Uint16    uint16
		Uint16Ptr *uint16
		Uint32    uint32
		Uint32Ptr *uint32
		Uint64    uint64
		Uint64Ptr *uint64
	}

	u := uint(123)
	u8 := uint8(123)
	u16 := uint16(123)
	u32 := uint32(123)
	u64 := uint64(123)

	t.Setenv("UINT", fmt.Sprintf("%d", u))
	t.Setenv("UINT_PTR", fmt.Sprintf("%d", u))
	t.Setenv("UINT8", fmt.Sprintf("%d", u8))
	t.Setenv("UINT8_PTR", fmt.Sprintf("%d", u8))
	t.Setenv("UINT16", fmt.Sprintf("%d", u16))
	t.Setenv("UINT16_PTR", fmt.Sprintf("%d", u16))
	t.Setenv("UINT32", fmt.Sprintf("%d", u32))
	t.Setenv("UINT32_PTR", fmt.Sprintf("%d", u32))
	t.Setenv("UINT64", fmt.Sprintf("%d", u64))
	t.Setenv("UINT64_PTR", fmt.Sprintf("%d", u64))

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		Uint:      123,
		UintPtr:   &u,
		Uint8:     123,
		Uint8Ptr:  &u8,
		Uint16:    123,
		Uint16Ptr: &u16,
		Uint32:    123,
		Uint32Ptr: &u32,
		Uint64:    123,
		Uint64Ptr: &u64,
	}, cfg)
}

func TestParseFloat(t *testing.T) {
	type Config struct {
		Float32    float32
		Float32Ptr *float32
		Float64    float64
		Float64Ptr *float64
	}

	f32 := float32(123)
	f64 := float64(123)

	t.Setenv("FLOAT32", fmt.Sprintf("%f", f32))
	t.Setenv("FLOAT32_PTR", fmt.Sprintf("%f", f32))
	t.Setenv("FLOAT64", fmt.Sprintf("%f", f64))
	t.Setenv("FLOAT64_PTR", fmt.Sprintf("%f", f64))

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		Float32:    f32,
		Float32Ptr: &f32,
		Float64:    f64,
		Float64Ptr: &f64,
	}, cfg)
}

func TestParseSlice(t *testing.T) {
	type Config struct {
		Slice         []string
		PtrSlice      []*string
		IndexSlice    []string
		PtrIndexSlice []*string

		IntSlice         []int
		PtrIntSlice      []*int
		IndexIntSlice    []int
		PtrIndexIntSlice []*int

		FloatSlice         []float64
		PtrFloatSlice      []*float64
		IndexFloatSlice    []float64
		PtrIndexFloatSlice []*float64

		BoolSlice         []bool
		PtrBoolSlice      []*bool
		IndexBoolSlice    []bool
		PtrIndexBoolSlice []*bool

		DurationSlice         []time.Duration
		PtrDurationSlice      []*time.Duration
		IndexDurationSlice    []time.Duration
		PtrIndexDurationSlice []*time.Duration

		URLSlice         []url.URL
		PtrURLSlice      []*url.URL
		IndexURLSlice    []url.URL
		PtrIndexURLSlice []*url.URL
	}

	t.Setenv("SLICE", "a,b,c")
	t.Setenv("PTR_SLICE_0", "a")
	t.Setenv("PTR_SLICE_1", "b")
	t.Setenv("PTR_SLICE_2", "c")
	t.Setenv("INDEX_SLICE_0", "a")
	t.Setenv("INDEX_SLICE_1", "b")
	t.Setenv("INDEX_SLICE_2", "c")
	t.Setenv("PTR_INDEX_SLICE_0", "a")
	t.Setenv("PTR_INDEX_SLICE_1", "b")
	t.Setenv("PTR_INDEX_SLICE_2", "c")

	t.Setenv("INT_SLICE", "1,2,3")
	t.Setenv("PTR_INT_SLICE", "1,2,3")
	t.Setenv("INDEX_INT_SLICE_0", "1")
	t.Setenv("INDEX_INT_SLICE_1", "2")
	t.Setenv("INDEX_INT_SLICE_2", "3")
	t.Setenv("PTR_INDEX_INT_SLICE_0", "1")
	t.Setenv("PTR_INDEX_INT_SLICE_1", "2")
	t.Setenv("PTR_INDEX_INT_SLICE_2", "3")

	t.Setenv("FLOAT_SLICE", "1.1,2.2,3.3")
	t.Setenv("PTR_FLOAT_SLICE", "1.1,2.2,3.3")
	t.Setenv("INDEX_FLOAT_SLICE_0", "1.1")
	t.Setenv("INDEX_FLOAT_SLICE_1", "2.2")
	t.Setenv("INDEX_FLOAT_SLICE_2", "3.3")
	t.Setenv("PTR_INDEX_FLOAT_SLICE_0", "1.1")
	t.Setenv("PTR_INDEX_FLOAT_SLICE_1", "2.2")
	t.Setenv("PTR_INDEX_FLOAT_SLICE_2", "3.3")

	t.Setenv("BOOL_SLICE", "true,false,true")
	t.Setenv("PTR_BOOL_SLICE", "true,false,true")
	t.Setenv("INDEX_BOOL_SLICE_0", "true")
	t.Setenv("INDEX_BOOL_SLICE_1", "false")
	t.Setenv("INDEX_BOOL_SLICE_2", "true")
	t.Setenv("PTR_INDEX_BOOL_SLICE_0", "true")
	t.Setenv("PTR_INDEX_BOOL_SLICE_1", "false")
	t.Setenv("PTR_INDEX_BOOL_SLICE_2", "true")

	t.Setenv("DURATION_SLICE", "1s,2s,3s")
	t.Setenv("PTR_DURATION_SLICE", "1s,2s,3s")
	t.Setenv("INDEX_DURATION_SLICE_0", "1s")
	t.Setenv("INDEX_DURATION_SLICE_1", "2s")
	t.Setenv("INDEX_DURATION_SLICE_2", "3s")
	t.Setenv("PTR_INDEX_DURATION_SLICE_0", "1s")
	t.Setenv("PTR_INDEX_DURATION_SLICE_1", "2s")
	t.Setenv("PTR_INDEX_DURATION_SLICE_2", "3s")

	t.Setenv("URL_SLICE", "http://example.com,https://example.com,http://example.com")
	t.Setenv("PTR_URL_SLICE", "http://example.com,https://example.com,http://example.com")
	t.Setenv("INDEX_URL_SLICE_0", "http://example.com")
	t.Setenv("INDEX_URL_SLICE_1", "https://example.com")
	t.Setenv("INDEX_URL_SLICE_2", "http://example.com")
	t.Setenv("PTR_INDEX_URL_SLICE_0", "http://example.com")
	t.Setenv("PTR_INDEX_URL_SLICE_1", "https://example.com")
	t.Setenv("PTR_INDEX_URL_SLICE_2", "http://example.com")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		Slice:                 []string{"a", "b", "c"},
		PtrSlice:              []*string{strPtr("a"), strPtr("b"), strPtr("c")},
		IndexSlice:            []string{"a", "b", "c"},
		PtrIndexSlice:         []*string{strPtr("a"), strPtr("b"), strPtr("c")},
		IntSlice:              []int{1, 2, 3},
		PtrIntSlice:           []*int{intPtr(1), intPtr(2), intPtr(3)},
		IndexIntSlice:         []int{1, 2, 3},
		PtrIndexIntSlice:      []*int{intPtr(1), intPtr(2), intPtr(3)},
		FloatSlice:            []float64{1.1, 2.2, 3.3},
		PtrFloatSlice:         []*float64{float64Ptr(1.1), float64Ptr(2.2), float64Ptr(3.3)},
		IndexFloatSlice:       []float64{1.1, 2.2, 3.3},
		PtrIndexFloatSlice:    []*float64{float64Ptr(1.1), float64Ptr(2.2), float64Ptr(3.3)},
		BoolSlice:             []bool{true, false, true},
		PtrBoolSlice:          []*bool{boolPtr(true), boolPtr(false), boolPtr(true)},
		IndexBoolSlice:        []bool{true, false, true},
		PtrIndexBoolSlice:     []*bool{boolPtr(true), boolPtr(false), boolPtr(true)},
		DurationSlice:         []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
		PtrDurationSlice:      []*time.Duration{durationPtr(1 * time.Second), durationPtr(2 * time.Second), durationPtr(3 * time.Second)},
		IndexDurationSlice:    []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
		PtrIndexDurationSlice: []*time.Duration{durationPtr(1 * time.Second), durationPtr(2 * time.Second), durationPtr(3 * time.Second)},
		URLSlice: []url.URL{
			{Scheme: "http", Host: "example.com"},
			{Scheme: "https", Host: "example.com"},
			{Scheme: "http", Host: "example.com"},
		},
		PtrURLSlice: []*url.URL{
			{Scheme: "http", Host: "example.com"},
			{Scheme: "https", Host: "example.com"},
			{Scheme: "http", Host: "example.com"},
		},
		IndexURLSlice: []url.URL{
			{Scheme: "http", Host: "example.com"},
			{Scheme: "https", Host: "example.com"},
			{Scheme: "http", Host: "example.com"},
		},
		PtrIndexURLSlice: []*url.URL{
			{Scheme: "http", Host: "example.com"},
			{Scheme: "https", Host: "example.com"},
			{Scheme: "http", Host: "example.com"},
		},
	}, cfg)
}

func TestParseSliceOfStructs(t *testing.T) {
	type Inner struct {
		Value      string
		CamelValue string
		TagValue   string `env:"CUSTOM_VALUE"`
	}

	type Config struct {
		Slice []Inner
	}

	t.Setenv("SLICE_0_VALUE", "a")
	t.Setenv("SLICE_0_CAMEL_VALUE", "a")
	t.Setenv("SLICE_0_CUSTOM_VALUE", "a")

	t.Setenv("SLICE_1_VALUE", "b")
	t.Setenv("SLICE_1_CAMEL_VALUE", "b")
	t.Setenv("SLICE_1_CUSTOM_VALUE", "b")

	t.Setenv("SLICE_2_VALUE", "c")
	t.Setenv("SLICE_2_CAMEL_VALUE", "c")
	t.Setenv("SLICE_2_CUSTOM_VALUE", "c")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{Slice: []Inner{
		{Value: "a", CamelValue: "a", TagValue: "a"},
		{Value: "b", CamelValue: "b", TagValue: "b"},
		{Value: "c", CamelValue: "c", TagValue: "c"},
	}}, cfg)
}

func TestParseMap(t *testing.T) {
	type Config struct {
		StringMap map[string]string
		IntMap    map[int]int
	}

	t.Setenv("STRING_MAP", "a:1,b:2,c:3")
	t.Setenv("INT_MAP", "1:1,2:2,3:3")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		StringMap: map[string]string{"a": "1", "b": "2", "c": "3"},
		IntMap:    map[int]int{1: 1, 2: 2, 3: 3},
	}, cfg)
}

func TestMapOfStructs(t *testing.T) {
	type Inner struct {
		Key      string
		CamelKey string
		SnakeKey string
		TagKey   string `env:"CUSTOM_KEY"`
	}

	type Config struct {
		Map map[string]Inner
	}

	t.Setenv("MAP_A_KEY", "a")
	t.Setenv("MAP_A_CAMELKEY", "a")
	t.Setenv("MAP_A_SNAKE_KEY", "a")
	t.Setenv("MAP_A_CUSTOM_KEY", "a")

	t.Setenv("MAP_B_KEY", "b")
	t.Setenv("MAP_B_CAMELKEY", "b")
	t.Setenv("MAP_B_SNAKE_KEY", "b")
	t.Setenv("MAP_B_CUSTOM_KEY", "b")

	t.Setenv("MAP_C_KEY", "c")
	t.Setenv("MAP_C_CAMELKEY", "c")
	t.Setenv("MAP_C_SNAKE_KEY", "c")
	t.Setenv("MAP_C_CUSTOM_KEY", "c")

	t.Setenv("MAP_D_D_KEY", "d")
	t.Setenv("MAP_D_D_CAMELKEY", "d")
	t.Setenv("MAP_D_D_SNAKE_KEY", "d")
	t.Setenv("MAP_D_D_CUSTOM_KEY", "d")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		Map: map[string]Inner{
			"a":   {Key: "a", CamelKey: "a", SnakeKey: "a", TagKey: "a"},
			"b":   {Key: "b", CamelKey: "b", SnakeKey: "b", TagKey: "b"},
			"c":   {Key: "c", CamelKey: "c", SnakeKey: "c", TagKey: "c"},
			"d_d": {Key: "d", CamelKey: "d", SnakeKey: "d", TagKey: "d"},
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
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		Config: &Inner{String: "hello"},
		Empty:  nil,
	}, cfg)
}

func TestParseNested(t *testing.T) {
	type Inner struct {
		String      string
		CamelString string
		TagString   string `env:"CUSTOM_STRING"`
	}

	type Config struct {
		Inner Inner
	}

	t.Setenv("INNER_STRING", "hello")
	t.Setenv("INNER_CAMEL_STRING", "hello")
	t.Setenv("INNER_CUSTOM_STRING", "hello")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{Inner: Inner{String: "hello", CamelString: "hello", TagString: "hello"}}, cfg)
}

func TestParseNestedPtr(t *testing.T) {
	type Config struct {
		String string
	}

	type Outer struct {
		Config *Config
	}

	t.Setenv("CONFIG_STRING", "hello")

	cfg := Outer{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Outer{Config: &Config{String: "hello"}}, cfg)
}

func TestParseNestedPrefix(t *testing.T) {
	type Config struct {
		String string
	}

	type Outer struct {
		Config Config
	}

	t.Setenv("CONFIG_STRING", "hello")

	cfg := Outer{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Outer{Config: Config{String: "hello"}}, cfg)
}

func TestParseNestedTag(t *testing.T) {
	type Inner struct {
		String string `what:"MY_STRING"`
	}

	type Config struct {
		Inner Inner `what:"MY_INNER"`
	}

	t.Setenv("MY_INNER_MY_STRING", "hello")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{Inner: Inner{String: "hello"}}, cfg)
}

func TestParseEmbedded(t *testing.T) {
	type Base struct {
		BaseField string
	}

	type Config struct {
		Base
	}

	t.Setenv("BASE_BASE_FIELD", "base")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		Base: Base{
			BaseField: "base",
		},
	}, cfg)
}

func TestParseTags(t *testing.T) {
	type Config struct {
		EnvString          string `env:"MY_STRING"`
		JSONString         string `json:"my_json_string"`
		TomlString         string `toml:"my_toml_string"`
		MapstructureString string `mapstructure:"my_mapstructure_string"`
	}

	t.Setenv("MY_STRING", "hello")
	t.Setenv("MY_JSON_STRING", "hello")
	t.Setenv("MY_TOML_STRING", "hello")
	t.Setenv("MY_MAPSTRUCTURE_STRING", "hello")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		EnvString:          "hello",
		JSONString:         "hello",
		TomlString:         "hello",
		MapstructureString: "hello",
	}, cfg)
}

func TestParseWithParserFunc(t *testing.T) {
	type Inter interface{}

	type Impl struct {
		Value string
	}

	type Config struct {
		Inter Inter
	}

	t.Setenv("INTER", "value")

	cfg := Config{}

	err := envcfg.Parse(&cfg, envcfg.WithTypeParser(reflect.TypeOf((*Inter)(nil)).Elem(), func(value string) (any, error) {
		return &Impl{Value: value}, nil
	}))

	require.NoError(t, err)
	assert.Equal(t, Config{Inter: &Impl{Value: "value"}}, cfg)
}

func TestParseDuration(t *testing.T) {
	type Config struct {
		Duration    time.Duration
		DurationPtr *time.Duration
	}

	t.Setenv("DURATION", "1s")
	t.Setenv("DURATION_PTR", "1s")
	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		Duration:    1 * time.Second,
		DurationPtr: durationPtr(1 * time.Second),
	}, cfg)
}

func TestParseURL(t *testing.T) {
	type Inner struct {
		Url  url.URL
		Urls []url.URL
	}

	type Config struct {
		URL    url.URL
		URLPtr *url.URL
		Nested Inner
	}

	t.Setenv("URL", "http://example.com")
	t.Setenv("URL_PTR", "http://example.com")

	t.Setenv("NESTED_URL", "http://nested.com")
	t.Setenv("NESTED_URLS_0", "http://nested.com")
	t.Setenv("NESTED_URLS_1", "https://nested.com")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		URL:    url.URL{Scheme: "http", Host: "example.com"},
		URLPtr: &url.URL{Scheme: "http", Host: "example.com"},
		Nested: Inner{
			Url: url.URL{Scheme: "http", Host: "nested.com"},
			Urls: []url.URL{
				{Scheme: "http", Host: "nested.com"},
				{Scheme: "https", Host: "nested.com"},
			},
		},
	}, cfg)
}

func TestParseInitVars(t *testing.T) {
	type Inner struct {
		Value string
	}

	type Config struct {
		String      string
		StringPtr   *string
		Int         int
		IntPtr      *int
		Bool        bool
		BoolPtr     *bool
		Float       float64
		FloatPtr    *float64
		Struct      Inner
		StructPtr   *Inner
		Slice       []string
		SlicePtr    *[]string
		Map         map[string]string
		PtrMap      *map[string]string
		Duration    time.Duration
		DurationPtr *time.Duration
		URL         url.URL
		URLPtr      *url.URL

		EmptyString      string
		EmptyStringPtr   *string
		EmptyInt         int
		EmptyIntPtr      *int
		EmptyBool        bool
		EmptyBoolPtr     *bool
		EmptyFloat       float64
		EmptyFloatPtr    *float64
		EmptyStruct      Inner
		EmptyStructPtr   *Inner
		EmptySlice       []string
		EmptySlicePtr    *[]string
		EmptyMap         map[string]string
		EmptyMapPtr      *map[string]string
		EmptyDuration    time.Duration
		EmptyDurationPtr *time.Duration
		EmptyURL         url.URL
		EmptyURLPtr      *url.URL
	}

	t.Setenv("STRING", "hello")
	t.Setenv("STRING_PTR", "hello")
	t.Setenv("INT", "1")
	t.Setenv("INT_PTR", "1")
	t.Setenv("BOOL", "true")
	t.Setenv("BOOL_PTR", "true")
	t.Setenv("FLOAT", "1.1")
	t.Setenv("FLOAT_PTR", "1.1")
	t.Setenv("STRUCT_VALUE", "hello")
	t.Setenv("STRUCT_PTR_VALUE", "hello")
	t.Setenv("SLICE_0", "hello")
	t.Setenv("SLICE_PTR_0", "hello")
	t.Setenv("MAP_KEY", "hello")
	t.Setenv("PTR_MAP_KEY", "hello")
	t.Setenv("DURATION", "10s")
	t.Setenv("DURATION_PTR", "10s")
	t.Setenv("URL", "http://example.com")
	t.Setenv("URL_PTR", "https://example.com")

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	require.NoError(t, err)
	assert.Equal(t, Config{
		String:      "hello",
		StringPtr:   strPtr("hello"),
		Int:         1,
		IntPtr:      intPtr(1),
		Bool:        true,
		BoolPtr:     boolPtr(true),
		Float:       1.1,
		FloatPtr:    float64Ptr(1.1),
		Struct:      Inner{Value: "hello"},
		StructPtr:   &Inner{Value: "hello"},
		Slice:       []string{"hello"},
		SlicePtr:    &[]string{"hello"},
		Map:         map[string]string{"key": "hello"},
		PtrMap:      &map[string]string{"key": "hello"},
		Duration:    10 * time.Second,
		DurationPtr: durationPtr(10 * time.Second),
		URL:         url.URL{Scheme: "http", Host: "example.com"},
		URLPtr:      &url.URL{Scheme: "https", Host: "example.com"},

		EmptyString:      "",
		EmptyStringPtr:   nil,
		EmptyInt:         0,
		EmptyIntPtr:      nil,
		EmptyBool:        false,
		EmptyBoolPtr:     nil,
		EmptyFloat:       0,
		EmptyFloatPtr:    nil,
		EmptyStruct:      Inner{},
		EmptyStructPtr:   nil,
		EmptySlice:       nil,
		EmptySlicePtr:    nil,
		EmptyMap:         nil,
		EmptyMapPtr:      nil,
		EmptyDuration:    0,
		EmptyDurationPtr: nil,
		EmptyURL:         url.URL{},
		EmptyURLPtr:      nil,
	}, cfg)
}

func TestParseInitAny(t *testing.T) {
	type Inner struct {
		Value string
	}

	type DefaultInner struct {
		Value string `default:"hello"`
	}

	type Config struct {
		String      string
		StringPtr   *string
		Int         int
		IntPtr      *int
		Bool        bool
		BoolPtr     *bool
		Float       float64
		FloatPtr    *float64
		Struct      Inner
		StructPtr   *Inner
		Slice       []string
		SlicePtr    *[]string
		Map         map[string]string
		PtrMap      *map[string]string
		Duration    time.Duration
		DurationPtr *time.Duration
		URL         url.URL
		URLPtr      *url.URL

		DefaultString      string   `default:"hello"`
		DefaultStringPtr   *string  `default:"hello"`
		DefaultInt         int      `default:"1"`
		DefaultIntPtr      *int     `default:"1"`
		DefaultBool        bool     `default:"true"`
		DefaultBoolPtr     *bool    `default:"true"`
		DefaultFloat       float64  `default:"1.1"`
		DefaultFloatPtr    *float64 `default:"1.1"`
		DefaultStruct      DefaultInner
		DefaultStructPtr   *DefaultInner
		DefaultSlice       []string           `default:"hello"`
		DefaultSlicePtr    *[]string          `default:"hello"`
		DefaultMap         map[string]string  `default:"key:hello"`
		DefaultMapPtr      *map[string]string `default:"key:hello"`
		DefaultDuration    time.Duration      `default:"10s"`
		DefaultDurationPtr *time.Duration     `default:"10s"`
		DefaultURL         url.URL            `default:"http://example.com"`
		DefaultURLPtr      *url.URL           `default:"https://example.com"`

		EmptyString      string
		EmptyStringPtr   *string
		EmptyInt         int
		EmptyIntPtr      *int
		EmptyBool        bool
		EmptyBoolPtr     *bool
		EmptyFloat       float64
		EmptyFloatPtr    *float64
		EmptyStruct      Inner
		EmptyStructPtr   *Inner
		EmptySlice       []string
		EmptySlicePtr    *[]string
		EmptyMap         map[string]string
		EmptyMapPtr      *map[string]string
		EmptyDuration    time.Duration
		EmptyDurationPtr *time.Duration
		EmptyURL         url.URL
		EmptyURLPtr      *url.URL
	}

	t.Setenv("STRING", "hello")
	t.Setenv("STRING_PTR", "hello")
	t.Setenv("INT", "1")
	t.Setenv("INT_PTR", "1")
	t.Setenv("BOOL", "true")
	t.Setenv("BOOL_PTR", "true")
	t.Setenv("FLOAT", "1.1")
	t.Setenv("FLOAT_PTR", "1.1")
	t.Setenv("STRUCT_VALUE", "hello")
	t.Setenv("STRUCT_PTR_VALUE", "hello")
	t.Setenv("SLICE_0", "hello")
	t.Setenv("SLICE_PTR_0", "hello")
	t.Setenv("MAP_KEY", "hello")
	t.Setenv("PTR_MAP_KEY", "hello")
	t.Setenv("DURATION", "10s")
	t.Setenv("DURATION_PTR", "10s")
	t.Setenv("URL", "http://example.com")
	t.Setenv("URL_PTR", "https://example.com")

	cfg := Config{}
	err := envcfg.Parse(&cfg, envcfg.WithInitAny())

	require.NoError(t, err)
	assert.Equal(t, Config{
		String:      "hello",
		StringPtr:   strPtr("hello"),
		Int:         1,
		IntPtr:      intPtr(1),
		Bool:        true,
		BoolPtr:     boolPtr(true),
		Float:       1.1,
		FloatPtr:    float64Ptr(1.1),
		Struct:      Inner{Value: "hello"},
		StructPtr:   &Inner{Value: "hello"},
		Slice:       []string{"hello"},
		SlicePtr:    &[]string{"hello"},
		Map:         map[string]string{"key": "hello"},
		PtrMap:      &map[string]string{"key": "hello"},
		Duration:    10 * time.Second,
		DurationPtr: durationPtr(10 * time.Second),
		URL:         url.URL{Scheme: "http", Host: "example.com"},
		URLPtr:      &url.URL{Scheme: "https", Host: "example.com"},

		DefaultString:      "hello",
		DefaultStringPtr:   strPtr("hello"),
		DefaultInt:         1,
		DefaultIntPtr:      intPtr(1),
		DefaultBool:        true,
		DefaultBoolPtr:     boolPtr(true),
		DefaultFloat:       1.1,
		DefaultFloatPtr:    float64Ptr(1.1),
		DefaultStruct:      DefaultInner{Value: "hello"},
		DefaultStructPtr:   &DefaultInner{Value: "hello"},
		DefaultSlice:       []string{"hello"},
		DefaultSlicePtr:    &[]string{"hello"},
		DefaultMap:         map[string]string{"key": "hello"},
		DefaultMapPtr:      &map[string]string{"key": "hello"},
		DefaultDuration:    10 * time.Second,
		DefaultDurationPtr: durationPtr(10 * time.Second),
		DefaultURL:         url.URL{Scheme: "http", Host: "example.com"},
		DefaultURLPtr:      &url.URL{Scheme: "https", Host: "example.com"},

		EmptyString:      "",
		EmptyStringPtr:   nil,
		EmptyInt:         0,
		EmptyIntPtr:      nil,
		EmptyBool:        false,
		EmptyBoolPtr:     nil,
		EmptyFloat:       0,
		EmptyFloatPtr:    nil,
		EmptyStruct:      Inner{},
		EmptyStructPtr:   nil,
		EmptySlice:       nil,
		EmptySlicePtr:    nil,
		EmptyMap:         nil,
		EmptyMapPtr:      nil,
		EmptyDuration:    0,
		EmptyDurationPtr: nil,
		EmptyURL:         url.URL{},
		EmptyURLPtr:      nil,
	}, cfg)
}

func TestParseInitNever(t *testing.T) {
	type Inner struct {
		Value string
	}

	type Config struct {
		EmptyStringPtr   *string
		EmptyIntPtr      *int
		EmptyBoolPtr     *bool
		EmptyFloatPtr    *float64
		EmptyStructPtr   *Inner
		EmptySlicePtr    *[]string
		EmptyMapPtr      *map[string]string
		EmptyDurationPtr *time.Duration
		EmptyURLPtr      *url.URL
	}

	t.Setenv("EMPTY_STRING", "hello")
	t.Setenv("EMPTY_INT", "1")
	t.Setenv("EMPTY_BOOL", "true")
	t.Setenv("EMPTY_FLOAT", "1.1")
	t.Setenv("EMPTY_DURATION", "1s")
	t.Setenv("EMPTY_URL", "http://example.com")
	t.Setenv("EMPTY_SLICE_0", "hello")
	t.Setenv("EMPTY_MAP_KEY", "hello")
	t.Setenv("EMPTY_DURATION", "10s")
	t.Setenv("EMPTY_URL", "http://example.com")
	t.Setenv("EMPTY_STRUCT_VALUE", "hello")

	cfg := Config{}
	err := envcfg.Parse(&cfg, envcfg.WithInitNever())

	require.NoError(t, err)
	assert.Equal(t, Config{
		EmptyStringPtr:   nil,
		EmptyIntPtr:      nil,
		EmptyBoolPtr:     nil,
		EmptyFloatPtr:    nil,
		EmptyStructPtr:   nil,
		EmptySlicePtr:    nil,
		EmptyMapPtr:      nil,
		EmptyDurationPtr: nil,
		EmptyURLPtr:      nil,
	}, cfg)
}

func TestParseInitAlways(t *testing.T) {
	type Inner struct {
		Value string
	}

	type Config struct {
		EmptyString   *string
		EmptyInt      *int
		EmptyBool     *bool
		EmptyFloat    *float64
		EmptyStruct   *Inner
		EmptyDuration *time.Duration
		EmptyURL      *url.URL
	}

	cfg := Config{}
	err := envcfg.Parse(&cfg, envcfg.WithInitAlways())

	require.NoError(t, err)
	assert.Equal(t, Config{
		EmptyString:   strPtr(""),
		EmptyInt:      intPtr(0),
		EmptyBool:     boolPtr(false),
		EmptyFloat:    float64Ptr(0),
		EmptyStruct:   &Inner{},
		EmptyDuration: durationPtr(0),
		EmptyURL:      &url.URL{},
	}, cfg)
}

func TestParseDecodeUnset(t *testing.T) {
	type Config struct {
		Unset Unset `env:",decodeunset"`
	}

	cfg := Config{}
	err := envcfg.Parse(&cfg)
	require.Error(t, err)
}

type Unset string

func (u *Unset) UnmarshalText(text []byte) error {
	return fmt.Errorf("unset")
}

func strPtr(s string) *string {
	return &s
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}

func float64Ptr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
