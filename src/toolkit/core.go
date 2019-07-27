package kit

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var DisableLog = false

func Pwd() string {
	wd, _ := os.Getwd()
	return wd
}
func Env(key string) {
	os.Getenv(key)
}
func Log(action string, str string, args ...interface{}) {
	if DisableLog {
		return
	}

	if len(args) > 0 {
		str = fmt.Sprintf(str, args...)
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", action, str)
}
func Errorf(str string, args ...interface{}) {
	Log("error", str, args...)
}
func Debugf(str string, args ...interface{}) {
	Log("debug", str, args...)
}

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
		"2006-01-02",
		"2006/01/02",
		"01-02 15:04",
	} {
		if t, e := time.ParseInLocation(v, arg[0], time.Local); e == nil {
			return Int(t)
		}
	}
	return 0
}
func Duration(arg ...string) time.Duration {
	d, _ := time.ParseDuration(arg[0])
	return d
}
func Hash(arg ...interface{}) (string, []string) {
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

			// Log("error", "chain [%v %v] [%v %v] [%v/%v %v/%v] %v", parent_key, parent_index, key, index, i, len(args), j, len(keys), data)

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
func Chains(root interface{}, args ...interface{}) string {
	return Format(Chain(root, args...))
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
				if b && arg != "" {
					return arg
				}
				return value
			}
		}
		if arg != "" {
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
func Slice(arg []string, args ...interface{}) ([]string, string) {
	if len(arg) == 0 {
		return arg, ""
	}
	if len(args) == 0 {
		return arg[1:], arg[0]
	}

	result := ""
	switch v := args[0].(type) {
	case int:
	case string:
		if arg[0] == v && len(arg) > 1 {
			return arg[2:], arg[1]
		}
		if len(args) > 1 {
			return arg, Format(args[1])
		}
	}

	return arg, result
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

func Create(p string) (*os.File, string, error) {
	if dir, _ := path.Split(p); dir != "" {
		if e := os.MkdirAll(dir, 0777); e != nil {
			return nil, p, e
		}
	}
	f, e := os.Create(p)
	return f, p, e
}
