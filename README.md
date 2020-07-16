# config

[![GoDoc](https://godoc.org/github.com/gookit/config?status.svg)](https://godoc.org/github.com/gookit/config)
[![Build Status](https://travis-ci.org/gookit/config.svg?branch=master)](https://travis-ci.org/gookit/config)
[![Coverage Status](https://coveralls.io/repos/github/gookit/config/badge.svg?branch=master)](https://coveralls.io/github/gookit/config?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/config)](https://goreportcard.com/report/github.com/gookit/config)

Golang application config manage tool library. 

**The v2 version is recommended, the v1 version is no longer updated**

> **[中文说明](README_cn.md)**

- Support multi format: `JSON`(default), `INI`, `YAML`, `TOML`, `HCL`
  - `JSON` content support comments. will auto clear comments
- Support multi-file and multi-data loading
- Support for loading configuration data from remote URLs
- Support for setting configuration data from command line arguments(`flags`)
- Support data overlay and merge, automatically load by key when loading multiple copies of data
- Support get sub value by path, like `map.key` `arr.2`
- Support parse ENV name. like `envKey: ${SHELL}` -> `envKey: /bin/zsh`
- Generic api `Get` `Int` `String` `Bool` `Ints` `IntMap` `Strings` `StringMap` ...
- complete unit test(code coverage > 95%)

## Only use INI

> If you just want to use INI for simple config management, recommended use [gookit/ini](https://github.com/gookit/ini)

## GoDoc

- [godoc for github](https://godoc.org/github.com/gookit/config)

## Usage

Here using the yaml format as an example(`testdata/yml_other.yml`):

```yaml
name: app2
debug: false
baseKey: value2
shell: ${SHELL}
envKey1: ${NotExist|defValue}

map1:
    key: val2
    key2: val20

arr1:
    - val1
    - val21
```

### Load data

> examples code please see [_examples/yaml.go](_examples/yaml.go):

```go
package main

import (
    "github.com/gookit/config"
    "github.com/gookit/config/yaml"
)

// go run ./examples/yaml.go
func main() {
	config.WithOptions(config.ParseEnv)
	
	// add driver for support yaml content
	config.AddDriver(yaml.Driver)
	// config.SetDecoder(config.Yaml, yaml.Decoder)

	err := config.LoadFiles("testdata/yml_base.yml")
	if err != nil {
		panic(err)
	}

	// fmt.Printf("config data: \n %#v\n", config.Data())

	// load more files
	err = config.LoadFiles("testdata/yml_other.yml")
	// can also load multi at once
	// err := config.LoadFiles("testdata/yml_base.yml", "testdata/yml_other.yml")
	if err != nil {
		panic(err)
	}
}
```

### Read data

- get integer

```go
age, ok := config.Int("age")
fmt.Print(ok, age) // true 100
```

- Get bool

```go
val, ok := config.Bool("debug")
fmt.Print(ok, val) // true true
```

- Get string

```go
name, ok := config.String("name")
fmt.Print(ok, name) // true inhere
```

- Get strings(slice)

```go
arr1, ok := config.Strings("arr1")
fmt.Printf("%v %#v", ok, arr1) // true []string{"val1", "val21"}
```

- Get string map

```go
val, ok := config.StringMap("map1")
fmt.Printf("%v %#v",ok, val) // true map[string]string{"key":"val2", "key2":"val20"}
```

- Value contains ENV var

```go
value, ok := config.String("shell")
fmt.Print(ok, value) // true /bin/zsh
```

- Get value by key path

```go
// from array
value, ok := config.String("arr1.0")
fmt.Print(ok, value) // true "val1"

// from map
value, ok := config.String("map1.key")
fmt.Print(ok, value) // true "val2"
```

- Setting new value

```go
// set value
config.Set("name", "new name")
name, ok = config.String("name")
fmt.Print(ok, name) // true "new name"
```

## API Methods Refer

### Load Config

- `LoadData(dataSource ...interface{}) (err error)` Load from struts or maps
- `LoadFlags(keys []string) (err error)` Load from cli flags
- `LoadExists(sourceFiles ...string) (err error)` 
- `LoadFiles(sourceFiles ...string) (err error)`
- `LoadRemote(format, url string) (err error)`
- `LoadSources(format string, src []byte, more ...[]byte) (err error)`
- `LoadStrings(format string, str string, more ...string) (err error)`

### Getting Values

> `DefXXX` get value with default value

- `Bool(key string) (value bool, ok bool)`
- `DefBool(key string, defVal ...bool) bool`
- `Int(key string) (value int, ok bool)`
- `DefInt(key string, defVal ...int) int`
- `Int64(key string) (value int64, ok bool)`
- `DefInt64(key string, defVal ...int64)`
- `Ints(key string) (arr []int, ok bool)`
- `IntMap(key string) (mp map[string]int, ok bool)`
- `Float(key string) (value float64, ok bool)`
- `DefFloat(key string, defVal ...float64) float64`
- `String(key string) (value string, ok bool)`
- `DefString(key string, defVal ...string) string`
- `Strings(key string) (arr []string, ok bool)`
- `StringMap(key string) (mp map[string]string, ok bool)`
- `Get(key string, findByPath ...bool) (value interface{}, ok bool)`

### Setting Values

- `Set(key string, val interface{}, setByPath ...bool) (err error)`

### Useful Methods

- `AddDriver(driver Driver)`
- `Data() map[string]interface{}`
- `DumpTo(out io.Writer, format string) (n int64, err error)`

## Run Tests

```bash
go test -cover
// contains all sub-folder
go test -cover ./...
```

## Related Packages

- Ini parse [gookit/ini/parser](https://github.com/gookit/ini/tree/master/parser)
- Ini config [gookit/ini](https://github.com/gookit/ini)
- Yaml parse [go-yaml](https://github.com/go-yaml/yaml)
- Toml parse [go toml](https://github.com/BurntSushi/toml)
- Data merge [mergo](https://github.com/imdario/mergo)

### Ini Config Use

- [gookit/ini](https://github.com/gookit/ini) ini config manage

## License

**MIT**
