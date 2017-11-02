package ctx // {{{
// }}}
import ( // {{{
	"errors"
	"fmt"
	"log"
	"os"
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

	Name string
	*Context

	Target *Context
	Index  int
}

func (m *Message) Add(key string, value ...string) string { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if _, ok := m.Meta[key]; !ok {
		m.Meta[key] = make([]string, 0, 3)
	}

	m.Meta[key] = append(m.Meta[key], value...)
	return value[0]
}

// }}}
func (m *Message) Put(key string, value interface{}) interface{} { // {{{
	if m.Data == nil {
		m.Data = make(map[string]interface{})
	}

	m.Data[key] = value
	return value
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
func (m *Message) Echo(str string, arg ...interface{}) string { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if _, ok := m.Meta["result"]; !ok {
		m.Meta["result"] = make([]string, 0, 3)
	}

	s := fmt.Sprintf(str, arg...)
	m.Meta["result"] = append(m.Meta["result"], s)
	return s
}

// }}}
func (m *Message) End(s bool) { // {{{
	// log.Println(m.Name, "end", m.Code, ":", m.Meta["detail"])
	if m.Wait != nil {
		m.Wait <- s
	}
	m.Wait = nil
}

// }}}
// }}}
type Server interface { // {{{
	Begin() bool
	Start() bool
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

	Root     *Context
	Context  *Context
	Contexts map[string]*Context

	Index   map[string]*Context
	Shadows map[string]*Context

	Session  map[string]*Message
	Resource []*Message
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

func (c *Context) Request(arg ...string) bool { // {{{
	if c.Session == nil {
		c.Session = make(map[string]*Message)
	}

	m := &Message{
		Code: c.Capi("nmessage", 1),
		Time: time.Now(),
		Meta: map[string][]string{},
		Data: map[string]interface{}{},
		Name: arg[0],
	}
	m.Context = c
	return true
}

// }}}
func (c *Context) Release(key string) bool { // {{{

	return true
}

// }}}
func (c *Context) Destroy(index int) bool { // {{{

	return true
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

func (c *Context) Find(name []string) *Context { // {{{
	if x, ok := c.Contexts[name[0]]; ok {
		log.Println(c.Name, "find:", x.Name)
		if len(name) == 1 {
			return x
		}
		return x.Find(name[1:])
	}

	log.Println(c.Name, "not find:", name[0])
	return nil
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

	if len(arg) > 0 {
		root.Conf("log", arg[0])
	} else {
		root.Conf("log", root.Conf("log"))
	}

	if len(arg) > 1 {
		root.Conf("init.sh", arg[1])
	} else {
		root.Conf("init.sh", root.Conf("init.sh"))
	}

	cs := []*Context{root}

	for i := 0; i < len(cs); i++ {
		cs[i].Root = root
		cs[i].Begin()

		for _, v := range cs[i].Contexts {
			cs = append(cs, v)
		}
	}

	if len(arg) > 2 {
		for _, v := range arg[2:] {
			cs = root.Search(v)
			for _, s := range cs {
				log.Println(v, "match start:", s.Name)
				go s.Start()
			}
		}
	} else {
		go root.Find(strings.Split(root.Conf("default"), ".")).Start()
	}

	<-make(chan bool)
}

// }}}

func (c *Context) Begin() bool { // {{{
	for k, v := range c.Configs {
		c.Conf(k, v.Value)
	}

	if c.Server != nil && c.Server.Begin() {
		c.Root.Capi("ncontext", 1)
		return true
	}

	return false
}

// }}}
func (c *Context) Start() bool { // {{{
	if c.Server != nil && c.Cap("status") != "start" {
		c.Cap("status", "status", "start", "服务状态")
		defer c.Cap("status", "stop")

		c.Root.Capi("nserver", 1)
		defer c.Root.Capi("nserver", -1)

		log.Println(c.Name, "start:")
		c.Server.Start()
		log.Println(c.Name, "stop:")
	}

	return true
}

// }}}
func (c *Context) Spawn(arg ...string) *Context { // {{{
	s := &Context{Name: arg[0], Help: c.Help}
	c.Register(s, c.Server.Spawn(s, arg...))
	s.Begin()

	log.Println(c.Name, "spawn:", s.Name)
	return s
}

// }}}
func (c *Context) Get() *Message { // {{{
	if c.Messages == nil {
		c.Messages = make(chan *Message, c.Confi("MessageQueueSize"))
	}

	select {
	case msg := <-c.Messages:
		// log.Println(c.Name, "get", msg.Code, ":", msg.Meta["detail"])
		return msg
	}

	return nil
}

// }}}
func (c *Context) Post(m *Message) bool { // {{{
	if c.Messages == nil {
		c.Messages = make(chan *Message, c.Confi("MessageQueueSize"))
	}

	m.Code = c.Root.Capi("nmessage", 1)
	// log.Println(c.Name, "post", m.Code, ":", m.Meta["detail"])
	// defer log.Println(c.Name, "done", m.Code, ":", m.Meta["detail"])

	c.Messages <- m
	if m.Wait != nil {
		return <-m.Wait
	}
	return true
}

// }}}

func (c *Context) Cmd(m *Message, arg ...string) string { // {{{
	if x, ok := c.Commands[arg[0]]; ok {
		log.Println(c.Name, "command:", arg)
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
		if v, ok := c.Caches[arg[0]]; ok {
			panic(errors.New(v.Name + "缓存项已存在"))
		}

		c.Caches[arg[0]] = &Cache{arg[1], arg[2], arg[3], nil}
		log.Println(c.Name, "cache:", arg)
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
	// log.Println(c.Name, "cache:", n+value)
	return n
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

		"cert": &Config{Name: "cert", Value: "etc/cert.pem", Help: "证书文件"},
		"key":  &Config{Name: "key", Value: "etc/key.pem", Help: "私钥文件"},

		"debug":   &Config{Name: "debug", Value: "on", Help: "调试模式"},
		"default": &Config{Name: "default", Value: "cli", Help: "默认启动模块"},
		"init.sh": &Config{Name: "init.sh", Value: "etc/hi.sh", Help: "默认启动脚本"},
		"log": &Config{Name: "log", Value: "var/bench.log", Help: "默认日志文件", Hand: func(c *Context, arg string) string {
			if l, e := os.Create(arg); e == nil { // {{{
				log.SetOutput(l)
			} else {
				log.Println("log", arg, "create error")
			}
			return arg
			// }}}
		}},
	},
	Commands: map[string]*Command{},
}
