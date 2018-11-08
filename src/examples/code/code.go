package code

import (
	"contexts/ctx"
	"contexts/web"
)

var Index = &ctx.Context{Name: "code", Help: "代码中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"web_site": &ctx.Config{Name: "web_site", Value: []interface{}{
			map[string]interface{}{"_name": "MDN", "site": "https://developer.mozilla.org"},
			map[string]interface{}{"_name": "github", "site": "https://github.com"},
		}, Help: "web_site"},
		"componet_command": &ctx.Config{Name: "component_command", Value: "pwd", Help: "默认命令"},
		"componet_group":   &ctx.Config{Name: "component_group", Value: "index", Help: "默认组件"},
		"componet": &ctx.Config{Name: "componet", Value: map[string]interface{}{
			"login": []interface{}{
				map[string]interface{}{"name": "head", "template": "head"},
				map[string]interface{}{"name": "userinfo", "help": "userinfo",
					"context": "aaa", "command": "userinfo", "arguments": []interface{}{"@sessid"},
				},
				map[string]interface{}{"name": "login", "help": "login", "template": "componet",
					"context": "aaa", "command": "login", "arguments": []interface{}{"@username", "@password"},
					"inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "username", "label": "username"},
						map[string]interface{}{"type": "password", "name": "password", "label": "password"},
						map[string]interface{}{"type": "button", "label": "login"},
					},
					"display_append": "", "display_result": "", "result_reload": "10",
				},
				map[string]interface{}{"name": "tail", "template": "tail"},
			},
			"index": []interface{}{
				map[string]interface{}{"name": "head", "template": "head"},
				map[string]interface{}{"name": "sysinfo", "help": "sysinfo", "template": "componet",
					"context": "cli", "command": "sysinfo",
					"inputs": []interface{}{
						map[string]interface{}{"type": "button", "label": "refresh"},
					},
					"pre_run":        true,
					"display_result": "",
				},
				map[string]interface{}{"name": "runtime", "help": "runtime", "template": "componet",
					"context": "cli", "command": "runtime",
					"inputs": []interface{}{
						map[string]interface{}{"type": "button", "label": "refresh"},
					},
					"pre_run":        true,
					"display_result": "",
				},
				map[string]interface{}{"name": "develop", "help": "develop", "template": "componet",
					"context": "cli", "command": "develop",
					"inputs": []interface{}{
						map[string]interface{}{"type": "button", "label": "refresh"},
					},
					"pre_run":        true,
					"display_result": "",
				},
				map[string]interface{}{"name": "clipbaord", "help": "clipbaord", "template": "clipboard"},
				map[string]interface{}{"name": "buffer", "help": "buffer", "template": "componet",
					"context": "cli", "command": "tmux", "arguments": []interface{}{"buffer"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "limit", "label": "limit", "value": "3"},
						map[string]interface{}{"type": "text", "name": "index", "label": "index"},
						map[string]interface{}{"type": "button", "label": "refresh"},
					},
					"pre_run": true,
				},
				map[string]interface{}{"name": "command", "help": "command", "template": "componet",
					"context": "cli.shy", "command": "source", "arguments": []interface{}{"@cmd"},
					"inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "cmd", "value": "",
							"class": "cmd", "clipstack": "clistack",
						},
					},
				},
				map[string]interface{}{"name": "time", "help": "time", "template": "componet",
					"context": "cli", "command": "time", "arguments": []interface{}{"@string"},
					"inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "time_format",
							"label": "format", "value": "2006-01-02 15:04:05",
						},
						map[string]interface{}{"type": "text", "name": "string", "label": "string"},
						map[string]interface{}{"type": "button", "label": "refresh"},
					},
				},
				map[string]interface{}{"name": "json", "help": "json", "template": "componet",
					"context": "nfs", "command": "json", "arguments": []interface{}{"@string"},
					"inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "string", "label": "string"},
						map[string]interface{}{"type": "button", "label": "refresh"},
					},
				},
				map[string]interface{}{"name": "upload", "help": "upload", "template": "componet",
					"context": "web", "command": "upload", "form_type": "upload",
					"inputs": []interface{}{
						map[string]interface{}{"type": "file", "name": "upload"},
						map[string]interface{}{"type": "submit", "value": "submit"},
					},
					"display_result": "",
				},
				map[string]interface{}{"name": "dir", "help": "dir", "template": "componet",
					"context": "nfs", "command": "dir", "arguments": []interface{}{"@dir",
						"dir_deep", "no", "dir_name", "name", "dir_info", "",
						"dir_link", "<a class='download' data-type='%s'>%s<a>",
					},
					"pre_run": true,
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
								map[string]interface{}{"name": "is_dir", "value": "is_dir"},
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
				map[string]interface{}{"name": "tail", "template": "tail"},
			},
		}, Help: "组件列表"},
	},
	Commands: map[string]*ctx.Command{},
}

func init() {
	code := &web.WEB{}
	code.Context = Index
	web.Index.Register(Index, code)
}
