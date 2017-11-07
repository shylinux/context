package web // {{{
// }}}
import ( // {{{
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
)

// }}}

type WEB struct {
	Server *http.Server
	Mux    *http.ServeMux
	hands  map[string]bool

	run bool

	*ctx.Context
}

func (web *WEB) Handle(p string, h http.Handler) { // {{{
	if web.hands == nil {
		web.hands = make(map[string]bool)
	}

	p = path.Clean(p)
	if _, ok := web.hands[p]; ok {
		panic(fmt.Sprintln(web.Name, "handle exits", p))
	}
	web.hands[p] = true
	if p == "/" {
		panic(fmt.Sprintln(web.Name, "handle exits", p))
	}

	web.Mux.Handle(p+"/", http.StripPrefix(p, h))
}

// }}}
func (web *WEB) HandleFunc(p string, h func(http.ResponseWriter, *http.Request)) { // {{{
	if web.hands == nil {
		web.hands = make(map[string]bool)
	}

	p = path.Clean(p)
	if _, ok := web.hands[p]; ok {
		panic(fmt.Sprintln(web.Name, "handle exits", p))
	}
	web.hands[p] = true
	if p == "/" {
		panic(fmt.Sprintln(web.Name, "handle exits", p))
	}

	web.Mux.HandleFunc(p, h)
	log.Println(web.Name, "hand:", p)
}

// }}}
func (web *WEB) ServeHTTP(w http.ResponseWriter, r *http.Request) { // {{{
	log.Println()
	log.Println(web.Name, r.RemoteAddr, r.Method, r.URL.Path)
	web.Mux.ServeHTTP(w, r)
}

// }}}

func (web *WEB) Begin(m *ctx.Message) ctx.Server { // {{{
	return web
}

// }}}
func (web *WEB) Start(m *ctx.Message) bool { // {{{

	if !web.run {
		if web.Conf("directory") != "" {
			web.Mux.Handle("/", http.FileServer(http.Dir(web.Conf("directory"))))
			log.Println(web.Name, "directory:", web.Conf("directory"))
		}

		for _, v := range web.Contexts {
			if s, ok := v.Server.(*WEB); ok && s.Mux != nil {
				log.Println(web.Name, "route:", s.Conf("route"), "->", s.Name)
				web.Handle(s.Conf("route"), s.Mux)
				s.Start(m)
			}
		}
	}
	web.run = true

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
		"directory": &ctx.Config{Name: "directory", Value: "./", Help: "服务目录"},
		"protocol":  &ctx.Config{Name: "protocol", Value: "http", Help: "服务协议"},
		"address":   &ctx.Config{Name: "address", Value: ":9393", Help: "监听地址"},
		"route":     &ctx.Config{Name: "route", Value: "/", Help: "请求路径"},
		"default":   &ctx.Config{Name: "default", Value: "hello web world", Help: "默认响应体"},
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

	web.Mux = http.NewServeMux()
	web.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(web.Conf("default")))
	})
}
