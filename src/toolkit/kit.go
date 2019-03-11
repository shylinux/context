package kit

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"
)

func Errorf(str string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", str)
		return
	}
	fmt.Fprintf(os.Stderr, str, args...)
}
func Log(action string, str string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "%s", str)
	} else {
		fmt.Fprintf(os.Stderr, str, args...)
	}
	fmt.Fprintln(os.Stderr)
}

func Int(arg ...interface{}) int {
	result := 0
	for _, v := range arg {
		switch val := v.(type) {
		case int:
			result += val
		case int8:
			result += int(val)
		case int16:
			result += int(val)
		// case int32:
		// 	result += int(val)
		case int64:
			result += int(val)
		// case uint8:
		// 	result += int(val)
		case uint16:
			result += int(val)
		case uint32:
			result += int(val)
		case uint64:
			result += int(val)
		case float64:
			result += int(val)
		case byte: // uint8
			result += int(val)
		case rune: // int32
			result += int(val)
		case string:
			if i, e := strconv.Atoi(val); e == nil {
				result += i
			}
		case bool:
			if val {
				result += 1
			}
		case time.Time:
			result += int(val.Unix())
		case []string:
			result += len(val)
		case map[string]string:
			result += len(val)
		case []interface{}:
			result += len(val)
		case map[string]interface{}:
			result += len(val)
		}
	}
	return result
}
func Right(arg ...interface{}) bool {
	result := false
	for _, v := range arg {
		switch val := v.(type) {
		case int:
			result = result || val != 0
		case bool:
			result = result || val
		case string:
			switch val {
			case "", "0", "false", "off", "no", "error: ":
				result = result || false
			default:
				result = result || true
			}
		case error:
			result = result || false
		case []string:
			result = result || len(val) > 0
		case map[string]string:
			result = result || len(val) > 0
		case []interface{}:
			result = result || len(val) > 0
		case map[string]interface{}:
			result = result || len(val) > 0
		default:
			result = result || val != nil
		}
	}
	return result
}
func Format(arg ...interface{}) string {
	result := []string{}
	for _, v := range arg {
		switch val := v.(type) {
		case nil:
			result = result[:0]
		case int, int8, int16, int32, int64:
			result = append(result, fmt.Sprintf("%d", val))
		case uint, uint8, uint16, uint32, uint64:
			result = append(result, fmt.Sprintf("%d", val))
		case float64:
			result = append(result, fmt.Sprintf("%d", int(val)))
		case bool:
			result = append(result, fmt.Sprintf("%t", val))
		case string:
			result = append(result, val)
		case []rune:
			result = append(result, string(val))
		case []string:
			result = append(result, val...)
		case time.Time:
			result = append(result, fmt.Sprintf("%s", val.Format("2006-01-02 15:03:04")))
		case *os.File:
			if s, e := val.Stat(); e == nil {
				result = append(result, fmt.Sprintf("%T [name: %s]", v, s.Name()))
			} else {
				result = append(result, fmt.Sprintf("%T", v))
			}
		// case error:
		// 	result = append(result, fmt.Sprintf("%v", val))
		default:
			if b, e := json.Marshal(val); e == nil {
				result = append(result, string(b))
			}
		}
	}

	if len(result) > 1 {
		args := []interface{}{}
		if n := strings.Count(result[0], "%") - strings.Count(result[0], "%%"); len(result) > n {
			for i := 1; i < n+1; i++ {
				args = append(args, result[i])
			}
			return fmt.Sprintf(result[0], args...) + strings.Join(result[n+1:], "")
		} else if len(result) == n+1 {
			for i := 1; i < len(result); i++ {
				args = append(args, result[i])
			}
			return fmt.Sprintf(result[0], args...)
		}
	}
	return strings.Join(result, "")
}
func Formats(arg ...interface{}) string {
	result := []string{}
	for _, v := range arg {
		switch val := v.(type) {
		// case []interface{}:
		// 	for _, v := range val {
		// 		result = append(result, Format(v))
		// 	}
		default:
			if b, e := json.MarshalIndent(val, "", "  "); e == nil {
				result = append(result, string(b))
			} else {
				result = append(result, fmt.Sprintf("%#v", val))
			}
		}
	}
	return strings.Join(result, " ")
}
func Trans(arg ...interface{}) []string {
	ls := []string{}
	for _, v := range arg {
		switch val := v.(type) {
		// case *Message:
		// 	if val.Hand {
		// 		ls = append(ls, val.Meta["result"]...)
		// 	} else {
		// 		ls = append(ls, val.Meta["detail"]...)
		// 	}
		case nil:
		case []float64:
			for _, v := range val {
				ls = append(ls, fmt.Sprintf("%d", int(v)))
			}
		case []int:
			for _, v := range val {
				ls = append(ls, fmt.Sprintf("%d", v))
			}
		case []bool:
			for _, v := range val {
				ls = append(ls, fmt.Sprintf("%t", v))
			}
		case []string:
			ls = append(ls, val...)
		case map[string]string:
			for k, v := range val {
				ls = append(ls, k, v)
			}
		case map[string]interface{}:
			for k, v := range val {
				ls = append(ls, k, Format(v))
			}
		case []interface{}:
			for _, v := range val {
				ls = append(ls, Format(v))
			}
		default:
			ls = append(ls, Format(val))
		}
	}
	return ls
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
func Elect(last interface{}, args ...interface{}) string {
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case []string:
			index := 0
			if len(args) > 1 {
				switch a := args[1].(type) {
				case string:
					i, e := strconv.Atoi(a)
					if e == nil {
						index = i
					}
				case int:
					index = a
				}
			}

			if 0 <= index && index < len(arg) && arg[index] != "" {
				return arg[index]
			}
		case string:
			if arg != "" {
				return arg
			}
		}
	}

	switch l := last.(type) {
	case string:
		return l
	}
	return ""
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

func Link(name string, url string) string {
	return fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>", url, name)
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
func FileName(name string, meta ...string) string {
	result, app := strings.Split(name, "."), ""
	if len(result) > 1 {
		app, result = result[len(result)-1], result[:len(result)-1]
	}

	for _, v := range meta {
		switch v {
		case "year":
			result = append(result, "_", time.Now().Format("2006"))
		case "date":
			result = append(result, "_", time.Now().Format("0102"))
		case "time":
			result = append(result, "_", time.Now().Format("2006_0102_1504"))
		case "rand":
			result = append(result, "_", Format(rand.Int()))
		case "uniq":
			result = append(result, "_", Format(Time()))
			result = append(result, "_", Format(rand.Int()))
		}
	}

	if app != "" {
		result = append(result, ".", app)
	}
	return strings.Join(result, "")
}
func FmtSize(size uint64) string {
	if size > 1<<30 {
		return fmt.Sprintf("%d.%dG", size>>30, (size>>20)%1024*100/1024)
	}

	if size > 1<<20 {
		return fmt.Sprintf("%d.%dM", size>>20, (size>>10)%1024*100/1024)
	}

	if size > 1<<10 {
		return fmt.Sprintf("%d.%dK", size>>10, size%1024*100/1024)
	}

	return fmt.Sprintf("%dB", size)
}
func FmtNano(nano int64) string {
	if nano > 1000000000 {
		return fmt.Sprintf("%d.%ds", nano/1000000000, nano/100000000%100)
	}

	if nano > 1000000 {
		return fmt.Sprintf("%d.%dms", nano/100000, nano/100000%100)
	}

	if nano > 1000 {
		return fmt.Sprintf("%d.%dus", nano/1000, nano/100%100)
	}

	return fmt.Sprintf("%dns", nano)
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
			if s, e := os.Stat(v); e == nil && !s.IsDir() {
				if f, e := os.Open(v); e == nil {
					defer f.Close()
					m := md5.New()
					io.Copy(m, f)
					h := m.Sum(nil)
					args = append(args, hex.EncodeToString(h[:]))
					break
				}
			}
			args = append(args, v)
		}
	}

	h := md5.Sum([]byte(strings.Join(args, "")))
	return hex.EncodeToString(h[:]), args
}

func Block(root interface{}, args ...interface{}) interface{} {

	return root
}

func Check(e error) bool {
	if e != nil {
		panic(e)
	}
	return true
}
func DirWalk(file string, hand func(file string)) {
	s, e := os.Stat(file)
	Check(e)
	hand(file)

	if s.IsDir() {
		fs, e := ioutil.ReadDir(file)
		Check(e)

		for _, v := range fs {
			DirWalk(path.Join(file, v.Name()), hand)
		}
	}
}
