package web

import (
	"bytes"
	"contexts/ctx"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type MUX interface {
	Handle(string, http.Handler)
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
	HandleCmd(*ctx.Message, string, func(*ctx.Message, *ctx.Context, string, ...string))
	ServeHTTP(http.ResponseWriter, *http.Request)
}
type WEB struct {
	client *http.Client
	*http.ServeMux
	*http.Server

	*ctx.Context
}

func (web *WEB) Merge(m *ctx.Message, uri string, arg ...string) string {
	add, e := url.Parse(uri)
	m.Assert(e)
	adds := []string{m.Confx("protocol", add.Scheme, "%s://"), m.Confx("hostname", add.Host)}

	if dir, file := path.Split(add.EscapedPath()); path.IsAbs(dir) {
		adds = append(adds, dir)
		adds = append(adds, file)
	} else {
		adds = append(adds, m.Conf("path"))
		if dir == "" && file == "" {
			adds = append(adds, m.Conf("file"))
		} else {
			adds = append(adds, dir)
			adds = append(adds, file)
		}
	}

	args := []string{}
	for i := 0; i < len(arg)-1; i += 2 {
		value := arg[i+1]
		if len(arg[i+1]) > 1 {
			switch arg[i+1][0] {
			case '$':
				value = m.Cap(arg[i+1][1:])
			case '@':
				value = m.Conf(arg[i+1][1:])
			}
		}
		args = append(args, arg[i]+"="+url.QueryEscape(value))
	}

	query := strings.Join(args, "&")
	if query == "" {
		query = add.RawQuery
	} else if add.RawQuery != "" {
		query = add.RawQuery + "&" + query
	}
	adds = append(adds, m.Confx("query", query, "?%s"))

	return strings.Join(adds, "")
}
func (web *WEB) HandleCmd(m *ctx.Message, key string, hand func(*ctx.Message, *ctx.Context, string, ...string)) {
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		msg := m.Spawn()
		msg.TryCatch(msg, true, func(msg *ctx.Message) {
			msg.Option("method", r.Method)
			msg.Option("path", r.URL.Path)
			msg.Option("referer", r.Header.Get("Referer"))

			remote := r.RemoteAddr
			if r.Header.Get("X-Real-Ip") != "" {
				remote = r.Header.Get("X-Real-Ip")
			}

			count := 1
			if m.Confv("record", []interface{}{r.URL.Path}) == nil {
				m.Confv("record", []interface{}{r.URL.Path}, map[string]interface{}{
					"remote": map[string]interface{}{remote: map[string]interface{}{"time": time.Now().Format("2006/01/02 15:04:05")}},
					"count":  count,
				})
			} else {
				switch v := m.Confv("record", []interface{}{r.URL.Path, "count"}).(type) {
				case int:
					count = v
				case float64:
					count = int(v)
				default:
					count = 0
				}

				if m.Confv("record", []interface{}{r.URL.Path, "remote", remote}) == nil {
					m.Confv("record", []interface{}{r.URL.Path, "count"}, count+1)
				} else {
					msg.Option("last_record_time", m.Confv("record", []interface{}{r.URL.Path, "remote", remote, "time"}))
				}
				m.Confv("record", []interface{}{r.URL.Path, "remote", remote}, map[string]interface{}{"time": time.Now().Format("2006/01/02 15:04:05")})
			}
			msg.Option("record_count", count)

			if r.ParseForm(); len(r.PostForm) > 0 {
				for k, v := range r.PostForm {
					m.Log("info", "%s: %v", k, v)
				}
				m.Log("info", "")
			}
			for _, v := range r.Cookies() {
				msg.Option(v.Name, v.Value)
			}
			for k, v := range r.Form {
				msg.Add("option", k, v)
			}

			msg.Log("cmd", "%s [] %v", key, msg.Meta["option"])
			msg.Put("option", "request", r).Put("option", "response", w)
			hand(msg, msg.Target(), msg.Option("path"))

			switch {
			case msg.Has("redirect"):
				http.Redirect(w, r, msg.Append("redirect"), http.StatusFound)
			case msg.Has("directory"):
				http.ServeFile(w, r, msg.Append("directory"))
			case msg.Has("template"):
				msg.Spawn().Cmd("/render", msg.Meta["template"])
			case msg.Has("append"):
				msg.Spawn().Copy(msg, "result").Copy(msg, "append").Cmd("/json")
			default:
				for _, v := range msg.Meta["result"] {
					w.Write([]byte(v))
				}
			}
		})
	})
}
func (web *WEB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := web.Message().Log("info", "").Log("info", "%v %s %s", r.RemoteAddr, r.Method, r.URL)

	if m.Confs("logheaders") {
		for k, v := range r.Header {
			m.Log("info", "%s: %v", k, v)
		}
		m.Log("info", "")
	}

	if r.URL.Path == "/" && m.Confs("root_index") {
		http.Redirect(w, r, m.Conf("root_index"), http.StatusFound)
	} else {
		web.ServeMux.ServeHTTP(w, r)
	}

	if m.Confs("logheaders") {
		for k, v := range w.Header() {
			m.Log("info", "%s: %v", k, v)
		}
		m.Log("info", "")
	}
}

func (web *WEB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(WEB)
	s.Context = c
	return s
}
func (web *WEB) Begin(m *ctx.Message, arg ...string) ctx.Server {
	web.Configs["logheaders"] = &ctx.Config{Name: "logheaders(yes/no)", Value: "no", Help: "日志输出报文头"}
	web.Configs["root_index"] = &ctx.Config{Name: "root_index", Value: "/wiki/", Help: "默认路由"}
	web.Caches["directory"] = &ctx.Cache{Name: "directory", Value: m.Confx("directory", arg, 0), Help: "服务目录"}
	web.Caches["route"] = &ctx.Cache{Name: "route", Value: "/" + web.Context.Name + "/", Help: "模块路由"}
	web.Caches["register"] = &ctx.Cache{Name: "register(yes/no)", Value: "no", Help: "是否已初始化"}
	web.Caches["master"] = &ctx.Cache{Name: "master(yes/no)", Value: "no", Help: "服务入口"}
	web.ServeMux = http.NewServeMux()
	return web
}
func (web *WEB) Start(m *ctx.Message, arg ...string) bool {
	m.Cap("directory", m.Confx("directory", arg, 0))

	m.Travel(func(m *ctx.Message, i int) bool {
		if h, ok := m.Target().Server.(MUX); ok && m.Cap("register") == "no" {
			m.Cap("register", "yes")

			p := m.Target().Context()
			if s, ok := p.Server.(MUX); ok {
				m.Log("info", "route: /%s <- /%s", p.Name, m.Target().Name)
				s.Handle(m.Cap("route"), http.StripPrefix(path.Dir(m.Cap("route")), h))
			}

			for k, x := range m.Target().Commands {
				if k[0] == '/' {
					m.Log("info", "route: %s", k)
					h.HandleCmd(m, k, x.Hand)
					m.Capi("nroute", 1)
				}
			}

			if m.Cap("directory") != "" {
				m.Log("info", "route: %s <- [%s]\n", m.Cap("route"), m.Cap("directory"))
				h.Handle("/", http.FileServer(http.Dir(m.Cap("directory"))))
			}
		}
		return true
	})

	web.Configs["library_dir"] = &ctx.Config{Name: "library_dir", Value: "usr", Help: "通用模板路径"}
	web.Configs["template_dir"] = &ctx.Config{Name: "template_dir", Value: "usr/template/", Help: "通用模板路径"}
	web.Configs["common_tmpl"] = &ctx.Config{Name: "common_tmpl", Value: "common/*.html", Help: "通用模板路径"}
	web.Configs["common_main"] = &ctx.Config{Name: "common_main", Value: "main.html", Help: "通用模板框架"}
	web.Configs["upload_tmpl"] = &ctx.Config{Name: "upload_tmpl", Value: "upload.html", Help: "上传文件模板"}
	web.Configs["upload_main"] = &ctx.Config{Name: "upload_main", Value: "main.html", Help: "上传文件框架"}
	web.Configs["travel_tmpl"] = &ctx.Config{Name: "travel_tmpl", Value: "travel.html", Help: "浏览模块模板"}
	web.Configs["travel_main"] = &ctx.Config{Name: "travel_main", Value: "main.html", Help: "浏览模块框架"}

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

	web.Caches["protocol"] = &ctx.Cache{Name: "protocol", Value: m.Confx("protocol", arg, 2), Help: "服务协议"}
	web.Caches["address"] = &ctx.Cache{Name: "address", Value: m.Confx("address", arg, 1), Help: "服务地址"}
	m.Log("info", "%d %s://%s", m.Capi("nserve", 1), m.Cap("protocol"), m.Cap("stream", m.Cap("address")))
	web.Server = &http.Server{Addr: m.Cap("address"), Handler: web}

	if m.Caps("master", true); m.Cap("protocol") == "https" {
		web.Caches["cert"] = &ctx.Cache{Name: "cert", Value: m.Confx("cert", arg, 3), Help: "服务证书"}
		web.Caches["key"] = &ctx.Cache{Name: "key", Value: m.Confx("key", arg, 4), Help: "服务密钥"}
		m.Log("info", "cert [%s]", m.Cap("cert"))
		m.Log("info", "key [%s]", m.Cap("key"))
		web.Server.ListenAndServeTLS(m.Cap("cert"), m.Cap("key"))
	} else {
		web.Server.ListenAndServe()
	}
	return true
}
func (web *WEB) Close(m *ctx.Message, arg ...string) bool {
	switch web.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

var Index = &ctx.Context{Name: "web", Help: "应用中心",
	Caches: map[string]*ctx.Cache{
		"nserve": &ctx.Cache{Name: "nserve", Value: "0", Help: "主机数量"},
		"nroute": &ctx.Cache{Name: "nroute", Value: "0", Help: "路由数量"},
	},
	Configs: map[string]*ctx.Config{
		"body_response": &ctx.Config{Name: "body_response", Value: "response", Help: "响应缓存"},
		"method":        &ctx.Config{Name: "method", Value: "GET", Help: "请求方法"},
		"brow_home":     &ctx.Config{Name: "brow_home", Value: "http://localhost:9094", Help: "服务"},
		"directory":     &ctx.Config{Name: "directory", Value: "usr", Help: "服务目录"},
		"address":       &ctx.Config{Name: "address", Value: ":9094", Help: "服务地址"},
		"protocol":      &ctx.Config{Name: "protocol", Value: "http", Help: "服务协议"},
		"cert":          &ctx.Config{Name: "cert", Value: "etc/cert.pem", Help: "路由数量"},
		"key":           &ctx.Config{Name: "key", Value: "etc/key.pem", Help: "路由数量"},

		"record":       &ctx.Config{Name: "record", Value: map[string]interface{}{}, Help: "访问记录"},
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
		"client": &ctx.Command{Name: "client address [output [editor]]", Help: "添加浏览器配置, address: 默认地址, output: 输出路径, editor: 编辑器", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			uri, e := url.Parse(arg[0])
			m.Assert(e)
			m.Conf("method", "method", "GET", "请求方法")
			m.Conf("protocol", "protocol", uri.Scheme, "服务协议")
			m.Conf("hostname", "hostname", uri.Host, "服务主机")

			dir, file := path.Split(uri.EscapedPath())
			m.Conf("path", "path", dir, "服务路由")
			m.Conf("file", "file", file, "服务文件")
			m.Conf("query", "query", uri.RawQuery, "服务参数")

			if m.Conf("output", "output", "stdout", "文件缓存"); len(arg) > 1 {
				m.Conf("output", arg[1])
			}
			if m.Conf("editor", "editor", "vim", "文件编辑器"); len(arg) > 2 {
				m.Conf("editor", arg[2])
			}
		}},
		"cookie": &ctx.Command{Name: "cookie [create]|[name [value]]", Help: "读写浏览器的Cookie, create: 创建cookiejar, name: 变量名, value: 变量值", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			switch len(arg) {
			case 0:
				for k, v := range m.Confv("cookie").(map[string]interface{}) {
					m.Echo("%s: %v\n", k, v.(*http.Cookie).Value)
				}
			case 1:
				if arg[0] == "create" {
					m.Target().Configs["cookie"] = &ctx.Config{Name: "cookie", Value: map[string]interface{}{}, Help: "cookie"}
					break
				}
				if v, ok := m.Confv("cookie", arg[0]).(*http.Cookie); ok {
					m.Echo("%s", v.Value)
				}
			default:
				if m.Confv("cookie") == nil {
					m.Target().Configs["cookie"] = &ctx.Config{Name: "cookie", Value: map[string]interface{}{}, Help: "cookie"}
				}
				if v, ok := m.Confv("cookie", arg[0]).(*http.Cookie); ok {
					v.Value = arg[1]
				} else {
					m.Confv("cookie", arg[0], &http.Cookie{Name: arg[0], Value: arg[1]})
				}
			}
		}},
		"get": &ctx.Command{Name: "get [method GET|POST] url arg...",
			Help: "访问服务, method: 请求方法, url: 请求地址, arg: 请求参数",
			Form: map[string]int{"method": 1, "headers": 2, "content_type": 1, "body": 1, "path_value": 1, "body_response": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if web, ok := m.Target().Server.(*WEB); m.Assert(ok) {
					if m.Has("path_value") {
						values := []interface{}{}
						for _, v := range strings.Split(m.Option("path_value"), " ") {
							if len(v) > 1 && v[0] == '$' {
								values = append(values, m.Cap(v[1:]))
							} else {
								values = append(values, v)
							}
						}
						arg[0] = fmt.Sprintf(arg[0], values...)
					}

					method := m.Confx("method")
					uri := web.Merge(m, arg[0], arg[1:]...)
					body, _ := m.Optionv("body").(io.Reader)

					if method == "POST" && body == nil {
						if index := strings.Index(uri, "?"); index > 0 {
							uri, body = uri[:index], strings.NewReader(uri[index+1:])
						}
					}

					req, e := http.NewRequest(method, uri, body)
					m.Assert(e)

					for i := 0; i < len(m.Meta["headers"]); i += 2 {
						req.Header.Set(m.Meta["headers"][i], m.Meta["headers"][i+1])
					}
					if m.Options("content_type") {
						req.Header.Set("Content-Type", m.Option("content_type"))
					}
					switch cs := m.Confv("cookie").(type) {
					case map[string]interface{}:
						for _, v := range cs {
							req.AddCookie(v.(*http.Cookie))
						}
					}

					m.Log("info", "%s %s", req.Method, req.URL)
					m.Echo("%s: %s\n", req.Method, req.URL)
					for k, v := range req.Header {
						m.Log("fuck", "%s: %s", k, v)
					}

					if web.client == nil {
						web.client = &http.Client{}
					}
					res, e := web.client.Do(req)
					m.Assert(e)

					for _, v := range res.Cookies() {
						m.Confv("cookie", v.Name, v)
						m.Log("info", "set-cookie %s: %v", v.Name, v.Value)
					}

					buf, e := ioutil.ReadAll(res.Body)
					m.Assert(e)

					var result interface{}
					ct := res.Header.Get("Content-Type")
					switch {
					case strings.HasPrefix(ct, "application/json"):
						json.Unmarshal(buf, &result)
					default:
						result = string(buf)
					}
					m.Target().Configs[m.Confx("body_response")] = &ctx.Config{Value: result}
					m.Echo(string(buf))
				}
			}},
		"post": &ctx.Command{Name: "post [file fieldname filename]", Help: "post请求",
			Form: map[string]int{"file": 2, "content_type": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				msg := m.Spawn()
				if m.Has("file") {
					file, e := os.Open(m.Meta["file"][1])
					m.Assert(e)
					defer file.Close()

					buf := &bytes.Buffer{}
					writer := multipart.NewWriter(buf)
					writer.SetBoundary(fmt.Sprintf("\r\n--%s--\r\n", writer.Boundary()))
					part, e := writer.CreateFormFile(m.Option("file"), filepath.Base(m.Meta["file"][1]))
					m.Assert(e)
					io.Copy(part, file)

					// for i := 0; i < len(arg)-1; i += 2 {
					// 	value := arg[i+1]
					// 	if len(arg[i+1]) > 1 {
					// 		switch arg[i+1][0] {
					// 		case '$':
					// 			value = m.Cap(arg[i+1][1:])
					// 		case '@':
					// 			value = m.Conf(arg[i+1][1:])
					// 		}
					// 	}
					// 	writer.WriteField(arg[i], value)
					// }

					writer.Close()
					msg.Optionv("body", buf)
					msg.Option("content_type", writer.FormDataContentType())
					msg.Option("headers", "Content-Length", buf.Len())
				} else if m.Option("content_type") == "json" {
					data := map[string]interface{}{}
					for i := 1; i < len(arg)-1; i += 2 {
						switch arg[i+1] {
						case "false":
							data[arg[i]] = false
						case "true":
							data[arg[i]] = true
						default:
							if len(arg[i+1]) > 1 && arg[i+1][0] == '$' {
								data[arg[i]] = m.Cap(arg[i+1][1:])
							} else {
								data[arg[i]] = arg[i+1]
							}
						}
					}
					b, e := json.Marshal(data)
					m.Assert(e)
					msg.Optionv("body", bytes.NewReader(b))
					msg.Option("content_type", "application/json")
					arg = arg[:1]
				} else if m.Option("content_type") == "json_fmt" {
					data := []interface{}{}
					for _, v := range arg[2:] {
						if len(v) > 1 && v[0] == '$' {
							v = m.Cap(v[1:])
						} else if len(v) > 1 && v[0] == '@' {
							v = m.Cap(v[1:])
						}
						data = append(data, v)
					}
					msg.Optionv("body", strings.NewReader(fmt.Sprintf(arg[1], data...)))
					msg.Option("content_type", "application/json")
					arg = arg[:1]
				} else {
					msg.Option("content_type", "application/x-www-form-urlencoded")
				}
				msg.Cmd("get", "method", "POST", arg)
				m.Copy(msg, "result").Copy(msg, "append")
			}},
		"brow": &ctx.Command{Name: "brow url", Help: "浏览网页", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			url := m.Confx("brow_home", arg, 0)
			switch runtime.GOOS {
			case "windows":
				m.Sess("cli").Cmd("system", "explorer", url)
			case "darwin":
				m.Sess("cli").Cmd("system", "open", url)
			case "linux":
				m.Spawn().Cmd("open", url)
			}
		}},
		"serve": &ctx.Command{Name: "serve [directory [address [protocol [cert [key]]]]", Help: "启动服务, directory: 服务路径, address: 服务地址, protocol: 服务协议(https/http), cert: 服务证书, key: 服务密钥", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Set("detail", arg...).Target().Start(m)
		}},
		"route": &ctx.Command{Name: "route script|template|directory route content", Help: "添加响应, script: 脚本响应, template: 模板响应, directory: 目录响应, route: 请求路由, content: 响应内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if mux, ok := m.Target().Server.(MUX); m.Assert(ok) {
				switch len(arg) {
				case 0:
					for k, v := range m.Target().Commands {
						if k[0] == '/' {
							m.Echo("%s: %s\n", k, v.Name)
						}
					}
				case 1:
					for k, v := range m.Target().Commands {
						if k == arg[0] {
							m.Echo("%s: %s\n%s", k, v.Name, v.Help)
						}
					}
				case 3:
					switch arg[0] {
					case "script":
						mux.HandleCmd(m, arg[1], func(m *ctx.Message, c *ctx.Context, key string, a ...string) {
							msg := m.Sess("cli").Cmd("source", arg[2])
							m.Copy(msg, "result").Copy(msg, "append")
						})
					case "template":
						mux.HandleCmd(m, arg[1], func(m *ctx.Message, c *ctx.Context, key string, a ...string) {
							w := m.Optionv("response").(http.ResponseWriter)
							if _, e := os.Stat(arg[2]); e == nil {
								template.Must(template.ParseGlob(arg[2])).Execute(w, m)
							} else {
								template.Must(template.New("temp").Parse(arg[2])).Execute(w, m)
							}
						})
					case "directory":
						mux.Handle(arg[1]+"/", http.StripPrefix(arg[1], http.FileServer(http.Dir(arg[2]))))
					}
				}
			}
		}},
		"upload": &ctx.Command{Name: "upload file", Help: "上传文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Spawn(m.Target())
			msg.Cmd("get", "/upload", "method", "POST", "file", "file", arg[0])
			m.Copy(msg, "result")

		}},
		"/library/": &ctx.Command{Name: "/library", Help: "网页门户", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)
			dir := path.Join(m.Conf("library_dir"), m.Option("path"))
			if s, e := os.Stat(dir); e == nil && !s.IsDir() {
				http.ServeFile(w, r, dir)
				return
			}

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
		"/json": &ctx.Command{Name: "/json", Help: "json响应", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			w := m.Optionv("response").(http.ResponseWriter)

			meta := map[string]interface{}{}
			if len(m.Meta["result"]) > 0 {
				meta["result"] = m.Meta["result"]
			}
			if len(m.Meta["append"]) > 0 {
				meta["append"] = m.Meta["append"]
				for _, v := range m.Meta["append"] {
					if _, ok := m.Data[v]; ok {
						meta[v] = m.Data[v]
					} else if _, ok := m.Meta[v]; ok {
						meta[v] = m.Meta[v]
					}
				}
			}

			if b, e := json.Marshal(meta); m.Assert(e) {
				w.Header().Set("Content-Type", "application/javascript")
				w.Write(b)
			}
		}},
		"/paste": &ctx.Command{Name: "/paste", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if login := m.Spawn().Cmd("/login"); login.Has("redirect") {
				m.Sess("cli").Cmd("system", "tmux", "set-buffer", "-b", "0", m.Option("content"))
			}

		}},
		"/upload": &ctx.Command{Name: "/upload", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			r := m.Optionv("request").(*http.Request)
			f, h, e := r.FormFile("file")
			lf, e := os.Create(fmt.Sprintf("tmp/%s", h.Filename))
			m.Assert(e)
			io.Copy(lf, f)
		}},
		"user": &ctx.Command{Name: "user", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			aaa := m.Sess("aaa")
			m.Spawn().Cmd("get", fmt.Sprintf("%suser/get", aaa.Conf("wx_api")), "access_token", aaa.Cap("access_token"))
		}},
		"temp": &ctx.Command{Name: "temp", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Spawn(m.Target())
			question := []string{}
			for i := 1; i < 21; i++ {
				question = append(question, fmt.Sprintf("{\"type\":\"1001\",\"title\":{\"text\":\"第%d题\"}}", i))
			}
			qs := "[" + strings.Join(question, ",") + "]"

			msg.Cmd("get", "method", "POST", "evaluating_add/", "questions", qs)
			m.Add("append", "hi", "hello")
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
		"old_get": &ctx.Command{
			Name: "get [method GET|POST] [file name filename] url arg...",
			Help: "访问服务, method: 请求方法, file: 发送文件, url: 请求地址, arg: 请求参数",
			Form: map[string]int{"method": 1, "content_type": 1, "headers": 2, "file": 2, "body_type": 1, "body": 1, "fields": 1, "value": 1, "json_route": 1, "json_key": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if web, ok := m.Target().Server.(*WEB); m.Assert(ok) {
					if web.client == nil {
						web.client = &http.Client{}
					}

					if m.Has("value") {
						args := strings.Split(m.Option("value"), " ")
						values := []interface{}{}
						for _, v := range args {
							if len(v) > 1 && v[0] == '$' {
								values = append(values, m.Cap(v[1:]))
							} else {
								values = append(values, v)
							}
						}
						arg[0] = fmt.Sprintf(arg[0], values...)
					}

					method := m.Confx("method")
					uri := web.Merge(m, arg[0], arg[1:]...)
					m.Log("info", "%s %s", method, uri)
					m.Echo("%s: %s\n", method, uri)

					var body io.Reader
					index := strings.Index(uri, "?")
					content_type := ""

					switch method {
					case "POST":
						if m.Options("file") {
							file, e := os.Open(m.Meta["file"][1])
							m.Assert(e)
							defer file.Close()

							if m.Option("body_type") == "json" {
								content_type = "application/json"
								body = file
								break
							}
							buf := &bytes.Buffer{}
							writer := multipart.NewWriter(buf)

							part, e := writer.CreateFormFile(m.Option("file"), filepath.Base(m.Meta["file"][1]))
							m.Assert(e)
							io.Copy(part, file)

							for i := 0; i < len(arg)-1; i += 2 {
								value := arg[i+1]
								if len(arg[i+1]) > 1 {
									switch arg[i+1][0] {
									case '$':
										value = m.Cap(arg[i+1][1:])
									case '@':
										value = m.Conf(arg[i+1][1:])
									}
								}
								writer.WriteField(arg[i], value)
							}

							content_type = writer.FormDataContentType()
							body = buf
							writer.Close()
						} else if m.Option("body_type") == "json" {
							if m.Options("body") {
								data := []interface{}{}
								for _, v := range arg[1:] {
									if len(v) > 1 && v[0] == '$' {
										v = m.Cap(v[1:])
									}
									data = append(data, v)
								}
								body = strings.NewReader(fmt.Sprintf(m.Option("body"), data...))
							} else {
								data := map[string]interface{}{}
								for i := 1; i < len(arg)-1; i += 2 {
									switch arg[i+1] {
									case "false":
										data[arg[i]] = false
									case "true":
										data[arg[i]] = true
									default:
										if len(arg[i+1]) > 1 && arg[i+1][0] == '$' {
											data[arg[i]] = m.Cap(arg[i+1][1:])
										} else {
											data[arg[i]] = arg[i+1]
										}
									}
								}

								b, e := json.Marshal(data)
								m.Assert(e)
								body = bytes.NewReader(b)
							}

							content_type = "application/json"
							if index > -1 {
								uri = uri[:index]
							}

						} else if index > 0 {
							content_type = "application/x-www-form-urlencoded"
							body = strings.NewReader(uri[index+1:])
							uri = uri[:index]
						}
					}

					req, e := http.NewRequest(method, uri, body)
					m.Assert(e)
					for i := 0; i < len(m.Meta["headers"]); i += 2 {
						req.Header.Set(m.Meta["headers"][i], m.Meta["headers"][i+1])
					}

					if len(content_type) > 0 {
						req.Header.Set("Content-Type", content_type)
						m.Log("info", "content-type: %s", content_type)
					}

					for _, v := range m.Confv("cookie").(map[string]interface{}) {
						req.AddCookie(v.(*http.Cookie))
					}

					res, e := web.client.Do(req)
					m.Assert(e)

					for _, v := range res.Cookies() {
						m.Confv("cookie", v.Name, v)
						m.Log("info", "set-cookie %s: %v", v.Name, v.Value)
					}

					if m.Confs("logheaders") {
						for k, v := range res.Header {
							m.Log("info", "%s: %v", k, v)
						}
					}

					if m.Confs("output") {
						if _, e := os.Stat(m.Conf("output")); e == nil {
							name := path.Join(m.Conf("output"), fmt.Sprintf("%d", time.Now().Unix()))
							f, e := os.Create(name)
							m.Assert(e)
							io.Copy(f, res.Body)
							if m.Confs("editor") {
								cmd := exec.Command(m.Conf("editor"), name)
								cmd.Stdin = os.Stdin
								cmd.Stdout = os.Stdout
								cmd.Stderr = os.Stderr
								cmd.Run()
							} else {
								m.Echo("write to %s\n", name)
							}
							return
						}
					}

					buf, e := ioutil.ReadAll(res.Body)
					m.Assert(e)

					ct := res.Header.Get("Content-Type")
					if len(ct) >= 16 && ct[:16] == "application/json" {
						var result interface{}
						json.Unmarshal(buf, &result)
						m.Option("response_json", result)
						if m.Has("json_route") {
							routes := strings.Split(m.Option("json_route"), ".")
							for _, k := range routes {
								if len(k) > 0 && k[0] == '$' {
									k = m.Cap(k[1:])
								}
								switch r := result.(type) {
								case map[string]interface{}:
									result = r[k]
								}
							}
						}

						fields := map[string]bool{}
						for _, k := range strings.Split(m.Option("fields"), " ") {
							if k == "" {
								continue
							}
							if fields[k] = true; len(fields) == 1 {
								m.Meta["append"] = append(m.Meta["append"], "index")
							}
							m.Meta["append"] = append(m.Meta["append"], k)
						}

						if len(fields) > 0 {

							switch ret := result.(type) {
							case map[string]interface{}:
								m.Append("index", "0")
								for k, v := range ret {
									switch value := v.(type) {
									case string:
										m.Append(k, strings.Replace(value, "\n", " ", -1))
									case float64:
										m.Append(k, fmt.Sprintf("%d", int(value)))
									default:
										if _, ok := fields[k]; ok {
											m.Append(k, fmt.Sprintf("%v", value))
										}
									}
								}
							case []interface{}:
								for i, r := range ret {
									m.Add("append", "index", i)
									if rr, ok := r.(map[string]interface{}); ok {
										for k, v := range rr {
											switch value := v.(type) {
											case string:
												if _, ok := fields[k]; len(fields) == 0 || ok {
													m.Add("append", k, strings.Replace(value, "\n", " ", -1))
												}
											case float64:
												if _, ok := fields[k]; len(fields) == 0 || ok {
													m.Add("append", k, fmt.Sprintf("%d", int64(value)))
												}
											case bool:
												if _, ok := fields[k]; len(fields) == 0 || ok {
													m.Add("append", k, fmt.Sprintf("%v", value))
												}
											case map[string]interface{}:
												for kk, vv := range value {
													key := k + "." + kk
													if _, ok := fields[key]; len(fields) == 0 || ok {
														m.Add("append", key, strings.Replace(fmt.Sprintf("%v", vv), "\n", " ", -1))
													}
												}
											default:
												if _, ok := fields[k]; ok {
													m.Add("append", k, fmt.Sprintf("%v", value))
												}
											}
										}
									}
								}

								if m.Has("json_key") {
									m.Sort(m.Option("json_key"))
								}
								m.Meta["index"] = nil
								for i, _ := range ret {
									m.Add("append", "index", i)
								}
							}
						}
					}

					if m.Table(); len(m.Meta["append"]) == 0 {
						m.Echo("%s", string(buf))
					}
				}
			}},
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)
}
