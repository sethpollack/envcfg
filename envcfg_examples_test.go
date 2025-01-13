package envcfg_test

import (
	"fmt"
	"os"
	"time"

	"github.com/sethpollack/envcfg"
)

func ExampleParse() {
	os.Setenv("NAME", "name")
	os.Setenv("PORT", "8080")
	os.Setenv("RATE", "1.23")
	os.Setenv("IS_ENABLED", "true")
	os.Setenv("TIMEOUT", "60s")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("TAGS", "tag1,tag2,tag3")
	os.Setenv("PORTS_0", "8080")
	os.Setenv("PORTS_1", "9090")
	os.Setenv("SERVERS_0_HOST", "localhost1")
	os.Setenv("SERVERS_0_PORT", "8080")
	os.Setenv("SERVERS_1_HOST", "localhost2")
	os.Setenv("SERVERS_1_PORT", "9090")
	os.Setenv("LABELS", "key1:value1,key2:value2")
	os.Setenv("SETTINGS_KEY1", "1")
	os.Setenv("SETTINGS_KEY2", "2")
	os.Setenv("DATABASES_PRIMARY_HOST", "localhost1")
	os.Setenv("DATABASES_PRIMARY_PORT", "6379")
	os.Setenv("DATABASES_SECONDARY_HOST", "localhost2")
	os.Setenv("DATABASES_SECONDARY_PORT", "6380")

	defer os.Clearenv()

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
		Redis     ServerConfig
		Tags      []string
		Ports     []int
		Servers   []ServerConfig
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

	// Output:
	// Config:
	//   Name: name
	//   Port: 8080
	//   Rate: 1.23
	//   IsEnabled: true
	//   Timeout: 1m0s
	//   Redis:
	//     Host: localhost
	//     Port: 6379
	//   Tags: [tag1 tag2 tag3]
	//   Ports: [8080 9090]
	//   Servers: [
	//     {Host: localhost1, Port: 8080},
	//     {Host: localhost2, Port: 9090}
	//   ]
	//   Labels: map[key1:value1 key2:value2]
	//   Settings: map[key1:1 key2:2]
	//   Databases: {
	//     primary: {Host: localhost1, Port: 6379},
	//     secondary: {Host: localhost2, Port: 6380}
	//   }
}

func ExampleParse_fieldMatching() {
	os.Setenv("FIELDNAME", "field name")
	os.Setenv("SNAKE_FIELD_NAME", "snake field name")
	os.Setenv("TOML_OVERRIDE", "toml override")
	os.Setenv("ENV_FIELD_NAME", "env field name")

	defer os.Clearenv()

	type Config struct {
		FieldName      string
		SnakeFieldName string
		TomlFieldName  string `toml:"toml_override"`
		EnvFieldName   string `env:"ENV_FIELD_NAME"`
	}

	cfg := Config{}
	if err := envcfg.Parse(&cfg); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", cfg)
	// Output: {FieldName:field name SnakeFieldName:snake field name TomlFieldName:toml override EnvFieldName:env field name}
}

func ExampleParse_tags() {
	tmpfile, err := os.CreateTemp("", "test.txt")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString("${FILE_CONTENTS}")
	if err != nil {
		panic(err)
	}

	os.Setenv("EXPAND", "${EXPAND_OTHER}")
	os.Setenv("EXPAND_OTHER", "expand other")
	os.Setenv("FILE_CONTENTS", "file contents")
	os.Setenv("FILE", tmpfile.Name())
	os.Setenv("EXPAND_FILE", tmpfile.Name())
	os.Setenv("IGNORE", "ignore")
	os.Setenv("INIT_VALUES_FIELD", "init values")
	os.Setenv("INIT_NEVER_FIELD", "init never")

	defer os.Clearenv()

	type Nested struct {
		Field string
	}

	type Config struct {
		Default       string  `default:"default value"`
		ExpandDefault string  `expand:"true" default:"${EXPAND_OTHER}"`
		Expand        string  `expand:"true"`
		File          string  `file:"true"`
		ExpandFile    string  `expand:"true" file:"true"`
		Ignore        string  `ignore:"true"`
		InitValues    *Nested `init:"values"`
		InitAlways    *Nested `init:"always"`
		InitNever     *Nested `init:"never"`
	}

	cfg := Config{}
	if err := envcfg.Parse(&cfg); err != nil {
		panic(err)
	}

	fmt.Printf("Default: %v, ExpandDefault: %v, Expand: %v, File: %v, ExpandFile: %v, Ignore: %v, InitValues: %v, InitAlways: %v, InitNever: %v",
		cfg.Default, cfg.ExpandDefault, cfg.Expand, cfg.File, cfg.ExpandFile, cfg.Ignore, cfg.InitValues, cfg.InitAlways, cfg.InitNever,
	)
	// Output: Default: default value, ExpandDefault: expand other, Expand: expand other, File: ${FILE_CONTENTS}, ExpandFile: file contents, Ignore: , InitValues: &{init values}, InitAlways: &{}, InitNever: <nil>
}

func ExampleParse_options() {
	tmpfile, err := os.CreateTemp("", "test.txt")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString("${FILE_CONTENTS}")
	if err != nil {
		panic(err)
	}

	os.Setenv("EXPAND", "${EXPAND_OTHER}")
	os.Setenv("EXPAND_OTHER", "expand other")
	os.Setenv("FILE_CONTENTS", "file contents")
	os.Setenv("FILE", tmpfile.Name())
	os.Setenv("EXPAND_FILE", tmpfile.Name())
	os.Setenv("IGNORE", "ignore")
	os.Setenv("INIT_VALUES_FIELD", "init values")
	os.Setenv("INIT_NEVER_FIELD", "init never")

	defer os.Clearenv()

	type Nested struct {
		Field string
	}

	type Config struct {
		Default       string  `env:",default=default value"`
		ExpandDefault string  `env:",expand,default=${EXPAND_OTHER}"`
		Expand        string  `env:",expand"`
		File          string  `env:",file"`
		ExpandFile    string  `env:",expand,file"`
		Ignore        string  `env:"-"`
		InitValues    *Nested `env:",init=values"`
		InitAlways    *Nested `env:",init=always"`
		InitNever     *Nested `env:",init=never"`
	}

	cfg := Config{}
	err = envcfg.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Default: %v, ExpandDefault: %v, Expand: %v, File: %v, ExpandFile: %v, Ignore: %v, InitValues: %v, InitAlways: %v, InitNever: %v",
		cfg.Default, cfg.ExpandDefault, cfg.Expand, cfg.File, cfg.ExpandFile, cfg.Ignore, cfg.InitValues, cfg.InitAlways, cfg.InitNever,
	)
	// Output: Default: default value, ExpandDefault: expand other, Expand: expand other, File: ${FILE_CONTENTS}, ExpandFile: file contents, Ignore: , InitValues: &{init values}, InitAlways: &{}, InitNever: <nil>
}

func ExampleParse_validationNotempty() {
	os.Setenv("NOT_EMPTY", "")
	defer os.Clearenv()

	type Config struct {
		NotEmpty string `notempty:"true"`
	}

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	fmt.Printf("%+v\n", err)
	// Output: environment variable is empty: NOT_EMPTY
}

func ExampleParse_validationRequired() {
	defer os.Clearenv()

	type Config struct {
		Required string `required:"true"`
	}

	cfg := Config{}
	err := envcfg.Parse(&cfg)

	fmt.Printf("%+v\n", err)
	// Output: required field not found: Required
}
