package web // {{{
// }}}
import ( // {{{
	"context"
	"log"
	"net/http"
	"path"
)

// }}}

type Request struct {
	R *http.Request
	W http.ResponseWriter
	X interface{}
	*WEB
}

type WEB struct {
	Run bool

	*Request

	*http.ServeMux
	*http.Server

	*ctx.Context
}

func (web *WEB) ServeHTTP(w http.ResponseWriter, r *http.Request) { // {{{
	log.Println()
	log.Println(web.Name, r.RemoteAddr, r.Method, r.URL)
	defer log.Println()

	if web.Conf("logheaders") == "yes" {
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

	web.ServeMux.ServeHTTP(w, r)

	if web.Conf("logheaders") == "yes" {
		for k, v := range w.Header() {
			log.Println(k+":", v[0])
		}
	}
}

// }}}

func (web *WEB) Begin(m *ctx.Message) ctx.Server { // {{{
	return web
}

// }}}
func (web *WEB) Start(m *ctx.Message) bool { // {{{

	if !web.Run {
		if web.Conf("directory") != "" {
			web.Handle("/", http.FileServer(http.Dir(web.Conf("directory"))))
			log.Println(web.Name, "directory:", web.Conf("directory"))
		}

		for _, v := range web.Contexts {
			if s, ok := v.Server.(http.Handler); ok {
				log.Println(web.Name, "route:", v.Conf("route"), "->", v.Name)
				web.Handle(v.Conf("route"), http.StripPrefix(path.Dir(v.Conf("route")), s))
				v.Start(m)
			}
		}
	}
	web.Run = true

	if m.Target != web.Context {
		return true
	}

	web.Server = &http.Server{Addr: web.Conf("address"), Handler: web}
	log.Println(web.Name, "protocol:", web.Conf("protocol"))
	log.Println(web.Name, "address:", web.Conf("address"))

	if web.Conf("protocol") == "https" {
		log.Println(web.Name, "cert:", web.Conf("cert"))
		log.Println(web.Name, "key:", web.Conf("key"))
		web.Server.ListenAndServeTLS(web.Conf("cert"), web.Conf("key"))
	} else {
		web.Server.ListenAndServe()
	}

	return true
}

// }}}
func (web *WEB) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}
	c.Commands = map[string]*ctx.Command{}

	s := new(WEB)
	s.Context = c
	return s
}

// }}}
func (web *WEB) Exit(m *ctx.Message, arg ...string) bool { // {{{
	return true
}

// }}}

var Index = &ctx.Context{Name: "web", Help: "网页服务",
	Caches: map[string]*ctx.Cache{
		"status": &ctx.Cache{Name: "status", Value: "stop", Help: "服务状态"},
	},
	Configs: map[string]*ctx.Config{
		"logheaders": &ctx.Config{Name: "logheaders", Value: "yes", Help: "日志输出请求头"},
		"directory":  &ctx.Config{Name: "directory", Value: "./", Help: "服务目录"},
		"protocol":   &ctx.Config{Name: "protocol", Value: "https", Help: "服务协议"},
		"address":    &ctx.Config{Name: "address", Value: ":443", Help: "监听地址"},
		"route":      &ctx.Config{Name: "route", Value: "/", Help: "请求路径"},
		"default":    &ctx.Config{Name: "default", Value: "hello web world", Help: "默认响应体"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen [route [address protocol [directory]]]", Help: "开启网页服务", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			s, ok := m.Target.Server.(*WEB) // {{{
			if !ok {
				return ""
			}

			if len(arg) > 0 {
				s.Conf("route", arg[0])
			}
			if len(arg) > 2 {
				s.Conf("address", arg[1])
				s.Conf("protocol", arg[2])
			}
			if len(arg) > 3 {
				s.Conf("directory", arg[3])
			}
			go s.Start(m)

			return ""
			// }}}
		}},
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)

	web.ServeMux = http.NewServeMux()
	web.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(web.Conf("default")))
	})
}
