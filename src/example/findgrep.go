package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"toolkit"
)

var files = ".*\\.(xml|html|css|js)$"
var words = "[[:^ascii:]]+"

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage", os.Args[0], "dirs [files [words]]")
		fmt.Println("在目录dirs中，查找匹配files的文件，并查找匹配words的单词")
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		files = os.Args[2]
	}
	if len(os.Args) > 3 {
		words = os.Args[3]
	}

	word, e := regexp.Compile(words)
	kit.Check(e)
	// out, e := os.Create(os.Args[2])
	// kit.Check(e)
	out := os.Stdout

	total := 0
	count := 0
	chars := 0
	kit.DirWalk(os.Args[1], func(file string) {
		s, _ := os.Stat(file)
		if s.IsDir() {
			return
		}
		if m, e := regexp.MatchString(files, file); !kit.Check(e) || !m {
			return
		}

		f, e := os.Open(file)
		kit.Check(e)
		bio := bufio.NewReader(f)

		fmt.Fprintln(out, kit.FmtSize(s.Size()), file)
		line := 0

		cs := 0
		for i := 1; true; i++ {
			l, e := bio.ReadString('\n')
			if e == io.EOF {
				break
			}
			kit.Check(e)
			if i == 1 {
				continue
			}

			a := word.FindAllString(l, 20)
			for _, v := range a {
				n := len([]rune(v))
				fmt.Fprintf(out, "l:%d c:%d %s\n", i, n, v)
				total++
				line++
				chars += n
				cs += n
			}
		}
		fmt.Fprintln(out, "lines:", line, "chars:", cs, file)
		fmt.Fprintln(out)
		if line > 0 {
			count++
		}
	})
	fmt.Fprintln(out, "files:", count, "lines:", total, "chars:", chars, os.Args[1])
	return
}
