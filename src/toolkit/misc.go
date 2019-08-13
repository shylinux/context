package kit

import (
	"fmt"
	"os"
	"path"
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
func FmtTime(time int64) string {
	if time > 1000000000 {
		return fmt.Sprintf("%d.%ds", time/1000000000, (time/1000000)%1000*100/1000)
	}
	if time > 1000000 {
		return fmt.Sprintf("%d.%dms", time/1000000, (time/1000)%1000*100/1000)
	}
	if time > 1000 {
		return fmt.Sprintf("%d.%dus", time/1000, (time/1000)%1000*100/1000)
	}
	return fmt.Sprintf("%dns", time)
}
