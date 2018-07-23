package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGet(t *testing.T) {
	st := assert.New(t)

	ClearAll()
	err := LoadStrings(Json, jsonStr)
	st.Nil(err)

	// fmt.Printf("%#v\n", Data())
	c := Default()

	// error on get
	_, ok := c.Get("")
	st.False(ok)

	_, ok = c.Get("notExist")
	st.False(ok)

	_, ok = c.Get("map1.key", false)
	st.False(ok)

	val, ok := Get("map1.notExist")
	st.False(ok)
	st.Nil(val)

	val, ok = c.Get("arr1.100")
	st.False(ok)
	st.Nil(val)

	val, ok = c.Get("arr1.notExist")
	st.False(ok)
	st.Nil(val)

	// get int
	val, ok = Get("age")
	st.True(ok)
	st.Equal(float64(123), val)
	st.Equal("float64", fmt.Sprintf("%T", val))

	iv, ok := Int("age")
	st.True(ok)
	st.Equal(123, iv)

	iv = DefInt("notExist", 34)
	st.Equal(34, iv)

	iv = c.MustInt("age")
	st.Equal(123, iv)
	iv = c.MustInt("notExist")
	st.Equal(0, iv)

	// get int64
	iv64, ok := Int64("age")
	st.True(ok)
	st.Equal(int64(123), iv64)

	iv64 = DefInt64("age", 34)
	st.Equal(int64(123), iv64)
	iv64 = DefInt64("notExist", 34)
	st.Equal(int64(34), iv64)

	iv64 = c.MustInt64("age")
	st.Equal(int64(123), iv64)
	iv64 = c.MustInt64("notExist")
	st.Equal(int64(0), iv64)

	// get bool
	val, ok = Get("debug")
	st.True(ok)
	st.Equal(true, val)

	bv, ok := Bool("debug")
	st.True(ok)
	st.Equal(true, bv)

	bv, ok = Bool("age")
	st.False(ok)
	st.Equal(false, bv)

	bv = DefBool("debug", false)
	st.Equal(true, bv)

	bv = DefBool("notExist", false)
	st.Equal(false, bv)

	bv = c.MustBool("debug")
	st.True(bv)
	bv = c.MustBool("notExist")
	st.False(bv)

	// get string
	val, ok = Get("name")
	st.True(ok)
	st.Equal("app", val)

	str, ok := String("notExists")
	st.False(ok)
	st.Equal("", str)

	str = DefString("notExists", "defVal")
	st.Equal("defVal", str)

	str = c.MustString("name")
	st.Equal("app", val)
	str = c.MustString("notExist")
	st.Equal("", str)

	// get string array
	arr, ok := Strings("notExist")
	st.False(ok)

	arr, ok = Strings("arr1")
	st.True(ok)
	st.Equal(`[]string{"val", "val1", "val2"}`, fmt.Sprintf("%#v", arr))

	val, ok = String("arr1.1")
	st.True(ok)
	st.Equal("val1", val)

	err = LoadStrings(Json, `{
"iArr": [12, 34, 36],
"iMap": {"k1": 12, "k2": 34, "k3": 36}
}`)
	st.Nil(err)

	// get int arr
	iarr, ok := Ints("notExist")
	st.False(ok)

	iarr, ok = Ints("iArr")
	st.True(ok)
	st.Equal(`[]int{12, 34, 36}`, fmt.Sprintf("%#v", iarr))

	iv, ok = Int("iArr.1")
	st.True(ok)
	st.Equal(34, iv)

	iv, ok = Int("iArr.100")
	st.False(ok)

	// get int map
	imp, ok := IntMap("notExist")
	st.False(ok)

	imp, ok = IntMap("iMap")
	st.True(ok)
	st.NotEmpty(imp)

	iv, ok = Int("iMap.k2")
	st.True(ok)
	st.Equal(34, iv)

	iv, ok = Int("iMap.notExist")
	st.False(ok)

	// get string map
	smp, ok := StringMap("map1")
	st.True(ok)
	st.Equal("val1", smp["key1"])

	// like load from yaml content
	// c = New("test")
	err = c.LoadData(map[string]interface{}{
		"newIArr": []int{2, 3},
		"newSArr": []string{"a", "b"},
		"yMap": map[interface{}]interface{}{
			"k0": "v0",
			"k1": 23,
		},
		"yMap1": map[interface{}]interface{}{
			"k": "v",
			"k1": 23,
			"k2": []interface{}{23, 45},
		},
		"yArr": []interface{}{23, 45, "val", map[string]interface{}{"k4": "v4"}},
	})
	st.Nil(err)

	iarr,ok = Ints("newIArr")
	st.True(ok)
	st.Equal("[2 3]", fmt.Sprintf("%v", iarr))

	iv, ok = Int("newIArr.1")
	st.True(ok)
	st.Equal(3, iv)

	iv, ok = Int("newIArr.200")
	st.False(ok)

	val, ok = String("newSArr.1")
	st.True(ok)
	st.Equal("b", val)

	val, ok = String("newSArr.100")
	st.False(ok)

	smp, ok = StringMap("yMap")
	st.True(ok)
	st.Equal("v0", smp["k0"])

	iarr, ok = Ints("yMap1.k2")
	st.True(ok)
	st.Equal("[23 45]", fmt.Sprintf("%v", iarr))
}

type user struct {
	Age    int
	Name   string
	Sports []string
}

func TestConfig_MapStructure(t *testing.T) {
	st := assert.New(t)

	cfg := New("test")
	err := cfg.LoadStrings(Json, `{
"age": 28,
"name": "inhere",
"sports": ["pingPong", "跑步"]
}`)

	st.Nil(err)

	user := &user{}
	err = cfg.MapStructure("", user)
	st.Nil(err)

	st.Equal(28, user.Age)
	st.Equal("inhere", user.Name)
	st.Equal("pingPong", user.Sports[0])
}

func TestEnableCache(t *testing.T) {
	at := assert.New(t)

	c := NewWithOptions("test", EnableCache)
	err := c.LoadStrings(Json, jsonStr)
	at.Nil(err)

	str, ok := c.String("name")
	at.True(ok)
	at.Equal("app", str)

	// re-get, from caches
	str, ok = c.String("name")
	at.True(ok)
	at.Equal("app", str)

	sArr, ok := c.Strings("arr1")
	at.True(ok)
	at.Equal("app", str)

	// re-get, from caches
	sArr, ok = c.Strings("arr1")
	at.True(ok)
	at.Equal("val1", sArr[1])

	sMap, ok := c.StringMap("map1")
	at.True(ok)
	at.Equal("val1", sMap["key1"])
	sMap, ok = c.StringMap("map1")
	at.True(ok)
	at.Equal("val1", sMap["key1"])
}
