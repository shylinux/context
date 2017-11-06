package web // {{{
// }}}
import ( // {{{
	"context"
	"log"
	"net/http"
)

// }}}

type WEB struct {
	*ctx.Context
}

func (web *WEB) Begin(m *ctx.Message) ctx.Server { // {{{
	return web
}

// }}}
func (web *WEB) Start(m *ctx.Message) bool { // {{{
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(web.Conf("directory"))))
	for _, v := range m.Meta["detail"][2:] {
		if h, ok := m.Data[v].(func(http.ResponseWriter, *http.Request)); ok {
			mux.HandleFunc(v, h)
		}
	}

	s := &http.Server{Addr: web.Conf("address"), Handler: mux}
	s.ListenAndServe()

	log.Println(s.ListenAndServe())
	return true
}

// }}}
func (web *WEB) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"status": &ctx.Cache{Name: "status", Value: "stop", Help: "服务状态"},
	}
	c.Configs = map[string]*ctx.Config{
		"address":   &ctx.Config{Name: "listen", Value: arg[0], Help: "监听地址"},
		"directory": &ctx.Config{Name: "directory", Value: arg[1], Help: "服务目录"},
	}
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

var Index = &ctx.Context{Name: "web", Help: "网页服务", // {{{
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{"listen address directory", "设置监听地址和目录", func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
			default:
				m.Add("detail", "/hi")
				m.Put("/hi", func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("hello context world!"))
				})

				m.Start(m.Meta["detail"][1:]...)
			}
			return ""
			// }}}
		}},
	},
}

// }}}

func init() { // {{{
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)
}

// }}}
