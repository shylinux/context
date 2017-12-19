package ctx // {{{
// }}}
import ( // {{{
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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
	Name string
	Help string

	Formats map[string]int
	Options map[string]string
	Appends map[string]string
	Hand    func(m *Message, c *Context, key string, arg ...string)
}

type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
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

	root     *Context
	context  *Context
	contexts map[string]*Context

	Master   *Context
	messages chan *Message

	Pulse    *Message
	Requests []*Message
	Historys []*Message
	Sessions map[string]*Message

	Index  map[string]*Context
	Groups map[string]*Context
	Owner  *Context
	Group  string

	Server
}

func (c *Context) Register(s *Context, x Server) *Context { // {{{
	if c.contexts == nil {
		c.contexts = make(map[string]*Context)
	}
	if x, ok := c.contexts[s.Name]; ok {
		panic(errors.New(c.Name + "上下文中已存在模块:" + x.Name))
	}

	c.contexts[s.Name] = s
	s.context = c
	s.Server = x
	return s
}

// }}}
func (c *Context) Spawn(m *Message, name string, help string) *Context { // {{{
	s := &Context{Name: name, Help: help, root: c.root, context: c}

	if m.Target = s; c.Server != nil {
		c.Register(s, c.Server.Spawn(m, s, m.Meta["detail"]...))
	} else {
		c.Register(s, nil)
	}

	if m.Template != nil {
		m.Template.Source = s
	}

	return s
}

// }}}
func (c *Context) Begin(m *Message) *Context { // {{{
	c.Caches["status"] = &Cache{Name: "服务状态(begin/start/close)", Value: "begin", Help: "服务状态，begin:初始完成，start:正在运行，close:未在运行"}
	c.Caches["stream"] = &Cache{Name: "服务数据", Value: "", Help: "服务数据"}

	m.Index = 1
	c.Pulse = m
	c.Requests = []*Message{m}
	c.Historys = []*Message{m}

	c.Master = m.Master.Master
	c.Owner = m.Master.Owner
	c.Group = m.Master.Group

	m.Log("begin", nil, "%d context %v", m.root.Capi("ncontext", 1), m.Meta["detail"])
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
	if c.Requests = append(c.Requests, m); m.Cap("status") != "start" {
		running := make(chan bool)
		go m.AssertOne(m, true, func(m *Message) {
			m.Log(m.Cap("status", "start"), nil, "%d server %v", m.root.Capi("nserver", 1), m.Meta["detail"])

			if running <- true; c.Server != nil && c.Server.Start(m, m.Meta["detail"]...) {
				c.Close(m, m.Meta["detail"]...)
			}
		})
		<-running
	}
	return true
}

// }}}
func (c *Context) Close(m *Message, arg ...string) bool { // {{{
	m.Log("close", c, "%v", arg)

	if m.Target == c {
		if m.Index == 0 {
			for i := len(c.Requests) - 1; i >= 0; i-- {
				v := c.Requests[i]
				if v.Index = -1; v.Source != c && !v.Source.Close(v, arg...) {
					v.Index = i
					return false
				}
				c.Requests = c.Requests[:i]
			}
		} else if m.Index > 0 {
			for i := m.Index - 1; i < len(c.Requests)-1; i++ {
				c.Requests[i] = c.Requests[i+1]
			}
			c.Requests = c.Requests[:len(c.Requests)-1]
		}
	}

	if c.Server != nil && !c.Server.Close(m, arg...) {
		return false
	}

	if m.Source == c && m.Target != c {
		if _, ok := c.Sessions[m.Name]; ok {
			delete(c.Sessions, m.Name)
		}
		return true
	}

	if len(c.Requests) > 1 {
		return false
	}

	if m.Cap("status") == "start" {
		m.Log(m.Cap("status", "close"), nil, "%d server %v", m.root.Capi("nserver", -1)+1, arg)
		for _, v := range c.Sessions {
			if v.Target != c {
				v.Target.Close(v, arg...)
			}
		}
	}

	if m.Index == 0 && c.context != nil && len(c.contexts) == 0 {
		m.Log("close", nil, "%d context %v", m.root.Capi("ncontext", -1)+1, arg)
		delete(c.context.contexts, c.Name)
	}
	return true
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
	code int
	time time.Time

	Recv chan bool
	Wait chan bool
	Meta map[string][]string
	Data map[string]interface{}

	messages []*Message
	message  *Message
	root     *Message

	Name   string
	Source *Context
	Master *Context
	Target *Context
	Index  int

	Template *Message
}

func (m *Message) Log(action string, ctx *Context, str string, arg ...interface{}) { // {{{
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
		if m.root.Conf("debug") != "on" {
			return
		}
	}

	if ctx == nil {
		ctx = m.Target
	}

	info := fmt.Sprintf("%s", ctx.Name)
	name := fmt.Sprintf("%s->%s", m.Source.Name, m.Target.Name)
	if m.Name != "" {
		name = fmt.Sprintf("%s:%s->%s.%d", m.Source.Name, m.Name, m.Target.Name, m.Index)
	}

	log.Printf("\033[%dm%d %s(%s) %s: %s\033[0m", color, m.code, action, name, info, fmt.Sprintf(str, arg...))
}

// }}}
func (m *Message) Check(s *Context, arg ...string) bool { // {{{
	if s.Owner == nil {
		return true
	}
	if m.Master.Owner == s.Owner {
		return true
	}
	if m.Master.Owner == s.root.Owner {
		return true
	}

	g, ok := s.Index[m.Master.Group]
	gg, gok := s.Index["void"]

	if len(arg) < 2 {
		if ok && g != nil {
			return true
		}

		m.Log("debug", s, "not auth: %s(%s)", m.Master.Name, m.Master.Group)
		if gok && gg != nil {
			return true
		}

		m.Log("debug", s, "not auth: %s(void)", m.Master.Name)
		return false
	}

	ok, gok = false, false
	switch arg[0] {
	case "commands":
		if g != nil {
			_, ok = g.Commands[arg[1]]
		}
		if gg != nil {
			_, gok = gg.Commands[arg[1]]
		}
	case "configs":
		if g != nil {
			_, ok = g.Configs[arg[1]]
		}
		if gg != nil {
			_, gok = gg.Configs[arg[1]]
		}
	case "caches":
		if g != nil {
			_, ok = g.Caches[arg[1]]
		}
		if gg != nil {
			_, gok = gg.Caches[arg[1]]
		}
	}

	if ok {
		return true
	}
	if g != nil {
		m.Log("debug", s, "%s:%s not auth: %s(%s)", arg[0], arg[1], m.Master.Name, m.Master.Group)
	}
	if gok {
		return true
	}
	m.Log("debug", s, "%s:%s not auth: %s(void)", arg[0], arg[1], m.Master.Name)
	return false
}

// }}}
func (m *Message) Assert(e interface{}, msg ...string) bool { // {{{
	switch e := e.(type) {
	case error:
	case bool:
		if e {
			return true
		}
	case string:
		if e != "error: " {
			return true
		}
	case *Context:
		if m.Check(e, msg...) {
			return true
		}
		if len(msg) > 2 {
			msg = msg[2:]
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

	m.Set("result", "error: ", fmt.Sprintln(e), "\n")
	panic(e)
}

// }}}
func (m *Message) AssertOne(msg *Message, safe bool, hand ...func(msg *Message)) *Message { // {{{
	defer func() {
		if e := recover(); e != nil {
			msg.Log("error", nil, "error: %v", e)
			if msg.root.Conf("debug") == "on" && e != io.EOF {
				fmt.Printf("\n\033[31m%s error: %v\033[0m\n", msg.Target.Name, e)
				debug.PrintStack()
				fmt.Printf("\033[31m%s error: %v\033[0m\n\n", msg.Target.Name, e)
			}

			if e == io.EOF {
				return
			} else if len(hand) > 1 {
				m.AssertOne(msg, safe, hand[1:]...)
			} else if !safe {
				msg.Assert(e)
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
		code:    m.root.Capi("nmessage", 1),
		time:    time.Now(),
		message: m,
		root:    m.root,
		Source:  m.Target,
		Master:  m.Target,
		Target:  c,
	}

	if m.messages == nil {
		m.messages = make([]*Message, 0, 10)
	}
	m.messages = append(m.messages, msg)

	if len(key) == 0 {
		return msg
	}

	if msg.Source.Sessions == nil {
		msg.Source.Sessions = make(map[string]*Message)
	}
	msg.Source.Sessions[key[0]] = msg
	msg.Name = key[0]
	return msg
}

// }}}
func (m *Message) Reply(key ...string) *Message { // {{{
	if m.Template == nil {
		m.Template = m.Spawn(m.Source, key...)
	}

	msg := m.Template
	if len(key) == 0 {
		return msg
	}

	if msg.Source.Sessions == nil {
		msg.Source.Sessions = make(map[string]*Message)
	}
	msg.Source.Sessions[key[0]] = msg
	msg.Name = key[0]
	return msg
}

// }}}
func (m *Message) Format() string { // {{{
	name := fmt.Sprintf("%s->%s", m.Source.Name, m.Target.Name)
	if m.Name != "" {
		name = fmt.Sprintf("%s.%s->%s.%d", m.Source.Name, m.Name, m.Target.Name, m.Index)
	}
	return fmt.Sprintf("%d(%s): %s %v", m.code, name, m.time.Format("15:04:05"), m.Meta["detail"])
}

// }}}

func (m *Message) BackTrace(hand func(m *Message) bool) { // {{{
	target := m.Target
	for s := target; s != nil; s = s.context {
		if m.Target = s; m.Check(s) && !hand(m) {
			break
		}
	}
	m.Target = target
}

// }}}
func (m *Message) Travel(c *Context, hand func(m *Message) bool) { // {{{
	if c == nil {
		c = m.Target
	}
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
func (m *Message) Search(key string, root ...bool) []*Message { // {{{
	reg, e := regexp.Compile(key)
	m.Assert(e)

	target := m.Target
	if len(root) > 0 && root[0] {
		target = m.Target.root
	}

	cs := make([]*Context, 0, 3)
	m.Travel(target, func(m *Message) bool {
		if reg.MatchString(m.Target.Name) || reg.FindString(m.Target.Help) != "" {
			m.Log("search", nil, "%d match [%s]", len(cs)+1, key)
			cs = append(cs, m.Target)
		}
		return true
	})

	ms := make([]*Message, len(cs))
	for i := 0; i < len(cs); i++ {
		ms[i] = m.Spawn(cs[i])
	}

	return ms
}

// }}}
func (m *Message) Find(name string, root ...bool) *Message { // {{{
	target := m.Target
	if len(root) > 0 && root[0] {
		target = m.Target.root
	}

	cs := target.contexts
	for _, v := range strings.Split(name, ".") {
		if x, ok := cs[v]; ok {
			target, cs = x, x.contexts
		} else {
			m.Log("find", target, "not find %s", v)
			return nil
		}
	}
	m.Log("find", nil, "find %s", name)
	return m.Spawn(target)
}

// }}}
func (m *Message) Start(name string, help string, arg ...string) bool { // {{{
	return m.Set("detail", arg...).Target.Spawn(m, name, help).Begin(m).Start(m)
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
			m.Meta[key] = make([]string, 0, 3)
			m.Meta[meta] = append(m.Meta[meta], key)
		}
		m.Meta[key] = append(m.Meta[key], value...)
	default:
		m.Assert(false, "消息参数错误")
	}

	return m
}

// }}}
func (m *Message) Set(meta string, arg ...string) *Message { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if _, ok := m.Meta[meta]; !ok {
		m.Meta[meta] = make([]string, 0, 3)
	}

	switch meta {
	case "detail", "result":
		m.Meta[meta] = arg
	case "option", "append":
		if len(arg) > 0 {
			if _, ok := m.Meta[arg[0]]; !ok {
				m.Meta[meta] = append(m.Meta[meta], arg[0])
			}
			m.Meta[arg[0]] = arg[1:]
		} else {
			for _, k := range m.Meta[meta] {
				delete(m.Meta, k)
				delete(m.Data, k)
			}
			delete(m.Meta, meta)
		}
	default:
		m.Assert(false, "消息参数错误")
	}

	return m
}

// }}}
func (m *Message) Put(meta string, key string, value interface{}) *Message { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}

	switch meta {
	case "option", "append":
		if m.Data == nil {
			m.Data = make(map[string]interface{})
		}
		if _, ok := m.Meta[meta]; !ok {
			m.Meta[meta] = make([]string, 0, 3)
		}
		if _, ok := m.Data[key]; !ok {
			m.Meta[meta] = append(m.Meta[meta], key)
		}
		m.Data[key] = value
	default:
		m.Assert(false, "消息参数错误")
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
	if meta, ok := m.Meta[key]; ok && len(meta) > 0 {
		return meta[0]
	}
	return ""
}

// }}}
func (m *Message) Echo(str string, arg ...interface{}) *Message { // {{{
	return m.Add("result", fmt.Sprintf(str, arg...))
}

// }}}
func (m *Message) End(s bool) { // {{{
	if m.Wait != nil {
		m.Wait <- s
	}
}

// }}}

func (m *Message) Exec(key string, arg ...string) string { // {{{

	for _, c := range []*Context{m.Target, m.Target.Master, m.Target.Owner, m.Source, m.Source.Master, m.Source.Owner} {
		for s := c; s != nil; s = s.context {

			m.Master = m.Source
			if x, ok := s.Commands[key]; ok && x.Hand != nil && m.Check(c, "commands", key) {
				m.AssertOne(m, true, func(m *Message) {
					m.Log("cmd", s, "%s %v", key, arg)

					if x.Options != nil {
						for _, v := range m.Meta["option"] {
							if _, ok := x.Options[v]; !ok {
								panic(errors.New(fmt.Sprintf("未知参数: %s", v)))
							}
						}
					}

					if x.Formats != nil {
						for i := 0; i < len(arg); i++ {
							n, ok := x.Formats[arg[i]]
							if !ok {
								m.Add("option", "args", arg[i])
								continue
							}

							if n < 0 {
								n += len(arg) - i
							}

							if x, ok := m.Meta[arg[i]]; ok && len(x) == n {
								m.Add("option", "args", arg[i])
								continue
							}

							m.Add("option", arg[i], arg[i+1:i+1+n]...)
							i += n
						}
						arg = m.Meta["args"]
					}

					x.Hand(m.Set("result").Set("append"), s, key, arg...)

					if x.Appends != nil {
						for _, v := range m.Meta["append"] {
							if _, ok := x.Appends[v]; !ok {
								panic(errors.New(fmt.Sprintf("未知参数: %s", v)))
							}
						}
					}

					if c.Historys == nil {
						c.Historys = make([]*Message, 0, 10)
					}
					c.Historys = append(c.Historys, m)
				})

				return m.Get("result")
			}
		}
	}

	m.Set("result", "error: ", "命令不存在")
	return ""
}

// }}}
func (m *Message) Deal(pre func(msg *Message, arg ...string) bool, post func(msg *Message, arg ...string) bool) { // {{{
	if m.Target.messages == nil {
		m.Target.messages = make(chan *Message, m.Confi("MessageQueueSize"))
	}

	for run := true; run; {
		m.AssertOne(<-m.Target.messages, true, func(msg *Message) {
			defer msg.End(true)

			if len(msg.Meta["detail"]) == 0 {
				return
			}

			if pre != nil && !pre(msg, msg.Meta["detail"]...) {
				run = false
				return
			}

			msg.Exec(msg.Meta["detail"][0], msg.Meta["detail"][1:]...)

			if post != nil && !post(msg, msg.Meta["result"]...) {
				run = false
				return
			}
		})
	}
}

// }}}
func (m *Message) Post(s *Context) string { // {{{
	if s == nil {
		s = m.Target.Master
	}

	m.Assert(s.messages != nil, s.Name+" 没有开启消息处理")

	if s.messages <- m; m.Wait != nil {
		<-m.Wait
	}

	return m.Get("result")
}

// }}}

func (m *Message) Cmd(arg ...string) string { // {{{
	if len(arg) > 0 {
		m.Set("detail", arg...)
	}

	if s := m.Target.Master; s != nil && s != m.Source.Master {
		return m.Post(s)
	}

	return m.Exec(m.Meta["detail"][0], m.Meta["detail"][1:]...)
}

// }}}
func (m *Message) Confi(key string, arg ...int) int { // {{{
	n, e := strconv.Atoi(m.Conf(key))
	m.Assert(e)

	if len(arg) > 0 {
		n, e = strconv.Atoi(m.Conf(key, fmt.Sprintf("%d", arg[0])))
		m.Assert(e)
	}

	return n
}

// }}}
func (m *Message) Conf(key string, arg ...string) string { // {{{
	var hand func(m *Message, x *Config, arg ...string) string
	for s := m.Target; s != nil; s = s.context {
		if x, ok := s.Configs[key]; ok {
			if !m.Check(s, "configs", key) {
				continue
			}

			switch len(arg) {
			case 3:
				if hand == nil {
					hand = x.Hand
				}
			case 1:
				m.Log("conf", s, "%s %v", key, arg)

				if x.Hand != nil {
					x.Value = x.Hand(m, x, arg[0])
				} else {
					x.Value = arg[0]
				}
				return x.Value
			case 0:
				if x.Hand != nil {
					return x.Hand(m, x)
				}
				return x.Value
			default:
				panic(errors.New(key + "配置项参数错误"))
			}
		}
	}

	if len(arg) == 3 && m.Check(m.Target, "configs", key) {
		if m.Target.Configs == nil {
			m.Target.Configs = make(map[string]*Config)
		}

		m.Target.Configs[key] = &Config{Name: arg[0], Value: arg[1], Help: arg[2], Hand: hand}
		m.Log("conf", nil, "%s %v", key, arg)
		return m.Conf(key, arg[1])
	}

	m.Assert(false, key+"配置项操作错误")
	return ""
}

// }}}
func (m *Message) Capi(key string, arg ...int) int { // {{{
	n, e := strconv.Atoi(m.Cap(key))
	m.Assert(e)

	if len(arg) > 0 {
		n, e = strconv.Atoi(m.Cap(key, fmt.Sprintf("%d", arg[0]+n)))
		m.Assert(e)
	}

	return n
}

// }}}
func (m *Message) Cap(key string, arg ...string) string { // {{{
	var hand func(m *Message, x *Cache, arg ...string) string
	for s := m.Target; s != nil; s = s.context {
		if x, ok := s.Caches[key]; ok {
			if !m.Check(s, "caches", key) {
				continue
			}

			switch len(arg) {
			case 3:
				if hand == nil {
					hand = x.Hand
				}
			case 1:
				if x.Hand != nil {
					x.Value = x.Hand(m, x, arg[0])
				} else {
					x.Value = arg[0]
				}
				return x.Value
			case 0:
				if x.Hand != nil {
					return x.Hand(m, x)
				}
				return x.Value
			default:
				panic(errors.New(key + "缓存项参数错误"))
			}
		}
	}

	if len(arg) == 3 && m.Check(m.Target, "caches", key) {
		if m.Target.Caches == nil {
			m.Target.Caches = make(map[string]*Cache)
		}

		m.Target.Caches[key] = &Cache{Name: arg[0], Value: arg[1], Help: arg[2], Hand: hand}
		m.Log("cap", nil, "%s %v", key, arg)
		return m.Cap(key, arg[1])
	}

	m.Assert(false, key+"缓存项操作错误")
	return ""

}

// }}}

var Pulse = &Message{code: 0, time: time.Now(), Wait: make(chan bool), Source: Index, Master: Index, Target: Index}
var Index = &Context{Name: "ctx", Help: "模块中心",
	Caches: map[string]*Cache{
		"nserver":  &Cache{Name: "服务数量", Value: "0", Help: "显示已经启动运行模块的数量"},
		"ncontext": &Cache{Name: "模块数量", Value: "0", Help: "显示功能树已经注册模块的数量"},
		"nmessage": &Cache{Name: "消息数量", Value: "0", Help: "显示模块启动时所创建消息的数量"},
	},
	Configs: map[string]*Config{
		"default": &Config{Name: "默认的搜索起点(root/back/home)", Value: "root", Help: "模块搜索的默认起点，root:从根模块，back:从父模块，home:从当前模块"},

		"start":   &Config{Name: "启动模块", Value: "cli", Help: "启动时自动运行的模块"},
		"init.sh": &Config{Name: "启动脚本", Value: "etc/init.sh", Help: "模块启动时自动运行的脚本"},
		"bench.log": &Config{Name: "日志文件", Value: "var/bench.log", Help: "模块日志输出的文件", Hand: func(m *Message, x *Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				if e := os.MkdirAll(path.Dir(arg[0]), os.ModePerm); e == nil {
					if l, e := os.Create(x.Value); e == nil {
						log.SetOutput(l)
					}
				}
				return arg[0]
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
				return arg[0]
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
			Hand: func(m *Message, c *Context, key string, arg ...string) {
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
				// }}}
			}},
		"server": &Command{Name: "server [start|exit|switch][args]", Help: "服务启动停止切换", Hand: func(m *Message, c *Context, key string, arg ...string) {
			switch len(arg) { // {{{
			case 0:
				m.Travel(m.Target.root, func(m *Message) bool {
					if x, ok := m.Target.Caches["status"]; ok {
						m.Echo("%s(%s): %s\n", m.Target.Name, x.Value, m.Target.Help)
					}
					return true
				})

			default:
				switch arg[0] {
				case "start":
					m.Meta = nil
					m.Set("detail", arg[1:]...).Target.Start(m)
				case "stop":
					m.Set("detail", arg[1:]...).Target.Close(m)
				case "switch":
				}
			}
			// }}}
		}},
		"message": &Command{Name: "message code", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			switch len(arg) { // {{{
			case 0:
				m.Echo("\033[31mrequests:\033[0m\n")
				for i, v := range m.Target.Requests {
					m.Echo("%d %s\n", i, v.Format())
					for i, v := range v.messages {
						m.Echo("  %d %s\n", i, v.Format())
					}
				}

				m.Echo("\033[32msessions:\033[0m\n")
				for k, v := range m.Target.Sessions {
					m.Echo("%s %s\n", k, v.Format())
				}

				m.Echo("\033[33mhistorys:\033[0m\n")
				for i, v := range m.Target.Historys {
					m.Echo("%d %s\n", i, v.Format())
					for i, v := range v.messages {
						m.Echo("  %d %s\n", i, v.Format())
					}
				}
			case 1:
				n, e := strconv.Atoi(arg[0])
				m.Assert(e)

				ms := []*Message{m.root}
				for i := 0; i < len(ms); i++ {
					if ms[i].code == n {
						if ms[i].message != nil {
							m.Echo("message: %d\n", ms[i].message.code)
						}

						m.Echo("%s\n", ms[i].Format())
						if len(ms[i].Meta["option"]) > 0 {
							m.Echo("option: %v\n", ms[i].Meta["option"])
						}
						for _, k := range ms[i].Meta["option"] {
							m.Echo("  %s: %v\n", k, ms[i].Meta[k])
						}

						if len(ms[i].Meta["result"]) > 0 {
							m.Echo("result: %v\n", ms[i].Meta["result"])
						}
						if len(ms[i].Meta["append"]) > 0 {
							m.Echo("append: %v\n", ms[i].Meta["append"])
						}
						for _, k := range ms[i].Meta["append"] {
							m.Echo("  %s: %v\n", k, ms[i].Meta[k])
						}

						if len(ms[i].messages) > 0 {
							m.Echo("messages: %d\n", len(ms[i].messages))
						}
						for _, v := range ms[i].messages {
							m.Echo("  %s\n", v.Format())
						}
						break
					}
					ms = append(ms, ms[i].messages...)
				}
			}

			// }}}
		}},
		"context": &Command{Name: "context [root] [[find|search] name] [list|show|spawn|start|switch|close][args]", Help: "查找并操作模块，\n查找起点root:根模块、back:父模块、home:本模块，\n查找方法find:路径匹配、search:模糊匹配，\n查找对象name:支持点分和正则，\n操作类型show:显示信息、switch:切换为当前、start:启动模块、spawn:分裂子模块，args:启动参数",
			Formats: map[string]int{"root": 0, "back": 0, "home": 0, "find": 1, "search": 1, "list": 0, "show": 0, "close": 0, "switch": 0, "start": 0, "spawn": 0},
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				root := true || m.Has("root") // {{{

				ms := []*Message{}
				switch {
				case m.Has("search"):
					if s := m.Search(m.Get("search"), root); len(s) > 0 {
						ms = append(ms, s...)
					}
				case m.Has("find"):
					if msg := m.Find(m.Get("find"), root); msg != nil {
						ms = append(ms, msg)
					}
				case m.Has("args"):
					if s := m.Search(m.Get("args"), root); len(s) > 0 {
						ms = append(ms, s...)
						arg = arg[1:]
						break
					}
					fallthrough
				default:
					ms = append(ms, m)
				}

				for _, v := range ms {
					m.Target = v.Target

					switch {
					case m.Has("switch"):
					case m.Has("spawn"):
						v.Set("detail", arg[2:]...).Target.Spawn(v, arg[0], arg[1]).Begin(v)
					case m.Has("start"):
						v.Set("detail", arg...).Target.Start(v)
					case m.Has("close"):
						v.Target.Close(v)
					case m.Has("show"):
						m.Echo("%s(%s): %s\n", v.Target.Name, v.Target.Owner.Name, v.Target.Help)
						if len(v.Target.Requests) > 0 {
							m.Echo("模块资源：\n")
							for i, v := range v.Target.Requests {
								m.Echo("  %d: <- %s %s\n", i, v.Source.Name, v.Meta["detail"])
								// for i, v := range v.Messages {
								// 	m.Echo("    %d: -> %s %s\n", i, v.Source.Name, v.Meta["detail"])
								// }
							}
						}
						if len(v.Target.Sessions) > 0 {
							m.Echo("模块引用：\n")
							for k, v := range v.Target.Sessions {
								m.Echo("  %s: -> %s %v\n", k, v.Target.Name, v.Meta["detail"])
							}
						}
					case m.Has("list") || len(m.Meta["detail"]) == 1:
						m.Travel(v.Target, func(msg *Message) bool {
							target := msg.Target
							m.Echo("%s(", target.Name)

							if target.context != nil {
								m.Echo("%s", target.context.Name)
							}
							m.Echo(":")

							if target.Master != nil {
								m.Echo("%s", target.Master.Name)
							}
							m.Echo(":")

							if target.Owner != nil {
								m.Echo("%s", target.Owner.Name)
							}
							m.Echo(":")

							msg.Target = msg.Target.Owner
							if msg.Target != nil && msg.Check(msg.Target, "caches", "username") && msg.Check(msg.Target, "caches", "group") {
								m.Echo("%s:%s", msg.Cap("username"), msg.Cap("group"))
							}
							m.Echo("): ")
							msg.Target = target

							if msg.Check(msg.Target, "caches", "status") && msg.Check(msg.Target, "caches", "stream") {
								m.Echo("%s(%s) ", msg.Cap("status"), msg.Cap("stream"))
							}
							m.Echo("%s\n", target.Help)
							return true
						})
					case len(arg) > 0 && v != m:
						v.Set("detail", arg...).Cmd()
					}
				}
				// }}}
			}},
		"command": &Command{Name: "command [all] [key [name help]]", Help: "查看或修改命令",
			Formats: map[string]int{"all": 0, "delete": 0, "void": 0},
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				all := m.Has("all") // {{{

				switch len(arg) {
				case 0:
					m.BackTrace(func(m *Message) bool {
						if all {
							m.Echo("%s comands:\n", m.Target.Name)
						}
						for k, x := range m.Target.Commands {
							if m.Check(m.Target, "commands", k) {
								if all {
									m.Echo("  ")
								}
								m.Echo("%s: %s\n", k, x.Name)
							}
						}
						return all
					})
				case 1:
					switch {
					case m.Has("delete"):
						if _, ok := m.Target.Commands[arg[0]]; ok {
							if m.Target.Owner == nil || m.Master.Owner == m.Target.Owner {
								delete(m.Target.Commands, arg[0])
							}
						}
					case m.Has("void"):
						if x, ok := m.Target.Commands[arg[0]]; ok {
							if m.Target.Owner == nil || m.Master.Owner == m.Target.Owner {
								x.Hand = nil
							}
						}
					}

					m.BackTrace(func(m *Message) bool {
						if all {
							m.Echo("%s commands:\n", m.Target.Name)
						}
						if x, ok := m.Target.Commands[arg[0]]; ok {
							if all {
								m.Echo("  ")
							}
							if m.Check(m.Target, "commands", arg[0]) {
								m.Echo("%s\n    %s\n", x.Name, x.Help)
							}
						}
						return all
					})
				case 3:
					cmd := &Command{}
					m.BackTrace(func(m *Message) bool {
						if x, ok := m.Target.Commands[arg[0]]; ok && x.Hand != nil {
							*cmd = *x
						}
						return all
					})

					if m.Check(m.Target, "commands", arg[0]) {
						if x, ok := m.Target.Commands[arg[0]]; ok {
							if m.Target.Owner == nil || m.Master.Owner == m.Target.Owner {
								x.Name = arg[1]
								x.Help = arg[2]
								m.Echo("%s\n    %s\n", x.Name, x.Help)
							}
						} else {
							if m.Target.Commands == nil {
								m.Target.Commands = map[string]*Command{}
							}
							cmd.Name = arg[1]
							cmd.Help = arg[2]
							m.Target.Commands[arg[0]] = cmd
						}
					}
				}
				// }}}
			}},
		"config": &Command{Name: "config [all] [[delete|void] key [value]|[name value help]]", Help: "删除、置空、查看、修改或添加配置",
			Formats: map[string]int{"all": 0, "delete": 0, "void": 0},
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				all := m.Has("all") // {{{

				switch len(arg) {
				case 0:
					m.BackTrace(func(m *Message) bool {
						if all {
							m.Echo("%s configs:\n", m.Target.Name)
						}
						for k, x := range m.Target.Configs {
							if m.Check(m.Target, "configs", k) {
								if all {
									m.Echo("  ")
								}
								m.Echo("%s(%s): %s\n", k, x.Value, x.Name)
							}
						}
						return all
					})
				case 1:
					switch {
					case m.Has("delete"):
						if _, ok := m.Target.Configs[arg[0]]; ok {
							if m.Target.Owner == nil || m.Master.Owner == m.Target.Owner {
								delete(m.Target.Configs, arg[0])
							}
						}
					case m.Has("void"):
						m.Conf(arg[0], "")
					}

					m.BackTrace(func(m *Message) bool {
						if all {
							m.Echo("%s config:\n", m.Target.Name)
						}
						if x, ok := m.Target.Configs[arg[0]]; ok {
							if m.Check(m.Target, "configs", arg[0]) {
								if all {
									m.Echo("  ")
								}
								m.Echo("%s: %s\n", x.Name, x.Help)
							}
						}
						return all
					})

				case 2:
					m.Conf(arg[0], arg[1])
				case 4:
					m.Conf(arg[0], arg[1:]...)
				}
				// }}}
			}},
		"cache": &Command{Name: "cache [all] [[delete|void] key [value]|[name value help]]", Help: "删除、置空、查看、修改或添加缓存",
			Formats: map[string]int{"all": 0, "delete": 0, "void": 0},
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				all := m.Has("all") // {{{

				switch len(arg) {
				case 0:
					m.BackTrace(func(m *Message) bool {
						if all {
							m.Echo("%s configs:\n", m.Target.Name)
						}
						for k, x := range m.Target.Caches {
							if m.Check(m.Target, "caches", k) {
								if all {
									m.Echo("  ")
								}
								m.Echo("%s(%s): %s\n", k, m.Cap(k), x.Name)
							}
						}
						return all
					})

				case 1:
					switch {
					case m.Has("delete"):
						if _, ok := m.Target.Caches[arg[0]]; ok {
							if m.Target.Owner == nil || m.Master.Owner == m.Target.Owner {
								delete(m.Target.Caches, arg[0])
							}
						}
					case m.Has("void"):
						m.Cap(arg[0], "")
					}

					m.BackTrace(func(m *Message) bool {
						if all {
							m.Echo("%s config:\n", m.Target.Name)
						}
						if x, ok := m.Target.Caches[arg[0]]; ok {
							if m.Check(m.Target, "caches", arg[0]) {
								if all {
									m.Echo("  ")
								}
								m.Echo("%s: %s\n", x.Name, x.Help)
							}
						}
						return all
					})
				case 2:
					m.Cap(arg[0], arg[1])
				case 4:
					m.Cap(arg[0], arg[1:]...)
				}
				// }}}
			}},
	},
	Index: map[string]*Context{
		"void": &Context{Name: "void",
			Caches:  map[string]*Cache{},
			Configs: map[string]*Config{},
			Commands: map[string]*Command{
				"message": &Command{},
				"context": &Command{},
				"command": &Command{},
				"config":  &Command{},
				"cache":   &Command{},
			},
		},
	},
}

func init() {
	Pulse.root = Pulse
	Index.root = Index
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

	log.Println("\n\n")
	Index.Group = "root"
	Index.Owner = Index.contexts["aaa"]
	Index.Master = Index.contexts["cli"]
	for _, m := range Pulse.Search("") {
		m.Target.root = Index
		m.Target.Begin(m)
	}
	Index.Requests = append(Index.Requests, Pulse)
	log.Println()

	for _, m := range Pulse.Search(Pulse.Conf("start")) {
		m.Put("option", "io", os.Stdout).Target.Start(m)
	}

	for <-Pulse.Wait; Pulse.Capi("nserver") > 0; <-Pulse.Wait {
	}
}
