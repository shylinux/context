package kit

import (
	"fmt"
	"strconv"
	"strings"

	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"
)

type TERM interface {
	Show(...interface{}) bool
}

var STDIO TERM

func Split(str string, n int) []string {
	res := []string{}
	for i, j := 0, 0; i < len(str); i++ {
		if str[i] == ' ' {
			continue
		}
		for j = i; j < len(str); j++ {
			if str[j] == ' ' {
				break
			}
		}
		if n == len(res)+1 {
			j = len(str)
		}
		res, i = append(res, str[i:j]), j
	}
	return res
}

func Width(str string, mul int) int {
	return len([]rune(str)) + (len(str)-len([]rune(str)))/2/mul
}
func Len(arg interface{}) int {
	switch arg := arg.(type) {
	case []interface{}:
		return len(arg)
	case map[string]interface{}:
		return len(arg)
	}
	return 0
}
func Simple(str string) string {
	return strings.Replace(strings.TrimSpace(str), "\n", "\\n", -1)
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

func Link(name string, url string) string {
	return fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>", url, name)
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
