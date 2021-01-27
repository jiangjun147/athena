package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	"github.com/rickone/athena/common"
	"gopkg.in/yaml.v2"
)

var (
	kvs = map[interface{}]interface{}{}
	mu  = sync.RWMutex{}
)

func Init(path ...string) {
	var configPath string
	if len(path) > 0 {
		configPath = path[0]
	} else {
		env := os.Getenv("ENV")
		if env == "" {
			env = "test"
		}

		configPath = fmt.Sprintf("./conf/%s.yml", env)
		if os.Getenv("DEBUG") != "" {
			configPath = fmt.Sprintf("../../conf/%s.yml", env)
		}
	}

	data, err := ioutil.ReadFile(configPath)
	common.AssertError(err)

	err = yaml.Unmarshal(data, &kvs)
	common.AssertError(err)

	go Watch()
}

type Value struct {
	val interface{}
}

func GetValue(fields ...interface{}) *Value {
	mu.RLock()
	defer mu.RUnlock()

	var v interface{} = kvs
	for _, field := range fields {
		switch val := v.(type) {
		case map[interface{}]interface{}:
			v = val[field]
		case []interface{}:
			v = val[field.(int)]
		}
	}

	if v == nil {
		return nil
	}
	return &Value{val: v}
}

func (v *Value) GetValue(fields ...interface{}) *Value {
	var sv interface{} = v.val
	for _, field := range fields {
		switch val := sv.(type) {
		case map[interface{}]interface{}:
			sv = val[field]
		case []interface{}:
			sv = val[field.(int)]
		}
	}

	if sv == nil {
		return nil
	}
	return &Value{val: sv}
}

func (v *Value) GetInt(field interface{}) int64 {
	if m, ok := v.val.(map[interface{}]interface{}); ok {
		f := m[field]
		if f != nil {
			switch n := f.(type) {
			case int:
				return int64(n)
			case int64:
				return n
			case string:
				i, _ := strconv.ParseInt(n, 10, 64)
				return i
			}
		}
	}
	return 0
}

func (v *Value) GetFloat(field interface{}) float64 {
	if m, ok := v.val.(map[interface{}]interface{}); ok {
		f := m[field]
		if f != nil {
			switch n := f.(type) {
			case int:
				return float64(n)
			case int64:
				return float64(n)
			case float32:
				return float64(n)
			case float64:
				return n
			case string:
				ft, _ := strconv.ParseFloat(n, 64)
				return ft
			}
		}
	}
	return 0
}

func (v *Value) GetString(field interface{}) string {
	if m, ok := v.val.(map[interface{}]interface{}); ok {
		f := m[field]
		if f != nil {
			switch s := f.(type) {
			case int:
				return strconv.Itoa(s)
			case int64:
				return strconv.FormatInt(s, 10)
			case string:
				return s
			}
		}
	}
	return ""
}

func (v *Value) ToMap() map[interface{}]interface{} {
	if m, ok := v.val.(map[interface{}]interface{}); ok {
		return m
	}
	return nil
}

func (v *Value) ToSlice() []interface{} {
	if s, ok := v.val.([]interface{}); ok {
		return s
	}
	return nil
}

func GetInt(fields ...interface{}) int64 {
	v := GetValue(fields...)
	if v == nil {
		return 0
	}

	switch n := v.val.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	}
	return 0
}

func GetFloat(fields ...interface{}) float64 {
	v := GetValue(fields...)
	if v == nil {
		return 0
	}

	switch n := v.val.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float32:
		return float64(n)
	case float64:
		return n
	case string:
		ft, _ := strconv.ParseFloat(n, 64)
		return ft
	}
	return 0
}

func GetString(fields ...interface{}) string {
	v := GetValue(fields...)
	if v == nil {
		return ""
	}
	return v.val.(string)
}

func updateValue(node map[interface{}]interface{}, key interface{}, val interface{}) {
	switch t := val.(type) {
	case map[interface{}]interface{}:
		inner, ok := node[key].(map[interface{}]interface{})
		if ok {
			for k, v := range t {
				updateValue(inner, k, v)
			}
		} else {
			node[key] = val
		}

	default:
		node[key] = val
	}
}

func UpdateValue(key interface{}, val interface{}) {
	mu.Lock()
	defer mu.Unlock()

	updateValue(kvs, key, val)
}
