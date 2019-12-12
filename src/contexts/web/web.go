package web

import (
	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
	"time"

	"contexts/ctx"
	"toolkit"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type MUX interface {
	Handle(string, http.Handler)
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
	HandleCmd(*ctx.Message, string, *ctx.Command)
	ServeHTTP(http.ResponseWriter, *http.Request)
}
type WEB struct {
	*http.Client

	*http.Server
	*http.ServeMux
	*template.Template

	*ctx.Context
}

func Cookie(msg *ctx.Message, w http.ResponseWriter, r *http.Request) {
	expire := time.Now().Add(kit.Duration(msg.Conf("login", "expire")))
	msg.Log("info", "expire %v", expire)
	http.SetCookie(w, &http.Cookie{Name: "sessid",
		Value: msg.Cmdx("aaa.user", "session", "select"), Path: "/", Expires: expire})
	return
}
func proxy(m *ctx.Message, url string) string {
	if strings.HasPrefix(url, "//") {
		return "proxy/https:" + url
	}

	return "proxy/" + url
}
func merge(m *ctx.Message, uri string, arg ...string) string {
	add, e := url.Parse(uri)
	m.Assert(e)

	args := []interface{}{}
	for i := 0; i < strings.Count(add.RawQuery, "%s") && len(arg) > 0; i++ {
		args, arg = append(args, arg[0]), arg[1:]
	}
	add.RawQuery = fmt.Sprintf(add.RawQuery, args...)

	query := add.Query()
	for i := 0; i < len(arg)-1; i += 2 {
		value := m.Parse(arg[i+1])

		if value == "" {
			query.Del(arg[i])
		} else {
			// query.Set(arg[i], value)
			query.Add(arg[i], value)
		}
	}
	add.RawQuery = query.Encode()
	return add.String()
}
func Merge(m *ctx.Message, client map[string]interface{}, uri string, arg ...string) string {
	add, e := url.Parse(uri)
	m.Assert(e)

	add.Scheme = kit.Select(kit.Format(client["protocol"]), add.Scheme)
	add.Host = kit.Select(kit.Format(client["hostname"]), add.Host)

	if add.Path == "" {
		add.Path = path.Join(kit.Format(client["path"]), kit.Format(client["file"]))
	} else if !path.IsAbs(add.Path) {
		add.Path = path.Join(kit.Format(client["path"]), add.Path)
		if strings.HasSuffix(uri, "/") {
			add.Path += "/"
		}
	}

	add.RawQuery = kit.Select(kit.Format(client["query"]), add.RawQuery)
	return merge(m, add.String(), arg...)
}

func (web *WEB) Login(msg *ctx.Message, w http.ResponseWriter, r *http.Request) bool {
	if !msg.Has("username") || !msg.Options("username") {
		msg.Option("username", "")
	}

	// 用户登录
	if msg.Options("username") && msg.Options("password") {
		if msg.Cmds("aaa.auth", "username", msg.Option("username"), "password", msg.Option("password")) {
			msg.Log("info", "login: %s", msg.Option("username"))
			if Cookie(msg, w, r); msg.Options("relay") {
				if role := msg.Cmdx("aaa.relay", "check", msg.Option("relay"), "userrole"); role != "" {
					msg.Cmd("aaa.role", role, "user", msg.Option("username"))
					msg.Log("info", "relay: %s", role)
				}
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
		return false

	} else if msg.Options("relay") {
		msg.Short("relay")

		relay := msg.Cmd("aaa.relay", "check", msg.Option("relay"))
		if relay.Appendi("count") == 0 {
			msg.Err("共享失效")
			return false
		}
		if relay.Appends("username") {
			name := msg.Cmdx("ssh._route", msg.Conf("runtime", "work.route"), "_check", "work", "create", relay.Append("username"), msg.Conf("runtime", "node.route"))
			msg.Log("info", "login: %s", msg.Option("username", name))
			Cookie(msg, w, r)
		}
		if role := relay.Append("userrole"); role != "" {
			msg.Cmd("aaa.role", role, "user", msg.Option("username"))
			msg.Log("info", "relay: %s", role)
		}
		if relay.Appends("url") {
			msg.Append("redirect", relay.Append("url"))
			return false
		}

		if relay.Appends("form") {
			form := kit.UnMarshalm(relay.Append("form"))
			for k, v := range form {
				msg.Log("info", "form %s:%s", k, msg.Option(k, kit.Format(v)))
			}
		}
	}

	// 用户访问
	if msg.Log("info", "sessid: %s", msg.Option("sessid")); msg.Options("sessid") {
		msg.Log("info", "username: %s", msg.Option("username", msg.Cmd("aaa.sess", "user").Append("meta")))
		msg.Log("info", "userrole: %v", msg.Option("userrole", msg.Cmd("aaa.user", "role").Append("meta")))
	}

	// 本地用户
	if !msg.Options("username") && kit.IsLocalIP(msg.Option("remote_ip")) && msg.Confs("web.login", "local") && !strings.HasPrefix(msg.Option("agent"), "curl") {
		msg.Cmd("aaa.role", "root", "user", msg.Cmdx("ssh.work", "create"))
		msg.Log("info", "%s: %s", msg.Option("remote_ip"), msg.Option("username", msg.Conf("runtime", "work.name")))
		Cookie(msg, w, r)
	}

	return true
}
func (web *WEB) HandleCmd(m *ctx.Message, key string, cmd *ctx.Command) {
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		m.TryCatch(m.Spawn(m.Conf("serve", "autofree")), true, func(msg *ctx.Message) {
			defer func() {
				msg.Log("time", "serve: %v", msg.Format("cost"))
			}()
			msg.Option("agent", r.Header.Get("User-Agent"))
			msg.Option("remote_addr", r.RemoteAddr)
			msg.Option("remote_ip", r.Header.Get("remote_ip"))
			msg.Option("index_url", r.Header.Get("index_url"))
			msg.Option("index_path", r.Header.Get("index_path"))
			msg.Option("referer", r.Header.Get("Referer"))
			msg.Option("accept", r.Header.Get("Accept"))
			msg.Option("method", r.Method)
			msg.Option("path", r.URL.Path)
			msg.Option("sessid", "")

			// 请求环境
			for _, v := range r.Cookies() {
				if v.Value != "" {
					msg.Option(v.Name, v.Value)
				}
			}

			// 请求参数
			r.ParseMultipartForm(int64(msg.Confi("serve", "form_size")))
			if r.ParseForm(); len(r.PostForm) > 0 {
				for k, v := range r.PostForm {
					msg.Log("info", "%s: %v", k, v)
				}
				msg.Log("info", "")
			}
			for k, v := range r.Form {
				msg.Add("option", k, v)
			}

			// 请求数据
			switch r.Header.Get("Content-Type") {
			case "application/json":
				var data interface{}
				if e := json.NewDecoder(r.Body).Decode(&data); e != nil {
					m.Log("warn", "%v", e)
				}
				msg.Optionv("content_data", data)
				m.Log("info", "%v", kit.Formats(data))

				switch d := data.(type) {
				case map[string]interface{}:
					for k, v := range d {
						for _, v := range kit.Trans(v) {
							msg.Add("option", k, v)
						}
					}
				}
			}

			msg.Short("river")

			// 用户登录
			if msg.Put("option", "request", r).Put("option", "response", w).Sess("web", msg); web.Login(msg, w, r) {
				msg.Log("cmd", "%s [] %v", key, msg.Meta["option"])
				cmd.Hand(msg, msg.Target(), msg.Option("path"))
			}

			// 返回响应
			switch {
			case msg.Has("redirect"):
				http.Redirect(w, r, msg.Append("redirect"), http.StatusTemporaryRedirect)

			case msg.Has("directory"):
				http.ServeFile(w, r, msg.Append("directory"))

			case msg.Has("qrcode"):
				if qr, e := qrcode.New(msg.Append("qrcode"), qrcode.Medium); m.Assert(e) {
					w.Header().Set("Content-Type", "image/png")
					m.Assert(qr.Write(256, w))
				}

			case msg.Has("append"):
				meta := map[string]interface{}{}
				if len(msg.Meta["result"]) > 0 {
					meta["result"] = msg.Meta["result"]
				}
				if len(msg.Meta["append"]) > 0 {
					meta["append"] = msg.Meta["append"]
					for _, v := range msg.Meta["append"] {
						if _, ok := msg.Data[v]; ok {
							meta[v] = msg.Data[v]
						} else if _, ok := msg.Meta[v]; ok {
							meta[v] = msg.Meta[v]
						}
					}
				}

				if b, e := json.Marshal(meta); msg.Assert(e) {
					w.Header().Set("Content-Type", "application/javascript")
					w.Write(b)
				}
			default:
				for _, v := range msg.Meta["result"] {
					w.Write([]byte(v))
				}
			}
		})
	})
}
func (web *WEB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := web.Message()

	index := r.Header.Get("index_module") == ""
	if index {
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			r.Header.Set("remote_ip", ip)
		} else if ip := r.Header.Get("X-Real-Ip"); ip != "" {
			r.Header.Set("remote_ip", ip)
		} else if strings.HasPrefix(r.RemoteAddr, "[") {
			r.Header.Set("remote_ip", strings.Split(r.RemoteAddr, "]")[0][1:])
		} else {
			r.Header.Set("remote_ip", strings.Split(r.RemoteAddr, ":")[0])
		}

		m.Log("info", "").Log("info", "%v %s %s", r.Header.Get("remote_ip"), r.Method, r.URL)
		r.Header.Set("index_module", m.Cap("module"))
		r.Header.Set("index_url", r.URL.String())
		r.Header.Set("index_path", r.URL.Path)
		if r.URL.Path == "/" && m.Confs("serve", "index") {
			r.URL.Path = m.Conf("serve", "index")
		}
	}

	if index && m.Confs("serve", "logheaders") {
		for k, v := range r.Header {
			m.Log("info", "%s: %v", k, v)
		}
		m.Log("info", "")
	}

	if r.URL.Path == "/" && m.Confs("route", "index") {
		r.URL.Path = m.Conf("route", "index")
	}
	web.ServeMux.ServeHTTP(w, r)

	if index && m.Confs("serve", "logheaders") {
		for k, v := range w.Header() {
			m.Log("info", "%s: %v", k, v)
		}
		m.Log("info", "")
	}
}

func (web *WEB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	return &WEB{Context: c}
}
func (web *WEB) Begin(m *ctx.Message, arg ...string) ctx.Server {
	web.Caches["master"] = &ctx.Cache{Name: "master(yes/no)", Value: "no", Help: "服务入口"}
	web.Caches["register"] = &ctx.Cache{Name: "register(yes/no)", Value: "no", Help: "是否已初始化"}
	web.Caches["route"] = &ctx.Cache{Name: "route", Value: "/" + web.Context.Name + "/", Help: "模块路由"}

	web.ServeMux = http.NewServeMux()
	web.Template = template.New("render").Funcs(ctx.CGI)
	web.Template.ParseGlob(path.Join(m.Conf("route", "template_dir"), "/*.tmpl"))
	web.Template.ParseGlob(path.Join(m.Conf("route", "template_dir"), m.Cap("route"), "/*.tmpl"))
	return web
}
func (web *WEB) Start(m *ctx.Message, arg ...string) bool {
	web.Caches["directory"] = &ctx.Cache{Name: "directory", Value: kit.Select(m.Conf("serve", "directory"), arg, 0), Help: "服务目录"}
	web.Caches["protocol"] = &ctx.Cache{Name: "protocol", Value: kit.Select(m.Conf("serve", "protocol"), arg, 2), Help: "服务协议"}
	web.Caches["address"] = &ctx.Cache{Name: "address", Value: kit.Select(m.Conf("runtime", "boot.web_port"), arg, 1), Help: "服务地址"}
	m.Log("info", "%d %s %s://%s", m.Capi("nserve", 1), m.Cap("directory"), m.Cap("protocol"), m.Cap("stream", m.Cap("address")))

	render := m.Target().Commands["/render"]
	proxy := m.Target().Commands["/proxy/"]

	m.Target().Travel(m, func(m *ctx.Message, i int) bool {
		if h, ok := m.Target().Server.(MUX); ok && m.Cap("register") == "no" {
			m.Cap("register", "yes")

			// 路由级联
			p := m.Target().Context()
			if s, ok := p.Server.(MUX); ok {
				m.Log("info", "route: /%s <- %s", p.Name, m.Cap("route"))
				s.Handle(m.Cap("route"), http.StripPrefix(path.Dir(m.Cap("route")), h))
			}

			// 模板文件
			if m.Target().Commands["/render"] == nil {
				m.Target().Commands["/render"] = render
			}
			if m.Target().Commands["/proxy/"] == nil {
				m.Target().Commands["/proxy/"] = proxy
			}

			// 动态文件
			msg := m.Target().Message()
			for k, x := range m.Target().Commands {
				if k[0] == '/' {
					m.Log("info", "%d route: %s", m.Capi("nroute", 1), k)
					h.HandleCmd(msg, k, x)
				}
			}

			// 静态文件
			if m.Cap("directory") != "" {
				m.Log("info", "route: %sstatic/ <- [%s]\n", m.Cap("route"), m.Cap("directory"))
				h.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(m.Cap("directory")))))
			}
		}
		return false
	})

	// 启动服务
	m.Log("info", "web: %s", m.Cap("address"))
	web.Server = &http.Server{Addr: m.Cap("address"), Handler: web}
	if m.Caps("master", true); m.Cap("protocol") == "https" {
		web.Server.ListenAndServeTLS(m.Conf("runtime", "node.cert"), m.Conf("runtime", "node.key"))
	} else {
		web.Server.ListenAndServe()
	}
	return true
}
func (web *WEB) Close(m *ctx.Message, arg ...string) bool {
	return true
}

var Index = &ctx.Context{Name: "web", Help: "应用中心",
	Caches: map[string]*ctx.Cache{
		"nserve": &ctx.Cache{Name: "nserve", Value: "0", Help: "主机数量"},
		"nroute": &ctx.Cache{Name: "nroute", Value: "0", Help: "路由数量"},
	},
	Configs: map[string]*ctx.Config{
		"spide": &ctx.Config{Name: "spide", Value: map[string]interface{}{
			"": map[string]interface{}{
				"client": map[string]interface{}{},
				"header": map[string]interface{}{},
				"cookie": map[string]interface{}{},
			},
		}, Help: "爬虫配置"},
		"serve": &ctx.Config{Name: "serve", Value: map[string]interface{}{
			"autofree":   false,
			"logheaders": false,
			"form_size":  "102400",
			"directory":  "usr",
			"protocol":   "http",
			"cert":       "etc/cert.pem",
			"key":        "etc/key.pem",
			"site":       "",
			// "index":      "/chat/",
			"index": "/static/volcanos/",
			"open":  []interface{}{},
		}, Help: "服务配置"},
		"route": &ctx.Config{Name: "route", Value: map[string]interface{}{
			"index":          "/render",
			"template_dir":   "usr/template",
			"template_debug": false,
			"componet_index": "index",
			"toolkit_view": map[string]interface{}{
				"top": 96, "left": 472, "width": 600, "height": 300,
			},
		}, Help: "功能配置"},
		"login": &ctx.Config{Name: "login", Value: map[string]interface{}{
			"expire":    "240h",
			"local":     true,
			"check":     true,
			"sess_void": false,
			"cas_uuid":  "email",
		}, Help: "认证配置"},
		"componet": &ctx.Config{Name: "componet", Value: map[string]interface{}{
			"login": []interface{}{},
			"index": []interface{}{
				map[string]interface{}{"name": "head", "template": "head"},
				map[string]interface{}{"name": "clipbaord", "template": "clipboard"},
				map[string]interface{}{"name": "tail", "template": "tail"},
			},
		}, Help: "组件列表"},
		"upload": &ctx.Config{Name: "upload", Value: map[string]interface{}{
			"path": "var/file",
		}, Help: "上件文件"},
		"toolkit": &ctx.Config{Name: "toolkit", Value: map[string]interface{}{
			"time": map[string]interface{}{"cmd": "time"},
		}, Help: "工具列表"},
		"wss": &ctx.Config{Name: "wss", Value: map[string]interface{}{}, Help: ""},
	},
	Commands: map[string]*ctx.Command{
		"_init": &ctx.Command{Name: "_init", Help: "post请求", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmd("web.spide", "cas", "client", "new", m.Conf("runtime", "boot.ctx_cas"))
			m.Cmd("web.spide", "dev", "client", "new", kit.Select(m.Conf("runtime", "boot.ctx_dev"), m.Conf("runtime", "boot.ctx_box")))
			return
		}},
		"spide": &ctx.Command{Name: "spide [which [client|header|cookie [name|new [value]]]]", Help: "爬虫配置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch len(arg) {
			case 0:
				m.Confm("spide", func(key string, value map[string]interface{}) {
					m.Push("key", key)
					m.Push("protocol", kit.Chains(value, "client.protocol"))
					m.Push("hostname", kit.Chains(value, "client.hostname"))
					m.Push("path", kit.Chains(value, "client.path"))
				})
				m.Sort("key").Table()
			case 1:
				m.Cmdy("ctx.config", "spide", arg[0])
			case 2:
				m.Cmdy("ctx.config", "spide", strings.Join(arg[:2], "."))
			default:
				switch arg[1] {
				case "client":
					if len(arg) == 3 {
						m.Cmdy("ctx.config", "spide", strings.Join(arg[:3], "."))
						break
					}

					if arg[2] == "new" {
						if uri, e := url.Parse(arg[3]); e == nil && arg[3] != "" {
							dir, file := path.Split(uri.EscapedPath())
							m.Confv("spide", arg[0], map[string]interface{}{
								"cookie": map[string]interface{}{},
								"header": map[string]interface{}{},
								"client": map[string]interface{}{
									"logheaders": false,
									"timeout":    "100s",
									"method":     "GET",
									"protocol":   uri.Scheme,
									"hostname":   uri.Host,
									"path":       dir,
									"file":       file,
									"query":      uri.RawQuery,
									"url":        arg[3],
								},
							})
						}
						break
					}

					m.Cmd("ctx.config", "spide", strings.Join(arg[:3], "."), arg[3])
				case "merge":
					m.Echo(Merge(m, m.Confm("spide", []string{arg[0], "client"}), arg[2], arg[3:]...))

				case "cookie", "header":
					if len(arg) > 3 {
						m.Cmd("ctx.config", "spide", strings.Join(arg[:3], "."), arg[3])
					}
					m.Cmdy("ctx.config", "spide", strings.Join(arg[:3], "."))

				default:
					m.Cmd("ctx.config", "spide", strings.Join(arg[:2], "."), arg[2])
				}
			}
			return
		}},
		"post": &ctx.Command{Name: "post args...", Help: "post请求", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy("web.get", "method", "POST", arg)
			return
		}},
		"get": &ctx.Command{Name: "get [which] name [method GET|POST] url arg...", Help: "访问服务, method: 请求方法, url: 请求地址, arg: 请求参数",
			Form: map[string]int{
				"which": 1, "method": 1, "args": 1, "headers": 2,
				"content_type": 1, "content_data": 1, "body": 1, "file": 2,
				"parse": 1, "temp": -1, "temp_expire": 1, "save": 1, "saves": 1,
			}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				// 查看配置
				if len(arg) == 0 {
					m.Cmdy("web.spide")
					return
				}
				if len(arg) == 1 {
					m.Cmdy("web.spide", arg[0])
					return
				}

				web, ok := m.Target().Server.(*WEB)
				m.Assert(ok)

				// 查找配置
				which, client := m.Option("which"), m.Confm("spide", []string{m.Option("which"), "client"})
				if c := m.Confm("spide", []string{arg[0], "client"}); c != nil {
					which, client, arg = arg[0], c, arg[1:]
				}

				method := kit.Select(kit.Format(client["method"]), m.Option("method"))
				uri := Merge(m, client, arg[0], arg[1:]...)
				uri_arg := ""

				body, ok := m.Optionv("body").(io.Reader)
				if method == "POST" && !ok {
					uuu, e := url.Parse(uri)
					m.Assert(e)

					if m.Has("content_data") { // POST file
						body = bytes.NewReader([]byte(m.Option("content_data")))

					} else if m.Has("file") { // POST file
						writer := multipart.NewWriter(&bytes.Buffer{})
						defer writer.Close()

						for k, v := range uuu.Query() {
							for _, v := range v {
								writer.WriteField(k, v)
							}
						}

						if file, e := os.Open(m.Cmdx("nfs.path", m.Meta["file"][1])); m.Assert(e) {
							defer file.Close()

							if part, e := writer.CreateFormFile(m.Option("file"), filepath.Base(m.Meta["file"][1])); m.Assert(e) {
								io.Copy(part, file)
								m.Option("content_type", writer.FormDataContentType())
							}
						}

					} else if index := strings.Index(uri, "?"); index > 0 {
						switch m.Option("content_type") {
						case "application/json": // POST json
							var data interface{}
							for k, v := range uuu.Query() {
								if len(v) == 1 {
									if i, e := strconv.Atoi(v[0]); e == nil {
										data = kit.Chain(data, k, i)
									} else {
										data = kit.Chain(data, k, v[0])
									}
								} else {
									for i, val := range v {
										if i, e := strconv.Atoi(v[i]); e == nil {
											data = kit.Chain(data, []string{k, "-2"}, i)
										} else {
											data = kit.Chain(data, []string{k, "-2"}, val)
										}
									}
								}
							}

							if b, e := json.Marshal(data); m.Assert(e) {
								m.Log("info", "json %v", string(b))

								if body = bytes.NewReader(b); m.Has("args") {
									uri = uri[:index] + "?" + m.Option("args")
									index = len(uri)
								}
							}

						default: // POST form
							m.Log("info", "body %v", string(uri[index+1:]))
							body = strings.NewReader(uri[index+1:])
							m.Option("content_type", "application/x-www-form-urlencoded")
							m.Option("content_length", len(uri[index+1:]))
						}
						uri, uri_arg = uri[:index], "?"+uuu.RawQuery
					}
				}

				// 查找缓存
				if m.Options("temp_expire") {
					if h := m.Cmdx("mdb.temp", "check", "url", uri+uri_arg); h != "" {
						m.Cmdy("mdb.temp", h, "data", "data", m.Meta["temp"])
						return
					}
				}

				// 构造请求
				req, e := http.NewRequest(method, uri, body)
				m.Assert(e)
				m.Log("info", "%s %s", req.Method, req.URL)

				m.Confm("spide", []string{which, "header"}, func(key string, value string) {
					if key != "" {
						req.Header.Set(key, value)
						m.Log("info", "header %v %v", key, value)
					}
				})
				for i := 0; i < len(m.Meta["headers"]); i += 2 {
					req.Header.Set(m.Meta["headers"][i], m.Meta["headers"][i+1])
				}
				if m.Options("content_type") {
					req.Header.Set("Content-Type", m.Option("content_type"))
				}
				if m.Options("content_length") {
					req.Header.Set("Content-Length", m.Option("content_length"))
				}
				// 请求头
				if kit.Right(client["logheaders"]) {
					for k, v := range req.Header {
						m.Log("info", "%s: %s", k, v)
					}
					m.Log("info", "")
				}

				// 请求cookie
				kit.Structm(m.Magic("user", []string{"cookie", which}), func(key string, value string) {
					req.AddCookie(&http.Cookie{Name: key, Value: value})
					m.Log("info", "set-cookie %s: %v", key, value)
				})

				// 发送请求
				if web.Client == nil {
					web.Client = &http.Client{Timeout: kit.Duration(kit.Format(client["timeout"]))}
				}
				res, e := web.Client.Do(req)
				if e != nil {
					m.Log("warn", "%v", e)
					m.Echo("%v", e)
					return e
				}

				// 响应结果
				if res.StatusCode != http.StatusOK {
					m.Log("warn", "%d: %s\n", res.StatusCode, res.Status)
					m.Echo("%d: %s\n", res.StatusCode, res.Status)
				}

				// 保存cookie
				for _, v := range res.Cookies() {
					if m.Log("info", "get-cookie %s: %v", v.Name, v.Value); v.Value != "" {
						m.Magic("user", []string{"cookie", which, v.Name}, v.Value)
					}
				}

				// 响应头
				if kit.Right(client["logheaders"]) {
					for k, v := range res.Header {
						m.Log("info", "%s: %v", k, v)
					}
				}

				// 保存响应
				if res.StatusCode == http.StatusOK && m.Options("save") {
					if f, p, e := kit.Create(m.Option("save")); m.Assert(e) {
						defer f.Close()

						if n, e := io.Copy(f, res.Body); m.Assert(e) {
							m.Log("info", "save %d %s", n, p)
							m.Echo(p)
						}
					}
					return
				}

				// 解析响应
				var result interface{}
				ct := res.Header.Get("Content-Type")
				parse := kit.Select(kit.Format(client["parse"]), m.Option("parse"))
				m.Log("info", "parse: %s content: %s", parse, ct)

				switch {
				case parse == "json" || strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "application/javascript"):
					// 解析数据
					if json.NewDecoder(res.Body).Decode(&result); m.Options("temp_expire") {
						if m.Log("info", "res: %v", kit.Format(result)); !m.Has("temp") {
							m.Option("temp", "")
						}
						// m.Put("option", "data", result).Cmdy("mdb.temp", "url", uri+uri_arg, "data", "data", m.Meta["temp"])
						m.Put("option", "data", result).Cmdy("mdb.temp", "url", uri+uri_arg, "data", "data", m.Meta["temp"])
						break
					} else if result != nil {
						if b, e := json.MarshalIndent(result, "", "  "); m.Assert(e) {
							m.Echo(string(b))
						}
						break
					}
					fallthrough
				default:
					// 输出数据
					if buf, e := ioutil.ReadAll(res.Body); m.Assert(e) {
						m.Echo(string(buf))
					}
				}
				if m.Options("saves") {
					f, p, e := kit.Create(m.Option("saves"))
					m.Assert(e)
					defer f.Close()
					for _, v := range m.Meta["result"] {
						f.WriteString(v)
					}
					m.Set("result").Echo(p)
				}
				return
			}},

		"serve": &ctx.Command{Name: "serve [directory [address [protocol [cert [key]]]]", Help: "启动服务, directory: 服务路径, address: 服务地址, protocol: 服务协议(https/http), cert: 服务证书, key: 服务密钥", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Set("detail", arg).Target().Start(m)
			return
		}},
		"route": &ctx.Command{Name: "route index content [help]", Help: "添加路由响应, index: 路由, context: 响应, help: 说明", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if mux, ok := m.Target().Server.(MUX); m.Assert(ok) {
				switch len(arg) {
				case 0:
					for k, v := range m.Target().Commands {
						if k[0] == '/' {
							m.Add("append", "route", k)
							m.Add("append", "name", v.Name)
						}
					}
					m.Sort("route").Table()
				case 1:
					for k, v := range m.Target().Commands {
						if k == arg[0] {
							m.Echo("%s: %s\n%s", k, v.Name, v.Help)
						}
					}
				default:
					help := kit.Select("dynamic route", arg, 2)
					hand := func(m *ctx.Message, c *ctx.Context, key string, a ...string) (e error) {
						w := m.Optionv("response").(http.ResponseWriter)
						template.Must(template.New("temp").Parse(arg[1])).Execute(w, m)
						return
					}

					if s, e := os.Stat(arg[1]); e == nil {
						if s.IsDir() {
							mux.Handle(arg[0]+"/", http.StripPrefix(arg[0], http.FileServer(http.Dir(arg[1]))))
						} else if strings.HasSuffix(arg[1], ".shy") {
							hand = func(m *ctx.Message, c *ctx.Context, key string, a ...string) (e error) {
								m.Cmdy("cli.source", arg[1])
								return
							}
						} else {
							hand = func(m *ctx.Message, c *ctx.Context, key string, a ...string) (e error) {
								w := m.Optionv("response").(http.ResponseWriter)
								template.Must(template.ParseGlob(arg[1])).Execute(w, m)
								return
							}
						}
					}

					if _, ok := m.Target().Commands[arg[0]]; ok {
						m.Target().Commands[arg[0]].Help = help
						m.Target().Commands[arg[0]].Name = arg[1]
						m.Target().Commands[arg[0]].Hand = hand
					} else {
						cmd := &ctx.Command{Name: arg[1], Help: help, Hand: hand}
						m.Target().Commands[arg[0]] = cmd
						mux.HandleCmd(m, arg[0], cmd)
					}
				}
			}
			return
		}},
		"template": &ctx.Command{Name: "template [name [file...]]", Help: "模板管理, name: 模板名, file: 模板文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if web, ok := m.Target().Server.(*WEB); m.Assert(ok) {
				if len(arg) == 0 {
					for _, v := range web.Template.Templates() {
						m.Add("append", "name", v.Name())
					}
					m.Sort("name").Table()
					return
				}

				tmpl := web.Template
				if len(arg) > 1 {
					tmpl = template.Must(web.Template.Clone())
					tmpl = template.Must(tmpl.ParseFiles(arg[1:]...))
				}

				buf := bytes.NewBuffer(make([]byte, 1024))
				tmpl.ExecuteTemplate(buf, arg[0], m)
				m.Echo(string(buf.Bytes()))
			}
			return
		}},
		"componet": &ctx.Command{Name: "componet [group [order [arg...]]]", Help: "添加组件, group: 组件分组, arg...: 组件参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) > 0 && arg[0] == "share" {
				m.Cmd("aaa.role", arg[1], "componet", arg[2], "command", arg[3:])
				m.Echo("%s%s?componet_group=%s&relay=%s", m.Conf("serve", "site"), m.Option("index_path"), arg[2], m.Cmdx("aaa.relay", "share", arg[1], "temp"))
				return
			}

			switch len(arg) {
			case 0:
				m.Cmdy("ctx.config", "componet")
			case 1:
				m.Confm("componet", arg[0], func(index int, val map[string]interface{}) {
					m.Add("append", "ctx", val["componet_ctx"])
					m.Add("append", "cmd", val["componet_cmd"])
					m.Add("append", "name", val["componet_name"])
					m.Add("append", "help", val["componet_help"])
					m.Add("append", "tmpl", val["componet_tmpl"])
					m.Add("append", "view", val["componet_view"])
					m.Add("append", "init", val["componet_init"])
				})
				m.Table()
			case 2:
				m.Cmdy("ctx.config", "componet", strings.Join(arg, "."))
			default:
				switch arg[0] {
				case "create":
					m.Confm("componet", arg[1:3], map[string]interface{}{
						"name": arg[3], "help": arg[4],
						"componet_ctx": arg[5], "componet_cmd": arg[6],
					})
				default:
					componet := m.Confm("componet", arg[:2])
					for i := 2; i < len(arg)-1; i += 2 {
						kit.Chain(componet, arg[i], arg[i+1])
					}
				}
			}
			return
		}},

		"/render": &ctx.Command{Name: "/render template", Help: "渲染模板, template: 模板名称", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 权限检查
			if m.Confs("login", "check") {
				if !m.Options("sessid") || !m.Options("username") || !m.Cmds("aaa.role", m.Option("userrole"), "check", m.Confx("group")) {
					m.Set("option", "group", "login").Set("option", "name", "")
				}
			}

			// 响应类型
			accept_json := strings.HasPrefix(m.Option("accept"), "application/json")
			w := m.Optionv("response").(http.ResponseWriter)
			if !accept_json {
				w.Header().Set("Content-Type", "text/html")
			} else {
				w.Header().Set("Content-Type", "application/json")
			}
			// w.Header().Set("Access-Control-Allow-Origin", "api.map.baidu.com")

			web, ok := m.Target().Server.(*WEB)
			m.Assert(ok)

			// 响应模板
			tmpl := web.Template
			if m.Confs("route", "template_debug") {
				tmpl = template.New("render").Funcs(ctx.CGI)
				tmpl.ParseGlob(path.Join(m.Conf("route", "template_dir"), "/*.tmpl"))
				tmpl.ParseGlob(path.Join(m.Conf("route", "template_dir"), m.Cap("route"), "/*.tmpl"))
			}

			m.Option("title", m.Conf("runtime", "boot.hostname"))
			// 响应数据
			group, order := m.Option("group", kit.Select(m.Conf("route", "componet_index"), m.Option("group"))), m.Option("names")
			list := []interface{}{}

			for _, v := range m.Confv("componet", group).([]interface{}) {
				val := v.(map[string]interface{})
				if order != "" && val["name"].(string) != order {
					continue
				}

				// 查找模块
				msg := m.Find(kit.Select(m.Cap("module"), val["ctx"]))

				// 默认变量
				for k, v := range val {
					switch value := v.(type) {
					case []string:
						msg.Set("option", k, value)
					case string:
						msg.Set("option", k, value)
					default:
						msg.Put("option", k, value)
					}
				}

				// 添加命令
				if kit.Right(val["cmd"]) {
					arg = append(arg, kit.Format(val["cmd"]))
				}
				// 添加参数
				if m.Has("cmds") {
					arg = append(arg, kit.Trans(m.Optionv("cmds"))...)
				} else {
					kit.Map(val["args"], "", func(index int, value map[string]interface{}) {
						if value["name"] != nil {
							arg = append(arg, kit.Select(msg.Option(value["name"].(string)), msg.Parse(value["value"])))
						}
					})
				}

				if len(arg) > 0 {
					if order != "" || kit.Right(val["pre_run"]) {
						// 权限检查
						if m.Confs("login", "check") && m.Cmds("aaa.role", m.Option("userrole"), "check", m.Option("componet_group"), arg[0]) {
							continue
						}
						// 执行命令
						msg.Cmd(arg)
					}
				}

				// 添加响应
				if msg.Appends("qrcode") {
					m.Append("qrcode", msg.Append("qrcode"))
				} else if msg.Appends("directory") {
					m.Append("download_file", fmt.Sprintf("/download/%s", msg.Append("directory")))
					return
				} else if accept_json {
					list = append(list, msg.Meta)
				} else if val["tmpl"] != nil {
					m.Assert(tmpl.ExecuteTemplate(w, val["tmpl"].(string), msg))
				}
			}

			// 生成响应
			if accept_json {
				en := json.NewEncoder(w)
				en.SetIndent("", "  ")
				en.Encode(list)
			}
			return
		}},
		"/upload": &ctx.Command{Name: "/upload key", Help: "上传文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			r := m.Optionv("request").(*http.Request)
			if f, h, e := r.FormFile(kit.Select("upload", arg, 0)); m.Assert(e) {
				defer f.Close()

				// 上传文件
				name := kit.Hashx(f)
				if o, p, e := kit.Create(path.Join(m.Conf("web.upload", "path"), "list", name)); m.Assert(e) {
					defer o.Close()

					f.Seek(0, os.SEEK_SET)
					if n, e := io.Copy(o, f); m.Assert(e) {
						m.Log("upload", "list: %s %d", p, n)

						// 文件摘要
						kind := strings.Split(h.Header.Get("Content-Type"), "/")[0]
						buf := bytes.NewBuffer(make([]byte, 0, 1024))
						fmt.Fprintf(buf, "create_time: %s\n", m.Time("2006-01-02 15:04"))
						fmt.Fprintf(buf, "create_user: %s\n", m.Option("username"))
						fmt.Fprintf(buf, "name: %s\n", h.Filename)
						fmt.Fprintf(buf, "type: %s\n", kind)
						fmt.Fprintf(buf, "hash: %s\n", name)
						fmt.Fprintf(buf, "size: %d\n", n)
						b := buf.Bytes()

						// 保存摘要
						code := kit.Hashs(string(b))
						if o, p, e := kit.Create(path.Join(m.Conf("web.upload", "path"), "meta", code)); m.Assert(e) {
							defer o.Close()

							if n, e := o.Write(b); m.Assert(e) {
								m.Log("upload", "meta: %s %d", p, n)
								m.Cmd("nfs.copy", path.Join(m.Conf("web.upload", "path"), kind, code), p)
							}
						}

						// 文件索引
						if m.Options("river") {
							prefix := []string{"ssh._route", m.Option("dream"), "ssh.data", "insert"}
							suffix := []string{"code", code, "kind", kind, "name", h.Filename, "hash", name, "size", kit.Format(n), "upload_time", m.Time("2006-01-02 15:04")}
							m.Cmd(prefix, kit.Select(kind, m.Option("table")), suffix)
							m.Cmd(prefix, "file", suffix)
						}

						// 返回信息
						if !strings.HasPrefix(m.Option("agent"), "curl") {
							m.Append("size", kit.FmtSize(n))
							m.Append("link", fmt.Sprintf(`<a href="/download/%s" target="_blank">%s</a>`, code, h.Filename))
							m.Append("type", kind)
							m.Append("hash", name)
						} else {
							m.Append("code", code)
							m.Append("type", kind)
							m.Append("name", h.Filename)
							m.Append("size", kit.FmtSize(n))
							m.Append("time", m.Time("2006-01-02 15:04"))
							m.Append("hash", name)
						}
					}
				}
			}
			return
		}},
		"/download/": &ctx.Command{Name: "/download/hash [meta [hash]]", Help: "下载文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)

			kind := kit.Select("meta", kit.Select(m.Option("meta"), arg, 0))
			file := kit.Select(strings.TrimPrefix(key, "/download/"), arg, 1)
			if file == "" {
				if m.Option("userrole") != "root" {
					return
				}
				// 文件列表
				m.Cmd("nfs.dir", path.Join(m.Conf("web.upload", "path"), kind)).Table(func(index int, value map[string]string) {
					name := path.Base(value["path"])
					meta := kit.Linex(value["path"])
					m.Push("time", meta["create_time"])
					m.Push("user", meta["create_user"])
					m.Push("size", kit.FmtSize(int64(kit.Int(meta["size"]))))
					if kind == "image" {
						m.Push("name", name)
					} else {
						m.Push("name", fmt.Sprintf(`<a href="/download/%s" target="_blank">%s</a>`, name, meta["name"]))
					}
					m.Push("hash", meta["hash"][:8])

				})
				m.Sort("time", "time_r").Table()
				return
			}

			if p := m.Cmdx("nfs.path", path.Join(m.Conf("web.upload", "path"), "list", file)); p != "" {
				// 直接下载
				m.Log("info", "download %s direct", p)
				http.ServeFile(w, r, p)
				return
			}
			if p := m.Cmdx("nfs.path", path.Join(m.Conf("web.upload", "path"), kind, file)); p != "" {
				// 下载文件
				meta := kit.Linex(p)
				if p := m.Cmdx("nfs.path", path.Join(m.Conf("web.upload", "path"), "list", meta["hash"])); p != "" {
					m.Log("info", "download %s %s", p, m.Cmdx("nfs.hash", meta["hash"]))
					w.Header().Set("Content-Disposition", fmt.Sprintf("filename=%s", meta["name"]))
					http.ServeFile(w, r, p)
				} else {
					http.NotFound(w, r)
				}
				return
			}

			if m.Option("userrole") != "root" {
				return
			}

			if p := m.Cmdx("nfs.path", file); p != "" {
				// 任意文件
				m.Log("info", "download %s %s", p, m.Cmdx("nfs.hash", p))
				http.ServeFile(w, r, p)
			}
			return
		}},
		"/require/": &ctx.Command{Name: "/require/", Help: "加载脚本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)
			file := strings.TrimPrefix(key, "/require/")

			if p := m.Cmdx("nfs.path", m.Conf("cli.project", "plugin.path"), file); p != "" {
				m.Log("info", "download %s direct", p)
				http.ServeFile(w, r, p)
				return
			}
			return
		}},
		"/proxy/": &ctx.Command{Name: "/proxy/which/method/url", Help: "服务代理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			fields := strings.Split(key, "/")
			m.Cmdy("web.get", "which", fields[2], "method", fields[3], strings.Join(fields, "/"))
			return
		}},
		"/login": &ctx.Command{Name: "/login", Help: "认证", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch {
			case m.Options("cert"): // 注册证书
				msg := m.Cmd("aaa.rsa", "info", m.Option("cert"))
				m.Cmd("aaa.auth", "nodes", msg.Append("route"), "cert", m.Option("cert"))
				m.Append("sess", m.Cmdx("aaa.sess", "nodes", "nodes", msg.Append("route")))

			case m.Options("pull"): // 下载证书
				sess := m.Cmd("aaa.auth", "nodes", m.Option("pull"), "session").Append("key")
				m.Add("append", "username", m.Cmd("aaa.auth", sess, "username").Append("meta"))
				m.Add("append", "cert", (m.Cmd("aaa.auth", "nodes", m.Option("pull"), "cert").Append("meta")))

			case m.Options("bind"): // 绑定设备
				sess := m.Cmd("aaa.auth", "nodes", m.Option("bind"), "session").Append("key")
				if m.Cmd("aaa.auth", sess, "username").Appends("meta") {
					return // 已经绑定
				}

				if m.Cmds("aaa.rsa", "verify", m.Cmd("aaa.auth", "username", m.Option("username"), "cert").Append("meta"), m.Option("code"), m.Option("bind")) {
					m.Cmd("aaa.login", sess, "username", m.Option("username"))
					m.Append("userrole", "root")
				}
			case m.Options("user.cert"): // 用户注册
				if !m.Cmds("aaa.auth", "username", m.Option("username"), "cert") {
					m.Cmd("aaa.auth", "username", m.Option("username"), "cert", m.Option("user.cert"))
				}
				m.Append("username", m.Option("username"))
			}
			return
		}},

		"/publish/": &ctx.Command{Name: "/publish/filename [upgrade script|plugin]", Help: "下载项目", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 下载程序
			key = strings.TrimPrefix(key, "/publish/")
			if key == "bench" {
				key = m.Conf("runtime", "boot.ctx_app") + "." + kit.Select(m.Conf("runtime", "host.GOOS"), m.Option("GOOS")) +
					"." + kit.Select(m.Conf("runtime", "host.GOARCH"), m.Option("GOARCH"))
			}

			p := ""
			if m.Option("upgrade") == "script" {
				// 下载脚本
				if m.Options("missyou") {
					p = m.Cmdx("nfs.path", path.Join(m.Conf("missyou", "path"), m.Option("missyou"), "usr/script", key))
				} else {
					p = m.Cmdx("nfs.path", path.Join("usr/script", key))
				}
			} else if m.Option("upgrade") == "plugin" {
				// 下载插件
				p = m.Cmdx("nfs.path", path.Join(m.Conf("publish", "path"), key, kit.Select("index.shy", m.Option("index"))))

			} else {
				// 下载系统
				if p = m.Cmdx("nfs.path", path.Join(m.Conf("publish", "path"), key)); p == "" {
					p = m.Cmdx("nfs.path", m.Conf("publish", []string{"list", kit.Key(key)}))
				}
			}

			// 下载文件
			if s, e := os.Stat(p); e == nil && !s.IsDir() {
				m.Log("info", "publish %s %s", kit.Hashs(p), p)
				http.ServeFile(m.Optionv("response").(http.ResponseWriter), m.Optionv("request").(*http.Request), p)
			}
			return
		}},
		"/shadow": &ctx.Command{Name: "/shadow", Help: "暗网", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Confm("runtime", "node.port", func(index int, value string) {
				m.Add("append", "ports", value)
			})
			return
		}},

		"/wss": &ctx.Command{Name: "/wss", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)
			agent := r.Header.Get("User-Agent")

			if s, e := websocket.Upgrade(w, r, nil, 4096, 4096); m.Assert(e) {
				h := m.Option("wssid")
				if h == "" || m.Confs("wss", h) {
					h = kit.Hashs("uniq")
				}
				p := make(chan *ctx.Message, 10)
				meta := map[string]interface{}{
					"create_time": m.Time(),
					"create_user": m.Option("username"),
					"agent":       agent,
					"sessid":      m.Option("sessid"),
					"socket":      s,
					"channel":     p,
				}
				m.Conf("wss", []string{m.Option("username"), h}, meta)
				m.Conf("wss", h, meta)
				p <- m.Spawn().Add("detail", "wssid", h)

				what := m
				m.Log("wss", "conn %v %s", h, agent)
				m.Gos(m.Spawn(), func(msg *ctx.Message) {
					for {
						if t, b, e := s.ReadMessage(); e == nil {
							var data interface{}
							if e := json.Unmarshal(b, &data); e == nil {
								m.Log("wss", "recv %s %d msg %v", h, t, data)
							} else {
								m.Log("wss", "recv %s %d msg %v", h, t, b)
								data = b
							}

							what.Optionv("data", data)
							what.Back(what)
						} else {
							m.Log("warn", "wss recv %s %d msg %v", h, t, e)
							close(p)
							break
						}
					}
				})

				for what = range p {
					s.WriteJSON(what.Meta)
				}

				m.Log("wss", "close %s %s", h, agent)
				m.Conf("wss", []string{m.Option("username"), h}, "")
				m.Conf("wss", h, "")
				s.Close()
			}
			return
		}},
		"wss": &ctx.Command{Name: "wss", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 || arg[0] == "" {
				m.Confm("wss", func(key string, value map[string]interface{}) {
					if value["agent"] == nil {
						return
					}
					m.Push("key", m.Cmdx("aaa.short", key))
					m.Push("create_time", value["create_time"])
					m.Push("create_user", value["create_user"])
					m.Push("sessid", kit.Format(value["sessid"])[:6])
					m.Push("agent", value["agent"])
				})
				m.Table()
				return
			}

			list := []string{}
			if strings.Contains(arg[0], ".") {
				vs := strings.SplitN(arg[0], ".", 2)
				m.Confm("wss", vs[0], func(key string, value map[string]interface{}) {
					if len(vs) == 1 || vs[1] == "*" || strings.Contains(kit.Format(value["agent"]), vs[1]) {
						list = append(list, key)
					}
				})
			} else {
				if len(arg[0]) != 32 {
					arg[0] = m.Cmdx("aaa.short", arg[0])
				}
				list = append(list, arg[0])
			}

			if len(arg) == 1 {
				m.Cmdy(".config", "wss", arg[0])
				return
			}

			for _, v := range list {
				if p, ok := m.Confv("wss", []string{v, "channel"}).(chan *ctx.Message); ok {
					if msg := m.Spawn(); arg[1] == "sync" {
						p <- msg.Add("detail", arg[2], arg[3:])
						msg.CallBack(true, func(msg *ctx.Message) *ctx.Message {
							if data, ok := msg.Optionv("data").(map[string]interface{}); ok {
								if len(list) == 1 && data["append"] != nil {
									for _, k := range kit.Trans(data["append"]) {
										m.Push(k, kit.Trans(data[k]))
									}
								} else {
									m.Push("time", m.Time())
									m.Push("key", m.Cmdx("aaa.short", v))
									m.Push("action", kit.Format(arg[2:]))

									res := kit.Trans(data["result"])
									m.Push("return", kit.Format(res))
									m.Log("wss", "result: %v", res)
								}

							} else {
								m.Push("time", m.Time())
								m.Push("key", m.Cmdx("aaa.short", v))
								m.Push("action", kit.Format(arg[2:]))
							}
							return m
						}, "skip")
						m.Table()
					} else {
						p <- msg.Add("detail", arg[1], arg[2:])
					}
				}
			}
			return
		}},
	},
}

func init() {
	ctx.Index.Register(Index, &WEB{Context: Index})
}
