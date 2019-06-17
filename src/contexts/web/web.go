package web

import (
	"bytes"
	"contexts/ctx"
	"encoding/json"
	"fmt"
	"github.com/skip2/go-qrcode"
	"strconv"
	// "github.com/PuerkitoBio/goquery"
	"github.com/go-cas/cas"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"toolkit"
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
	// if msg.Confs("skip_login", msg.Option("path")) {
	// 	return true
	// }
	defer func() {
		msg.Log("info", "access: %s", msg.Option("access", msg.Cmdx("aaa.sess", "access")))
	}()
	if msg.Confs("login", "cas") {
		if !cas.IsAuthenticated(r) {
			r.URL, _ = r.URL.Parse(r.Header.Get("index_url"))
			msg.Log("info", "cas_login %v %v", r.URL, msg.Conf("spide", "cas.client.url"))
			cas.RedirectToLogin(w, r)
			return false
		}

		for k, v := range cas.Attributes(r) {
			for _, val := range v {
				msg.Add("option", k, val)
			}
		}

		msg.Log("info", "cas_login %v", msg.Option("ticket"))
		if msg.Options("ticket") {
			msg.Option("username", cas.Username(r))
			msg.Log("info", "login: %s", msg.Option("username"))
			http.SetCookie(w, &http.Cookie{Name: "sessid", Value: msg.Option("sessid", msg.Cmdx("aaa.user", "session", "select")), Path: "/"})
			http.Redirect(w, r, merge(msg, r.Header.Get("index_url"), "ticket", ""), http.StatusTemporaryRedirect)
			return false
		}
	} else if msg.Options("username") && msg.Options("password") {
		if msg.Cmds("aaa.auth", "username", msg.Option("username"), "password", msg.Option("password")) {
			msg.Log("info", "login: %s", msg.Option("username"))
			http.SetCookie(w, &http.Cookie{Name: "sessid", Value: msg.Cmdx("aaa.user", "session", "select"), Path: "/"})
			if msg.Options("relay") {
				if role := msg.Cmdx("aaa.relay", "check", msg.Option("relay"), "userrole"); role != "" {
					msg.Cmd("aaa.role", role, "user", msg.Option("username"))
					msg.Log("info", "relay: %s", role)
				}
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
		return false
	}

	if msg.Log("info", "sessid: %s", msg.Option("sessid")); msg.Options("sessid") {
		if msg.Log("info", "username: %s", msg.Option("username", msg.Cmd("aaa.sess", "user").Append("meta"))); msg.Options("username") {
			if msg.Log("info", "nickname: %s", msg.Option("nickname", msg.Cmdx("aaa.auth", "username", msg.Option("username"), "data", "nickname"))); !msg.Options("nickname") {
				msg.Option("nickname", msg.Option("username"))
			}
		}
	}

	if !msg.Options("username") && msg.Options("relay") {
		if relay := msg.Cmd("aaa.relay", "check", msg.Option("relay")); relay.Appends("username") {
			if role := msg.Cmdx("aaa.relay", "check", msg.Option("relay"), "userrole"); role != "" {
				msg.Log("info", "login: %s", msg.Option("username", relay.Append("username")))
				http.SetCookie(w, &http.Cookie{Name: "sessid", Value: msg.Cmdx("aaa.user", "session", "select"), Path: "/"})
				msg.Cmd("aaa.role", role, "user", msg.Option("username"))
				msg.Log("info", "relay: %s", role)
			}
		}
	}

	return true
}
func (web *WEB) HandleCmd(m *ctx.Message, key string, cmd *ctx.Command) {
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		m.TryCatch(m.Spawn(m.Conf("serve", "autofree")), true, func(msg *ctx.Message) {
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
			msg.Option("dir_root", msg.Cap("directory"))
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
				json.NewDecoder(r.Body).Decode(&data)
				msg.Optionv("content_data", data)

				switch d := data.(type) {
				case map[string]interface{}:
					for k, v := range d {
						for _, v := range kit.Trans(v) {
							msg.Add("option", k, v)
						}
					}
				}
			}

			// 请求系统
			// msg.Option("GOOS", m.Conf("runtime", "host.GOOS"))
			// msg.Option("GOARCH", m.Conf("runtime", "host.GOARCH"))
			// agent := r.Header.Get("User-Agent")
			// switch {
			// case strings.Contains(agent, "Macintosh"):
			// 	msg.Option("GOOS", "darwin")
			// }
			// switch {
			// case strings.Contains(agent, "Intel"):
			// 	msg.Option("GOARCH", "386")
			// }
			//

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

			case msg.Has("componet"):
				msg.Spawn().Add("option", "componet_group", msg.Meta["componet"]).Cmd("/render")

			case msg.Has("qrcode"):
				w.Header().Set("Content-Type", "image/png")
				qr, e := qrcode.New(msg.Append("qrcode"), qrcode.Medium)
				m.Assert(e)
				m.Assert(qr.Write(256, w))

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
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(WEB)
	s.Context = c
	return s
}
func (web *WEB) Begin(m *ctx.Message, arg ...string) ctx.Server {
	web.Caches["master"] = &ctx.Cache{Name: "master(yes/no)", Value: "no", Help: "服务入口"}
	web.Caches["register"] = &ctx.Cache{Name: "register(yes/no)", Value: "no", Help: "是否已初始化"}
	web.Caches["route"] = &ctx.Cache{Name: "route", Value: "/" + web.Context.Name + "/", Help: "模块路由"}

	web.ServeMux = http.NewServeMux()
	web.Template = template.New("render").Funcs(ctx.CGI)
	web.Template.ParseGlob(path.Join(m.Cap("directory"), m.Conf("serve", "template_dir"), m.Cap("route"), "/*.tmpl"))
	return web
}
func (web *WEB) Start(m *ctx.Message, arg ...string) bool {
	web.Caches["directory"] = &ctx.Cache{Name: "directory", Value: kit.Select(m.Conf("serve", "directory"), arg, 0), Help: "服务目录"}
	web.Caches["protocol"] = &ctx.Cache{Name: "protocol", Value: kit.Select(m.Conf("serve", "protocol"), arg, 2), Help: "服务协议"}
	web.Caches["address"] = &ctx.Cache{Name: "address", Value: kit.Select(m.Conf("serve", "address"), arg, 1), Help: "服务地址"}
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

			// 通用响应
			if m.Target().Commands["/render"] == nil {
				m.Target().Commands["/render"] = render
			}
			if m.Target().Commands["/proxy/"] == nil {
				m.Target().Commands["/proxy/"] = proxy
			}

			// 路由节点
			msg := m.Target().Message()
			for k, x := range m.Target().Commands {
				if k[0] == '/' {
					m.Log("info", "%d route: %s", m.Capi("nroute", 1), k)
					h.HandleCmd(msg, k, x)
				}
			}

			// 路由文件
			if m.Cap("directory") != "" {
				m.Log("info", "route: %sstatic/ <- [%s]\n", m.Cap("route"), m.Cap("directory"))
				h.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(m.Cap("directory")))))
			}
		}
		return false
	})

	// SSO认证
	var handler http.Handler
	if cas_url, e := url.Parse(m.Cmdx("web.spide", "cas", "client", "url")); e == nil && cas_url.Host != "" {
		m.Log("info", "cas url: %s", cas_url)
		m.Conf("login", "cas", "true")
		client := cas.NewClient(&cas.Options{URL: cas_url})
		handler = client.Handle(web)
	} else {
		handler = web
	}

	m.Log("info", "web: %s", m.Cap("address"))
	// 启动服务
	web.Server = &http.Server{Addr: m.Cap("address"), Handler: handler}
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
			"address":    ":9094",
			"cert":       "etc/cert.pem",
			"key":        "etc/key.pem",
			"site":       "",
			"index":      "/code/",
		}, Help: "服务配置"},
		"route": &ctx.Config{Name: "route", Value: map[string]interface{}{
			"index":          "/render",
			"template_dir":   "template",
			"template_debug": true,
			"componet_index": "index",
			"toolkit_view": map[string]interface{}{
				"top": 96, "left": 472, "width": 600, "height": 300,
			},
		}, Help: "功能配置"},
		"login": &ctx.Config{Name: "login", Value: map[string]interface{}{
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
		"toolkit": &ctx.Config{Name: "toolkit", Value: map[string]interface{}{
			"time": map[string]interface{}{"cmd": "time"},
		}, Help: "工具列表"},
	},
	Commands: map[string]*ctx.Command{
		"_init": &ctx.Command{Name: "_init", Help: "post请求", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmd("web.spide", "cas", "client", "new", m.Conf("runtime", "boot.ctx_cas"))
			m.Cmd("web.spide", "dev", "client", "new", kit.Select(m.Conf("runtime", "boot.ctx_dev"), m.Conf("runtime", "boot.ctx_box")))
			return
		}},
		"spide": &ctx.Command{Name: "spide [which [client|cookie [name [value]]]]", Help: "爬虫配置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch len(arg) {
			case 0:
				m.Confm("spide", func(key string, value map[string]interface{}) {
					m.Add("append", "key", key)
					m.Add("append", "protocol", kit.Chains(value, "client.protocol"))
					m.Add("append", "hostname", kit.Chains(value, "client.hostname"))
					m.Add("append", "path", kit.Chains(value, "client.path"))
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
		"get": &ctx.Command{Name: "get [which] name [method GET|POST] url arg...", Help: "访问服务, method: 请求方法, url: 请求地址, arg: 请求参数",
			Form: map[string]int{"which": 1, "method": 1, "headers": 2, "content_type": 1, "file": 2, "body": 1, "parse": 1, "temp": -1, "save": 1, "temp_expire": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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

					if m.Has("file") { // POST file
						writer := multipart.NewWriter(&bytes.Buffer{})
						defer writer.Close()

						for k, v := range uuu.Query() {
							for _, v := range v {
								writer.WriteField(k, v)
							}
						}

						file, e := os.Open(m.Cmdx("nfs.path", m.Meta["file"][1]))
						m.Assert(e)
						defer file.Close()

						part, e := writer.CreateFormFile(m.Option("file"), filepath.Base(m.Meta["file"][1]))
						m.Assert(e)
						io.Copy(part, file)

						m.Option("content_type", writer.FormDataContentType())
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
									for _, val := range v {
										if i, e := strconv.Atoi(v[0]); e == nil {
											data = kit.Chain(data, []string{k, "-2"}, i)
										} else {
											data = kit.Chain(data, []string{k, "-2"}, val)
										}
									}
								}
							}

							b, e := json.Marshal(data)
							m.Assert(e)
							m.Log("info", "json %v", string(b))
							body = bytes.NewReader(b)

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

				// 请求cookie
				kit.Structm(m.Magic("user", []string{"cookie", which}), func(key string, value string) {
					req.AddCookie(&http.Cookie{Name: key, Value: value})
					m.Log("info", "set-cookie %s: %v", key, value)

				})

				if web.Client == nil {
					web.Client = &http.Client{Timeout: kit.Duration(kit.Format(client["timeout"]))}
				}

				// 请求日志
				if kit.Right(client["logheaders"]) {
					for k, v := range req.Header {
						m.Log("info", "%s: %s", k, v)
					}
					m.Log("info", "")
				}

				// 发送请求
				res, e := web.Client.Do(req)
				if e != nil {
					m.Log("warn", "%v", e)
					return e
				}

				// 响应日志
				var result interface{}
				ct := res.Header.Get("Content-Type")
				parse := kit.Select(kit.Format(client["parse"]), m.Option("parse"))
				m.Log("info", "status %s parse: %s content: %s", res.Status, parse, ct)
				if kit.Right(client["logheaders"]) {
					for k, v := range res.Header {
						m.Log("info", "%s: %v", k, v)
					}
				}

				// 响应失败
				if res.StatusCode != http.StatusOK {
					m.Echo("%d: %s", res.StatusCode, res.Status)
				}

				// 响应cookie
				for _, v := range res.Cookies() {
					if m.Log("info", "get-cookie %s: %v", v.Name, v.Value); v.Value != "" {
						m.Magic("user", []string{"cookie", which, v.Name}, v.Value)
					}
				}

				// 解析响应
				switch {
				case parse == "json" || strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "application/javascript"):
					if json.NewDecoder(res.Body).Decode(&result); !m.Has("temp") {
						m.Option("temp", "")
					}
					m.Put("option", "data", result).Cmdy("mdb.temp", "url", uri+uri_arg, "data", "data", m.Meta["temp"])

				case parse == "html":
					/*
						page, e := goquery.NewDocumentFromReader(res.Body)
						m.Assert(e)

						page.Find(kit.Select("html", m.Option("parse_chain"))).Each(func(n int, s *goquery.Selection) {
							if m.Options("parse_select") {
								for i := 0; i < len(m.Meta["parse_select"])-2; i += 3 {
									item := s.Find(m.Meta["parse_select"][i+1])
									if m.Meta["parse_select"][i+1] == "" {
										item = s
									}
									if v, ok := item.Attr(m.Meta["parse_select"][i+2]); ok {
										m.Add("append", m.Meta["parse_select"][i], v)
										m.Log("info", "item attr %v", v)
									} else {
										m.Add("append", m.Meta["parse_select"][i], strings.Replace(item.Text(), "\n", "", -1))
										m.Log("info", "item text %v", item.Text())
									}
								}
								return
							}

							s.Find("a").Each(func(n int, s *goquery.Selection) {
								if attr, ok := s.Attr("href"); ok {
									s.SetAttr("href", proxy(m, attr))
								}
							})
							s.Find("img").Each(func(n int, s *goquery.Selection) {
								if attr, ok := s.Attr("src"); ok {
									s.SetAttr("src", proxy(m, attr))
								}
								if attr, ok := s.Attr("r-lazyload"); ok {
									s.SetAttr("src", proxy(m, attr))
								}
							})
							s.Find("script").Each(func(n int, s *goquery.Selection) {
								if attr, ok := s.Attr("src"); ok {
									s.SetAttr("src", proxy(m, attr))
								}
							})

							if html, e := s.Html(); e == nil {
								m.Add("append", "html", html)
							}
						})
						m.Table()

					*/
				default:
					if res.StatusCode == http.StatusOK && m.Options("save") {
						dir := path.Dir(m.Option("save"))
						if _, e = os.Stat(dir); e != nil {
							m.Assert(os.MkdirAll(dir, 0777))
						}

						f, e := os.Create(m.Option("save"))
						m.Assert(e)
						defer f.Close()

						n, e := io.Copy(f, res.Body)
						m.Assert(e)
						m.Log("info", "save %d %s", n, m.Option("save"))
						m.Echo(m.Option("save"))
					} else {
						buf, e := ioutil.ReadAll(res.Body)
						m.Assert(e)
						m.Echo(string(buf))
					}
				}
				return
			}},
		"post": &ctx.Command{Name: "post args...", Help: "post请求", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy("web.get", "method", "POST", arg)
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
		"template": &ctx.Command{Name: "template [file [directory]]|[name [content]]", Help: "添加模板, content: 模板内容, directory: 模板目录", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if web, ok := m.Target().Server.(*WEB); m.Assert(ok) {
				if len(arg) == 0 {
					for _, v := range web.Template.Templates() {
						m.Add("append", "name", v.Name())
					}
					m.Sort("name").Table()
					return
				}

				if web.Template == nil {
					web.Template = template.New("render").Funcs(ctx.CGI)
				}

				dir := path.Join(m.Cap("directory"), m.Confx("template_dir", arg, 1), arg[0])
				if t, e := web.Template.ParseGlob(dir); e == nil {
					web.Template = t
				} else {
					m.Log("info", "%s", e)
					if len(arg) > 1 {
						web.Template = template.Must(web.Template.New(arg[0]).Parse(arg[1]))
					} else {
						tmpl, e := web.Template.Clone()
						m.Assert(e)
						tmpl.Funcs(ctx.CGI)

						buf := bytes.NewBuffer(make([]byte, 1024))
						tmpl.ExecuteTemplate(buf, arg[0], m)
						m.Echo(string(buf.Bytes()))
					}
				}
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
			if m.Options("toolkit") {
				if kit, ok := m.Confv("toolkit", m.Option("toolkit")).(map[string]interface{}); ok {
					m.Sess("cli").Cmd(kit["cmd"], m.Option("argument")).CopyTo(m)
				}
				return
			}

			if web, ok := m.Target().Server.(*WEB); m.Assert(ok) {
				// 响应类型
				accept_json := strings.HasPrefix(m.Option("accept"), "application/json")
				w := m.Optionv("response").(http.ResponseWriter)
				if accept_json {
					// w.Header().Add("Content-Type", "application/json")
				} else {
					w.Header().Add("Content-Type", "text/html")
				}

				// 响应数据
				list := []interface{}{}
				tmpl := web.Template
				if m.Confs("route", "template_debug") {
					tmpl = template.New("render").Funcs(ctx.CGI)
					tmpl.ParseGlob(path.Join(m.Cap("directory"), m.Conf("route", "template_dir"), "/*.tmpl"))
					tmpl.ParseGlob(path.Join(m.Cap("directory"), m.Conf("route", "template_dir"), m.Cap("route"), "/*.tmpl"))
				}

				// 权限检查
				if m.Confs("login", "check") {
					if m.Option("username") == "" { // 没有登录
						m.Set("option", "componet_group", "login").Set("option", "componet_name", "").Set("option", "bench", "")
					} else {
						// 创建空间
						if bench := m.Option("bench"); m.Option("bench", m.Cmdx("aaa.sess", "bench", "select")) != bench {
							m.Append("redirect", merge(m, m.Option("index_url"), "bench", m.Option("bench")))
							return
						}
						m.Optionv("bench_data", m.Confv("auth", []string{m.Option("bench"), "data"}))

						if !m.Cmds("aaa.work", "right", m.Confx("componet_group")) { // 没有权限
							m.Set("option", "componet_group", "login").Set("option", "componet_name", "").Set("option", "bench", "")
						}
					}
				}

				// 响应模板
				group, order := m.Option("componet_group", kit.Select(m.Conf("route", "componet_index"), m.Option("componet_group"))), m.Option("componet_name")

				for _, v := range m.Confv("componet", group).([]interface{}) {
					val := v.(map[string]interface{})
					if order != "" && val["componet_name"].(string) != order {
						continue
					}

					// 查找模块
					msg := m.Find(kit.Select(m.Cap("module"), val["componet_ctx"]))

					// 默认变量
					for k, v := range val {
						if msg.Option(k) != "" {
							continue
						}
						switch value := v.(type) {
						case []string:
							msg.Add("option", k, value)
						case string:
							msg.Add("option", k, value)
						default:
							msg.Put("option", k, value)
						}
					}
					// 默认参数
					if val["inputs"] != nil {
						for _, v := range val["inputs"].([]interface{}) {
							value := v.(map[string]interface{})
							if value["name"] != nil && msg.Option(value["name"].(string)) == "" {
								msg.Add("option", value["name"].(string), m.Parse(value["value"]))
							}
						}
					}

					// 添加设备
					arg = arg[:0]
					if kit.Right(val["componet_pod"]) {
						arg = append(arg, "remote", kit.Format(m.Magic("session", "current.pod")))
					}
					// 添加命令
					if kit.Right(val["componet_cmd"]) {
						arg = append(arg, kit.Format(val["componet_cmd"]))
					}
					if m.Has("cmds") {
						arg = append(arg, kit.Trans(m.Optionv("cmds"))...)
					}
					// 添加参数
					for _, v := range kit.Trans(val["componet_args"]) {
						arg = append(arg, msg.Parse(v))
					}

					if len(arg) > 0 {
						// 权限检查
						if m.Options("bench") && !m.Cmds("aaa.work", "right", m.Option("componet_group"), arg[0]) {
							continue
						}

						m.Option("remote", "true")

						// 执行命令
						if order != "" || kit.Right(val["pre_run"]) {
							if msg.Cmd(arg); m.Options("bench") {
								name_alias := "action." + kit.Select(msg.Option("componet_name"), msg.Option("componet_name_alias"))

								// 命令历史
								msg.Put("option", name_alias, map[string]interface{}{
									"cmd": arg, "order": m.Option("componet_name_order"), "action_time": msg.Time(),
								}).Cmd("aaa.work", m.Option("bench"), "data", "option", name_alias, "modify_time", msg.Time())
							}
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
					} else if val["componet_tmpl"] != nil {
						m.Assert(tmpl.ExecuteTemplate(w, val["componet_tmpl"].(string), msg))
					}
				}

				// 生成响应
				if accept_json {
					en := json.NewEncoder(w)
					en.SetIndent("", "  ")
					en.Encode(list)
				}
			}
			return
		}},
		"/upload": &ctx.Command{Name: "/upload", Help: "上传文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			r := m.Optionv("request").(*http.Request)
			f, h, e := r.FormFile("upload")
			m.Assert(e)
			defer f.Close()

			p := path.Join(m.Cmdx("nfs.path", m.Magic("session", "current.dir")), h.Filename)
			m.Log("upload", "file: %s", p)
			m.Echo("%s", p)

			o, e := os.Create(p)
			m.Assert(e)
			defer o.Close()

			io.Copy(o, f)
			return
		}},
		"/download/": &ctx.Command{Name: "/download/", Help: "下载文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)

			p := m.Cmdx("nfs.path", strings.TrimPrefix(key, "/download/"))
			m.Log("info", "download %s %s", p, m.Cmdx("aaa.hash", "file", p))

			http.ServeFile(w, r, p)
			return
		}},
		"/proxy/": &ctx.Command{Name: "/proxy/which/method/url", Help: "服务代理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			fields := strings.Split(key, "/")
			m.Cmdy("web.get", "which", fields[2], "method", fields[3], strings.Join(fields, "/"))
			return
		}},

		"/publish/": &ctx.Command{Name: "/publish/", Help: "下载文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			key = strings.TrimPrefix(key, "/publish/")
			if strings.HasSuffix(key, "bench") {
				key = key + "." + m.Option("GOOS") + "." + m.Option("GOARCH")
			}
			p := ""
			if m.Option("upgrade") == "script" {
				p = m.Cmdx("nfs.path", path.Join("usr/script", key))
			}
			key = strings.Replace(key, ".", "_", -1)
			if p == "" {
				p = m.Cmdx("nfs.path", path.Join(m.Conf("publish", "path"), key))
			}
			if p == "" {
				p = m.Cmdx("nfs.path", m.Conf("publish", []string{"list", key}))
			}

			m.Log("info", "publish %s %s", kit.Hashs(p), p)
			http.ServeFile(m.Optionv("response").(http.ResponseWriter), m.Optionv("request").(*http.Request), p)
			return
		}},
		"/shadow": &ctx.Command{Name: "/shadow", Help: "暗网", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Confm("runtime", "node.port", func(index int, value string) {
				m.Add("append", "ports", value)
			})
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
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)
}
