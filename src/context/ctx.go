package ctx // {{{
// }}}
import ( // {{{

	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// }}}

func Right(str string) bool {
	return str != "" && str != "0" && str != "false"
}

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

	master   *Context
	messages chan *Message

	Pulse    *Message
	Requests []*Message
	Historys []*Message
	Sessions map[string]*Message
	Exit     chan bool

	Index    map[string]*Context
	Groups   map[string]*Context
	Owner    *Context
	Group    string
	password string

	Server
}

func (c *Context) Password(meta string) string { // {{{
	bs := md5.Sum([]byte(fmt.Sprintln("%d%d%s", time.Now().Unix(), rand.Int(), meta)))
	sessid := hex.EncodeToString(bs[:])
	return sessid
}

// }}}
func (c *Context) Register(s *Context, x Server) (password string) { // {{{
	if c.contexts == nil {
		c.contexts = make(map[string]*Context)
	}
	if x, ok := c.contexts[s.Name]; ok {
		panic(errors.New(c.Name + "上下文中已存在模块:" + x.Name))
	}

	c.contexts[s.Name] = s
	s.context = c
	s.Server = x
	s.password = c.Password(s.Name)
	return s.password
}

// }}}
func (c *Context) Spawn(m *Message, name string, help string) *Context { // {{{
	s := &Context{Name: name, Help: help, root: c.root, context: c}

	if m.target = s; c.Server != nil {
		c.Register(s, c.Server.Spawn(m, s, m.Meta["detail"]...))
	} else {
		c.Register(s, nil)
	}

	if m.Template != nil {
		m.Template.source = s
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
	c.Sessions = map[string]*Message{}

	c.master = m.master.master
	c.Owner = m.master.Owner
	c.Group = m.master.Group

	m.Log("begin", nil, "%d context %v %v", m.root.Capi("ncontext", 1), m.Meta["detail"], m.Meta["option"])
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
	m.Hand = true

	if m != c.Requests[0] {
		c.Requests, m.Index = append(c.Requests, m), len(c.Requests)+1
	}

	if m.Cap("status") != "start" {
		running := make(chan bool)
		go m.AssertOne(m, true, func(m *Message) {
			m.Log(m.Cap("status", "start"), nil, "%d server %v %v", m.root.Capi("nserver", 1), m.Meta["detail"], m.Meta["option"])

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
	m.Log("close", c, "%d:%d %v", len(m.source.Sessions), len(m.target.Historys), arg)

	if m.target == c {
		if m.Index == 0 {
			for i := len(c.Requests) - 1; i >= 0; i-- {
				v := c.Requests[i]
				if v.Index = -1; v.source != c && !v.source.Close(v, arg...) {
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

	if m.source == c && m.target != c {
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
			if v.target != c {
				v.target.Close(v, arg...)
			}
		}
	}

	// if m.Index == 0 && c.context != nil && len(c.contexts) == 0 {
	if c.context != nil {
		m.Log("close", nil, "%d context %v", m.root.Capi("ncontext", -1)+1, arg)
		delete(c.context.contexts, c.Name)
		c.context = nil
		if c.Exit != nil {
			c.Exit <- true
		}
	}
	return true
}

// }}}

func (c *Context) Context() *Context { // {{{
	return c.context
}

// }}}
func (c *Context) Master(s ...*Context) *Context { // {{{
	if len(s) > 0 {
		switch s[0] {
		case nil, c:
			c.master = s[0]
		}
	}
	return c.master
}

// }}}
func (c *Context) Has(key ...string) bool { // {{{
	switch len(key) {
	case 1:
		if _, ok := c.Caches[key[0]]; ok {
			return true
		}
		if _, ok := c.Configs[key[0]]; ok {
			return true
		}
	case 2:
		if _, ok := c.Caches[key[0]]; ok && key[1] == "cache" {
			return true
		}
		if _, ok := c.Configs[key[0]]; ok && key[1] == "config" {
			return true
		}
		if _, ok := c.Commands[key[0]]; ok && key[1] == "command" {
			return true
		}
	}
	return false
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
	time time.Time
	code int
	Hand bool

	Recv chan bool
	Wait chan bool
	Meta map[string][]string
	Data map[string]interface{}

	messages []*Message
	message  *Message
	root     *Message

	Name   string
	source *Context
	master *Context
	target *Context
	Index  int

	Template *Message
}

func (m *Message) Code() int { // {{{
	return m.code
}

// }}}
func (m *Message) Message() *Message { // {{{
	return m.message
}

// }}}
func (m *Message) Source(s ...*Context) *Context { // {{{
	if len(s) > 0 {
		m.source = s[0]
	}
	return m.source
}

// }}}
func (m *Message) Master(s ...*Context) *Context { // {{{
	if len(s) > 0 && s[0] == m.source {
		m.master = m.source
	}
	return m.master
}

// }}}
func (m *Message) Target(s ...*Context) *Context { // {{{
	if len(s) > 0 {
		m.target = s[0]
	}
	return m.target
}

// }}}
var i = 0

func (m *Message) Log(action string, ctx *Context, str string, arg ...interface{}) { // {{{
	if !m.Confs("bench.log") {
		return
	}

	if !m.Options("log") {
		return
	}

	if l := m.Sess("log"); l != nil {
		if i++; i > 20000 {
			debug.PrintStack()
			os.Exit(1)
		}
		// l.Wait = nil
		l.Options("log", false)
		l.Cmd("log", action, fmt.Sprintf(str, arg...))
	} else {
		log.Printf(str, arg...)
		return
	}
}

// }}}
func (m *Message) Gdb(action string) { // {{{

}

// }}}
func (m *Message) Check(s *Context, arg ...string) bool { // {{{
	if s.Owner == nil {
		return true
	}
	if m.master.Owner == s.Owner {
		return true
	}
	if m.master.Owner == s.root.Owner {
		return true
	}

	g, ok := s.Index[m.master.Group]
	gg, gok := s.Index["void"]

	if len(arg) < 2 {
		if ok && g != nil {
			return true
		}

		m.Log("debug", s, "not auth: %s(%s)", m.master.Name, m.master.Group)
		if gok && gg != nil {
			return true
		}

		m.Log("debug", s, "not auth: %s(void)", m.master.Name)
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
		m.Log("debug", s, "%s:%s not auth: %s(%s)", arg[0], arg[1], m.master.Name, m.master.Group)
	}
	if gok {
		return true
	}
	m.Log("debug", s, "%s:%s not auth: %s(void)", arg[0], arg[1], m.master.Name)
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
	case *Message:
		os.Exit(1)
		panic(e)
	default:
		return true
	}

	if len(msg) > 1 {
		arg := make([]interface{}, 0, len(msg)-1)
		for _, m := range msg[1:] {
			arg = append(arg, m)
		}
		e = errors.New(fmt.Sprintf(msg[0], arg...))
	} else if len(msg) > 0 {
		e = errors.New(msg[0])
	}
	if _, ok := e.(error); !ok {
		e = errors.New("error")
	}

	m.Set("result", "error: ", fmt.Sprintln(e), "\n")
	os.Exit(1)
	panic(e)
}

// }}}
func (m *Message) AssertOne(msg *Message, safe bool, hand ...func(msg *Message)) *Message { // {{{
	defer func() {
		if e := recover(); e != nil {
			switch e.(type) {
			case *Message:
				os.Exit(1)
				panic(e)
			}

			msg.Log("error", nil, "error: %v", e)
			if msg.root.Conf("debug") == "on" && e != io.EOF {
				fmt.Printf("\n\033[31m%s error: %v\033[0m\n", msg.target.Name, e)
				debug.PrintStack()
				fmt.Printf("\033[31m%s error: %v\033[0m\n\n", msg.target.Name, e)
			}

			if len(hand) > 1 {
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
		source:  m.target,
		master:  m.target,
		target:  c,
	}

	if m.messages == nil {
		m.messages = make([]*Message, 0, 10)
	}
	m.messages = append(m.messages, msg)

	msg.Wait = make(chan bool)
	if len(key) == 0 {
		return msg
	}

	if msg.source.Sessions == nil {
		msg.source.Sessions = make(map[string]*Message)
	}
	msg.source.Sessions[key[0]] = msg
	msg.Name = key[0]
	return msg
}

// }}}
func (m *Message) Reply(key ...string) *Message { // {{{
	if m.Template == nil {
		m.Template = m.Spawn(m.source, key...)
	}

	msg := m.Template
	if len(key) == 0 {
		return msg
	}

	if msg.source.Sessions == nil {
		msg.source.Sessions = make(map[string]*Message)
	}
	msg.source.Sessions[key[0]] = msg
	msg.Name = key[0]
	return msg
}

// }}}
func (m *Message) Format() string { // {{{
	name := fmt.Sprintf("%s->%s", m.source.Name, m.target.Name)
	if m.Name != "" {
		name = fmt.Sprintf("%s.%s->%s.%d", m.source.Name, m.Name, m.target.Name, m.Index)
	}
	return fmt.Sprintf("%d(%s): %s %v", m.code, name, m.time.Format("15:04:05"), m.Meta["detail"])
}

// }}}

func (m *Message) BackTrace(hand func(m *Message) bool) { // {{{
	target := m.target
	for s := target; s != nil; s = s.context {
		if m.target = s; m.Check(s) && !hand(m) {
			break
		}
	}
	m.target = target
}

// }}}
func (m *Message) Travel(c *Context, hand func(m *Message) bool) { // {{{
	if c == nil {
		c = m.target
	}
	target := m.target

	cs := []*Context{c}
	for i := 0; i < len(cs); i++ {
		if m.target = cs[i]; m.Check(cs[i]) && !hand(m) {
			break
		}

		for _, v := range cs[i].contexts {
			cs = append(cs, v)
		}
	}

	m.target = target
}

// }}}
func (m *Message) Search(key string, root ...bool) []*Message { // {{{
	reg, e := regexp.Compile(key)
	m.Assert(e)

	target := m.target
	if len(root) > 0 && root[0] {
		target = m.target.root
	}

	cs := make([]*Context, 0, 3)
	m.Travel(target, func(m *Message) bool {
		if reg.MatchString(m.target.Name) || reg.FindString(m.target.Help) != "" {
			m.Log("search", nil, "%d match [%s]", len(cs)+1, key)
			cs = append(cs, m.target)
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
	target := m.target.root
	if len(root) > 0 && !root[0] {
		target = m.target
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
	return m.Set("detail", arg...).target.Spawn(m, name, help).Begin(m).Start(m)
}

// }}}

func (m *Message) Sess(key string, arg ...string) *Message { // {{{
	if len(arg) > 0 {
		root := true
		if len(arg) > 2 {
			root = Right(arg[2])
		}
		method := "find"
		if len(arg) > 1 {
			method = arg[1]
		}
		switch method {
		case "find":
			m.target.Sessions[key] = m.Find(arg[0], root)
		case "search":
			m.target.Sessions[key] = m.Search(arg[0], root)[0]
		}
		return m.target.Sessions[key]
	}

	for msg := m; msg != nil; msg = msg.message {
		if x, ok := msg.target.Sessions[key]; ok {
			return m.Spawn(x.target)
		}
	}

	return nil
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
		}
		m.Meta[key] = append(m.Meta[key], value...)

		for _, v := range m.Meta[meta] {
			if v == key {
				return m
			}
		}
		m.Meta[meta] = append(m.Meta[meta], key)

	default:
		m.Log("error", nil, "%s 消息参数错误", meta)
	}

	return m
}

// }}}
func (m *Message) Set(meta string, arg ...string) *Message { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}

	switch meta {
	case "detail", "result":
		delete(m.Meta, meta)
	case "option", "append":
		if len(arg) > 0 {
			delete(m.Meta, arg[0])
		} else {
			for _, k := range m.Meta[meta] {
				delete(m.Meta, k)
				delete(m.Data, k)
			}
			delete(m.Meta, meta)
		}
	default:
		m.Log("error", nil, "%s 消息参数错误", meta)
	}

	if len(arg) > 0 {
		m.Add(meta, arg[0], arg[1:]...)
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
		m.Data[key] = value

		if _, ok := m.Meta[meta]; !ok {
			m.Meta[meta] = make([]string, 0, 3)
		}
		for _, v := range m.Meta[meta] {
			if v == key {
				return m
			}
		}
		m.Meta[meta] = append(m.Meta[meta], key)

	default:
		m.Log("error", nil, "%s 消息参数错误", meta)
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
func (m *Message) Geti(key string) int { // {{{
	n, e := strconv.Atoi(m.Get(key))
	m.Assert(e)
	return n
}

// }}}
func (m *Message) Gets(key string) bool { // {{{
	return Right(m.Get(key))
}

// }}}
func (m *Message) Echo(str string, arg ...interface{}) *Message { // {{{
	return m.Add("result", fmt.Sprintf(str, arg...))
}

// }}}
func (m *Message) Copy(msg *Message, meta string, arg ...string) *Message { // {{{
	switch meta {
	case "detail", "result":
		m.Meta[meta] = append(m.Meta[meta][:0], msg.Meta[meta]...)
	case "option", "append":
		if len(arg) == 0 {
			arg = msg.Meta[meta]
		}

		for _, k := range arg {
			if v, ok := msg.Meta[k]; ok {
				m.Set(meta, k).Add(meta, k, v...)
			}
			if v, ok := msg.Data[k]; ok {
				m.Put(meta, k, v)
			}
		}
	}

	return m
}

// }}}

func (m *Message) Insert(meta string, index int, arg ...interface{}) string { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}

	str := []string{}
	for _, v := range arg {
		switch s := v.(type) {
		case string:
			str = append(str, s)
		case []string:
			str = append(str, s...)
		case []int:
			for _, v := range s {
				str = append(str, fmt.Sprintf("%d", v))
			}
		case []bool:
			for _, v := range s {
				str = append(str, fmt.Sprintf("%t", v))
			}
		default:
			str = append(str, fmt.Sprintf("%v", s))
		}
	}

	if index == -1 {
		index, m.Meta[meta] = 0, append(str, m.Meta[meta]...)
	} else if index == -2 {
		index, m.Meta[meta] = len(m.Meta[meta]), append(m.Meta[meta], str...)
	} else {
		if index < -2 {
			index += len(m.Meta[meta]) + 2
		}
		if index < 0 {
			index = 0
		}

		for i := len(m.Meta[meta]); i < index+len(str); i++ {
			m.Meta[meta] = append(m.Meta[meta], "")
		}
		for i := 0; i < len(str); i++ {
			m.Meta[meta][index+i] = str[i]
		}
	}

	if -1 < index && index < len(m.Meta[meta]) {
		return m.Meta[meta][index]
	}
	return ""
}

// }}}
func (m *Message) Detail(index int, arg ...interface{}) string { // {{{
	return m.Insert("detail", index, arg...)
}

// }}}
func (m *Message) Detaili(index int, arg ...int) int { // {{{
	i, e := strconv.Atoi(m.Insert("detail", index, arg))
	m.Assert(e)
	return i
}

// }}}
func (m *Message) Details(index int, arg ...bool) bool { // {{{
	return Right(m.Insert("detail", index, arg))
}

// }}}
func (m *Message) Result(index int, arg ...interface{}) string { // {{{
	return m.Insert("result", index, arg...)
}

// }}}
func (m *Message) Resulti(index int, arg ...int) int { // {{{
	i, e := strconv.Atoi(m.Insert("result", index, arg))
	m.Assert(e)
	return i
}

// }}}
func (m *Message) Results(index int, arg ...bool) bool { // {{{
	return Right(m.Insert("result", index, arg))
}

// }}}

func (m *Message) Option(key string, arg ...interface{}) string { // {{{
	m.Insert(key, 0, arg...)
	if _, ok := m.Meta[key]; ok {
		m.Add("option", key)
	}

	for msg := m; msg != nil; msg = msg.message {
		if msg.Has(key) {
			return msg.Get(key)
		}
	}
	return ""
}

// }}}
func (m *Message) Optioni(key string, arg ...int) int { // {{{
	i, e := strconv.Atoi(m.Option(key, arg))
	m.Assert(e)
	return i
}

// }}}
func (m *Message) Options(key string, arg ...bool) bool { // {{{
	return Right(m.Option(key, arg))
}

// }}}
func (m *Message) Append(key string, arg ...interface{}) string { // {{{
	m.Insert(key, 0, arg...)
	if _, ok := m.Meta[key]; ok {
		m.Add("append", key)
	}

	for msg := m; msg != nil; msg = msg.message {
		if m.Has(key) {
			return m.Get(key)
		}
	}
	return ""
}

// }}}
func (m *Message) Appendi(key string, arg ...int) int { // {{{
	i, e := strconv.Atoi(m.Append(key, arg))
	m.Assert(e)
	return i
}

// }}}
func (m *Message) Appends(key string, arg ...bool) bool { // {{{
	return Right(m.Append(key, arg))
}

// }}}

func (m *Message) Exec(key string, arg ...string) string { // {{{

	for _, c := range []*Context{m.target, m.target.master, m.target.Owner, m.source, m.source.master, m.source.Owner} {
		for s := c; s != nil; s = s.context {

			m.master = m.source
			if x, ok := s.Commands[key]; ok && x.Hand != nil && m.Check(c, "commands", key) {
				m.AssertOne(m, true, func(m *Message) {
					m.Log("cmd", s, "%d %s %v %v", len(m.target.Historys), key, arg, m.Meta["option"])

					if x.Options != nil {
						for _, v := range m.Meta["option"] {
							if _, ok := x.Options[v]; !ok {
								panic(errors.New(fmt.Sprintf("未知参数: %s", v)))
							}
						}
					}

					if m.Has("args") {
						m.Meta["args"] = nil
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

					m.Hand = true
					x.Hand(m.Set("result").Set("append"), s, key, arg...)

					if x.Appends != nil {
						for _, v := range m.Meta["append"] {
							if _, ok := x.Appends[v]; !ok {
								panic(errors.New(fmt.Sprintf("未知参数: %s", v)))
							}
						}
					}

					if m.target.Historys == nil {
						m.target.Historys = make([]*Message, 0, 10)
					}
					m.target.Historys = append(m.target.Historys, m)
				})

				return m.Get("result")
			}
		}
	}
	return ""
}

// }}}
func (m *Message) Deal(pre func(msg *Message, arg ...string) bool, post func(msg *Message, arg ...string) bool) { // {{{
	if m.target.messages == nil {
		m.target.messages = make(chan *Message, m.Confi("MessageQueueSize"))
	}

	for run := true; run; {
		m.AssertOne(<-m.target.messages, true, func(msg *Message) {
			defer func() {
				if msg.Wait != nil {
					msg.Wait <- true
				}
			}()

			if len(msg.Meta["detail"]) == 0 {
				return
			}

			if pre == nil || pre(msg, msg.Meta["detail"]...) {
				msg.Exec(msg.Meta["detail"][0], msg.Meta["detail"][1:]...)
			}

			if post != nil && !post(msg, msg.Meta["result"]...) {
				run = false
				return
			}
		})
	}
}

// }}}
func (m *Message) Post(s *Context, async ...bool) string { // {{{
	if s == nil {
		s = m.target.master
	}

	if s != nil && s.messages != nil {
		if s.messages <- m; m.Wait != nil {
			<-m.Wait
		}
		return m.Get("result")
	}

	return m.Exec(m.Meta["detail"][0], m.Meta["detail"][1:]...)
}

// }}}
func (m *Message) Cmd(arg ...interface{}) *Message { // {{{
	if m.Hand {
		if m.message != nil {
			m = m.message.Spawn(m.target)
		} else {
			msg := m.Spawn(m.target)
			msg.source = m.source
			m = msg
		}
	}

	if len(arg) > 0 {
		m.Set("detail")
		m.Detail(0, arg...)
	}

	if s := m.target.master; s != nil && s != m.source.master {
		m.Post(s)
	} else {
		m.Exec(m.Meta["detail"][0], m.Meta["detail"][1:]...)
	}

	return m
}

// }}}

func (m *Message) Confs(key string, arg ...bool) bool { // {{{
	if len(arg) > 0 {
		if arg[0] {
			m.Conf(key, "1")
		} else {
			m.Conf(key, "0")
		}
	}

	b := m.Conf(key)
	return b != "" && b != "0" && b != "false"
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

	for _, c := range []*Context{m.target, m.target.master, m.target.Owner, m.source, m.source.master, m.source.Owner} {
		for s := c; s != nil; s = s.context {
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
					if x.Hand != nil {
						x.Value = x.Hand(m, x, arg[0])
					} else {
						x.Value = arg[0]
					}
					// m.Log("conf", s, "%s %v", x.Name, x.Value)
					return x.Value
				case 0:
					if x.Hand != nil {
						return x.Hand(m, x)
					}
					return x.Value
				}
			}
		}
	}

	if len(arg) == 3 && m.Check(m.target, "configs", key) {
		if m.target.Configs == nil {
			m.target.Configs = make(map[string]*Config)
		}

		m.target.Configs[key] = &Config{Name: arg[0], Value: arg[1], Help: arg[2], Hand: hand}
		m.Log("conf", nil, "%s %v", key, arg)
		return m.Conf(key, arg[1])
	}

	m.Log("error", nil, "%s 配置项不存在", key)
	return ""
}

// }}}
func (m *Message) Caps(key string, arg ...bool) bool { // {{{
	if len(arg) > 0 {
		if arg[0] {
			m.Cap(key, "1")
		} else {
			m.Cap(key, "0")
		}
	}

	b := m.Cap(key)
	return b != "" && b != "0" && b != "false"
}

// }}}
func (m *Message) Capi(key string, arg ...int) int { // {{{
	n, e := strconv.Atoi(m.Cap(key))
	m.Assert(e)

	for _, i := range arg {
		if i == 0 {
			i = -n
		}
		n, e = strconv.Atoi(m.Cap(key, fmt.Sprintf("%d", n+i)))
		m.Assert(e)
	}

	return n
}

// }}}
func (m *Message) Cap(key string, arg ...string) string { // {{{
	var hand func(m *Message, x *Cache, arg ...string) string

	for _, c := range []*Context{m.target, m.target.master, m.target.Owner, m.source, m.source.master, m.source.Owner} {
		for s := c; s != nil; s = s.context {
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
					// m.Log("debug", s, "%s %s", x.Name, x.Value)
					return x.Value
				case 0:
					// m.Log("debug", s, "%s %s", x.Name, x.Value)
					if x.Hand != nil {
						return x.Hand(m, x)
					}
					return x.Value
				}
			}
		}
	}

	if len(arg) == 3 && m.Check(m.target, "caches", key) {
		if m.target.Caches == nil {
			m.target.Caches = make(map[string]*Cache)
		}

		m.target.Caches[key] = &Cache{Name: arg[0], Value: arg[1], Help: arg[2], Hand: hand}
		m.Log("cap", nil, "%s %v", key, arg)
		return m.Cap(key, arg[1])
	}

	m.Log("error", nil, "%s 缓存项不存在", key)
	return ""
}

// }}}

var Pulse = &Message{code: 0, time: time.Now(), Wait: make(chan bool), source: Index, master: Index, target: Index}
var Index = &Context{Name: "ctx", Help: "模块中心",
	Caches: map[string]*Cache{
		"debug":    &Cache{Name: "服务数量", Value: "true", Help: "显示已经启动运行模块的数量"},
		"nserver":  &Cache{Name: "服务数量", Value: "0", Help: "显示已经启动运行模块的数量"},
		"ncontext": &Cache{Name: "模块数量", Value: "0", Help: "显示功能树已经注册模块的数量"},
		"nmessage": &Cache{Name: "消息数量", Value: "0", Help: "显示模块启动时所创建消息的数量"},
	},
	Configs: map[string]*Config{
		"debug": &Config{Name: "调试模式(true/false)", Value: "true", Help: "是否打印错误信息，off:不打印，on:打印)"},

		"default": &Config{Name: "默认的搜索起点(root/back/home)", Value: "root", Help: "模块搜索的默认起点，root:从根模块，back:从父模块，home:从当前模块"},

		"start":    &Config{Name: "启动模块", Value: "cli", Help: "启动时自动运行的模块"},
		"init.shy": &Config{Name: "启动脚本", Value: "etc/init.shy", Help: "模块启动时自动运行的脚本"},
		"bench.log": &Config{Name: "日志文件", Value: "var/bench.log", Help: "模块日志输出的文件", Hand: func(m *Message, x *Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				// if e := os.MkdirAll(path.Dir(arg[0]), os.ModePerm); e == nil {
				if l, e := os.Create(x.Value); e == nil {
					log.SetOutput(l)
					return arg[0]
				}
				return ""
				// }
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
				}
				if e := os.Chdir(x.Value); e != nil {
					fmt.Println(e)
				}
				return arg[0]
			}

			return x.Value
			// }}}
		}},

		"ContextRequestSize": &Config{Name: "请求队列长度", Value: "10", Help: "每个模块可以被其它模块引用的的数量"},
		"ContextSessionSize": &Config{Name: "会话队列长度", Value: "10", Help: "每个模块可以启动其它模块的数量"},
		"MessageQueueSize":   &Config{Name: "消息队列长度", Value: "10", Help: "每个模块接收消息的队列长度"},

		"cert": &Config{Name: "证书文件", Value: "etc/cert.pem", Help: "证书文件"},
		"key":  &Config{Name: "私钥文件", Value: "etc/key.pem", Help: "私钥文件"},
	},
	Commands: map[string]*Command{
		"userinfo": &Command{Name: "userinfo [add|del [context key name help]|[command|config|cache group name]]", Help: "查看模块的用户信息",
			Formats: map[string]int{"add": -1, "del": -1},
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				switch { // {{{
				case m.Has("add"):
					m.target.Add(m.source.Group, m.Meta["add"]...)
				case m.Has("del"):
					m.target.Del(m.Meta["del"]...)
				default:
					target := m.target
					m.target = target.Owner
					if m.target != nil && m.Check(m.target) {
						m.Echo("%s %s\n", m.Cap("username"), m.Cap("group"))
					}
					m.target = target

					if len(m.Meta["args"]) > 0 {
						if g, ok := m.target.Index[m.Get("args")]; ok {
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
						for k, v := range m.target.Index {
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
				m.Travel(m.target.root, func(m *Message) bool {
					if x, ok := m.target.Caches["status"]; ok {
						m.Echo("%s(%s): %s\n", m.target.Name, x.Value, m.target.Help)
					}
					return true
				})

			default:
				switch arg[0] {
				case "start":
					m.Meta = nil
					m.Set("detail", arg[1:]...).target.Start(m)
				case "stop":
					m.Set("detail", arg[1:]...).target.Close(m)
				case "switch":
				}
			}
			// }}}
		}},
		"message": &Command{Name: "message code", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			switch len(arg) { // {{{
			case 0:
				m.Echo("\033[31mrequests:\033[0m\n")
				for i, v := range m.target.Requests {
					m.Echo("%d %s\n", i, v.Format())
					for i, v := range v.messages {
						m.Echo("  %d %s\n", i, v.Format())
					}
				}

				m.Echo("\033[32msessions:\033[0m\n")
				for k, v := range m.target.Sessions {
					m.Echo("%s %s\n", k, v.Format())
				}

				m.Echo("\033[33mhistorys:\033[0m\n")
				for i, v := range m.target.Historys {
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
		"detail": &Command{Name: "detail index val...", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.Spawn(m.Target())

			msg.Detail(1, "nie", 1, []string{"123", "123"}, true, []bool{false, true}, []int{1, 2, 2})

			m.Echo("%v", msg.Meta)
			msg.Detail(2, "nie")
			m.Echo("%v", msg.Meta)

		}},
		"option": &Command{Name: "option key val...", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			if len(arg) > 0 { // {{{
				m.Option(arg[0], arg[1:])
			} else {
				for msg := m; msg != nil; msg = msg.message {
					m.Echo("%d %s:%s->%s %v\n", msg.code, msg.time.Format("15:03:04"), msg.source.Name, msg.target.Name, msg.Meta["detail"])
					for _, k := range msg.Meta["option"] {
						m.Echo("%s: %v\n", k, msg.Meta[k])
					}
				}
			} // }}}
		}},
		"append": &Command{Name: "append key val...", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			if len(arg) > 0 { // {{{
				m.Append(arg[0], arg[1:])
			} else {
				for msg := m; msg != nil; msg = msg.message {
					m.Echo("%d %s:%s->%s %v\n", msg.code, msg.time.Format("15:03:04"), msg.source.Name, msg.target.Name, msg.Meta["result"])
					for _, k := range msg.Meta["append"] {
						m.Echo("%s: %v\n", k, msg.Meta[k])
					}
				}
			} // }}}
		}},
		"context": &Command{Name: "context back|[[home] [find|search] name] [info|list|show|spawn|start|switch|close][args]", Help: "查找并操作模块，\n查找起点root:根模块、back:父模块、home:本模块，\n查找方法find:路径匹配、search:模糊匹配，\n查找对象name:支持点分和正则，\n操作类型show:显示信息、switch:切换为当前、start:启动模块、spawn:分裂子模块，args:启动参数",
			Formats: map[string]int{"back": 0, "home": 0, "find": 1, "search": 1, "info": 1, "list": 0, "show": 0, "close": 0, "switch": 0, "start": 0, "spawn": 0},
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				if m.Has("back") { // {{{
					m.target = m.source
					return
				}
				root := !m.Has("home")

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
					// v.Meta = m.Meta
					// v.Data = m.Data
					switch {
					case m.Has("switch"), m.Has("back"):
						m.target = v.target
					case m.Has("spawn"):
						v.Set("detail", arg[2:]...).target.Spawn(v, arg[0], arg[1]).Begin(v)
						m.target = v.target
					case m.Has("start"):
						v.Set("detail", arg...).target.Start(v)
						m.target = v.target
					case m.Has("close"):
						v.target.Close(v)
					case m.Has("show"):
						m.Echo("%s(%s): %s\n", v.target.Name, v.target.Owner.Name, v.target.Help)
						if len(v.target.Requests) > 0 {
							m.Echo("模块资源：\n")
							for i, v := range v.target.Requests {
								m.Echo("  %d: <- %s %s\n", i, v.source.Name, v.Meta["detail"])
								// for i, v := range v.Messages {
								// 	m.Echo("    %d: -> %s %s\n", i, v.source.Name, v.Meta["detail"])
								// }
							}
						}
						if len(v.target.Sessions) > 0 {
							m.Echo("模块引用：\n")
							for k, v := range v.target.Sessions {
								m.Echo("  %s: -> %s %v\n", k, v.target.Name, v.Meta["detail"])
							}
						}
					case m.Has("info"):
						switch m.Get("info") {
						case "name":
							m.Echo("%s", v.target.Name)
						case "owner":
							m.Echo("%s", v.target.Owner.Name)
						default:
							m.Echo("%s(%s): %s\n", v.target.Name, v.target.Owner.Name, v.target.Help)
						}
					case m.Has("list") || len(m.Meta["detail"]) == 1:
						m.Travel(v.target, func(msg *Message) bool {
							target := msg.target
							m.Echo("%s(", target.Name)

							if target.context != nil {
								m.Echo("%s", target.context.Name)
							}
							m.Echo(":")

							if target.master != nil {
								m.Echo("%s", target.master.Name)
							}
							m.Echo(":")

							if target.Owner != nil {
								m.Echo("%s", target.Owner.Name)
							}
							m.Echo(":")

							msg.target = msg.target.Owner
							if msg.target != nil && msg.Check(msg.target, "caches", "username") && msg.Check(msg.target, "caches", "group") {
								m.Echo("%s:%s", msg.Cap("username"), msg.Cap("group"))
							}
							m.Echo("): ")
							msg.target = target

							if msg.Check(msg.target, "caches", "status") && msg.Check(msg.target, "caches", "stream") {
								m.Echo("%s(%s) ", msg.Cap("status"), msg.Cap("stream"))
							}
							m.Echo("%s\n", target.Help)
							return true
						})
					case len(arg) > 0 && v != m:
						v.Meta = m.Meta
						v.Cmd(arg)
						m.Meta = v.Meta
					default:
						m.target = v.target
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
							m.Echo("%s comands:\n", m.target.Name)
						}
						for k, x := range m.target.Commands {
							if m.Check(m.target, "commands", k) {
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
						if _, ok := m.target.Commands[arg[0]]; ok {
							if m.target.Owner == nil || m.master.Owner == m.target.Owner {
								delete(m.target.Commands, arg[0])
							}
						}
					case m.Has("void"):
						if x, ok := m.target.Commands[arg[0]]; ok {
							if m.target.Owner == nil || m.master.Owner == m.target.Owner {
								x.Hand = nil
							}
						}
					}

					m.BackTrace(func(m *Message) bool {
						if all {
							m.Echo("%s commands:\n", m.target.Name)
						}
						if x, ok := m.target.Commands[arg[0]]; ok {
							if all {
								m.Echo("  ")
							}
							if m.Check(m.target, "commands", arg[0]) {
								m.Echo("%s\n    %s\n", x.Name, x.Help)
							}
						}
						return all
					})
					m.Assert(m.Has("result"), "%s 命令不存在", arg[0])
				case 3:
					cmd := &Command{}
					m.BackTrace(func(m *Message) bool {
						if x, ok := m.target.Commands[arg[0]]; ok && x.Hand != nil {
							*cmd = *x
						}
						return all
					})

					if m.Check(m.target, "commands", arg[0]) {
						if x, ok := m.target.Commands[arg[0]]; ok {
							if m.target.Owner == nil || m.master.Owner == m.target.Owner {
								x.Name = arg[1]
								x.Help = arg[2]
								m.Echo("%s\n    %s\n", x.Name, x.Help)
							}
						} else {
							if m.target.Commands == nil {
								m.target.Commands = map[string]*Command{}
							}
							cmd.Name = arg[1]
							cmd.Help = arg[2]
							m.target.Commands[arg[0]] = cmd
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
							m.Echo("%s configs:\n", m.target.Name)
						}
						for k, x := range m.target.Configs {
							if m.Check(m.target, "configs", k) {
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
						if _, ok := m.target.Configs[arg[0]]; ok {
							if m.target.Owner == nil || m.master.Owner == m.target.Owner {
								delete(m.target.Configs, arg[0])
							}
						}
					case m.Has("void"):
						m.Conf(arg[0], "")
					}

					m.BackTrace(func(m *Message) bool {
						// if all {
						// 	m.Echo("%s config:\n", m.target.Name)
						// }
						if x, ok := m.target.Configs[arg[0]]; ok {
							if m.Check(m.target, "configs", arg[0]) {
								// if all {
								// 	m.Echo("  ")
								// }
								// m.Echo("%s: %s\n", x.Name, x.Help)
								m.Echo("%s", x.Value)
								return false
							}
						}
						return true
						// return all
					})

				case 2:
					m.Conf(arg[0], arg[1])
				case 3:
					m.Conf(arg[0], arg[2])
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
							m.Echo("%s configs:\n", m.target.Name)
						}
						for k, x := range m.target.Caches {
							if m.Check(m.target, "caches", k) {
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
						if _, ok := m.target.Caches[arg[0]]; ok {
							if m.target.Owner == nil || m.master.Owner == m.target.Owner {
								delete(m.target.Caches, arg[0])
							}
						}
					case m.Has("void"):
						m.Cap(arg[0], "")
					}

					if m.source == m.source.master {
						m.source, m.target = m.target, m.source
					}
					m.Echo("%s", m.Cap(arg[0]))
				case 2:
					if m.source == m.source.master {
						m.source, m.target = m.target, m.source
					}
					m.Cap(arg[0], arg[1])
				case 3:
					if m.source == m.source.master {
						m.source, m.target = m.target, m.source
					}
					m.Cap(arg[0], arg[2])
				case 4:
					if m.source == m.source.master {
						m.source, m.target = m.target, m.source
					}
					m.Cap(arg[0], arg[1:]...)
				}
				// }}}
			}},
		"pulse": &Command{Name: "arg name", Help: "查看日志", Hand: func(m *Message, c *Context, key string, arg ...string) {
			p := m.Target().Pulse
			m.Echo("%d\n", p.code)
			m.Echo("%v\n", p.Meta["detail"])
			m.Echo("%v\n", p.Meta["option"])
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
		Pulse.Conf("init.shy", args[1])
	}
	if len(args) > 2 {
		Pulse.Conf("bench.log", args[2])
	} else {
		Pulse.Conf("bench.log", Pulse.Conf("bench.log"))
	}
	if len(args) > 3 {
		Pulse.Conf("root", args[3])
	}

	Pulse.Options("log", true)

	// log.Println("\n\n")
	Index.Group = "root"
	Index.Owner = Index.contexts["aaa"]
	Index.master = Index.contexts["cli"]
	for _, m := range Pulse.Search("") {
		m.target.root = Index
		m.target.Begin(m)
	}
	// log.Println()
	Pulse.Sess("log", "log").Conf("bench.log", "var/bench.log")

	for _, m := range Pulse.Search(Pulse.Conf("start")) {
		m.Set("detail", Pulse.Conf("init.shy")).Set("option", "stdio").target.Start(m)
	}

	<-Index.master.Exit
}
