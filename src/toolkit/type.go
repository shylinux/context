package kit

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

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
func Int64(arg ...interface{}) int64 {
	var result int64
	for _, v := range arg {
		switch val := v.(type) {
		case int:
			result += int64(val)
		case int8:
			result += int64(val)
		case int16:
			result += int64(val)
		case int64:
			result += int64(val)
		case uint16:
			result += int64(val)
		case uint32:
			result += int64(val)
		case uint64:
			result += int64(val)
		case float64:
			result += int64(val)
		case byte: // uint8
			result += int64(val)
		case rune: // int32
			result += int64(val)
		case string:
			if i, e := strconv.ParseInt(val, 10, 64); e == nil {
				result += i
			}
		case bool:
			if val {
				result += 1
			}
		case time.Time:
			result += int64(val.Unix())
		case []string:
			result += int64(len(val))
		case map[string]string:
			result += int64(len(val))
		case []interface{}:
			result += int64(len(val))
		case map[string]interface{}:
			result += int64(len(val))
		}
	}
	return result
}
func Int(arg ...interface{}) int {
	return int(Int64(arg...))
}
func Key(name string) string {
	return strings.Replace(name, ".", "_", -1)
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
		case []byte:
			result = append(result, string(val))
		case []rune:
			result = append(result, string(val))
		case time.Time:
			result = append(result, fmt.Sprintf("%s", val.Format("2006-01-02 15:03:04")))
		case *os.File:
			if s, e := val.Stat(); e == nil {
				result = append(result, fmt.Sprintf("%T [name: %s]", v, s.Name()))
			} else {
				result = append(result, fmt.Sprintf("%T", v))
			}
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
				ls = append(ls, Trans(v)...)
			}
		default:
			ls = append(ls, Format(val))
		}
	}
	return ls
}
func Struct(arg ...interface{}) map[string]interface{} {
	value := map[string]interface{}{}
	if len(arg) == 0 {
		return value
	}
	switch val := arg[0].(type) {
	case map[string]interface{}:
		return val
	case string:
		json.Unmarshal([]byte(val), value)
	}

	return value
}
func Structm(args ...interface{}) map[string]interface{} {
	value := Struct(args...)
	for _, arg := range args {
		switch val := arg.(type) {
		case func(k string, v string):
			for k, v := range value {
				val(k, Format(v))
			}
		}
	}
	return value
}
