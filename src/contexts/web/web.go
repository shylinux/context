package web

import (
	"bytes"
	"contexts/ctx"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
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
func Merge(m *ctx.Message, uri string, arg ...string) string {
	uri = strings.Replace(uri, ":/", "://", -1)
	uri = strings.Replace(uri, ":///", "://", -1)
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

func (web *WEB) HandleCmd(m *ctx.Message, key string, cmd *ctx.Command) {
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		m.TryCatch(m.Spawn(), true, func(msg *ctx.Message) {
			msg.Option("remote_addr", r.RemoteAddr)
			msg.Option("remote_ip", r.Header.Get("remote_ip"))
			msg.Option("index_url", r.Header.Get("index_url"))
			msg.Option("index_path", r.Header.Get("index_path"))
			msg.Option("referer", r.Header.Get("Referer"))
			msg.Option("accept", r.Header.Get("Accept"))
			msg.Option("method", r.Method)
			msg.Option("path", r.URL.Path)

			msg.Option("dir_root", msg.Cap("directory"))
			for _, v := range r.Cookies() {
				msg.Option(v.Name, v.Value)
			}

			r.ParseMultipartForm(int64(msg.Confi("multipart_bsize")))
			if r.ParseForm(); len(r.PostForm) > 0 {
				for k, v := range r.PostForm {
					msg.Log("info", "%s: %v", k, v)
				}
				msg.Log("info", "")
			}
			for k, v := range r.Form {
				msg.Add("option", k, v)
			}

			msg.Put("option", "request", r).Put("option", "response", w).Sess("web", msg)
			if msg.Confs("cas_url") {
				if !cas.IsAuthenticated(r) && !msg.Confs("skip_cas") {
					r.URL, _ = r.URL.Parse(r.Header.Get("index_url"))
					cas.RedirectToLogin(w, r)
					return
				}

				msg.Option("username", cas.Username(r))
				for k, v := range cas.Attributes(r) {
					for _, val := range v {
						msg.Add("option", k, val)
					}
				}

				if msg.Options("ticket") {
					msg.Option("uuid", msg.Option(msg.Conf("cas_uuid")))
					msg.Option("sessid", msg.Spawn().Cmd("session", "uuid").Result(0))

					uri, _ := r.URL.Parse(r.Header.Get("index_url"))
					redirect := uri.Path
					if b := uri.Query().Get("bench"); b != "" {
						redirect += "?bench=" + b
					}
					http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
					return
				}
			} else if msg.Options("username") && msg.Options("password") {
				if sessid := msg.Spawn().Cmd("session", "password").Result(0); sessid != "" {
					msg.Option("sessid", sessid)
					msg.Option("password", "")
				} else {
					w.WriteHeader(http.StatusUnauthorized)
					msg.Option("username", "")
				}
				return
			}

			msg.Log("cmd", "%s [] %v", key, msg.Meta["option"])
			cmd.Hand(msg, msg.Target(), msg.Option("path"))

			switch {
			case msg.Has("redirect"):
				http.Redirect(w, r, msg.Append("redirect"), http.StatusTemporaryRedirect)
			case msg.Has("directory"):
				http.ServeFile(w, r, msg.Append("directory"))
			case msg.Has("componet"):
				msg.Spawn().Add("option", "componet_group", msg.Meta["componet"]).Cmd("/render")
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
		} else {
			r.Header.Set("remote_ip", strings.Split(r.RemoteAddr, ":")[0])
		}

		m.Log("info", "").Log("info", "%v %s %s", r.Header.Get("remote_ip"), r.Method, r.URL)
		r.Header.Set("index_module", m.Cap("module"))
		r.Header.Set("index_url", r.URL.String())
		r.Header.Set("index_path", r.URL.Path)
	}

	if index && m.Confs("logheaders") {
		for k, v := range r.Header {
			m.Log("info", "%s: %v", k, v)
		}
		m.Log("info", "")
	}

	if r.URL.Path == "/" && m.Confs("root_index") {
		r.URL.Path = m.Conf("root_index")
	}
	web.ServeMux.ServeHTTP(w, r)

	if index && m.Confs("logheaders") {
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
	web.Configs["skip_cas"] = &ctx.Config{Name: "skip_cas", Value: "false", Help: "默认路由"}
	web.Configs["root_index"] = &ctx.Config{Name: "root_index", Value: "/render", Help: "默认路由"}
	web.Configs["logheaders"] = &ctx.Config{Name: "logheaders(yes/no)", Value: "no", Help: "日志输出报文头"}
	web.Configs["template_sub"] = &ctx.Config{Name: "template_sub", Value: web.Context.Name, Help: "模板文件"}

	web.Caches["directory"] = &ctx.Cache{Name: "directory", Value: m.Confx("directory", arg, 0), Help: "服务目录"}
	web.Caches["route"] = &ctx.Cache{Name: "route", Value: "/" + web.Context.Name + "/", Help: "模块路由"}
	web.Caches["register"] = &ctx.Cache{Name: "register(yes/no)", Value: "no", Help: "是否已初始化"}
	web.Caches["master"] = &ctx.Cache{Name: "master(yes/no)", Value: "no", Help: "服务入口"}

	web.ServeMux = http.NewServeMux()
	web.Template = template.New("render").Funcs(ctx.CGI)
	web.Template.ParseGlob(path.Join(m.Conf("template_dir"), m.Conf("template_sub"), "/*.tmpl"))
	return web
}
func (web *WEB) Start(m *ctx.Message, arg ...string) bool {
	m.Cap("directory", m.Confx("directory", arg, 0))

	render := m.Target().Commands["/render"]
	proxy := m.Target().Commands["/proxy/"]

	m.Travel(func(m *ctx.Message, i int) bool {
		if h, ok := m.Target().Server.(MUX); ok && m.Cap("register") == "no" {
			m.Cap("register", "yes")

			p := m.Target().Context()
			if s, ok := p.Server.(MUX); ok {
				m.Log("info", "route: /%s <- /%s", p.Name, m.Target().Name)
				s.Handle(m.Cap("route"), http.StripPrefix(path.Dir(m.Cap("route")), h))
			}

			if m.Target().Commands["/render"] == nil {
				m.Target().Commands["/render"] = render
			}
			if m.Target().Commands["/proxy/"] == nil {
				m.Target().Commands["/proxy/"] = proxy
			}

			msg := m.Target().Message()
			for k, x := range m.Target().Commands {
				if k[0] == '/' {
					m.Log("info", "route: %s", k)
					h.HandleCmd(msg, k, x)
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

	web.Caches["protocol"] = &ctx.Cache{Name: "protocol", Value: m.Confx("protocol", arg, 2), Help: "服务协议"}
	web.Caches["address"] = &ctx.Cache{Name: "address", Value: m.Confx("address", arg, 1), Help: "服务地址"}
	m.Log("info", "%d %s://%s", m.Capi("nserve", 1), m.Cap("protocol"), m.Cap("stream", m.Cap("address")))

	var handler http.Handler
	if cas_url, e := url.Parse(m.Conf("cas_url")); e == nil && m.Confs("cas_url") {
		m.Log("info", "cas url: %s", m.Conf("cas_url"))
		client := cas.NewClient(&cas.Options{URL: cas_url})
		handler = client.Handle(web)
	} else {
		handler = web
	}

	web.Server = &http.Server{Addr: m.Cap("address"), Handler: handler}

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
		"log_uri":         &ctx.Config{Name: "log_uri", Value: "false", Help: "输出请求"},
		"method":          &ctx.Config{Name: "method", Value: "GET", Help: "请求方法"},
		"body_response":   &ctx.Config{Name: "body_response", Value: "response", Help: "响应缓存"},
		"multipart_bsize": &ctx.Config{Name: "multipart_bsize", Value: "102400", Help: "缓存大小"},
		"brow_home":       &ctx.Config{Name: "brow_home", Value: "http://localhost:9094", Help: "浏览服务"},
		"spide":           &ctx.Config{Name: "spide", Value: map[string]interface{}{}, Help: "浏览服务"},
		"client_timeout":  &ctx.Config{Name: "client_timeout", Value: "100s", Help: "浏览服务"},

		"directory": &ctx.Config{Name: "directory", Value: "usr", Help: "服务目录"},
		"protocol":  &ctx.Config{Name: "protocol", Value: "http", Help: "服务协议"},
		"address":   &ctx.Config{Name: "address", Value: ":9094", Help: "服务地址"},
		"site":      &ctx.Config{Name: "site", Value: "", Help: "网站地址"},
		"cert":      &ctx.Config{Name: "cert", Value: "etc/cert.pem", Help: "证书"},
		"key":       &ctx.Config{Name: "key", Value: "etc/key.pem", Help: "密钥"},

		"login_right": &ctx.Config{Name: "login_right", Value: "1", Help: "登录认证"},
		"sess_void":   &ctx.Config{Name: "sess_void", Value: "0", Help: "匿名会话"},
		"cas_url":     &ctx.Config{Name: "cas_url", Value: "", Help: "单点登录"},
		"cas_uuid":    &ctx.Config{Name: "cas_uuid", Value: "email", Help: "单点登录"},

		"toolkit": &ctx.Config{Name: "toolkit", Value: map[string]interface{}{
			"time": map[string]interface{}{
				"cmd": "time",
			},
		}, Help: "小工具"},

		"library_dir":      &ctx.Config{Name: "library_dir", Value: "usr/librarys", Help: "脚本目录"},
		"template_dir":     &ctx.Config{Name: "template_dir", Value: "usr/template", Help: "模板目录"},
		"template_debug":   &ctx.Config{Name: "template_debug", Value: "true", Help: "模板调试"},
		"componet_context": &ctx.Config{Name: "component_context", Value: "nfs", Help: "默认模块"},
		"componet_command": &ctx.Config{Name: "component_command", Value: "pwd", Help: "默认命令"},
		"componet_group":   &ctx.Config{Name: "component_group", Value: "index", Help: "默认组件"},
		"componet": &ctx.Config{Name: "componet", Value: map[string]interface{}{
			"index": []interface{}{
				map[string]interface{}{"componet_name": "head", "template": "head"},
				map[string]interface{}{"componet_name": "clipbaord", "componet_help": "clipbaord", "template": "clipboard"},
				map[string]interface{}{"componet_name": "time", "componet_help": "time", "template": "componet",
					"componet_ctx": "cli", "componet_cmd": "time", "arguments": []interface{}{"@string"},
					"inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "time_format",
							"label": "format", "value": "2006-01-02 15:04:05",
						},
						map[string]interface{}{"type": "text", "name": "string", "label": "string"},
						map[string]interface{}{"type": "button", "label": "refresh"},
					},
				},
				map[string]interface{}{"componet_name": "tail", "template": "tail"},
			},
		}, Help: "组件列表"},

		"bench_disable": &ctx.Config{Name: "bench_disable", Value: "false", Help: "工作流"},
		"bench":         &ctx.Config{Name: "bench", Value: map[string]interface{}{}, Help: "工作流"},
		"bench_view": &ctx.Config{Name: "bench_view", Value: map[string]interface{}{
			"base": []interface{}{"key", "share", "comment", "creator", "create_time", "modify_time", "commands"},
			"link": []interface{}{"share", "comment", "creator", "link"},
		}, Help: "工作流"},
		"docker_view": &ctx.Config{Name: "docker_view", Value: map[string]interface{}{
			"top": 200, "left": 10, "width": 480, "height": 320,
		}, Help: "工作流"},
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
			Form: map[string]int{"method": 1, "headers": 2, "content_type": 1, "body": 1, "path_value": 1, "body_response": 1, "parse": 1, "sub_parse": 3, "save": 1},
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

					method := m.Conf("method")
					if m.Has("method") {
						method = m.Option("method")
					}
					uri := Merge(m, arg[0], arg[1:]...)
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
					if m.Confs("log_uri") {
						m.Echo("%s: %s\n", req.Method, req.URL)
					}
					if m.Confs("logheaders") {
						for k, v := range req.Header {
							m.Log("info", "%s: %s", k, v)
						}
					}

					if web.Client == nil {
						d, e := time.ParseDuration(m.Conf("client_timeout"))
						m.Assert(e)
						web.Client = &http.Client{Timeout: d}
					}
					res, e := web.Client.Do(req)
					if e != nil {
						m.Log("info", "get error %v", e)
						return
					}
					m.Assert(e)
					if m.Confs("logheaders") {
						for k, v := range res.Header {
							m.Log("info", "%s: %v", k, v)
						}
					}

					for _, v := range res.Cookies() {
						m.Confv("cookie", v.Name, v)
						m.Log("info", "set-cookie %s: %v", v.Name, v.Value)
					}

					var result interface{}
					defer func() {
						m.Target().Configs[m.Confx("body_response")] = &ctx.Config{Value: result}
						m.Log("info", "cache %s", m.Confx("body_response"))
					}()

					if m.Has("save") {
						p := m.Option("save")
						if !strings.Contains(m.Option("save"), "/") {
							p = path.Join(m.Sess("nfs").Cmd("pwd").Result(0), m.Option("save"))
						}

						f, e := os.Create(p)
						m.Assert(e)
						io.Copy(f, res.Body)
						defer f.Close()
						m.Log("info", "save file %s %s", p, m.Sess("aaa").Cmd("hash", "file", p).Result(0))
						m.Echo(p)
						return
					}

					ct := res.Header.Get("Content-Type")
					m.Log("info", "content: %s", ct)

					switch {
					case strings.HasPrefix(ct, "application/json"):
						json.NewDecoder(res.Body).Decode(&result)
						if m.Has("parse") {
							msg := m.Spawn().Put("option", "data", result).Cmd("trans", "data", m.Option("parse"))
							m.Copy(msg, "append").Copy(msg, "result")
							return
						}
						b, _ := json.MarshalIndent(result, "", "  ")
						result = string(b)
					case strings.HasPrefix(ct, "text/html"):
						page, e := goquery.NewDocumentFromReader(res.Body)
						m.Assert(e)

						query := page.Find("html")
						if m.Has("parse") {
							query = query.Find(m.Option("parse"))
						}

						query.Each(func(n int, s *goquery.Selection) {
							if m.Options("sub_parse") {
								for i := 0; i < len(m.Meta["sub_parse"])-2; i += 3 {
									item := s.Find(m.Meta["sub_parse"][i+1])
									if m.Meta["sub_parse"][i+1] == "" {
										item = s
									}
									if v, ok := item.Attr(m.Meta["sub_parse"][i+2]); ok {
										m.Add("append", m.Meta["sub_parse"][i], v)
										m.Log("info", "item attr %v", v)
									} else {
										m.Add("append", m.Meta["sub_parse"][i], strings.Replace(item.Text(), "\n", "", -1))
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
						result = ""
					case strings.HasPrefix(ct, "text"):
						buf, e := ioutil.ReadAll(res.Body)
						m.Assert(e)
						if m.Has("parse") {
							json.Unmarshal(buf, &result)
							msg := m.Spawn()
							msg.Put("option", "response", result)
							msg.Cmd("trans", "response", m.Option("parse"))
							m.Copy(msg, "append").Copy(msg, "result")
							return
						}
						result = string(buf)
					default:
						if w, ok := m.Optionv("response").(http.ResponseWriter); ok {
							header := w.Header()
							for k, v := range res.Header {
								header.Add(k, v[0])
							}
							io.Copy(w, res.Body)
							return
						} else {
							buf, e := ioutil.ReadAll(res.Body)
							m.Assert(e)

							m.Append("Content-Type", ct)
							result = string(buf)
						}

					}
					m.Echo("%v", result)
				}
			}},
		"post": &ctx.Command{Name: "post [file fieldname filename]", Help: "post请求",
			Form: map[string]int{"file": 2, "content_type": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				msg := m.Spawn()
				parse := "_"
				if m.Has("file") {
					file, e := os.Open(m.Meta["file"][1])
					m.Assert(e)
					defer file.Close()

					buf := &bytes.Buffer{}
					writer := multipart.NewWriter(buf)
					writer.SetBoundary(fmt.Sprintf("\r\n--%s--\r\n", writer.Boundary()))
					part, e := writer.CreateFormFile(m.Option("file"), filepath.Base(m.Meta["file"][1]))
					m.Assert(e)

					for i := 1; i < len(arg)-1; i += 2 {
						if arg[i] == "parse" {
							parse = arg[i+1]
							continue
						}
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

					io.Copy(part, file)
					writer.Close()
					msg.Optionv("body", buf)
					msg.Option("content_type", writer.FormDataContentType())
					msg.Option("headers", "Content-Length", buf.Len())
				} else if m.Option("content_type") == "json" {
					data := map[string]interface{}{}
					for i := 1; i < len(arg)-1; i += 2 {
						if arg[i] == "parse" {
							parse = arg[i+1]
							continue
						}
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
					m.Log("info", "json %v", string(b))

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

				if parse != "_" {
					arg = append(arg, "parse")
					arg = append(arg, parse)
				}
				msg.Cmd("get", "method", "POST", arg)
				m.Copy(msg, "result").Copy(msg, "append")
			}},
		"brow": &ctx.Command{Name: "brow url", Help: "浏览网页", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Optionv("request").(*http.Request); ok {
				action := false
				m.Travel(func(m *ctx.Message, i int) bool {
					for key, v := range m.Target().Commands {
						method, url := "", ""
						if strings.HasPrefix(v.Name, "get ") {
							method, url = "get", strings.TrimPrefix(v.Name, "get ")
						} else if strings.HasPrefix(v.Name, "post ") {
							method, url = "post", strings.TrimPrefix(v.Name, "post ")
						} else {
							continue
						}

						if len(arg) == 0 {
							m.Add("append", "method", method)
							m.Add("append", "request", url)
						} else if strings.HasPrefix(url, arg[0]) {
							msg := m.Spawn().Cmd(key, arg[1:])
							m.Copy(msg, "append").Copy(msg, "result")
							action = true
							return false
						}
					}
					return true
				})

				if !action {
					msg := m.Spawn().Cmd("get", arg)
					m.Copy(msg, "append").Copy(msg, "result")
				}
				return
			}

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
		"route": &ctx.Command{Name: "route index content [help]", Help: "添加路由响应, index: 路由, context: 响应, help: 说明", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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
					help := "dynamic route"
					if len(arg) > 2 {
						help = arg[2]
					}
					hand := func(m *ctx.Message, c *ctx.Context, key string, a ...string) {
						w := m.Optionv("response").(http.ResponseWriter)
						template.Must(template.New("temp").Parse(arg[1])).Execute(w, m)
					}

					if s, e := os.Stat(arg[1]); e == nil {
						if s.IsDir() {
							mux.Handle(arg[0]+"/", http.StripPrefix(arg[0], http.FileServer(http.Dir(arg[1]))))
						} else if strings.HasSuffix(arg[1], ".shy") {
							hand = func(m *ctx.Message, c *ctx.Context, key string, a ...string) {
								msg := m.Sess("cli").Cmd("source", arg[1])
								m.Copy(msg, "result").Copy(msg, "append")
							}
						} else {
							hand = func(m *ctx.Message, c *ctx.Context, key string, a ...string) {
								w := m.Optionv("response").(http.ResponseWriter)
								template.Must(template.ParseGlob(arg[1])).Execute(w, m)
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
		}},
		"template": &ctx.Command{Name: "template [file [directory]]|[name [content]]", Help: "添加模板, content: 模板内容, directory: 模板目录", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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

				dir := path.Join(m.Confx("template_dir", arg, 1), arg[0])
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
		}},
		"componet": &ctx.Command{Name: "componet [group [order [arg...]]]", Help: "添加组件, group: 组件分组, arg...: 组件参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			switch len(arg) {
			case 0:
				for k, v := range m.Confv("componet").(map[string]interface{}) {
					for i, val := range v.([]interface{}) {
						value := val.(map[string]interface{})
						m.Add("append", "group", k)
						m.Add("append", "order", i)
						m.Add("append", "componet_name", value["componet_name"])
						m.Add("append", "componet_help", value["componet_help"])
						m.Add("append", "componet_ctx", value["componet_ctx"])
						m.Add("append", "componet_cmd", value["componet_cmd"])
					}
				}
				m.Sort("group").Table()
			case 1:
				for i, val := range m.Confv("componet", arg[0]).([]interface{}) {
					value := val.(map[string]interface{})
					m.Add("append", "order", i)
					m.Add("append", "componet_name", value["componet_name"])
					m.Add("append", "componet_help", value["componet_help"])
					m.Add("append", "componet_ctx", value["componet_ctx"])
					m.Add("append", "componet_cmd", value["componet_cmd"])
				}
				m.Table()
			case 2:
				value := m.Confv("componet", []interface{}{arg[0], arg[1]}).(map[string]interface{})
				for k, v := range value {
					m.Add("append", k, v)
				}
				m.Table()
			default:
				if com, ok := m.Confv("componet", []interface{}{arg[0], arg[1]}).(map[string]interface{}); ok {
					for i := 2; i < len(arg)-1; i += 2 {
						com[arg[i]] = arg[i+1]
					}
				} else {
					m.Confv("componet", []interface{}{arg[0], arg[1]}, map[string]interface{}{
						"componet_name": arg[2], "componet_help": arg[3],
						"componet_ctx": m.Confx("componet_context", arg, 4),
						"componet_cmd": m.Confx("componet_command", arg, 5),
					})
					break
				}
			}
		}},
		"session": &ctx.Command{Name: "session [secrete]", Help: "用户登录", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			sessid := m.Option("sessid")
			if sessid == "" || !m.Cmds("aaa.sess", sessid) {
				if !m.Confs("sess_void") && !m.Options("username") {
					return
				}

				// 创建会话
				sessid = m.Cmdx("aaa.sess", "create", "web", "ip", m.Option("remote_ip"))
				http.SetCookie(m.Optionv("response").(http.ResponseWriter), &http.Cookie{Name: "sessid", Value: sessid, Path: "/"})
			}

			if len(arg) > 0 { // 用户认证
				if m.Options("username") && m.Options(arg[0]) && m.Cmds("aaa.sess", sessid, m.Option("username"), arg[0], m.Option(arg[0])) {
					m.Echo(sessid)
				}
				return
			}

			// 用户角色
			m.Cmdy("aaa.sess", m.Option("sessid"), "userrole")
			m.Log("info", "username: %v userrole: %v", m.Meta["username"], m.Meta["userrole"])
		}},
		"bench": &ctx.Command{Name: "bench", Help: "任务列表", Form: map[string]int{"view": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			default_com := "default bench"
			default_cmd := map[string]interface{}{}
			if len(arg) > 0 && arg[0] == "copy" { // 复制工作流
				bench := m.Confv("bench", arg[1]).(map[string]interface{})
				for k, v := range bench["commands"].(map[string]interface{}) {
					default_cmd[k] = v
				}
				arg[0] = "create"
				default_com = bench["comment"].(string)
				m.Option("bench", "")
			}

			if len(arg) > 0 && arg[0] == "create" { // 创建工作流
				create_time := time.Now().Format(m.Conf("time_format"))
				key := m.Option("bench")
				if key == "" {
					key = m.Sess("aaa").Cmd("hash", "bench", m.Option("remote_addr"), "time", "rand").Result(0)
				}

				link := fmt.Sprintf("%s?bench=%s", m.Conf("site"), key)
				if _, ok := m.Confv("bench", key).(map[string]interface{}); !ok {
					m.Log("info", "%s create bench  %s", m.Option("username"), key)
					m.Confv("bench", key, map[string]interface{}{
						"remote_addr": m.Option("remote_addr"),
						"modify_time": create_time,
						"create_time": create_time,
						"creator":     m.Option("username"),
						"share":       "protected",
						"link":        link,
						"comment":     default_com,
						"key":         key,
						"commands":    default_cmd,
					})
				}

				m.Append("key", key)
				m.Echo(link)
				return
			}

			if len(arg) > 0 && arg[0] == "check" { // 检查工作流
				bench := m.Confv("bench", m.Option("bench")).(map[string]interface{})
				if bench["creator"].(string) != arg[1] {
					switch bench["share"].(string) {
					case "private":
						m.Echo("private")
						return
					case "protected":
						m.Echo("protected")
					case "public":
					}
				}
				return
			}

			bench := m.Confv("bench").(map[string]interface{})
			if len(arg) > 0 && arg[0] == "delete" { // 删除工作流
				delete(bench, arg[1])
				arg = arg[2:]
			}

			if len(arg) == 0 { // 查看工作流
				view := "base"
				if m.Has("view") {
					view = m.Option("view")
				}
				for _, v := range bench {
					val := v.(map[string]interface{})
					for _, k := range m.Confv("bench_view", view).([]interface{}) {
						switch v := val[k.(string)].(type) {
						case map[string]interface{}:
							b, _ := json.Marshal(v)
							m.Add("append", "commands", string(b))
						case nil:
							m.Add("append", k.(string), "")
						default:
							m.Add("append", k.(string), v)
						}
					}
				}
				m.Table()
				return
			}

			if len(arg) > 1 { // 编辑工作流
				m.Confv("bench", strings.Split(arg[0], "."), arg[1])
			}

			// 查看工作流
			msg := m.Spawn().Put("option", "_cache", bench).Cmd("trans", "_cache", arg[0])
			m.Copy(msg, "append").Copy(msg, "result")
		}},

		"/render": &ctx.Command{Name: "/render template", Help: "渲染模板, template: 模板名称", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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
					w.Header().Add("Content-Type", "application/json")
				} else {
					w.Header().Add("Content-Type", "text/html")
				}

				// 响应数据
				list := []interface{}{}
				tmpl := web.Template
				if m.Confs("template_debug") {
					tmpl = template.New("render").Funcs(ctx.CGI)
					tmpl.ParseGlob(path.Join(m.Conf("template_dir"), "/*.tmpl"))
					tmpl.ParseGlob(path.Join(m.Conf("template_dir"), m.Conf("template_sub"), "/*.tmpl"))
				}

				// 响应模板
				group, order := m.Option("componet_group", m.Confx("componet_group")), m.Option("componet_name")
				userrole := m.Option("userrole", m.Cmd("web.session").Append("userrole"))

				// 会话检查
				if userrole == "" { // 用户登录
					group, order = m.Option("componet_group", "login"), m.Option("componet_name", "")
				} else if group == "login" { // 登录成功
					return
				} else if !m.Options("bench") || !m.Cmds("aaa.work", m.Option("bench")) { // 创建空间
					m.Append("redirect", fmt.Sprintf("%s?bench=%s", m.Option("index_path"), m.Cmdx("aaa.work", m.Option("sessid"))))
					return
				} else if !m.Options("right", !m.Confs("login_right") || group == "login" ||
					m.Cmds("aaa.work", m.Option("bench"), "right", userrole, "componet", m.Option("componet_group"))) { // 没有权限
					group, order = m.Option("componet_group", "login"), m.Option("componet_name", "")
				} else { //n访问成功
					m.Cmd("aaa.auth", m.Option("bench"), "data", "access_time", m.Time())
					m.Optionv("bench_data", m.Confv("auth", []string{m.Option("bench"), "data"}))
				}

				m.Log("info", "json: %v group: %v order: %v userrole: %v right: %v", accept_json, group, order, userrole, m.Option("right"))

				for _, v := range m.Confv("componet", group).([]interface{}) {
					val := v.(map[string]interface{})
					if order != "" && val["componet_name"].(string) != order {
						continue
					}

					// 查找模块
					context := m.Cap("module")
					if val["componet_ctx"] != nil {
						context = val["componet_ctx"].(string)
					}
					msg := m.Find(context)

					// 添加固定值
					if msg != nil {
						msg.Option("componet_name", val["componet_name"].(string))
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
					}
					pre_run, _ := val["pre_run"].(bool)
					if (!pre_run && order == "") || msg == nil {
						if msg == nil {
							msg = m
						}
						if accept_json {
							list = append(list, msg.Meta)
						} else if val["template"] != nil {
							m.Assert(tmpl.ExecuteTemplate(w, val["template"].(string), msg))
						}
						continue
					}

					// 添加输入值
					if val["inputs"] != nil {
						for _, v := range val["inputs"].([]interface{}) {
							value := v.(map[string]interface{})
							if value["name"] != nil && msg.Option(value["name"].(string)) == "" {
								msg.Add("option", value["name"].(string), value["value"])
							}
						}
					}

					// 添加参数值
					args := []string{}
					if val["componet_cmd"] != nil {
						args = append(args, val["componet_cmd"].(string))
					}
					if val["arguments"] != nil {
						for _, v := range val["arguments"].([]interface{}) {
							switch value := v.(type) {
							case string:
								args = append(args, msg.Parse(value))
							}
						}
					}

					// 执行命令
					if pre_run || m.Cmds("aaa.work", m.Option("bench"), "right", userrole, "componet", m.Option("componet_group"), "command", args[0]) {
						msg.Cmd(args)

						name_alias := "action." + msg.Option("componet_name")
						if msg.Options("componet_name_alias") {
							name_alias = "action." + msg.Option("componet_name_alias")
						}

						msg.Put("option", name_alias, map[string]interface{}{
							"action_time": msg.Time(), "order": m.Option("componet_name_order"), "cmd": args,
						}).Cmd("aaa.auth", m.Option("bench"), "data", "option", name_alias, "modify_time", msg.Time())
					}

					// 添加响应
					if msg.Appends("directory") {
						m.Append("download_file", fmt.Sprintf("/download/%s", msg.Append("directory")))
						return
					}

					if accept_json {
						list = append(list, msg.Meta)
					} else if val["template"] != nil {
						m.Assert(tmpl.ExecuteTemplate(w, val["template"].(string), msg))
					}
				}

				// 生成响应
				if accept_json {
					en := json.NewEncoder(w)
					en.SetIndent("", "  ")
					en.Encode(list)
				}
			}
		}},
		"/upload": &ctx.Command{Name: "/upload", Help: "上传文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			r := m.Optionv("request").(*http.Request)
			f, h, e := r.FormFile("upload")
			m.Assert(e)
			defer f.Close()

			p := path.Join(m.Conf("directory"), m.Option("download_dir"), h.Filename)
			o, e := os.Create(p)
			m.Assert(e)
			defer o.Close()

			io.Copy(o, f)
			m.Log("upload", "file: %s", p)
			m.Append("redirect", m.Option("referer"))
		}},
		"/download/": &ctx.Command{Name: "/download/", Help: "下载文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)
			p := m.Sess("nfs").Cmd("path", strings.TrimPrefix(m.Option("path"), "/download/")).Result(0)
			m.Log("info", "download %s %s", p, m.Sess("aaa").Cmd("hash", "file", p).Result(0))
			http.ServeFile(w, r, p)
		}},
		"/proxy/": &ctx.Command{Name: "/proxy/", Help: "服务代理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Spawn().Cmd("get", strings.TrimPrefix(key, "/proxy/"), arg)
			m.Copy(msg, "append").Copy(msg, "result")
		}},

		"spide": &ctx.Command{Name: "spide", Help: "spide", Form: map[string]int{"fields": 1, "limit": 1, "offset": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if len(arg) == 0 {
				for k, v := range m.Confv("spide").(map[string]interface{}) {
					val := v.(map[string]interface{})
					m.Add("append", "key", k)
					m.Add("append", "limit", val["limit"])
					m.Add("append", "site", val["site"])
				}
				m.Sort("key", "str").Table()
				return
			}

			if len(arg) > 0 && arg[0] == "add" {
				m.Confv("spide", arg[1], map[string]interface{}{
					"site":      arg[2],
					"narg":      strings.Count(arg[2], "%s"),
					"parse":     arg[3],
					"sub_parse": arg[4:],
					"fields":    m.Option("fields"),
					"offset":    m.Option("offset"),
					"limit":     m.Option("limit"),
				})
				return
			}

			spide := m.Confv("spide", arg[0]).(map[string]interface{})
			arg = arg[1:]

			sub_parse := spide["sub_parse"].([]string)

			args := []interface{}{}
			for i := 0; i < spide["narg"].(int); i++ {
				args = append(args, url.QueryEscape(arg[i]))
			}

			msg := m.Spawn().Cmd("get", fmt.Sprintf(spide["site"].(string), args...),
				"parse", spide["parse"], sub_parse)
			arg = arg[spide["narg"].(int):]

			offset := spide["offset"].(string)
			if m.Has("offset") {
				offset = m.Option("offset")
			}
			if len(arg) > 0 {
				offset, arg = arg[0], arg[1:]
			}

			limit := spide["limit"].(string)
			if m.Has("limit") {
				limit = m.Option("limit")
			}
			if len(arg) > 0 {
				limit, arg = arg[0], arg[1:]
			}

			fields := spide["fields"].(string)
			if m.Has("fields") {
				fields = m.Option("fields")
			}
			if len(arg) > 0 {
				fields, arg = arg[0], arg[1:]
			}

			m.Copy(msg, "append").Copy(msg, "result")
			m.Cmd("select", "limit", limit, "offset", offset, "fields", fields)

		}},
		"12306": &ctx.Command{Name: "12306", Help: "12306", Form: map[string]int{"fields": 1, "limit": 1, "offset": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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

			msg := m.Spawn().Cmd("get", fmt.Sprintf("https://kyfw.12306.cn/otn/leftTicket/queryX?leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=ADULT", date, from, to), "parse", "data.result")
			for _, v := range msg.Meta["value"] {
				fields := strings.Split(v, "|")
				m.Add("append", "车次--", fields[3])
				m.Add("append", "出发----", fields[8])
				m.Add("append", "到站----", fields[9])
				m.Add("append", "时长----", fields[10])
				m.Add("append", "二等座", fields[30])
				m.Add("append", "一等座", fields[31])
			}
			m.Table()
		}},
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)
}
