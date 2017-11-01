package ctx // {{{
// }}}
import ( // {{{
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
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
}

// }}}
type Command struct { // {{{
	Name string
	Help string
	Hand func(c *Context, m *Message, arg ...string) string
}

// }}}
type Message struct { // {{{
	Meta map[string][]string
	Data map[string]interface{}
	Wait chan bool

	*Context
}

func (m *Message) Add(key string, value ...string) string { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if _, ok := m.Meta[key]; !ok {
		m.Meta[key] = []string{}
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
		m.Meta["result"] = []string{}
	}

	s := fmt.Sprintf(str, arg...)
	m.Meta["result"] = append(m.Meta["result"], s)
	return s
}

// }}}
// }}}
type Server interface { // {{{
	Begin() bool
	Start() bool
	Spawn(c *Context, key string) Server
	Fork(c *Context, key string) Server
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
}

func (c *Context) Check(e error) bool { // {{{
	if e != nil {
		log.Println(c.Name, "error:", e)
		debug.PrintStack()
		panic(e)
	}
	return true
}

// }}}
func (c *Context) Find(name []string) *Context { // {{{
	if x, ok := c.Contexts[name[0]]; ok {
		if len(name) == 1 {
			return x
		}
		return x.Find(name[1:])
	}

	return nil
}

// }}}
func (c *Context) Search(name string) []*Context { // {{{
	ps := make([]*Context, 0, 3)

	cs := []*Context{c.Root}
	for i := 0; i < len(cs); i++ {
		for _, v := range cs[i].Contexts {
			cs = append(cs, v)
		}

		if cs[i].Name == name {
			ps = append(ps, cs[i])
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
		panic(errors.New(x.Name + "上下文已存在"))
	}

	c.Contexts[s.Name] = s
	s.Root = c.Root
	s.Context = c
	s.Server = self
	return true
}

// }}}
func (c *Context) Init(arg ...string) { // {{{
	if c.Root == nil {
		c.Root = c
		cs := []*Context{c}
		for i := 0; i < len(cs); i++ {
			for _, v := range cs[i].Contexts {
				cs = append(cs, v)
			}
			cs[i].Init()
		}

		cs = c.Search(arg[0])
		cs[0].Begin()
		cs[0].Start()
	} else {
		for _, v := range c.Contexts {
			v.Root = c.Root
			v.Context = c
		}
	}
}

// }}}
func (c *Context) Fork(key string) { // {{{
	cs := []*Context{new(Context)}
	*cs[0] = *c

	for i := 0; i < len(cs); i++ {
		cs[i].Name = cs[i].Name + key
		cs[i].Messages = make(chan *Message, len(cs[i].Messages))
		cs[i].Context.Register(cs[i], cs[i].Server.Fork(cs[i], key))

		for _, v := range cs[i].Contexts {
			s := new(Context)
			*s = *v
			s.Context = cs[i]
			cs = append(cs, s)
		}
	}
}

// }}}
func (c *Context) Spawn(key string) *Context { // {{{
	s := new(Context)
	s.Name = c.Name + key
	s.Help = c.Help

	s.Caches = make(map[string]*Cache)
	s.Configs = make(map[string]*Config)
	s.Commands = make(map[string]*Command)
	s.Messages = make(chan *Message, len(c.Messages))
	c.Register(s, c.Server.Spawn(s, key))
	log.Println(c.Name, "spawn", c.Contexts[s.Name].Name)
	return s
}

// }}}

func (c *Context) Cap(arg ...string) string { // {{{
	switch len(arg) {
	case 1:
		if v, ok := c.Caches[arg[0]]; ok {
			if v.Hand != nil {
				v.Value = v.Hand(c, v.Value)
			}
			log.Println(c.Name, "cache:", arg)
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
			return v.Value
		}

		if c.Context != nil {
			return c.Context.Cap(arg...)
		}
	case 3:
		if v, ok := c.Caches[arg[0]]; ok {
			panic(errors.New(v.Name + "缓存项已存在"))
		}

		c.Caches[arg[0]] = &Cache{arg[0], arg[1], arg[2], nil}
		log.Println(c.Name, "cache:", arg)
		return arg[1]
	default:
		panic(errors.New(arg[0] + "缓存项参数错误"))
	}

	panic(errors.New(arg[0] + "缓存项不存在"))
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
				v.Hand(c, v.Value)
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

		c.Configs[arg[0]] = &Config{arg[1], arg[2], arg[3], nil}
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
func (c *Context) Post(m *Message) bool { // {{{
	if c.Messages == nil {
		c.Messages = make(chan *Message, 10)
	}

	c.Messages <- m
	if m.Wait != nil {
		return <-m.Wait
	}
	log.Println(c.Context.Name, "message", m.Meta["detail"])
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
// }}}

var Index = &Context{Name: "ctx", Help: "根文", // {{{
	Caches: map[string]*Cache{},
	Configs: map[string]*Config{
		"开场白":   &Config{"开场白", "你好，上下文", "开场白", nil},
		"结束语":   &Config{"结束语", "再见，上下文", "结束语", nil},
		"debug": &Config{"debug", "on", "调试模式", nil},
	},
	Commands: map[string]*Command{},
}

// }}}

func init() { // {{{
	Index.Root = Index
}

// }}}
