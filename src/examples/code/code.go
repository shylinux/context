package code

import (
	"contexts/ctx"
	"contexts/web"
	"fmt"
	"strconv"
)

var Index = &ctx.Context{Name: "code", Help: "代码中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"counter": &ctx.Config{Name: "counter", Value: map[string]interface{}{
			"nopen": "0", "nsave": "0",
		}, Help: "counter"},
		"counter_service": &ctx.Config{Name: "counter_service", Value: "http://localhost:9094/code/counter", Help: "counter"},
		"web_site": &ctx.Config{Name: "web_site", Value: []interface{}{
			map[string]interface{}{"_name": "MDN", "site": "https://developer.mozilla.org"},
			map[string]interface{}{"_name": "github", "site": "https://github.com"},
		}, Help: "web_site"},
		"componet_command": &ctx.Config{Name: "component_command", Value: "pwd", Help: "默认命令"},
		"componet_group":   &ctx.Config{Name: "component_group", Value: "index", Help: "默认组件"},
		"componet": &ctx.Config{Name: "componet", Value: map[string]interface{}{
			"login": []interface{}{
				map[string]interface{}{"componet_name": "head", "template": "head"},
				map[string]interface{}{"componet_name": "login", "componet_help": "login", "template": "componet",
					"componet_ctx": "aaa", "componet_cmd": "auth", "arguments": []interface{}{"@sessid", "ship", "username", "@username", "password", "@password"},
					"inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "username", "label": "username"},
						map[string]interface{}{"type": "password", "name": "password", "label": "password"},
						map[string]interface{}{"type": "button", "value": "login"},
					},
					"display_append": "", "display_result": "",
				},
				map[string]interface{}{"componet_name": "tail", "template": "tail"},
			},
			"index": []interface{}{
				map[string]interface{}{"componet_name": "head", "template": "head"},
				map[string]interface{}{"componet_name": "docker", "componet_help": "docker", "template": "docker"},
				// map[string]interface{}{"componet_name": "login", "componet_help": "login", "template": "componet",
				// 	"componet_ctx": "aaa", "componet_cmd": "login", "arguments": []interface{}{"@username", "@password"},
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "text", "name": "username", "label": "username"},
				// 		map[string]interface{}{"type": "password", "name": "password", "label": "password"},
				// 		map[string]interface{}{"type": "button", "value": "login"},
				// 	},
				// 	"display_append": "", "display_result": "",
				// },
				// map[string]interface{}{"componet_name": "userinfo", "componet_help": "userinfo", "template": "componet",
				// 	"componet_ctx": "aaa", "componet_cmd": "login", "arguments": []interface{}{"@sessid"},
				// 	"pre_run": true,
				// },
				map[string]interface{}{"componet_name": "clipboard", "componet_help": "clipboard", "template": "clipboard"},
				map[string]interface{}{"componet_name": "buffer", "componet_help": "buffer", "template": "componet",
					"componet_ctx": "cli", "componet_cmd": "tmux", "arguments": []interface{}{"buffer"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "limit", "label": "limit", "value": "3"},
						map[string]interface{}{"type": "text", "name": "index", "label": "index"},
						map[string]interface{}{"type": "button", "value": "refresh"},
					},
					"pre_run": true,
				},
				// map[string]interface{}{"componet_name": "time", "componet_help": "time", "template": "componet",
				// 	"componet_ctx": "cli", "componet_cmd": "time", "arguments": []interface{}{"@string"},
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "text", "name": "time_format",
				// 			"label": "format", "value": "2006-01-02 15:04:05",
				// 		},
				// 		map[string]interface{}{"type": "text", "name": "string", "label": "string"},
				// 		map[string]interface{}{"type": "button", "value": "refresh"},
				// 	},
				// },
				// map[string]interface{}{"componet_name": "json", "componet_help": "json", "template": "componet",
				// 	"componet_ctx": "nfs", "componet_cmd": "json", "arguments": []interface{}{"@string"},
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "text", "name": "string", "label": "string"},
				// 		map[string]interface{}{"type": "button", "value": "refresh"},
				// 	},
				// },
				map[string]interface{}{"componet_name": "dir", "componet_help": "dir", "template": "componet",
					"componet_ctx": "nfs", "componet_cmd": "dir", "arguments": []interface{}{"@dir", "dir_sort", "@sort_order", "@sort_field"},
					"pre_run": true, "display_result": "",
					"inputs": []interface{}{
						map[string]interface{}{"type": "choice", "name": "dir_type",
							"label": "dir_type", "value": "both", "choice": []interface{}{
								map[string]interface{}{"name": "both", "value": "both"},
								map[string]interface{}{"name": "file", "value": "file"},
								map[string]interface{}{"name": "dir", "value": "dir"},
							},
						},
						map[string]interface{}{"type": "choice", "name": "sort_field",
							"label": "sort_field", "value": "time", "choice": []interface{}{
								map[string]interface{}{"name": "filename", "value": "filename"},
								map[string]interface{}{"name": "is_dir", "value": "type"},
								map[string]interface{}{"name": "line", "value": "line"},
								map[string]interface{}{"name": "size", "value": "size"},
								map[string]interface{}{"name": "time", "value": "time"},
							},
						},
						map[string]interface{}{"type": "choice", "name": "sort_order",
							"label": "sort_order", "value": "time_r", "choice": []interface{}{
								map[string]interface{}{"name": "str", "value": "str"},
								map[string]interface{}{"name": "str_r", "value": "str_r"},
								map[string]interface{}{"name": "int", "value": "int"},
								map[string]interface{}{"name": "int_r", "value": "int_r"},
								map[string]interface{}{"name": "time", "value": "time"},
								map[string]interface{}{"name": "time_r", "value": "time_r"},
							},
						},
						map[string]interface{}{"type": "text", "name": "dir", "label": "dir"},
					},
				},
				map[string]interface{}{"componet_name": "upload", "componet_help": "upload", "template": "componet",
					"componet_ctx": "web", "componet_cmd": "upload", "form_type": "upload",
					"inputs": []interface{}{
						map[string]interface{}{"type": "file", "name": "upload"},
						map[string]interface{}{"type": "submit", "value": "submit"},
					},
					"display_result": "",
				},
				// map[string]interface{}{"componet_name": "download", "componet_help": "download", "template": "componet",
				// 	"componet_ctx": "cli.shy", "componet_cmd": "source", "arguments": []interface{}{"@cmds"},,
				// 	"display_result": "", "download_file": "",
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "text", "name": "download_file", "value": "data_2006_0102_1504.txt", "class": "file_name"},
				// 		map[string]interface{}{"type": "text", "name": "cmds", "value": "",
				// 			"class": "file_cmd", "clipstack": "clistack",
				// 		},
				// 	},
				// },
				map[string]interface{}{"componet_name": "command", "componet_help": "command", "template": "componet",
					"componet_ctx": "cli.shy", "componet_cmd": "source", "arguments": []interface{}{"@cmd"},
					"inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "cmd", "value": "",
							"class": "cmd", "clipstack": "void",
						},
					},
				},
				map[string]interface{}{"componet_name": "ctx", "componet_help": "ctx", "template": "componet",
					"componet_ctx": "cli.shy", "componet_cmd": "context", "arguments": []interface{}{"@ctx", "list"},
					"display_result": "",
					"inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "ctx", "value": "shy"},
						map[string]interface{}{"type": "button", "value": "refresh"},
					},
				},
				// map[string]interface{}{"componet_name": "ccc", "componet_help": "ccc", "template": "componet",
				// 	"componet_ctx": "cli.shy", "componet_cmd": "context", "arguments": []interface{}{"@current_ctx", "@ccc"},
				// 	"display_result": "",
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "choice", "name": "ccc",
				// 			"label": "ccc", "value": "command", "choice": []interface{}{
				// 				map[string]interface{}{"name": "command", "value": "command"},
				// 				map[string]interface{}{"name": "config", "value": "config"},
				// 				map[string]interface{}{"name": "cache", "value": "cache"},
				// 			},
				// 		},
				// 		map[string]interface{}{"type": "button", "value": "refresh"},
				// 	},
				// },
				// map[string]interface{}{"componet_name": "cmd", "componet_help": "cmd", "template": "componet",
				// 	"componet_ctx": "cli.shy", "componet_cmd": "context", "arguments": []interface{}{"@current_ctx", "command", "list"},
				// 	"pre_run": true, "display_result": "",
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "button", "value": "refresh"},
				// 	},
				// },
				// map[string]interface{}{"componet_name": "history", "componet_help": "history", "template": "componet",
				// 	"componet_ctx": "cli", "componet_cmd": "config", "arguments": []interface{}{"source_list"},
				// 	"pre_run": true, "display_result": "",
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "button", "value": "refresh"},
				// 	},
				// },
				// map[string]interface{}{"componet_name": "develop", "componet_help": "develop", "template": "componet",
				// 	"componet_ctx": "web.code", "componet_cmd": "config", "arguments": []interface{}{"counter"},
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "button", "value": "refresh"},
				// 	},
				// 	"pre_run":        true,
				// 	"display_result": "",
				// },
				// map[string]interface{}{"componet_name": "windows", "componet_help": "windows", "template": "componet",
				// 	"componet_ctx": "cli", "componet_cmd": "windows",
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "button", "value": "refresh"},
				// 	},
				// 	"pre_run":        true,
				// 	"display_result": "",
				// },
				map[string]interface{}{"componet_name": "runtime", "componet_help": "runtime", "template": "componet",
					"componet_ctx": "cli", "componet_cmd": "runtime",
					"inputs": []interface{}{
						map[string]interface{}{"type": "button", "value": "refresh"},
					},
					"pre_run":        true,
					"display_result": "",
				},
				// map[string]interface{}{"componet_name": "sysinfo", "componet_help": "sysinfo", "template": "componet",
				// 	"componet_ctx": "cli", "componet_cmd": "sysinfo",
				// 	"inputs": []interface{}{
				// 		map[string]interface{}{"type": "button", "value": "refresh"},
				// 	},
				// 	"pre_run":        true,
				// 	"display_result": "",
				// },
				map[string]interface{}{"componet_name": "tail", "template": "tail"},
			},
		}, Help: "组件列表"},
	},
	Commands: map[string]*ctx.Command{
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
	},
}

func init() {
	code := &web.WEB{}
	code.Context = Index
	web.Index.Register(Index, code)
}
