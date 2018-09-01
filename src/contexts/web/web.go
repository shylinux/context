package web // {{{
// }}}
import ( // {{{
	"bufio"
	"contexts"
	"github.com/gomarkdown/markdown"
	"path/filepath"
	"runtime"

	"encoding/json"
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"bytes"
	"mime/multipart"

	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// }}}

type MUX interface {
	Handle(string, http.Handler)
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
	Trans(*ctx.Message, string, func(*ctx.Message, *ctx.Context, string, ...string))
}

type WEB struct {
	*http.ServeMux
	*http.Server

	client *http.Client
	cookie map[string]*http.Cookie

	*ctx.Context
}

func (web *WEB) Merge(m *ctx.Message, uri string, arg ...string) string { // {{{
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

// }}}
func (web *WEB) Trans(m *ctx.Message, key string, hand func(*ctx.Message, *ctx.Context, string, ...string)) { // {{{
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		msg := m.Spawn()
		msg.Option("method", r.Method)
		msg.Option("path", r.URL.Path)
		msg.Option("referer", r.Header.Get("Referer"))

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
		case msg.Has("directory"):
			http.ServeFile(w, r, msg.Append("directory"))
		case msg.Has("redirect"):
			http.Redirect(w, r, msg.Append("redirect"), http.StatusFound)
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
}

// }}}
func (web *WEB) ServeHTTP(w http.ResponseWriter, r *http.Request) { // {{{
	m := web.Message().Log("info", "").Log("info", "%v %s %s", r.RemoteAddr, r.Method, r.URL)

	if m.Confs("logheaders") {
		for k, v := range r.Header {
			m.Log("info", "%s: %v", k, v)
		}
		m.Log("info", "")
	}

	if r.ParseForm(); len(r.PostForm) > 0 {
		for k, v := range r.PostForm {
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

// }}}

func (web *WEB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(WEB)
	s.Context = c
	s.cookie = web.cookie
	return s
}

// }}}
func (web *WEB) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	web.Caches["route"] = &ctx.Cache{Name: "请求路径", Value: "/" + web.Context.Name + "/", Help: "请求路径"}
	web.Caches["register"] = &ctx.Cache{Name: "已初始化(yes/no)", Value: "no", Help: "模块是否已初始化"}
	web.Caches["master"] = &ctx.Cache{Name: "服务入口(yes/no)", Value: "no", Help: "服务入口"}
	web.Caches["directory"] = &ctx.Cache{Name: "服务目录", Value: "usr", Help: "服务目录"}
	if len(arg) > 0 {
		m.Cap("directory", arg[0])
	}

	web.ServeMux = http.NewServeMux()
	if mux, ok := m.Target().Server.(MUX); ok {
		for k, x := range web.Commands {
			if k[0] == '/' {
				mux.Trans(m, k, x.Hand)
			}
		}
	}
	return web
}

// }}}
func (web *WEB) Start(m *ctx.Message, arg ...string) bool { // {{{
	if len(arg) > 0 {
		m.Cap("directory", arg[0])
	}

	m.Travel(func(m *ctx.Message, i int) bool {
		if h, ok := m.Target().Server.(http.Handler); ok && m.Cap("register") == "no" {
			m.Cap("register", "yes")
			m.Capi("nroute", 1)

			p, i := m.Target(), 0
			m.BackTrace(func(m *ctx.Message) bool {
				p = m.Target()
				if i++; i == 2 {
					return false
				}
				return true
			})

			if s, ok := p.Server.(MUX); ok {
				m.Log("info", "route %s -> %s", m.Cap("route"), m.Target().Name)
				s.Handle(m.Cap("route"), http.StripPrefix(path.Dir(m.Cap("route")), h))
			}

			if s, ok := m.Target().Server.(MUX); ok && m.Cap("directory") != "" {
				m.Log("info", "dir / -> [%s]", m.Cap("directory"))
				s.Handle("/", http.FileServer(http.Dir(m.Cap("directory"))))
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

	web.Caches["address"] = &ctx.Cache{Name: "服务地址", Value: ":9191", Help: "服务地址"}
	web.Caches["protocol"] = &ctx.Cache{Name: "服务协议", Value: "http", Help: "服务协议"}
	if len(arg) > 1 {
		m.Cap("address", arg[1])
	}
	if len(arg) > 2 {
		m.Cap("protocol", arg[2])
	}

	m.Cap("master", "yes")
	m.Cap("stream", m.Cap("address"))
	m.Log("info", "address [%s]", m.Cap("address"))
	m.Log("info", "protocol [%s]", m.Cap("protocol"))
	web.Server = &http.Server{Addr: m.Cap("address"), Handler: web}

	web.Configs["logheaders"] = &ctx.Config{Name: "日志输出报文头(yes/no)", Value: "no", Help: "日志输出报文头"}
	m.Capi("nserve", 1)

	// yac := m.Sess("tags", m.Sess("yac").Cmd("scan"))
	// yac.Cmd("train", "void", "void", "[\t ]+")
	// yac.Cmd("train", "other", "other", "[^\n]+")
	// yac.Cmd("train", "key", "key", "[A-Za-z_][A-Za-z_0-9]*")
	// yac.Cmd("train", "code", "struct", "struct", "key", "\\{")
	// yac.Cmd("train", "code", "struct", "\\}", "key", ";")
	// yac.Cmd("train", "code", "struct", "typedef", "struct", "key", "key", ";")
	// yac.Cmd("train", "code", "function", "key", "\\*", "key", "(", "other")
	// yac.Cmd("train", "code", "function", "key", "key", "(", "other")
	// yac.Cmd("train", "code", "variable", "struct", "key", "key", "other")
	// yac.Cmd("train", "code", "define", "#define", "key", "other")
	//
	if m.Cap("protocol") == "https" {
		web.Caches["cert"] = &ctx.Cache{Name: "服务证书", Value: m.Conf("cert"), Help: "服务证书"}
		web.Caches["key"] = &ctx.Cache{Name: "服务密钥", Value: m.Conf("key"), Help: "服务密钥"}
		m.Log("info", "cert [%s]", m.Cap("cert"))
		m.Log("info", "key [%s]", m.Cap("key"))

		web.Server.ListenAndServeTLS(m.Cap("cert"), m.Cap("key"))
	} else {
		web.Server.ListenAndServe()
	}

	return true
}

// }}}
func (web *WEB) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch web.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

// }}}

var Index = &ctx.Context{Name: "web", Help: "应用中心",
	Caches: map[string]*ctx.Cache{
		"nserve": &ctx.Cache{Name: "nserve", Value: "0", Help: "主机数量"},
		"nroute": &ctx.Cache{Name: "nroute", Value: "0", Help: "路由数量"},
	},
	Configs: map[string]*ctx.Config{
		"cmd":      &ctx.Config{Name: "cmd", Value: "tmux", Help: "路由数量"},
		"cert":     &ctx.Config{Name: "cert", Value: "etc/cert.pem", Help: "路由数量"},
		"key":      &ctx.Config{Name: "key", Value: "etc/key.pem", Help: "路由数量"},
		"wiki_dir": &ctx.Config{Name: "wiki_dir", Value: "usr/wiki", Help: "路由数量"},
		"wiki_list_show": &ctx.Config{Name: "wiki_list_show", Value: map[string]interface{}{
			"md": true,
		}, Help: "路由数量"},
		"which":        &ctx.Config{Name: "which", Value: "redis.note", Help: "路由数量"},
		"root_index":   &ctx.Config{Name: "root_index", Value: "/wiki/", Help: "路由数量"},
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
				//聊天服务
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "aaa", "detail": []interface{}{"lark"},
					"template": "detail", "title": "list_lark",
					// "option": map[string]interface{}{"auto_refresh": true},
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "aaa", "detail": []interface{}{"lark"},
					"template": "detail", "title": "send_lark",
					"option": map[string]interface{}{"ninput": 2},
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
				//格式转换
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "cli", "detail": []interface{}{"time"},
					"template": "detail", "title": "time",
					"option": map[string]interface{}{"refresh": true, "ninput": 1},
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "nfs", "detail": []interface{}{"pwd"},
					"template": "detail", "title": "pwd",
					"option": map[string]interface{}{"refresh": true},
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "nfs", "detail": []interface{}{"json"},
					"template": "detail", "title": "json",
					"option": map[string]interface{}{"ninput": 1},
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "cli", "command": "system",
					"argument": []interface{}{"tmux", "show-buffer"},
					"template": "result", "title": "buffer",
				},
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "web", "command": "/share",
					"argument": []interface{}{},
					"template": "share", "title": "share",
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
				map[string]interface{}{
					"from": "root", "to": []interface{}{},
					"module": "nfs", "command": "git",
					"argument": []interface{}{},
					"template": "result", "title": "git",
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
		"client": &ctx.Command{
			Name: "client address [output [editor]]",
			Help: "添加请求配置, address: 默认地址, output: 输出路径, editor: 编辑器",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if _, e := m.Target().Server.(*WEB); m.Assert(e) { // {{{
					if len(arg) == 0 {
						return
					}

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
				} // }}}
			}},
		"cookie": &ctx.Command{
			Name: "cookie [create]|[name [value]]",
			Help: "读写请求的Cookie, name: 变量名, value: 变量值",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if web, ok := m.Target().Server.(*WEB); m.Assert(ok) { // {{{
					switch len(arg) {
					case 0:
						for k, v := range web.cookie {
							m.Echo("%s: %v\n", k, v.Value)
						}
					case 1:
						if arg[0] == "create" {
							web.cookie = make(map[string]*http.Cookie)
							break
						}
						if v, ok := web.cookie[arg[0]]; ok {
							m.Echo("%s", v.Value)
						}
					default:
						if web.cookie == nil {
							web.cookie = make(map[string]*http.Cookie)
						}
						if v, ok := web.cookie[arg[0]]; ok {
							v.Value = arg[1]
						} else {
							web.cookie[arg[0]] = &http.Cookie{Name: arg[0], Value: arg[1]}
						}
					}
				} // }}}
			}},
		"get": &ctx.Command{
			Name: "get [method GET|POST] [file name filename] url arg...",
			Help: "访问服务, method: 请求方法, file: 发送文件, url: 请求地址, arg: 请求参数",
			Form: map[string]int{"method": 1, "file": 2, "type": 1, "body": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if web, ok := m.Target().Server.(*WEB); m.Assert(ok) { // {{{
					if web.client == nil {
						web.client = &http.Client{}
					}

					method := m.Confx("method")
					uri := web.Merge(m, arg[0], arg[1:]...)
					m.Log("info", "%s %s", method, uri)
					m.Echo("%s: %s\n", method, uri)

					var body io.Reader
					index := strings.Index(uri, "?")
					contenttype := ""

					switch method {
					case "POST":
						if m.Options("file") {
							file, e := os.Open(m.Meta["file"][1])
							m.Assert(e)
							defer file.Close()

							if m.Option("type") == "json" {
								contenttype = "application/json"
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

							contenttype = writer.FormDataContentType()
							body = buf
							writer.Close()
						} else if m.Option("type") == "json" {
							if m.Options("body") {
								data := []interface{}{}
								for _, v := range arg[1:] {
									data = append(data, v)
								}
								m.Log("body", "%v", fmt.Sprintf(m.Option("body"), data...))
								body = strings.NewReader(fmt.Sprintf(m.Option("body"), data...))
							} else {
								data := map[string]string{}
								for i := 1; i < len(arg)-1; i++ {
									data[arg[i]] = arg[i+1]
								}

								b, e := json.Marshal(data)
								m.Assert(e)
								body = bytes.NewReader(b)
							}

							contenttype = "application/json"
							if index > -1 {
								uri = uri[:index]
							}

						} else if index > 0 {
							contenttype = "application/x-www-form-urlencoded"
							body = strings.NewReader(uri[index+1:])
							uri = uri[:index]
						}
					}

					m.Log("info", "content-type: %s", contenttype)
					req, e := http.NewRequest(method, uri, body)
					m.Assert(e)

					if len(contenttype) > 0 {
						req.Header.Set("Content-Type", contenttype)
					}

					for _, v := range web.cookie {
						req.AddCookie(v)
					}

					res, e := web.client.Do(req)
					m.Assert(e)

					if web.cookie == nil {
						web.cookie = make(map[string]*http.Cookie)
					}
					for _, v := range res.Cookies() {
						web.cookie[v.Name] = v
					}

					for k, v := range res.Header {
						m.Log("info", "%s: %v", k, v)
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

					if res.Header.Get("Content-Type") == "application/json" {
						result := map[string]interface{}{}
						json.Unmarshal(buf, &result)
						for k, v := range result {
							switch value := v.(type) {
							case string:
								m.Append(k, value)
							case float64:
								m.Append(k, fmt.Sprintf("%d", int(value)))
							default:
								m.Put("append", k, value)
							}
						}
					}

					result := string(buf)
					m.Echo("%s", result)
					m.Append("response", result)
				} // }}}
			}},
		"post": &ctx.Command{Name: "post", Help: "访问服务",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				msg := m.Spawn().Cmd("get", "method", "POST", arg)
				m.Copy(msg, "result").Copy(msg, "append")
			}},
		"brow": &ctx.Command{Name: "brow url", Help: "浏览器网页", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			url := fmt.Sprintf("http://localhost:9094") //{{{
			if len(arg) > 0 {
				url = arg[0]
			}
			switch runtime.GOOS {
			case "windows":
				m.Find("cli").Cmd("system", "explorer", url)
			case "darwin":
				m.Find("cli").Cmd("system", "open", url)
			case "linux":
				m.Spawn().Cmd("open", url)
			}
			// }}}
		}},
		"serve": &ctx.Command{
			Name: "serve [directory [address [protocol]]]",
			Help: "启动服务, directory: 服务路径, address: 服务地址, protocol: 服务协议(https/http)",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				m.Set("detail", arg...).Target().Start(m)
			}},
		"route": &ctx.Command{
			Name: "route script|template|directory route content",
			Help: "添加响应, script: 脚本响应, template: 模板响应, directory: 目录响应, route: 请求路由, content: 响应内容",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if mux, ok := m.Target().Server.(MUX); m.Assert(ok) { // {{{
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
							mux.Trans(m, arg[1], func(m *ctx.Message, c *ctx.Context, key string, a ...string) {
								msg := m.Find("cli").Cmd("source", arg[2])
								m.Copy(msg, "result").Copy(msg, "append")
							})
						case "template":
							mux.Trans(m, arg[1], func(m *ctx.Message, c *ctx.Context, key string, a ...string) {
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
				} // }}}
			}},
		"upload": &ctx.Command{Name: "upload file", Help: "上传文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Spawn(m.Target()) // {{{
			msg.Cmd("get", "/upload", "method", "POST", "file", "file", arg[0])
			m.Copy(msg, "result")
			// }}}
		}},
		"/library/": &ctx.Command{Name: "/library", Help: "网页门户", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			r := m.Optionv("request").(*http.Request) // {{{
			w := m.Optionv("response").(http.ResponseWriter)
			dir := path.Join(m.Conf("library_dir"), m.Option("path"))
			if s, e := os.Stat(dir); e == nil && !s.IsDir() {
				http.ServeFile(w, r, dir)
				return
			}
			// }}}
		}},
		"/travel": &ctx.Command{Name: "/travel", Help: "文件上传", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			// r := m.Optionv("request").(*http.Request) // {{{
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
			// }}}
		}},
		"/index/": &ctx.Command{Name: "/index", Help: "网页门户", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			r := m.Optionv("request").(*http.Request) // {{{
			w := m.Optionv("response").(http.ResponseWriter)

			if login := m.Spawn().Cmd("/login"); login.Has("template") {
				m.Echo("no").Copy(login, "append")
				return
			}
			m.Option("username", m.Append("username"))

			//权限检查
			dir := m.Option("dir", path.Join(m.Cap("directory"), "local", m.Option("username"), m.Option("dir", strings.TrimPrefix(m.Option("path"), "/index"))))
			if check := m.Spawn(c).Cmd("/check", "command", "/index/", "dir", dir); !check.Results(0) {
				m.Copy(check, "append")
				return
			}

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
			// }}}
		}},
		"/create": &ctx.Command{Name: "/create", Help: "创建目录或文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if check := m.Spawn().Cmd("/share", "/upload", "dir", m.Option("dir")); !check.Results(0) { // {{{
				m.Copy(check, "append")
				return
			}

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
			// }}}
		}},
		"/share": &ctx.Command{Name: "/share arg...", Help: "资源共享", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if check := m.Spawn().Cmd("/check", "target", m.Option("module"), m.Optionv("share")); !check.Results(0) { // {{{
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
			// }}}
		}},
		"/check": &ctx.Command{Name: "/check arg...", Help: "权限检查, cache|config|command: 接口类型, name: 接口名称, args: 其它参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if login := m.Spawn().Cmd("/login"); login.Has("template") { // {{{
				m.Echo("no").Copy(login, "append")
				return
			}

			if msg := m.Spawn().Cmd("right", m.Append("username"), "check", arg); !msg.Results(0) {
				m.Echo("no").Append("message", "no right, please contact manager")
				return
			}

			m.Echo("ok")
			// }}}
		}},
		"/login": &ctx.Command{Name: "/login", Help: "用户登录", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if m.Options("sessid") { // {{{
				if aaa := m.Sess("aaa").Cmd("login", m.Option("sessid")); aaa.Results(0) {
					m.Append("redirect", m.Option("referer"))
					m.Append("username", aaa.Cap("username"))
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
			// }}}
		}},
		"/render": &ctx.Command{Name: "/render index", Help: "模板响应, main: 模板入口, tmpl: 附加模板", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			w := m.Optionv("response").(http.ResponseWriter) // {{{
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
			// }}}
		}},
		"/json": &ctx.Command{Name: "/json", Help: "json响应", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			w := m.Optionv("response").(http.ResponseWriter) // {{{

			meta := map[string][]string{}
			if len(m.Meta["result"]) > 0 {
				meta["result"] = m.Meta["result"]
			}
			if len(m.Meta["append"]) > 0 {
				meta["append"] = m.Meta["append"]
				for _, v := range m.Meta["append"] {
					meta[v] = m.Meta[v]
				}
			}

			if b, e := json.Marshal(meta); m.Assert(e) {
				w.Header().Set("Content-Type", "application/javascript")
				w.Write(b)
			}
			// }}}
		}},
		"/paste": &ctx.Command{Name: "/paste", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if login := m.Spawn().Cmd("/login"); login.Has("redirect") {
				m.Sess("cli").Cmd("system", "tmux", "set-buffer", "-b", "0", m.Option("content"))
			}
		}},
		"/blog": &ctx.Command{Name: "/blog", Help: "博客", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if m.Has("title") && m.Has("content") {
			}

			m.Echo("blog service")
		}},
		"/wiki_tags": &ctx.Command{Name: "/wiki_tags ", Help: "博客", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if len(arg) > 0 {
				m.Option("dir", arg[0])
			}

			yac := m.Find("yac.parse4", true)

			msg := m.Sess("nfs").Cmd("dir", path.Join(m.Conf("wiki_dir"), m.Option("dir")), "dir_name", "path")
			for i, v := range msg.Meta["filename"] {
				name := strings.TrimSpace(v)
				es := strings.Split(name, ".")
				switch es[len(es)-1] {
				case "pyc", "o", "gz", "tar":
					continue
				case "c":
				case "h":
				default:
					continue
				}

				f, e := os.Open(name)
				m.Assert(e)
				defer f.Close()
				m.Log("fuck", "%d/%d %s", i, len(msg.Meta["filename"]), v)

				bio := bufio.NewScanner(f)
				for line := 1; bio.Scan(); line++ {
					yac.Options("silent", true)
					l := yac.Cmd("parse", "code", "void", bio.Text())

					key := ""
					switch l.Result(1) {
					case "struct":
						switch l.Result(2) {
						case "struct", "}":
							key = l.Result(3)
						case "typedef":
							if l.Result(3) == "struct" {
								key = l.Result(5)
							}
						}
					case "function":
						switch l.Result(3) {
						case "*":
							key = l.Result(4)
						default:
							key = l.Result(3)
						}
					case "variable":
						switch l.Result(2) {
						case "struct":
							key = l.Result(4)
						}
					case "define":
						key = l.Result(3)
					}
					if key != "" {
						m.Confv("define", strings.Join([]string{key, "position", "-2"}, "."), map[string]interface{}{
							"file": strings.TrimPrefix(name, m.Confx("wiki_dir")),
							"line": line,
							"type": l.Result(1),
						})
					}

					yac.Meta = nil
				}
			}
			m.Log("fuck", "parse %s", time.Now().Format("2006-01-02 15:04:05"))
		}},
		"/wiki_body": &ctx.Command{Name: "/wiki_body", Help: "维基", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if ls, e := ioutil.ReadFile(path.Join(m.Conf("wiki_dir"), m.Confx("which"))); e == nil {
				pre := false
				es := strings.Split(m.Confx("which"), ".")
				if len(es) > 0 {
					switch es[len(es)-1] {
					case "md":
						ls = markdown.ToHTML(ls, nil, nil)
					default:
						pre = true
					}
				}

				if pre {
					m.Option("nline", bytes.Count(ls, []byte("\n")))
					m.Option("nbyte", len(ls))
					m.Add("append", "code", string(ls))
					m.Add("append", "body", "")
				} else {
					m.Add("append", "body", string(ls))
					m.Add("append", "code", "")
				}
				return
			}

			if m.Options("query") {
				if v, ok := m.Confv("define", m.Option("query")).(map[string]interface{}); ok {
					for _, val := range v["position"].([]interface{}) {
						value := val.(map[string]interface{})
						m.Add("append", "name", fmt.Sprintf("%v#hash_%v", value["file"], value["line"]))
					}
					return
				}
				msg := m.Sess("nfs").Cmd("dir", path.Join(m.Conf("wiki_dir"), m.Option("dir")), "dir_name", "path")
				for _, v := range msg.Meta["filename"] {
					name := strings.TrimPrefix(strings.TrimSpace(v), m.Conf("wiki_dir"))
					es := strings.Split(name, ".")
					switch es[len(es)-1] {
					case "pyc", "o", "gz", "tar":
						continue
					}
					if strings.Contains(name, m.Option("query")) {
						m.Add("append", "name", name)
					}
				}
				return
			}

			msg := m.Spawn().Cmd("/wiki_list")
			m.Copy(msg, "append").Copy(msg, "option")
		}},
		"/wiki_list": &ctx.Command{Name: "/wiki_list", Help: "维基", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			ls, e := ioutil.ReadDir(path.Join(m.Conf("wiki_dir"), m.Option("which")))
			m.Option("dir", m.Option("which"))
			if e != nil {
				dir, _ := path.Split(m.Option("which"))
				m.Option("dir", dir)
				ls, e = ioutil.ReadDir(path.Join(m.Conf("wiki_dir"), dir))
			}

			parent, _ := path.Split(strings.TrimSuffix(m.Option("dir"), "/"))
			m.Option("parent", parent)
			for _, l := range ls {
				if l.Name()[0] == '.' {
					continue
				}
				if !l.IsDir() {
					es := strings.Split(l.Name(), ".")
					if len(es) > 0 {
						if show, ok := m.Confv("wiki_list_show", es[len(es)-1]).(bool); !ok || !show {
							continue
						}
					}
				}

				m.Add("append", "name", l.Name())
				m.Add("append", "time", l.ModTime().Format("2006-01-02 15:04:05"))
				if l.IsDir() {
					m.Add("append", "pend", "/")
				} else {
					m.Add("append", "pend", "")
				}
				m.Option("time_layout", "2006-01-02 15:04:05")
				m.Sort("time", "time_r")
			}
		}},
		"/wiki/": &ctx.Command{Name: "/wiki", Help: "维基", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Option("which", strings.TrimPrefix(key, "/wiki/"))
			if f, e := os.Stat(path.Join(m.Conf("wiki_dir"), m.Option("which"))); e == nil && !f.IsDir() && (strings.HasSuffix(m.Option("which"), ".json") || strings.HasSuffix(m.Option("which"), ".js") || strings.HasSuffix(m.Option("which"), ".css")) {
				m.Append("directory", path.Join(m.Conf("wiki_dir"), m.Option("which")))
				return
			}

			m.Append("template", "wiki")
		}},
		"/wx/": &ctx.Command{Name: "/wx/", Help: "微信", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if !m.Sess("aaa").Cmd("wx").Results(0) {
				return
			}
			if m.Has("echostr") {
				m.Echo(m.Option("echostr"))
				return
			}
			r := m.Optionv("request").(*http.Request)

			switch r.Header.Get("Content-Type") {
			case "text/xml":
				type Article struct {
					XMLName     xml.Name `xml:"item"`
					PicUrl      string
					Title       string
					Description string
					Url         string
				}
				type WXMsg struct {
					XMLName      xml.Name `xml:"xml"`
					ToUserName   string
					FromUserName string
					CreateTime   int32
					MsgId        int64
					MsgType      string

					Event    string
					EventKey string

					Content string

					Format      string
					Recognition string

					PicUrl  string
					MediaId string

					Location_X float64
					Location_Y float64
					Scale      int64
					Label      string

					ArticleCount int
					Articles     struct {
						XMLName  xml.Name `xml:"Articles"`
						Articles []*Article
					}
				}

				var data WXMsg

				b, e := ioutil.ReadAll(r.Body)
				m.Log("fuck", "b: %v", string(b))
				e = xml.Unmarshal(b, &data)
				m.Log("fuck", "b: %#v", data)

				// de := xml.NewDecoder(r.Body)
				// e := de.Decode(&data)
				m.Assert(e)

				var echo WXMsg
				echo.FromUserName = data.ToUserName
				echo.ToUserName = data.FromUserName
				echo.CreateTime = data.CreateTime

				fs, e := ioutil.ReadDir("usr/wiki")
				m.Assert(e)
				msg := m.Spawn()
				for _, f := range fs {
					if !strings.HasSuffix(f.Name(), ".md") {
						continue
					}
					msg.Add("append", "name", f.Name())
					msg.Add("append", "title", strings.TrimSuffix(f.Name(), ".md")+"源码解析")
					msg.Add("append", "time", f.ModTime().Format("01/02 15:03"))
				}
				msg.Option("time_layout", "01/02 15:03")
				msg.Sort("time", "time_r")

				articles := []*Article{}
				articles = append(articles, &Article{PicUrl: "http://mmbiz.qpic.cn/mmbiz_jpg/sCJZHmp0V0doWEFBe6gS2HjgB0abiaK7H5WjkXGTvAI0CkCFrVJDEBBbJX8Kz0VegZ54ZoCo4We0sKJUOTuf1Tw/0",
					Title: "wiki首页", Description: "技术文章", Url: "https://shylinux.com/wiki/"})
				for i, v := range msg.Meta["title"] {
					if i > 6 {
						continue
					}

					articles = append(articles, &Article{PicUrl: "http://mmbiz.qpic.cn/mmbiz_jpg/sCJZHmp0V0doWEFBe6gS2HjgB0abiaK7H5WjkXGTvAI0CkCFrVJDEBBbJX8Kz0VegZ54ZoCo4We0sKJUOTuf1Tw/0",
						Title: msg.Meta["time"][i] + " " + v, Description: "技术文章", Url: "https://shylinux.com/wiki/" + msg.Meta["name"][i]})
				}

				switch data.MsgType {
				case "event":
					echo.MsgType = "news"
					echo.Articles.Articles = articles
					echo.ArticleCount = len(echo.Articles.Articles)
				case "text":
					echo.MsgType = "news"
					echo.Articles.Articles = articles
					echo.ArticleCount = len(echo.Articles.Articles)
				case "voice":
					echo.MsgType = "text"
					echo.Content = "你好"
				case "image":
					echo.MsgType = "text"
					echo.Content = "你好"
				case "location":
					echo.MsgType = "text"
					echo.Content = "你好"
				}

				b, e = xml.Marshal(echo)
				m.Echo(string(b))
			}
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
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)
}
