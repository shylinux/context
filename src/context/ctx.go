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
	Hand  func(m *Message, x *Cache, arg ...string) string
}

// }}}
type Config struct { // {{{
	Name  string
	Value string
	Help  string
	Hand  func(m *Message, x *Config, arg ...string) string
}

// }}}
type Command struct { // {{{
	Name    string
	Help    string
	Formats map[string]int
	Options map[string]string
	Appends map[string]string
	Hand    func(c *Context, m *Message, key string, arg ...string) string
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

	contexts map[string]*Context
	Context  *Context
	Root     *Context

	Server

	Message  *Message
	Messages chan *Message

	Sessions map[string]*Message
	Requests []*Message
	Master   *Context

	Owner  *Context
	Group  string
	Index  map[string]*Context
	Groups map[string]*Context
}

func (c *Context) Register(s *Context, x Server) *Context { // {{{
	Index.Root = Index

	if c.contexts == nil {
		c.contexts = make(map[string]*Context)
	}
	if x, ok := c.contexts[s.Name]; ok {
		panic(errors.New(c.Name + " 上下文已存在" + x.Name))
	}

	s.Server = x
	s.Context = c
	c.contexts[s.Name] = s
	if c.Root != nil {
		s.Root = c.Root
	} else {
		s.Root = Index
	}

	log.Printf("%s sub(%d): %s", c.Name, Pulse.Capi("ncontext", 1), s.Name)
	return s
}

// }}}
func (c *Context) Begin(m *Message) *Context { // {{{
	for k, x := range c.Configs {
		if x.Hand != nil {
			m.Conf(k, x.Value)
		}
	}

	if c.Server != nil {
		c.Server.Begin(m, m.Meta["detail"]...)
	}

	return c
}

// }}}
func (c *Context) Start(m *Message) bool { // {{{

	if _, ok := c.Caches["status"]; !ok {
		c.Caches["status"] = &Cache{Name: "服务状态", Value: "stop", Help: "服务状态，start:正在运行，stop:未在运行"}
	}

	if m.Cap("status") != "start" && c.Server != nil {
		m.AssertOne(m, true, func(m *Message) {
			m.Cap("status", "start")
			defer m.Cap("status", "stop")

			log.Printf("%d start(%d): %s %s %v", m.Code, m.Root.Capi("nserver", 1), c.Name, c.Help, m.Meta["detail"])
			defer m.Root.Capi("nserver", -1)
			defer log.Printf("%d stop(%s): %s %s", m.Code, m.Root.Cap("nserver"), c.Name, c.Help)

			c.Requests = []*Message{m}
			c.Server.Start(m, m.Meta["detail"]...)
		})
	}
	m.Root.Wait <- true

	return true
}

// }}}
func (c *Context) Spawn(m *Message, key string) *Context { // {{{
	// s := &Context{Name: key, Help: c.Help}
	s := &Context{Name: key, Help: c.Help, Owner: m.Source.Owner}
	m.Target = s
	if m.Template != nil {
		m.Template.Source = s
	}

	if c.Server != nil {
		c.Register(s, c.Server.Spawn(s, m, m.Meta["detail"]...))
	} else {
		c.Register(s, nil)
	}
	return s
}

// }}}
func (c *Context) Exit(m *Message, arg ...string) { // {{{
	if m.Code != 0 {
		log.Printf("%d exit(%s:%s->%s.%d): %s %v", m.Code, m.Source.Name, m.Name, m.Target.Name, m.Index, c.Name, arg)
	} else {
		log.Printf("%d exit(%s->%s): %s %v", m.Code, m.Source.Name, m.Target.Name, c.Name, arg)
	}

	if m.Target == c {
		for _, v := range c.Sessions {
			if v.Target != c {
				v.Target.Exit(v, arg...)
			}
		}

		if c.Server != nil && c.Server.Exit(m, arg...) {
			if len(c.Sessions) == 0 && c.Context != nil {
				delete(c.Context.contexts, c.Name)
			}
		}

		for _, v := range c.Requests {
			if v.Source != c {
				v.Source.Exit(v, arg...)
			}
		}
	} else if m.Source == c {
		delete(c.Sessions, m.Name)

		if c.Server != nil && c.Server.Exit(m, arg...) {
			if len(c.Sessions) == 0 && c.Context != nil {
				delete(c.Context.contexts, c.Name)
			}
		}
	}
}

// }}}

func (c *Context) Add(group string, arg ...string) { // {{{
	if c.Index == nil {
		c.Index = make(map[string]*Context)
	}
	if c.Groups == nil {
		c.Groups = make(map[string]*Context)
	}
	s := c
	if group != "" && group != "root" {
		if g, ok := c.Index[group]; ok {
			s = g
		} else {
			panic(errors.New(group + "上下文不存在"))
		}
	}

	switch arg[0] {
	case "context":
		if len(arg) != 4 {
			panic(errors.New("参数错误"))
		}
		if v, ok := c.Index[arg[1]]; ok {
			panic(errors.New(v.Name + "上下文已存在"))
		}

		s.Groups[arg[1]] = &Context{Name: arg[2], Help: arg[3], Index: c.Index}
		c.Index[arg[1]] = s.Groups[arg[1]]

		log.Println(c.Name, "add context:", arg[1:])
	case "command":
		if len(arg) != 3 {
			panic(errors.New("参数错误"))
		}

		if v, ok := s.Groups[arg[1]]; ok {
			if v.Commands == nil {
				v.Commands = make(map[string]*Command)
			}
			if x, ok := v.Commands[arg[2]]; ok {
				panic(errors.New(x.Name + "命令已存在"))
			}
			if x, ok := s.Commands[arg[2]]; ok {
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

		if v, ok := s.Groups[arg[1]]; ok {
			if v.Configs == nil {
				v.Configs = make(map[string]*Config)
			}
			if x, ok := v.Configs[arg[2]]; ok {
				panic(errors.New(x.Name + "配置项已存在"))
			}
			if x, ok := s.Configs[arg[2]]; ok {
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

		if v, ok := s.Groups[arg[1]]; ok {
			if v.Caches == nil {
				v.Caches = make(map[string]*Cache)
			}
			if x, ok := v.Caches[arg[2]]; ok {
				panic(errors.New(x.Name + "缓存项已存在"))
			}
			if x, ok := s.Caches[arg[2]]; ok {
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

func (c *Context) BackTrace(hand func(s *Context) bool) { // {{{
	for cs := c; cs != nil; cs = cs.Context {
		if !hand(cs) {
			return
		}
	}
}

// }}}
func (c *Context) Find(name string) (s *Context) { // {{{
	ns := strings.Split(name, ".")
	cs := c.contexts
	for _, v := range ns {
		if x, ok := cs[v]; ok {
			cs = x.contexts
			s = x
		} else {
			log.Println(c.Name, "not find:", name)
			return nil
			panic(errors.New(c.Name + " not find: " + name))
		}
	}
	log.Println(c.Name, "find:", name)
	return s
}

// }}}

type Message struct {
	Code int
	Time time.Time

	Meta map[string][]string
	Data map[string]interface{}
	Wait chan bool

	Messages []*Message
	Message  *Message
	Root     *Message

	Name   string
	Source *Context
	Master *Context
	Target *Context
	Index  int

	Template *Message
}

func (m *Message) Assert(e error) bool { // {{{
	if e != nil {
		m.Set("result", "error:", fmt.Sprintln(e))

		log.Println(m.Code, "error:", e)
		if m.Conf("debug") == "on" {
			fmt.Println(m.Code, "error:", e)
		}

		panic(e)
	}
	return true
}

// }}}
func (m *Message) AssertOne(msg *Message, safe bool, hand ...func(msg *Message)) *Message { // {{{
	defer func() {
		if e := recover(); e != nil {
			log.Println(msg.Target.Name, e)
			if msg.Conf("debug") == "on" && e != io.EOF {
				fmt.Println(msg.Target.Name, "error:", e)
				debug.PrintStack()
			}

			if e == io.EOF {
				return
			}

			if len(hand) > 1 {
				m.AssertOne(msg, safe, hand[1:]...)
			} else {
				if !safe {
					log.Println(msg.Target.Name, "error:", e)
					panic(e)
				}
			}
		}
		// Pulse.Wait <- true
	}()

	if len(hand) > 0 {
		hand[0](msg)
	}

	return m
}

// }}}
func (m *Message) Spawn(c *Context, key ...string) *Message { // {{{

	msg := &Message{
		Time:    time.Now(),
		Message: m,
		Root:    m.Root,
		Source:  m.Target,
		Master:  c,
		Target:  c,
	}

	if len(key) == 0 {
		return msg
	}

	msg.Code = m.Capi("nmessage", 1)

	if m.Messages == nil {
		m.Messages = make([]*Message, 0, 10)
	}
	m.Messages = append(m.Messages, msg)

	if msg.Source.Sessions == nil {
		msg.Source.Sessions = make(map[string]*Message)
	}
	msg.Source.Sessions[key[0]] = msg
	msg.Name = key[0]

	log.Printf("%d spawn %d: %s.%s->%s.%d", m.Code, msg.Code, msg.Source.Name, msg.Name, msg.Target.Name, msg.Index)
	return msg
}

// }}}
func (m *Message) Reply(key ...string) *Message { // {{{
	if m.Template == nil {
		return m.Spawn(m.Source, key...)
	}

	msg := m.Template
	msg.Time = time.Now()
	if len(key) == 0 {
		return msg
	}

	msg.Code = m.Capi("nmessage", 1)

	if m.Messages == nil {
		m.Messages = make([]*Message, 0, 10)
	}
	m.Messages = append(m.Messages, msg)

	if msg.Source.Sessions == nil {
		msg.Source.Sessions = make(map[string]*Message)
	}
	msg.Source.Sessions[key[0]] = msg
	msg.Name = key[0]

	log.Printf("%d spawn %d: %s.%s->%s.%d", m.Code, msg.Code, msg.Source.Name, msg.Name, msg.Target.Name, msg.Index)
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
func (m *Message) Set(meta string, arg ...string) *Message { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if len(arg) > 0 {
		m.Meta[meta] = arg
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
		if len(meta) > 0 {
			return meta[0]
		}
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

func (m *Message) Travel(c *Context, hand func(m *Message) bool) { // {{{
	target := m.Target

	cs := []*Context{c}
	for i := 0; i < len(cs); i++ {
		for _, v := range cs[i].contexts {
			cs = append(cs, v)
		}
		m.Target = cs[i]
		if m.Check(m.Target) && !hand(m) {
			break
		}
	}
	m.Target = target
}

// }}}
func (m *Message) Search(c *Context, name string) []*Context { // {{{
	cs := make([]*Context, 0, 3)

	m.Travel(c, func(m *Message) bool {
		if strings.Contains(m.Target.Name, name) || strings.Contains(m.Target.Help, name) {
			cs = append(cs, m.Target)
			log.Println(c.Name, "search:", m.Target.Name, "[match]", name)
		}
		return true
	})
	return cs
}

// }}}

func (m *Message) Start(key string, arg ...string) bool { // {{{
	m.Set("detail", arg...)
	m.Target.Spawn(m, key).Begin(m).Start(m)
	return true
}

// }}}
func (m *Message) Exec(arg ...string) string { // {{{
	cs := []*Context{m.Target, m.Target.Master, m.Source, m.Source.Master}
	for _, c := range cs {
		if c == nil {
			continue
		}
		for s := c; s != nil; s = s.Context {
			if x, ok := s.Commands[arg[0]]; ok {
				if !m.Check(s, "commands", arg[0]) {
					panic(errors.New(fmt.Sprintf("没有权限:" + arg[0])))
				}
				m.Master = s

				success := false
				m.AssertOne(m, true, func(m *Message) {
					if m.Code != 0 {
						log.Printf("%d cmd(%s:%s->%s.%d): %s %v", m.Code, m.Source.Name, m.Name, m.Target.Name, m.Index, c.Name, arg)
					} else {
						log.Printf("%d cmd(%s->%s): %s %v", m.Code, m.Source.Name, m.Target.Name, c.Name, arg)
					}

					if x.Options != nil {
						for _, v := range m.Meta["option"] {
							if _, ok := x.Options[v]; !ok {
								panic(errors.New(fmt.Sprintf("未知参数:" + v)))
							}
						}
					}

					if x.Formats != nil {
						for i, args := 1, m.Meta["detail"]; i < len(args); i++ {
							n, ok := x.Formats[args[i]]
							if !ok {
								m.Add("option", "args", arg[i])
								continue
							}

							if n < 0 {
								n += len(args) - i
							}

							m.Add("option", args[i], arg[i+1:i+1+n]...)
							i += n
						}
					}

					m.Meta["result"] = nil
					ret := x.Hand(c, m, arg[0], arg[1:]...)
					if ret != "" {
						m.Echo(ret)
					}

					if x.Appends != nil {
						for _, v := range m.Meta["append"] {
							if _, ok := x.Appends[v]; !ok {
								panic(errors.New(fmt.Sprintf("未知参数:" + v)))
							}
						}
					}

					success = true
				})

				return m.Get("result")
			}
		}
	}

	m.AssertOne(m, true, func(m *Message) {
		log.Printf("system command(%s->%s): %v", m.Source.Name, m.Target.Name, arg)
		cmd := exec.Command(arg[0], arg[1:]...)
		v, e := cmd.CombinedOutput()
		if e != nil {
			m.Echo("%s\n", e)
		} else {
			m.Echo(string(v))
		}
	})
	return ""
}

// }}}
func (m *Message) Deal(pre func(msg *Message, arg ...string) bool, post func(msg *Message, arg ...string) bool) (live bool) { // {{{
	if m.Target.Messages == nil {
		m.Target.Messages = make(chan *Message, m.Confi("MessageQueueSize"))
	}

	msg := <-m.Target.Messages
	defer msg.End(true)

	if len(msg.Meta["detail"]) == 0 {
		return true
	}

	if pre != nil && !pre(msg, msg.Meta["detail"]...) {
		return false
	}

	m.AssertOne(msg, true, func(msg *Message) {
		msg.Exec(msg.Meta["detail"]...)
	})

	if post != nil && !post(msg, msg.Meta["result"]...) {
		return false
	}

	return true
}

// }}}
func (m *Message) Post(s *Context) string { // {{{
	if s.Messages == nil {
		panic(s.Name + " 没有开启消息处理")
	}

	s.Messages <- m
	if m.Wait != nil {
		<-m.Wait
	}

	return m.Get("result")
}

// }}}

func (m *Message) Check(s *Context, arg ...string) bool { // {{{
	if s.Owner == nil {
		return true
	}

	if m.Master.Owner == s.Owner {
		return true
	}
	if m.Master.Owner == s.Root.Owner {
		return true
	}

	g, ok := s.Index[m.Master.Group]
	if !ok {
		if g, ok = s.Index["void"]; !ok {
			if m.Master.Owner != nil {
				log.Printf("%s(%s:%s) not auth: %s(%s)", m.Master.Name, m.Master.Owner.Name, m.Master.Group, s.Name, s.Owner.Name)
			} else {
				log.Printf("%s() not auth: %s(%s)", m.Master.Name, s.Name, s.Owner.Name)
			}

			return false
		}
	}

	if len(arg) < 2 {
		return true
	}

	switch arg[0] {
	case "commands":
		_, ok = g.Commands[arg[1]]
	case "configs":
		_, ok = g.Configs[arg[1]]
	case "caches":
		_, ok = g.Caches[arg[1]]
	}

	if !ok {
		if m.Master.Owner != nil {
			log.Printf("%s(%s:%s) not auth: %s(%s) %s %s", m.Master.Name, m.Master.Owner.Name, m.Master.Group, s.Name, s.Owner.Name, g.Name, arg[1])
		} else {
			log.Printf("%s() not auth: %s(%s) %s %s", m.Master.Name, s.Name, s.Owner.Name, g.Name, arg[1])
		}
		return false
	}

	return true
}

// }}}

func (m *Message) Cmd(arg ...string) string { // {{{
	m.Set("detail", arg...)

	if s := m.Target.Master; s != nil && s != m.Source.Master {
		m.Post(s)
	}

	return m.Exec(m.Meta["detail"]...)
}

// }}}
func (m *Message) Conf(key string, arg ...string) string { // {{{
	for s := m.Target; s != nil; s = s.Context {
		if x, ok := s.Configs[key]; ok {
			if !m.Check(s, "configs", key) {
				panic(errors.New(fmt.Sprintf("没有权限:" + key)))
			}

			switch len(arg) {
			case 0:
				if x.Hand != nil {
					return x.Hand(m, x)
				}
				return x.Value
			case 1:
				if m.Code != 0 {
					log.Printf("%d conf(%s:%s->%s.%d): %s %v", m.Code, m.Source.Name, m.Name, m.Target.Name, m.Index, key, arg)
				} else {
					log.Printf("%d conf(%s->%s): %s %v", m.Code, m.Source.Name, m.Target.Name, key, arg)
				}

				x.Value = arg[0]
				if x.Hand != nil {
					x.Hand(m, x, x.Value)
				}
				return x.Value
			case 3:
				if m.Code != 0 {
					log.Printf("%d conf(%s:%s->%s.%d): %s %v", m.Code, m.Source.Name, m.Name, m.Target.Name, m.Index, key, arg)
				} else {
					log.Printf("%d conf(%s->%s): %s %v", m.Code, m.Source.Name, m.Target.Name, key, arg)
				}

				if s == m.Target {
					panic(errors.New(key + "配置项已存在"))
				}
				if m.Target.Configs == nil {
					m.Target.Configs = make(map[string]*Config)
				}
				m.Target.Configs[key] = &Config{Name: arg[0], Value: arg[1], Help: arg[2], Hand: x.Hand}
				return arg[1]
			default:
				panic(errors.New(key + "配置项参数错误"))
			}
		}
	}

	if len(arg) == 3 {
		log.Println(m.Target.Name, "conf:", key, arg)
		if m.Target.Configs == nil {
			m.Target.Configs = make(map[string]*Config)
		}
		m.Target.Configs[key] = &Config{Name: arg[0], Value: arg[1], Help: arg[2]}
		return arg[1]
	}

	panic(errors.New(key + "配置项不存在"))
}

// }}}
func (m *Message) Confi(key string, arg ...int) int { // {{{
	if len(arg) > 0 {
		n, e := strconv.Atoi(m.Conf(key, fmt.Sprintf("%d", arg[0])))
		m.Assert(e)
		return n
	}

	n, e := strconv.Atoi(m.Conf(key))
	m.Assert(e)
	return n
}

// }}}
func (m *Message) Cap(key string, arg ...string) string { // {{{
	// if m.Code != 0 {
	// 	log.Printf("%d cap(%s:%s->%s.%d): %s %v", m.Code, m.Context.Name, m.Name, m.Master.Name, m.Index, key, arg)
	// } else {
	// 	log.Printf("%d cap(%s->%s): %s %v", m.Code, m.Context.Name, m.Master.Name, key, arg)
	// }
	//
	for s := m.Target; s != nil; s = s.Context {
		if x, ok := s.Caches[key]; ok {
			if !m.Check(s, "caches", key) {
				panic(errors.New(fmt.Sprintf("没有权限:" + key)))
			}

			switch len(arg) {
			case 0:
				if x.Hand != nil {
					x.Value = x.Hand(m, x)
				}
				return x.Value
			case 1:
				if x.Hand != nil {
					x.Value = x.Hand(m, x, arg[0])
				} else {
					x.Value = arg[0]
				}
				return x.Value
			case 3:
				log.Println(m.Target.Name, "cap:", key, arg)
				if s == m.Target {
					panic(errors.New(key + "缓存项已存在"))
				}
				if m.Target.Caches == nil {
					m.Target.Caches = make(map[string]*Cache)
				}
				m.Target.Caches[key] = &Cache{arg[0], arg[1], arg[2], x.Hand}
				return arg[1]
			default:
				panic(errors.New(key + "缓存项参数错误"))
			}
		}
	}
	if len(arg) == 3 {
		log.Println(m.Target.Name, "cap:", key, arg)
		if m.Target.Caches == nil {
			m.Target.Caches = make(map[string]*Cache)
		}
		m.Target.Caches[key] = &Cache{arg[0], arg[1], arg[2], nil}
		return arg[1]
	}

	panic(errors.New(key + "缓存项不存在"))
}

// }}}
func (m *Message) Capi(key string, arg ...int) int { // {{{
	n, e := strconv.Atoi(m.Cap(key))
	m.Assert(e)
	if len(arg) > 0 {
		n += arg[0]
		m.Cap(key, strconv.Itoa(n))
	}
	return n
}

// }}}

var Pulse = &Message{Code: 0, Time: time.Now(), Source: Index, Master: Index, Target: Index}

var Index = &Context{Name: "ctx", Help: "根模块",
	Caches: map[string]*Cache{
		"nserver":  &Cache{Name: "服务数量", Value: "0", Help: "显示已经启动运行模块的数量"},
		"ncontext": &Cache{Name: "模块数量", Value: "0", Help: "显示功能树已经注册模块的数量"},
		"nmessage": &Cache{Name: "消息数量", Value: "0", Help: "显示模块启动时所创建消息的数量"},
	},
	Configs: map[string]*Config{
		"start":   &Config{Name: "启动模块", Value: "cli", Help: "启动时自动运行的模块"},
		"init.sh": &Config{Name: "启动脚本", Value: "etc/init.sh", Help: "模块启动时自动运行的脚本"},
		"bench.log": &Config{Name: "日志文件", Value: "var/bench.log", Help: "模块日志输出的文件", Hand: func(m *Message, x *Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				if e := os.MkdirAll(path.Dir(arg[0]), os.ModePerm); e == nil {
					if l, e := os.Create(x.Value); e == nil {
						log.SetOutput(l)
						log.Println("\n\n")
					}
				}
			}
			return x.Value
			// }}}
		}},
		"root": &Config{Name: "工作目录", Value: ".", Help: "所有模块的当前目录", Hand: func(m *Message, x *Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				if !path.IsAbs(x.Value) {
					wd, e := os.Getwd()
					m.Assert(e)
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

		"debug": &Config{Name: "调试模式(off/on)", Value: "on", Help: "是否打印错误信息，off:不打印，on:打印)"},
		"cert":  &Config{Name: "证书文件", Value: "etc/cert.pem", Help: "证书文件"},
		"key":   &Config{Name: "私钥文件", Value: "etc/key.pem", Help: "私钥文件"},
	},
	Commands: map[string]*Command{
		"userinfo": &Command{Name: "userinfo [add|del [context key name help]|[command|config|cache group name]]", Help: "查看模块的用户信息",
			Formats: map[string]int{"add": -1, "del": -1},
			Hand: func(c *Context, m *Message, key string, arg ...string) string {
				log.Println(m.Meta)
				switch {
				case m.Has("add"):
					m.Target.Add(m.Source.Group, m.Meta["add"]...)
				case m.Has("del"):
					m.Target.Del(m.Meta["del"]...)
				default:
					target := m.Target
					m.Target = target.Owner
					if m.Target != nil && m.Check(m.Target) {
						m.Echo("%s %s\n", m.Cap("username"), m.Cap("group"))
					}
					m.Target = target

					if len(m.Meta["args"]) > 0 {
						if g, ok := m.Target.Index[m.Get("args")]; ok {
							for k, _ := range g.Commands {
								m.Echo("cmd: %s\n", k)
							}
							for k, _ := range g.Configs {
								m.Echo("cfg: %s\n", k)
							}
							for k, _ := range g.Caches {
								m.Echo("cap: %s\n", k)
							}
						}
					} else {
						for k, v := range m.Target.Index {
							m.Echo("%s", k)
							m.Echo(": %s %s\n", v.Name, v.Help)
						}
					}
				}
				return ""
			}},
		"server": &Command{Name: "server [start|exit|switch][args]", Help: "服务启动停止切换", Hand: func(c *Context, m *Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				m.Travel(m.Target.Root, func(m *Message) bool {
					if x, ok := m.Target.Caches["status"]; ok {
						m.Echo("%s(%s): %s\n", m.Target.Name, x.Value, m.Target.Help)
					}
					return true
				})

			default:
				switch arg[0] {
				case "start":
					go m.Set("detail", arg[1:]...).Target.Start(m)
				case "stop":
					m.Set("detail", arg[1:]...).Target.Exit(m)
				case "switch":
				}
			}
			return ""
			// }}}
		}},
		"message": &Command{Name: "message", Help: "查看消息", Hand: func(c *Context, m *Message, key string, arg ...string) string {
			ms := []*Message{m.Root} // {{{
			for i := 0; i < len(ms); i++ {
				if ms[i].Code != 0 {
					m.Echo("%d %s.%s -> %s.%d: %s %v\n", ms[i].Code, ms[i].Source.Name, ms[i].Name, ms[i].Target.Name, ms[i].Index, ms[i].Time.Format("15:04:05"), ms[i].Meta["detail"])
				}
				ms = append(ms, ms[i].Messages...)
			}
			return ""
			// }}}
		}},
		"command": &Command{Name: "command [all] [key [name help]]", Help: "查看或修改命令", Hand: func(c *Context, m *Message, key string, arg ...string) string {
			all := false // {{{
			if len(arg) > 0 && arg[0] == "all" {
				arg = arg[1:]
				all = true
			}

			m.Target.BackTrace(func(s *Context) bool {
				switch len(arg) {
				case 0:
					for k, v := range s.Commands {
						if m.Check(m.Target, "commands", k) {
							m.Echo("%s: %s\n", k, v.Name)
						}
					}
				case 1:
					if v, ok := s.Commands[arg[0]]; ok {
						if m.Check(m.Target, "commands", arg[0]) {
							m.Echo("%s\n%s\n", v.Name, v.Help)
						}
					}
				case 3:
					if v, ok := s.Commands[arg[0]]; ok {
						if m.Check(m.Target, "commands", arg[0]) {
							v.Name = arg[1]
							v.Help = arg[2]
							m.Echo("%s\n%s\n", v.Name, v.Help)
						}
					}
					return false
				}
				return all
			})
			return ""
			// }}}
		}},
		"config": &Command{Name: "config [all] [[delete|void] key [value]|[name value help]]", Help: "删除、空值、查看、修改或添加配置", Hand: func(c *Context, m *Message, key string, arg ...string) string {
			all := false // {{{
			if len(arg) > 0 && arg[0] == "all" {
				arg = arg[1:]
				all = true
			}

			m.Target.BackTrace(func(s *Context) bool {
				switch len(arg) {
				case 0:
					for k, v := range s.Configs {
						if m.Check(m.Target, "configs", k) {
							m.Echo("%s(%s): %s\n", k, v.Value, v.Name)
						}
					}
				case 1:
					if v, ok := s.Configs[arg[0]]; ok {
						if m.Check(m.Target, "configs", arg[0]) {
							m.Echo("%s: %s\n", v.Name, v.Help)
						}
					}
				case 2:
					if s != m.Target {
						m.Echo("请到%s模块上下文中操作配置%v", s.Name, arg)
						return false
					}

					switch arg[0] {
					case "void":
						if m.Check(m.Target, "configs", arg[1]) {
							m.Conf(arg[1], "")
						}
					case "delete":
						if _, ok := s.Configs[arg[1]]; ok {
							if m.Check(m.Target, "configs", arg[1]) {
								delete(s.Configs, arg[1])
							}
						}
					default:
						if m.Check(m.Target, "configs", arg[0]) {
							m.Conf(arg[0], arg[1])
						}
					}
				case 4:
					if m.Check(m.Target) {
						m.Conf(arg[0], arg[1:]...)
					}
					return false
				}
				return all
			})
			return ""
			// }}}
		}},
		"cache": &Command{Name: "cache [all] [[delete] key [value]|[name value help]]", Help: "删除、查看、修改或添加配置", Hand: func(c *Context, m *Message, key string, arg ...string) string {
			all := false // {{{
			if len(arg) > 0 && arg[0] == "all" {
				arg = arg[1:]
				all = true
			}

			m.Target.BackTrace(func(s *Context) bool {
				switch len(arg) {
				case 0:
					for k, v := range s.Caches {
						if m.Check(m.Target, "caches", k) {
							m.Echo("%s(%s): %s\n", k, v.Value, v.Name)
						}
					}
				case 1:
					if v, ok := s.Caches[arg[0]]; ok {
						if m.Check(m.Target, "caches", arg[0]) {
							m.Echo("%s: %s\n", v.Name, v.Help)
						}
					}
				case 2:
					if s != m.Target {
						m.Echo("请到%s模块上下文中操作缓存%v", s.Name, arg)
						return false
					}

					switch arg[0] {
					case "delete":
						if m.Check(m.Target, "caches", arg[1]) {
							if _, ok := s.Caches[arg[1]]; ok {
								delete(s.Caches, arg[1])
							}
						}
					default:
						if _, ok := s.Caches[arg[0]]; ok {
							if m.Check(m.Target, "caches", arg[0]) {
								m.Echo("%s: %s\n", arg[0], m.Cap(arg[0], arg[1]))
							}
						}
					}
				case 4:
					if m.Check(m.Target) {
						m.Cap(arg[0], arg[1:]...)
					}
					return false
				}

				return all
			})
			return ""
			// }}}
		}},
	},
	Index: map[string]*Context{
		"void": &Context{Name: "void",
			Configs: map[string]*Config{
				"debug": &Config{},
			},
			Commands: map[string]*Command{
				"command": &Command{},
				"config":  &Command{},
				"cache":   &Command{},
			},
		},
	},
}

func init() {
	Pulse.Root = Pulse
	Pulse.Wait = make(chan bool, 10)
}

func Start(args ...string) {
	if len(args) > 0 {
		Pulse.Conf("start", args[0])
	}
	if len(args) > 1 {
		Pulse.Conf("init.sh", args[1])
	}
	if len(args) > 2 {
		Pulse.Conf("bench.log", args[2])
	} else {
		Pulse.Conf("bench.log", Pulse.Conf("bench.log"))
	}
	if len(args) > 3 {
		Pulse.Conf("root", args[3])
	}
	log.Println("\n\n\n")

	Pulse.Travel(Index, func(m *Message) bool {
		m.Target.Begin(m)
		return true
	})
	Pulse.Target = Index

	if n := 0; Pulse.Conf("start") != "" {
		for _, s := range Index.contexts {
			if ok, _ := regexp.MatchString(Pulse.Conf("start"), s.Name); ok {
				go s.Start(Pulse.Spawn(s, s.Name).Put("option", "io", os.Stdout))
				n++
			}
		}

		for n > 0 || Pulse.Capi("nserver") > 0 {
			<-Pulse.Wait
			n--
		}
	}
}
