package cli

var Template = []interface{}{
	map[string]interface{}{"name": "index.go", "text": `package main

import (
    "contexts/cli"
    "contexts/ctx"
    "toolkit"

    "fmt"
    "os"
)

var Index = &ctx.Context{Name: "test", Help: "测试工具",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"_index": &ctx.Config{Name: "index", Value: []interface{}{
			map[string]interface{}{"name": "demo", "help": "demo",
				"tmpl": "componet", "view": "componet", "init": "",
				"type": "public", "ctx": "demo", "cmd": "demo",
				"args": []interface{}{}, "inputs": []interface{}{
					map[string]interface{}{"type": "text", "name": "pod", "value": "hello world"},
					map[string]interface{}{"type": "button", "value": "执行"},
				},
			},
		}},
	},
	Commands: map[string]*ctx.Command{
		"demo": {Name: "demo", Help: "demo", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
            m.Echo(kit.Select("hello world", arg, 0))
            return
		}},
	},
}

func main() {
	fmt.Print(cli.Index.Plugin(Index, os.Args[1:]))
}
`}, map[string]interface{}{"name": "index.shy", "text": `
fun hello world "" "" \
	public \
	text "" \
	button "执行"
	copy pwd
end
`}, map[string]interface{}{"name": "index.js", "text": `
{init: function(page, pane, field, option, output) {
    kit.Log("hello world")
}}
`}, map[string]interface{}{"name": "local.shy", "text": ` `},
}