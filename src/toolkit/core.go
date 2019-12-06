package kit

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Time(arg ...string) int {
	if len(arg) == 0 {
		return Int(time.Now())
	}

	if len(arg) > 1 {
		if t, e := time.ParseInLocation(arg[1], arg[0], time.Local); e == nil {
			return Int(t)
		}
	}

	for _, v := range []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"01-02 15:04",
		"2006-01-02",
		"2006/01/02",
		"15:04:05",
	} {
		if t, e := time.ParseInLocation(v, arg[0], time.Local); e == nil {
			return Int(t)
		}
	}
	return 0
}
func Times(arg ...string) time.Time {
	return time.Unix(int64(Time(arg...)), 0)
}
func Duration(arg ...string) time.Duration {
	d, _ := time.ParseDuration(arg[0])
	return d
}
func Hash(arg ...interface{}) (string, []string) {
	if len(arg) == 0 {
		arg = append(arg, "uniq")
	}
	args := []string{}
	for _, v := range Trans(arg...) {
		switch v {
		case "time":
			args = append(args, Format(time.Now()))
		case "rand":
			args = append(args, Format(rand.Int()))
		case "uniq":
			args = append(args, Format(time.Now()))
			args = append(args, Format(rand.Int()))
		default:
			args = append(args, v)
		}
	}

	h := md5.Sum([]byte(strings.Join(args, "")))
	return hex.EncodeToString(h[:]), args
}
func Hashs(arg ...interface{}) string {
	h, _ := Hash(arg...)
	return h
}

func Select(value string, args ...interface{}) string {
	if len(args) == 0 {
		return value
	}

	switch arg := args[0].(type) {
	case string:
		if len(args) > 1 {
			switch b := args[1].(type) {
			case bool:
				if b && Right(arg) {
					return arg
				}
				return value
			}
		}
		if Right(arg) {
			return arg
		}
	case []interface{}:
		index := 0
		if len(args) > 1 {
			index = Int(args[1])
		}
		if index < len(arg) && Format(arg[index]) != "" {
			return Format(arg[index])
		}
	case []string:
		index := 0
		if len(args) > 1 {
			index = Int(args[1])
		}
		if index < len(arg) && arg[index] != "" {
			return arg[index]
		}
	default:
		if v := Format(args...); v != "" {
			return v
		}
	}
	return value
}
func Chains(root interface{}, args ...interface{}) string {
	return Format(Chain(root, args...))
}
func Chain(root interface{}, args ...interface{}) interface{} {
	for i := 0; i < len(args); i += 2 {
		if arg, ok := args[i].(map[string]interface{}); ok {
			argn := []interface{}{}
			for k, v := range arg {
				argn = append(argn, k, v)
			}
			argn = append(argn, args[i+1:])
			args, i = argn, -2
			continue
		}

		var parent interface{}
		parent_key, parent_index := "", 0

		keys := []string{}
		for _, v := range Trans(args[i]) {
			keys = append(keys, strings.Split(v, ".")...)
		}

		data := root
		for j, key := range keys {
			index, e := strconv.Atoi(key)

			var next interface{}
			switch value := data.(type) {
			case nil:
				if i == len(args)-1 {
					return nil
				}
				if j == len(keys)-1 {
					next = args[i+1]
				}

				if e == nil {
					data, index = []interface{}{next}, 0
				} else {
					data = map[string]interface{}{key: next}
				}
			case []string:
				index = (index+2+len(value)+2)%(len(value)+2) - 2

				if j == len(keys)-1 {
					if i == len(args)-1 {
						if index < 0 {
							return ""
						}
						return value[index]
					}
					next = args[i+1]
				}

				if index == -1 {
					data, index = append([]string{Format(next)}, value...), 0
				} else {
					data, index = append(value, Format(next)), len(value)
				}
				next = value[index]
			case map[string]string:
				if j == len(keys)-1 {
					if i == len(args)-1 {
						return value[key] // 读取数据
					}
					value[key] = Format(next) // 修改数据
				}
				next = value[key]
			case map[string]interface{}:
				if j == len(keys)-1 {
					if i == len(args)-1 {
						return value[key] // 读取数据
					}
					value[key] = args[i+1] // 修改数据
					if !Right(args[i+1]) {
						delete(value, key)
					}
				}
				next = value[key]
			case []interface{}:
				index = (index+2+len(value)+2)%(len(value)+2) - 2

				if j == len(keys)-1 {
					if i == len(args)-1 {
						if index < 0 {
							return nil
						}
						return value[index] // 读取数据
					}
					next = args[i+1] // 修改数据
				}

				if index == -1 {
					value, index = append([]interface{}{next}, value...), 0
				} else if index == -2 {
					value, index = append(value, next), len(value)
				} else if j == len(keys)-1 {
					value[index] = next
				}
				data, next = value, value[index]
			}

			switch p := parent.(type) {
			case map[string]interface{}:
				p[parent_key] = data
			case []interface{}:
				p[parent_index] = data
			case nil:
				root = data
			}

			parent, data = data, next
			parent_key, parent_index = key, index
		}
	}

	return root
}
func Array(list []string, index int, arg ...interface{}) []string {
	if len(arg) == 0 {
		if -1 < index && index < len(list) {
			return []string{list[index]}
		}
		return []string{""}
	}

	str := Trans(arg...)

	index = (index+2)%(len(list)+2) - 2
	if index == -1 {
		list = append(str, list...)
	} else if index == -2 {
		list = append(list, str...)
	} else {
		if index < -2 {
			index += len(list) + 2
		}
		if index < 0 {
			index = 0
		}

		for i := len(list); i < index+len(str); i++ {
			list = append(list, "")
		}
		for i := 0; i < len(str); i++ {
			list[index+i] = str[i]
		}
	}

	return list
}
func Width(str string, mul int) int {
	return len([]rune(str)) + (len(str)-len([]rune(str)))/2/mul
}
func View(args []string, conf map[string]interface{}) []string {
	if len(args) == 0 {
		args = append(args, "default")
	}

	keys := []string{}
	for _, k := range args {
		if v, ok := conf[k]; ok {
			keys = append(keys, Trans(v)...)
		} else {
			keys = append(keys, k)
		}
	}
	return keys
}
func Map(v interface{}, random string, args ...interface{}) map[string]interface{} {
	table, _ := v.([]interface{})
	value, _ := v.(map[string]interface{})
	if len(args) == 0 {
		return value
	}

	switch fun := args[0].(type) {
	case func(int, string):
		for i, v := range table {
			fun(i, Format(v))
		}
	case func(int, string) bool:
		for i, v := range table {
			if fun(i, Format(v)) {
				break
			}
		}
	case func(string, string):
		for k, v := range value {
			fun(k, Format(v))
		}
	case func(string, string) bool:
		for k, v := range value {
			if fun(k, Format(v)) {
				break
			}
		}
	case func(map[string]interface{}):
		if len(value) == 0 {
			return nil
		}
		fun(value)

	case func(int, map[string]interface{}):
		for i := 0; i < len(table); i++ {
			if val, ok := table[i].(map[string]interface{}); ok {
				fun(i, val)
			}
		}
	case func(string, []interface{}):
		for k, v := range value {
			if val, ok := v.([]interface{}); ok {
				fun(k, val)
			}
		}
	case func(string, int, string):
		for k, v := range value {
			if val, ok := v.([]interface{}); ok {
				for i, v := range val {
					fun(k, i, Format(v))
				}
			}
		}
	case func(string, map[string]interface{}):
		switch random {
		case "%":
			n, i := rand.Intn(len(value)), 0
			for k, v := range value {
				if val, ok := v.(map[string]interface{}); i == n && ok {
					fun(k, val)
					break
				}
				i++
			}
		case "*":
			fallthrough
		default:
			for k, v := range value {
				if val, ok := v.(map[string]interface{}); ok {
					fun(k, val)
				}
			}
		}
	case func(string, map[string]interface{}) bool:
		for k, v := range value {
			if val, ok := v.(map[string]interface{}); ok {
				if fun(k, val) {
					break
				}
			}
		}
	case func(string, int, map[string]interface{}):
		keys := make([]string, 0, len(value))
		for k, _ := range value {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := value[k]
			if val, ok := v.([]interface{}); ok {
				for i, v := range val {
					if val, ok := v.(map[string]interface{}); ok {
						fun(k, i, val)
					}
				}
			}
		}
	case func(key string, meta map[string]interface{}, index int, value map[string]interface{}):
		keys := make([]string, 0, len(value))
		for k, _ := range value {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := value[k].(map[string]interface{})
			meta := v["meta"].(map[string]interface{})
			list := v["list"].([]interface{})
			for i, v := range list {
				if val, ok := v.(map[string]interface{}); ok {
					fun(k, meta, i, val)
				}
			}
		}
	case func(meta map[string]interface{}, index int, value map[string]interface{}):
		meta := value["meta"].(map[string]interface{})
		list := value["list"].([]interface{})

		for i := 0; i < len(list); i++ {
			if val, ok := list[i].(map[string]interface{}); ok {
				fun(meta, i, val)
			}
		}
	}
	return value
}
