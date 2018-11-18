package ctx

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"os"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Right(arg interface{}) bool {
	switch str := arg.(type) {
	case nil:
		return false
	case bool:
		return str
	case string:
		switch str {
		case "", "0", "false", "off", "no", "error: ":
			return false
		}
		return true
	}
	return false
}
func Trans(arg ...interface{}) []string {
	ls := []string{}
	for _, v := range arg {
		switch val := v.(type) {
		case *Message:
			if val.Hand {
				ls = append(ls, val.Meta["result"]...)
			} else {
				ls = append(ls, val.Meta["detail"]...)
			}
		case string:
			ls = append(ls, val)
		case bool:
			ls = append(ls, fmt.Sprintf("%t", val))
		case int, int8, int16, int32, int64:
			ls = append(ls, fmt.Sprintf("%d", val))
		case float64:
			ls = append(ls, fmt.Sprintf("%d", int(val)))
		case []interface{}:
			for _, v := range val {
				switch val := v.(type) {
				case string:
					ls = append(ls, val)
				case bool:
					ls = append(ls, fmt.Sprintf("%t", val))
				case int, int8, int16, int32, int64:
					ls = append(ls, fmt.Sprintf("%d", val))
				}
			}
		case []string:
			ls = append(ls, val...)
		case []bool:
			for _, v := range val {
				ls = append(ls, fmt.Sprintf("%t", v))
			}
		case []int:
			for _, v := range val {
				ls = append(ls, fmt.Sprintf("%d", v))
			}
		case nil:
		default:
			ls = append(ls, fmt.Sprintf("%v", val))
		}
	}
	return ls
}
func Chain(m *Message, data interface{}, args ...interface{}) interface{} {
	// if len(args) == 1 {
	// 	if arg, ok := args[0].([]string); ok {
	// 		args = args[:0]
	// 		for _, v := range arg {
	// 			args = append(args, v)
	// 		}
	// 	}
	// }
	//
	root := data
	for i := 0; i < len(args); i += 2 {
		var parent interface{}
		parent_key, parent_index := "", 0
		data, root = root, nil

		keys := []string{}
		switch arg := args[i].(type) {
		case map[string]interface{}:
			args = args[:0]
			for k, v := range arg {
				args = append(args, k, v)
			}
			i = -2
			continue
		case []interface{}:
			keys = []string{}
			for _, v := range arg {
				keys = append(keys, fmt.Sprintf("%v", v))
			}
		case []string:
			keys = arg
			keys = strings.Split(strings.Join(arg, "."), ".")
		case string:
			keys = strings.Split(arg, ".")
		case nil:
			continue
		default:
			keys = append(keys, fmt.Sprintf("%v", arg))
		}

		for j, k := range keys {
			switch value := data.(type) {
			case nil:
				if i == len(args)-1 {
					return nil
				}

				if _, e := strconv.Atoi(k); e == nil {
					node := []interface{}{nil}
					switch p := parent.(type) {
					case map[string]interface{}:
						p[parent_key] = node
					case []interface{}:
						p[parent_index] = node
					}
					if data, parent_index = node, 0; j == len(keys)-1 {
						node[0] = args[i+1]
					}
				} else {
					node := map[string]interface{}{}
					switch p := parent.(type) {
					case map[string]interface{}:
						p[parent_key] = node
					case []interface{}:
						p[parent_index] = node
					}
					if data, parent_key = node, k; j == len(keys)-1 {
						node[k] = args[i+1]
					}
				}

				parent, data = data, nil
			case []string:
				if index, e := strconv.Atoi(k); e == nil {
					index = (index+2+len(value)+2)%(len(value)+2) - 2
					if i == len(args)-1 {
						if index < 0 {
							return ""
						}
						return value[index]
					}
					switch index {
					case -1:
						return append([]string{args[i+1].(string)}, value...)
					case -2:
						return append(value, args[i+1].(string))
					default:
						value[index] = args[i+1].(string)
					}
				}

			case map[string]string:
				if i == len(args)-1 {
					return value[k]
				}
				value[k] = args[i+1].(string)
			case map[string]interface{}:
				if j == len(keys)-1 {
					if i == len(args)-1 {
						return value[k]
					}
					value[k] = args[i+1]
				}
				parent, data, parent_key = data, value[k], k
			case []interface{}:
				index, e := strconv.Atoi(k)
				if e != nil {
					return nil
				}
				index = (index+2+len(value)+2)%(len(value)+2) - 2

				if i == len(args)-1 {
					if index < 0 {
						return nil
					}
					if j == len(keys)-1 {
						return value[index]
					}
				} else {
					if index == -1 {
						value = append([]interface{}{nil}, value...)
						index = 0
					} else if index == -2 {
						value = append(value, nil)
						index = len(value) - 1
					}

					if j == len(keys)-1 {
						value[index] = args[i+1]
					}
				}
				switch p := parent.(type) {
				case map[string]interface{}:
					p[parent_key] = value
				case []interface{}:
					p[parent_index] = value
				}
				parent, data, parent_index = value, value[index], index
			}

			if root == nil {
				root = parent
			}
		}
	}

	return root
}
func Elect(last interface{}, args ...interface{}) string {
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case []string:
			index := 0
			if len(args) > 1 {
				switch a := args[1].(type) {
				case string:
					i, e := strconv.Atoi(a)
					if e == nil {
						index = i
					}
				case int:
					index = a
				}
			}

			if 0 <= index && index < len(arg) && arg[index] != "" {
				return arg[index]
			}
		case string:
			if arg != "" {
				return arg
			}
		}
	}

	switch l := last.(type) {
	case string:
		return l
	}
	return ""
}

func Array(m *Message, list []string, index int, arg ...interface{}) []string {
	if len(arg) == 0 {
		if -1 < index && index < len(list) {
			return []string{list[index]}
		}
		return []string{""}
	}

	index = (index+2)%(len(list)+2) - 2

	str := Trans(arg...)

	if index == -1 {
		index, list = 0, append(str, list...)
	} else if index == -2 {
		index, list = len(list), append(list, str...)
	} else {
		if index < -2 {
			index += len(list) + 2
		}
		if index < 0 {
			index = 0
		}

		for i := len(list); i < index+len(str); i++ {
			list = append(list, "")
		}
		for i := 0; i < len(str); i++ {
			list[index+i] = str[i]
		}
	}

	return list
}

type Cache struct {
	Name  string
	Value string
	Help  string
	Hand  func(m *Message, x *Cache, arg ...string) string
}
type Config struct {
	Name  string
	Value interface{}
	Help  string
	Hand  func(m *Message, x *Config, arg ...string) string
}
type Command struct {
	Name string
	Help interface{}
	Form map[string]int
	Hand func(m *Message, c *Context, key string, arg ...string)

	Shares map[string][]string
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
	Index    map[string]*Context

	requests []*Message
	sessions []*Message
	message  *Message

	contexts map[string]*Context
	context  *Context
	root     *Context

	exit chan bool
	Server
}

func (c *Context) Register(s *Context, x Server) {
	if c.contexts == nil {
		c.contexts = make(map[string]*Context)
	}
	if x, ok := c.contexts[s.Name]; ok {
		panic(errors.New(c.Name + "上下文中已存在模块:" + x.Name))
	}

	c.contexts[s.Name] = s
	s.context = c
	s.Server = x
}
func (c *Context) Spawn(m *Message, name string, help string) *Context {
	s := &Context{Name: name, Help: help, root: c.root, context: c}

	if m.target = s; c.Server != nil {
		c.Register(s, c.Server.Spawn(m, s, m.Meta["detail"]...))
	} else {
		c.Register(s, nil)
		s.Caches = map[string]*Cache{}
		s.Configs = map[string]*Config{}
	}

	item := []string{name}
	for s := c; s != nil; s = s.context {
		item = append(item, s.Name)
	}
	for i := 0; i < len(item)/2; i++ {
		item[i], item[len(item)-i-1] = item[len(item)-i-1], item[i]
	}
	s.Caches["module"] = &Cache{Name: "module", Value: strings.Join(item, "."), Help: "模块域名"}
	s.message = m
	return s
}
func (c *Context) Begin(m *Message, arg ...string) *Context {
	if len(arg) > 0 {
		m.Meta["detail"] = arg
	}

	item := []string{}
	for s := c; s != nil; s = s.context {
		item = append(item, s.Name)
	}
	for i := 0; i < len(item)/2; i++ {
		item[i], item[len(item)-i-1] = item[len(item)-i-1], item[i]
	}
	c.Caches["module"] = &Cache{Name: "module", Value: strings.Join(item, "."), Help: "模块域名"}
	c.Caches["status"] = &Cache{Name: "status(begin/start/close)", Value: "begin", Help: "模块状态，begin:初始完成，start:正在运行，close:未在运行"}
	c.Caches["stream"] = &Cache{Name: "stream", Value: "", Help: "模块数据"}

	c.requests = append(c.requests, m)
	m.source.sessions = append(m.source.sessions, m)
	c.message = m

	m.Log("begin", "%d context %v %v", m.root.Capi("ncontext", 1), m.Meta["detail"], m.Meta["option"])
	for k, x := range c.Configs {
		if x.Hand != nil {
			m.Conf(k, x.Value.(string))
		}
	}

	if c.Server != nil {
		c.Server.Begin(m, m.Meta["detail"]...)
	}

	return c
}
func (c *Context) Start(m *Message, arg ...string) bool {
	if len(arg) > 0 {
		m.Set("detail", arg...)
	}

	c.requests = append(c.requests, m)
	m.source.sessions = append(m.source.sessions, m)
	if m.Hand = true; m.Cap("status") == "start" {
		return true
	}

	m.Sess("log", m.Sess("log"))
	m.Sess("lex", m.Sess("lex"))
	if c.Name == "aaa" {
		debug.PrintStack()
		panic(1)
	}

	running := make(chan bool)
	go m.TryCatch(m, true, func(m *Message) {
		m.Log(m.Cap("status", "start"), "%d server %v %v", m.root.Capi("nserver", 1), m.Meta["detail"], m.Meta["option"])

		c.message = m
		c.exit = make(chan bool, 2)
		if running <- true; c.Server != nil && c.Server.Start(m, m.Meta["detail"]...) {
			c.Close(m, m.Meta["detail"]...)
		}
	})
	return <-running
}
func (c *Context) Close(m *Message, arg ...string) bool {
	if len(c.requests) == 0 {
		return true
	}
	m.Log("close", "%d:%d %v", len(c.requests), len(c.sessions), arg)

	if m.target == c {
		for i := len(c.requests) - 1; i >= 0; i-- {
			if msg := c.requests[i]; msg.code == m.code {
				if c.Server == nil || c.Server.Close(m, arg...) {
					for j := i; j < len(c.requests)-1; j++ {
						m.Log("close", "requests: %v %s", j, c.requests[j].Format())
						c.requests[j] = c.requests[j+1]
					}
					c.requests = c.requests[:len(c.requests)-1]
				}
			}
		}
	}

	if len(c.requests) > 0 {
		return false
	}

	if m.Cap("status") == "start" {
		m.Log(m.Cap("status", "close"), "%d server %v", m.root.Capi("nserver", -1)+1, arg)
		for _, msg := range c.sessions {
			if msg.Cap("status") == "start" {
				msg.target.Close(msg, arg...)
			}
		}
	}

	if c.context != nil {
		m.Log("close", "%d context %v", m.root.Capi("ncontext", -1)+1, arg)
		if c.Name != "stdio" {
			delete(c.context.contexts, c.Name)
		}
		c.exit <- true
	}
	return true
}
func (c *Context) Context() *Context {
	return c.context
}
func (c *Context) Message() *Message {
	return c.message
}
func (c *Context) Has(key ...string) bool {
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

type Message struct {
	code int
	time time.Time

	source *Context
	target *Context

	Hand bool
	Meta map[string][]string
	Data map[string]interface{}

	callback func(msg *Message) (sub *Message)
	Sessions map[string]*Message

	messages []*Message
	message  *Message
	root     *Message

	Remote chan bool
}

func (m *Message) Code() int {
	return m.code
}
func (m *Message) Time() string {
	return m.time.Format("2006-01-02 15:04:05")
}
func (m *Message) Message() *Message {
	return m.message
}
func (m *Message) Source() *Context {
	return m.source
}
func (m *Message) Target() *Context {
	return m.target
}
func (m *Message) Format() string {
	return fmt.Sprintf("%d(%s->%s): %s %v %v", m.code, m.source.Name, m.target.Name, m.time.Format("15:04:05"), m.Meta["detail"], m.Meta["option"])
}

func (m *Message) Tree(code int) *Message {
	ms := []*Message{m}
	for i := 0; i < len(ms); i++ {
		if ms[i].Code() == code {
			return ms[i]
		}
		ms = append(ms, ms[i].messages...)
	}
	return nil
}
func (m *Message) Copy(msg *Message, meta string, arg ...string) *Message {
	switch meta {
	case "target":
		m.target = msg.target
	case "callback":
		m.callback = msg.callback
	case "session":
		if len(arg) == 0 {
			for k, v := range msg.Sessions {
				m.Sessions[k] = v
			}
		} else {
			for _, k := range arg {
				m.Sessions[k] = msg.Sessions[k]
			}
		}
	case "detail", "result":
		if len(msg.Meta[meta]) > 0 {
			m.Add(meta, msg.Meta[meta][0], msg.Meta[meta][1:])
		}
	case "option", "append":
		if len(arg) == 0 {
			arg = msg.Meta[meta]
		}

		for _, k := range arg {
			if v, ok := msg.Data[k]; ok {
				m.Put(meta, k, v)
			}
			if v, ok := msg.Meta[k]; ok {
				m.Add(meta, k, v)
			}
		}
	}

	return m
}
func (m *Message) Log(action string, str string, arg ...interface{}) *Message {
	l := m.Sess("log", !m.Confs("compact_log"))
	if l == nil || m.Detail(0) == "log" || m.Detail(0) == "write" || m.Options("silent") {
		return m
	}

	l.Optionv("msg", m)
	l.Cmd("log", action, fmt.Sprintf(str, arg...))
	return m
}

func (m *Message) Assert(e interface{}, msg ...string) bool {
	switch v := e.(type) {
	case nil:
		return true
	case bool:
		if v {
			return true
		}
	case string:
		if Right(v) {
			return true
		}
	case *Message:
		if result, ok := v.Meta["result"]; ok && len(result) > 0 && result[0] == "error: " {
			e = v.Result(1)
			break
		}
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

	m.Log("error", "%s", fmt.Sprintln(e))
	panic(m.Set("result", "error: ", fmt.Sprintln(e), "\n"))
}
func (m *Message) TryCatch(msg *Message, safe bool, hand ...func(msg *Message)) *Message {
	defer func() {
		if e := recover(); e != nil && e != io.EOF {
			// switch v := e.(type) {
			// case *Message:
			// 	// msg.Log("error", "error: %v", strings.Join(v.Meta["result"][1:], ""))
			// default:
			// 	// msg.Log("error", "error: %v", e)
			// }

			if msg.root.Conf("debug") == "on" {
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

func (m *Message) Instance(msg *Message, source ...bool) bool {
	c := m.target
	if len(source) > 0 && source[0] == true {
		c = m.source
	}

	for s := c; s != nil; s = s.context {
		if s == msg.target {
			return true
		}
	}
	return false
}
func (m *Message) BackTrace(hand func(m *Message) bool, c ...*Context) *Message {
	target := m.target
	if len(c) > 0 {
		m.target = c[0]
	}
	for s := m.target; s != nil; s = s.context {
		if m.target = s; !hand(m) {
			break
		}
	}
	m.target = target
	return m
}
func (m *Message) Travel(hand func(m *Message, i int) bool, c ...*Context) {
	target := m.target
	if len(c) > 0 {
		m.target = c[0]
	}

	cs := []*Context{m.target}
	for i := 0; i < len(cs); i++ {
		if m.target = cs[i]; !hand(m, i) {
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

func (m *Message) Spawn(arg ...interface{}) *Message {
	c := m.target
	if len(arg) > 0 {
		switch v := arg[0].(type) {
		case *Context:
			c = v
		case *Message:
			c = v.target
		}
	}

	msg := &Message{
		code:    m.root.Capi("nmessage", 1),
		time:    time.Now(),
		source:  m.target,
		target:  c,
		message: m,
		root:    m.root,
	}

	m.messages = append(m.messages, msg)
	return msg
}
func (m *Message) Find(name string, root ...bool) *Message {
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
			m.Log("find", "not find %s", v)
			return nil
		}
	}
	m.Log("find", "find %s", name)
	return m.Spawn(target)
}
func (m *Message) Search(key string, root ...bool) []*Message {
	reg, e := regexp.Compile(key)
	m.Assert(e)

	target := m.target
	if len(root) > 0 && root[0] {
		target = m.target.root
	}

	cs := make([]*Context, 0, 3)
	m.Travel(func(m *Message, i int) bool {
		if reg.MatchString(m.target.Name) || reg.FindString(m.target.Help) != "" {
			m.Log("search", "%d match [%s]", len(cs)+1, key)
			cs = append(cs, m.target)
		}
		return true
	}, target)

	ms := make([]*Message, len(cs))
	for i := 0; i < len(cs); i++ {
		ms[i] = m.Spawn(cs[i])
	}
	if len(ms) == 0 {
		ms = append(ms, nil)
	}

	return ms
}
func (m *Message) Sess(key string, arg ...interface{}) *Message {
	spawn := true
	if len(arg) > 0 {
		switch v := arg[0].(type) {
		case bool:
			spawn, arg = v, arg[1:]
		}
	}

	if len(arg) > 0 {
		if m.Sessions == nil {
			m.Sessions = make(map[string]*Message)
		}

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
				switch v := arg[2].(type) {
				case string:
					root = Right(v)
				case bool:
					root = v
				}
			}

			method := "find"
			if len(arg) > 1 {
				switch v := arg[1].(type) {
				case string:
					method = v
				}
			}

			switch method {
			case "find":
				m.Sessions[key] = m.Find(value, root)
			case "search":
				m.Sessions[key] = m.Search(value, root)[0]
			}
			return m.Sessions[key]
		case nil:
			m.Sessions[key] = nil
			return nil
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
func (m *Message) Call(cb func(msg *Message) (sub *Message), arg ...interface{}) *Message {
	if m.callback = cb; len(arg) > 0 || len(m.Meta["detail"]) > 0 {
		m.Cmd(arg...)
	}
	return m
}
func (m *Message) Back(msg *Message) *Message {
	if msg == nil || m.callback == nil {
		return m
	}

	if msg.Hand {
		m.Log("cb", "%d %v %v", msg.code, msg.Meta["result"], msg.Meta["append"])
	} else {
		m.Log("cb", "%d %v %v", msg.code, msg.Meta["detail"], msg.Meta["option"])
	}

	if sub := m.callback(msg); sub != nil && m.message != nil && m.message != m {
		m.message.Back(sub)
	}

	return m
}
func (m *Message) CallBack(sync bool, cb func(msg *Message) (sub *Message), arg ...interface{}) *Message {
	if !sync {
		return m.Call(cb, arg...)
	}

	wait := make(chan *Message)
	go m.Call(func(sub *Message) *Message {
		msg := cb(sub)
		wait <- m
		return msg
	}, arg...)

	return <-wait
}

func (m *Message) Add(meta string, key string, value ...interface{}) *Message {
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if _, ok := m.Meta[meta]; !ok {
		m.Meta[meta] = make([]string, 0, 3)
	}

	switch meta {
	case "detail", "result":
		m.Meta[meta] = append(m.Meta[meta], key)
		m.Meta[meta] = append(m.Meta[meta], Trans(value...)...)

	case "option", "append":
		if _, ok := m.Meta[key]; !ok {
			m.Meta[key] = make([]string, 0, 3)
		}
		m.Meta[key] = append(m.Meta[key], Trans(value...)...)

		for _, v := range m.Meta[meta] {
			if v == key {
				return m
			}
		}
		m.Meta[meta] = append(m.Meta[meta], key)

	default:
		return m
		m.Assert(true, "%s 消息参数错误", meta)
	}

	return m
}
func (m *Message) Set(meta string, arg ...string) *Message {
	switch meta {
	case "detail", "result":
		if m != nil && m.Meta != nil {
			delete(m.Meta, meta)
		}
	case "option", "append":
		if len(arg) > 0 {
			delete(m.Meta, arg[0])
		} else {
			for _, k := range m.Meta[meta] {
				delete(m.Data, k)
				delete(m.Meta, k)
			}
			delete(m.Meta, meta)
		}
	default:
		return m
		m.Assert(true, "%s 消息参数错误", meta)
	}

	if len(arg) > 0 {
		m.Add(meta, arg[0], arg[1:])
	}
	return m
}
func (m *Message) Put(meta string, key string, value interface{}) *Message {
	switch meta {
	case "option", "append":
		if m.Set(meta, key); m.Data == nil {
			m.Data = make(map[string]interface{})
		}
		m.Data[key] = value

	default:
		return m
		m.Assert(true, "%s 消息参数错误", meta)
	}
	return m
}
func (m *Message) Has(key ...string) bool {
	switch len(key) {
	case 1:
		if _, ok := m.Data[key[0]]; ok {
			return true
		}
		if _, ok := m.Meta[key[0]]; ok {
			return true
		}
	}
	return false
}
func (m *Message) Get(key string) string {
	if meta, ok := m.Meta[key]; ok && len(meta) > 0 {
		return meta[0]
	}
	return ""
}
func (m *Message) Geti(key string) int {
	n, e := strconv.Atoi(m.Get(key))
	m.Assert(e)
	return n
}
func (m *Message) Gets(key string) bool {
	return Right(m.Get(key))
}
func (m *Message) Echo(str string, arg ...interface{}) *Message {
	if len(arg) > 0 {
		return m.Add("result", fmt.Sprintf(str, arg...))
	}
	return m.Add("result", str)
}
func (m *Message) Color(color int, str string, arg ...interface{}) *Message {
	if len(arg) > 0 {
		str = fmt.Sprintf(str, arg...)
	}
	if m.Options("terminal_color") {
		str = fmt.Sprintf("\033[%dm%s\033[0m", color, str)
	}
	return m.Add("result", str)
}
func (m *Message) Table(cbs ...func(maps map[string]string, list []string, line int) (goon bool)) *Message {
	var cb func(maps map[string]string, list []string, line int) (goon bool)
	if len(cbs) > 0 {
		cb = cbs[0]
	} else {
		row := m.Confx("table_row_sep")
		col := m.Confx("table_col_sep")
		compact := Right(m.Confx("table_compact"))
		cb = func(maps map[string]string, lists []string, line int) bool {
			for i, v := range lists {
				if k := m.Meta["append"][i]; compact {
					v = maps[k]
				}

				if m.Echo(v); i < len(lists)-1 {
					m.Echo(col)
				}
			}
			m.Echo(row)
			return true
		}
	}
	if len(m.Meta["append"]) == 0 {
		return m
	}

	//计算列宽
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

	//输出字段名
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
			//解析extra字段
			if m.Options("extras") && k == "extra" {
				var extra interface{}
				json.Unmarshal([]byte(data), &extra)
				for _, k := range m.Meta["extras"] {
					if i, e := strconv.Atoi(k); e == nil && i >= 0 {
						if d, ok := extra.([]interface{}); ok && i < len(d) {
							extra = d[i]
							continue
						}
					}

					if d, ok := extra.(map[string]interface{}); ok {
						extra = d[k]
						continue
					}

					extra = nil
					break
				}

				if extra == nil {
					data = ""
				} else {
					format := m.Confx("extra_format")
					if format == "" {
						format = "%v"
					}
					data = fmt.Sprintf(format, extra)
				}
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
func (m *Message) Matrix(index int, arg ...interface{}) string {
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
func (m *Message) Sort(key string, arg ...string) *Message {
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
			case "str":
				if table[i][key] > table[j][key] {
					result = true
				}
			case "str_r":
				if table[i][key] < table[j][key] {
					result = true
				}
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
			case "time":
				ti, e := time.ParseInLocation(m.Confx("time_format"), table[i][key], time.Local)
				m.Assert(e)
				tj, e := time.ParseInLocation(m.Confx("time_format"), table[j][key], time.Local)
				m.Assert(e)
				if tj.Before(ti) {
					result = true
				}
			case "time_r":
				ti, e := time.ParseInLocation(m.Confx("time_format"), table[i][key], time.Local)
				m.Assert(e)
				tj, e := time.ParseInLocation(m.Confx("time_format"), table[j][key], time.Local)
				m.Assert(e)
				if ti.Before(tj) {
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
	return m
}

func (m *Message) Insert(meta string, index int, arg ...interface{}) string {
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	if len(arg) == 0 {
		if -1 < index && index < len(m.Meta[meta]) {
			return m.Meta[meta][index]
		}
		return ""
	}

	str := Trans(arg...)
	if m.Confs("insert_limit") {
		index = (index+2)%(len(m.Meta[meta])+2) - 2
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
func (m *Message) Detail(arg ...interface{}) string {
	noset, index := true, 0
	if len(arg) > 0 {
		switch v := arg[0].(type) {
		case int:
			noset, index, arg = false, v, arg[1:]
		}
	}
	if noset && len(arg) > 0 {
		index = -2
	}

	return m.Insert("detail", index, arg...)
}
func (m *Message) Detaili(arg ...interface{}) int {
	i, e := strconv.Atoi(m.Detail(arg...))
	m.Assert(e)
	return i
}
func (m *Message) Details(arg ...interface{}) bool {
	return Right(m.Detail(arg...))
}
func (m *Message) Result(arg ...interface{}) string {
	noset, index := true, 0
	if len(arg) > 0 {
		switch v := arg[0].(type) {
		case int:
			noset, index, arg = false, v, arg[1:]
		}
	}
	if noset && len(arg) > 0 {
		index = -2
	}

	return m.Insert("result", index, arg...)
}
func (m *Message) Resulti(arg ...interface{}) int {
	i, e := strconv.Atoi(m.Result(arg...))
	m.Assert(e)
	return i
}
func (m *Message) Results(arg ...interface{}) bool {
	return Right(m.Result(arg...))
}
func (m *Message) Option(key string, arg ...interface{}) string {
	if len(arg) > 0 {
		m.Insert(key, 0, arg...)
		if _, ok := m.Meta[key]; ok {
			m.Add("option", key)
		}
	}

	for msg := m; msg != nil; msg = msg.message {
		if !msg.Has(key) {
			continue
		}
		for _, k := range msg.Meta["option"] {
			if k == key {
				return msg.Get(key)
			}
		}
	}
	return ""
}
func (m *Message) Optioni(key string, arg ...interface{}) int {
	i, e := strconv.Atoi(m.Option(key, arg...))
	m.Assert(e)
	return i
}
func (m *Message) Options(key string, arg ...interface{}) bool {
	return Right(m.Option(key, arg...))
}
func (m *Message) Optionv(key string, arg ...interface{}) interface{} {
	if len(arg) > 0 {
		m.Put("option", key, arg[0])
	}

	for msg := m; msg != nil; msg = msg.message {
		if msg.Meta == nil || !msg.Has(key) {
			continue
		}
		for _, k := range msg.Meta["option"] {
			if k == key {
				if v, ok := msg.Data[key]; ok {
					return v
				}
				return msg.Meta[key]
			}
		}
	}
	return nil
}
func (m *Message) Optionx(key string, format string) interface{} {
	if value := m.Option(key); value != "" {
		return fmt.Sprintf(format, value)
	}
	return ""
}
func (m *Message) Append(key string, arg ...interface{}) string {
	if len(arg) > 0 {
		m.Insert(key, 0, arg...)
		if _, ok := m.Meta[key]; ok {
			m.Add("append", key)
		}
	}

	ms := []*Message{m}
	for i := 0; i < len(ms); i++ {
		ms = append(ms, ms[i].messages...)
		if !ms[i].Has(key) {
			continue
		}
		for _, k := range ms[i].Meta["append"] {
			if k == key {
				return ms[i].Get(key)
			}
		}
	}
	return ""
}
func (m *Message) Appendi(key string, arg ...interface{}) int {
	i, e := strconv.Atoi(m.Append(key, arg...))
	m.Assert(e)
	return i
}
func (m *Message) Appends(key string, arg ...interface{}) bool {
	return Right(m.Append(key, arg...))
}
func (m *Message) Appendv(key string, arg ...interface{}) interface{} {
	if len(arg) > 0 {
		m.Put("append", key, arg[0])
	}

	ms := []*Message{m}
	for i := 0; i < len(ms); i++ {
		ms = append(ms, ms[i].messages...)
		if !ms[i].Has(key) {
			continue
		}
		for _, k := range ms[i].Meta["append"] {
			if k == key {
				if v, ok := ms[i].Data[key]; ok {
					return v
				}
				return ms[i].Meta[key]
			}
		}
	}
	return nil
}

func (m *Message) Parse(arg interface{}) string {
	switch str := arg.(type) {
	case string:
		if len(str) > 1 && str[0] == '$' {
			return m.Cap(str[1:])
		}
		if len(str) > 1 && str[0] == '@' {
			return m.Confx(str[1:])
		}
		return str
	}
	return ""
}
func (m *Message) Wait() bool {
	if m.target.exit != nil {
		return <-m.target.exit
	}
	return true
}
func (m *Message) Start(name string, help string, arg ...string) bool {
	return m.Set("detail", arg...).target.Spawn(m, name, help).Begin(m).Start(m)
}
func (m *Message) Cmd(args ...interface{}) *Message {
	if m == nil {
		return m
	}
	if len(args) > 0 {
		m.Set("detail", Trans(args...)...)
	}
	key, arg := m.Meta["detail"][0], m.Meta["detail"][1:]

	for _, c := range []*Context{m.target, m.source} {
		for s := c; s != nil; s = s.context {

			if x, ok := s.Commands[key]; ok && x.Hand != nil {
				m.TryCatch(m, true, func(m *Message) {
					m.Log("cmd", "%s:%s %v %v", s.Name, c.Name, m.Meta["detail"], m.Meta["option"])
					if args := []string{}; x.Form != nil {
						for i := 0; i < len(arg); i++ {
							n, ok := x.Form[arg[i]]
							if !ok {
								args = append(args, arg[i])
								continue
							}

							if n < 0 {
								n += len(arg) - i
							}

							m.Add("option", arg[i], arg[i+1:i+1+n])
							i += n
						}
						arg = args
					}

					m.Hand = true
					x.Hand(m, s, key, arg...)

				})
				return m
			}
		}
	}
	return m
}
func (m *Message) Confx(key string, arg ...interface{}) string {
	conf := m.Conf(key)
	if len(arg) == 0 {
		value := m.Option(key)
		if value == "" {
			value = conf
		}
		return value
	}

	value := ""
	switch v := arg[0].(type) {
	case map[string]interface{}:
		conf = v[key].(string)
		value = m.Option(key)
		arg = arg[1:]
	case string:
		value, arg = v, arg[1:]
	case []string:
		which := 0
		if len(arg) > 1 {
			if x, ok := arg[1].(int); ok {
				which, arg = x, arg[2:]
			}
		}
		if which < len(v) {
			value = v[which]
		}
	default:
		if x := fmt.Sprintf("%v", v); v != nil && x != "" {
			value, arg = x, arg[1:]
		}
	}

	force := false
	if len(arg) > 0 {
		if v, ok := arg[0].(bool); ok {
			force, arg = v, arg[1:]
		}
	}
	if !force && value == "" {
		value = conf
	}

	format := "%s"
	if len(arg) > 0 {
		if v, ok := arg[0].(string); ok {
			format, arg = v, arg[1:]
		}
	}
	if value != "" {
		args := []interface{}{value}
		for _, v := range arg {
			args = append(args, v)
		}
		value = fmt.Sprintf(format, args...)
	}
	return value
}
func (m *Message) Confs(key string, arg ...interface{}) bool {
	index, value := "", m.Conf(key)
	if len(arg) > 0 {
		switch v := arg[0].(type) {
		case string:
			index, arg, value = v, arg[1:], m.Conf(key, v)
		case []string:
			index = strings.Join(v, ".")
			arg, value = arg[1:], m.Conf(key, index)
		}
	}

	if len(arg) > 0 {
		val := "0"
		if t, ok := arg[0].(bool); ok && t {
			val = "1"
		}

		if index != "" {
			value = m.Conf(key, index, val)
		} else {
			value = m.Conf(key, val)
		}
	}

	return Right(value)
}
func (m *Message) Confi(key string, arg ...interface{}) int {
	index, value := "", m.Conf(key)
	if len(arg) > 0 {
		if i, ok := arg[0].(string); ok {
			arg, index, value = arg[1:], i, m.Conf(key, i)
		}
	}

	n, e := strconv.Atoi(value)
	m.Assert(e)

	if len(arg) > 0 {
		if index != "" {
			n, e = strconv.Atoi(m.Conf(key, index, fmt.Sprintf("%d", arg[0])))
		} else {
			n, e = strconv.Atoi(m.Conf(key, fmt.Sprintf("%d", arg[0])))
		}
		m.Assert(e)
	}

	return n
}
func (m *Message) Confv(key string, args ...interface{}) interface{} {
	var hand func(m *Message, x *Config, arg ...string) string
	arg := Trans(args...)

	for _, c := range []*Context{m.target, m.source} {
		for s := c; s != nil; s = s.context {
			if x, ok := s.Configs[key]; ok {
				if len(args) == 0 {
					return x.Value
				}
				if len(arg) == 3 {
					hand = x.Hand
				}

				switch x.Value.(type) {
				case string:
					x.Value = fmt.Sprintf("%v", arg[0])
				case bool:
					x.Value = Right(fmt.Sprintf("%v", args[0]))
				case int:
					i, e := strconv.Atoi(fmt.Sprintf("%v", args[0]))
					m.Assert(e)
					x.Value = i
				case nil:
					x.Value = args[0]
				default:
					for i := 0; i < len(args); i += 2 {
						if i < len(args)-1 {
							x.Value = Chain(m, x.Value, args[i], args[i+1])
						}
						if i == len(args)-2 {
							return Chain(m, x.Value, args[len(args)-2])
						}
						if i == len(args)-1 {
							return Chain(m, x.Value, args[len(args)-1])
						}
					}
				}
			}
		}
	}

	if len(args) == 0 {
		return nil
	}

	m.Log("conf", "%s %v", key, args)
	if m.target.Configs == nil {
		m.target.Configs = make(map[string]*Config)
	}
	if len(arg) == 3 {
		m.target.Configs[key] = &Config{Name: arg[0], Value: arg[1], Help: arg[2], Hand: hand}
		return m.Conf(key, arg[1])
	}
	if !m.Confs("auto_make") {
		return nil
	}

	if len(arg) == 1 {
		m.target.Configs[key] = &Config{Name: key, Value: arg[0], Help: "auto make", Hand: hand}
		return m.Conf(key, arg[0])
	}
	m.target.Configs[key] = &Config{Name: key, Value: Chain(m, nil, args), Help: "auto make", Hand: hand}
	return Chain(m, key, args[len(args)-2+(len(args)%2)])
}
func (m *Message) Conf(key string, args ...interface{}) string {
	var hand func(m *Message, x *Config, arg ...string) string

	for _, c := range []*Context{m.target, m.source} {
		for s := c; s != nil; s = s.context {
			if x, ok := s.Configs[key]; ok {
				switch value := x.Value.(type) {
				case string:
					val := ""
					if len(args) > 0 {
						switch v := args[0].(type) {
						case string:
							val = v
						case nil:
						default:
							val = fmt.Sprintf("%v", v)

						}
					}
					switch len(args) {
					case 0:
						if x.Hand != nil {
							return x.Hand(m, x)
						}
						return value
					case 1:
						if x.Hand != nil {
							x.Value = x.Hand(m, x, val)
						} else {
							x.Value = val
						}
						return value
					default:
						if hand == nil {
							hand = x.Hand
						}
					}
				case bool:
				case int:
				default:
					values := ""
					for i := 0; i < len(args); i += 2 {
						if i < len(args)-1 {
							x.Value = Chain(m, x.Value, args[i], args[i+1])
						}

						if val := Chain(m, x.Value, args[i]); val != nil {
							values = fmt.Sprintf("%v", val)
						}
					}

					if len(args) == 0 && x.Value != nil {
						values = fmt.Sprintf("%T", x.Value)
					}
					return values
				}
			}
		}
	}

	if len(args) > 0 {
		m.Log("conf", "%s %v", key, args)
		if m.target.Configs == nil {
			m.target.Configs = make(map[string]*Config)
		}

		arg := Trans(args...)
		if len(arg) == 3 {
			m.target.Configs[key] = &Config{Name: arg[0], Value: arg[1], Help: arg[2], Hand: hand}
			return m.Conf(key, arg[1])
		}
		if !m.Confs("auto_make") {
			return ""
		}
		if len(arg) == 1 {
			m.target.Configs[key] = &Config{Name: key, Value: arg[0], Help: "auto make", Hand: hand}
			return m.Conf(key, arg[0])
		}

		var value interface{}
		for i := 0; i < len(args)-1; i += 2 {
			value = Chain(m, value, args[i], args[i+1])
		}
		m.target.Configs[key] = &Config{Name: key, Value: value, Help: "auto make", Hand: hand}
		if val := Chain(m, key, args[len(args)-2]); val != nil {
			return fmt.Sprintf("%v", val)
		}
	}

	return ""
}
func (m *Message) Caps(key string, arg ...bool) bool {
	if len(arg) > 0 {
		if arg[0] {
			m.Cap(key, "1")
		} else {
			m.Cap(key, "0")
		}
	}

	return Right(m.Cap(key))
}
func (m *Message) Capi(key string, arg ...int) int {
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
func (m *Message) Cap(key string, arg ...string) string {
	var hand func(m *Message, x *Cache, arg ...string) string

	for _, c := range []*Context{m.target, m.source} {
		for s := c; s != nil; s = s.context {
			if x, ok := s.Caches[key]; ok {
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
				}
			}
		}
	}

	if len(arg) == 3 {
		if m.target.Caches == nil {
			m.target.Caches = make(map[string]*Cache)
		}

		m.target.Caches[key] = &Cache{Name: arg[0], Value: arg[1], Help: arg[2], Hand: hand}
		m.Log("cap", "%s %v", key, arg)
		return m.Cap(key, arg[1])
	}

	return ""
}

var CGI = template.FuncMap{
	"meta": func(arg ...interface{}) string {
		//meta meta [key [index]]
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
	},
	"sess": func(arg ...interface{}) string {
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return fmt.Sprintf("%v", m)
			}

			switch which := arg[1].(type) {
			case string:
				m.Sess(which, arg[2:]...)
				return ""
			}
		}
		return ""
	},

	"ctx": func(arg ...interface{}) string {
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
	},
	"msg": func(arg ...interface{}) interface{} {
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				return fmt.Sprintf("%v", m.Format())
			}

			switch which := arg[1].(type) {
			case string:
				switch which {
				case "spawn":
					return m.Spawn()
				case "code":
					return m.code
				case "time":
					return m.time.Format("2006-01-02 15:04:05")
				case "source":
					return m.source.Name
				case "target":
					return m.target.Name
				case "message":
					return m.message
				case "messages":
					return m.messages
				case "sessions":
					return m.Sessions
				default:
					return m.Sess(which)
				}
			case int:
				ms := []*Message{m}
				for i := 0; i < len(ms); i++ {
					if ms[i].code == which {
						return ms[i]
					}
					ms = append(ms, ms[i].messages...)
				}
			}
		}
		return ""
	},

	"cap": func(arg ...interface{}) string {
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
	},
	"conf": func(arg ...interface{}) interface{} {
		if len(arg) == 0 {
			return ""
		}

		if m, ok := arg[0].(*Message); ok {
			if len(arg) == 1 {
				list := []string{}
				for k, _ := range m.target.Configs {
					list = append(list, k)
				}
				return list
			}

			switch which := arg[1].(type) {
			case string:
				if len(arg) == 2 {
					return m.Confv(which)
				}
				return m.Confv(which, arg[2:]...)
			}
		}
		return ""
	},
	"cmd": func(arg ...interface{}) string {
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
	},

	"detail": func(arg ...interface{}) interface{} {
		if len(arg) == 0 {
			return ""
		}

		switch m := arg[0].(type) {
		case *Message:
			if len(arg) == 1 {
				return m.Meta["detail"]
			}

			index := 0
			switch value := arg[1].(type) {
			case int:
				index = value
			case string:
				i, e := strconv.Atoi(value)
				m.Assert(e)
				index = i
			}

			if len(arg) == 2 {
				return m.Detail(index)
			}

			return m.Detail(index, arg[2])
		case map[string][]string:
			return strings.Join(m["detail"], "")
		case []string:
			return strings.Join(m, "")
		default:
			return m
		}
		return ""
	},
	"option": func(arg ...interface{}) interface{} {
		if len(arg) == 0 {
			return ""
		}

		switch m := arg[0].(type) {
		case *Message:
			if len(arg) == 1 {
				return m.Meta["option"]
			}

			switch value := arg[1].(type) {
			case int:
				if 0 <= value && value < len(m.Meta["option"]) {
					return m.Meta["option"][value]
				}
			case string:
				if len(arg) == 2 {
					return m.Optionv(value)
				}

				switch val := arg[2].(type) {
				case int:
					if 0 <= val && val < len(m.Meta[value]) {
						return m.Meta[value][val]
					}
				}
			}
		case map[string][]string:
			if len(arg) == 1 {
				return strings.Join(m["option"], "")
			}
			switch value := arg[1].(type) {
			case string:
				return strings.Join(m[value], "")
			}
		case []string:
			return strings.Join(m, "")
		default:
			return m
		}
		return ""
	},
	"result": func(arg ...interface{}) interface{} {
		if len(arg) == 0 {
			return ""
		}

		switch m := arg[0].(type) {
		case *Message:
			if len(arg) == 1 {
				return m.Meta["result"]
			}

			index := 0
			switch value := arg[1].(type) {
			case int:
				index = value
			case string:
				i, e := strconv.Atoi(value)
				m.Assert(e)
				index = i
			}

			if len(arg) == 2 {
				return m.Result(index)
			}

			return m.Result(index, arg[2])
		case map[string][]string:
			return strings.Join(m["result"], "")
		case []string:
			return strings.Join(m, "")
		default:
			return m
		}
		return ""
	},
	"append": func(arg ...interface{}) interface{} {
		if len(arg) == 0 {
			return ""
		}

		switch m := arg[0].(type) {
		case *Message:
			if len(arg) == 1 {
				return m.Meta["append"]
			}

			switch value := arg[1].(type) {
			case int:
				if 0 <= value && value < len(m.Meta["append"]) {
					return m.Meta["append"][value]
				}
			case string:
				if len(arg) == 2 {
					return m.Meta[value]
				}

				switch val := arg[2].(type) {
				case int:
					if 0 <= val && val < len(m.Meta[value]) {
						return m.Meta[value][val]
					}
				}
			}
		case map[string][]string:
			if len(arg) == 1 {
				return strings.Join(m["append"], "")
			}
			switch value := arg[1].(type) {
			case string:
				return strings.Join(m[value], "")
			}
		case []string:
			return strings.Join(m, "")
		default:
			return m
		}
		return ""
	},
	"table": func(arg ...interface{}) []interface{} {
		if len(arg) == 0 {
			return []interface{}{}
		}

		switch m := arg[0].(type) {
		case *Message:
			if len(m.Meta["append"]) == 0 {
				return []interface{}{}
			}
			if len(arg) == 1 {
				data := []interface{}{}
				nrow := len(m.Meta[m.Meta["append"][0]])
				for i := 0; i < nrow; i++ {
					line := map[string]string{}
					for _, k := range m.Meta["append"] {
						line[k] = m.Meta[k][i]
						if len(m.Meta[k]) != i {
							continue
						}
					}
					data = append(data, line)
				}

				return data
			}
		case map[string][]string:
			if len(arg) == 1 {
				data := []interface{}{}
				nrow := len(m[m["append"][0]])

				for i := 0; i < nrow; i++ {
					line := map[string]string{}
					for _, k := range m["append"] {
						line[k] = m[k][i]
					}
					data = append(data, line)
				}

				return data
			}
		}
		return []interface{}{}
	},
	"unescape": func(str string) interface{} {
		return template.HTML(str)
	},

	"list": func(arg interface{}) interface{} {
		n := 0
		switch v := arg.(type) {
		case string:
			i, e := strconv.Atoi(v)
			if e == nil {
				n = i
			}
		case int:
			n = v
		}

		list := make([]int, n)
		for i := 1; i <= n; i++ {
			list[i-1] = i
		}
		return list
	},
}
var Pulse = &Message{code: 0, time: time.Now(), source: Index, target: Index, Meta: map[string][]string{}}
var Index = &Context{Name: "ctx", Help: "模块中心",
	Caches: map[string]*Cache{
		"nserver":  &Cache{Name: "nserver", Value: "0", Help: "服务数量"},
		"ncontext": &Cache{Name: "ncontext", Value: "0", Help: "模块数量"},
		"nmessage": &Cache{Name: "nmessage", Value: "0", Help: "消息数量"},
	},
	Configs: map[string]*Config{
		"chain":       &Config{Name: "chain", Value: map[string]interface{}{}, Help: "调试模式，on:打印，off:不打印)"},
		"compact_log": &Config{Name: "compact_log(true/false)", Value: "true", Help: "调试模式，on:打印，off:不打印)"},
		"auto_make":   &Config{Name: "auto_make(true/false)", Value: "true", Help: "调试模式，on:打印，off:不打印)"},
		"debug":       &Config{Name: "debug(on/off)", Value: "on", Help: "调试模式，on:打印，off:不打印)"},

		"search_method": &Config{Name: "search_method(find/search)", Value: "search", Help: "搜索方法, find: 模块名精确匹配, search: 模块名或帮助信息模糊匹配"},
		"search_choice": &Config{Name: "search_choice(first/last/rand/magic)", Value: "magic", Help: "搜索匹配, first: 匹配第一个模块, last: 匹配最后一个模块, rand: 随机选择, magic: 加权选择"},
		"search_action": &Config{Name: "search_action(list/switch)", Value: "switch", Help: "搜索操作, list: 输出模块列表, switch: 模块切换"},
		"search_root":   &Config{Name: "search_root(true/false)", Value: "true", Help: "搜索起点, true: 根模块, false: 当前模块"},

		"insert_limit": &Config{Name: "insert_limit(true/false)", Value: "true", Help: "参数的索引"},
		"detail_index": &Config{Name: "detail_index", Value: "0", Help: "参数的索引"},
		"result_index": &Config{Name: "result_index", Value: "-2", Help: "返回值的索引"},

		"list_help":     &Config{Name: "list_help", Value: "list command", Help: "命令列表帮助"},
		"table_compact": &Config{Name: "table_compact", Value: "false", Help: "命令列表帮助"},
		"table_col_sep": &Config{Name: "table_col_sep", Value: "\t", Help: "命令列表帮助"},
		"table_row_sep": &Config{Name: "table_row_sep", Value: "\n", Help: "命令列表帮助"},

		"page_offset": &Config{Name: "page_offset", Value: "0", Help: "列表偏移"},
		"page_limit":  &Config{Name: "page_limit", Value: "100", Help: "列表大小"},

		"time_format": &Config{Name: "time_format", Value: "2006-01-02 15:04:05", Help: "时间格式"},
	},
	Commands: map[string]*Command{
		"help": &Command{Name: "help topic", Help: "帮助", Hand: func(m *Message, c *Context, key string, arg ...string) {
			if len(arg) == 0 {
				m.Echo("usage: help context [module [command|config|cache name]]\n")
				m.Echo("     : 查看模块信息, module: 模块名, command: 模块命令, config: 模块配置, cache: 模块缓存, name: 模块参数\n")
				m.Echo("usage: help command [name]\n")
				m.Echo("     : 查看当前环境下命令, name: 命令名\n")
				m.Echo("usage: help config [name]\n")
				m.Echo("     : 查看当前环境下配置, name: 配置名\n")
				m.Echo("usage: help cache [name]\n")
				m.Echo("     : 查看当前环境下缓存, name: 缓存名\n")
				m.Echo("\n")

				m.Echo("^_^  Welcome to context world  ^_^\n")
				m.Echo("Version: 1.0 A New Language, A New Framework\n")
				m.Echo("More: github.com/shylinux/context/README.md\n")
				m.Color(31, "       c\n")
				m.Color(31, "     sh").Color(33, " go\n")
				m.Color(31, "   vi").Color(32, " php").Color(32, " js\n")
				m.Echo(" ARM Linux HTTP\n")
				m.Color(31, "Context ").Color(32, "Message\n")
				m.Color(32, "ctx ").Color(33, "cli ").Color(31, "aaa ").Color(33, "web\n")
				m.Color(32, "lex ").Color(33, "yac ").Color(31, "log ").Color(33, "gdb\n")
				m.Color(32, "tcp ").Color(33, "nfs ").Color(31, "ssh ").Color(33, "mdb\n")
				m.Color(31, "script ").Color(32, "template\n")
				return
			}

			switch arg[0] {
			case "context":
				switch len(arg) {
				case 1:
					keys := []string{}
					values := map[string]*Context{}
					m.Travel(func(m *Message, i int) bool {
						if _, ok := values[m.Cap("module")]; !ok {
							keys = append(keys, m.Cap("module"))
							values[m.Cap("module")] = m.Target()
						}
						return true
					}, m.Target().root)
					sort.Strings(keys)
					for _, k := range keys {
						m.Echo("%s: %s %s\n", k, values[k].Name, values[k].Help)
					}
					break
				case 2:
					if msg := m.Find(arg[1]); msg != nil {
						m.Echo("%s: %s %s\n", arg[1], msg.Target().Name, msg.Target().Help)
						m.Echo("commands:\n")
						for k, v := range msg.Target().Commands {
							m.Echo("  %s: %s\n", k, v.Name)
						}
						m.Echo("configs:\n")
						for k, v := range msg.Target().Configs {
							m.Echo("  %s: %s\n", k, v.Name)
						}
						m.Echo("caches:\n")
						for k, v := range msg.Target().Caches {
							m.Echo("  %s: %s\n", k, v.Name)
						}
					}
				default:
					if msg := m.Find(arg[1]); msg != nil {
						m.Echo("%s: %s %s\n", arg[1], msg.Target().Name, msg.Target().Help)
						switch arg[2] {
						case "command":
							for k, v := range msg.Target().Commands {
								if k == arg[3] {
									m.Echo("%s: %s\n%s\n", k, v.Name, v.Help)
								}
							}
						case "config":
							for k, v := range msg.Target().Configs {
								if k == arg[3] {
									m.Echo("%s: %s\n  %s\n", k, v.Name, v.Help)
								}
							}
						case "cache":
							for k, v := range msg.Target().Caches {
								if k == arg[3] {
									m.Echo("%s: %s\n  %s\n", k, v.Name, v.Help)
								}
							}
						}
					}
				}
			case "command":
				keys := []string{}
				values := map[string]*Command{}
				for s := m.Target(); s != nil; s = s.context {
					for k, v := range s.Commands {
						if _, ok := values[k]; ok {
							continue
						}
						if len(arg) > 1 && k == arg[1] {
							m.Echo("%s: %s\n%s\n", k, v.Name, v.Help)
							for k, v := range v.Form {
								m.Echo("  option: %s(%d)\n", k, v)
							}
							return
						}
						keys = append(keys, k)
						values[k] = v
					}
				}
				sort.Strings(keys)
				for _, k := range keys {
					m.Echo("%s: %s\n", k, values[k].Name)
				}
			case "config":
				keys := []string{}
				values := map[string]*Config{}
				for s := m.Target(); s != nil; s = s.context {
					for k, v := range s.Configs {
						if _, ok := values[k]; ok {
							continue
						}
						if len(arg) > 1 && k == arg[1] {
							m.Echo("%s(%s): %s %s\n", k, v.Value, v.Name, v.Help)
							return
						}
						keys = append(keys, k)
						values[k] = v
					}
				}
				sort.Strings(keys)
				for _, k := range keys {
					m.Echo("%s(%s): %s\n", k, values[k].Value, values[k].Name)
				}
			case "cache":
				keys := []string{}
				values := map[string]*Cache{}
				for s := m.Target(); s != nil; s = s.context {
					for k, v := range s.Caches {
						if _, ok := values[k]; ok {
							continue
						}
						if len(arg) > 1 && k == arg[1] {
							m.Echo("%s(%s): %s %s\n", k, v.Value, v.Name, v.Help)
							return
						}
						keys = append(keys, k)
						values[k] = v
					}
				}
				sort.Strings(keys)
				for _, k := range keys {
					m.Echo("%s(%s): %s\n", k, values[k].Value, values[k].Name)
				}
			}

		}},

		"message": &Command{Name: "message [code] [cmd...]", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m
			if ms := m.Find(m.Cap("ps_target")); ms != nil {
				msg = ms
			}

			if len(arg) > 0 {
				if code, e := strconv.Atoi(arg[0]); e == nil {
					if msg = m.root.Tree(code); msg != nil {
						arg = arg[1:]
					}
				}
			}

			if len(arg) > 0 && arg[0] == "spawn" {
				sub := msg.Spawn()
				m.Echo("%d", sub.code)
				return
			}

			if len(arg) > 0 {
				msg = msg.Spawn().Cmd(arg)
				m.Copy(msg, "append").Copy(msg, "result")
				return
			}

			if msg.message != nil {
				m.Add("append", "time", msg.message.time.Format("15:04:05"))
				m.Add("append", "code", msg.message.code)
				m.Add("append", "source", msg.message.source.Name)
				m.Add("append", "target", msg.message.target.Name)
				if msg.message.Meta != nil {
					m.Add("append", "details", fmt.Sprintf("%v", msg.message.Meta["detail"]))
					m.Add("append", "options", fmt.Sprintf("%v", msg.message.Meta["option"]))
				} else {
					m.Add("append", "details", "")
					m.Add("append", "options", "")
				}
			} else {
				m.Add("append", "time", "")
				m.Add("append", "code", "")
				m.Add("append", "source", "")
				m.Add("append", "target", "")
				m.Add("append", "details", "")
				m.Add("append", "options", "")
			}
			m.Add("append", "time", msg.time.Format("15:04:05"))
			m.Add("append", "code", msg.code)
			m.Add("append", "source", msg.source.Name)
			m.Add("append", "target", msg.target.Name)
			m.Add("append", "details", fmt.Sprintf("%v", msg.Meta["detail"]))
			m.Add("append", "options", fmt.Sprintf("%v", msg.Meta["option"]))
			for _, v := range msg.messages {
				m.Add("append", "time", v.time.Format("15:04:05"))
				m.Add("append", "code", v.code)
				m.Add("append", "source", v.source.Name)
				m.Add("append", "target", v.target.Name)
				m.Add("append", "details", fmt.Sprintf("%v", v.Meta["detail"]))
				m.Add("append", "options", fmt.Sprintf("%v", v.Meta["option"]))
			}
			m.Table()
		}},
		"detail": &Command{Name: "detail [index] [value...]", Help: "查看或添加参数", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.message
			if len(arg) == 0 {
				for i, v := range msg.Meta["detail"] {
					m.Add("append", "index", i)
					m.Add("append", "value", v)
				}
				m.Table()
				return
			}

			index := m.Confi("detail_index")
			if i, e := strconv.Atoi(arg[0]); e == nil {
				index, arg = i, arg[1:]
			}
			m.Echo("%s", msg.Detail(index, arg))
		}},
		"option": &Command{Name: "option [all] [key [index] [value...]]", Help: "查看或添加选项", Hand: func(m *Message, c *Context, key string, arg ...string) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			index := -100
			if len(arg) > 1 {
				if i, e := strconv.Atoi(arg[1]); e == nil {
					index = i
					for i := 1; i < len(arg)-1; i++ {
						arg[i] = arg[i+1]
					}
					arg = arg[:len(arg)-1]
				}
			}

			msg := m.message
			for msg = msg; msg != nil; msg = msg.message {
				for _, k := range msg.Meta["option"] {
					if len(arg) == 0 {
						m.Add("append", "key", k)
						m.Add("append", "value", fmt.Sprintf("%v", msg.Meta[k]))
						continue
					}

					if k != arg[0] {
						continue
					}

					if len(arg) > 1 {
						msg.Meta[k] = Array(m, msg.Meta[k], index, arg[1:])
						m.Echo("%v", msg.Meta[k])
						return
					}

					if index != -100 {
						m.Echo(Array(m, msg.Meta[k], index)[0])
						return
					}

					for i, v := range msg.Meta[k] {
						m.Add("append", "index", i)
						m.Add("append", "value", v)
					}
					m.Table()
					return
				}

				if !all {
					break
				}
			}
			m.Table()
		}},
		"result": &Command{Name: "result [index] [value...]", Help: "查看或添加返回值", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.message
			if len(arg) == 0 {
				for i, v := range msg.Meta["result"] {
					m.Add("append", "index", i)
					m.Add("append", "value", strings.Replace(v, "\n", "\\n", -1))
				}
				m.Table()
				return
			}

			index := m.Confi("result_index")
			if i, e := strconv.Atoi(arg[0]); e == nil {
				index, arg = i, arg[1:]
			}
			m.Echo("%s", msg.Result(index, arg))
		}},
		"append": &Command{Name: "append [all] [key [index] [value...]]", Help: "查看或添加附加值", Hand: func(m *Message, c *Context, key string, arg ...string) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			index := -100
			if len(arg) > 1 {
				if i, e := strconv.Atoi(arg[1]); e == nil {
					index = i
					for i := 1; i < len(arg)-1; i++ {
						arg[i] = arg[i+1]
					}
					arg = arg[:len(arg)-1]
				}
			}

			msg := m.message
			for msg = msg; msg != nil; msg = msg.message {
				for _, k := range msg.Meta["append"] {
					if len(arg) == 0 {
						m.Add("append", "key", k)
						m.Add("append", "value", fmt.Sprintf("%v", msg.Meta[k]))
						continue
					}

					if k != arg[0] {
						continue
					}

					if len(arg) > 1 {
						msg.Meta[k] = Array(m, msg.Meta[k], index, arg[1:])
						m.Echo("%v", msg.Meta[k])
						return
					}

					if index != -100 {
						m.Echo(Array(m, msg.Meta[k], index)[0])
						return
					}

					for i, v := range msg.Meta[k] {
						m.Add("append", "index", i)
						m.Add("append", "value", v)
					}
					m.Table()
					return
				}

				if !all {
					break
				}
			}
			m.Table()
		}},
		"session": &Command{Name: "session [all] [key [module]]", Help: "查看或添加会话", Hand: func(m *Message, c *Context, key string, arg ...string) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			msg := m.message
			for msg = msg; msg != nil; msg = msg.message {
				for k, v := range msg.Sessions {
					if len(arg) > 1 {
						msg.Sessions[arg[0]] = msg.Find(arg[1])
						return
					} else if len(arg) > 0 {
						if k == arg[0] {
							m.Echo("%d", v.code)
							return
						}
						continue
					}

					m.Add("append", "key", k)
					m.Add("append", "time", v.time.Format("15:04:05"))
					m.Add("append", "code", v.code)
					m.Add("append", "source", v.source.Name)
					m.Add("append", "target", v.target.Name)
					m.Add("append", "details", fmt.Sprintf("%v", v.Meta["detail"]))
					m.Add("append", "options", fmt.Sprintf("%v", v.Meta["option"]))
				}

				if len(arg) == 0 && !all {
					break
				}
			}
			m.Table()
		}},
		"callback": &Command{Name: "callback", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.message
			for msg := msg; msg != nil; msg = msg.message {
				m.Add("append", "msg", msg.code)
				m.Add("append", "fun", msg.callback)
			}
			m.Table()
		}},

		"context": &Command{Name: "context [find|search] [root|back|home] [first|last|rand|magic] [module] [cmd|switch|list|spawn|start|close]",
			Help: "查找并操作模块;\n查找方法, find: 精确查找, search: 模糊搜索;\n查找起点, root: 根模块, back: 父模块, home: 本模块;\n过滤结果, first: 取第一个, last: 取最后一个, rand: 随机选择, magic: 智能选择;\n操作方法, cmd: 执行命令, switch: 切换为当前, list: 查看所有子模块, spwan: 创建子模块并初始化, start: 启动模块, close: 结束模块",
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				if len(arg) == 1 && arg[0] == "~" && m.target.context != nil {
					m.target = m.target.context
					return
				}

				action := "switch"
				if len(arg) == 0 {
					action = "list"
				}

				method := "search"
				if len(arg) > 0 {
					switch arg[0] {
					case "find", "search":
						method, arg = arg[0], arg[1:]
					}
				}

				root := true
				if len(arg) > 0 {
					switch arg[0] {
					case "root":
						root, arg = true, arg[1:]
					case "home":
						root, arg = false, arg[1:]
					case "back":
						root, arg = false, arg[1:]
						if m.target.context != nil {
							m.target = m.target.context
						}
					}
				}

				ms := []*Message{}
				if len(arg) > 0 {
					switch method {
					case "find":
						if msg := m.Find(arg[0], root); msg != nil {
							ms, arg = append(ms, msg), arg[1:]
						}
					case "search":
						msg := m.Search(arg[0], root)
						if len(msg) > 1 || msg[0] != nil {
							if len(arg) > 1 {
								switch arg[1] {
								case "first":
									ms, arg = append(ms, msg[0]), arg[2:]
								case "last":
									ms, arg = append(ms, msg[len(msg)-1]), arg[2:]
								case "rand":
									ms, arg = append(ms, msg[rand.Intn(len(msg))]), arg[2:]
								case "magic":
									ms, arg = append(ms, msg...), arg[2:]
								default:
									ms, arg = append(ms, msg[0]), arg[1:]
								}
							} else {
								ms, arg = append(ms, msg[0]), arg[1:]
							}
						}

					}
				}

				if len(ms) == 0 {
					ms = append(ms, m)
				}

				if len(arg) > 0 {
					switch arg[0] {
					case "switch", "list", "spawn", "start", "close":
						action, arg = arg[0], arg[1:]
					default:
						action = "cmd"
					}
				}

				for _, msg := range ms {
					if msg == nil {
						continue
					}

					switch action {
					case "cmd":
						msg.Cmd(arg)
						if !msg.Hand {
							msg = msg.Sess("cli").Cmd("cmd", arg)
						}
						m.Copy(msg, "append").Copy(msg, "result")
					case "switch":
						m.target = msg.target
					case "list":
						msg.Travel(func(msg *Message, n int) bool {
							m.Add("append", "name", msg.target.Name)
							if msg.target.context != nil {
								m.Add("append", "ctx", msg.target.context.Name)
							} else {
								m.Add("append", "ctx", "")
							}
							m.Add("append", "msg", msg.target.message.code)
							m.Add("append", "status", msg.Cap("status"))
							m.Add("append", "stream", msg.Cap("stream"))
							m.Add("append", "help", msg.target.Help)
							return true
						})
					case "spawn":
						msg.target.Spawn(msg, arg[0], arg[1]).Begin(msg, arg[2:]...)
						m.Copy(msg, "append").Copy(msg, "result").Copy(msg, "target")
					case "start":
						msg.target.Start(msg, arg...)
						m.Copy(msg, "append").Copy(msg, "result").Copy(msg, "target")
					case "close":
						msg.target.Close(msg, arg...)
					}
				}

				if action == "list" {
					m.Table()
				}
			}},
		"command": &Command{Name: "command [all] [show]|[list [begin [end]] [prefix] test [key val]...]|[add [list_name name] [list_help help] cmd...]|[delete cmd]",
			Help: "查看或修改命令, show: 查看命令;\nlist: 查看列表命令, begin: 起始索引, end: 截止索引, prefix: 过滤前缀, test: 执行命令;\nadd: 添加命令, list_name: 命令别名, list_help: 命令帮助;\ndelete: 删除命令",
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				all := false
				if len(arg) > 0 && arg[0] == "all" {
					all, arg = true, arg[1:]
				}

				action := "show"
				if len(arg) > 0 {
					switch arg[0] {
					case "show", "list", "add", "delete":
						action, arg = arg[0], arg[1:]
					}
				}

				switch action {
				case "show":
					m.BackTrace(func(m *Message) bool {
						for k, v := range m.target.Commands {
							if len(arg) > 0 {
								if k == arg[0] {
									m.Add("append", "key", k)
									m.Add("append", "name", v.Name)
									m.Add("append", "help", v.Name)
								}
							} else {
								m.Add("append", "key", k)
								m.Add("append", "name", v.Name)
							}
						}

						return all
					})
					m.Table()
				case "list":
					if m.Cap("list_count") == "" {
						break
					}
					begin, end := 0, m.Capi("list_count")
					if len(arg) > 0 {
						if n, e := strconv.Atoi(arg[0]); e == nil {
							begin, arg = n, arg[1:]
						}
					}
					if len(arg) > 0 {
						if n, e := strconv.Atoi(arg[0]); e == nil {
							end, arg = n, arg[1:]
						}
					}
					prefix := ""
					if len(arg) > 0 && arg[0] != "test" {
						prefix, arg = arg[0], arg[1:]
					}

					test := false
					if len(arg) > 0 && arg[0] == "test" {
						test, arg = true, arg[1:]
						for i := 0; i < len(arg)-1; i += 2 {
							m.Add("option", arg[i], arg[i+1])
						}
					}

					for i := begin; i < end; i++ {
						index := fmt.Sprintf("%d", i)
						if c, ok := m.target.Commands[index]; ok {
							if prefix != "" && !strings.HasPrefix(c.Help.(string), prefix) {
								continue
							}

							if test {
								msg := m.Spawn().Cmd(index)
								m.Add("append", "index", i)
								m.Add("append", "help", c.Help)
								m.Add("append", "msg", msg.messages[0].code)
								m.Add("append", "res", msg.Result(0))
							} else {
								m.Add("append", "index", i)
								m.Add("append", "help", fmt.Sprintf("%s", c.Help))
								m.Add("append", "command", fmt.Sprintf("%s", strings.Replace(c.Name, "\n", "\\n", -1)))
							}
						}
					}
					m.Table()
				case "add":
					if m.target.Caches == nil {
						m.target.Caches = map[string]*Cache{}
					}
					if _, ok := m.target.Caches["list_count"]; !ok {
						m.target.Caches["list_count"] = &Cache{Name: "list_count", Value: "0", Help: "list_count"}
					}
					if m.target.Commands == nil {
						m.target.Commands = map[string]*Command{}
					}

					list_name, list_help := "", "list_cmd"
					if len(arg) > 1 && arg[0] == "list_name" {
						list_name, arg = arg[1], arg[2:]
					}
					if len(arg) > 1 && arg[0] == "list_help" {
						list_help, arg = arg[1], arg[2:]
					}

					m.target.Commands[m.Cap("list_count")] = &Command{Name: strings.Join(arg, " "), Help: list_help, Hand: func(cmd *Message, c *Context, key string, args ...string) {
						list := []string{}
						for _, v := range arg {
							if v == "__" {
								if len(args) > 0 {
									v, args = args[0], args[1:]
								} else {
									continue
								}
							} else if strings.HasPrefix(v, "_") {
								if len(args) > 0 {
									v, args = args[0], args[1:]
								} else if len(v) > 1 {
									v = v[1:]
								} else {
									v = "''"
								}
							}
							list = append(list, v)
						}
						list = append(list, args...)

						msg := cmd.Sess("cli").Cmd("source", strings.Join(list, " "))
						cmd.Copy(msg, "append").Copy(msg, "result").Copy(msg, "target")
					}}

					if list_name != "" {
						m.target.Commands[list_name] = m.target.Commands[m.Cap("list_count")]
					}
					m.Capi("list_count", 1)
				case "delete":
					m.BackTrace(func(m *Message) bool {
						delete(m.target.Commands, arg[0])
						return all
					})
				}
			}},
		"config": &Command{Name: "config [all] [export key..] [save|load file key...] [create map|list|string key name help] [delete key]",
			Help: "配置管理, export: 导出配置, save: 保存配置到文件, load: 从文件加载配置, create: 创建配置, delete: 删除配置",
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				all := false
				if len(arg) > 0 && arg[0] == "all" {
					arg, all = arg[1:], true
				}

				action, which := "", "-1"
				have := map[string]bool{}
				if len(arg) > 0 {
					switch arg[0] {
					case "export":
						action, arg = arg[0], arg[1:]
						for _, v := range arg {
							have[v] = true
						}
					case "save", "load":
						action, which, arg = arg[0], arg[1], arg[2:]
						for _, v := range arg {
							have[v] = true
						}
					case "create", "delete":
						action, arg = arg[0], arg[1:]
					}
				}

				if len(arg) == 0 || action != "" {
					save := map[string]interface{}{}
					if action == "load" {
						f, e := os.Open(m.Sess("nfs").Cmd("path", which).Result(0))
						if e != nil {
							return
						}
						defer f.Close()

						de := json.NewDecoder(f)
						if e = de.Decode(&save); e != nil {
							m.Log("info", "e: %v", e)
						}
					}

					m.BackTrace(func(m *Message) bool {
						for k, v := range m.target.Configs {
							switch action {
							case "export", "save":
								if len(have) == 0 || have[k] {
									save[k] = v.Value
								}
							case "load":
								if x, ok := save[k]; ok && (len(have) == 0 || have[k]) {
									v.Value = x
								}
							case "create":
								m.Assert(k != arg[1], "%s exists", arg[1])
							case "delete":
								if k == arg[0] {
									delete(m.target.Configs, k)
								}
								fallthrough
							default:
								m.Add("append", "key", k)
								m.Add("append", "value", strings.Replace(strings.Replace(m.Conf(k), "\n", "\\n", -1), "\t", "\\t", -1))
								m.Add("append", "name", v.Name)
							}
						}
						switch action {
						case "create":
							var value interface{}
							switch arg[0] {
							case "map":
								value = map[string]interface{}{}
							case "list":
								value = []interface{}{}
							default:
								value = ""
							}
							m.target.Configs[arg[1]] = &Config{Name: arg[2], Value: value, Help: arg[3]}
						}
						return all
					})
					m.Sort("key", "str").Table()

					switch action {
					case "save":
						// f, e := os.Create(which)
						// m.Assert(e)
						// defer f.Close()
						//
						buf, e := json.MarshalIndent(save, "", "  ")
						m.Assert(e)
						m.Sess("nfs").Add("option", "data", string(buf)).Cmd("save", which)

						// f.Write(buf)
					case "export":
						buf, e := json.MarshalIndent(save, "", "  ")
						m.Assert(e)
						m.Echo("%s", string(buf))
					}
					return
				}

				var value interface{}
				if len(arg) > 2 {
					value = m.Confv(arg[0], arg[1], arg[2])
				} else if len(arg) > 1 {
					value = m.Confv(arg[0], arg[1])
				} else {
					value = m.Confv(arg[0])
				}

				switch val := value.(type) {
				case map[string]interface{}:
					for k, v := range val {
						m.Add("append", k, v)
					}
					sort.Strings(m.Meta["append"])
					m.Table()
					// for k, v := range val {
					// 	m.Add("append", "key", k)
					// 	m.Add("append", "value", v)
					// }
					// m.Sort("key", "str").Table()
				case map[string]string:
					for k, v := range val {
						m.Add("append", k, v)
					}
					sort.Strings(m.Meta["append"])
					m.Table()
					// for k, v := range val {
					// 	m.Add("append", "key", k)
					// 	m.Add("append", "value", v)
					// }
					// m.Sort("key", "str").Table()
				case []interface{}:
					for i, v := range val {
						switch value := v.(type) {
						case map[string]interface{}:
							for k, v := range value {
								m.Add("append", k, v)
							}
							sort.Strings(m.Meta["append"])
						case map[string]string:
							for k, v := range value {
								m.Add("append", k, v)
							}
							sort.Strings(m.Meta["append"])
						default:
							m.Add("append", "index", i)
							m.Add("append", "value", v)
						}
					}
					m.Table()
				case []string:
					for i, v := range val {
						m.Add("append", "index", i)
						m.Add("append", "value", v)
					}
					m.Table()
				default:
					m.Echo("%v", value)
				}
			}},
		"cache": &Command{Name: "cache [all] |key [value]|key = value|key name value help|delete key]",
			Help: "查看、读写、赋值、新建、删除缓存变量",
			Hand: func(m *Message, c *Context, key string, arg ...string) {
				all := false
				if len(arg) > 0 && arg[0] == "all" {
					arg, all = arg[1:], true
				}

				switch len(arg) {
				case 0:
					m.BackTrace(func(m *Message) bool {
						for k, v := range m.target.Caches {
							m.Add("append", "key", k)
							m.Add("append", "value", m.Cap(k))
							m.Add("append", "name", v.Name)
						}
						return all
					})
					m.Sort("key", "str").Table()
					return
				case 2:
					if arg[0] == "delete" {
						delete(m.target.Caches, arg[1])
						return
					}
					m.Cap(arg[0], arg[1])
				case 3:
					m.Cap(arg[0], arg[0], arg[2], arg[0])
				default:
					m.Echo(m.Cap(arg[0], arg[1:]...))
					return
				}
			}},

		"trans": &Command{Name: "trans key index", Help: "数据转换", Hand: func(m *Message, c *Context, key string, arg ...string) {
			value := m.Data[arg[0]]
			if arg[1] != "" {
				v := Chain(m, value, arg[1])
				value = v
			}

			switch val := value.(type) {
			case map[string]interface{}:
				for k, v := range val {
					m.Add("append", "key", k)
					switch value := v.(type) {
					case float64:
						m.Add("append", "value", fmt.Sprintf("%d", int(value)))
					default:
						m.Add("append", "value", fmt.Sprintf("%v", value))
					}
				}
				m.Sort("key", "str").Table()
			case map[string]string:
				for k, v := range val {
					m.Add("append", "key", k)
					m.Add("append", "value", v)
				}
				m.Sort("key", "str").Table()
			case []interface{}:
				for i, v := range val {
					switch value := v.(type) {
					case map[string]interface{}:
						for k, v := range value {
							m.Add("append", k, fmt.Sprintf("%v", v))
						}
					default:
						m.Add("append", "index", i)
						m.Add("append", "value", fmt.Sprintf("%v", v))
					}
				}
				m.Table()
			case []string:
				for i, v := range val {
					m.Add("append", "index", i)
					m.Add("append", "value", v)
				}
				m.Table()
			case float64:
				m.Echo("%d", int(val))
			case nil:
				m.Echo("")
			default:
				m.Echo("%v", val)
			}
		}},
		"select": &Command{Name: "select key value field",
			Form: map[string]int{"parse": 2, "group": 1, "order": 2, "limit": 1, "offset": 1, "fields": 1, "format": 2, "trans_map": 3, "vertical": 0},
			Help: "选取数据", Hand: func(m *Message, c *Context, key string, arg ...string) {
				msg := m.Spawn()
				m.Set("result")

				nrow := len(m.Meta[m.Meta["append"][0]])
				for i := 0; i < nrow; i++ {
					if len(arg) == 0 || strings.Contains(m.Meta[arg[0]][i], arg[1]) {
						for _, k := range m.Meta["append"] {
							if m.Has("parse") && m.Option("parse") == k {
								var value interface{}
								m.Assert(json.Unmarshal([]byte(m.Meta[k][i]), &value))
								if m.Meta["parse"][1] != "" {
									value = Chain(m, value, m.Meta["parse"][1])
								}

								switch val := value.(type) {
								case map[string]interface{}:
									for k, v := range val {
										msg.Add("append", k, v)
									}
								case []interface{}:
								default:
								}
							} else {
								msg.Add("append", k, m.Meta[k][i])
							}
						}
					}
				}

				if m.Set("append"); m.Has("group") {

					group := m.Option("group")

					nrow := len(msg.Meta[msg.Meta["append"][0]])

					for i := 0; i < nrow; i++ {
						count := 1

						if group != "" && msg.Meta[group][i] == "" {
							msg.Add("append", "count", 0)
							continue
						}

						for j := i + 1; j < nrow; j++ {
							if group == "" || msg.Meta[group][i] == msg.Meta[group][j] {
								count++
								for _, k := range msg.Meta["append"] {
									if k == "count" {
										continue
									}
									if k == group {
										continue
									}
									m, e := strconv.Atoi(msg.Meta[k][i])
									if e != nil {
										continue
									}
									n, e := strconv.Atoi(msg.Meta[k][j])
									if e != nil {
										continue
									}
									msg.Meta[k][i] = fmt.Sprintf("%d", m+n)

								}

								if group != "" {
									msg.Meta[group][j] = ""
								}
							}
						}

						msg.Add("append", "count", count)
						for _, k := range msg.Meta["append"] {
							m.Add("append", k, msg.Meta[k][i])
						}
						if group == "" {
							break
						}
					}
				} else {
					m.Copy(msg, "append")
				}

				if m.Has("order") {
					m.Sort(m.Meta["order"][1], m.Option("order"))
				}

				offset := 0
				limit := m.Confi("page_limit")
				if m.Has("limit") {
					limit = m.Optioni("limit")
				}
				if m.Has("offset") {
					offset = m.Optioni("offset")
				}
				nrow = len(m.Meta[m.Meta["append"][0]])
				if offset > nrow {
					offset = nrow
				}
				if limit+offset > nrow {
					limit = nrow - offset
				}
				for _, k := range m.Meta["append"] {
					m.Meta[k] = m.Meta[k][offset : offset+limit]
				}

				if m.Option("fields") != "" {
					msg = m.Spawn()
					msg.Copy(m, "append", strings.Split(m.Option("fields"), " ")...)
					m.Set("append").Copy(msg, "append")
				}

				for i := 0; i < len(m.Meta["format"])-1; i += 2 {
					format := m.Meta["format"]
					for j, v := range m.Meta[format[i]] {
						m.Meta[format[i]][j] = fmt.Sprintf(format[i+1], v)
					}
				}

				if m.Has("vertical") {
					msg := m.Spawn()
					nrow := len(m.Meta[m.Meta["append"][0]])
					sort.Strings(m.Meta["append"])
					msg.Add("append", "field", "")
					msg.Add("append", "value", "")
					for i := 0; i < nrow; i++ {
						for _, k := range m.Meta["append"] {
							msg.Add("append", "field", k)
							msg.Add("append", "value", m.Meta[k][i])
						}
						msg.Add("append", "field", "")
						msg.Add("append", "value", "")
					}
					m.Set("append").Copy(msg, "append")
				}

				if len(arg) > 2 {
					if len(m.Meta[arg[2]]) > 0 {
						m.Echo(m.Meta[arg[2]][0])
					}
					return
				}

				m.Set("result").Table()
			}},
		"import": &Command{Name: "import filename", Help: "导入数据", Hand: func(m *Message, c *Context, key string, arg ...string) {
			f, e := os.Open(arg[0])
			m.Assert(e)
			defer f.Close()

			switch {
			case strings.HasSuffix(arg[0], ".json"):
				var data interface{}
				de := json.NewDecoder(f)
				de.Decode(&data)
				m.Optionv("data", data)

				switch d := data.(type) {
				case []interface{}:
					for _, value := range d {
						switch val := value.(type) {
						case map[string]interface{}:
							for k, v := range val {
								m.Add("append", k, v)
							}
						}
					}
				}
			case strings.HasSuffix(arg[0], ".csv"):
				r := csv.NewReader(f)
				l, e := r.Read()
				m.Assert(e)
				m.Meta["append"] = l

				for {
					l, e := r.Read()
					if e != nil {
						break
					}
					for i, v := range l {
						m.Add("append", m.Meta["append"][i], v)
					}
				}
			}
			m.Table()
		}},
	},
}

func Start() {
	Index.root = Index
	Pulse.root = Pulse

	for _, m := range Pulse.Search("") {
		m.target.root = Index
		m.target.Begin(m)
	}

	for k, c := range Index.contexts {
		Pulse.Sess(k, c)
	}

	args := os.Args[1:]
	if len(args) > 0 {
		if strings.HasSuffix(args[0], ".log") {
			Pulse.Sess("log", false).Conf("bench.log", args[0])
			args = args[1:]
		}
	}

	Pulse.Sess("yac", false).Cmd("init")
	for _, v := range Pulse.Sess("cli", false).Cmd("source", args).Meta["result"] {
		fmt.Printf("%v", v)
	}
}
