package code

import (
	"contexts/ctx"
	"contexts/web"
	"fmt"
	"path"
	"plugin"
	"runtime"
	"strconv"
	"strings"
	"time"
	"toolkit"
)

var Index = &ctx.Context{Name: "code", Help: "代码中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"skip_login": &ctx.Config{Name: "skip_login", Value: map[string]interface{}{"/consul": "true"}, Help: "免密登录"},
		"componet": &ctx.Config{Name: "componet", Value: map[string]interface{}{
			"login": []interface{}{
				map[string]interface{}{"componet_name": "code", "componet_tmpl": "head", "metas": []interface{}{
					map[string]interface{}{"name": "viewport", "content": "width=device-width, initial-scale=0.7, user-scalable=no"},
				}, "favicon": "favicon.ico", "styles": []interface{}{"example.css", "code.css"}},

				map[string]interface{}{"componet_name": "login", "componet_help": "login",
					"componet_tmpl": "componet", "componet_init": "initLogin",
					"componet_ctx": "aaa", "componet_cmd": "auth", "componet_args": []interface{}{"@sessid", "ship", "username", "@username", "password", "@password"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "username", "value": "", "label": "username"},
						map[string]interface{}{"type": "password", "name": "password", "value": "", "label": "password"},
						map[string]interface{}{"type": "button", "value": "login"},
					},
					"display_append": "", "display_result": "",
				},

				map[string]interface{}{"componet_name": "tail", "componet_tmpl": "tail",
					"scripts": []interface{}{"toolkit.js", "context.js", "example.js", "code.js"},
				},
			},
			"flash": []interface{}{
				map[string]interface{}{"componet_name": "flash", "componet_tmpl": "head", "metas": []interface{}{
					map[string]interface{}{"name": "viewport", "content": "width=device-width, initial-scale=0.7, user-scalable=no"},
				}, "favicon": "favicon.ico", "styles": []interface{}{"example.css", "code.css"}},

				map[string]interface{}{"componet_name": "ask", "componet_help": "ask", "componet_tmpl": "componet",
					"componet_view": "FlashText", "componet_init": "initFlashText",
					"componet_ctx": "web.code", "componet_cmd": "flash", "componet_args": []interface{}{"text", "@text"}, "inputs": []interface{}{
						map[string]interface{}{"type": "textarea", "name": "text", "value": "", "cols": 50, "rows": 5},
						map[string]interface{}{"type": "button", "value": "添加请求"},
					},
					"display_result": "", "display_append": "",
				},
				map[string]interface{}{"componet_name": "tip", "componet_help": "tip", "componet_tmpl": "componet",
					"componet_view": "FlashList", "componet_init": "initFlashList",
					"componet_ctx": "web.code", "componet_cmd": "flash",
					"display_result": "", "display_append": "",
				},

				map[string]interface{}{"componet_name": "tail", "componet_tmpl": "tail",
					"scripts": []interface{}{"toolkit.js", "context.js", "example.js", "code.js"},
				},
			},
			"schedule": []interface{}{
				map[string]interface{}{"componet_name": "flash", "componet_tmpl": "head", "metas": []interface{}{
					map[string]interface{}{"name": "viewport", "content": "width=device-width, initial-scale=0.7, user-scalable=no"},
				}, "favicon": "favicon.ico", "styles": []interface{}{"example.css", "code.css"}},

				map[string]interface{}{"componet_name": "com", "componet_help": "com", "componet_tmpl": "componet",
					"componet_view": "ComList", "componet_init": "initComList",
					"componet_ctx": "web.code", "componet_cmd": "componet", "componet_args": []interface{}{"share", "@role", "@componet_group", "@tips"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "role", "value": "tech", "label": "role"},
						map[string]interface{}{"type": "text", "name": "tips", "value": "schedule", "label": "tips"},
						map[string]interface{}{"type": "button", "value": "共享页面"},
					},
				},

				map[string]interface{}{"componet_name": "text", "componet_help": "text", "componet_tmpl": "componet",
					"componet_view": "ScheduleText", "componet_init": "initScheduleText",
					"componet_ctx": "web.code", "componet_cmd": "schedule",
					"componet_args": []interface{}{"@time", "@name", "@place"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "time", "value": "", "label": "time"},
						map[string]interface{}{"type": "text", "name": "name", "value": "", "label": "name"},
						map[string]interface{}{"type": "text", "name": "place", "value": "", "label": "place"},
						map[string]interface{}{"type": "button", "value": "添加行程"},
					},
					"display_result": "", "display_append": "",
				},
				map[string]interface{}{"componet_name": "list", "componet_help": "list", "componet_tmpl": "componet",
					"componet_view": "ScheduleList", "componet_init": "initScheduleList",
					"componet_ctx": "web.code", "componet_cmd": "schedule",
					"inputs": []interface{}{
						map[string]interface{}{"type": "choice", "name": "view", "value": "default", "label": "显示字段", "choice": []interface{}{
							map[string]interface{}{"name": "默认", "value": "default"},
							map[string]interface{}{"name": "行程", "value": "order"},
							map[string]interface{}{"name": "总结", "value": "summary"},
						}},
						map[string]interface{}{"type": "button", "value": "刷新行程"},
					},
					"display_result": "",
				},

				map[string]interface{}{"componet_name": "tail", "componet_tmpl": "tail",
					"scripts": []interface{}{"toolkit.js", "context.js", "example.js", "code.js"},
				},
			},
			"index": []interface{}{
				map[string]interface{}{"componet_name": "code", "componet_tmpl": "head", "metas": []interface{}{
					map[string]interface{}{"name": "viewport", "content": "width=device-width, initial-scale=0.7, user-scalable=no"},
				}, "favicon": "favicon.ico", "styles": []interface{}{"example.css", "code.css"}},
				map[string]interface{}{"componet_name": "banner", "componet_help": "banner", "componet_tmpl": "banner",
					"componet_view": "Banner", "componet_init": "initBanner",
				},

				map[string]interface{}{"componet_name": "toolkit", "componet_help": "Ctrl+B", "componet_tmpl": "toolkit",
					"componet_view": "KitList", "componet_init": "initKitList",
				},
				// map[string]interface{}{"componet_name": "login", "componet_help": "login", "componet_tmpl": "componet",
				// 	"componet_ctx": "aaa", "componet_cmd": "login", "componet_args": []interface{}{"@username", "@password"},
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "text", "name": "username", "label": "username"},
				// 		map[string]interface{}{"type": "password", "name": "password", "label": "password"},
				// 		map[string]interface{}{"type": "button", "value": "login"},
				// 	},
				// 	"display_append": "", "display_result": "",
				// },
				// map[string]interface{}{"componet_name": "userinfo", "componet_help": "userinfo", "componet_tmpl": "componet",
				// 	"componet_ctx": "aaa", "componet_cmd": "login", "componet_args": []interface{}{"@sessid"},
				// 	"pre_run": true,
				// },
				map[string]interface{}{"componet_name": "buffer", "componet_help": "buffer", "componet_tmpl": "componet",
					"componet_view": "BufList", "componet_init": "initBufList",
					"componet_ctx": "cli", "componet_cmd": "tmux", "componet_args": []interface{}{"buffer"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "limit", "value": "3", "label": "limit"},
						map[string]interface{}{"type": "text", "name": "index", "value": "0", "label": "index"},
						map[string]interface{}{"type": "button", "value": "refresh"},
					},
					"pre_run": true,
				},
				map[string]interface{}{"componet_name": "dir", "componet_help": "dir", "componet_tmpl": "componet",
					"componet_view": "DirList", "componet_init": "initDirList",
					"componet_ctx": "nfs", "componet_cmd": "dir", "componet_args": []interface{}{"@dir", "dir_sort", "@sort_field", "@sort_order"}, "inputs": []interface{}{
						map[string]interface{}{"type": "choice", "name": "dir_type", "value": "both", "label": "dir_type", "choice": []interface{}{
							map[string]interface{}{"name": "all", "value": "all"},
							map[string]interface{}{"name": "both", "value": "both"},
							map[string]interface{}{"name": "file", "value": "file"},
							map[string]interface{}{"name": "dir", "value": "dir"},
						}},
						map[string]interface{}{"type": "text", "name": "dir", "value": "@current.dir", "label": "dir"},
						map[string]interface{}{"type": "button", "value": "refresh"},
					},
					"pre_run": false, "display_result": "",
				},
				map[string]interface{}{"componet_name": "upload", "componet_help": "upload", "componet_tmpl": "componet",
					"componet_view": "PutFile", "componet_init": "initPutFile",
					"componet_ctx": "web", "componet_cmd": "upload", "form_type": "upload", "inputs": []interface{}{
						map[string]interface{}{"type": "file", "name": "upload"},
						map[string]interface{}{"type": "submit", "value": "submit"},
					},
					"display_result": "",
				},
				map[string]interface{}{"componet_name": "pod", "componet_help": "pod", "componet_tmpl": "componet",
					"componet_view": "PodList", "componet_init": "initPodList",
					"componet_ctx": "ssh", "componet_cmd": "node", "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "value": "@current.pod"},
						map[string]interface{}{"type": "button", "value": "refresh"},
					},
					"pre_run": true, "display_result": "",
				},
				map[string]interface{}{"componet_name": "ctx", "componet_help": "ctx", "componet_tmpl": "componet",
					"componet_view": "CtxList", "componet_init": "initCtxList",
					"componet_pod": "true", "componet_ctx": "ssh", "componet_cmd": "context", "componet_args": []interface{}{"@ctx", "list"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "ctx", "value": "@current.ctx"},
						map[string]interface{}{"type": "button", "value": "refresh"},
					},
					"display_result": "",
				},
				map[string]interface{}{"componet_name": "cmd", "componet_help": "cmd", "componet_tmpl": "componet",
					"componet_view": "CmdList", "componet_init": "initCmdList",
					"componet_ctx": "cli.shy", "componet_cmd": "source", "componet_args": []interface{}{"@cmd"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "cmd", "value": "", "class": "cmd", "clipstack": "void"},
					},
				},
				// map[string]interface{}{"componet_name": "mp", "componet_tmpl": "mp"},
				map[string]interface{}{"componet_name": "tail", "componet_tmpl": "tail",
					"scripts": []interface{}{"toolkit.js", "context.js", "example.js", "code.js"},
				},
			},
		}, Help: "组件列表"},
		"componet_group": &ctx.Config{Name: "component_group", Value: "index", Help: "默认组件"},

		"make": &ctx.Config{Name: "make", Value: map[string]interface{}{
			"go": map[string]interface{}{
				"build":  []interface{}{"go", "build"},
				"plugin": []interface{}{"go", "build", "-buildmode=plugin"},
				"load":   []interface{}{"load"},
			},
			"so": map[string]interface{}{
				"load": []interface{}{"load"},
			},
		}, Help: "免密登录"},

		"flash": &ctx.Config{Name: "flash", Value: map[string]interface{}{
			"data": []interface{}{},
			"view": map[string]interface{}{"default": []interface{}{"index", "time", "text", "code", "output"}},
		}, Help: "闪存"},
		"schedule": &ctx.Config{Name: "schedule", Value: map[string]interface{}{
			"data": []interface{}{},
			"view": map[string]interface{}{
				"default": []interface{}{"面试时间", "面试公司", "面试地点", "面试轮次", "题目类型", "面试题目", "面试总结"},
				"summary": []interface{}{"面试公司", "面试轮次", "题目类型", "面试题目", "面试总结"},
				"order":   []interface{}{"面试时间", "面试公司", "面试地点"},
			},
			"maps": map[string]interface{}{"baidu": "<a href='baidumap://map/direction?region=&origin=&destination=%s'>%s</a>"},
		}, Help: "闪存"},

		"counter": &ctx.Config{Name: "counter", Value: map[string]interface{}{
			"nopen": "0", "nsave": "0",
		}, Help: "counter"},
		"counter_service": &ctx.Config{Name: "counter_service", Value: "http://localhost:9094/code/counter", Help: "counter"},
	},
	Commands: map[string]*ctx.Command{
		"make": &ctx.Command{Name: "make [action] file [args...]", Help: "更新代码", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			target, action, suffix := path.Join(m.Conf("runtime", "boot.ctx_home"), "src/examples/app/bench.go"), "build", "go"
			if len(arg) == 0 {
				arg = append(arg, target)
			}

			if cs := strings.Split(arg[0], "."); len(cs) > 1 {
				suffix = cs[len(cs)-1]
			} else if cs := strings.Split(arg[1], "."); len(cs) > 1 {
				action, suffix, arg = arg[0], cs[len(cs)-1], arg[1:]
			}

			target = m.Cmdx("nfs.path", arg[0])
			if target == "" {
				target = m.Cmdx("nfs.path", path.Join("src/plugin/", arg[0]))
			}

			cook := m.Confv("make", []string{suffix, action})
			switch kit.Chains(cook, "0") {
			case "load":
				if suffix == "go" {
					so := strings.Replace(target, ".go", ".so", -1)
					m.Cmd("cli.system", m.Confv("make", "go.plugin"), "-o", so, target)
					arg[0] = so
				}

				if p, e := plugin.Open(arg[0]); m.Assert(e) {
					s, e := p.Lookup("Index")
					m.Assert(e)
					w := *(s.(**ctx.Context))
					c.Register(w, nil, true)
				}
			default:
				m.Cmdy("cli.system", cook, arg)
			}
			return
		}},
		"flash": &ctx.Command{Name: "flash", Help: "闪存", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			total := len(m.Confv("flash", "data").([]interface{}))
			// 查看列表
			if len(arg) == 0 {
				if index := m.Option("flash_index"); index != "" {
					arg = append(arg, index)
				} else {
					m.Confm("flash", "data", func(index int, item map[string]interface{}) {
						for _, k := range kit.View([]string{}, m.Confm("flash", "view")) {
							m.Add("append", k, kit.Format(item[k]))
						}
					})
					m.Table()
					return
				}

			}

			index, item := -1, map[string]interface{}{"time": m.Time()}
			if i, e := strconv.Atoi(arg[0]); e == nil && 0 <= i && i < total {
				// 查看索引
				index, arg = total-1-i, arg[1:]
				if item = m.Confm("flash", []interface{}{"data", index}); len(arg) == 0 {
					// 查看数据
					for _, k := range kit.View([]string{}, m.Confm("flash", "view")) {
						m.Add("append", k, kit.Format(item[k]))
					}
					m.Table()
					return e
				}
			}

			switch arg[0] {
			case "vim": // 编辑数据
				name := m.Cmdx("nfs.temp", kit.Format(item[kit.Select("code", arg, 1)]))
				m.Cmd("cli.system", "vi", name)
				item[kit.Select("code", arg, 1)] = m.Cmdx("nfs.load", name)
				m.Cmd("nfs.trash", name)

			case "run": // 运行代码
				code := kit.Format(item[kit.Select("code", arg, 1)])
				if code == "" {
					break
				}
				name := m.Cmdx("nfs.temp", code)
				m.Cmdy("cli.system", "python", name)
				item["output"] = m.Result(0)
				m.Cmd("nfs.trash", name)

			default:
				// 修改数据
				for i := 0; i < len(arg)-1; i += 2 {
					item[arg[i]] = arg[i+1]
				}
				m.Conf("flash", []interface{}{"data", index}, item)
				item["index"] = total - 1 - index
				m.Echo("%d", total-1-index)
			}

			return
		}},
		"schedule": &ctx.Command{Name: "schedule [time name place]", Help: "行程安排", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 { // 会话列表
				m.Confm("schedule", "data", func(index int, value map[string]interface{}) {
					for _, v := range kit.View([]string{m.Option("view")}, m.Confm("schedule", "view")) {
						if v == "面试地点" {
							m.Add("append", "面试地点", fmt.Sprintf(m.Conf("schedule", "maps.baidu"), value["面试地点"], value["面试地点"]))
							continue
						}
						m.Add("append", v, kit.Format(value[v]))
					}
				})
				m.Table()
				return
			}

			view := "default"
			if m.Confs("schedule", arg[0]) {
				view, arg = arg[0], arg[1:]
			}

			data := map[string]interface{}{}
			for _, k := range kit.View([]string{view}, m.Confm("schedule", "view")) {
				if len(arg) == 0 {
					data[k] = ""
					continue
				}
				data[k], arg = arg[0], arg[1:]
			}

			extra := map[string]interface{}{}
			for i := 0; i < len(arg)-1; i += 2 {
				data[arg[i]] = arg[i+1]
			}
			data["extra"] = extra

			m.Conf("schedule", []string{"data", "-1"}, data)
			return
		}},
		"12306": &ctx.Command{Name: "12306", Help: "12306", Form: map[string]int{"fields": 1, "limit": 1, "offset": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			date := time.Now().Add(time.Hour * 24).Format("2006-01-02")
			if len(arg) > 0 {
				date, arg = arg[0], arg[1:]
			}
			to := "QFK"
			if len(arg) > 0 {
				to, arg = arg[0], arg[1:]
			}
			from := "BJP"
			if len(arg) > 0 {
				from, arg = arg[0], arg[1:]
			}
			m.Echo("%s->%s %s\n", from, to, date)

			m.Cmd("web.get", fmt.Sprintf("https://kyfw.12306.cn/otn/leftTicket/queryX?leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=ADULT", date, from, to), "temp", "data.result")
			for _, v := range m.Meta["value"] {
				fields := strings.Split(v, "|")
				m.Add("append", "车次--", fields[3])
				m.Add("append", "出发----", fields[8])
				m.Add("append", "到站----", fields[9])
				m.Add("append", "时长----", fields[10])
				m.Add("append", "二等座", fields[30])
				m.Add("append", "一等座", fields[31])
			}
			m.Table()
			return
		}},
		"brow": &ctx.Command{Name: "brow url", Help: "浏览网页", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmd("tcp.ifconfig").Table(func(index int, value map[string]string) {
					m.Append("index", index)
					m.Append("site", fmt.Sprintf("%s://%s%s", m.Conf("serve", "protocol"), value["ip"], m.Conf("runtime", "boot.web_port")))
				})
				m.Table()
				return
			}

			switch runtime.GOOS {
			case "windows":
				m.Cmd("cli.system", "explorer", arg[0])
			case "darwin":
				m.Cmd("cli.system", "open", arg[0])
			default:
				m.Cmd("web.get", arg[0])
			}
			return
		}},

		"/counter": &ctx.Command{Name: "/counter", Help: "/counter", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) > 0 {
				m.Option("name", arg[0])
			}
			if len(arg) > 1 {
				m.Option("count", arg[1])
			}

			count := m.Optioni("count")
			switch v := m.Confv("counter", m.Option("name")).(type) {
			case string:
				i, e := strconv.Atoi(v)
				m.Assert(e)
				count += i
			}
			m.Log("info", "%v: %v", m.Option("name"), m.Confv("counter", m.Option("name"), fmt.Sprintf("%d", count)))
			m.Echo("%d", count)
			return
		}},
		"counter": &ctx.Command{Name: "counter name count", Help: "counter", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) > 1 {
				m.Copy(m.Spawn().Cmd("get", m.Conf("counter_service"), "name", arg[0], "count", arg[1]), "result")
			}
			return
		}},
		"tmux": &ctx.Command{Name: "tmux buffer", Help: "终端管理, buffer: 查看复制", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch arg[0] {
			case "buffer":
				bufs := strings.Split(m.Spawn().Cmd("system", "tmux", "list-buffers").Result(0), "\n")

				n := 3
				if m.Option("limit") != "" {
					n = m.Optioni("limit")
				}

				for i, b := range bufs {
					if i >= n {
						break
					}
					bs := strings.SplitN(b, ": ", 3)
					if len(bs) > 1 {
						m.Add("append", "buffer", bs[0][:len(bs[0])])
						m.Add("append", "length", bs[1][:len(bs[1])-6])
						m.Add("append", "strings", bs[2][1:len(bs[2])-1])
					}
				}

				if m.Option("index") == "" {
					m.Echo(m.Spawn().Cmd("system", "tmux", "show-buffer").Result(0))
				} else {
					m.Echo(m.Spawn().Cmd("system", "tmux", "show-buffer", "-b", m.Option("index")).Result(0))
				}
			}
			return
		}},
		"windows": &ctx.Command{Name: "windows", Help: "windows", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Append("nclient", strings.Count(m.Spawn().Cmd("system", "tmux", "list-clients").Result(0), "\n"))
			m.Append("nsession", strings.Count(m.Spawn().Cmd("system", "tmux", "list-sessions").Result(0), "\n"))
			m.Append("nwindow", strings.Count(m.Spawn().Cmd("system", "tmux", "list-windows", "-a").Result(0), "\n"))
			m.Append("npane", strings.Count(m.Spawn().Cmd("system", "tmux", "list-panes", "-a").Result(0), "\n"))

			m.Append("nbuf", strings.Count(m.Spawn().Cmd("system", "tmux", "list-buffers").Result(0), "\n"))
			m.Append("ncmd", strings.Count(m.Spawn().Cmd("system", "tmux", "list-commands").Result(0), "\n"))
			m.Append("nkey", strings.Count(m.Spawn().Cmd("system", "tmux", "list-keys").Result(0), "\n"))
			m.Table()
			return
		}},
		"notice": &ctx.Command{Name: "notice", Help: "睡眠, time(ns/us/ms/s/m/h): 时间值(纳秒/微秒/毫秒/秒/分钟/小时)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmd("cli.system", "osascript", "-e", fmt.Sprintf("display notification \"%s\"", kit.Select("", arg, 0)))
			return
		}},
	},
}

func init() {
	code := &web.WEB{}
	code.Context = Index
	web.Index.Register(Index, code)
}
