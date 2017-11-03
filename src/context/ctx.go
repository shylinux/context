package ctx // {{{
// }}}
import ( // {{{
	"errors"
	"fmt"
	"log"
	"os"
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
	Hand  func(c *Context, arg string) string
}

// }}}
type Config struct { // {{{
	Name  string
	Value string
	Help  string
	Hand  func(c *Context, arg string) string
	Spawn bool
}

// }}}
type Command struct { // {{{
	Name string
	Help string
	Hand func(c *Context, m *Message, arg ...string) string
}

// }}}
type Message struct { // {{{
	Code int
	Time time.Time

	Meta map[string][]string
	Data map[string]interface{}
	Wait chan bool

	*Context
	Name   string
	Target *Context
	Index  int

	Messages []*Message
	Message  *Message
	Root     *Message
}

func (m *Message) Spawn(c *Context, key string, index int) *Message { // {{{
	msg := &Message{
		Code:    m.Capi("nmessage", 1),
		Time:    time.Now(),
		Target:  c,
		Name:    key,
		Message: m,
		Root:    m.Root,
	}

	msg.Context = m.Target
	if msg.Session == nil {
		msg.Session = make(map[string]*Message)
	}
	msg.Session[key] = msg

	if m.Messages == nil {
		m.Messages = make([]*Message, 0, 10)
	}
	m.Messages = append(m.Messages, msg)
	log.Printf("%d spawn %d: %s.%s->%s.%d", m.Code, msg.Code, msg.Context.Name, msg.Name, msg.Target.Name, msg.Index)
	return msg

}

// }}}
func (m *Message) Add(key string, value ...string) *Message { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if _, ok := m.Meta[key]; !ok {
		m.Meta[key] = make([]string, 0, 3)
	}

	m.Meta[key] = append(m.Meta[key], value...)
	return m
}

// }}}
func (m *Message) Put(key string, value interface{}) *Message { // {{{
	if m.Data == nil {
		m.Data = make(map[string]interface{})
	}

	m.Data[key] = value
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

	s := fmt.Sprintf(str, arg...)
	m.Meta["result"] = append(m.Meta["result"], s)
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
func (m *Message) Cmd() string { // {{{
	return m.Target.Cmd(m, m.Meta["detail"]...)
}

// }}}
func (m *Message) Post(c *Context) bool { // {{{
	if c.Messages == nil {
		c.Messages = make(chan *Message, c.Confi("MessageQueueSize"))
	}

	c.Messages <- m
	if m.Wait != nil {
		return <-m.Wait
	}
	return true

}

// }}}
func (m *Message) Start(arg ...string) bool { // {{{
	s := m.Target.Spawn(arg...).Begin()
	m.Target = s
	go s.Start(m)
	return true
}

// }}}
// }}}
type Server interface { // {{{
	Begin() Server
	Start(m *Message) bool
	Spawn(c *Context, arg ...string) Server
}

// }}}
type Context struct { // {{{
	Name string
	Help string

	Caches   map[string]*Cache
	Configs  map[string]*Config
	Commands map[string]*Command

	Messages chan *Message
	Message  *Message
	Server

	Resource []*Message
	Session  map[string]*Message

	Index   map[string]*Context
	Shadows map[string]*Context

	Contexts map[string]*Context
	Context  *Context
	Root     *Context
}

func (c *Context) Check(e error) bool { // {{{
	if e != nil {
		log.Println(c.Name, "error:", e)
		if c.Conf("debug") == "on" {
			debug.PrintStack()
		}
		panic(e)
	}
	return true
}

// }}}
func (c *Context) Register(s *Context, self Server) bool { // {{{
	if c.Contexts == nil {
		c.Contexts = make(map[string]*Context)
	}

	if x, ok := c.Contexts[s.Name]; ok {
		panic(errors.New(c.Name + " 上下文已存在" + x.Name))
	}

	c.Contexts[s.Name] = s
	s.Context = c
	s.Root = c.Root
	s.Server = self

	log.Println(c.Name, "register:", s.Name)
	return true
}

// }}}
func (c *Context) Init(arg ...string) { // {{{
	if c.Root != nil {
		return
	}

	root := c
	for root.Context != nil {
		root = root.Context
	}

	cs := []*Context{root}

	for i := 0; i < len(cs); i++ {
		cs[i].Root = root
		cs[i].Begin()

		for _, v := range cs[i].Contexts {
			cs = append(cs, v)
		}
	}

	Pulse.Root = Pulse
	Pulse.Target = Index
	Pulse.Context = Index

	for _, s := range root.Contexts {
		if ok, _ := regexp.MatchString(root.Conf("start"), s.Name); ok {
			go s.Start(Pulse.Spawn(s, s.Name, 0).Put("detail", os.Stdout))
		}
	}

	Pulse.Wait = make(chan bool)
	for {
		<-Pulse.Wait
		if Index.Capi("nserver", 0) == 0 {
			return
		}
	}
}

// }}}
func (c *Context) Find(name []string) (s *Context) { // {{{
	log.Println(c.Name, "find:", name)

	cs := c.Contexts
	for _, v := range name {
		if x, ok := cs[v]; ok {
			s = x
			cs = x.Contexts
			continue
		}
		panic(errors.New("not find: " + v))
	}
	return
}

// }}}
func (c *Context) Search(name string) []*Context { // {{{
	ps := make([]*Context, 0, 3)

	cs := []*Context{c}
	for i := 0; i < len(cs); i++ {
		for _, v := range cs[i].Contexts {
			cs = append(cs, v)
		}

		if strings.Contains(cs[i].Name, name) || strings.Contains(cs[i].Help, name) {
			ps = append(ps, cs[i])
			log.Println(c.Name, "search:", i, cs[i].Name, "[match]")
		} else {
			log.Println(c.Name, "search:", i, cs[i].Name)
		}
	}

	return ps
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

		if c.Shadows == nil {
			c.Shadows = make(map[string]*Context)
		}
		c.Shadows[arg[1]] = &Context{Name: arg[2], Help: arg[3], Index: c.Index}
		c.Index[arg[1]] = c.Shadows[arg[1]]
		log.Println(c.Name, "add context:", arg[1:])
	case "command":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Shadows[arg[1]]; ok {
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

		if v, ok := c.Shadows[arg[1]]; ok {
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

		if v, ok := c.Shadows[arg[1]]; ok {
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

		if v, ok := c.Shadows[arg[1]]; ok {
			cs = append(cs, v)
			delete(c.Index, arg[1])
			delete(c.Shadows, arg[1])
			log.Println(c.Name, "del context:", arg[1])
		}
		for i := 0; i < len(cs); i++ {
			for k, v := range cs[i].Shadows {
				cs = append(cs, v)
				delete(c.Index, k)
				log.Println(c.Name, "del context:", k)
			}
		}
	case "command":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Shadows[arg[1]]; ok {
			cs = append(cs, v)
			delete(v.Commands, arg[2])
			log.Println(v.Name, "del command:", arg[2])
		}
		for i := 0; i < len(cs); i++ {
			for _, v := range cs[i].Shadows {
				cs = append(cs, v)
				delete(v.Commands, arg[2])
				log.Println(v.Name, "del command:", arg[2])
			}
		}
	case "config":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Shadows[arg[1]]; ok {
			cs = append(cs, v)
			delete(v.Configs, arg[2])
			log.Println(v.Name, "del config:", arg[2])
		}
		for i := 0; i < len(cs); i++ {
			for _, v := range cs[i].Shadows {
				cs = append(cs, v)
				delete(v.Configs, arg[2])
				log.Println(v.Name, "del config:", arg[2])
			}
		}
	case "cache":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := c.Shadows[arg[1]]; ok {
			cs = append(cs, v)
			delete(v.Caches, arg[2])
			log.Println(v.Name, "del cache:", arg[2])
		}
		for i := 0; i < len(cs); i++ {
			for _, v := range cs[i].Shadows {
				cs = append(cs, v)
				delete(v.Caches, arg[2])
				log.Println(v.Name, "del cache:", arg[2])
			}
		}
	}
}

// }}}

func (c *Context) Begin() *Context { // {{{
	c.Root.Capi("ncontext", 1)
	for _, v := range c.Configs {
		if v.Hand != nil {
			v.Hand(c, v.Value)
		}
	}

	if c.Server != nil {
		c.Server.Begin()
	}
	return c
}

// }}}
func (c *Context) Start(m *Message) bool { // {{{
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
		}
		Pulse.Wait <- true
	}()

	if c.Server != nil && c.Cap("status") != "start" {

		c.Cap("status", "status", "start", "服务状态")
		defer c.Cap("status", "stop")

		c.Root.Capi("nserver", 1)
		defer c.Root.Capi("nserver", -1)

		c.Resource = []*Message{m}
		log.Println(m.Code, "start:", c.Name)
		c.Server.Start(m)
		log.Println(m.Code, "stop:", c.Name)
	}

	return true
}

// }}}
func (c *Context) Spawn(arg ...string) *Context { // {{{
	s := &Context{Name: arg[0], Help: c.Help}
	c.Register(s, c.Server.Spawn(s, arg...))
	s.Begin()
	return s
}

// }}}

func (c *Context) Post(m *Message) bool { // {{{
	if c.Messages == nil {
		c.Messages = make(chan *Message, c.Confi("MessageQueueSize"))
	}

	c.Messages <- m
	if m.Wait != nil {
		return <-m.Wait
	}
	return true
}

// }}}
func (c *Context) Get() *Message { // {{{
	if c.Messages == nil {
		c.Messages = make(chan *Message, c.Confi("MessageQueueSize"))
	}

	return <-c.Messages
}

// }}}

func (c *Context) Cmd(m *Message, arg ...string) string { // {{{
	if x, ok := c.Commands[arg[0]]; ok {
		return x.Hand(c, m, arg...)
	}

	if c.Context != nil {
		return c.Context.Cmd(m, arg...)
	}

	panic(errors.New(fmt.Sprintf(arg[0] + "命令项不存在")))
}

// }}}
func (c *Context) Conf(arg ...string) string { // {{{
	switch len(arg) {
	case 1:
		if v, ok := c.Configs[arg[0]]; ok {
			if v.Hand != nil {
				return v.Hand(c, v.Value)
			}
			return v.Value
		}

		if c.Context != nil {
			return c.Context.Conf(arg...)
		}
	case 2:
		if v, ok := c.Configs[arg[0]]; ok {
			v.Value = arg[1]
			if v.Hand != nil {
				return v.Hand(c, v.Value)
			}
			log.Println(c.Name, "config:", arg)
			return v.Value
		}

		if c.Context != nil {
			return c.Context.Conf(arg...)
		}
	case 4:
		if v, ok := c.Configs[arg[0]]; ok {
			panic(errors.New(v.Name + "配置项已存在"))
		}

		c.Configs[arg[0]] = &Config{Name: arg[1], Value: arg[2], Help: arg[3]}
		log.Println(c.Name, "config:", arg)
		return arg[2]
	default:
		panic(errors.New(arg[0] + "配置项参数错误"))
	}

	panic(errors.New(arg[0] + "配置项不存在"))
}

// }}}
func (c *Context) Confi(arg ...string) int { // {{{
	n, e := strconv.Atoi(c.Conf(arg...))
	c.Check(e)
	return n
}

// }}}
func (c *Context) Cap(arg ...string) string { // {{{
	switch len(arg) {
	case 1:
		if v, ok := c.Caches[arg[0]]; ok {
			if v.Hand != nil {
				v.Value = v.Hand(c, v.Value)
			}
			// log.Println(c.Name, "cache:", arg, v.Value)
			return v.Value
		}

		if c.Context != nil {
			return c.Context.Cap(arg...)
		}
	case 2:
		if v, ok := c.Caches[arg[0]]; ok {
			v.Value = arg[1]
			if v.Hand != nil {
				v.Value = v.Hand(c, v.Value)
			}
			// log.Println(c.Name, "cache:", arg)
			return v.Value
		}

		if c.Context != nil {
			return c.Context.Cap(arg...)
		}
	case 4:
		// if v, ok := c.Caches[arg[0]]; ok {
		// 	panic(errors.New(v.Name + "缓存项已存在"))
		// }
		//
		c.Caches[arg[0]] = &Cache{arg[1], arg[2], arg[3], nil}
		// log.Println(c.Name, "cache:", arg)
		return arg[2]
	default:
		panic(errors.New(arg[0] + "缓存项参数错误"))
	}

	panic(errors.New(arg[0] + "缓存项不存在"))
}

// }}}
func (c *Context) Capi(key string, value int) int { // {{{
	n, e := strconv.Atoi(c.Cap(key))
	c.Check(e)
	c.Cap(key, strconv.Itoa(n+value))
	return n + value
}

// }}}
// }}}

var Index = &Context{Name: "ctx", Help: "根上下文",
	Caches: map[string]*Cache{
		"status":   &Cache{Name: "status", Value: "stop", Help: "服务状态"},
		"nserver":  &Cache{Name: "nserver", Value: "0", Help: "服务数量"},
		"ncontext": &Cache{Name: "ncontext", Value: "0", Help: "上下文数量"},
		"nmessage": &Cache{Name: "nmessage", Value: "0", Help: "消息发送数量"},
	},
	Configs: map[string]*Config{
		"开场白": &Config{Name: "开场白", Value: "你好，上下文", Help: "开场白"},
		"结束语": &Config{Name: "结束语", Value: "再见，上下文", Help: "结束语"},

		"MessageQueueSize": &Config{Name: "MessageQueueSize", Value: "10", Help: "默认消息队列长度"},
		"MessageListSize":  &Config{Name: "MessageListSize", Value: "10", Help: "默认消息列表长度"},

		"cert": &Config{Name: "cert", Value: "etc/cert.pem", Help: "证书文件"},
		"key":  &Config{Name: "key", Value: "etc/key.pem", Help: "私钥文件"},

		"debug":   &Config{Name: "debug", Value: "on", Help: "调试模式"},
		"start":   &Config{Name: "start", Value: "cli", Help: "默认启动模块"},
		"init.sh": &Config{Name: "init.sh", Value: "etc/init.sh", Help: "默认启动脚本"},
		"bench.log": &Config{Name: "bench.log", Value: "var/bench.log", Help: "默认日志文件", Hand: func(c *Context, arg string) string {
			l, e := os.Create(arg) // {{{
			c.Check(e)
			log.SetOutput(l)
			return arg
			// }}}
		}},
	},
	Commands: map[string]*Command{},
	Session:  map[string]*Message{"root": Pulse},
	Resource: []*Message{Pulse},
}

var Pulse = &Message{Code: 0, Time: time.Now(), Index: 0, Name: "root"}

func init() {
	if len(os.Args) > 1 {
		Index.Conf("bench.log", os.Args[1])
	} else {
		Index.Conf("bench.log", Index.Conf("bench.log"))
	}

	if len(os.Args) > 2 {
		Index.Conf("init.sh", os.Args[2])
	}

	if len(os.Args) > 3 {
		Index.Conf("start", os.Args[3])
	}
}

func Start() {
	Index.Init()
}
