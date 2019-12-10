package wiki

import (
	"github.com/gomarkdown/markdown"
	mis "github.com/shylinux/toolkits"

	"contexts/ctx"
	"contexts/web"
	"toolkit"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"text/template"
)

type opt struct {
	font_size  int
	font_color string
	background string
	padding    int
	margin     int
}

func show(m *ctx.Message, str string) (res []string) {
	miss := []int{}
	list := mis.Split(str, "\n")
	for _, line := range list {
		dep := 0
	loop:
		for _, v := range []rune(line) {
			switch v {
			case ' ':
				dep++
			case '\t':
				dep += 4
			default:
				break loop
			}
		}
		if len(miss) > 0 {
			if miss[len(miss)-1] > dep {
				for i := len(miss) - 1; i >= 0; i-- {
					if miss[i] < dep {
						break
					}
					m.Log("show", "pop %d %v %v", dep, mis.Format(miss), line)
					res = append(res, "]", "}")
					miss = miss[:i]
				}
				m.Log("show", "push %d %v %v", dep, mis.Format(miss), line)
				miss = append(miss, dep)
			} else if miss[len(miss)-1] < dep {
				m.Log("show", "push %d %v %v", dep, mis.Format(miss), line)
				miss = append(miss, dep)
			} else {
				res = append(res, "]", "}")
			}
		} else {
			m.Log("show", "push %d %v", dep, mis.Format(miss))
			miss = append(miss, dep)
		}

		word := mis.Split(line)
		res = append(res, "{", "meta", "{", "text")
		res = append(res, word...)
		res = append(res, "}", "list", "[")

	}
	m.Log("haha", "%v %v", str, res)
	return
}

func size(m *ctx.Message, root map[string]interface{}, depth int, width map[int]int) int {
	text := kit.Format(kit.Chain(root, "meta.text"))
	if len(text) > width[depth] {
		width[depth] = len(text)
	}

	if list, ok := root["list"].([]interface{}); !ok || len(list) == 0 {
		kit.Chain(root, "meta.height", 1)
		m.Log("fuck", "size %v %d", kit.Chain(root, "meta.text"), 1)
		return 1
	}

	height := 0
	kit.Map(root["list"], "", func(index int, value map[string]interface{}) {
		height += size(m, value, depth+1, width)
	})
	kit.Chain(root, "meta.height", height)
	m.Log("fuck", "size %v %d", kit.Chain(root, "meta.text"), height)
	return height
}
func draw(m *ctx.Message, root map[string]interface{}, depth int, width map[int]int, x int, y int, opt *opt) {
	meta := root["meta"].(map[string]interface{})
	m.Log("fuck", "draw %v %d", meta["text"], y)
	height := kit.Int(meta["height"])
	p := (height - 1) * (opt.font_size + opt.margin + opt.padding)

	m.Echo(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`,
		x, y+p/2, width[depth]*opt.font_size/2+opt.padding, opt.font_size+opt.padding, kit.Select(opt.background, meta["bg"]))
	m.Echo(`<text x="%d" y="%d" font-size="%d" text-anchor="middle" fill="%s">%v</text>`,
		x+width[depth]*opt.font_size/2/2+opt.padding/2, y+p/2+opt.font_size-opt.padding/2, opt.font_size, kit.Select(opt.font_color, meta["fg"]), meta["text"])

	kit.Map(root["list"], "", func(index int, value map[string]interface{}) {
		draw(m, value, depth+1, width, x+width[depth]*opt.font_size/2+opt.margin+opt.padding, y, opt)
		y += kit.Int(kit.Chain(value, "meta.height")) * (opt.font_size + opt.margin + opt.padding)
	})
}

var Index = &ctx.Context{Name: "wiki", Help: "文档中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"login": {Name: "login", Value: map[string]interface{}{"check": "false", "meta": map[string]interface{}{
			"script": "usr/script",
		}}, Help: "用户登录"},
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

		"story": {Name: "story", Value: map[string]interface{}{
			"data": map[string]interface{}{},
			"node": map[string]interface{}{},
			"head": map[string]interface{}{},
		}, Help: "故事会"},

		"template": {Name: "template", Value: map[string]interface{}{
			"list": []interface{}{
				`{{define "raw"}}{{.|results}}{{end}}`,
				`{{define "title"}}{{.|results}}{{end}}`,
				`{{define "chapter"}}{{.|results}}{{end}}`,
				`{{define "section"}}{{.|results}}{{end}}`,
				`{{define "block"}}<div>{{.|results}}<div>{{end}}`,
			},
		}, Help: "故事会"},
	},
	Commands: map[string]*ctx.Command{
		"tree": {Name: "tree", Help: "目录", Form: map[string]int{"level": 1, "class": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy("nfs.dir", path.Join(m.Confx("level"), m.Confx("class", arg, 0)),
				"time", "size", "line", "file", "dir_sort", "time", "time_r").Set("result")
			return
		}},
		"text": {Name: "text", Help: "文章", Form: map[string]int{"level": 1, "class": 1, "favor": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			which := m.Cmdx("nfs.path", path.Join(m.Confx("level"), m.Confx("class", arg, 1), m.Confx("favor", arg, 0)))

			tmpl := template.New("render").Funcs(*ctx.LocalCGI(m, c))
			m.Confm("template", "list", func(index int, value string) { tmpl = template.Must(tmpl.Parse(value)) })
			tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Cmdx("nfs.path", m.Conf("route", "template_dir")), "/*.tmpl")))
			tmpl = template.Must(tmpl.ParseGlob(path.Join(m.Cmdx("nfs.path", m.Conf("route", "template_dir")), m.Cap("route"), "/*.tmpl")))
			tmpl = template.Must(tmpl.ParseFiles(which))
			for i, v := range tmpl.Templates() {
				m.Log("info", "%v, %v", i, v.Name())
			}
			m.Optionv("title", map[string]int{})
			m.Optionv("tmpl", tmpl)
			m.Option("render", "")

			buffer := bytes.NewBuffer([]byte{})
			m.Assert(tmpl.ExecuteTemplate(buffer, m.Option("filename", path.Base(which)), m))
			if f, p, e := kit.Create(path.Join("var/tmp/file", which)); e == nil {
				defer f.Close()
				f.Write(buffer.Bytes())
				m.Log("info", "save %v", p)
			}
			data := markdown.ToHTML(buffer.Bytes(), nil, nil)
			m.Echo(string(data))
			return
		}},
		"note": {Name: "note file|favor|commit", Help: "笔记", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmd("tree")
				return
			}

			switch arg[0] {
			case "favor", "commit":
				m.Cmd("story", arg[0], arg[1:])
			default:
				m.Cmd(kit.Select("tree", "text", strings.HasSuffix(arg[0], ".md")), arg[0])
			}
			return
		}},

		"story": {Name: "story favor|commit|branch|remote story scene enjoy happy", Help: "故事会", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			switch arg[0] {
			case "favor":
				if len(arg) < 4 {
					m.Cmdy("ssh.data", "show", arg[1:])
					break
				}

				head := kit.Hashs(arg[2], arg[4])
				prev := m.Conf(cmd, []string{"head", head, "node"})
				m.Cmdy("ssh.data", "insert", arg[1], "story", arg[2], "scene", arg[3], "enjoy", arg[4], "node", prev)

			case "commit":
				head := kit.Hashs(arg[1], arg[3])
				prev := m.Conf(cmd, []string{"head", head, "node"})
				m.Log("info", "head: %v %#v", head, prev)

				if len(arg) > 4 {
					data := kit.Hashs(arg[4])
					m.Log("info", "data: %v %v", data, arg[4])
					if m.Conf(cmd, []string{"node", prev, "data"}) != data {
						m.Conf(cmd, []string{"data", data}, arg[4])

						meta := map[string]interface{}{
							"time":  m.Time(),
							"story": arg[1],
							"scene": arg[2],
							"enjoy": arg[3],
							"data":  data,
							"prev":  prev,
						}
						node := kit.Hashs(kit.Format(meta))
						m.Log("info", "node: %v %v", node, meta)
						m.Conf(cmd, []string{"node", node}, meta)

						m.Log("info", "head: %v %v", head, node)
						m.Conf(cmd, []string{"head", head, "node"}, node)
						m.Echo("%v", kit.Formats(meta))
						break
					}
				}

				for prev != "" {
					node := m.Confm(cmd, []string{"node", prev})
					m.Push("node", kit.Short(prev, 6))
					m.Push("time", node["time"])
					m.Push("data", m.Conf(cmd, []string{"data", kit.Format(node["data"])}))
					prev = kit.Format(node["prev"])
				}
				m.Table()
				return

			case "branch":
				m.Confm(cmd, "head", func(key string, value map[string]interface{}) {
					node := kit.Format(value["node"])
					m.Push("key", kit.Short(key, 6))
					m.Push("story", m.Conf(cmd, []string{"node", node, "story"}))
					m.Push("scene", m.Conf(cmd, []string{"node", node, "scene"}))
					m.Push("enjoy", m.Conf(cmd, []string{"node", node, "enjoy"}))
					m.Push("node", kit.Short(value["node"], 6))
				})
				m.Table()
			case "remote":
			}
			return
		}},
		"index": {Name: "index name|hash", Help: "索引", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			scene := ""
			if hash := m.Conf("story", []string{"head", kit.Hashs(m.Option("filename"), arg[0]), "node"}); hash != "" {
				arg[0] = hash
			} else if hash := m.Conf("story", []string{"head", arg[0], "node"}); hash != "" {
				arg[0] = hash
			}
			if hash := m.Conf("story", []string{"node", arg[0], "data"}); hash != "" {
				scene = m.Conf("story", []string{"node", arg[0], "scene"})
				arg[0] = hash
			}
			if data := m.Conf("story", []string{"data", arg[0]}); data != "" {
				arg[0] = data
			}
			if scene != "" {
				m.Cmdy(scene, "", arg[0])
			} else {
				m.Echo(arg[0])
			}
			return
		}},
		"table": {Name: "table name data", Help: "表格", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			if len(arg) < 2 {
				return
			}
			m.Option("scene", cmd)
			m.Option("enjoy", arg[0])
			m.Option("happy", arg[1])
			m.Option("render", cmd)

			head := []string{}
			for i, l := range strings.Split(strings.TrimSpace(arg[1]), "\n") {
				if i == 0 {
					head = kit.Split(l, ' ', 100)
					continue
				}
				for j, v := range kit.Split(l, ' ', 100) {
					m.Push(head[j], v)
				}
			}
			return
		}},
		"order": {Name: "order", Help: "列表", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			if len(arg) < 2 {
				return
			}
			m.Option("scene", cmd)
			m.Option("enjoy", arg[0])
			m.Option("happy", arg[1])
			m.Option("render", cmd)

			for _, l := range strings.Split(strings.TrimSpace(arg[1]), "\n") {
				m.Push("list", l)
			}
			return
		}},
		"refer": {Name: "refer", Help: "链接地址", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			if len(arg) == 1 {
				cmd, arg = arg[0], arg[1:]
				for _, l := range strings.Split(strings.TrimSpace(cmd), "\n") {
					if l = strings.TrimSpace(l); len(l) > 0 {
						arg = append(arg, kit.Split(l, ' ', 2)...)
					}
				}
			}

			m.Set("option", "render", "order")
			for i := 0; i < len(arg)-1; i += 2 {
				m.Push("list", fmt.Sprintf(`%s: <a href="%s" target="_blank">%s</a>`, arg[i], arg[i+1], arg[i+1]))
			}
			return
		}},
		"favor": {Name: "favor type tab", Help: "链接地址", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			msg := m.Cmd("ssh.data", "show", "tip", "", "1000", "0", "tab", arg[1])

			switch arg[0] {
			case "script":
				m.Set("option", "render", "code")
				if b, e := ioutil.ReadFile(path.Join(m.Conf("login", "meta.script"), arg[1])); e == nil {
					m.Echo(string(b))
				}

			case "li":
				m.Set("option", "render", "order")
				msg.Table(func(index int, value map[string]string) {
					m.Push("list", fmt.Sprintf(`%s: <a href="%s" target="_blank">%s</a>`, value["note"], value["word"], value["word"]))
				})

			case "sh":
				m.Set("option", "render", "code")
				m.Echo("#! /bin/sh\n")
				m.Echo("# %v\n", arg[1])
				m.Echo("\n")
				msg.Table(func(index int, value map[string]string) {
					m.Echo("# %d %v\n%v\n\n", index, value["note"], value["word"])
				})
			}
			return
		}},
		"shell": {Name: "shell dir cmd", Help: "命令行", Form: map[string]int{"style": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Option("render", "code")
			m.Echo("$ %v\n", strings.Join(arg[1:], " "))
			m.Cmdy("cli.system", "cmd_dir", arg[0], "bash", "-c", strings.Join(arg[1:], " "))
			return
		}},
		"chart": {Name: "chart type text", Help: "绘图", Form: map[string]int{"style": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Option("render", "raw")
			var chart Chart
			switch arg[0] {
			case "block":
				chart = &Block{}
			case "chain":
				chart = &Chain{}
			case "table":
				chart = &Table{}
			}
			arg[1] = strings.TrimSpace(arg[1])

			chart.Init(m, arg[1:]...)
			m.Echo(`<svg vertion="1.1" xmlns="http://www.w3.org/2000/svg" width="%d", height="%d" style="%s">`,
				chart.GetWidth(), chart.GetHeight(), m.Option("style"))
			m.Echo("\n")
			chart.Draw(m, 0, 0)
			m.Echo(`</svg>`)
			m.Echo("\n")
			return
		}},

		"title": {Name: "title text", Help: "一级标题", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			ns := strings.Split(m.Conf("runtime", "node.name"), "-")
			m.Set("option", "render", cmd).Echo(kit.Select(ns[len(ns)-1], arg, 0))
			return
		}},
		"chapter": {Name: "chaper text", Help: "二级标题", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			prefix := ""
			if title, ok := m.Optionv("title").(map[string]int); ok {
				title["chapter"]++
				title["section"] = 0
				prefix = strconv.Itoa(title["chapter"]) + " "
			}
			m.Set("option", "render", cmd).Echo(prefix + kit.Select("", arg, 0))
			return
		}},
		"section": {Name: "section text", Help: "三级标题", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			prefix := ""
			if title, ok := m.Optionv("title").(map[string]int); ok {
				title["section"]++
				prefix = strconv.Itoa(title["chapter"]) + "." + strconv.Itoa(title["section"]) + " "
			}
			m.Set("option", "render", cmd).Echo(prefix + kit.Select("", arg, 0))
			return
		}},

		"run": {Name: "run", Help: "便签", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Option("render", "raw")
			m.Cmdy(arg)
			return
		}},
		"time": {Name: "time", Help: "便签", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy("cli.time", "show").Set("append")
			return
		}},

		"svg": {Name: "svg", Help: "绘图", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			data := map[string]interface{}{"meta": map[string]interface{}{"text": "chat"}, "list": []interface{}{
				map[string]interface{}{"meta": map[string]interface{}{"text": "ocean"}},
				map[string]interface{}{"meta": map[string]interface{}{"text": "river"}},
				map[string]interface{}{"meta": map[string]interface{}{"text": "dream"}, "list": []interface{}{
					map[string]interface{}{"meta": map[string]interface{}{"text": "zsh"}, "list": []interface{}{
						map[string]interface{}{"meta": map[string]interface{}{"text": "auto.sh"}},
					}},
					map[string]interface{}{"meta": map[string]interface{}{"text": "tmux"}},
					map[string]interface{}{"meta": map[string]interface{}{"text": "docker"}},
					map[string]interface{}{"meta": map[string]interface{}{"text": "git"}},
					map[string]interface{}{"meta": map[string]interface{}{"text": "vim"}, "list": []interface{}{
						map[string]interface{}{"meta": map[string]interface{}{"text": "auto.vim"}},
					}},
				}},
				map[string]interface{}{"meta": map[string]interface{}{"text": "storm"}},
				map[string]interface{}{"meta": map[string]interface{}{"text": "steam"}},
			}}

			opt := &opt{
				font_size:  kit.Int(kit.Select("60", arg, 0)),
				font_color: kit.Select("red", arg, 1),
				background: kit.Select("green", arg, 2),
				padding:    10,
				margin:     20,
			}
			if len(arg) > 3 {
				data = mis.Parse(nil, "", show(m, arg[3])...).(map[string]interface{})
			}

			max := map[int]int{}
			num := size(m, data, 0, max)
			width := 0
			for _, v := range max {
				width += v*opt.font_size/2 + opt.margin
			}

			m.Echo(`<svg vertion="1.1" xmlns="http://www.w3.org/2000/svg" width="%d", height="%d">`,
				width, num*(opt.font_size+opt.padding+opt.margin)-opt.margin)
			draw(m, data, 0, max, 0, 0, opt)
			m.Echo(`</svg>`)

			// m.Echo(`<rect width="100%" height="100%" fill="red"/>`)
			// m.Echo(`<circle cx="150" cy="100" r="80" fill="green"/>`)
			// m.Echo(`<text x="150" y="100" font-size="60" text-anchor="middle" fill="black">SVG</text>`)
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
