package ctx // {{{
// }}}
import ( // {{{
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"os"
	"path"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

// }}}

func Right(str string) bool {
	switch str {
	case "", "0", "false", "off", "no", "error: ":
		return false
	}
	return true
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

	Shares  map[string][]string
	Options map[string]string
	Appends map[string]string

	Form map[string]int
	Hand func(m *Message, c *Context, key string, arg ...string)
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

	Owner    *Context
	Index    map[string]*Context
	password string

	Server
}

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

	item := []string{}
	m.BackTrace(func(m *Message) bool {
		item = append(item, m.target.Name)
		return true
	})
	for i := 0; i < len(item)/2; i++ {
		item[i], item[len(item)-i-1] = item[len(item)-i-1], item[i]
	}
	c.Caches["module"] = &Cache{Name: "服务数据", Value: strings.Join(item, "."), Help: "服务数据"}

	m.Index = 1
	c.Pulse = m
	c.Requests = []*Message{m}
	c.Historys = []*Message{m}
	c.Sessions = map[string]*Message{}

	c.master = m.master.master
	c.Owner = m.master.Owner

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
		go m.TryCatch(m, true, func(m *Message) {
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
			m.Log("info", nil, "before exit<-")
			c.Exit <- true
			m.Log("info", nil, "after exit<-")
		}
	}
	return true
}

// }}}

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
func (c *Context) Check(m *Message, arg ...string) bool { // {{{
	if g, ok := c.Index["void"]; ok && g != nil {
		if len(arg) < 2 {
			return true
		}

		switch arg[0] {
		case "caches":
			_, ok = g.Caches[arg[1]]
		case "configs":
			_, ok = g.Configs[arg[1]]
		case "commands":
			_, ok = g.Commands[arg[1]]
		}

		if ok {
			return true
		}
	}

	aaa := m
	for msg := m; msg != nil && msg.code != 0 && msg != msg.message; msg = msg.message {
		aaa = nil
		if a, ok := msg.Sessions["aaa"]; ok {
			aaa = a
			break
		}
	}

	if aaa == nil {
		return true
	}

	if c.Owner == nil {
		return true
	}

	if c.Owner == aaa.target {
		return true
	}

	group := aaa.Cap("group")
	if group == aaa.Conf("rootname") {
		return true
	}

	if g, ok := c.Index[group]; ok && g != nil {
		if len(arg) < 2 {
			return true
		}

		switch arg[0] {
		case "caches":
			_, ok = g.Caches[arg[1]]
		case "configs":
			_, ok = g.Configs[arg[1]]
		case "commands":
			_, ok = g.Commands[arg[1]]
		}

		if ok {
			return true
		}
	}

	m.Log("check", nil, "%s %d %v failure", c.Name, m.code, arg)
	return false
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

type Callback struct {
	ncall int
	hand  func(msg *Message) (sub *Message)
}

type Message struct {
	code int
	time time.Time

	Name   string
	source *Context
	master *Context
	target *Context
	Index  int

	Meta map[string][]string
	Data map[string]interface{}

	Sessions map[string]*Message
	messages []*Message
	message  *Message
	root     *Message

	callback Callback

	Wait chan bool
	Recv chan bool
	Hand bool

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

func (m *Message) Login(aaa *Message) { // {{{
	m.source.Owner = m.target
	for msg := m; msg != nil; msg = msg.message {
		if nfs, ok := msg.Sessions["nfs"]; ok {
			nfs.Sessions["aaa"] = aaa
			return
		}
	}
}

// }}}
func (m *Message) Log(action string, ctx *Context, str string, arg ...interface{}) *Message { // {{{
	if !m.Options("log") {
		return m
	}

	if l := m.Sesss("log"); l != nil {
		l.Options("log", false)
		l.Cmd("log", action, fmt.Sprintf(str, arg...))
	}
	return m
}

// }}}
func (m *Message) Check(s *Context, arg ...string) bool { // {{{
	return true
	if m.root.target.Sessions == nil || m.root.target.Sessions["aaa"] == nil {
		return true
	}

	if g, ok := s.Index["void"]; ok && g != nil {
		if len(arg) < 2 {
			return true
		}

		switch arg[0] {
		case "caches":
			_, ok = g.Caches[arg[1]]
		case "configs":
			_, ok = g.Configs[arg[1]]
		case "commands":
			_, ok = g.Commands[arg[1]]
		}

		if ok {
			return true
		}
	}

	aaa := m
	for msg := m; msg != nil && msg.code != 0 && msg != msg.message; msg = msg.message {
		if a, ok := msg.Target().Sessions["aaa"]; ok {
			aaa = a
			break
		}
	}

	if aaa == nil {
		return true
	}

	group := aaa.Cap("group")
	if group == aaa.Conf("rootname") {
		return true
	}

	if g, ok := s.Index[group]; ok && g != nil {
		if len(arg) < 2 {
			return true
		}

		switch arg[0] {
		case "caches":
			_, ok = g.Caches[arg[1]]
		case "configs":
			_, ok = g.Configs[arg[1]]
		case "commands":
			_, ok = g.Commands[arg[1]]
		}

		if ok {
			return true
		}
	}

	fmt.Printf("check %s %s %v false\n", group, s.Name, arg)
	return false

	if aaa.target.Caches == nil {
		return true
	}
	if aaa.target.Caches["group"] == nil {
		return true
	}

	g, ok := s.Index[group]
	gg, gok := s.Index["void"]

	if len(arg) < 2 {
		if ok && g != nil {
			return true
		}

		if gok && gg != nil {
			return true
		}

		return false
	}
	fmt.Printf("%v\n", arg)

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
	if gok {
		return true
	}
	return false
}

// }}}
func (m *Message) Permit(s *Context, arg ...string) bool { // {{{

	if m.root.target.Sessions == nil || m.root.target.Sessions["aaa"] == nil {
		return true
	}

	if aaa := m.Sesss("aaa"); aaa != nil {

		if g, ok := s.Index["void"]; ok && g != nil {
			if len(arg) < 2 {
				return true
			}

			switch arg[0] {
			case "caches":
				_, ok = g.Caches[arg[1]]
			case "configs":
				_, ok = g.Configs[arg[1]]
			case "commands":
				_, ok = g.Commands[arg[1]]
			}

			if ok {
				return true
			}
		}

		group := aaa.Cap("group")
		if group == aaa.Conf("rootname") {
			return true
		}

		if g, ok := s.Index[group]; ok && g != nil {
			if len(arg) < 2 {
				return true
			}

			switch arg[0] {
			case "caches":
				_, ok = g.Caches[arg[1]]
			case "configs":
				_, ok = g.Configs[arg[1]]
			case "commands":
				_, ok = g.Commands[arg[1]]
			}

			if ok {
				return true
			}
		}
		return true
	}

	return true
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
		if result, ok := e.Meta["result"]; ok && len(result) > 0 && result[0] == "error: " {
			panic(e)
		}
		return true
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
	panic(e)
}

// }}}
func (m *Message) TryCatch(msg *Message, safe bool, hand ...func(msg *Message)) *Message { // {{{
	defer func() {
		if e := recover(); e != nil {
			switch v := e.(type) {
			case *Message:
				e = strings.Join(v.Meta["result"][1:], "")
			}

			msg.Log("error", nil, "error: %v", e)
			if msg.root.Conf("debug") == "on" && e != io.EOF {
				fmt.Printf("\n\033[31m%s error: %v\033[0m\n", msg.target.Name, e)
				debug.PrintStack()
				fmt.Printf("\033[31m%s error: %v\033[0m\n\n", msg.target.Name, e)
			}

			if len(hand) > 1 {
				m.TryCatch(msg, safe, hand[1:]...)
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

func (m *Message) Spawn(arg ...interface{}) *Message { // {{{
	c := m.target
	if len(arg) > 0 {
		if v, ok := arg[0].(*Context); ok {
			c, arg = v, arg[1:]
		}
	}
	key := ""
	if len(arg) > 0 {
		if v, ok := arg[0].(string); ok {
			key, arg = v, arg[1:]
		}
	}

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
	msg.Sessions = make(map[string]*Message)

	msg.Wait = make(chan bool)
	if key == "" {
		return msg
	}

	if msg.source.Sessions == nil {
		msg.source.Sessions = make(map[string]*Message)
	}
	msg.source.Sessions[key] = msg
	msg.Name = key
	return msg
}

// }}}
func (m *Message) Reply(key ...string) *Message { // {{{
	if m.Template == nil {
		m.Template = m.Spawn(m.source)
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

		keys := []string{}
		for k, _ := range cs[i].contexts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			cs = append(cs, cs[i].contexts[k])
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
		} else if target.Name == v {
			continue
		} else {
			m.Log("find", target, "not find %s", v)
			return nil
		}
	}
	m.Log("find", nil, "find %s", name)
	return m.Spawn(target)
}

// }}}
func (m *Message) Sess(key string, arg ...string) *Message { // {{{

	if _, ok := m.target.Sessions[key]; !ok && len(arg) > 0 {
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
func (m *Message) Sesss(key string, arg ...interface{}) *Message { // {{{
	spawn := true
	if _, ok := m.Sessions[key]; !ok && len(arg) > 0 {
		switch value := arg[0].(type) {
		case *Message:
			m.Sessions[key] = value
			return m.Sessions[key]
		case *Context:
			m.Sessions[key] = m.Spawn(value)
			return m.Sessions[key]
		case string:
			root := true
			if len(arg) > 2 {
				root = Right(arg[2].(string))
			}
			method := "find"
			if len(arg) > 1 {
				method = arg[1].(string)
			}

			switch method {
			case "find":
				m.Sessions[key] = m.Find(value, root)
			case "search":
				m.Sessions[key] = m.Search(value, root)[0]
			}
			return m.Sessions[key]
		case bool:
			spawn = value
		}
	}

	for msg := m; msg != nil; msg = msg.message {
		if x, ok := msg.Sessions[key]; ok {
			if spawn {
				x = m.Spawn(x.target)
			}
			return x
		}
	}

	return nil
}

// }}}

func (m *Message) Call(cb func(msg *Message) (sub *Message), arg ...interface{}) *Message { // {{{
	m.callback.hand = cb
	m.Wait = nil
	m.Cmd(arg...)
	return m
}

// }}}
func (m *Message) Back(msg *Message) *Message { // {{{
	if msg == nil || m.callback.hand == nil {
		return m
	}

	m.Log("cb", nil, "%d %v %v", msg.code, msg.Meta["result"], msg.Meta["append"])

	m.callback.ncall++
	if sub := m.callback.hand(msg); sub != nil && m.message != nil && m.message != m {
		m.message.Back(sub)
	}

	return m
}

// }}}
func (m *Message) CallBack(sync bool, cb func(msg *Message) (sub *Message), arg ...interface{}) *Message { // {{{
	if !sync {
		m.Call(cb, arg...)
		return m
	}

	wait := make(chan bool)

	go m.Call(func(sub *Message) *Message {
		msg := cb(sub)
		m.Log("lock", nil, "before done %v", arg)
		wait <- true
		m.Log("lock", nil, "after done %v", arg)
		return msg
	}, arg...)

	m.Log("lock", nil, "before wait %v", arg)
	<-wait
	m.Log("lock", nil, "after wait %v", arg)
	return m
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
func (m *Message) Color(color int, str string, arg ...interface{}) *Message { // {{{
	return m.Add("result", fmt.Sprintf("\033[%dm%s\033[0m", color, fmt.Sprintf(str, arg...)))
}

// }}}
func (m *Message) Copy(msg *Message, meta string, arg ...string) *Message { // {{{
	switch meta {
	case "session":
	case "detail", "result":
		m.Set(meta, msg.Meta[meta]...)
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
func (m *Message) Table(cb func(map[string]string, []string, int) bool) *Message { // {{{
	if len(m.Meta["append"]) == 0 {
		return m
	}

	width := make(map[string]int, len(m.Meta[m.Meta["append"][0]]))
	for _, k := range m.Meta["append"] {
		title := k
		if m.Options("extras") && k == "extra" {
			title = "extra." + m.Option("extras")
		}
		width[k] = len(title)
	}
	for i := 0; i < len(m.Meta[m.Meta["append"][0]]); i++ {
		for _, k := range m.Meta["append"] {
			data := m.Meta[k][i]
			if len(data) > width[k] {
				width[k] = len(data)
			}
		}
	}

	row := map[string]string{}
	wor := []string{}
	for _, k := range m.Meta["append"] {
		title := k
		if m.Options("extras") && k == "extra" {
			title = "extra." + m.Option("extras")
		}
		row[k] = title
		title += strings.Repeat(" ", width[k]-len(title))
		wor = append(wor, title)
	}
	if !cb(row, wor, -1) {
		return m
	}

	for i := 0; i < len(m.Meta[m.Meta["append"][0]]); i++ {
		row := map[string]string{}
		wor := []string{}
		for _, k := range m.Meta["append"] {
			data := m.Meta[k][i]
			if m.Options("extras") && k == "extra" {
				extra := map[string]interface{}{}
				json.Unmarshal([]byte(data), &extra)
				data = fmt.Sprintf("%v", extra[m.Option("extras")])
			}

			if i < len(m.Meta[k]) {
				row[k] = data
				data += strings.Repeat(" ", width[k]-len(data))
				wor = append(wor, data)
			}
		}
		if !cb(row, wor, i) {
			break
		}
	}

	return m
}

// }}}
func (m *Message) Matrix(index int, arg ...interface{}) string { // {{{
	if len(m.Meta["append"]) == 0 || index < 0 {
		return ""
	}

	key := m.Meta["append"][0]
	if len(arg) > 0 {
		switch v := arg[0].(type) {
		case string:
			for _, k := range m.Meta["append"] {
				if k == v {
					key = v
				}
			}
			if key != v {
				return ""
			}
		case int:
			if v < len(m.Meta["append"]) {
				key = m.Meta["append"][v]
			} else {
				return ""
			}
		}
	}
	if index < len(m.Meta[key]) {
		return m.Meta[key][index]

	}
	return ""
}

// }}}
func (m *Message) Sort(key string, arg ...string) { // {{{
	table := []map[string]string{}
	m.Table(func(line map[string]string, lists []string, index int) bool {
		if index != -1 {
			table = append(table, line)
		}
		return true
	})

	cmp := "string"
	if len(arg) > 0 {
		cmp = arg[0]
	}

	for i := 0; i < len(table)-1; i++ {
		for j := i + 1; j < len(table); j++ {
			result := false
			switch cmp {
			case "int":
				a, e := strconv.Atoi(table[i][key])
				m.Assert(e)
				b, e := strconv.Atoi(table[j][key])
				m.Assert(e)
				if a > b {
					result = true
				}
			case "int_r":
				a, e := strconv.Atoi(table[i][key])
				m.Assert(e)
				b, e := strconv.Atoi(table[j][key])
				m.Assert(e)
				if a < b {
					result = true
				}
			case "string":
				if table[i][key] > table[j][key] {
					result = true
				}
			case "string_r":
				if table[i][key] < table[j][key] {
					result = true
				}
			}

			if result {
				table[i], table[j] = table[j], table[i]
			}
		}
	}

	for _, k := range m.Meta["append"] {
		delete(m.Meta, k)
	}

	for _, v := range table {
		for _, k := range m.Meta["append"] {
			m.Add("append", k, v[k])
		}
	}
}

// }}}

func (m *Message) Insert(meta string, index int, arg ...interface{}) string { // {{{
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if len(arg) == 0 {
		if -1 < index && index < len(m.Meta[meta]) {
			return m.Meta[meta][index]
		}
		return ""
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
func (m *Message) Optionv(key string, arg ...interface{}) interface{} { // {{{
	if len(arg) > 0 {
		m.Put("option", key, arg[0])
		return arg[0]
	}

	for msg := m; msg != nil; msg = msg.message {
		if msg.Data == nil || msg.Data[key] == nil {
			continue
		}
		if msg.Meta["option"] == nil || len(msg.Meta["option"]) == 0 {
			continue
		}
		for _, v := range msg.Meta["option"] {
			if v == key {
				return msg.Data[key]
			}
		}
	}
	return nil
}

// }}}
func (m *Message) Optionx(key string, format string) interface{} { // {{{
	value := m.Option(key)
	if value != "" {
		return fmt.Sprintf(format, value)
	}
	return ""
}

// }}}
func (m *Message) Append(key string, arg ...interface{}) string { // {{{
	m.Insert(key, 0, arg...)
	if _, ok := m.Meta[key]; ok {
		m.Add("append", key)
	}

	ms := []*Message{m}
	for i := 0; i < len(ms); i++ {
		ms = append(ms, ms[i].messages...)
		if ms[i].Has(key) {
			for j := 0; j < len(ms[i].Meta["append"]); j++ {
				if ms[i].Meta["append"][j] == key {
					return ms[i].Get(key)
				}
			}
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
func (m *Message) Appendv(key string, arg ...interface{}) interface{} { // {{{
	if len(arg) > 0 {
		m.Put("append", key, arg[0])
	}

	ms := []*Message{m}
	for i := 0; i < len(ms); i++ {
		ms = append(ms, ms[i].messages...)
		if ms[i].Data == nil || ms[i].Data[key] == nil {
			continue
		}
		if ms[i].Meta["append"] == nil || len(ms[i].Meta["append"]) == 0 {
			continue
		}
		for _, v := range ms[i].Meta["append"] {
			if v == key {
				return ms[i].Data[key]
			}
		}
	}
	return nil
}

// }}}

func (m *Message) Start(name string, help string, arg ...string) bool { // {{{
	return m.Set("detail", arg...).target.Spawn(m, name, help).Begin(m).Start(m)
}

// }}}
func (m *Message) Starts(name string, help string, arg ...string) bool { // {{{
	return m.Set("detail", arg...).target.Spawn(m, name, help).Begin(m).Start(m)
}

// }}}

func (m *Message) Exec(key string, arg ...string) string { // {{{

	for _, c := range []*Context{m.target, m.target.master, m.target.Owner, m.source, m.source.master, m.source.Owner} {
		for s := c; s != nil; s = s.context {

			if x, ok := s.Commands[key]; ok && x.Hand != nil && c.Check(m, "commands", key) {
				m.TryCatch(m, true, func(m *Message) {
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
					if x.Form != nil {
						for i := 0; i < len(arg); i++ {
							n, ok := x.Form[arg[i]]
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
					// x.Hand(m.Set("result").Set("append"), s, key, arg...)
					x.Hand(m, s, key, arg...)

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
		m.TryCatch(<-m.target.messages, true, func(msg *Message) {
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

	if s, t := m.source.master, m.target.master; t != nil && s != t {
		m.Post(s)
	} else {
		m.Exec(m.Meta["detail"][0], m.Meta["detail"][1:]...)
	}

	return m
}

// }}}

func (m *Message) Confx(key string, arg ...interface{}) string { // {{{
	if len(arg) == 0 {
		value := m.Option(key)
		if value == "" {
			value = m.Conf(key)
		}
		return value
	}

	value := ""
	switch v := arg[0].(type) {
	case string:
		value = v
	case []string:
		which := 0
		if len(arg) > 1 {
			if x, ok := arg[1].(int); ok {
				which = x
				arg = arg[1:]
			}
		}
		if which < len(v) {
			value = v[which]
		}
	default:
		x := fmt.Sprintf("%v", v)
		if v != nil && x != "" {
			value = x
		}
	}

	force := false
	if len(arg) > 1 {
		if v, ok := arg[1].(bool); ok {
			arg = arg[1:]
			force = v
		}
	}
	if !force && value == "" {
		value = m.Conf(key)
	}

	format := "%s"
	if len(arg) > 1 {
		if v, ok := arg[1].(string); ok {
			arg = arg[1:]
			format = v
		}
	}
	if value != "" {
		args := []interface{}{value}
		for _, v := range arg[1:] {
			args = append(args, v)
		}
		value = fmt.Sprintf(format, args...)
	}
	return value
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
func (m *Message) Capx(key string, arg ...interface{}) string { // {{{
	if len(arg) == 0 {
		value := m.Option(key)
		if value == "" {
			value = m.Cap(key)
		}
		return value
	}

	skip := false
	if len(arg) > 1 {
		if v, ok := arg[1].(bool); ok && !v {
			skip = true
		}
	}

	if len(arg) > 0 {
		switch v := arg[0].(type) {
		case string:
			if skip || v == "" {
				return m.Cap(key)
			}
			return v
		case []string:
			which := 0
			if len(arg) > 1 {
				if x, ok := arg[1].(int); ok {
					which = x
				}
			}
			if which < len(v) {
				return v[which]
			}
			return m.Cap(key)
		default:
			x := fmt.Sprintf("%v", v)
			if skip || v == nil || x == "" {
				return m.Cap(key)
			}
			return x
		}
	}

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

var CGI = template.FuncMap{
	"meta": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}

		up := ""

		list := []string{}
		switch data := arg[0].(type) {
		case map[string][]string:
			if len(arg) == 1 {
				list = append(list, fmt.Sprintf("detail: %s\n", data["detail"]))
				list = append(list, fmt.Sprintf("option: %s\n", data["option"]))
				list = append(list, fmt.Sprintf("result: %s\n", data["result"]))
				list = append(list, fmt.Sprintf("append: %s\n", data["append"]))
				break
			}
			if key, ok := arg[1].(string); ok {
				if list, ok = data[key]; ok {
					arg = arg[1:]
				} else {
					return up
				}
			} else {
				return fmt.Sprintf("%v", data)
			}
		case []string:
			list = data
		default:
			if data == nil {
				return ""
			}
			return fmt.Sprintf("%v", data)
		}

		if len(arg) == 1 {
			return strings.Join(list, "")
		}

		index, ok := arg[1].(int)
		if !ok {
			return strings.Join(list, "")
		}

		if index >= len(list) {
			return ""
		}
		return list[index]
	}, // }}}
	"sess": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return fmt.Sprintf("%v", m)
			}

			switch which := arg[1].(type) {
			case string:
				m.Sesss(which, arg[2:]...)
				return ""
			}
		}
		return ""
	}, // }}}

	"ctx": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}
		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return fmt.Sprintf("%v", m)
			}

			switch which := arg[1].(type) {
			case string:
				switch which {
				case "name":
					return fmt.Sprintf("%s", m.target.Name)
				case "help":
					return fmt.Sprintf("%s", m.target.Help)
				case "context":
					return fmt.Sprintf("%s", m.target.context.Name)
				case "contexts":
					ctx := []string{}
					for _, v := range m.target.contexts {
						ctx = append(ctx, fmt.Sprintf("%d", v.Name))
					}
					return strings.Join(ctx, " ")
				case "time":
					return m.time.Format("2006-01-02 15:04:05")
				case "source":
					return m.source.Name
				case "target":
					return m.target.Name
				case "message":
					return fmt.Sprintf("%d", m.message.code)
				case "messages":
				case "sessions":
					msg := []string{}
					for k, _ := range m.Sessions {
						msg = append(msg, fmt.Sprintf("%s", k))
					}
					return strings.Join(msg, " ")
				}
			case int:
			}
		}
		return ""
	}, // }}}
	"msg": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}
		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return fmt.Sprintf("%v", m)
			}

			switch which := arg[1].(type) {
			case string:
				switch which {
				case "code":
					return fmt.Sprintf("%d", m.code)
				case "time":
					return m.time.Format("2006-01-02 15:04:05")
				case "source":
					return m.source.Name
				case "target":
					return m.target.Name
				case "message":
					return fmt.Sprintf("%d", m.message.code)
				case "messages":
					msg := []string{}
					for _, v := range m.messages {
						msg = append(msg, fmt.Sprintf("%d", v.code))
					}
					return strings.Join(msg, " ")
				case "sessions":
					msg := []string{}
					for k, _ := range m.Sessions {
						msg = append(msg, fmt.Sprintf("%s", k))
					}
					return strings.Join(msg, " ")
				}
			case int:
			}
		}
		return ""
	}, // }}}

	"cap": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return fmt.Sprintf("%v", m)
			}

			switch which := arg[1].(type) {
			case string:
				if len(arg) == 2 {
					return m.Cap(which)
				}

				switch value := arg[2].(type) {
				case string:
					return m.Cap(which, value)
				case int:
					return fmt.Sprintf("%d", m.Capi(which, value))
				case bool:
					return fmt.Sprintf("%t", m.Caps(which, value))
				default:
					return m.Cap(which, fmt.Sprintf("%v", arg[2]))
				}
			}
		}
		return ""
	}, // }}}
	"conf": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return fmt.Sprintf("%v", m)
			}

			switch which := arg[1].(type) {
			case string:
				if len(arg) == 2 {
					return m.Conf(which)
				}

				switch value := arg[2].(type) {
				case string:
					return m.Conf(which, value)
				case int:
					return fmt.Sprintf("%d", m.Confi(which, value))
				case bool:
					return fmt.Sprintf("%t", m.Confs(which, value))
				default:
					return m.Conf(which, fmt.Sprintf("%v", arg[2]))
				}
			}
		}
		return ""
	}, // }}}
	"cmd": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return fmt.Sprintf("%v", m)
			}

			msg := m.Spawn(m.Target()).Cmd(arg[1:]...)
			return strings.Join(msg.Meta["result"], "")
		}
		return ""
	}, // }}}

	"detail": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return strings.Join(m.Meta["detail"], "")
			}
			return m.Detail(0, arg[1:]...)
		}
		return ""
	}, // }}}
	"option": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return fmt.Sprintf("%v", m)
			}
			switch which := arg[1].(type) {
			case string:
				if len(arg) == 2 {
					return m.Option(which)
				}

				return m.Option(which, arg[2:]...)
			}
		}
		return ""
	}, // }}}
	"result": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return strings.Join(m.Meta["result"], "")
			}
			return m.Result(0, arg[1:]...)
		}
		return ""
	}, // }}}
	"append": func(arg ...interface{}) string { // {{{
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return fmt.Sprintf("%v", m)
			}
			switch which := arg[1].(type) {
			case string:
				if len(arg) == 2 {
					return m.Append(which)
				}

				return m.Append(which, arg[2:]...)
			}
		}
		return ""
	}, // }}}
}

var Pulse = &Message{
	code:     0,
	time:     time.Now(),
	Wait:     make(chan bool),
	source:   Index,
	master:   Index,
	target:   Index,
	Sessions: make(map[string]*Message),
}
var Index = &Context{Name: "ctx", Help: "模块中心",
	Caches: map[string]*Cache{
		"debug":    &Cache{Name: "服务数量", Value: "true", Help: "显示已经启动运行模块的数量"},
		"nserver":  &Cache{Name: "服务数量", Value: "0", Help: "显示已经启动运行模块的数量"},
		"ncontext": &Cache{Name: "模块数量", Value: "0", Help: "显示功能树已经注册模块的数量"},
		"nmessage": &Cache{Name: "消息数量", Value: "0", Help: "显示模块启动时所创建消息的数量"},
	},
	Configs: map[string]*Config{
		"debug":     &Config{Name: "调试模式(true/false)", Value: "true", Help: "是否打印错误信息，off:不打印，on:打印)"},
		"default":   &Config{Name: "默认的搜索起点(root/back/home)", Value: "root", Help: "模块搜索的默认起点，root:从根模块，back:从父模块，home:从当前模块"},
		"start":     &Config{Name: "启动模块", Value: "cli", Help: "启动时自动运行的模块"},
		"init.shy":  &Config{Name: "启动脚本", Value: "etc/init.shy", Help: "模块启动时自动运行的脚本"},
		"bench.log": &Config{Name: "日志文件", Value: "var/bench.log", Help: "模块日志输出的文件"},
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

		"command_list_base": &Config{Name: "命令列表的起始位置", Value: "0", Help: "命令列表的起始位置"},

		"search_method": &Config{Name: "search_method(find/search)", Value: "search", Help: "搜索方法, find: 模块名精确匹配, search: 模块名或帮助信息模糊匹配"},
		"search_choice": &Config{Name: "search_choice(first/last/rand/magic)", Value: "magic", Help: "搜索方法, find: 模块名精确匹配, search: 模块名或帮助信息模糊匹配"},
		"search_action": &Config{Name: "search_method(list/switch)", Value: "switch", Help: "搜索方法, find: 模块名精确匹配, search: 模块名或帮助信息模糊匹配"},
		"search_root":   &Config{Name: "search_root(true/false)", Value: "true", Help: "搜索方法, find: 模块名精确匹配, search: 模块名或帮助信息模糊匹配"},
		"search_index":  &Config{Name: "search_index", Value: "0", Help: "搜索索引"},

		"detail_index": &Config{Name: "detail_index", Value: "0", Help: "参数的索引"},
		"result_index": &Config{Name: "result_index", Value: "-2", Help: "返回值的索引"},
	},
	Commands: map[string]*Command{
		"help": &Command{Name: "help topic", Help: "帮助", Hand: func(m *Message, c *Context, key string, arg ...string) {
			if len(arg) == 0 { // {{{
				m.Echo("^_^  Welcome to context world  ^_^\n")
				m.Echo("\n")
				m.Echo("Context is to be a distributed operating system, try to simple everything in work and life. ")
				m.Echo("In context you will find all kinds of tools, and you can also make new tool in a quick and easy way.\n")
				m.Echo("Here is just a simple introduce, you can look github.com/shylinux/context/README.md for more information.\n")
				m.Echo("\n")
				m.Color(31, "       c\n")
				m.Color(31, "     sh").Color(33, " go\n")
				m.Color(31, "   vi").Color(32, " php").Color(32, " js\n")
				m.Echo(" ARM Linux HTTP\n")
				m.Echo("\n")

				m.Color(31, "Context ").Color(32, "Message\n")
				m.Color(32, "ctx ").Color(33, "cli ").Color(31, "aaa ").Color(33, "web\n")
				m.Color(32, "lex ").Color(33, "yac ").Color(31, "log ").Color(33, "gdb\n")
				m.Color(32, "tcp ").Color(33, "nfs ").Color(31, "ssh ").Color(33, "mdb\n")
				m.Color(31, "script ").Color(32, "template\n")
				return
			}

			switch arg[0] {
			case "example":
			case "context":
			case "message":
			}
			// }}}
		}},
		"message": &Command{Name: "message code meta index", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			switch len(arg) { // {{{
			case 0:
				pulse := m.target.Pulse
				if pulse != nil {
					m.Echo("\033[31mPulse:\033[0m\n")
					m.Echo("%d %s\n", pulse.code, pulse.Format())
				}

				m.Echo("\033[31mrequests:\033[0m\n")
				for i, v := range m.target.Requests {
					m.Echo("%d %s\n", i, v.Format())
					for i, v := range v.messages {
						if v.Detail(0) != "log" {
							m.Echo("  %d %s\n", i, v.Format())
						}
					}
				}

				m.Echo("\033[32msessions:\033[0m\n")
				for k, v := range m.Sessions {
					m.Echo("%s %s\n", k, v.Format())
				}

				m.Echo("\033[33mhistorys:\033[0m\n")
				for i, v := range m.target.Historys {
					m.Echo("%d %s\n", i, v.Format())
					for i, v := range v.messages {
						if v.Detail(0) != "log" {
							m.Echo("  %d %s\n", i, v.Format())
						}
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

						m.Echo("sessions:\n")
						for k, v := range ms[i].Sessions {
							m.Echo("  %s: %s\n", k, v.Format())
						}

						if ms[i].callback.hand != nil {
							m.Echo("callback: %d\n", ms[i].callback.ncall)
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
			default:
				n, e := strconv.Atoi(arg[0])
				m.Assert(e)

				ms := []*Message{m.root}
				for i := 0; i < len(ms); i++ {
					if ms[i].code == n {
						switch arg[1] {
						case "ncol":
							m.Echo("%d", len(ms[i].Meta["append"]))
							return
						case "nrow":
							m.Echo("%d", len(ms[i].Meta[ms[i].Meta["append"][0]]))
							return
						case "option", "session", "callback", "feedback":
							msg := ms[i].Spawn(ms[i].target)
							msg.Cmd(arg[1:])
							m.Copy(msg, "result")
							return
						default:
							index := 0
							if len(arg) > 2 {
								n, e := strconv.Atoi(arg[2])
								m.Assert(e)
								index = n
							}

							if meta, ok := ms[i].Meta[arg[1]]; ok {
								m.Echo(meta[index])
							}
						}

					}
					ms = append(ms, ms[i].messages...)
				}
			}

			// }}}
		}},
		"detail": &Command{Name: "detail [index] [value...]", Help: "查看或添加参数", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.Sesss("cli", false) // {{{
			if len(arg) == 0 {
				m.Echo("%v\n", msg.Meta["detail"])
				return
			}

			index := m.Confi("detail_index")
			if i, e := strconv.Atoi(arg[0]); e == nil {
				index, arg = i, arg[1:]
			}
			m.Echo("%s", msg.Detail(index, arg))
			// }}}
		}},
		"option": &Command{Name: "option [key [value...]]", Help: "查看或添加选项", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.Sesss("cli", false) // {{{
			if len(arg) == 0 {
				keys := []string{}
				values := map[string][]string{}

				for msg = msg; msg != nil; msg = msg.message {
					for _, k := range msg.Meta["option"] {
						if _, ok := values[k]; ok {
							continue
						}
						keys = append(keys, k)
						values[k] = msg.Meta[k]
					}
				}

				sort.Strings(keys)
				for _, k := range keys {
					m.Echo("%s: %v\n", k, values[k])
				}
				return
			}

			m.Echo("%s", msg.Option(arg[0], arg[1:]))
			// }}}
		}},
		"result": &Command{Name: "result [value...]", Help: "查看或添加返回值", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.Sesss("cli", false) // {{{
			if len(arg) == 0 {
				m.Echo("%v\n", msg.Meta["result"])
				return
			}

			index := m.Confi("result_index")
			if i, e := strconv.Atoi(arg[0]); e == nil {
				index, arg = i, arg[1:]
			}
			m.Echo("%s", msg.Result(index, arg))
			// }}}
		}},
		"append": &Command{Name: "append [key [value...]]", Help: "查看或添加附加值", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.Sesss("cli", false) // {{{
			if len(arg) == 0 {
				keys := []string{}
				values := map[string][]string{}

				ms := []*Message{msg}
				for i := 0; i < len(ms); i++ {
					ms = append(ms, ms[i].messages...)
					for _, k := range ms[i].Meta["append"] {
						if _, ok := values[k]; ok {
							continue
						}

						keys = append(keys, k)
						values[k] = ms[i].Meta[k]
					}
				}

				sort.Strings(keys)
				for _, k := range keys {
					m.Echo("%s: %v\n", k, values[k])
				}
				return
			}

			m.Echo("%s", msg.Append(arg[0], arg[1:]))
			// }}}
		}},
		"session": &Command{Name: "session [key [cmd...]]", Help: "查看或添加会话", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.Sesss("cli", false) // {{{
			if len(arg) == 0 {
				for msg = msg; msg != nil; msg = msg.message {
					for k, v := range msg.Sessions {
						m.Echo("%s: %d(%s->%s) %v %v\n", k, v.code, v.source.Name, v.target.Name, v.Meta["detail"], v.Meta["option"])
					}
				}
				return
			}

			var sub *Message
			for m := msg; m != nil; m = m.message {
				for k, v := range m.Sessions {
					if k == arg[0] {
						sub = v
					}
				}
			}

			if len(arg) == 1 {
				if sub != nil {
					m.Echo("%d", sub.code)
				}
				return
			}

			if sub == nil {
				sub = msg
			}

			cmd := msg.Spawn(sub.target).Cmd(arg[1:])
			m.Copy(cmd, "result").Copy(cmd, "append")
			msg.Sessions[arg[0]] = cmd
			// }}}
		}},
		"callback": &Command{Name: "callback index", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			index := 0
			for msg := m.Sesss("cli", false); msg != nil; msg = msg.message { // {{{
				index--
				switch len(arg) {
				case 0:
					if msg.callback.hand == nil {
						return
					}

					m.Echo("%d: %v\n", index, msg.callback.hand)
				case 1:
					if msg.callback.hand == nil {
						return
					}

					if i, e := strconv.Atoi(arg[0]); e == nil && i == index {
						m.Echo("%v", msg.callback.hand)
					}
				default:
					if i, e := strconv.Atoi(arg[0]); e == nil && i == index {
						msg.callback.hand = func(msg *Message) *Message {
							return msg
						}
						m.Echo("%v", msg.callback.hand)
					}
					return
				}
				continue

				if len(arg) == 0 {
					m.Echo("msg(%s->%s): %d(%s) %v\n", msg.source.Name, msg.target.Name, msg.code, msg.time.Format("15:04:05"), msg.Meta["detail"])
					if msg.callback.hand != nil {
						m.Echo("  hand(%d): %v\n", msg.callback.ncall, msg.callback.hand)
					}
				} else {
					switch arg[0] {
					case "del":
						msg.message.callback.hand = nil
					case "add":
						msg.message.callback.hand = func(msg *Message) *Message {
							msg.Log("info", nil, "callback default")
							return msg
						}
						return
					default:
						m.Result(0, arg)
						msg.message.Back(m)
						return
					}
				}
			}
			// }}}
		}},
		"context": &Command{
			Name: "context [[find [root|home]|search [root|home] [name|help] [magic|rand|first|last]] name] [list|info|cache|config|command|switch] [args]",
			Help: "查找并操作模块，\n查找起点root:根模块、back:父模块、home:本模块，\n查找方法find:路径匹配、search:模糊匹配，\n查找对象name:支持点分和正则，\n操作类型show:显示信息、switch:切换为当前、start:启动模块、spawn:分裂子模块，args:启动参数",
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				action := m.Conf("search_action") // {{{
				if len(arg) == 0 {
					action = "list"
				}

				method := m.Conf("search_method")
				if len(arg) > 0 {
					switch arg[0] {
					case "find", "search":
						method = arg[0]
						arg = arg[1:]
					}
				}

				root := m.Confs("search_root")
				if len(arg) > 0 {
					switch arg[0] {
					case "root":
						root = true
						arg = arg[1:]
					case "home":
						root = false
						arg = arg[1:]
					}
				}

				ms := []*Message{}
				if len(arg) > 0 {
					switch method {
					case "find":
						if msg := m.Find(arg[0], root); msg != nil {
							ms = append(ms, msg)
						}
						arg = arg[1:]
					case "search":
						choice := m.Conf("search_choice")
						switch arg[0] {
						case "magic", "rand", "first", "last":
							choice = arg[0]
							arg = arg[1:]
						}

						if s := m.Search(arg[0], root); len(s) > 0 {
							switch choice {
							case "first":
								ms = append(ms, s[0])
							case "last":
								ms = append(ms, s[len(s)-1])
							case "rand":
								ms = append(ms, s[rand.Intn(len(s))])
							case "magic":
								ms = append(ms, s...)
							}
						}
						arg = arg[1:]
					}
				} else {
					ms = append(ms, m)
				}

				if len(arg) == 0 {
					arg = []string{action}
				}

				for _, msg := range ms {
					switch arg[0] {
					case "switch":
						m.target = msg.target
					case "list":
						which := ""
						if len(arg) > 1 {
							which = arg[1]
						}
						m.Log("fuck", nil, "what %s", which)
						switch which {
						case "cache":
							for k, v := range msg.target.Caches {
								m.Add("append", "key", k)
								m.Add("append", "name", v.Name)
								m.Add("append", "value", v.Value)
								m.Add("append", "help", v.Help)
							}
						case "config":
							for k, v := range msg.target.Configs {
								m.Add("append", "key", k)
								m.Add("append", "name", v.Name)
								m.Add("append", "value", v.Value)
								m.Add("append", "help", v.Help)
							}
						case "command":
							for k, v := range msg.target.Commands {
								m.Add("append", "key", k)
								m.Add("append", "name", v.Name)
								m.Add("append", "help", v.Help)
							}
						case "module":
							m.Travel(msg.target, func(msg *Message) bool {
								m.Add("append", "name", msg.target.Name)
								m.Add("append", "help", msg.target.Help)
								m.Add("append", "module", msg.Cap("module"))
								m.Add("append", "status", msg.Cap("status"))
								m.Add("append", "stream", msg.Cap("stream"))
								return true
							})
						case "domain":
							msg := m.Find("ssh", true)
							msg.Travel(msg.target, func(msg *Message) bool {
								m.Add("append", "name", msg.target.Name)
								m.Add("append", "help", msg.target.Help)
								m.Add("append", "domain", msg.Cap("domain")+"."+msg.Conf("domains"))
								return true
							})
						default:
							msg.Travel(msg.target, func(msg *Message) bool {
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
						}
					default:
						msg.Cmd(arg)
						m.Meta["result"] = append(m.Meta["result"], msg.Meta["result"]...)
						m.Copy(msg, "append")
						m.target = msg.target
					}
				}
				// }}}
			}},
		"server": &Command{
			Name: "server [spawn|begin|start|close][args]",
			Help: "查看、新建、初始化、启动、停止服务",
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				switch len(arg) { // {{{
				case 0:
					m.Travel(m.target.root, func(m *Message) bool {
						if m.Cap("status") == "start" {
							m.Echo("%s(%s): %s\n", m.target.Name, m.Cap("stream"), m.target.Help)
						}
						return true
					})

				default:
					switch arg[0] {
					case "spawn":
						if len(arg) > 2 {
							msg := m.Spawn().Set("detail", arg[3:]...)
							msg.target.Spawn(msg, arg[1], arg[2])
							m.target = msg.target
						}
					case "begin":
						msg := m.Spawn().Set("detail", arg...)
						msg.target.Begin(msg)
					case "start":
						msg := m.Spawn().Set("detail", arg...)
						msg.target.Start(msg)
					case "close":
						msg := m.Spawn().Set("detail", arg...)
						m.target = msg.target.context
						msg.target.Close(msg)
					}
				}
				// }}}
			}},
		"command": &Command{
			Name: "command [all|add cmd arg...|list [begin [end]]|test [begin [end]]|delete cmd]", Help: "查看或修改命令",
			Form: map[string]int{"condition": -1},
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				if len(arg) == 0 { // {{{
					keys := []string{}
					for k, _ := range m.target.Commands {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						m.Echo("%s: %s\n", k, m.target.Commands[k].Name)
					}
					return
				}
				switch arg[0] {
				case "all":
					keys := []string{}
					values := map[string]*Command{}
					for s := m.target; s != nil; s = s.context {
						for k, v := range s.Commands {
							if _, ok := values[k]; !ok {
								keys = append(keys, k)
								values[k] = v
							}
						}
					}
					sort.Strings(keys)
					for _, k := range keys {
						m.Echo("%s: %s\n", k, values[k].Name)
					}
					return
				case "add":
					if m.target.Caches == nil {
						m.target.Caches = map[string]*Cache{}
					}
					if _, ok := m.target.Caches["list_count"]; !ok {
						m.target.Caches["list_count"] = &Cache{Name: "list_count", Value: "0", Help: "list_count"}
						m.target.Caches["list_begin"] = &Cache{Name: "list_begin", Value: "0", Help: "list_begin"}
					}

					if m.target.Commands == nil {
						m.target.Commands = map[string]*Command{}
					}
					m.target.Commands[m.Cap("list_count")] = &Command{
						Name: strings.Join(arg[1:], " "),
						Help: "list_command",
						Hand: func(m *Message, c *Context, key string, args ...string) {
							list := []string{}
							j := 0
							for i := 1; i < len(arg); i++ {
								if arg[i] == "_" && j < len(args) {
									list = append(list, args[j])
									j++
									continue
								}
								list = append(list, arg[i])
							}
							list = append(list, args[j:]...)

							msg := m.Spawn().Cmd(list)
							m.Copy(msg, "result").Copy(msg, "append")
						},
					}
					m.Capi("list_count", 1)
					return
				case "list":
					begin, end := m.Capi("list_begin"), m.Capi("list_count")
					if len(arg) > 1 {
						n, e := strconv.Atoi(arg[1])
						m.Assert(e)
						begin = n
					}
					if len(arg) > 2 {
						n, e := strconv.Atoi(arg[2])
						m.Assert(e)
						end = n
					}
					for i := begin; i < end; i++ {
						if c, ok := m.target.Commands[fmt.Sprintf("%d", i)]; ok {
							m.Echo("%d: %s\n", i, c.Name)
						}
					}
					return
				case "test":
					begin, end := 0, m.Capi("list_count")
					if len(arg) > 1 {
						n, e := strconv.Atoi(arg[1])
						m.Assert(e)
						begin = n
					}
					if len(arg) > 2 {
						n, e := strconv.Atoi(arg[2])
						m.Assert(e)
						end = n
					}

					success, failure := 0, 0
					for i := begin; i < end; i++ {
						key := fmt.Sprintf("%d", i)
						if c, ok := m.target.Commands[key]; ok {
							msg := m.Spawn().Cmd(key)
							if m.Options("condition") {
								done := true
								condition := m.Meta["condition"]
								for j := 0; j < len(condition)-1; j += 2 {
									if !msg.Has(condition[j]) || msg.Append(condition[j]) != condition[j+1] {
										m.Color(31, "%s %s %s\n", key, " fail", c.Name)
										done = false
										failure++
									}
								}
								if done {
									// m.Echo("%s %s\n", key, " done")
									m.Echo("%s %s %s\n", key, " done", c.Name)
									success++
								}
							} else {
								for _, v := range msg.Meta["result"] {
									m.Echo("%v", v)
								}
								m.Echo("\n")
								success++
							}
						}
					}
					m.Color(32, "success: %d, ", success)
					m.Color(31, "failure: %d, ", failure)
					m.Color(33, "total: %d", success+failure)
					return
				case "delete":
					if _, ok := m.target.Commands[arg[1]]; ok {
						delete(m.target.Commands, arg[1])
					}
				}
				// }}}
			}},
		"config": &Command{
			Name: "config [all|key [value]|key name value help|delete key]",
			Help: "查看、读写、添加配置变量",
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				switch len(arg) { //{{{
				case 0:
					keys := []string{}
					for k, _ := range m.target.Configs {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						m.Echo("%s(%s): %s\n", k, m.Conf(k), m.target.Configs[k].Name)
					}
					return
				case 1:
					if arg[0] == "all" {
						keys := []string{}
						values := map[string]*Config{}
						for s := m.target; s != nil; s = s.context {
							for k, v := range s.Configs {
								if _, ok := values[k]; !ok {
									keys = append(keys, k)
									values[k] = v
								}
							}
						}
						sort.Strings(keys)
						for _, k := range keys {
							m.Echo("%s(%s): %s\n", k, m.Conf(k), values[k].Name)
						}
						return
					}
				case 2:
					if arg[0] == "delete" {
						if _, ok := m.target.Configs[arg[1]]; ok {
							delete(m.target.Configs, arg[1])
						}
						return
					}
					m.Conf(arg[0], arg[1])
				case 3:
					m.Conf(arg[0], arg[0], arg[2], arg[0])
				default:
					m.Conf(arg[0], arg[1:]...)
				}
				m.Echo("%s", m.Conf(arg[0]))
				// }}}
			}},
		"cache": &Command{
			Name: "cache [all|key [value]|key = value|key name value help|delete key]",
			Help: "查看、读写、赋值、新建、删除缓存变量",
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				switch len(arg) { //{{{
				case 0:
					keys := []string{}
					for k, _ := range m.target.Caches {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						m.Echo("%s(%s): %s\n", k, m.Cap(k), m.target.Caches[k].Name)
					}
					return
				case 1:
					if arg[0] == "all" {
						keys := []string{}
						values := map[string]*Cache{}
						for s := m.target; s != nil; s = s.context {
							for k, v := range s.Caches {
								if _, ok := values[k]; !ok {
									keys = append(keys, k)
									values[k] = v
								}
							}
						}
						sort.Strings(keys)
						for _, k := range keys {
							m.Echo("%s(%s): %s\n", k, m.Cap(k), values[k].Name)
						}
						return
					}
				case 2:
					if arg[0] == "delete" {
						if _, ok := m.target.Caches[arg[1]]; ok {
							delete(m.target.Caches, arg[1])
						}
						return
					}
					m.Cap(arg[0], arg[1])
				case 3:
					m.Cap(arg[0], arg[0], arg[2], arg[0])
				default:
					m.Cap(arg[0], arg[1:]...)
				}
				m.Echo("%s", m.Cap(arg[0]))
				// }}}
			}},
		"right": &Command{
			Name: "right [share|add|del group [cache|config|command item]]",
			Help: "用户组管理，查看、添加、删除用户组或是接口",
			Form: map[string]int{"check": 0, "add": 0, "del": 0, "brow": 0, "cache": 0, "config": 0, "command": 0},
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				index := m.Target().Index // {{{
				if index == nil {
					m.Target().Index = map[string]*Context{}
				}

				current := m.Target()
				aaa := m.Sesss("aaa")
				void := index["void"]
				if aaa != nil && aaa.Cap("group") != aaa.Conf("rootname") {
					if current = index[aaa.Cap("group")]; current == nil {
						if void != nil {
							m.Echo("%s:caches\n", void.Name)
							for k, c := range void.Caches {
								m.Echo("  %s: %s\n", k, c.Value)
							}
							m.Echo("%s:configs\n", void.Name)
							for k, c := range void.Configs {
								m.Echo("  %s: %s\n", k, c.Value)
							}
							m.Echo("%s:commands\n", void.Name)
							for k, c := range void.Commands {
								m.Echo("  %s: %s\n", k, c.Name)
							}
							m.Echo("%s:contexts\n", void.Name)
							for k, c := range void.Index {
								m.Echo("  %s: %s\n", k, c.Name)
							}
						}
						return
					}
				}

				group := current
				if len(arg) > 0 {
					group = current.Index[arg[0]]
				}

				item := ""
				if len(arg) > 1 {
					item = arg[1]
				}

				switch {
				case m.Has("check"):
					if group != nil {
						switch {
						case m.Has("cache"):
							if _, ok := group.Caches[item]; ok {
								m.Echo("ok")
								return
							}
						case m.Has("config"):
							if _, ok := group.Configs[item]; ok {
								m.Echo("ok")
								return
							}
						case m.Has("command"):

							if len(arg) > 1 {
								if _, ok := group.Commands[item]; !ok {
									m.Echo("no")
									return
								}
							}
							if len(arg) > 2 {
								if _, ok := group.Commands[item].Shares[arg[2]]; !ok {
									m.Echo("no")
									return
								}
							}
							if len(arg) > 3 {
								for _, v := range group.Commands[item].Shares[arg[2]] {
									match, e := regexp.MatchString(v, arg[3])
									m.Assert(e)
									if match {
										m.Echo("ok")
										return
									}
								}
								m.Echo("no")
								return
							}
							m.Echo("ok")
							return
						default:
							m.Echo("ok")
							return
						}
					}
					m.Echo("no")
					return
				case m.Has("add"):
					if group == nil {
						if _, ok := index[arg[0]]; ok {
							break
						}
						group = &Context{Name: arg[0]}
					}

					switch {
					case m.Has("cache"):
						if x, ok := current.Caches[item]; ok {
							if group.Caches == nil {
								group.Caches = map[string]*Cache{}
							}
							group.Caches[item] = x
						}
					case m.Has("config"):
						if x, ok := current.Configs[item]; ok {
							if group.Configs == nil {
								group.Configs = map[string]*Config{}
							}
							group.Configs[item] = x
						}
					case m.Has("command"):
						if _, ok := current.Commands[item]; ok {
							if group.Commands == nil {
								group.Commands = map[string]*Command{}
							}

							command, ok := group.Commands[item]
							if !ok {
								command = &Command{Shares: map[string][]string{}}
								group.Commands[item] = command
							}

							for i := 2; i < len(arg)-1; i += 2 {
								command.Shares[arg[i]] = append(command.Shares[arg[i]], arg[i+1])
							}

							// group.Commands[item] = x
						}
					}

					if current.Index == nil {
						current.Index = map[string]*Context{}
					}
					current.Index[arg[0]] = group
					index[arg[0]] = group

				case m.Has("del"):
					if group == nil {
						break
					}

					gs := []*Context{group}
					for i := 0; i < len(gs); i++ {
						for _, g := range gs[i].Index {
							gs = append(gs, g)
						}

						switch {
						case m.Has("cache"):
							delete(gs[i].Caches, item)
						case m.Has("config"):
							delete(gs[i].Configs, item)
						case m.Has("command"):
							if gs[i].Commands == nil {
								break
							}
							if len(arg) == 2 {
								delete(gs[i].Commands, item)
								break
							}

							if gs[i].Commands[item] == nil {
								break
							}
							shares := gs[i].Commands[item].Shares
							if shares == nil {
								break
							}
							if len(arg) == 3 {
								delete(shares, arg[2])
								break
							}

							for i := 0; i < len(shares[arg[2]]); i++ {
								if shares[arg[2]][i] == arg[3] {
									for ; i < len(shares[arg[2]])-1; i++ {
										shares[arg[2]][i] = shares[arg[2]][i+1]
									}
									shares[arg[2]] = shares[arg[2]][:i]
								}
							}
							gs[i].Commands[item].Shares = shares

						default:
							delete(index, gs[i].Name)
							delete(current.Index, gs[i].Name)
						}
					}

				default:
					m.Echo("%s:caches\n", group.Name)
					if void != nil {
						for k, c := range void.Caches {
							m.Echo("  %s: %s\n", k, c.Value)
						}
					}
					for k, c := range group.Caches {
						m.Echo("  %s: %s\n", k, c.Value)
					}
					m.Echo("%s:configs\n", group.Name)
					if void != nil {
						for k, c := range void.Configs {
							m.Echo("  %s: %s\n", k, c.Value)
						}
					}
					for k, c := range group.Configs {
						m.Echo("  %s: %s\n", k, c.Value)
					}
					m.Echo("%s:commands\n", group.Name)
					if void != nil {
						for k, c := range void.Commands {
							m.Echo("  %s: %s\n", k, c.Name)
						}
					}
					for k, c := range group.Commands {
						m.Echo("  %s: %s %v\n", k, c.Name, c.Shares)
					}
					m.Echo("%s:contexts\n", group.Name)
					if void != nil {
						for k, c := range void.Index {
							m.Echo("  %s: %s\n", k, c.Name)
						}
					}
					for k, c := range group.Index {
						m.Echo("  %s: %s\n", k, c.Name)
					}
				} // }}}
			}},
	},
	Index: map[string]*Context{
		"void": &Context{Name: "void",
			Caches: map[string]*Cache{
				"nmessage": &Cache{},
				"ncontext": &Cache{},
				"nserver":  &Cache{},
			},
			Configs: map[string]*Config{
				"debug":     &Config{},
				"bench.log": &Config{},
			},
			Commands: map[string]*Command{
				"message":  &Command{},
				"option":   &Command{},
				"session":  &Command{},
				"callback": &Command{},
				"feedback": &Command{},

				"context": &Command{},
				"server":  &Command{},
				"command": &Command{},
				"config":  &Command{},
				"cache":   &Command{},

				"right": &Command{},
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
		Pulse.Conf("init.shy", args[0])
	}
	if len(args) > 1 {
		Pulse.Conf("bench.log", args[1])
	}
	if len(args) > 2 {
		Pulse.Conf("root", args[2])
	}
	if len(args) > 3 {
		Pulse.Conf("start", args[3])
	}

	Index.Owner = Index.contexts["aaa"]
	Index.master = Index.contexts["cli"]
	for _, m := range Pulse.Search("") {
		m.target.root = Index
		m.target.Begin(m)
	}

	Pulse.Options("log", true)
	Pulse.Sesss("log", "log").Conf("bench.log", Pulse.Conf("bench.log"))

	for _, m := range Pulse.Search(Pulse.Conf("start")) {
		m.Set("detail", Pulse.Conf("init.shy")).Set("option", "stdio").target.Start(m)
	}

	<-Index.master.Exit
}
