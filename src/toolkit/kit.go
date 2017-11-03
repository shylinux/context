package kit

import (
	"fmt"
	"io/ioutil"
	// "log"
	"os"
	"path"
)

func FmtSize(size int64) string {
	if size > 1000000000 {
		return fmt.Sprintf("%d.%dG", size/1000000000, size/100000000%100)
	}

	if size > 1000000 {
		return fmt.Sprintf("%d.%dM", size/100000, size/100000%100)
	}

	if size > 1000 {
		return fmt.Sprintf("%d.%dK", size/1000, size/100%100)
	}

	return fmt.Sprintf("%dB", size)
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
