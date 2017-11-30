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

type Cache struct {
	Name  string
	Value string
	Help  string
	Hand  func(m *Message, x *Cache, arg ...string) string
}

type Config struct {
	Name  string
	Value string
	Help  string
	Hand  func(m *Message, x *Config, arg ...string) string
}

type Command struct {
	Name    string
	Help    string
	Formats map[string]int
	Options map[string]string
	Appends map[string]string
	Hand    func(c *Context, m *Message, key string, arg ...string) string
}

type Server interface {
	Spawn(c *Context, m *Message, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
}

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
	if c.contexts == nil {
		c.contexts = make(map[string]*Context)
	}
	if x, ok := c.contexts[s.Name]; ok {
		panic(errors.New(c.Name + "上下文中已存在模块:" + x.Name))
	}

	c.contexts[s.Name] = s
	s.Context = c
	s.Server = x

	s.Root = Index
	if c.Root != nil {
		s.Root = c.Root
	}
	return s
}

// }}}
func (c *Context) Spawn(m *Message, name string, help string) *Context { // {{{
	s := &Context{Name: name, Help: help, Context: c, Owner: m.Source.Owner}
	m.Log("spawn", "%s: %s", name, help)
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
func (c *Context) Begin(m *Message) *Context { // {{{
	m.Log("begin", "%s: %d %v", c.Name, Pulse.Capi("ncontext", 1), m.Meta["detail"])
	for k, x := range c.Configs {
		if x.Hand != nil {
			m.Conf(k, x.Value)
		}
	}

	c.Requests = []*Message{m}

	if c.Server != nil {
		c.Server.Begin(m, m.Meta["detail"]...)
	}

	return c
}

// }}}
func (c *Context) Start(m *Message) bool { // {{{
	if _, ok := c.Caches["status"]; !ok {
		c.Caches["status"] = &Cache{Name: "服务状态(start/stop)", Value: "stop", Help: "服务状态，start:正在运行，stop:未在运行"}
	}

	if m.Cap("status") != "start" && c.Server != nil {
		running := make(chan bool)
		go m.AssertOne(m, true, func(m *Message) {
			m.Log(m.Cap("status", "start"), "%s: %d %v", c.Name, m.Root.Capi("nserver", 1), m.Meta["detail"])

			running <- true
			c.Requests = append(c.Requests, m)
			if c.Server.Start(m, m.Meta["detail"]...) {
				c.Close(m, m.Meta["detail"]...)
			}
		})
		<-running
	}

	return true
}

// }}}
func (c *Context) Close(m *Message, arg ...string) { // {{{
	if c.Server != nil && c.Server.Close(m, arg...) {
		m.Log("close", "%s: %d %v", c.Name, m.Root.Capi("ncontext", -1)+1, arg)
		delete(c.Context.contexts, c.Name)
	}
	return

	if m.Target == c {
		for _, v := range c.Sessions {
			if v.Target != c {
				v.Target.Close(v, arg...)
			}
		}

		if c.Server != nil && c.Server.Close(m, arg...) {
			if len(c.Sessions) == 0 && c.Context != nil {
				delete(c.Context.contexts, c.Name)
			}
		}

		for _, v := range c.Requests {
			if v.Source != c {
				v.Source.Close(v, arg...)
			}
		}
	} else if m.Source == c {
		delete(c.Sessions, m.Name)

		if c.Server != nil && c.Server.Close(m, arg...) {
			if len(c.Sessions) == 0 && c.Context != nil {
				delete(c.Context.contexts, c.Name)
			}
		}
	}

	if len(c.Requests) == 0 {
		m.Log(m.Cap("status", "stop"), "%s: %d %v", c.Name, m.Root.Capi("nserver", -1), m.Meta["detail"])
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

func (m *Message) Log(action, str string, arg ...interface{}) { // {{{
	color := 0
	switch action {
	case "error", "check":
		color = 31
	case "cmd":
		color = 32
	case "conf":
		color = 33
	case "search", "find", "spawn":
		color = 35
	case "begin", "start", "close":
		color = 36
	case "debug":
		if m.Conf("debug") != "on" {
			return
		}
	}

	if m.Name != "" {
		log.Printf("\033[%dm%d %s(%s:%s->%s.%d) %s\033[0m", color, m.Code, action, m.Source.Name, m.Name, m.Target.Name, m.Index, fmt.Sprintf(str, arg...))
	} else {
		log.Printf("\033[%dm%d %s(%s->%s) %s\033[0m", color, m.Code, action, m.Source.Name, m.Target.Name, fmt.Sprintf(str, arg...))
	}
}

// }}}
func (m *Message) Assert(e interface{}, msg ...string) bool { // {{{
	switch e := e.(type) {
	case bool:
		if e {
			return true
		}
	case string:
		if e != "error:" {
			return true
		}
	case error:
		if e == nil {
			return true
		}
	default:
		return true
	}

	if len(msg) > 0 {
		e = errors.New(msg[0])
	}
	if _, ok := e.(error); !ok {
		e = errors.New("error")
	}

	m.Set("result", "error:", fmt.Sprintln(e))
	panic(e)
}

// }}}
func (m *Message) AssertOne(msg *Message, safe bool, hand ...func(msg *Message)) *Message { // {{{
	defer func() {
		if e := recover(); e != nil {
			msg.Log("error", "error: %v", e)
			if msg.Conf("debug") == "on" && e != io.EOF {
				fmt.Println(msg.Target.Name, "error:", e)
				debug.PrintStack()
			}

			if e == io.EOF {
				return
			} else if len(hand) > 1 {
				m.AssertOne(msg, safe, hand[1:]...)
			} else if !safe {
				panic(e)
			}
		}
	}()

	if len(hand) > 0 {
		hand[0](msg)
	}

	return m
}

// }}}
func (m *Message) Spawn(c *Context, key ...string) *Message { // {{{
	msg := &Message{
		Code:    m.Capi("nmessage", 1),
		Time:    time.Now(),
		Message: m,
		Root:    m.Root,
		Source:  m.Target,
		Master:  m.Target,
		Target:  c,
	}

	if m.Messages == nil {
		m.Messages = make([]*Message, 0, 10)
	}
	m.Messages = append(m.Messages, msg)

	if len(key) == 0 {
		return msg
	}

	if msg.Source.Sessions == nil {
		msg.Source.Sessions = make(map[string]*Message)
	}
	msg.Source.Sessions[key[0]] = msg
	msg.Name = key[0]

	m.Log("spawn", "%d: %s.%s->%s.%d", msg.Code, msg.Source.Name, msg.Name, msg.Target.Name, msg.Index)
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

	m.Log("spawn", "%d: %s.%s->%s.%d", msg.Code, msg.Source.Name, msg.Name, msg.Target.Name, msg.Index)
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

func (m *Message) BackTrace(hand func(m *Message) bool) { // {{{
	target := m.Target
	for cs := target; cs != nil; cs = cs.Context {
		if m.Check(cs) && !hand(m) {
			break
		}
	}
	m.Target = target
}

// }}}
func (m *Message) Travel(c *Context, hand func(m *Message) bool) { // {{{
	target := m.Target

	cs := []*Context{c}
	for i := 0; i < len(cs); i++ {
		if m.Target = cs[i]; m.Check(cs[i]) && !hand(m) {
			break
		}

		for _, v := range cs[i].contexts {
			cs = append(cs, v)
		}
	}

	m.Target = target
}

// }}}
func (m *Message) Search(key string, begin ...*Context) []*Message { // {{{
	target := m.Target
	if len(begin) > 0 {
		target = begin[0]
	}

	reg, e := regexp.Compile(key)
	m.Assert(e)

	ms := make([]*Message, 0, 3)
	m.Travel(target, func(m *Message) bool {
		if reg.MatchString(m.Target.Name) || reg.FindString(m.Target.Help) != "" {
			m.Log("search", "%s: %d match [%s]", m.Target.Name, len(ms)+1, key)
			ms = append(ms, m.Spawn(m.Target))
		}
		return true
	})

	for _, v := range ms {
		v.Source = m.Target
	}

	return ms
}

// }}}
func (m *Message) Find(name string, begin ...*Context) *Message { // {{{
	target := m.Target
	if len(begin) > 0 {
		target = begin[0]
	}

	cs := target.contexts
	for _, v := range strings.Split(name, ".") {
		if x, ok := cs[v]; ok {
			cs = x.contexts
			target = x
		} else {
			m.Log("find", "%s: not find %s", target.Name, v)
			return nil
		}
	}
	m.Log("find", "%s: find %s", m.Target.Name, name)
	return m.Spawn(target)
}

// }}}

func (m *Message) Start(name string, help string, arg ...string) bool { // {{{
	return m.Set("detail", arg...).Target.Spawn(m, name, help).Begin(m).Start(m)
}

// }}}
func (m *Message) Exec(key string, arg ...string) string { // {{{
	cs := []*Context{m.Target, m.Target.Master, m.Source, m.Source.Master}
	for _, c := range cs {
		if c == nil {
			continue
		}
		for s := c; s != nil; s = s.Context {
			if x, ok := s.Commands[key]; ok {
				if !m.Check(s, "commands", key) {
					break
				}

				success := false
				m.AssertOne(m, true, func(m *Message) {
					m.Log("cmd", "%s: %s %v", s.Name, key, arg)

					if x.Options != nil {
						for _, v := range m.Meta["option"] {
							if _, ok := x.Options[v]; !ok {
								panic(errors.New(fmt.Sprintf("未知参数:" + v)))
							}
						}
					}

					if x.Formats != nil {
						for i, args := 0, arg; i < len(args); i++ {
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
						arg = m.Meta["args"]
					}

					m.Meta["result"] = nil
					ret := x.Hand(c, m, key, arg...)
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

					if c.Requests == nil {
						c.Requests = make([]*Message, 0, 10)
					}
					c.Requests = append(c.Requests, m)

					success = true
				})

				return m.Get("result")
			}
		}
	}

	m.AssertOne(m, true, func(m *Message) {
		m.Log("system", ": %s %v", key, arg)
		cmd := exec.Command(key, arg...)
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
		msg.Exec(msg.Meta["detail"][0], msg.Meta["detail"][1:]...)
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
				m.Log("check", "%s(%s:%s) not auth: %s(%s)", m.Master.Name, m.Master.Owner.Name, m.Master.Group, s.Name, s.Owner.Name)
			} else {
				m.Log("check", "%s() not auth: %s(%s)", m.Master.Name, s.Name, s.Owner.Name)
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
			m.Log("check", "%s(%s:%s) not auth: %s(%s) %s %s", m.Master.Name, m.Master.Owner.Name, m.Master.Group, s.Name, s.Owner.Name, g.Name, arg[1])
		} else {
			m.Log("check", "%s() not auth: %s(%s) %s %s", m.Master.Name, s.Name, s.Owner.Name, g.Name, arg[1])
		}
		return false
	}

	return true
}

// }}}

func (m *Message) Cmd(arg ...string) string { // {{{
	m.Set("detail", arg...)

	if s := m.Target.Master; s != nil && s != m.Source.Master {
		if s.Messages == nil {
			panic(s.Name + " 没有开启消息处理")
		}

		s.Messages <- m
		if m.Wait != nil {
			<-m.Wait
		}

		return m.Get("result")
	}

	return m.Exec(m.Meta["detail"][0], m.Meta["detail"][1:]...)
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
				m.Log("conf", "%s: %s %v", s.Name, key, arg)

				x.Value = arg[0]
				if x.Hand != nil {
					x.Hand(m, x, x.Value)
				}
				return x.Value
			case 3:
				m.Log("conf", "%s: %s %v", s.Name, key, arg)

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
		m.Log("conf", "%s: %s %v", m.Target.Name, key, arg)
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
				m.Log("cap", "%s: %s %v", m.Target.Name, key, arg)
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
		m.Log("cap", "%s: %s %v", m.Target.Name, key, arg)
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

		"debug": &Config{Name: "调试模式(off/on)", Value: "off", Help: "是否打印错误信息，off:不打印，on:打印)"},
		"cert":  &Config{Name: "证书文件", Value: "etc/cert.pem", Help: "证书文件"},
		"key":   &Config{Name: "私钥文件", Value: "etc/key.pem", Help: "私钥文件"},
	},
	Commands: map[string]*Command{
		"userinfo": &Command{Name: "userinfo [add|del [context key name help]|[command|config|cache group name]]", Help: "查看模块的用户信息",
			Formats: map[string]int{"add": -1, "del": -1},
			Hand: func(c *Context, m *Message, key string, arg ...string) string {
				switch { // {{{
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
				// }}}
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
					m.Set("detail", arg[1:]...).Target.Start(m)
				case "stop":
					m.Set("detail", arg[1:]...).Target.Close(m)
				case "switch":
				}
			}
			return ""
			// }}}
		}},
		"message": &Command{Name: "message [index|home] [order]", Help: "查看消息", Hand: func(c *Context, m *Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				for k, v := range m.Target.Sessions {
					if v.Name != "" {
						m.Echo("%s %s.%s -> %s.%d: %s %v\n", k, v.Source.Name, v.Name, v.Target.Name, v.Index, v.Time.Format("15:04:05"), v.Meta["detail"])
					} else {
						m.Echo("%s %s -> %s: %s %v\n", k, v.Source.Name, v.Target.Name, v.Time.Format("15:04:05"), v.Meta["detail"])
					}
				}

				for i, v := range m.Target.Requests {
					if v.Name != "" {
						m.Echo("%d %s.%s -> %s.%d: %s %v\n", i, v.Source.Name, v.Name, v.Target.Name, v.Index, v.Time.Format("15:04:05"), v.Meta["detail"])
					} else {
						m.Echo("%d %s -> %s: %s %v\n", i, v.Source.Name, v.Target.Name, v.Time.Format("15:04:05"), v.Meta["detail"])
					}
					for i, v := range v.Messages {
						if v.Name != "" {
							m.Echo("  %d %s.%s -> %s.%d: %s %v\n", i, v.Source.Name, v.Name, v.Target.Name, v.Index, v.Time.Format("15:04:05"), v.Meta["detail"])
						} else {
							m.Echo("  %d %s -> %s: %s %v\n", i, v.Source.Name, v.Target.Name, v.Time.Format("15:04:05"), v.Meta["detail"])
						}
					}
				}
			case 1, 2:
				n, e := strconv.Atoi(arg[0])
				v := m
				if e == nil && 0 <= n && n < len(m.Target.Requests) {
					v = m.Target.Requests[n]
				} else {
					v = m.Target.Sessions[arg[0]]
				}

				if v != nil {
					if len(arg) > 1 {
						if n, e = strconv.Atoi(arg[1]); e == nil && 0 <= n && n < len(v.Messages) {
							v = v.Messages[n]
						}
					}

					if v.Name != "" {
						m.Echo("%s.%s -> %s.%d: %s %v\n", v.Source.Name, v.Name, v.Target.Name, v.Index, v.Time.Format("15:04:05"), v.Meta["detail"])
					} else {
						m.Echo("%s -> %s: %s %v\n", v.Source.Name, v.Target.Name, v.Time.Format("15:04:05"), v.Meta["detail"])
					}

					if len(v.Meta["option"]) > 0 {
						m.Echo("option:\n")
					}
					for _, k := range v.Meta["option"] {
						m.Echo("  %s: %v\n", k, v.Meta[k])
					}
					if len(v.Meta["result"]) > 0 {
						m.Echo("result: %v\n", v.Meta["result"])
					}
					if len(v.Meta["append"]) > 0 {
						m.Echo("append:\n")
					}
					for _, k := range v.Meta["append"] {
						m.Echo("  %s: %v\n", k, v.Meta[k])
					}
				}
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

			m.BackTrace(func(m *Message) bool {
				switch len(arg) {
				case 0:
					for k, v := range m.Target.Commands {
						if m.Check(m.Target, "commands", k) {
							m.Echo("%s: %s\n", k, v.Name)
						}
					}
				case 1:
					if v, ok := m.Target.Commands[arg[0]]; ok {
						if m.Check(m.Target, "commands", arg[0]) {
							m.Echo("%s\n%s\n", v.Name, v.Help)
						}
					}
				case 3:
					if v, ok := m.Target.Commands[arg[0]]; ok {
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
		"config": &Command{Name: "config [all] [[delete|void] key [value]|[name value help]]", Help: "删除、空值、查看、修改或添加配置",
			Formats: map[string]int{"all": 0, "delete": 0, "void": 0},
			Hand: func(c *Context, m *Message, key string, arg ...string) string {
				all := m.Has("all") // {{{

				switch len(arg) {
				case 0:
					m.BackTrace(func(m *Message) bool {
						if all {
							m.Echo("%s configs:\n", m.Target.Name)
						}
						for k, v := range m.Target.Configs {
							if m.Check(m.Target, "configs", k) {
								if all {
									m.Echo("  ")
								}
								m.Echo("%s(%s): %s\n", k, v.Value, v.Name)
							}
						}
						return all
					})
				case 1:
					m.BackTrace(func(m *Message) bool {
						if all {
							m.Echo("%s config:\n", m.Target.Name)
						}
						if v, ok := m.Target.Configs[arg[0]]; ok {
							if m.Check(m.Target, "configs", arg[0]) {
								if all {
									m.Echo("  ")
								}
								m.Echo("%s: %s\n", v.Name, v.Help)
							}
						}
						return all
					})

				case 2:
					switch arg[0] {
					case "delete":
						if _, ok := m.Target.Configs[arg[1]]; ok {
							if m.Check(m.Target, "configs", arg[1]) {
								delete(m.Target.Configs, arg[1])
							}
						}
					case "void":
						m.Conf(arg[1], "")
					default:
						m.Conf(arg[0], arg[1])
					}
				case 4:
					m.Conf(arg[0], arg[1:]...)
				}
				return ""
				// }}}
			}},
		"cache": &Command{Name: "cache [all] [[delete] key [value]|[name value help]]", Help: "删除、查看、修改或添加配置", Hand: func(c *Context, m *Message, key string, arg ...string) string {
			all := false // {{{
			if len(arg) > 0 && arg[0] == "all" {
				arg = arg[1:]
				all = true
			}

			m.BackTrace(func(m *Message) bool {
				switch len(arg) {
				case 0:
					for k, v := range m.Target.Caches {
						if m.Check(m.Target, "caches", k) {
							m.Echo("%s(%s): %s\n", k, m.Cap(k), v.Name)
						}
					}

				case 1:
					if v, ok := m.Target.Caches[arg[0]]; ok {
						if m.Check(m.Target, "caches", arg[0]) {
							m.Echo("%s: %s\n", v.Name, v.Help)
						}
					}
				case 2:
					switch arg[0] {
					case "delete":
						if m.Check(m.Target, "caches", arg[1]) {
							if _, ok := m.Target.Caches[arg[1]]; ok {
								delete(m.Target.Caches, arg[1])
							}
						}
					default:
						if _, ok := m.Target.Caches[arg[0]]; ok {
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
			Caches: map[string]*Cache{
				"nmessage": &Cache{},
			},
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

var Pulse = &Message{Code: 0, Time: time.Now(), Source: Index, Master: Index, Target: Index}

func init() {
	Pulse.Wait = make(chan bool, 10)
	Pulse.Root = Pulse
	Index.Root = Index
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

	for _, m := range Pulse.Search("") {
		m.Target.Begin(m)
	}

	for _, m := range Pulse.Search(Pulse.Conf("start")) {
		m.Put("option", "io", os.Stdout).Target.Start(m)
	}

	<-Pulse.Wait
	for Pulse.Capi("nserver") > 0 {
		<-Pulse.Wait
	}
}
