package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	errInvalidKey = errors.New("invalid config key string")
	// errNotFound   = errors.New("this key does not exist in the configuration data")
)

// Exists key exists check
func Exists(key string, findByPath ...bool) (ok bool) {
	return dc.Exists(key, findByPath...)
}

// Exists key exists check
func (c *Config) Exists(key string, findByPath ...bool) (ok bool) {
	if key = formatKey(key); key == "" {
		return
	}

	if _, ok = c.data[key]; ok {
		return
	}

	// disable find by path.
	if len(findByPath) > 0 && !findByPath[0] {
		return
	}

	// has sub key? eg. "lang.dir"
	if !strings.Contains(key, ".") {
		return
	}

	keys := strings.Split(key, ".")
	topK := keys[0]

	// find top item data based on top key
	var item interface{}
	if item, ok = c.data[topK]; !ok {
		return
	}
	for _, k := range keys[1:] {
		switch typeData := item.(type) {
		case map[string]string: // is map(from Set)
			if item, ok = typeData[k]; !ok {
				return
			}
		case map[string]interface{}: // is map(decode from toml/json)
			if item, ok = typeData[k]; !ok {
				return
			}
		case map[interface{}]interface{}: // is map(decode from yaml)
			if item, ok = typeData[k]; !ok {
				return
			}
		case []int: // is array(is from Set)
			i, err := strconv.Atoi(k)

			// check slice index
			if err != nil || len(typeData) < i {
				ok = false
				return
			}
		case []string: // is array(is from Set)
			i, err := strconv.Atoi(k)
			if err != nil || len(typeData) < i {
				ok = false
				return
			}
		case []interface{}: // is array(load from file)
			i, err := strconv.Atoi(k)
			if err != nil || len(typeData) < i {
				ok = false
				return
			}
		default: // error
			ok = false
			return
		}
	}
	return true
}

/*************************************************************
 * read config data
 *************************************************************/

// Get config value by key string, support get sub-value by key path(eg. 'map.key'),
// ok is true, find value from config
// ok is false, not found or error
func Get(key string, findByPath ...bool) interface{} { return dc.Get(key, findByPath...) }

// Get config value by key
func (c *Config) Get(key string, findByPath ...bool) interface{} {
	val, _ := c.GetValue(key, findByPath...)
	return val
}

// GetValue get value by given key string.
func GetValue(key string, findByPath ...bool) (value interface{}, ok bool) {
	return dc.GetValue(key, findByPath...)
}

// GetValue get value by given key string.
func (c *Config) GetValue(key string, findByPath ...bool) (value interface{}, ok bool) {
	key = formatKey(key)
	if key == "" {
		c.addError(errInvalidKey)
		return
	}

	// if not is readonly
	if !c.opts.Readonly {
		c.lock.RLock()
		defer c.lock.RUnlock()
	}

	// is top key
	if value, ok = c.data[key]; ok {
		return
	}

	// disable find by path.
	if len(findByPath) > 0 && !findByPath[0] {
		// c.addError(errNotFound)
		return
	}

	// has sub key? eg. "lang.dir"
	if !strings.Contains(key, ".") {
		// c.addError(errNotFound)
		return
	}

	keys := strings.Split(key, ".")
	topK := keys[0]

	// find top item data based on top key
	var item interface{}
	if item, ok = c.data[topK]; !ok {
		// c.addError(errNotFound)
		return
	}

	// find child
	// NOTICE: don't merge case, will result in an error.
	// e.g. case []int, []string
	// OR
	// case []int:
	// case []string:
	for _, k := range keys[1:] {
		switch typeData := item.(type) {
		case map[string]string: // is map(from Set)
			if item, ok = typeData[k]; !ok {
				return
			}
		case map[string]interface{}: // is map(decode from toml/json)
			if item, ok = typeData[k]; !ok {
				return
			}
		case map[interface{}]interface{}: // is map(decode from yaml)
			if item, ok = typeData[k]; !ok {
				return
			}
		case []int: // is array(is from Set)
			i, err := strconv.Atoi(k)

			// check slice index
			if err != nil || len(typeData) < i {
				ok = false
				c.addError(err)
				return
			}

			item = typeData[i]
		case []string: // is array(is from Set)
			i, err := strconv.Atoi(k)
			if err != nil || len(typeData) < i {
				ok = false
				c.addError(err)
				return
			}

			item = typeData[i]
		case []interface{}: // is array(load from file)
			i, err := strconv.Atoi(k)
			if err != nil || len(typeData) < i {
				ok = false
				c.addError(err)
				return
			}

			item = typeData[i]
		default: // error
			ok = false
			c.addErrorf("cannot get value of the key '%s'", key)
			return
		}
	}

	return item, true
}

/*************************************************************
 * read config (basic data type)
 *************************************************************/

// String get a string by key
func String(key string, defVal ...string) string { return dc.String(key, defVal...) }

// String get a string by key, if not found return default value
func (c *Config) String(key string, defVal ...string) (value string) {
	var ok bool

	value, ok = c.getString(key)
	if !ok && len(defVal) > 0 { // give default value
		value = defVal[0]
	}
	return
}

func (c *Config) getString(key string) (value string, ok bool) {
	// find from cache
	if c.opts.EnableCache && len(c.strCache) > 0 {
		value, ok = c.strCache[key]
		if ok {
			return
		}
	}

	val, ok := c.GetValue(key)
	if !ok {
		return
	}

	switch val.(type) {
	// from json int always is float64
	case bool, int, uint, int8, uint8, int16, uint16, int32, uint64, int64, float32, float64:
		value = fmt.Sprintf("%v", val)
	case string:
		value = fmt.Sprintf("%v", val)

		if c.opts.ParseEnv {
			value = c.parseEnvValue(value)
		}
	default:
		value = fmt.Sprintf("%v", val)
	}

	// add cache
	if ok && c.opts.EnableCache {
		if c.strCache == nil {
			c.strCache = make(map[string]string)
		}
		c.strCache[key] = value
	}
	return
}

// Int get a int by key
func Int(key string, defVal ...int) int { return dc.Int(key, defVal...) }

// Int get a int value, if not found return default value
func (c *Config) Int(key string, defVal ...int) (value int) {
	rawVal, exist := c.getString(key)
	if !exist {
		if len(defVal) > 0 {
			return defVal[0]
		}
		return
	}

	value, err := strconv.Atoi(rawVal)
	if err != nil {
		c.addError(err)
	}
	return
}

// Int64 get a int value, if not found return default value
func Int64(key string, defVal ...int64) int64 { return dc.Int64(key, defVal...) }

// Int64 get a int value, if not found return default value
func (c *Config) Int64(key string, defVal ...int64) (value int64) {
	rawVal, exist := c.getString(key)
	if !exist {
		if len(defVal) > 0 {
			return defVal[0]
		}
		return
	}

	value, err := strconv.ParseInt(rawVal, 10, 0)
	if err != nil {
		c.addError(err)
	}
	return
}

// Float get a float64 value, if not found return default value
func Float(key string, defVal ...float64) float64 { return dc.Float(key, defVal...) }

// Float get a float64 by key
func (c *Config) Float(key string, defVal ...float64) (value float64) {
	str, ok := c.getString(key)
	if !ok {
		if len(defVal) > 0 {
			return defVal[0]
		}
		return
	}

	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		c.addError(err)
	}
	return
}

// Bool get a bool value, if not found return default value
func Bool(key string, defVal ...bool) bool { return dc.Bool(key, defVal...) }

// Bool looks up a value for a key in this section and attempts to parse that value as a boolean,
// along with a boolean result similar to a map lookup.
// of following(case insensitive):
//  - true
//  - yes
//  - false
//  - no
//  - 1
//  - 0
// The `ok` boolean will be false in the event that the value could not be parsed as a bool
func (c *Config) Bool(key string, defVal ...bool) (value bool) {
	rawVal, ok := c.getString(key)
	if !ok {
		if len(defVal) > 0 {
			return defVal[0]
		}
		return
	}

	lowerCase := strings.ToLower(rawVal)
	switch lowerCase {
	case "", "0", "false", "no":
		value = false
	case "1", "true", "yes":
		value = true
	default:
		c.addErrorf("the value '%s' cannot be convert to bool", lowerCase)
	}
	return
}

/*************************************************************
 * read config (complex data type)
 *************************************************************/

// Ints get config data as a int slice/array
func Ints(key string) []int { return dc.Ints(key) }

// Ints get config data as a int slice/array
func (c *Config) Ints(key string) (arr []int) {
	rawVal, ok := c.GetValue(key)
	if !ok {
		return
	}

	switch typeData := rawVal.(type) {
	case []int:
		arr = typeData
	case []interface{}:
		for _, v := range typeData {
			// iv, err := strconv.Atoi(v.(string))
			iv, err := strconv.Atoi(fmt.Sprintf("%v", v))
			if err != nil {
				ok = false
				c.addError(err)
				return
			}

			arr = append(arr, iv)
		}
	default:
		c.addErrorf("value cannot be convert to []int, key is '%s'", key)
	}
	return
}

// IntMap get config data as a map[string]int
func IntMap(key string) map[string]int { return dc.IntMap(key) }

// IntMap get config data as a map[string]int
func (c *Config) IntMap(key string) (mp map[string]int) {
	rawVal, ok := c.GetValue(key)
	if !ok {
		return
	}

	switch typeData := rawVal.(type) {
	case map[string]int: // from Set
		mp = typeData
	case map[string]interface{}: // decode from json,toml
		mp = make(map[string]int)
		for k, v := range typeData {
			iv, err := strconv.Atoi(fmt.Sprintf("%v", v))
			if err != nil {
				c.addError(err)
				return
			}
			mp[k] = iv
		}
	case map[interface{}]interface{}: // if decode from yaml
		mp = make(map[string]int)
		for k, v := range typeData {
			iv, err := strconv.Atoi(fmt.Sprintf("%v", v))
			if err != nil {
				c.addError(err)
				return
			}

			sk := fmt.Sprintf("%v", k)
			mp[sk] = iv
		}
	default:
		c.addErrorf("value cannot be convert to map[string]int, key is '%s'", key)
	}
	return
}

// Strings get strings by key
func Strings(key string) []string { return dc.Strings(key) }

// Strings get config data as a string slice/array
func (c *Config) Strings(key string) (arr []string) {
	var ok bool
	// find from cache
	if c.opts.EnableCache && len(c.sArrCache) > 0 {
		arr, ok = c.sArrCache[key]
		if ok {
			return
		}
	}

	rawVal, ok := c.GetValue(key)
	if !ok {
		return
	}

	switch typeData := rawVal.(type) {
	case []string:
		arr = typeData
	case []interface{}:
		for _, v := range typeData {
			arr = append(arr, fmt.Sprintf("%v", v))
		}
	default:
		c.addErrorf("value cannot be convert to []string, key is '%s'", key)
		return
	}

	// add cache
	if c.opts.EnableCache {
		if c.sArrCache == nil {
			c.sArrCache = make(map[string]strArr)
		}
		c.sArrCache[key] = arr
	}
	return
}

// StringMap get config data as a map[string]string
func StringMap(key string) map[string]string { return dc.StringMap(key) }

// StringMap get config data as a map[string]string
func (c *Config) StringMap(key string) (mp map[string]string) {
	var ok bool

	// find from cache
	if c.opts.EnableCache && len(c.sMapCache) > 0 {
		mp, ok = c.sMapCache[key]
		if ok {
			return
		}
	}

	rawVal, ok := c.GetValue(key)
	if !ok {
		return
	}

	switch typeData := rawVal.(type) {
	case map[string]string: // from Set
		mp = typeData
	case map[string]interface{}: // decode from json,toml
		mp = make(map[string]string)
		for k, v := range typeData {
			mp[k] = fmt.Sprintf("%v", v)
		}
	case map[interface{}]interface{}: // if decode from yaml
		mp = make(map[string]string)
		for k, v := range typeData {
			sk := fmt.Sprintf("%v", k)
			mp[sk] = fmt.Sprintf("%v", v)
		}
	default:
		c.addErrorf("value cannot be convert to map[string]string, key is '%s'", key)
		return
	}

	// add cache
	if c.opts.EnableCache {
		if c.sMapCache == nil {
			c.sMapCache = make(map[string]strMap)
		}
		c.sMapCache[key] = mp
	}
	return
}

// MapStruct alias method of the 'Structure'
func MapStruct(key string, v interface{}) error { return dc.Structure(key, v) }

// MapStruct alias method of the 'Structure'
func (c *Config) MapStruct(key string, v interface{}) (err error) {
	return c.Structure(key, v)
}

// MapStructure alias method of the 'Structure'
func (c *Config) MapStructure(key string, v interface{}) (err error) {
	return c.Structure(key, v)
}

// Structure get config data and map to a structure.
// usage:
// 	dbInfo := Db{}
// 	config.Structure("db", &dbInfo)
func (c *Config) Structure(key string, v interface{}) (err error) {
	var ok bool
	var data interface{}

	// map all data
	if key == "" {
		ok = true
		data = c.data
	} else {
		data, ok = c.GetValue(key)
	}

	if ok {
		blob, err := JSONEncoder(data)
		if err != nil {
			return err
		}

		err = JSONDecoder(blob, v)
	}
	return
}

// parse env value, eg: "${SHELL}" ${NotExist|defValue}
var envRegex = regexp.MustCompile(`\${([\w-| ]+)}`)

// parse Env Value
func (c *Config) parseEnvValue(val string) string {
	if strings.Index(val, "${") == -1 {
		return val
	}

	// nodes like: ${VAR} -> [${VAR}]
	// val = "${GOPATH}/${APP_ENV | prod}/dir" -> [${GOPATH} ${APP_ENV | prod}]
	vars := envRegex.FindAllString(val, -1)
	if len(vars) == 0 {
		return val
	}

	var oldNew []string
	var name, def string
	for _, fVar := range vars {
		ss := strings.SplitN(fVar[2:len(fVar)-1], "|", 2)

		// has default ${NotExist|defValue}
		if len(ss) == 2 {
			name, def = strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])
		} else {
			def = fVar
			name = ss[0]
		}

		envVal := os.Getenv(name)
		if envVal == "" {
			envVal = def
		}

		oldNew = append(oldNew, fVar, envVal)
	}

	return strings.NewReplacer(oldNew...).Replace(val)
}

// format key
func formatKey(key string) string {
	return strings.Trim(strings.TrimSpace(key), ".")
}
