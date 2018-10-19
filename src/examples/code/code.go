package code

import (
	"bytes"
	"contexts/ctx"
	"contexts/web"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type CODE struct {
	web.WEB
}

// yac := m.Sess("tags", m.Sess("yac").Cmd("scan"))
// yac.Cmd("train", "void", "void", "[\t ]+")
// yac.Cmd("train", "other", "other", "[^\n]+")
// yac.Cmd("train", "key", "key", "[A-Za-z_][A-Za-z_0-9]*")
// yac.Cmd("train", "code", "def", "def", "key", "(", "other")
// yac.Cmd("train", "code", "def", "class", "key", "other")
// yac.Cmd("train", "code", "struct", "struct", "key", "\\{")
// yac.Cmd("train", "code", "struct", "\\}", "key", ";")
// yac.Cmd("train", "code", "struct", "typedef", "struct", "key", "key", ";")
// yac.Cmd("train", "code", "function", "key", "\\*", "key", "(", "other")
// yac.Cmd("train", "code", "function", "key", "key", "(", "other")
// yac.Cmd("train", "code", "variable", "struct", "key", "key", "other")
// yac.Cmd("train", "code", "define", "#define", "key", "other")
//

var Index = &ctx.Context{Name: "code", Help: "代码中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"library_dir":  &ctx.Config{Name: "library_dir", Value: "usr", Help: "通用模板路径"},
		"template_dir": &ctx.Config{Name: "template_dir", Value: "usr/template/", Help: "通用模板路径"},
		"common_tmpl":  &ctx.Config{Name: "common_tmpl", Value: "common/*.html", Help: "通用模板路径"},
		"common_main":  &ctx.Config{Name: "common_main", Value: "main.html", Help: "通用模板框架"},
		"upload_tmpl":  &ctx.Config{Name: "upload_tmpl", Value: "upload.html", Help: "上传文件模板"},
		"upload_main":  &ctx.Config{Name: "upload_main", Value: "main.html", Help: "上传文件框架"},
		"travel_tmpl":  &ctx.Config{Name: "travel_tmpl", Value: "travel.html", Help: "浏览模块模板"},
		"travel_main":  &ctx.Config{Name: "travel_main", Value: "main.html", Help: "浏览模块框架"},

		"check": &ctx.Config{Name: "check", Value: map[string]interface{}{
			"login": []interface{}{
				map[string]interface{}{
					"session": "aaa",
					"module":  "aaa", "command": "login",
					"variable": []interface{}{"$sessid"},
					"template": "login", "title": "login",
				},
				map[string]interface{}{
					"module": "aaa", "command": "login",
					"variable": []interface{}{"$username", "$password"},
					"template": "login", "title": "login",
				},
			},
			"right": []interface{}{
				map[string]interface{}{
					"module": "web", "command": "right",
					"variable": []interface{}{"$username", "check", "command", "/index", "dir", "$dir"},
					"template": "notice", "title": "notice",
				},
				map[string]interface{}{
					"module": "aaa", "command": "login",
					"variable": []interface{}{"username", "password"},
					"template": "login", "title": "login",
				},
			},
		}, Help: "执行条件"},
		"auto_create":  &ctx.Config{Name: "auto_create(true/false)", Value: "true", Help: "路由数量"},
		"refresh_time": &ctx.Config{Name: "refresh_time(ms)", Value: "1000", Help: "路由数量"},
		"define": &ctx.Config{Name: "define", Value: map[string]interface{}{
			"ngx_command_t": map[string]interface{}{
				"position": []interface{}{map[string]interface{}{
					"file": "nginx-1.15.2/src/core/ngx_core.h",
					"line": "22",
				}},
			},
			"ngx_command_s": map[string]interface{}{
				"position": map[string]interface{}{
					"file": "nginx-1.15.2/src/core/ngx_conf_file.h",
					"line": "77",
				},
			},
		}, Help: "路由数量"},
		"index": &ctx.Config{Name: "index", Value: map[string]interface{}{
			"duyu": []interface{}{
				map[string]interface{}{
					"template": "userinfo", "title": "userinfo",
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "aaa", "command": "lark",
					"argument": []interface{}{},
					"template": "append", "title": "lark_friend",
				},
				map[string]interface{}{
					"module": "aaa", "detail": []interface{}{"lark"},
					"template": "detail", "title": "send_lark",
					"option": map[string]interface{}{"ninput": 2},
				},
				map[string]interface{}{
					"module": "aaa", "command": "lark",
					"argument": []interface{}{"duyu"},
					"template": "append", "title": "lark",
				},
				map[string]interface{}{
					"module": "nfs", "command": "dir",
					"argument": []interface{}{"dir_type", "all", "dir_deep", "false", "dir_field", "time size line filename", "sort_field", "time", "sort_order", "time_r"},
					"template": "append", "title": "",
				},
			},
			"shy": []interface{}{
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"template": "userinfo", "title": "userinfo",
				},
				//文件服务
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "nfs", "command": "dir",
					"argument": []interface{}{"dir_type", "all", "dir_deep", "false", "dir_field", "time size line filename", "sort_field", "time", "sort_order", "time_r"},
					"template": "append", "title": "",
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"template": "upload", "title": "upload",
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"template": "create", "title": "create",
				},
				//会话服务
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "cli", "command": "system",
					"argument": []interface{}{"tmux", "show-buffer"},
					"template": "result", "title": "buffer",
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "cli", "command": "system",
					"argument": []interface{}{"tmux", "list-clients"},
					"template": "result", "title": "client",
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "cli", "command": "system",
					"argument": []interface{}{"tmux", "list-sessions"},
					"template": "result", "title": "session",
				},
				//格式转换
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "cli", "detail": []interface{}{"time"},
					"template": "detail", "title": "time",
					"option": map[string]interface{}{"refresh": true, "ninput": 1},
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "nfs", "detail": []interface{}{"json"},
					"template": "detail", "title": "json",
					"option": map[string]interface{}{"ninput": 1},
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "nfs", "detail": []interface{}{"pwd"},
					"template": "detail", "title": "pwd",
					"option": map[string]interface{}{"refresh": true},
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "nfs", "command": "git",
					"argument": []interface{}{},
					"template": "result", "title": "git",
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "web", "command": "/share",
					"argument": []interface{}{},
					"template": "share", "title": "share",
				},
			},
			"notice": []interface{}{
				map[string]interface{}{
					"template": "userinfo", "title": "userinfo",
				},
				map[string]interface{}{
					"template": "notice", "title": "notice",
				},
			},
			"login": []interface{}{
				map[string]interface{}{
					"template": "login", "title": "login",
				},
			},
			"wiki": []interface{}{
				map[string]interface{}{
					"template": "wiki_head", "title": "wiki_head",
				},
				map[string]interface{}{
					"template": "wiki_menu", "title": "wiki_menu",
				},
				map[string]interface{}{
					"module": "web", "command": "/wiki_list",
					"template": "wiki_list", "title": "wiki_list",
				},
				map[string]interface{}{
					"module": "web", "command": "/wiki_body",
					"template": "wiki_body", "title": "wiki_body",
				},
			},
		}, Help: "资源列表"},
	},
	Commands: map[string]*ctx.Command{
		"/demo": &ctx.Command{Name: "/demo", Help: "demo", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Echo("demo")
		}},
		"/render": &ctx.Command{Name: "/render index", Help: "模板响应, main: 模板入口, tmpl: 附加模板", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			w := m.Optionv("response").(http.ResponseWriter)
			w.Header().Add("Content-Type", "text/html")
			m.Optioni("ninput", 0)

			tpl := template.New("render").Funcs(ctx.CGI)
			tpl = template.Must(tpl.ParseGlob(path.Join(m.Conf("template_dir"), m.Conf("common_tmpl"))))
			tpl = template.Must(tpl.ParseGlob(path.Join(m.Conf("template_dir"), m.Conf("upload_tmpl"))))

			replace := [][]byte{
				[]byte{27, 91, 51, 50, 109}, []byte("<span style='color:red'>"),
				[]byte{27, 91, 51, 49, 109}, []byte("<span style='color:green'>"),
				[]byte{27, 91, 109}, []byte("</span>"),
			}

			if m.Confv("index", arg[0]) == nil {
				arg[0] = "notice"
			}

			m.Assert(tpl.ExecuteTemplate(w, "head", m))
			for _, v := range m.Confv("index", arg[0]).([]interface{}) {
				if v == nil {
					continue
				}
				val := v.(map[string]interface{})
				//命令模板
				if detail, ok := val["detail"].([]interface{}); ok {
					msg := m.Spawn().Add("detail", detail[0].(string), detail[1:])
					msg.Option("module", val["module"])
					msg.Option("title", val["title"])
					if option, ok := val["option"].(map[string]interface{}); ok {
						for k, v := range option {
							msg.Option(k, v)
						}
					}

					m.Assert(tpl.ExecuteTemplate(w, val["template"].(string), msg))
					continue
				}

				//执行命令
				if _, ok := val["command"]; ok {
					msg := m.Find(val["module"].(string)).Cmd(val["command"], val["argument"])
					for i, v := range msg.Meta["result"] {
						b := []byte(v)
						for i := 0; i < len(replace)-1; i += 2 {
							b = bytes.Replace(b, replace[i], replace[i+1], -1)
						}
						msg.Meta["result"][i] = string(b)
					}
					if msg.Option("title", val["title"]) == "" {
						msg.Option("title", m.Option("dir"))
					}
					m.Assert(tpl.ExecuteTemplate(w, val["template"].(string), msg))
					continue
				}

				//解析模板
				if _, ok := val["template"]; ok {
					m.Assert(tpl.ExecuteTemplate(w, val["template"].(string), m))
				}
			}
			m.Assert(tpl.ExecuteTemplate(w, "tail", m))
		}},

		"/check": &ctx.Command{Name: "/check arg...", Help: "权限检查, cache|config|command: 接口类型, name: 接口名称, args: 其它参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if login := m.Spawn().Cmd("/login"); login.Has("template") {
				m.Echo("no").Copy(login, "append")
				return
			}

			if msg := m.Spawn().Cmd("right", m.Append("username"), "check", arg); !msg.Results(0) {
				m.Echo("no").Append("message", "no right, please contact manager")
				return
			}

			m.Echo("ok")

		}},
		"/login": &ctx.Command{Name: "/login", Help: "用户登录", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if m.Options("sessid") {
				if aaa := m.Sess("aaa").Cmd("login", m.Option("sessid")); aaa.Results(0) {
					m.Append("redirect", m.Option("referer"))
					m.Append("username", aaa.Result(0))
					return
				}
			}

			w := m.Optionv("response").(http.ResponseWriter)
			if m.Options("username") && m.Options("password") {
				if aaa := m.Sess("aaa").Cmd("login", m.Option("username"), m.Option("password")); aaa.Results(0) {
					http.SetCookie(w, &http.Cookie{Name: "sessid", Value: aaa.Result(0)})
					m.Append("redirect", m.Option("referer"))
					m.Append("username", m.Option("username"))
					return
				}
			}

			w.WriteHeader(http.StatusUnauthorized)
			m.Append("template", "login")

		}},
		"/lookup": &ctx.Command{Name: "user", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if len(arg) > 0 {
				m.Option("service", arg[0])
			}
			msg := m.Sess("cli").Cmd("system", "sd", "lookup", m.Option("service"))

			rs := strings.Split(msg.Result(0), "\n")
			i := 0
			for ; i < len(rs); i++ {
				if len(rs[i]) == 0 {
					break
				}
				fields := strings.SplitN(rs[i], ": ", 2)
				m.Append(fields[0], fields[1])
			}

			lists := []interface{}{}
			for i += 2; i < len(rs); i++ {
				fields := strings.SplitN(rs[i], "  ", 3)
				if len(fields) < 3 {
					break
				}
				lists = append(lists, map[string]interface{}{
					"ip":   fields[0],
					"port": fields[1],
					"tags": fields[2],
				})
			}

			m.Appendv("lists", lists)
			m.Log("log", "%v", lists)
		}},
		"upload": &ctx.Command{Name: "upload file", Help: "上传文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Spawn(m.Target())
			msg.Cmd("get", "/upload", "method", "POST", "file", "file", arg[0])
			m.Copy(msg, "result")

		}},
		"/travel": &ctx.Command{Name: "/travel", Help: "文件上传", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			// r := m.Optionv("request").(*http.Request)
			// w := m.Optionv("response").(http.ResponseWriter)

			if !m.Options("dir") {
				m.Option("dir", "ctx")
			}

			check := m.Spawn().Cmd("/share", "/travel", "dir", m.Option("dir"))
			if !check.Results(0) {
				m.Copy(check, "append")
				return
			}

			// 权限检查
			if m.Option("method") == "POST" {
				if m.Options("domain") {
					msg := m.Find("ssh", true)
					msg.Detail(0, "send", "domain", m.Option("domain"), "context", "find", m.Option("dir"))
					if m.Option("name") != "" {
						msg.Add("detail", m.Option("name"))
					}
					if m.Options("value") {
						value := []string{}
						json.Unmarshal([]byte(m.Option("value")), &value)
						if len(value) > 0 {
							msg.Add("detail", value[0], value[1:])
						}
					}

					msg.CallBack(true, func(sub *ctx.Message) *ctx.Message {
						m.Copy(sub, "result").Copy(sub, "append")
						return nil
					})
					return
				}

				msg := m.Find(m.Option("dir"), true)
				if msg == nil {
					return
				}

				switch m.Option("ccc") {
				case "cache":
					m.Echo(msg.Cap(m.Option("name")))
				case "config":
					if m.Has("value") {
						m.Echo(msg.Conf(m.Option("name"), m.Option("value")))
					} else {
						m.Echo(msg.Conf(m.Option("name")))
					}
				case "command":
					msg = msg.Spawn(msg.Target())
					msg.Detail(0, m.Option("name"))
					if m.Options("value") {
						value := []string{}
						json.Unmarshal([]byte(m.Option("value")), &value)
						if len(value) > 0 {
							msg.Add("detail", value[0], value[1:])
						}
					}

					msg.Cmd()
					m.Copy(msg, "result").Copy(msg, "append")
				}
				return
			}

			// 解析模板
			m.Set("append", "tmpl", "userinfo", "share")
			msg := m
			for _, v := range []string{"cache", "config", "command", "module", "domain"} {
				if m.Options("domain") {
					msg = m.Find("ssh", true)
					msg.Detail(0, "send", "domain", m.Option("domain"), "context", "find", m.Option("dir"), "list", v)
					msg.CallBack(true, func(sub *ctx.Message) *ctx.Message {
						msg.Copy(sub, "result").Copy(sub, "append")
						return nil
					})
				} else {
					msg = m.Spawn()
					msg.Cmd("context", "find", msg.Option("dir"), "list", v)
				}

				if len(msg.Meta["append"]) > 0 {
					msg.Option("current_module", m.Option("dir"))
					msg.Option("current_domain", m.Option("domain"))
					m.Add("option", "tmpl", v)
					m.Sess(v, msg)
				}
			}
			m.Append("template", m.Conf("travel_main"), m.Conf("travel_tmpl"))

		}},
		"/index/": &ctx.Command{Name: "/index", Help: "网页门户", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)

			if login := m.Spawn().Cmd("/login"); login.Has("template") {
				m.Echo("no").Copy(login, "append")
				return
			}
			m.Option("username", m.Append("username"))

			//权限检查
			dir := m.Option("dir", path.Join(m.Cap("directory"), "local", m.Option("username"), m.Option("dir", strings.TrimPrefix(m.Option("path"), "/index"))))
			// if check := m.Spawn(c).Cmd("/check", "command", "/index/", "dir", dir); !check.Results(0) {
			// 	m.Copy(check, "append")
			// 	return
			// }

			//执行命令
			if m.Has("details") {
				if m.Confs("check_right") {
					if check := m.Spawn().Cmd("/check", "target", m.Option("module"), "command", m.Option("details")); !check.Results(0) {
						m.Copy(check, "append")
						return
					}
				}

				msg := m.Find(m.Option("module")).Cmd(m.Optionv("details"))
				m.Copy(msg, "result").Copy(msg, "append")
				return
			}

			//下载文件
			if s, e := os.Stat(dir); e == nil && !s.IsDir() {
				http.ServeFile(w, r, dir)
				return
			}

			if !m.Options("module") {
				m.Option("module", "web")
			}
			//浏览目录
			m.Append("template", m.Append("username"))
			m.Option("page_title", "index")

		}},
		"/create": &ctx.Command{Name: "/create", Help: "创建目录或文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			// if check := m.Spawn().Cmd("/share", "/upload", "dir", m.Option("dir")); !check.Results(0) {
			// 	m.Copy(check, "append")
			// 	return
			// }

			r := m.Optionv("request").(*http.Request)
			if m.Option("method") == "POST" {
				if m.Options("filename") { //添加文件或目录
					name := path.Join(m.Option("dir"), m.Option("filename"))
					if _, e := os.Stat(name); e != nil {
						if m.Options("content") {
							f, e := os.Create(name)
							m.Assert(e)
							defer f.Close()

							_, e = f.WriteString(m.Option("content"))
							m.Assert(e)
						} else {
							e = os.Mkdir(name, 0766)
							m.Assert(e)
						}
						m.Append("message", name, " create success!")
					} else {
						m.Append("message", name, " already exist!")
					}
				} else { //上传文件
					file, header, e := r.FormFile("file")
					m.Assert(e)

					name := path.Join(m.Option("dir"), header.Filename)

					if _, e := os.Stat(name); e != nil {
						f, e := os.Create(name)
						m.Assert(e)
						defer f.Close()

						_, e = io.Copy(f, file)
						m.Assert(e)
						m.Append("message", name, " upload success!")
					} else {
						m.Append("message", name, " already exist!")
					}
				}
			}
			m.Append("redirect", m.Option("referer"))

		}},
		"/share": &ctx.Command{Name: "/share arg...", Help: "资源共享", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if check := m.Spawn().Cmd("/check", "target", m.Option("module"), m.Optionv("share")); !check.Results(0) {
				m.Copy(check, "append")
				return
			}
			m.Option("username", m.Append("username"))

			// if m.Options("friend") && m.Options("module") {
			// 	m.Copy(m.Appendv("aaa").(*ctx.Message).Find(m.Option("module")).Cmd("right", m.Option("friend"), m.Option("action"), m.Optionv("share")), "result")
			// 	if m.Confv("index", m.Option("friend")) == nil {
			// 		m.Confv("index", m.Option("friend"), m.Confv("index", m.Append("username")))
			// 	}
			// 	return
			// }
			//
			// msg := m.Spawn().Cmd("right", "target", m.Option("module"), m.Append("username"), "show", "context")
			// m.Copy(msg, "append")
			if m.Options("friend") && m.Options("template") && m.Options("title") {
				for i, v := range m.Confv("index", m.Option("username")).([]interface{}) {
					if v == nil {
						continue
					}
					val := v.(map[string]interface{})
					if val["template"].(string) == m.Option("template") && val["title"].(string) == m.Option("title") {
						if m.Option("action") == "del" {
							friends := m.Confv("index", strings.Join([]string{m.Option("username"), fmt.Sprintf("%d", i), "to"}, ".")).([]interface{})
							for j, x := range friends {
								if x.(string) == m.Option("friend") {
									m.Confv("index", strings.Join([]string{m.Option("username"), fmt.Sprintf("%d", i), "to", fmt.Sprintf("%d", j)}, "."), nil)
								}
							}

							temps := m.Confv("index", strings.Join([]string{m.Option("friend")}, ".")).([]interface{})
							for j, x := range temps {
								if x == nil {
									continue
								}
								val = x.(map[string]interface{})
								if val["template"].(string) == m.Option("template") && val["title"].(string) == m.Option("title") {
									m.Confv("index", strings.Join([]string{m.Option("friend"), fmt.Sprintf("%d", j)}, "."), nil)
								}
							}

							break
						}

						if m.Confv("index", m.Option("friend")) == nil && !m.Confs("auto_create") {
							break
						}
						m.Confv("index", strings.Join([]string{m.Option("username"), fmt.Sprintf("%d", i), "to", "-2"}, "."), m.Option("friend"))

						item := map[string]interface{}{
							"template": val["template"],
							"title":    val["title"],
							"from":     m.Option("username"),
						}
						if val["command"] != nil {
							item["module"] = val["module"]
							item["command"] = val["command"]
							item["argument"] = val["argument"]
						} else if val["detail"] != nil {
							item["module"] = val["module"]
							item["detail"] = val["detail"]
							item["option"] = val["option"]
						}

						m.Confv("index", strings.Join([]string{m.Option("friend"), fmt.Sprintf("%d", -2)}, "."), item)
						m.Appendv("aaa").(*ctx.Message).Spawn(c).Cmd("right", m.Option("friend"), "add", "command", "/index/", "dir", m.Cap("directory"))
						os.Mkdir(path.Join(m.Cap("directory"), m.Option("friend")), 0666)
						break
					}
				}
				return
			}
			for _, v := range m.Confv("index", m.Option("username")).([]interface{}) {
				val := v.(map[string]interface{})
				m.Add("append", "template", val["template"])
				m.Add("append", "titles", val["title"])
				m.Add("append", "from", val["from"])
				m.Add("append", "to", "")
				if val["to"] == nil {
					continue
				}
				for _, u := range val["to"].([]interface{}) {
					m.Add("append", "template", val["template"])
					m.Add("append", "titles", val["title"])
					m.Add("append", "from", val["from"])
					m.Add("append", "to", u)
				}
			}

		}},
	},
}

func init() {
	code := &CODE{}
	code.Context = Index
	web.Index.Register(Index, code)
}
