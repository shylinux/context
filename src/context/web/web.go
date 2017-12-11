package web // {{{
// }}}
import ( // {{{
	"context"

	"html/template"
	"net/http"

	"bufio"
	"log"
	"os"
	"path"
	"strings"
)

// }}}

type MUX interface {
	Handle(string, http.Handler)
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
	Trans(*ctx.Message, string, func(*ctx.Message, *ctx.Context, string, ...string) string)
}

type WEB struct {
	*http.ServeMux
	*http.Server

	*ctx.Message
	*ctx.Context
}

func (web *WEB) Trans(m *ctx.Message, key string, hand func(*ctx.Message, *ctx.Context, string, ...string) string) { // {{{
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		msg := m.Spawn(m.Target)
		msg.Set("detail", key)
		for k, v := range r.Form {
			msg.Add("option", k)
			msg.Meta[k] = v
		}
		for _, v := range r.Cookies() {
			msg.Add("option", v.Name, v.Value)
		}

		msg.Log("cmd", nil, "%s %v", key, msg.Meta["option"])
		msg.Put("option", "request", r)
		msg.Put("option", "response", w)

		ret := hand(msg, msg.Target, key)
		if ret != "" {
			msg.Echo(ret)
		}

		header := w.Header()
		for _, k := range msg.Meta["append"] {
			ce := &http.Cookie{Name: k, Value: msg.Get(k)}
			header.Add("Set-Cookie", ce.String())
		}
		for _, v := range msg.Meta["result"] {
			w.Write([]byte(v))
		}
	})
}

// }}}
func (web *WEB) ServeHTTP(w http.ResponseWriter, r *http.Request) { // {{{
	if web.Message != nil {
		log.Println()
		web.Log("cmd", nil, "%v %s %s", r.RemoteAddr, r.Method, r.URL)
		defer log.Println()

		if web.Cap("logheaders") == "yes" {
			for k, v := range r.Header {
				log.Println(k+":", v[0])
			}
			log.Println()
		}

		r.ParseForm()
		if len(r.PostForm) > 0 {
			for k, v := range r.PostForm {
				log.Printf("%s: %s", k, v[0])
			}
			log.Println()
		}
	}

	web.ServeMux.ServeHTTP(w, r)

	if web.Message != nil {
		if web.Cap("logheaders") == "yes" {
			for k, v := range w.Header() {
				log.Println(k+":", v[0])
			}
		}
	}
}

// }}}

func (web *WEB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"directory": &ctx.Cache{Name: "directory", Value: "usr", Help: "服务目录"},
	}
	c.Configs = map[string]*ctx.Config{}

	s := new(WEB)
	s.Context = c
	return s
}

// }}}
func (web *WEB) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	if len(arg) > 0 {
		m.Cap("directory", arg[0])
	}

	web.Caches["route"] = &ctx.Cache{Name: "route", Value: "/" + web.Context.Name + "/", Help: "请求路径"}
	web.Caches["register"] = &ctx.Cache{Name: "已初始化(yes/no)", Value: "no", Help: "模块是否已注册"}
	web.Caches["master"] = &ctx.Cache{Name: "master(yes/no)", Value: "no", Help: "日志输出请求头"}
	m.Cap("stream", m.Cap("route")+" -> "+m.Cap("directory"))

	web.ServeMux = http.NewServeMux()
	if mux, ok := m.Target.Server.(MUX); ok {
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

	web.Message = m
	m.Cap("master", "yes")

	m.Travel(m.Target, func(m *ctx.Message) bool {
		if h, ok := m.Target.Server.(http.Handler); ok && m.Cap("register") == "no" {
			m.Cap("register", "yes")

			p, i := m.Target, 0
			m.BackTrace(func(m *ctx.Message) bool {
				p = m.Target
				if i++; i == 2 {
					return false
				}
				return true
			})

			if s, ok := p.Server.(MUX); ok {
				m.Log("info", p, "route %s -> %s", m.Cap("route"), m.Target.Name)
				s.Handle(m.Cap("route"), http.StripPrefix(path.Dir(m.Cap("route")), h))
			}

			if s, ok := m.Target.Server.(MUX); ok && m.Cap("directory") != "" {
				m.Log("info", nil, "dir / -> [%s]", m.Cap("directory"))
				s.Handle("/", http.FileServer(http.Dir(m.Cap("directory"))))
			}
		}
		return true
	})

	web.Caches["address"] = &ctx.Cache{Name: "address", Value: ":9393", Help: "监听地址"}
	web.Caches["protocol"] = &ctx.Cache{Name: "protocol", Value: "http", Help: "服务协议"}
	if len(arg) > 1 {
		m.Cap("address", arg[1])
	}
	if len(arg) > 2 {
		m.Cap("protocol", arg[2])
	}

	m.Cap("stream", m.Cap("address"))
	m.Log("info", nil, "address [%s]", m.Cap("address"))
	m.Log("info", nil, "protocol [%s]", m.Cap("protocol"))
	web.Server = &http.Server{Addr: m.Cap("address"), Handler: web}

	web.Caches["logheaders"] = &ctx.Cache{Name: "日志输出报文头(yes/no)", Value: "yes", Help: "日志输出请求头"}

	if m.Cap("protocol") == "https" {
		m.Log("info", nil, "key [%s]", m.Cap("key"))
		m.Log("info", nil, "cert [%s]", m.Cap("cert"))
		web.Server.ListenAndServeTLS(m.Cap("cert"), m.Cap("key"))
	} else {
		web.Server.ListenAndServe()
	}

	return true
}

// }}}
func (web *WEB) Close(m *ctx.Message, arg ...string) bool { // {{{
	return false
}

// }}}

var Index = &ctx.Context{Name: "web", Help: "网页服务",
	Caches: map[string]*ctx.Cache{
		"directory": &ctx.Cache{Name: "directory", Value: "usr", Help: "服务目录"},
	},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen [directory [address [protocol]]]", Help: "开启网页服务", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			m.Meta["detail"] = arg // {{{
			m.Target.Start(m)
			return ""
			// }}}
		}},
		"route": &ctx.Command{Name: "route [directory|template|script] route string", Help: "添加响应", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			mux, ok := m.Target.Server.(MUX) // {{{
			m.Assert(ok, "模块类型错误")
			m.Assert(len(arg) == 3, "缺少参数")

			switch arg[0] {
			case "directory":
				mux.Handle(arg[1], http.FileServer(http.Dir(arg[2])))
			case "template":
				mux.Trans(m, arg[1], func(m *ctx.Message, c *ctx.Context, key string, a ...string) string { // {{{
					w := m.Data["response"].(http.ResponseWriter)

					if _, e := os.Stat(arg[2]); e == nil {
						template.Must(template.ParseGlob(arg[2])).Execute(w, m)
					} else {
						template.Must(template.New("temp").Parse(arg[2])).Execute(w, m)
					}

					return ""
				})
				// }}}
			case "script":
				cli := m.Find("cli", true) // {{{
				lex := m.Find("lex", true)
				mux.Trans(m, arg[1], func(m *ctx.Message, c *ctx.Context, key string, a ...string) string {
					f, e := os.Open(arg[2])
					line, bio := "", bufio.NewReader(f)

					if e != nil {
						line = arg[2]
					}

					for {
						if line = strings.TrimSpace(line); line != "" {
							lex.Cmd("split", line, "void")
							cli.Wait = make(chan bool)
							cli.Cmd(lex.Meta["result"]...)
							m.Meta["result"] = cli.Meta["result"]
						}

						if line, e = bio.ReadString('\n'); e != nil {
							break
						}
					}

					return ""
				})
				// }}}
			}

			return ""
			// }}}
		}},
		"/hi": &ctx.Command{Name: "/hi", Help: "添加响应", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			m.Add("append", "hi", "hello")
			return "hello"
		}},
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)
}
