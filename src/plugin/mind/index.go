package main

import (
    "contexts/cli"
    "contexts/ctx"
    "toolkit"

	"encoding/json"
    "fmt"
    "os"
)

var Index = &ctx.Context{Name: "mind", Help: "思维导图",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"doc": {Name: "doc", Help: "文档", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
            return
		}},
		"xls": {Name: "xls", Help: "表格", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch len(arg) {
			case 0:
				m.Cmdy("ssh.data", "show", "xls")
				m.Meta["append"] = []string{"id", "title"}

			case 1:
				var data map[int]map[int]string
				what := m.Cmd("ssh.data", "show", "xls", arg[0], "format", "object").Append("content")
				json.Unmarshal([]byte(what), &data)

				max, n := 0, 0
				for i, v := range data {
					if i > n {
						n = i
					}
					for i := range v {
						if i > max {
							max = i
						}
					}
				}
				m.Log("info", "m: %d n: %d", m, n)

				for k := 0; k < n + 2; k++ {
					for i := 0; i < max + 2; i++ {
						m.Push(kit.Format(k), kit.Format(data[k][i]))
					}
				}

			case 2:
				m.Cmdy("ssh.data", "insert", "xls", "title", arg[0], "content", arg[1])

			default:
				data := map[int]map[int]string{}
				what := m.Cmd("ssh.data", "show", "xls", arg[0], "format", "object").Append("content")
				json.Unmarshal([]byte(what), &data)

				for i := 1; i < len(arg) - 2; i += 3 {
					if _, ok := data[kit.Int(arg[i])]; !ok {
						data[kit.Int(arg[i])] = make(map[int]string)
					}
					data[kit.Int(arg[i])][kit.Int(arg[i+1])] = arg[i+2]
				}
				m.Cmdy("ssh.data", "update", "xls", arg[0], "content", kit.Format(data))
			}
            return
		}},
		"ppt": {Name: "ppt", Help: "文稿", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
            return
		}},
	},
}

func main() {
	fmt.Print(cli.Index.Plugin(Index, os.Args[1:]))
}
