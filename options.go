package config

import (
	"github.com/mitchellh/mapstructure"
	"strings"
)

// there are some event names for config data changed.
const (
	OnSetValue  = "set.value"
	OnSetData   = "set.data"
	OnLoadData  = "load.data"
	OnCleanData = "clean.data"
)

// HookFunc on config data changed.
type HookFunc func(event string, c *Config)

// Options config options
type Options struct {
	// ParseEnv parse env value. like: "${EnvName}" "${EnvName|default}"
	ParseEnv bool
	// ParseTime parses a duration string to time.Duration
	// eg: 10s, 2m
	ParseTime bool
	// Readonly config is readonly
	Readonly bool
	// EnableCache enable config data cache
	EnableCache bool
	// ParseKey parse key path, allow find value by key path. eg: 'key.sub' will find `map[key]sub`
	ParseKey bool
	// TagName tag name for binding data to struct
	// Deprecated: please set tag name by DecoderConfig
	TagName string
	// Delimiter the delimiter char for split key path, if `FindByPath=true`. default is '.'
	Delimiter byte
	// DumpFormat default write format
	DumpFormat string
	// ReadFormat default input format
	ReadFormat string
	// DecoderConfig setting for binding data to struct. such as: TagName
	DecoderConfig *mapstructure.DecoderConfig
	// HookFunc on data changed.
	HookFunc HookFunc
}

func newDefaultOption() *Options {
	return &Options{
		ParseKey:  true,
		TagName:   defaultStructTag,
		Delimiter: defaultDelimiter,
		// for export
		DumpFormat: JSON,
		ReadFormat: JSON,
		// struct decoder config
		DecoderConfig: newDefaultDecoderConfig(),
	}
}

func newDefaultDecoderConfig() *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		// tag name for binding struct
		TagName: defaultStructTag,
		// will auto convert string to int/uint
		WeaklyTypedInput: true,
		// DecodeHook: ValDecodeHookFunc,
	}
}

func (o *Options) shouldAddHookFunc() bool {
	return o.ParseTime || o.ParseEnv
}

/*************************************************************
 * config setting
 *************************************************************/

// ParseEnv set parse env value
func ParseEnv(opts *Options) { opts.ParseEnv = true }

// ParseTime set parse time string.
func ParseTime(opts *Options) { opts.ParseTime = true }

// Readonly set readonly
func Readonly(opts *Options) { opts.Readonly = true }

// Delimiter set delimiter char
func Delimiter(sep byte) func(*Options) {
	return func(opts *Options) {
		opts.Delimiter = sep
	}
}

// WithSetSaveFile set hook func
func WithSetSaveFile(fileName string, format string) func(options *Options) {
	return func(opts *Options) {
		opts.HookFunc = func(event string, c *Config) {
			if strings.HasPrefix(event, "set.") {
				err := c.DumpToFile(fileName, format)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

// WithHookFunc set hook func
func WithHookFunc(fn HookFunc) func(*Options) {
	return func(opts *Options) {
		opts.HookFunc = fn
	}
}

// EnableCache set readonly
func EnableCache(opts *Options) { opts.EnableCache = true }

// WithOptions with options
func WithOptions(opts ...func(*Options)) { dc.WithOptions(opts...) }

// WithOptions apply some options
func (c *Config) WithOptions(opts ...func(*Options)) *Config {
	if !c.IsEmpty() {
		panic("config: Cannot set options after data has been loaded")
	}

	// apply options
	for _, opt := range opts {
		opt(c.opts)
	}
	return c
}

// GetOptions get options
func GetOptions() *Options { return dc.Options() }

// Options get
func (c *Config) Options() *Options {
	return c.opts
}

// With apply some options
func (c *Config) With(fn func(c *Config)) *Config {
	fn(c)
	return c
}

// Readonly disable set data to config.
//
// Usage:
//
//	config.LoadFiles(a, b, c)
//	config.Readonly()
func (c *Config) Readonly() {
	c.opts.Readonly = true
}
