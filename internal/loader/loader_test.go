package loader

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSource(t *testing.T) {
	loader := New()
	err := loader.Build(WithSource(&envSource{map[string]string{"TEST_KEY": "value"}}))
	assert.NoError(t, err)

	envs, err := loader.Load()

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"TEST_KEY": "value"}, envs)
}

func TestWithDefaultSources(t *testing.T) {
	t.Setenv("TEST_KEY", "value")

	loader := New()
	err := loader.Build(
		WithPrefix("TEST_"),
	)
	assert.NoError(t, err)

	envs, err := loader.Load()

	require.NoError(t, err)
	assert.EqualValues(t, map[string]string{"KEY": "value"}, envs)
}

func TestWithEnvVarsSource(t *testing.T) {
	customEnvs := map[string]string{
		"CUSTOM_KEY": "custom_value",
	}

	loader := New()
	err := loader.Build(WithEnvVarsSource(customEnvs))
	require.NoError(t, err)

	envs, err := loader.Load()

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"CUSTOM_KEY": "custom_value"}, envs)
}

func TestFileSource(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "env")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("FILE_KEY=file_value")
	assert.NoError(t, err)

	loader := New()
	err = loader.Build(WithFileSource(tmpFile.Name()))
	require.NoError(t, err)

	envs, err := loader.Load()
	assert.Equal(t, map[string]string{"FILE_KEY": "file_value"}, envs)
}

func TestWithOsEnvSource(t *testing.T) {
	loader := New()
	err := loader.Build(WithOsEnvSource())
	require.NoError(t, err)

	t.Setenv("TEST_KEY", "value")

	envs, err := loader.Load()
	assert.Contains(t, envs, "TEST_KEY")
}

func TestWithDefaults(t *testing.T) {
	defaults := map[string]string{
		"DEFAULT_KEY": "default_value",
		"OTHER_KEY":   "default_other_value",
	}

	loader := New()
	err := loader.Build(
		WithEnvVarsSource(map[string]string{
			"OTHER_KEY": "other_value",
		}),
		WithDefaults(defaults),
	)
	require.NoError(t, err)

	envs, err := loader.Load()

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"DEFAULT_KEY": "default_value",
		"OTHER_KEY":   "other_value",
	}, envs)
}

func TestWithFilter(t *testing.T) {
	loader := New()
	err := loader.Build(
		WithEnvVarsSource(map[string]string{"TEST_KEY": "value"}),
		WithFilter(func(key string) bool {
			return strings.HasPrefix(key, "TEST_")
		}),
	)
	require.NoError(t, err)

	envs, err := loader.Load()
	assert.Equal(t, map[string]string{"TEST_KEY": "value"}, envs)
}

func TestWithHasPrefix(t *testing.T) {
	loader := New()
	err := loader.Build(
		WithEnvVarsSource(map[string]string{
			"APP_TEST":   "value",
			"OTHER_TEST": "other",
		}),
		WithHasPrefix("APP_"),
	)
	require.NoError(t, err)

	envs, err := loader.Load()
	assert.Equal(t, map[string]string{"APP_TEST": "value"}, envs)
}

func TestWithHasSuffix(t *testing.T) {
	loader := New()
	err := loader.Build(
		WithEnvVarsSource(map[string]string{
			"APP_TEST":   "value",
			"APP_TEST_2": "other",
		}),
		WithHasSuffix("_TEST"),
	)
	require.NoError(t, err)

	envs, err := loader.Load()
	assert.Equal(t, map[string]string{"APP_TEST": "value"}, envs)
}

func TestWithHasMatch(t *testing.T) {
	loader := New()
	err := loader.Build(
		WithEnvVarsSource(map[string]string{
			"TEST_123": "value1",
			"TEST_ABC": "value2",
		}),
		WithHasMatch(`TEST_\d+`),
	)
	require.NoError(t, err)

	envs, err := loader.Load()
	assert.Equal(t, map[string]string{"TEST_123": "value1"}, envs)
}

func TestWithTransform(t *testing.T) {
	loader := New()
	err := loader.Build(
		WithEnvVarsSource(map[string]string{"TEST_KEY": "value"}),
		WithTransform(func(key string) string {
			return "PREFIX_" + key
		}))
	require.NoError(t, err)

	envs, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"PREFIX_TEST_KEY": "value"}, envs)
}

func TestWithTrimPrefix(t *testing.T) {
	loader := New()
	err := loader.Build(
		WithEnvVarsSource(map[string]string{"APP_TEST": "value"}),
		WithTrimPrefix("APP_"),
	)
	require.NoError(t, err)

	envs, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"TEST": "value"}, envs)
}

func TestWithTrimSuffix(t *testing.T) {
	loader := New()
	err := loader.Build(
		WithEnvVarsSource(map[string]string{
			"APP_TEST":   "value",
			"APP_TEST_2": "other",
		}),
		WithTrimSuffix("_TEST"),
	)
	require.NoError(t, err)

	envs, err := loader.Load()
	assert.Equal(t, map[string]string{
		"APP":        "value",
		"APP_TEST_2": "other",
	}, envs)
}

func TestWithPrefix(t *testing.T) {
	loader := New()
	err := loader.Build(
		WithEnvVarsSource(map[string]string{
			"APP_TEST":   "value",
			"OTHER_TEST": "other",
		}),
		WithPrefix("APP_"))
	require.NoError(t, err)

	envs, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"TEST": "value"}, envs)
}

func TestWithSuffix(t *testing.T) {
	loader := New()
	err := loader.Build(
		WithEnvVarsSource(map[string]string{
			"APP_TEST":   "value",
			"APP_TEST_2": "other",
		}),
		WithSuffix("_TEST"),
	)
	require.NoError(t, err)

	envs, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"APP": "value",
	}, envs)
}
