# envcfg

Parse environment variables into Go structs with minimal boilerplate and first-class support for complex data structures

[![Go Reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/sethpollack/envcfg)
[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen?logo=go&style=flat-square)](https://goreportcard.com/report/github.com/sethpollack/envcfg)
[![Coverage](https://img.shields.io/codecov/c/github/sethpollack/envcfg?logo=codecov&style=flat-square)](https://codecov.io/gh/sethpollack/envcfg)
[![Test Status](https://img.shields.io/github/actions/workflow/status/sethpollack/envcfg/test.yml?logo=github&style=flat-square)](https://github.com/sethpollack/envcfg/actions)
[![Release](https://img.shields.io/github/v/release/sethpollack/envcfg?logo=github&style=flat-square)](https://github.com/sethpollack/envcfg/releases/latest)
[![License](https://img.shields.io/github/license/sethpollack/envcfg?logo=opensourceinitiative&logoColor=white&style=flat-square)](https://github.com/sethpollack/envcfg/blob/main/LICENSE)


## Key Features

### ✨ Zero Configuration Philosophy
- Field names just work (snake or camel)
- Nested structs automatically create prefixes
- No special tags needed

### 🔄 Smart Tag Fallbacks
- When field name doesn't match, tries existing json/yaml/toml etc.. tags
- Explicit env tags available and take highest precedence, but rarely needed

### 🎯 Simple Environment Variables
- One string value per variable
- No need to encode complex data in environment values

### 🧩 Complex Types Made Simple
- Maps: `SETTINGS_KEY=value`
- Arrays: `PORTS_0=8080`
- Nested structs: `REDIS_HOST=localhost`
- Slice of structs: `SERVERS_0_HOST=host1`
- Map of structs: `DATABASES_PRIMARY_HOST=host1`

### 📦 Compatible With Popular Libraries
- Most features from popular libraries available

### 🛠️ Fully Configurable
- No lock-in to specific tag formats, tags are all configurable
- Default settings are all configurable
- Parsers/decoders are all configurable

## Table of Contents
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Types](#types)
- [Decoders](#decoders)
- [Struct Tags](#struct-tags)
  - [Init Options](#init-options)
- [Field Name Mapping](#field-name-mapping)
- [Functions](#functions)
  - [Configuration Options](#configuration-options)
    - [Tag Overrides](#tag-overrides)
    - [Default Overrides](#default-overrides)
    - [Custom Parser Functions](#custom-parser-functions)
    - [Custom Decoder Functions](#custom-decoder-functions)
    - [Loaders](#loaders)
    - [Loader Options](#loader-options)
    - [Configuration Sources](#configuration-sources)
    - [Source Ordering](#source-ordering)

## Installation

```bash
go get github.com/sethpollack/envcfg
```

## Quick Start

Set the following environment variables:
```bash
# string
export NAME=name
# integer
export PORT=8080
# float
export RATE=1.23
# boolean
export IS_ENABLED=true
# duration
export TIMEOUT=60s
# nested struct
export REDIS_HOST=localhost
export REDIS_PORT=6379
# delimited slice
export TAGS=tag1,tag2,tag3
# indexed slice
export PORTS_0=8080
export PORTS_1=9090
# slice of structs
export SERVERS_0_HOST=localhost1
export SERVERS_0_PORT=8080
export SERVERS_1_HOST=localhost2
export SERVERS_1_PORT=9090
# key-value map
export LABELS=key1:value1,key2:value2
# flat map
export SETTINGS_KEY1=1
export SETTINGS_KEY2=2
# map of structs
export DATABASES_PRIMARY_HOST=localhost1
export DATABASES_PRIMARY_PORT=6379
export DATABASES_SECONDARY_HOST=localhost2
export DATABASES_SECONDARY_PORT=6380
```

Parse the environment variables into a struct:

```go
package main

import (
	"fmt"
	"time"

	"github.com/sethpollack/envcfg"
)

func main() {
	type ServerConfig struct {
		Host string
		Port int
	}

	type Config struct {
		Name      string
		Port      int
		Rate      float64
		IsEnabled bool
		Timeout   time.Duration

		Redis ServerConfig

		Tags    []string
		Ports   []int
		Servers []ServerConfig

		Labels    map[string]string
		Settings  map[string]int
		Databases map[string]ServerConfig

	}

	cfg := Config{}
	if err := envcfg.Parse(&cfg); err != nil {
		panic(err)
	}

	fmt.Printf(`Config:
  Name: %s
  Port: %d
  Rate: %.2f
  IsEnabled: %t
  Timeout: %s
  Redis:
    Host: %s
    Port: %d
  Tags: %v
  Ports: %v
  Servers: [
    {Host: %s, Port: %d},
    {Host: %s, Port: %d}
  ]
  Labels: %v
  Settings: %v
  Databases: {
    primary: {Host: %s, Port: %d},
    secondary: {Host: %s, Port: %d}
  }
`,
		cfg.Name, cfg.Port, cfg.Rate, cfg.IsEnabled, cfg.Timeout,
		cfg.Redis.Host, cfg.Redis.Port,
		cfg.Tags, cfg.Ports,
		cfg.Servers[0].Host, cfg.Servers[0].Port,
		cfg.Servers[1].Host, cfg.Servers[1].Port,
		cfg.Labels, cfg.Settings,
		cfg.Databases["primary"].Host, cfg.Databases["primary"].Port,
		cfg.Databases["secondary"].Host, cfg.Databases["secondary"].Port,
	)
}
```

**Results**:

```bash
Config:
  Name: name
  Port: 8080
  Rate: 1.23
  IsEnabled: true
  Timeout: 1m0s
  Redis:
    Host: localhost
    Port: 6379
  Tags: [tag1 tag2 tag3]
  Ports: [8080 9090]
  Servers: [
    {Host: localhost1, Port: 8080},
    {Host: localhost2, Port: 9090}
  ]
  Labels: map[key1:value1 key2:value2]
  Settings: map[key1:1 key2:2]
  Databases: {
    primary: {Host: localhost1, Port: 6379},
    secondary: {Host: localhost2, Port: 6380}
  }
```

> [!TIP]
> The example above demonstrates the three supported syntaxes for both slices and maps:
> - **Slices**:
>   - delimited `TAGS=tag1,tag2,tag3`
>   - indexed `PORTS_0=8080`
>   - struct `SERVERS_0_HOST=localhost`
> - **Maps**:
>   - key-value pairs `LABELS=key1:value1`
>   - flat keys `SETTINGS_KEY1=1`
>   - struct `DATABASES_PRIMARY_HOST=localhost`


## Types

- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `bool`
- `float32`, `float64`
- `time.Duration`
- `structs`
- `slices`
- `maps`

> [!NOTE]
> Type support can be extended using the `WithKindParser` and `WithTypeParser` options.

## Decoders

- `envcfg.Decoder`
- `flag.Value`
- `encoding.TextUnmarshaler`
- `encoding.BinaryUnmarshaler`

> [!NOTE]
> Decoder support can be extended using the `WithDecoder` option.


## Struct Tags

| Name | Description | Default | Example Tag | Example Option |
|-----|-------------|---------|-----|--------|
| `default` | Default value when environment variable is not set | - | `default:"8080"` | `env:",default=8080"` |
| `required` | Mark field as required | `false` | `required:"true"` | `env:",required"` |
| `notempty` | Ensure value is not empty | `false` | `notempty:"true"` | `env:",notempty"` |
| `expand` | Expand environment variables in value | `false` | `expand:"true"` | `env:",expand"` |
| `file` | Load value from file | `false` | `file:"true"` | `env:",file"` |
| `delim` | Delimiter for array values | `,` | `delim:";"` | `env:",delim=;"` |
| `sep` | Separator for map key-value pairs | `:` | `sep:"="` | `env:",sep="` |
| `init` | Initialize nil pointers | `vars` | `init:"always"` | `env:",init=always"` |
| `ignore` | Ignore field | `false` | `ignore:"true"` | `env:",ignore"` or `env:"-"` |
| `decodeunset` | Decode unset environment variables | `false` | `decodeunset:"true"` | `env:",decodeunset"` |

> [!WARNING]
> When setting default values for slices, avoid using the comma as it conflicts with tag parsing. Either use a different delimiter or set array values using environment variables:
> ```go
> type Config struct {
>     Tags []string `env:"TAGS,default=tag1,tag2"` // This will fail!
> }
>
> // Do this instead:
> type Config struct {
>     Tags []string `env:"TAGS,delim=;,default=tag1;tag2"` // Use a different delimiter
> }
> ```

### Init Options
- `vars` - Initialize when values are present with the exception of structs that only have default values. (default)
- `any` - Same as `vars`, but also initialize structs that only have default values.
- `always` - Always initialize
- `never` - Never initialize


> [!TIP]
> All defaults and tag names can be customized using the `With*` options. See [Configuration Options](#configuration-options) for more details.


## Field Name Mapping

By default, `envcfg` will search for environment variables using multiple naming patterns until a match is found. Use `WithDisableFallback` to restrict matching to only the `env` tag value.

For example:

```go
os.Setenv("CUSTOM_URL",  "value") // Matches env tag
os.Setenv("DATABASEURL",  "value") // Matches struct field
os.Setenv("DATABASE_URL", "value") // Matches snake-case
os.Setenv("DB_URL",      "value") // Matches json tag
os.Setenv("DATA_SOURCE", "value") // Matches yaml tag
// ...

type Config struct {
    DatabaseURL string `json:"db_url" yaml:"data_source" env:"custom_url"`
}
```

> [!TIP]
> All environment variable matching is case __insensitive__.

## Functions
 - `Parse` - Parse environment variables into a struct pointer
 - `MustParse` - Same as `Parse`, but panics on error
 - `ParseAs` - Parse environment variables into a specific type
 - `MustParseAs` - Same as `ParseAs`, but panics on error

> [!IMPORTANT]
> `envcfg` only parses __exported__ fields.

### Configuration Options

#### Tag Overrides

| Option | Description | Default |
|--------|-------------|---------|
| `WithTagName` | Tag name for environment variables | `env` |
| `WithDelimiterTag` | Tag name for delimiter | `delim` |
| `WithSeparatorTag` | Tag name for separator | `sep` |
| `WithDecodeUnsetTag` | Tag name for decoding unset environment variables | `decodeunset` |
| `WithDefaultTag` | Tag name for default values | `default` |
| `WithExpandTag` | Tag name for expandable variables | `expand` |
| `WithFileTag` | Tag name for file variables | `file` |
| `WithNotEmptyTag` | Tag name for not empty variables | `notempty` |
| `WithRequiredTag` | Tag name for required variables | `required` |
| `WithIgnoreTag` | Tag name for ignored variables | `ignore` |
| `WithInitTag` | Tag name for initialization strategy | `init` |

#### Default Overrides

| Option | Description | Default |
|--------|-------------|---------|
| `WithDelimiter` | Sets the default delimiter for array and map values | `,` |
| `WithSeparator` | Sets the default separator for map key-value pairs | `:` |
| `WithDecodeUnset` | Enables decoding unset environment variables by default | `false` |
| `WithInitAny` | Sets the initialization strategy to `any` | `vars` |
| `WithInitNever` | Sets the initialization strategy to `never` | `vars` |
| `WithInitAlways` | Sets the initialization strategy to `always` | `vars` |
| `WithExpand` | Enables environment variable expansion by default | `false` |
| `WithNotEmpty` | Enables validating that values are not empty by default | `false` |
| `WithRequired` | Enables marking fields as required by default | `false` |
| `WithDisableFallback` | Enables strict matching using the `env` tag | `false` |

#### Custom Parser Functions

| Option | Description |
|--------|-------------|
| `WithTypeParser` | Registers a custom type parser |
| `WithTypeParsers` | Registers custom type parsers |
| `WithKindParser` | Registers a custom kind parser |
| `WithKindParsers` | Registers custom kind parsers |

#### Custom Decoder Functions

| Option | Description |
|--------|-------------|
| `WithDecoder` | Registers a custom decoder function for a specific interface |

#### Loaders

| Option | Description |
|--------|-------------|
| `WithLoader` | Registers a loader |

#### Loader Options

Loader options configure how environment variables are loaded and filtered before being matched to struct fields. These options allow you to:
- Add different sources of configuration (environment variables, files, external services)
- Filter which environment variables are considered
- Transform environment variable names before matching
- Set default values for when environment variables are not found

| Option | Description |
|--------|-------------|
| `WithSource` | Adds a source to the loader |
| `WithSources` | Adds multiple sources to the loader |
| `WithOSEnvSource` | Adds OS environment variables as a source |
| `WithMapEnvSource` | Uses the provided map of environment variables as a source |
| `WithDotEnvSource` | Adds environment variables from a file as a source |
| `WithPrefix` | Combines `WithTrimPrefix` and `WithHasPrefix` |
| `WithSuffix` | Combines `WithTrimSuffix` and `WithHasSuffix` |
| `WithTransform` | Adds a transform function that modifies environment variable keys |
| `WithTrimPrefix` | Removes a prefix from environment variable names |
| `WithTrimSuffix` | Removes a suffix from environment variable names |
| `WithFilter` | Adds a custom filter to the loader |
| `WithHasPrefix` | Adds a prefix filter to the loader |
| `WithHasSuffix` | Adds a suffix filter to the loader |
| `WithHasMatch` | Adds a pattern filter to the loader |
| `WithKeys` | Adds a filter to the loader that only includes specific keys |

#### Configuration Sources

envcfg supports multiple configuration sources that can be used individually or combined. Built-in sources have dedicated `With*` functions, while custom sources and those maintained as separate Go modules (to keep dependencies isolated) can be added using `WithSource`.


| Option | Description |
|--------|-------------|
| `WithOSEnvSource` | Adds OS environment variables as a source |
| `WithMapEnvSource` | Uses the provided map of environment variables as a source |
| `WithDotEnvSource` | Adds environment variables from a .env file as a source |
| `WithSource(awssm.New(...))` | Adds AWS Secrets Manager as a source |


#### Source Ordering

Sources are processed in the order they are added, with later sources taking precedence over earlier ones. This ordering allows you to:
- Set default values by adding a source with defaults first
- Override values by adding sources with higher precedence later
- Create a hierarchy of configuration (e.g., defaults → shared config → environment-specific config → local overrides)

```go
  envcfg.Parse(&cfg,
    envcfg.WithLoader(
      envcfg.WithMapEnvSource(map[string]string{
          "APP_PORT": "8080",
        "LOG_LEVEL": "info",
      }),
      envcfg.WithDotEnvSource(".env"),
      envcfg.WithSource(awssm.New(
        awssm.WithRegion("us-west-2"),
        awssm.WithSecretID("myapp/config"),
      )),
      envcfg.WithOSEnvSource(),
    ),
)
```

In this setup:
1. Default values are loaded first
2. Then values from the .env file are loaded next
3. Then values from AWS Secrets Manager
4. Finally, OS environment variables take precedence over both defaults and .env file

