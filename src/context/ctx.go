package ctx // {{{
// }}}
import ( // {{{
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// }}}

type Cache struct { // {{{
	Name  string
	Value string
	Help  string
	Hand  func(c *Context, x *Cache, arg ...string) string
}

// }}}
type Config struct { // {{{
	Name  string
	Value string
	Help  string
	Hand  func(c *Context, x *Config, arg ...string) string
}

// }}}
type Command struct { // {{{
	Name    string
	Help    string
	Options map[string]string
	Appends map[string]string
	Hand    func(c *Context, m *Message, key string, arg ...string) string
}

// }}}

type Message struct {
	Code int
	Time time.Time

	Meta map[string][]string
	Data map[string]interface{}
	Wait chan bool

	Name string
	*Context
	Index  int
	Target *Context

	Messages []*Message
	Message  *Message
	Root     *Message
}

func (m *Message) Spawn(c *Context, key ...string) *Message { // {{{

	msg := &Message{
		Time:    time.Now(),
		Target:  c,
		Message: m,
		Root:    m.Root,
	}
	msg.Context = m.Target

	if len(key) == 0 {
		return msg
	}

	if m.Messages == nil {
		m.Messages = make([]*Message, 0, 10)
	}
	m.Messages = append(m.Messages, msg)
	msg.Code = m.Capi("nmessage", 1)

	if msg.Sessions == nil {
		msg.Sessions = make(map[string]*Message)
	}
	msg.Sessions[key[0]] = msg
	msg.Name = key[0]

	log.Printf("%d spawn %d: %s.%s->%s.%d", m.Code, msg.Code, msg.Context.Name, msg.Name, msg.Target.Name, msg.Index)
	return msg

}

// }}}

func (m *Message) Add(meta string, key string, value ...string) *Message { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if _, ok := m.Meta[meta]; !ok {
		m.Meta[meta] = make([]string, 0, 3)
	}

	switch meta {
	case "detail", "result":
		m.Meta[meta] = append(m.Meta[meta], key)
		m.Meta[meta] = append(m.Meta[meta], value...)
	case "option", "append":
		if _, ok := m.Meta[key]; !ok {
			m.Meta[meta] = append(m.Meta[meta], key)
			m.Meta[key] = make([]string, 0, 3)
		}
		m.Meta[key] = append(m.Meta[key], value...)
	default:
		panic(errors.New("消息参数错误"))
	}

	return m
}

// }}}
func (m *Message) Put(meta string, key string, value interface{}) *Message { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if m.Data == nil {
		m.Data = make(map[string]interface{})
	}

	switch meta {
	case "option", "append":
		if _, ok := m.Meta[meta]; !ok {
			m.Meta[meta] = make([]string, 0, 3)
		}
		if _, ok := m.Data[key]; !ok {
			m.Meta[meta] = append(m.Meta[meta], key)
		}
		m.Data[key] = value
	default:
		panic(errors.New("消息参数错误"))
	}

	return m
}

// }}}
func (m *Message) Has(key string) bool { // {{{
	if _, ok := m.Meta[key]; ok {
		return true
	}
	if _, ok := m.Data[key]; ok {
		return true
	}
	return false
}

// }}}
func (m *Message) Get(key string) string { // {{{
	if meta, ok := m.Meta[key]; ok {
		return meta[0]
	}
	return ""
}

// }}}
func (m *Message) Echo(str string, arg ...interface{}) *Message { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if _, ok := m.Meta["result"]; !ok {
		m.Meta["result"] = make([]string, 0, 3)
	}

	m.Meta["result"] = append(m.Meta["result"], fmt.Sprintf(str, arg...))
	return m
}

// }}}
func (m *Message) End(s bool) { // {{{
	if m.Wait != nil {
		m.Wait <- s
	}
	m.Wait = nil
}

// }}}

func (m *Message) Cmd(arg ...string) string { // {{{
	if len(arg) > 0 {
		if m.Meta == nil {
			m.Meta = make(map[string][]string)
		}
		m.Meta["detail"] = arg
	}

	return m.Target.Cmd(m, m.Meta["detail"][0], m.Meta["detail"][1:]...)
}

// }}}
func (m *Message) Post(c *Context, arg ...string) bool { // {{{
	if len(arg) > 0 {
		if m.Meta == nil {
			m.Meta = make(map[string][]string)
		}
		m.Meta["detail"] = arg
	}

	if c.Messages == nil {
		panic(m.Target.Name + " 没有开启消息处理")
	}

	c.Messages <- m
	if m.Wait != nil {
		return <-m.Wait
	}
	return true

}

// }}}
func (m *Message) Start(arg ...string) bool { // {{{
	if len(arg) > 0 {
		if m.Meta == nil {
			m.Meta = make(map[string][]string)
		}
		m.Meta["detail"] = arg
	}

	go m.Target.Spawn(m, m.Meta["detail"][0], m.Meta["detail"][1:]...).Begin(m).Start(m)

	return true
}

// }}}

type Server interface { // {{{
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Spawn(c *Context, m *Message, arg ...string) Server
	Exit(m *Message, arg ...string) bool
}

// }}}

type Context struct {
	Name string
	Help string

	Caches   map[string]*Cache
	Configs  map[string]*Config
	Commands map[string]*Command

	Messages chan *Message
	Message  *Message
	Server

	Requests []*Message
	Sessions map[string]*Message

	Index  map[string]*Context
	Groups map[string]*Context

	Contexts map[string]*Context
	Context  *Context
	Root     *Context
}

func (c *Context) Assert(e error) bool { // {{{
	if e != nil {
		log.Println(c.Name, "error:", e)
		if c.Conf("debug") == "on" {
			fmt.Println(c.Name, "error:", e)
		}
		panic(e)
	}
	return true
}

// }}}
func (c *Context) AssertOne(m *Message, safe bool, hand ...func(c *Context, m *Message)) *Context { // {{{
	defer func() {
		if e := recover(); e != nil {
			if c.Conf("debug") == "on" && e != io.EOF {
				fmt.Println(c.Name, "error:", e)
				debug.PrintStack()
			}

			if e == io.EOF {
				return
			}

			if len(hand) > 1 {
				c.AssertOne(m, safe, hand[1:]...)
			} else {
				if !safe {
					log.Println(c.Name, "error:", e)
					panic(e)
				}
			}
		}
		// Pulse.Wait <- true
	}()

	if len(hand) > 0 {
		hand[0](c, m)
	}

	return c
}

// }}}

func (c *Context) Register(s *Context, x Server) *Context { // {{{
	if c.Contexts == nil {
		c.Contexts = make(map[string]*Context)
	}
	if x, ok := c.Contexts[s.Name]; ok {
		panic(errors.New(c.Name + " 上下文已存在" + x.Name))
	}

	c.Contexts[s.Name] = s
	s.Root = c.Root
	s.Context = c
	s.Server = x

	log.Printf("%s sub(%d): %s", c.Name, Index.Capi("ncontext", 1), s.Name)
	return s
}

// }}}
func (c *Context) Begin(m *Message) *Context { // {{{
	for k, x := range c.Configs {
		if x.Hand != nil {
			c.Conf(k, x.Value)
		}
	}

	if c.Server != nil {
		if m != nil {
			c.Server.Begin(m, m.Meta["detail"]...)
		} else {
			c.Server.Begin(m)
		}
	}

	return c
}

// }}}
func (c *Context) Start(m *Message) bool { // {{{
	if _, ok := c.Caches["status"]; !ok {
		c.Caches["status"] = &Cache{Name: "服务状态", Value: "stop", Help: "服务状态，start:正在运行，stop:未在运行"}
	}

	if c.Cap("status") != "start" && c.Server != nil {
		c.AssertOne(m, true, func(c *Context, m *Message) {
			c.Cap("status", "start")
			defer c.Cap("status", "stop")

			log.Printf("%d start(%d): %s %s", m.Code, Index.Capi("nserver", 1), c.Name, c.Help)
			defer Index.Capi("nserver", -1)
			defer log.Printf("%d stop(%s): %s %s", m.Code, Index.Cap("nserver"), c.Name, c.Help)

			c.Requests = []*Message{m}
			c.Server.Start(m, m.Meta["detail"]...)
		})
	}

	return true
}

// }}}
func (c *Context) Spawn(m *Message, key string, arg ...string) *Context { // {{{
	s := &Context{Name: key, Help: c.Help}
	m.Target = s
	if c.Server != nil {
		c.Register(s, c.Server.Spawn(s, m, arg...)).Begin(m)
	} else {
		c.Register(s, nil).Begin(m)
	}
	return s
}

// }}}

func (c *Context) Deal(pre func(m *Message, arg ...string) bool, post func(m *Message, arg ...string) bool) (live bool) { // {{{

	if c.Messages == nil {
		c.Messages = make(chan *Message, c.Confi("MessageQueueSize"))
	}
	m := <-c.Messages
	defer m.End(true)

	if len(m.Meta["detail"]) == 0 {
		return true
	}

	if pre != nil && !pre(m, m.Meta["detail"]...) {
		return false
	}

	arg := m.Meta["detail"]
	c.AssertOne(m, true, func(c *Context, m *Message) {
		m.Add("result", c.Cmd(m, arg[0], arg[1:]...))

	}, func(c *Context, m *Message) {
		m.Add("result", m.Cmd())

	}, func(c *Context, m *Message) {
		log.Printf("system command(%s->%s): %v", m.Context.Name, m.Target.Name, arg)
		cmd := exec.Command(arg[0], arg[1:]...)
		v, e := cmd.CombinedOutput()
		if e != nil {
			m.Echo("%s\n", e)
		} else {
			m.Echo(string(v))
		}
	})

	if post != nil && !post(m, m.Meta["result"]...) {
		return false
	}

	return true
}

// }}}
func (c *Context) Exit(m *Message, arg ...string) { // {{{
	if m.Target == c {
		for _, v := range c.Sessions {
			if v.Name != "" {
				v.Name = ""
				v.Target.Exit(v, arg...)
				log.Println(c.Name, c.Help, "exit: session", v.Code, v.Target.Name, v.Target.Help)
			}
		}

		if c.Server != nil {
			c.Server.Exit(m, arg...)
			log.Println(c.Name, c.Help, "exit: self", m.Code)
		}

		for _, v := range c.Requests {
			if v.Index != -1 {
				v.Context.Exit(v, arg...)
				log.Println(c.Name, c.Help, "exit: resource", v.Code, v.Context.Name, v.Context.Help)
				v.Index = -1
			}
		}
	} else if m.Context == c {
		if m.Name != "" {
			m.Target.Exit(m, arg...)
			log.Println(c.Name, c.Help, "exit: session", m.Code, m.Target.Name, m.Target.Help)
			m.Name = ""
		}
	}

}

// }}}

func (c *Context) Add(arg ...string) { // {{{
	switch arg[0] {
	case "context":
		if len(arg) != 4 {
			panic(errors.New("参数错误"))
		}
		if c.Index == nil {
			panic(errors.New("索引表不存在"))
		}
		if v, ok := c.Index[arg[1]]; ok {
			panic(errors.New(v.Name + "上下文已存在"))
		}

		if c.Groups == nil {
			c.Groups = make(map[string]*Context)
		}
		c.Groups[arg[1]] = &Context{Name: arg[2], Help: arg[3], Index: c.Index}
		c.Index[arg[1]] = c.Groups[arg[1]]
		log.Println(c.Name, "add context:", arg[1:])
	case "command":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Groups[arg[1]]; ok {
			if v.Commands == nil {
				v.Commands = make(map[string]*Command)
			}
			if x, ok := v.Commands[arg[2]]; ok {
				panic(errors.New(x.Name + "命令已存在"))
			}
			if x, ok := c.Commands[arg[2]]; ok {
				log.Println(v.Name, "add command:", arg[2])
				v.Commands[arg[2]] = x
			} else {
				panic(errors.New(arg[2] + "命令不存在"))
			}
		} else {
			panic(errors.New(arg[1] + "上下文不存在"))
		}
	case "config":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Groups[arg[1]]; ok {
			if v.Configs == nil {
				v.Configs = make(map[string]*Config)
			}
			if x, ok := v.Configs[arg[2]]; ok {
				panic(errors.New(x.Name + "配置项已存在"))
			}
			if x, ok := c.Configs[arg[2]]; ok {
				log.Println(v.Name, "add config:", arg[2])
				v.Configs[arg[2]] = x
			} else {
				panic(errors.New(arg[2] + "配置项不存在"))
			}
		} else {
			panic(errors.New(arg[1] + "上下文不存在"))
		}
	case "cache":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Groups[arg[1]]; ok {
			if v.Caches == nil {
				v.Caches = make(map[string]*Cache)
			}
			if x, ok := v.Caches[arg[2]]; ok {
				panic(errors.New(x.Name + "缓存项已存在"))
			}
			if x, ok := c.Caches[arg[2]]; ok {
				log.Println(v.Name, "add cache:", arg[2])
				v.Caches[arg[2]] = x
			} else {
				panic(errors.New(arg[2] + "缓存项不存在"))
			}
		} else {
			panic(errors.New(arg[1] + "上下文不存在"))
		}
	}
}

// }}}
func (c *Context) Del(arg ...string) { // {{{
	cs := make([]*Context, 0, 5)

	switch arg[0] {
	case "context":
		if len(arg) != 2 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Groups[arg[1]]; ok {
			cs = append(cs, v)
			delete(c.Index, arg[1])
			delete(c.Groups, arg[1])
			log.Println(c.Name, "del context:", arg[1])
		}
		for i := 0; i < len(cs); i++ {
			for k, v := range cs[i].Groups {
				cs = append(cs, v)
				delete(c.Index, k)
				log.Println(c.Name, "del context:", k)
			}
		}
	case "command":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Groups[arg[1]]; ok {
			cs = append(cs, v)
			delete(v.Commands, arg[2])
			log.Println(v.Name, "del command:", arg[2])
		}
		for i := 0; i < len(cs); i++ {
			for _, v := range cs[i].Groups {
				cs = append(cs, v)
				delete(v.Commands, arg[2])
				log.Println(v.Name, "del command:", arg[2])
			}
		}
	case "config":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Groups[arg[1]]; ok {
			cs = append(cs, v)
			delete(v.Configs, arg[2])
			log.Println(v.Name, "del config:", arg[2])
		}
		for i := 0; i < len(cs); i++ {
			for _, v := range cs[i].Groups {
				cs = append(cs, v)
				delete(v.Configs, arg[2])
				log.Println(v.Name, "del config:", arg[2])
			}
		}
	case "cache":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Groups[arg[1]]; ok {
			cs = append(cs, v)
			delete(v.Caches, arg[2])
			log.Println(v.Name, "del cache:", arg[2])
		}
		for i := 0; i < len(cs); i++ {
			for _, v := range cs[i].Groups {
				cs = append(cs, v)
				delete(v.Caches, arg[2])
				log.Println(v.Name, "del cache:", arg[2])
			}
		}
	}
}

// }}}

func (c *Context) Travel(hand func(s *Context) bool) { // {{{
	cs := []*Context{c}
	for i := 0; i < len(cs); i++ {
		for _, v := range cs[i].Contexts {
			cs = append(cs, v)
		}

		if !hand(cs[i]) {
			return
		}
	}
}

// }}}
func (c *Context) BackTrace(hand func(s *Context) bool) { // {{{
	for cs := c; cs != nil; cs = cs.Context {
		if !hand(cs) {
			return
		}
	}
}

// }}}
func (c *Context) Search(name string) []*Context { // {{{
	cs := make([]*Context, 0, 3)

	c.Travel(func(s *Context) bool {
		if strings.Contains(s.Name, name) || strings.Contains(s.Help, name) {
			cs = append(cs, s)
			log.Println(c.Name, "search:", s.Name, "[match]", name)
		} else {
			log.Println(c.Name, "search:", s.Name)
		}
		return true
	})
	return cs
}

// }}}
func (c *Context) Find(name string) (s *Context) { // {{{
	ns := strings.Split(name, ".")
	cs := c.Contexts
	for _, v := range ns {
		if x, ok := cs[v]; ok {
			cs = x.Contexts
			s = x
		} else {
			panic(errors.New(c.Name + " not find: " + name))
		}
	}
	return
}

// }}}

func (c *Context) Cmd(m *Message, key string, arg ...string) string { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	m.Meta["detail"] = nil
	m.Add("detail", key, arg...)

	for s := c; s != nil; s = s.Context {
		if x, ok := s.Commands[key]; ok {
			log.Printf("%s cmd(%s->%s): %v", c.Name, m.Context.Name, m.Target.Name, m.Meta["detail"])

			if x.Hand == nil {
				panic(errors.New(fmt.Sprintf(key + "没有权限")))
			}

			for _, v := range m.Meta["option"] {
				if _, ok := x.Options[v]; !ok {
					panic(errors.New(fmt.Sprintf(v + "未知参数")))
				}
			}

			value := x.Hand(c, m, key, arg...)

			for _, v := range m.Meta["append"] {
				if _, ok := x.Appends[v]; !ok {
					panic(errors.New(fmt.Sprintf(v + "未知参数")))
				}
			}

			return value
		}
	}

	panic(errors.New(fmt.Sprintf(key + "命令项不存在")))
}

// }}}
func (c *Context) Conf(key string, arg ...string) string { // {{{
	for s := c; s != nil; s = s.Context {
		if x, ok := s.Configs[key]; ok {
			switch len(arg) {
			case 0:
				if x.Hand != nil {
					return x.Hand(c, x)
				}
				return x.Value
			case 1:
				log.Println(c.Name, "conf:", key, arg[0])
				x.Value = arg[0]
				if x.Hand != nil {
					x.Hand(c, x, x.Value)
				}
				return x.Value
			case 3:
				log.Println(c.Name, "conf:", key, arg)
				if s == c {
					panic(errors.New(key + "配置项已存在"))
				}
				if c.Configs == nil {
					c.Configs = make(map[string]*Config)
				}
				c.Configs[key] = &Config{Name: arg[0], Value: arg[1], Help: arg[2], Hand: x.Hand}
				return arg[1]
			default:
				panic(errors.New(key + "配置项参数错误"))
			}
		}
	}

	if len(arg) == 3 {
		log.Println(c.Name, "conf:", key, arg)
		if c.Configs == nil {
			c.Configs = make(map[string]*Config)
		}
		c.Configs[key] = &Config{Name: arg[0], Value: arg[1], Help: arg[2]}
		return arg[1]
	}

	panic(errors.New(key + "配置项不存在"))
}

// }}}
func (c *Context) Confi(key string, arg ...int) int { // {{{
	if len(arg) > 0 {
		n, e := strconv.Atoi(c.Conf(key, fmt.Sprintf("%d", arg[0])))
		c.Assert(e)
		return n
	}

	n, e := strconv.Atoi(c.Conf(key))
	c.Assert(e)
	return n
}

// }}}
func (c *Context) Cap(key string, arg ...string) string { // {{{
	for s := c; s != nil; s = s.Context {
		if x, ok := s.Caches[key]; ok {
			switch len(arg) {
			case 0:
				if x.Hand != nil {
					x.Value = x.Hand(c, x)
				}
				return x.Value
			case 1:
				if x.Hand != nil {
					x.Value = x.Hand(c, x, arg[0])
				} else {
					x.Value = arg[0]
				}
				return x.Value
			case 3:
				log.Println(c.Name, "cap:", key, arg)
				if s == c {
					panic(errors.New(key + "缓存项已存在"))
				}
				if c.Caches == nil {
					c.Caches = make(map[string]*Cache)
				}
				c.Caches[key] = &Cache{arg[0], arg[1], arg[2], x.Hand}
				return arg[1]
			default:
				panic(errors.New(key + "缓存项参数错误"))
			}
		}
	}
	if len(arg) == 3 {
		log.Println(c.Name, "cap:", key, arg)
		if c.Caches == nil {
			c.Caches = make(map[string]*Cache)
		}
		c.Caches[key] = &Cache{arg[0], arg[1], arg[2], nil}
		return arg[1]
	}

	panic(errors.New(key + "缓存项不存在"))
}

// }}}
func (c *Context) Capi(key string, arg ...int) int { // {{{
	n, e := strconv.Atoi(c.Cap(key))
	c.Assert(e)
	if len(arg) > 0 {
		n += arg[0]
		c.Cap(key, strconv.Itoa(n))
	}
	return n
}

// }}}

var Pulse = &Message{Code: 1, Time: time.Now(), Name: "ctx", Index: 0}

var Index = &Context{Name: "ctx", Help: "根模块",
	Caches: map[string]*Cache{
		"nserver":  &Cache{Name: "服务数量", Value: "0", Help: "显示已经启动运行模块的数量"},
		"ncontext": &Cache{Name: "模块数量", Value: "1", Help: "显示功能树已经注册模块的数量"},
		"nmessage": &Cache{Name: "消息数量", Value: "1", Help: "显示模块启动时所创建消息的数量"},
	},
	Configs: map[string]*Config{
		"start":   &Config{Name: "启动模块", Value: "cli", Help: "启动时自动运行的模块"},
		"init.sh": &Config{Name: "启动脚本", Value: "etc/init.sh", Help: "模块启动时自动运行的脚本"},
		"bench.log": &Config{Name: "日志文件", Value: "var/bench.log", Help: "模块日志输出的文件", Hand: func(c *Context, x *Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				if e := os.MkdirAll(path.Dir(arg[0]), os.ModePerm); e == nil {
					if l, e := os.Create(x.Value); e == nil {
						log.SetOutput(l)
					}
				}
			}
			return x.Value
			// }}}
		}},
		"root": &Config{Name: "工作目录", Value: ".", Help: "所有模块的当前目录", Hand: func(c *Context, x *Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				if !path.IsAbs(x.Value) {
					wd, e := os.Getwd()
					c.Assert(e)
					x.Value = path.Join(wd, x.Value)
				}

				if e := os.MkdirAll(x.Value, os.ModePerm); e != nil {
					fmt.Println(e)
					os.Exit(1)
				}
				if e := os.Chdir(x.Value); e != nil {
					fmt.Println(e)
					os.Exit(1)
				}
			}

			return x.Value
			// }}}
		}},

		"ContextRequestSize": &Config{Name: "请求队列长度", Value: "10", Help: "每个模块可以被其它模块引用的的数量"},
		"ContextSessionSize": &Config{Name: "会话队列长度", Value: "10", Help: "每个模块可以启动其它模块的数量"},
		"MessageQueueSize":   &Config{Name: "消息队列长度", Value: "10", Help: "每个模块接收消息的队列长度"},

		"debug": &Config{Name: "调试模式(off/no)", Value: "off", Help: "是否打印错误信息，off:不打印，on:打印)"},
		"cert":  &Config{Name: "证书文件", Value: "etc/cert.pem", Help: "证书文件"},
		"key":   &Config{Name: "私钥文件", Value: "etc/key.pem", Help: "私钥文件"},
	},
	Commands: map[string]*Command{
		"show": &Command{Name: "输出简单的信息", Help: "建立远程连接", Hand: func(c *Context, m *Message, key string, arg ...string) string {
			m.Echo("nserver: %s\n", c.Cap("nserver")) // {{{
			m.Echo("ncontext: %s\n", c.Cap("ncontext"))
			m.Echo("nmessage: %s\n", c.Cap("nmessage"))
			return ""
			// }}}
		}},
	},
	Requests: []*Message{Pulse},
	Sessions: map[string]*Message{"root": Pulse},
}

func init() {
	if len(os.Args) > 1 {
		Index.Conf("start", os.Args[1])
	}

	if len(os.Args) > 2 {
		Index.Conf("init.sh", os.Args[2])
	}

	if len(os.Args) > 3 {
		Index.Conf("bench.log", os.Args[3])
	} else {
		Index.Conf("bench.log", Index.Conf("bench.log"))
	}

	if len(os.Args) > 4 {
		Index.Conf("root", os.Args[4])
	}

	log.Println("\n\n\n")
}

func Start() {
	Pulse.Root = Pulse
	Pulse.Target = Index
	Pulse.Context = Index
	Pulse.Wait = make(chan bool, 10)

	Index.Travel(func(s *Context) bool {
		s.Root = Index
		s.Begin(nil)
		return true
	})

	n := 0
	for _, s := range Index.Contexts {
		if ok, _ := regexp.MatchString(Index.Conf("start"), s.Name); ok {
			n++
			go s.Start(Pulse.Spawn(s, s.Name).Put("option", "io", os.Stdout))
		}
	}

	for {
		<-Pulse.Wait
		n--
		if n <= 0 && Index.Capi("nserver", 0) == 0 {
			return
		}
	}
}
