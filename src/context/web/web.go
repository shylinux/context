package web // {{{
// }}}
import ( // {{{
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
)

// }}}

type WEB struct {
	Run    bool
	Master bool

	*http.ServeMux
	*http.Server

	M *ctx.Message
	*ctx.Context
}

type MUX interface {
	Handle(string, http.Handler)
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
	Trans(*ctx.Message, string, ...string)
}

func (web *WEB) Trans(m *ctx.Message, key string, arg ...string) { // {{{
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		m.Set("detail", key)
		for k, v := range r.Form {
			m.Add("option", k)
			m.Meta[k] = v
		}
		for _, v := range r.Cookies() {
			m.Add("option", v.Name, v.Value)
		}

		m.Cmd(arg...)

		header := w.Header()
		for _, k := range m.Meta["append"] {
			ce := &http.Cookie{Name: k, Value: m.Get(k)}
			header.Add("Set-Cookie", ce.String())
		}
		for _, v := range m.Meta["result"] {
			w.Write([]byte(v))
		}
	})
}

// }}}
func (web *WEB) ServeHTTP(w http.ResponseWriter, r *http.Request) { // {{{
	if web.Master {
		log.Println()
		web.M.Log("info", "%s: %v %s %s", web.Name, r.RemoteAddr, r.Method, r.URL)
		defer log.Println()

		if web.M.Conf("logheaders") == "yes" {
			for k, v := range r.Header {
				log.Println(k+":", v[0])
			}
			log.Println()
		}
	}

	r.ParseForm()
	if web.Master {
		if len(r.PostForm) > 0 {
			for k, v := range r.PostForm {
				log.Printf("%s: %s", k, v[0])
			}
			log.Println()
		}
	}

	web.ServeMux.ServeHTTP(w, r)

	if web.Master {
		if web.M.Conf("logheaders") == "yes" {
			for k, v := range w.Header() {
				log.Println(k+":", v[0])
			}
		}
	}
}

// }}}

func (web *WEB) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	web.Configs["logheaders"] = &ctx.Config{Name: "logheaders", Value: "yes", Help: "日志输出请求头"}
	web.Configs["directory"] = &ctx.Config{Name: "directory", Value: "usr", Help: "服务目录"}
	web.Configs["protocol"] = &ctx.Config{Name: "protocol", Value: "http", Help: "服务协议"}
	web.Configs["address"] = &ctx.Config{Name: "address", Value: ":9393", Help: "监听地址"}
	web.Configs["route"] = &ctx.Config{Name: "route", Value: "/" + web.Name + "/", Help: "请求路径"}

	web.ServeMux = http.NewServeMux()
	mux := m.Target.Server.(MUX)
	for k, _ := range web.Commands {
		if k[0] == '/' {
			mux.Trans(m, k)
		}
	}

	return web
}

// }}}
func (web *WEB) Start(m *ctx.Message, arg ...string) bool { // {{{
	web.M = m

	if !web.Run {
		web.Run = true

		if s, ok := m.Target.Context.Server.(MUX); ok {
			if h, ok := m.Target.Server.(http.Handler); ok {
				m.Log("info", "%s: route %s -> %s", web.Context.Context.Name, m.Conf("route"), web.Name)
				s.Handle(m.Conf("route"), http.StripPrefix(path.Dir(m.Conf("route")), h))
			}
		}

		if m.Conf("directory") != "" {
			m.Log("info", "%s: directory [%s]", web.Name, m.Conf("directory"))
			web.Handle("/", http.FileServer(http.Dir(m.Conf("directory"))))
		}

		m.Set("detail", "slaver")
		m.Travel(web.Context, func(m *ctx.Message) bool {
			if m.Target != web.Context {
				m.Target.Start(m)
			}
			return true
		})
	}
	if len(arg) > 0 && arg[0] == "slaver" {
		return true
	}

	if m.Conf("address") == "" {
		return true
	}

	web.Server = &http.Server{Addr: m.Conf("address"), Handler: web}

	m.Log("info", "%s: protocol [%s]", web.Name, m.Conf("protocol"))
	m.Log("info", "%s: address [%s]", web.Name, m.Conf("address"))
	web.Master = true

	if m.Conf("protocol") == "https" {
		m.Log("info", "%s: cert [%s]", web.Name, m.Conf("cert"))
		m.Log("info", "%s: key [%s]", web.Name, m.Conf("key"))
		web.Server.ListenAndServeTLS(m.Conf("cert"), m.Conf("key"))
	} else {
		web.Server.ListenAndServe()
	}

	return true
}

// }}}
func (web *WEB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(WEB)
	s.Context = c
	return s
}

// }}}
func (web *WEB) Close(m *ctx.Message, arg ...string) bool { // {{{
	return true
}

// }}}

var Index = &ctx.Context{Name: "web", Help: "网页服务",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen [address [protocol [directory]]]", Help: "开启网页服务", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			if len(arg) > 0 { // {{{
				m.Conf("address", arg[0])
			}
			if len(arg) > 1 {
				m.Conf("protocol", arg[1])
			}
			if len(arg) > 2 {
				m.Conf("directory", arg[2])
			}
			go m.Target.Start(m)
			return ""
			// }}}
		}},
		"content": &ctx.Command{Name: "content route template", Help: "添加响应", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			mux, ok := m.Target.Server.(MUX) // {{{
			if !ok {
				return ""
			}

			mux.HandleFunc(arg[0], func(w http.ResponseWriter, r *http.Request) {
				if _, e := os.Stat(arg[1]); e == nil {
					template.Must(template.ParseGlob(arg[1])).Execute(w, m.Target)
				} else {
					template.Must(template.New("temp").Parse(arg[1])).Execute(w, m.Target)
				}
			})
			return ""
			// }}}
		}},
		"/hi": &ctx.Command{Name: "/hi", Help: "添加响应", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			m.Echo("hello\n")
			m.Add("append", "hi", "hello")
			m.Add("append", "hi", "hello")
			return "hello\n"
		}},
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)
}
