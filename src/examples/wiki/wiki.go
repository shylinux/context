package wiki

import (
	"github.com/gomarkdown/markdown"

	"contexts/ctx"
	"contexts/web"
	"toolkit"

	"bytes"
	"encoding/json"
	"path"
	"strings"
	"text/template"
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

		"level": {Name: "level", Value: "usr/local/wiki", Help: "文档路径"},
		"class": {Name: "class", Value: "", Help: "文档目录"},
		"favor": {Name: "favor", Value: "index.md", Help: "默认文档"},

		"commit": {Name: "data", Value: map[string]interface{}{
			"data": map[string]interface{}{},
			"ship": map[string]interface{}{},
			"head": map[string]interface{}{},
		}, Help: "数据"},
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
			tmpl := template.New("render").Funcs(*ctx.LocalCGI(m, c))

			tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Conf("route", "template_dir"), "/*.tmpl")))
			tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Conf("route", "template_dir"), m.Cap("route"), "/*.tmpl")))
			tmpl = template.Must(tmpl.ParseFiles(which))

			m.Optionv("tmpl", tmpl)
			m.Assert(tmpl.ExecuteTemplate(buffer, m.Option("filename", path.Base(which)), m))
			m.Echo(string(markdown.ToHTML(buffer.Bytes(), nil, nil)))
			return
		}},
		"note": {Name: "note file", Help: "便签", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) > 1 && arg[0] == "commit" {
				m.Cmd("commit", arg[1:])
			} else if len(arg) > 0 {
				m.Cmd(kit.Select("tree", "text", strings.HasSuffix(arg[0], ".md")), arg[0])
			} else {
				m.Cmd("tree")
			}
			return
		}},
		"commit": {Name: "commit file name type text", Help: "提交", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			head := kit.Hashs(arg[0], arg[1])
			prev := m.Conf("commit", []string{"head", head, "ship"})
			m.Log("info", "head: %v %v", head, m.Conf("commit", []string{"head", head}))
			if len(arg) == 2 {
				meta := m.Confm("commit", []string{"ship", prev})
				m.Push("time", meta["time"])
				m.Push("data", m.Conf("commit", []string{"data", kit.Format(meta["data"])}))
				m.Table()
				return
			}

			data := kit.Hashs(arg[3])
			m.Log("info", "data: %v %v", data, arg[3])
			m.Conf("commit", []string{"data", data}, arg[3])

			meta := map[string]interface{}{
				"prev": prev,
				"time": m.Time(),
				"file": arg[0],
				"name": arg[1],
				"type": arg[2],
				"data": data,
			}
			ship := kit.Hashs(kit.Format(meta))
			m.Log("info", "ship: %v %v", ship, meta)
			m.Conf("commit", []string{"ship", ship}, meta)

			m.Log("info", "head: %v %v", head, ship)
			m.Conf("commit", []string{"head", head, "ship"}, ship)
			m.Echo("%v", kit.Formats(meta))
			return
		}},
		"table": {Name: "table", Help: "表格", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				return
			}
			switch arg[1] {
			case "data":
				arg = []string{arg[0], m.Conf("commit", []string{"data", arg[2]})}

			default:
				msg := m.Spawn().Cmd("commit", m.Option("filename"), arg[0])
				m.Option("prev_data", msg.Append("data"))
				m.Option("prev_time", msg.Append("time"))
				m.Option("file", m.Option("filename"))
				m.Option("name", arg[0])
				m.Option("data", arg[1])
			}

			head := []string{}
			for i, l := range strings.Split(strings.TrimSpace(arg[1]), "\n") {
				if i == 0 {
					head = kit.Split(l, ' ', 100)
					continue
				}
				for j, v := range strings.Split(l, " ") {
					m.Push(head[j], v)
				}
			}
			return
		}},
		"runs": {Name: "run", Help: "便签", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy(arg).Set("append")
			return
		}},
		"run": {Name: "run", Help: "便签", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy(arg)
			return
		}},
		"time": {Name: "time", Help: "便签", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy("cli.time", "show").Set("append")
			return
		}},

		"svg": {Name: "svg", Help: "绘图", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo(arg[0])
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
