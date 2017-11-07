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
	hands map[string]bool
	Mux   *http.ServeMux
	*ctx.Context
}

func (web *WEB) Handle(p string, h http.Handler) { // {{{
	if web.hands == nil {
		web.hands = make(map[string]bool)
	}

	if _, ok := web.hands[p]; ok {
		panic(fmt.Sprintln(web.Name, "handle exits", p))
	}

	web.Mux.Handle(path.Clean(p)+"/", http.StripPrefix(path.Clean(p), h))
	log.Println(web.Name, "mux:", path.Clean(p)+"/")
}

// }}}
func (web *WEB) HandleFunc(p string, h func(http.ResponseWriter, *http.Request)) { // {{{
	if web.hands == nil {
		web.hands = make(map[string]bool)
	}

	if _, ok := web.hands[p]; ok {
		panic(fmt.Sprintln(web.Name, "handle exits", p))
	}

	web.Mux.HandleFunc(p, h)
	log.Println(web.Name, "mux:", p)
}

// }}}
func (web *WEB) Begin(m *ctx.Message) ctx.Server { // {{{
	return web
}

// }}}
func (web *WEB) Start(m *ctx.Message) bool { // {{{

	web.Mux.Handle("/", http.FileServer(http.Dir(web.Conf("directory"))))

	log.Println(web.Name, "listen:", web.Conf("address"))
	log.Println(web.Name, "https:", web.Conf("https"))

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
		}
	}()

	if web.Conf("https") == "yes" {
		log.Println(web.Name, "cert:", web.Conf("cert"))
		log.Println(web.Name, "key:", web.Conf("key"))
		http.ListenAndServeTLS(web.Conf("address"), web.Conf("cert"), web.Conf("key"), web.Mux)
	} else {
		http.ListenAndServe(web.Conf("address"), web.Mux)
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
		"address":   &ctx.Config{Name: "listen", Value: ":9090", Help: "监听地址"},
		"directory": &ctx.Config{Name: "directory", Value: "./", Help: "服务目录"},
		"https":     &ctx.Config{Name: "https", Value: "yes", Help: "开启安全连接"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{"listen address directory", "设置监听地址和目录", func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
			default:
				c.Conf("address", arg[0])
				c.Conf("directory", arg[1])
				go c.Start(m)
			}
			return ""
			// }}}
		}},
	},
}

func init() { // {{{
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)

	web.Mux = http.NewServeMux()
}

// }}}
