package kit

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type TERM interface {
	Show(...interface{}) bool
}

var STDIO TERM
var DisableLog = false

func Log(action string, str string, args ...interface{}) {
	if DisableLog {
		return
	}

	if len(args) > 0 {
		str = fmt.Sprintf(str, args...)
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", action, str)
}
func Env(key string) {
	os.Getenv(key)
}
func Pwd() string {
	wd, _ := os.Getwd()
	return wd
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

func Split(str string, c byte, n int) []string {
	res := []string{}
	for i, j := 0, 0; i < len(str); i++ {
		if str[i] == c {
			continue
		}
		for j = i; j < len(str); j++ {
			if str[j] == c {
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
func FmtSize(size int64) string {
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
func FmtTime(t int64) string {
	sign, time := "", t
	if time < 0 {
		sign, time = "-", -t
	}
	if time > 1000000000 {
		return fmt.Sprintf("%s%d.%ds", sign, time/1000000000, (time/1000000)%1000*100/1000)
	}
	if time > 1000000 {
		return fmt.Sprintf("%s%d.%dms", sign, time/1000000, (time/1000)%1000*100/1000)
	}
	if time > 1000 {
		return fmt.Sprintf("%s%d.%dus", sign, time/1000, (time/1000)%1000*100/1000)
	}
	return fmt.Sprintf("%s%dns", sign, time)
}

func Marshal(data interface{}, arg ...interface{}) string {
	if len(arg) > 0 {
		switch arg := arg[0].(type) {
		case string:
			if f, p, e := Create(arg); e == nil {
				defer f.Close()

				switch {
				case strings.HasSuffix(arg, ".json"):
					b, _ := json.MarshalIndent(data, "", "  ")
					if n, e := f.Write(b); e == nil && n == len(b) {
						return p
					}

				case strings.HasSuffix(arg, ".csv"):
					switch data := data.(type) {
					case []interface{}:
						w := csv.NewWriter(f)
						head := []string{}
						for _, v := range data {
							switch v := v.(type) {
							case map[string]interface{}:
								if len(head) == 0 {
									for k, _ := range v {
										head = append(head, k)
									}
									w.Write(head)
								}

								fields := []string{}
								for _, k := range head {
									fields = append(fields, Format(v[k]))
								}
								w.Write(fields)
							}
						}
						w.Flush()
					}
				}
			}
		}
	}

	b, _ := json.MarshalIndent(data, "", "  ")
	return string(b)
}
func UnMarshal(data string) interface{} {
	var res interface{}
	if strings.HasSuffix(data, ".json") {
		if b, e := ioutil.ReadFile(data); e == nil {
			json.Unmarshal(b, &res)
		}
	} else {
		json.Unmarshal([]byte(data), &res)
	}
	return res
}
func UnMarshalm(data string) map[string]interface{} {
	res, _ := UnMarshal(data).(map[string]interface{})
	return res
}
