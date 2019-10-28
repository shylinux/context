package wiki

import (
	"github.com/gomarkdown/markdown"

	"contexts/ctx"
	"contexts/web"
	"toolkit"

	"bytes"
	"encoding/json"
	"html/template"
	"path"
)

var Index = &ctx.Context{Name: "wiki", Help: "文档中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"login": {Name: "login", Value: map[string]interface{}{"check": "false"}, Help: "用户登录"},
		"componet": {Name: "componet", Value: map[string]interface{}{
			"index": []interface{}{
				map[string]interface{}{"name": "wiki",
					"tmpl": "head", "metas": []interface{}{map[string]interface{}{
						"name": "viewport", "content": "width=device-width, initial-scale=0.7, user-scalable=no",
					}}, "favicon": "favicon.ico", "styles": []interface{}{
						"example.css", "wiki.css",
					}},
				map[string]interface{}{"name": "header",
					"tmpl": "fieldset", "view": "Header", "init": "initHeader",
				},
				map[string]interface{}{"name": "tree",
					"tmpl": "fieldset", "view": "Tree", "init": "initTree",
					"ctx": "web.wiki", "cmd": "tree",
				},
				map[string]interface{}{"name": "text",
					"tmpl": "fieldset", "view": "Text", "init": "initText",
					"ctx": "web.wiki", "cmd": "text",
				},
				map[string]interface{}{"name": "footer",
					"tmpl": "fieldset", "view": "Footer", "init": "initFooter",
				},
				map[string]interface{}{"name": "tail",
					"tmpl": "tail", "scripts": []interface{}{
						"toolkit.js", "context.js", "example.js", "wiki.js",
					},
				},
			},
		}, Help: "组件列表"},

		"level": {Name: "level", Value: "local/wiki/自然/编程", Help: "路由数量"},
		"class": {Name: "class", Value: "", Help: "路由数量"},
		"favor": {Name: "favor", Value: "index.md", Help: "路由数量"},
	},
	Commands: map[string]*ctx.Command{
		"tree": {Name: "tree", Help: "目录", Form: map[string]int{"level": 1, "class": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy("nfs.dir", path.Join(m.Confx("level"), m.Confx("class", arg, 0)),
				"time", "size", "line", "file", "dir_sort", "time", "time_r")
			return
		}},
		"text": {Name: "text", Help: "文章", Form: map[string]int{"level": 1, "class": 1, "favor": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			which := m.Cmdx("nfs.path", path.Join(m.Confx("level"), m.Confx("class", arg, 1), m.Confx("favor", arg, 0)))

			buffer := bytes.NewBuffer([]byte{})
			template.Must(template.ParseFiles(which)).Funcs(ctx.CGI).Execute(buffer, m)
			m.Echo(string(markdown.ToHTML(buffer.Bytes(), nil, nil)))
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
				m.Log("info", "m: %d n: %d", max, n)

				for k := 0; k < n+2; k++ {
					for i := 0; i < max+2; i++ {
						m.Push(kit.Format(k), kit.Format(data[k][i]))
					}
				}

			case 2:
				m.Cmdy("ssh.data", "insert", "xls", "title", arg[0], "content", arg[1])

			default:
				data := map[int]map[int]string{}
				what := m.Cmd("ssh.data", "show", "xls", arg[0], "format", "object").Append("content")
				json.Unmarshal([]byte(what), &data)

				for i := 1; i < len(arg)-2; i += 3 {
					if _, ok := data[kit.Int(arg[i])]; !ok {
						data[kit.Int(arg[i])] = make(map[int]string)
					}
					data[kit.Int(arg[i])][kit.Int(arg[i+1])] = arg[i+2]
				}
				m.Cmdy("ssh.data", "update", "xls", arg[0], "content", kit.Format(data))
			}
			return
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.WEB{Context: Index})
}
