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

	"bytes"
	"mime/multipart"
	"path/filepath"

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

	client *http.Client
	cookie map[string]*http.Cookie

	list     map[string][]string
	list_key []string

	*ctx.Message
	*ctx.Context
}

func (web *WEB) generate(m *ctx.Message, uri string, arg ...string) string { // {{{
	add, e := url.Parse(uri)
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

	args := []string{}
	for i := 0; i < len(arg)-1; i += 2 {
		args = append(args, arg[i]+"="+arg[i+1])
	}
	p := strings.Join(args, "&")

	if add.RawQuery != "" {
		adds = append(adds, "?")
		adds = append(adds, add.RawQuery)
		if p != "" {
			adds = append(adds, "&")
			adds = append(adds, p)
		}
	} else if m.Confs("query") {
		adds = append(adds, "?")
		adds = append(adds, m.Conf("query"))
		if p != "" {
			adds = append(adds, "&")
			adds = append(adds, p)
		}
	} else {
		if p != "" {
			adds = append(adds, "?")
			adds = append(adds, p)
		}
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
	web.Log("fuck", nil, "why")
	if web.Message != nil {
		log.Println()
		web.Log("cmd", nil, "%v %s %s", r.RemoteAddr, r.Method, r.URL)

		if web.Conf("logheaders") == "yes" {
			for k, v := range r.Header {
				web.Log("info", nil, "%s: %v", k, v)
			}
			web.Log("info", nil, "")
		}

		if r.ParseForm(); len(r.PostForm) > 0 {
			for k, v := range r.PostForm {
				web.Log("info", nil, "%s: %v", k, v)
			}
			web.Log("info", nil, "")
		}
	}

	web.ServeMux.ServeHTTP(w, r)

	if web.Message != nil && web.Conf("logheaders") == "yes" {
		for k, v := range w.Header() {
			web.Log("info", nil, "%s: %v", k, v)
		}
		web.Log("info", nil, "")
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

	web.list = map[string][]string{}

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
		"output":   &ctx.Config{Name: "output", Value: "stdout", Help: "响应输出"},
		"editor":   &ctx.Config{Name: "editor", Value: "vim", Help: "响应编辑器"},
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
		"cookie": &ctx.Command{Name: "cookie add|del arg...", Help: "访问URL", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			web, ok := m.Target().Server.(*WEB) // {{{
			m.Assert(ok)

			switch len(arg) {
			case 0:
				for k, v := range web.cookie {
					m.Echo("%s: %v\n", k, v.Value)
				}
			case 1:
				if v, ok := web.cookie[arg[0]]; ok {
					m.Echo("%s", v.Value)
				}
			default:
				if web.cookie == nil {
					web.cookie = make(map[string]*http.Cookie)
				}
				if v, ok := web.cookie[arg[0]]; ok {
					v.Value = arg[1]
				} else {
					web.cookie[arg[0]] = &http.Cookie{Name: arg[0], Value: arg[1]}
				}
			}
			// }}}
		}},
		"get": &ctx.Command{Name: "get [method GET|POST] [file filename] arg...", Help: "访问URL",
			Formats: map[string]int{"method": 1, "file": 2},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				web, ok := m.Target().Server.(*WEB) // {{{
				m.Assert(ok)

				if web.client == nil {
					web.client = &http.Client{}
				}

				method := "GET"
				if m.Options("method") {
					method = m.Option("method")
				}

				uri := web.generate(m, arg[0], arg[1:]...)
				m.Log("info", nil, "GET %s", uri)
				m.Echo("%s\n", uri)

				var body io.Reader
				index := strings.Index(uri, "?")
				contenttype := ""

				switch method {
				case "POST":
					if m.Options("file") {
						file, e := os.Open(m.Meta["file"][1])
						m.Assert(e)
						defer file.Close()

						buf := &bytes.Buffer{}
						writer := multipart.NewWriter(buf)

						part, e := writer.CreateFormFile(m.Option("file"), filepath.Base(m.Meta["file"][1]))
						m.Assert(e)

						io.Copy(part, file)
						for i := 0; i < len(arg)-1; i += 2 {
							writer.WriteField(arg[0], arg[1])
						}

						contenttype = writer.FormDataContentType()
						body = buf
						writer.Close()
					} else if index > 0 {
						contenttype = "application/x-www-form-urlencoded"
						body = strings.NewReader(uri[index+1:])
						uri = uri[:index]
					}
				}

				req, e := http.NewRequest(method, uri, body)
				m.Assert(e)

				if len(contenttype) > 0 {
					req.Header.Set("Content-Type", contenttype)
				}

				for _, v := range web.cookie {
					req.AddCookie(v)
				}

				res, e := web.client.Do(req)
				m.Assert(e)

				if web.cookie == nil {
					web.cookie = make(map[string]*http.Cookie)
				}
				for _, v := range res.Cookies() {
					web.cookie[v.Name] = v
				}

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
							m.Echo("write to %s\n", name)
						}
						return
					}
				}

				buf, e := ioutil.ReadAll(res.Body)
				m.Assert(e)
				m.Echo(string(buf))
				// }}}
			}},
		"list": &ctx.Command{Name: "list [set|add|del [url]]", Help: "查看、访问、添加url", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			web, ok := m.Target().Server.(*WEB) // {{{
			m.Assert(ok)

			switch len(arg) {
			case 0:
				for _, k := range web.list_key {
					if v, ok := web.list[k]; ok {
						m.Echo("%s: %s\n", k, v)
					}
				}
			case 1:
				msg := m.Spawn(m.Target()).Cmd("get", web.list[arg[0]])
				m.Copy(msg, "result")
			default:
				switch arg[0] {
				case "add":
					web.list[m.Cap("count")] = arg[1:]
					web.list_key = append(web.list_key, m.Cap("count"))
					m.Capi("count", 1)
				case "del":
					delete(web.list, arg[1])
				case "set":
					web.list[arg[1]] = arg[2:]
				default:
					list := []string{}
					j := 1
					for _, v := range web.list[arg[0]] {
						if v == "_" && j < len(arg) {
							list = append(list, arg[j])
							j++
						} else {
							list = append(list, v)
						}
					}
					for ; j < len(arg); j++ {
						list = append(list, arg[j])
					}

					msg := m.Spawn(m.Target()).Cmd("get", list)
					m.Copy(msg, "result")
				}
			} // }}}
		}},
		"/demo": &ctx.Command{Name: "/demo", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			r := m.Data["request"].(*http.Request)
			file, _, e := r.FormFile("file")
			m.Assert(e)
			buf, e := ioutil.ReadAll(file)
			m.Assert(e)
			m.Echo(string(buf))
			m.Add("append", "hi", "hello")
		}},
		"temp": &ctx.Command{Name: "temp", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Spawn(m.Target())
			question := []string{}
			for i := 1; i < 21; i++ {
				question = append(question, fmt.Sprintf("{\"type\":\"1001\",\"title\":{\"text\":\"第%d题\"}}", i))
			}
			qs := "[" + strings.Join(question, ",") + "]"

			msg.Cmd("get", "method", "POST", "evaluating_add/", "questions", qs)
			m.Add("append", "hi", "hello")
		}},
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)
}
