package web // {{{
// }}}
import ( // {{{
	"contexts"

	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"bufio"
	"fmt"
	"io"
	"log"
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

	list map[string]string

	*ctx.Message
	*ctx.Context
}

func (web *WEB) generate(m *ctx.Message, arg string) string { // {{{
	add, e := url.Parse(arg)
	m.Assert(e)

	adds := []string{}

	if add.Scheme != "" {
		adds = append(adds, add.Scheme)
	} else if m.Confs("protocol") {
		adds = append(adds, m.Conf("protocol"))
	}
	adds = append(adds, "://")

	if add.Host != "" {
		adds = append(adds, add.Host)
	} else if m.Confs("hostname") {
		adds = append(adds, m.Conf("hostname"))
		if m.Confs("port") {
			adds = append(adds, ":")
			adds = append(adds, m.Conf("port"))
		}
	}

	dir, file := path.Split(add.EscapedPath())
	if path.IsAbs(dir) {
		adds = append(adds, dir)
		adds = append(adds, file)
	} else {
		adds = append(adds, m.Conf("dir"))
		if dir == "" && file == "" {
			adds = append(adds, m.Conf("file"))
		} else {
			adds = append(adds, dir)
			adds = append(adds, file)
		}
	}

	if add.RawQuery != "" {
		adds = append(adds, "?")
		adds = append(adds, add.RawQuery)
	} else if m.Confs("query") {
		adds = append(adds, "?")
		adds = append(adds, m.Conf("query"))
	}

	return strings.Join(adds, "")
}

// }}}

func (web *WEB) AppendJson(msg *ctx.Message) string { // {{{
	meta := map[string][]string{}
	meta["result"] = msg.Meta["result"]
	meta["append"] = msg.Meta["append"]
	for _, v := range msg.Meta["append"] {
		meta[v] = msg.Meta[v]
	}

	b, e := json.Marshal(meta)
	msg.Assert(e)
	return string(b)

	result := []string{"{"}
	for i, k := range msg.Meta["append"] {
		result = append(result, fmt.Sprintf("\"%s\": [", k))
		for j, v := range msg.Meta[k] {
			result = append(result, fmt.Sprintf("\"%s\"", url.QueryEscape(v)))
			if j < len(msg.Meta[k])-1 {
				result = append(result, ",")
			}
		}
		result = append(result, "]")
		if i < len(msg.Meta["append"])-1 {
			result = append(result, ", ")
		}
	}
	result = append(result, "}")

	return strings.Join(result, "")
}

// }}}
func (web *WEB) Trans(m *ctx.Message, key string, hand func(*ctx.Message, *ctx.Context, string, ...string)) { // {{{
	web.HandleFunc(key, func(w http.ResponseWriter, r *http.Request) {
		msg := m.Spawn(m.Target()).Set("detail", key)

		for k, v := range r.Form {
			msg.Add("option", k, v...)
		}
		for _, v := range r.Cookies() {
			msg.Add("option", v.Name, v.Value)
		}
		msg.Log("cmd", nil, "%s [] %v", key, msg.Meta["option"])

		msg.Put("option", "request", r).Put("option", "response", w)
		if hand(msg, msg.Target(), key); len(msg.Meta["append"]) > 0 {
			msg.Set("result", web.AppendJson(msg))
		}

		for _, v := range msg.Meta["result"] {
			msg.Log("info", nil, "%s", v)
			w.Write([]byte(v))
		}
	})
}

// }}}
func (web *WEB) ServeHTTP(w http.ResponseWriter, r *http.Request) { // {{{
	if web.Message != nil {
		log.Println()
		web.Log("cmd", nil, "%v %s %s", r.RemoteAddr, r.Method, r.URL)

		if web.Conf("logheaders") == "yes" {
			for k, v := range r.Header {
				log.Printf("%s: %v", k, v)
			}
			log.Println()
		}

		if r.ParseForm(); len(r.PostForm) > 0 {
			for k, v := range r.PostForm {
				log.Printf("%s: %v", k, v)
			}
			log.Println()
		}
	}

	web.ServeMux.ServeHTTP(w, r)

	if web.Message != nil && web.Conf("logheaders") == "yes" {
		for k, v := range w.Header() {
			log.Printf("%s: %v", k, v)
		}
		log.Println()
	}
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
func (web *WEB) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	web.Context.Master(nil)
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

	web.list = map[string]string{}

	return web
}

// }}}
func (web *WEB) Start(m *ctx.Message, arg ...string) bool { // {{{
	if len(arg) > 0 {
		m.Cap("directory", arg[0])
	}

	m.Travel(m.Target(), func(m *ctx.Message) bool {
		if h, ok := m.Target().Server.(http.Handler); ok && m.Cap("register") == "no" {
			m.Cap("register", "yes")

			p, i := m.Target(), 0
			m.BackTrace(func(m *ctx.Message) bool {
				p = m.Target()
				if i++; i == 2 {
					return false
				}
				return true
			})

			if s, ok := p.Server.(MUX); ok {
				m.Log("info", p, "route %s -> %s", m.Cap("route"), m.Target().Name)
				s.Handle(m.Cap("route"), http.StripPrefix(path.Dir(m.Cap("route")), h))
			}

			if s, ok := m.Target().Server.(MUX); ok && m.Cap("directory") != "" {
				m.Log("info", nil, "dir / -> [%s]", m.Cap("directory"))
				s.Handle("/", http.FileServer(http.Dir(m.Cap("directory"))))
			}
		}
		return true
	})

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
	m.Log("info", nil, "address [%s]", m.Cap("address"))
	m.Log("info", nil, "protocol [%s]", m.Cap("protocol"))
	web.Server = &http.Server{Addr: m.Cap("address"), Handler: web}

	web.Configs["logheaders"] = &ctx.Config{Name: "日志输出报文头(yes/no)", Value: "yes", Help: "日志输出报文头"}

	if web.Message = m; m.Cap("protocol") == "https" {
		web.Caches["cert"] = &ctx.Cache{Name: "服务证书", Value: m.Conf("cert"), Help: "服务证书"}
		web.Caches["key"] = &ctx.Cache{Name: "服务密钥", Value: m.Conf("key"), Help: "服务密钥"}
		m.Log("info", nil, "cert [%s]", m.Cap("cert"))
		m.Log("info", nil, "key [%s]", m.Cap("key"))

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
		"count": &ctx.Cache{Name: "count", Value: "0", Help: "主机协议"},
	},
	Configs: map[string]*ctx.Config{
		"protocol": &ctx.Config{Name: "protocol", Value: "", Help: "主机协议"},
		"hostname": &ctx.Config{Name: "hostname", Value: "", Help: "主机地址"},
		"port":     &ctx.Config{Name: "port", Value: "", Help: "主机端口"},
		"dir":      &ctx.Config{Name: "dir", Value: "/", Help: "主机路由"},
		"file":     &ctx.Config{Name: "file", Value: "", Help: "主机文件"},
		"query":    &ctx.Config{Name: "query", Value: "", Help: "主机参数"},
		"output":   &ctx.Config{Name: "output", Value: "", Help: "响应输出"},
		"editor":   &ctx.Config{Name: "editor", Value: "", Help: "响应编辑器"},
	},
	Commands: map[string]*ctx.Command{
		"serve": &ctx.Command{Name: "serve [directory [address [protocol]]]", Help: "开启应用服务", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Set("detail", arg...).Target().Start(m)
		}},
		"route": &ctx.Command{Name: "route directory|template|script route content", Help: "添加应用内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			mux, ok := m.Target().Server.(MUX) // {{{
			m.Assert(ok, "模块类型错误")
			m.Assert(len(arg) == 3, "缺少参数")

			switch arg[0] {
			case "directory":
				mux.Handle(arg[1]+"/", http.StripPrefix(arg[1], http.FileServer(http.Dir(arg[2]))))
			case "template":
				mux.Trans(m, arg[1], func(m *ctx.Message, c *ctx.Context, key string, a ...string) {
					w := m.Data["response"].(http.ResponseWriter)

					if _, e := os.Stat(arg[2]); e == nil {
						template.Must(template.ParseGlob(arg[2])).Execute(w, m)
					} else {
						template.Must(template.New("temp").Parse(arg[2])).Execute(w, m)
					}

				})
			case "script":
				cli := m.Find("cli")
				lex := m.Find("lex")
				mux.Trans(m, arg[1], func(m *ctx.Message, c *ctx.Context, key string, a ...string) {
					f, e := os.Open(arg[2])
					line, bio := "", bufio.NewReader(f)
					if e != nil {
						line = arg[2]
					}

					for {
						if line = strings.TrimSpace(line); line != "" {
							lex.Cmd("split", line, "void")
							cli.Wait = make(chan bool)
							cli.Cmd(lex.Meta["result"])
							m.Meta["result"] = cli.Meta["result"]
						}

						if line, e = bio.ReadString('\n'); e != nil {
							break
						}
					}
				})
			} // }}}
		}},
		"get": &ctx.Command{Name: "get url", Help: "访问URL", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			web, ok := m.Target().Server.(*WEB) // {{{
			m.Assert(ok)

			uri := web.generate(m, arg[0])
			m.Log("info", nil, "GET %s", uri)

			res, e := http.Get(uri)
			m.Assert(e)

			for k, v := range res.Header {
				m.Log("info", nil, "%s: %v", k, v)
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
						m.Echo("write to %s", name)
					}
					return
				}
			}

			buf, e := ioutil.ReadAll(res.Body)
			m.Assert(e)
			m.Echo(string(buf))
			// }}}
		}},
		"list": &ctx.Command{Name: "list [index|add|del [url]]", Help: "查看、访问、添加url", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			web, ok := m.Target().Server.(*WEB) // {{{
			m.Assert(ok)

			switch len(arg) {
			case 0:
				for k, v := range web.list {
					m.Echo("%s: %s\n", k, v)
				}
			case 1:
				msg := m.Spawn(m.Target()).Cmd("get", web.list[arg[0]])
				m.Copy(msg, "result")
			case 2:
				switch arg[0] {
				case "add":
					m.Capi("count", 1)
					web.list[m.Cap("count")] = arg[1]
				case "del":
					delete(web.list, arg[1])
				default:
					web.list[arg[0]] = arg[1]
				}
			} // }}}
		}},
		"/demo": &ctx.Command{Name: "/demo", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Add("append", "hi", "hello")
		}},
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)
}
