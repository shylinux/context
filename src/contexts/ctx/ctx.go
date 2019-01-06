package ctx

import (
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"errors"
	"io"
	"math/rand"
	"os"
	"runtime/debug"
	"sort"
	"time"
	"toolkit"
)

type Cache struct {
	Value string
	Name  string
	Help  string
	Hand  func(m *Message, x *Cache, arg ...string) string
}
type Config struct {
	Value interface{}
	Name  string
	Help  string
	Hand  func(m *Message, x *Config, arg ...string) string
}
type Command struct {
	Form map[string]int
	Name string
	Help interface{}
	Hand func(m *Message, c *Context, key string, arg ...string) (e error)
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

	message  *Message
	requests []*Message
	sessions []*Message

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
	s := &Context{Name: name, Help: help, root: c.root, context: c, message: m,
		Caches:   map[string]*Cache{},
		Configs:  map[string]*Config{},
		Commands: map[string]*Command{},
	}

	if m.target = s; c.Server != nil {
		c.Register(s, c.Server.Spawn(m, s, m.Meta["detail"]...))
	} else {
		c.Register(s, nil)
	}
	return s
}
func (c *Context) Begin(m *Message, arg ...string) *Context {
	if len(arg) > 0 {
		m.Set("detail", arg)
	}

	module := c.Name
	if c.context != nil {
		module = c.context.Name + "." + c.Name
	}

	c.Caches["module"] = &Cache{Name: "module", Value: module, Help: "模块域名"}
	c.Caches["status"] = &Cache{Name: "status(begin/start/close)", Value: "begin", Help: "模块状态, begin: 初始完成, start: 正在运行, close: 运行结束"}
	c.Caches["stream"] = &Cache{Name: "stream", Value: "", Help: "模块数据"}

	c.message = m
	c.requests = append(c.requests, m)
	m.source.sessions = append(m.source.sessions, m)

	m.Log("begin", "%d context %v %v", m.Capi("ncontext", 1), m.Meta["detail"], m.Meta["option"])
	for k, x := range c.Configs {
		if x.Hand != nil {
			m.Log("begin", "%s config %v", k, m.Conf(k, x.Value))
		}
	}

	if c.Server != nil {
		c.Server.Begin(m, m.Meta["detail"]...)
	}

	return c
}
func (c *Context) Start(m *Message, arg ...string) bool {
	sync := false
	if len(arg) > 0 && arg[0] == "sync" {
		sync, arg = true, arg[1:]
	}
	m.Set("detail", arg)

	c.requests = append(c.requests, m)
	m.source.sessions = append(m.source.sessions, m)
	if m.Hand = true; m.Cap("status") == "start" {
		return true
	}

	c.exit = make(chan bool, 2)
	go m.TryCatch(m, true, func(m *Message) {
		m.Log(m.Cap("status", "start"), "%d server %v %v", m.root.Capi("nserver", 1), m.Meta["detail"], m.Meta["option"])

		c.message = m
		if c.exit <- false; c.Server == nil || c.Server.Start(m, m.Meta["detail"]...) {
			c.Close(m, m.Meta["detail"]...)
			c.exit <- true
		}
	})

	if sync {
		for !<-c.exit {
		}
		return true
	}
	return <-c.exit
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
						c.requests[j] = c.requests[j+1]
					}
					c.requests = c.requests[:len(c.requests)-1]
				}
			}
		}
	}

	m.Log("close", "%d:%d %v", len(c.requests), len(c.sessions), arg)
	if len(c.requests) > 0 {
		return false
	}

	if m.Cap("status") == "start" {
		m.Log(m.Cap("status", "close"), "%d server %v", m.root.Capi("nserver", -1), arg)
		for _, msg := range c.sessions {
			if msg.Cap("status") != "close" {
				msg.target.Close(msg, arg...)
			}
		}
	}

	if c.context != nil {
		m.Log("close", "%d context %v", m.root.Capi("ncontext", -1), arg)
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
	case 2:
		if _, ok := c.Commands[key[0]]; ok && key[1] == "command" {
			return true
		}
		if _, ok := c.Configs[key[0]]; ok && key[1] == "config" {
			return true
		}
		if _, ok := c.Caches[key[0]]; ok && key[1] == "cache" {
			return true
		}
	case 1:
		if _, ok := c.Commands[key[0]]; ok {
			return true
		}
		if _, ok := c.Configs[key[0]]; ok {
			return true
		}
		if _, ok := c.Caches[key[0]]; ok {
			return true
		}
	}
	return false
}
func (c *Context) Travel(m *Message, hand func(m *Message, n int) (stop bool)) *Context {
	target := m.target

	cs := []*Context{c}
	for i := 0; i < len(cs); i++ {
		if m.target = cs[i]; hand(m, i) {
			return cs[i]
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
	return target
}
func (c *Context) BackTrace(m *Message, hand func(m *Message) (stop bool)) *Context {
	target := m.target

	for s := m.target; s != nil; s = s.context {
		if m.target = s; hand(m) {
			return s
		}
	}

	m.target = target
	return target
}

type LOGGER interface {
	LOG(*Message, string, string)
}
type DEBUG interface {
	Wait(*Message, ...interface{}) interface{}
	Goon(interface{}, ...interface{})
}
type Message struct {
	time time.Time
	code int

	source *Context
	target *Context

	Meta map[string][]string
	Data map[string]interface{}

	callback func(msg *Message) (sub *Message)
	Sessions map[string]*Message

	messages []*Message
	message  *Message
	root     *Message

	Remote chan bool
	Hand   bool
}

func (m *Message) Log(action string, str string, arg ...interface{}) *Message {
	kit.Errorf(fmt.Sprintf("%s %s %s", m.Format(), action, fmt.Sprintf(str, arg...)))
	return m
	if action == "error" {
		kit.Errorf(str, arg...)
	}

	if l := m.Sess("log", false); l != nil {
		if log, ok := l.target.Server.(LOGGER); ok {
			log.LOG(m, action, fmt.Sprintf(str, arg...))
		}
	}

	return m
}
func (m *Message) Gdb(arg ...interface{}) interface{} {
	if g := m.Sess("gdb", false); g != nil {
		if gdb, ok := g.target.Server.(DEBUG); ok {
			return gdb.Wait(m, arg...)
		}
	}
	return nil
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
		time:    time.Now(),
		code:    m.root.Capi("nmessage", 1),
		source:  m.target,
		target:  c,
		message: m,
		root:    m.root,
	}

	m.messages = append(m.messages, msg)
	return msg
}
func (m *Message) Time(arg ...interface{}) string {
	t := m.time

	if len(arg) > 0 {
		if d, e := time.ParseDuration(arg[0].(string)); e == nil {
			arg = arg[1:]
			t.Add(d)
		}
	}

	str := m.Conf("time_format")
	if len(arg) > 1 {
		str = fmt.Sprintf(arg[0].(string), arg[1:]...)
	} else if len(arg) > 0 {
		str = fmt.Sprintf("%v", arg[0])
	}
	return t.Format(str)
}
func (m *Message) Code() int {
	return m.code
}
func (m *Message) Source() *Context {
	return m.source
}
func (m *Message) Target() *Context {
	return m.target
}
func (m *Message) Message() *Message {
	return m.message
}
func (m *Message) Format(arg ...string) string {
	if len(arg) == 0 {
		arg = append(arg, "time", "ship")
	}

	meta := []string{}
	for _, v := range arg {
		switch v {
		case "time":
			meta = append(meta, m.Time())
		case "code":
			meta = append(meta, kit.Format(m.code))
		case "ship":
			meta = append(meta, fmt.Sprintf("%d(%s->%s)", m.code, m.source.Name, m.target.Name))
		case "detail":
			meta = append(meta, fmt.Sprintf("%v", m.Meta["detail"]))
		case "option":
			meta = append(meta, fmt.Sprintf("%v", m.Meta["option"]))
		case "append":
			meta = append(meta, fmt.Sprintf("%v", m.Meta["append"]))
		case "result":
			meta = append(meta, fmt.Sprintf("%v", m.Meta["result"]))
		}
	}
	return strings.Join(meta, " ")
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
		m.Meta[meta] = append(m.Meta[meta], kit.Trans(value...)...)

	case "option", "append":
		if _, ok := m.Meta[key]; !ok {
			m.Meta[key] = make([]string, 0, 3)
		}
		m.Meta[key] = append(m.Meta[key], kit.Trans(value...)...)

		for _, v := range m.Meta[meta] {
			if v == key {
				return m
			}
		}
		m.Meta[meta] = append(m.Meta[meta], key)

	default:
		m.Log("error", "add meta error %s %s %v", meta, key, value)
	}

	return m
}
func (m *Message) Set(meta string, arg ...interface{}) *Message {
	switch meta {
	case "detail", "result":
		if m != nil && m.Meta != nil {
			delete(m.Meta, meta)
		}
	case "option", "append":
		if len(arg) > 0 {
			delete(m.Meta, kit.Format(arg[0]))
		} else {
			for _, k := range m.Meta[meta] {
				delete(m.Data, k)
				delete(m.Meta, k)
			}
			delete(m.Meta, meta)
		}
	default:
		m.Log("error", "set meta error %s %s %v", meta, arg)
	}

	if args := kit.Trans(arg...); len(args) > 0 {
		m.Add(meta, args[0], args[1:])
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
		m.Log("error", "put data error %s %s %v", meta, key, value)
	}
	return m
}
func (m *Message) Get(key string, arg ...interface{}) string {
	if meta, ok := m.Meta[key]; ok && len(meta) > 0 {
		index := 0
		if len(arg) > 0 {
			index = kit.Int(arg[0])
		}

		index = (index+2)%(len(meta)+2) - 2
		if index > 0 && index < len(meta) {
			return meta[index]
		}
	}
	return ""
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
func (m *Message) CopyTo(msg *Message, arg ...string) *Message {
	if m == msg {
		return m
	}
	if len(arg) == 0 {
		if msg.Hand {
			msg.Copy(m, "append").Copy(m, "result")
		} else {
			msg.Copy(m, "option")
		}
	} else {
		msg.Copy(m, arg...)
	}
	return m
}
func (m *Message) Copy(msg *Message, arg ...string) *Message {
	if m == msg {
		return m
	}
	if len(arg) == 0 {
		if msg.Hand {
			arg = append(arg, "append")
		} else {
			arg = append(arg, "option")
		}
	}

	for i := 0; i < len(arg); i++ {
		meta := arg[i]

		switch meta {
		case "target":
			m.target = msg.target
		case "callback":
			m.callback = msg.callback
		// case "session":
		// 	if len(arg) == 0 {
		// 		for k, v := range msg.Sessions {
		// 			m.Sessions[k] = v
		// 		}
		// 	} else {
		// 		for _, k := range arg {
		// 			m.Sessions[k] = msg.Sessions[k]
		// 		}
		// 	}
		case "detail", "result":
			if len(msg.Meta[meta]) > 0 {
				m.Add(meta, msg.Meta[meta][0], msg.Meta[meta][1:])
			}
		case "option", "append":
			if i == len(arg)-1 {
				arg = append(arg, msg.Meta[meta]...)
			}

			for i++; i < len(arg); i++ {
				if v, ok := msg.Data[arg[i]]; ok {
					m.Put(meta, arg[i], v)
				}
				if v, ok := msg.Meta[arg[i]]; ok {
					m.Add(meta, arg[i], v)
				}
			}
		default:
			if msg.Hand {
				meta = "append"
			} else {
				meta = "option"
			}

			if v, ok := msg.Data[arg[i]]; ok {
				m.Put(meta, arg[i], v)
			}
			if v, ok := msg.Meta[arg[i]]; ok {
				m.Add(meta, arg[i], v)
			}
		}
	}

	return m
}
func (m *Message) Echo(str string, arg ...interface{}) *Message {
	if len(arg) > 0 {
		return m.Add("result", fmt.Sprintf(str, arg...))
	}
	return m.Add("result", str)
}

func (m *Message) Insert(meta string, index int, arg ...interface{}) string {
	if m.Meta == nil {
		m.Meta = make(map[string][]string)
	}
	m.Meta[meta] = kit.Array(m.Meta[meta], index, arg)

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
	return kit.Int(m.Detail(arg...))
}
func (m *Message) Details(arg ...interface{}) bool {
	return kit.Right(m.Detail(arg...))
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
	return kit.Int(m.Result(arg...))
}
func (m *Message) Results(arg ...interface{}) bool {
	return kit.Right(m.Result(arg...))
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
	return kit.Int(m.Option(key, arg...))

}
func (m *Message) Options(key string, arg ...interface{}) bool {
	return kit.Right(m.Option(key, arg...))
}
func (m *Message) Optionv(key string, arg ...interface{}) interface{} {
	if len(arg) > 0 {
		switch arg[0].(type) {
		case nil:
		// case []string:
		// 	m.Option(key, v...)
		// case string:
		// 	m.Option(key, v)
		default:
			m.Put("option", key, arg[0])
		}
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
func (m *Message) Optionx(key string, arg ...string) interface{} {
	value := m.Conf(key)
	if value == "" {
		value = m.Option(key)
	}

	if len(arg) > 0 {
		value = fmt.Sprintf(arg[0], value)
	}
	return value
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
	return kit.Int(m.Append(key, arg...))

}
func (m *Message) Appends(key string, arg ...interface{}) bool {
	return kit.Right(m.Append(key, arg...))
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
func (m *Message) Table(cbs ...func(maps map[string]string, list []string, line int) (goon bool)) *Message {
	if len(m.Meta["append"]) == 0 {
		return m
	}

	//计算列宽
	depth, width := 0, map[string]int{}
	for _, k := range m.Meta["append"] {
		if len(m.Meta[k]) > depth {
			depth = len(m.Meta[k])
		}
		for _, v := range m.Meta[k] {
			if len(v) > width[k] {
				width[k] = len(v)
			}
		}
	}

	space := m.Confx("table_space")
	var cb func(maps map[string]string, list []string, line int) (goon bool)
	if len(cbs) > 0 {
		cb = cbs[0]
	} else {
		row := m.Confx("table_row_sep")
		col := m.Confx("table_col_sep")
		compact := kit.Right(m.Confx("table_compact"))
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

	// 输出表头
	row := map[string]string{}
	wor := []string{}
	for _, k := range m.Meta["append"] {
		row[k], wor = k, append(wor, k+strings.Repeat(space, width[k]-len(k)))
	}
	if !cb(row, wor, -1) {
		return m
	}

	// 输出数据
	for i := 0; i < depth; i++ {
		row := map[string]string{}
		wor := []string{}
		for _, k := range m.Meta["append"] {
			data := ""
			if i < len(m.Meta[k]) {
				data = m.Meta[k][i]
			}

			row[k], wor = data, append(wor, data+strings.Repeat(space, width[k]-len(data)))
		}
		if !cb(row, wor, i) {
			break
		}
	}

	return m
}
func (m *Message) Sort(key string, arg ...string) *Message {
	cmp := "string"
	if len(arg) > 0 {
		cmp = arg[0]
	}

	number := map[int]int{}
	table := []map[string]string{}
	m.Table(func(line map[string]string, lists []string, index int) bool {
		if index != -1 {
			table = append(table, line)
			switch cmp {
			case "int":
				number[index] = kit.Int(line[key])
			case "int_r":
				number[index] = -kit.Int(line[key])
			case "time":
				number[index] = kit.Time(line[key])
			case "time_r":
				number[index] = -kit.Time(line[key])
			}
		}
		return true
	})

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
			default:
				if number[i] > number[j] {
					result = true
				}
			}

			if result {
				table[i], table[j] = table[j], table[i]
				number[i], number[j] = number[j], number[i]
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
func (m *Message) Parse(arg interface{}) string {
	switch str := arg.(type) {
	case string:
		if len(str) > 1 && str[0] == '$' {
			return m.Cap(str[1:])
		}
		if len(str) > 1 && str[0] == '@' {
			return m.Confx(str[1:])
		}
		return m.Cmdx(str)
	}
	return ""
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
			m.Log("error", "context not find %s", name)
			return nil
		}
	}
	return m.Spawn(target)
}
func (m *Message) Search(key string, root ...bool) []*Message {
	reg, e := regexp.Compile(key)
	m.Assert(e)

	target := m.target
	if target == nil {
		return []*Message{nil}
	}
	if len(root) > 0 && root[0] {
		target = m.target.root
	}

	cs := make([]*Context, 0, 3)
	target.Travel(m, func(m *Message, i int) bool {
		if reg.MatchString(m.target.Name) || reg.FindString(m.target.Help) != "" {
			m.Log("search", "%d %s match [%s]", len(cs), m.target.Name, key)
			cs = append(cs, m.target)
		}
		return false
	})

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
	if key == "" {
		return m.Spawn()
	}

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
			root := len(arg) < 3 || kit.Right(arg[2])

			method := "find"
			if len(arg) > 1 {
				method = kit.Format(arg[1])
			}

			switch method {
			case "find":
				m.Sessions[key] = m.Find(value, root)
			case "search":
				m.Sessions[key] = m.Search(value, root)[0]
			}
			return m.Sessions[key]
		case nil:
			delete(m.Sessions, key)
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
func (m *Message) Match(key string, spawn bool, hand func(m *Message, s *Context, c *Context, key string) bool) *Message {
	if strings.Contains(key, ".") {
		arg := strings.Split(key, ".")
		m, key = m.Sess(arg[0], spawn), arg[1]
	}

	context := []*Context{m.target}
	for _, v := range []string{"aaa", "cli"} {
		if msg := m.Sess(v, false); msg != nil && msg.target != nil {
			context = append(context, msg.target)
		}
	}
	context = append(context, m.source)

	for _, s := range context {
		for c := s; c != nil && !hand(m, s, c, key) && c != c.context; c = c.context {
		}
	}
	return m
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

func (m *Message) Assert(e interface{}, msg ...string) bool {
	switch v := e.(type) {
	case nil:
		return true
	case *Message:
		if v.Result(0) != "error: " {
			return true
		}
		e = v.Result(1)
	default:
		if kit.Right(v) {
			return true
		}
	}

	switch e.(type) {
	case error:
	default:
		e = errors.New(kit.Format(msg))
	}

	m.Log("error", "%v", e)
	panic(m.Set("result", "error: ", kit.Format(e), "\n"))
}
func (m *Message) TryCatch(msg *Message, safe bool, hand ...func(msg *Message)) *Message {
	defer func() {
		e := recover()
		switch e {
		case io.EOF:
		case nil:
		default:
			if len(hand) > 1 {
				m.TryCatch(msg, safe, hand[1:]...)
			} else if !safe {
				m.Log("error", "%s not catch %v", msg.Format(), e)
				debug.PrintStack()
				msg.Assert(e)
			}
		}
	}()

	if len(hand) > 0 {
		hand[0](msg)
	}

	return m
}
func (m *Message) Start(name string, help string, arg ...string) bool {
	return m.Set("detail", arg).target.Spawn(m, name, help).Begin(m).Start(m)
}
func (m *Message) Wait() bool {
	if m.target.exit != nil {
		return <-m.target.exit
	}
	return true
}

func (m *Message) Cmdy(args ...interface{}) *Message {
	m.Cmd(args...).CopyTo(m)
	return m
}
func (m *Message) Cmdx(args ...interface{}) string {
	return m.Cmd(args...).Result(0)
}
func (m *Message) Cmds(args ...interface{}) bool {
	return m.Cmd(args...).Results(0)
}
func (m *Message) Cmd(args ...interface{}) *Message {
	if m == nil {
		return m
	}

	if len(args) > 0 {
		m.Set("detail", kit.Trans(args...))
	}

	key, arg := m.Meta["detail"][0], m.Meta["detail"][1:]

	m = m.Match(key, true, func(m *Message, s *Context, c *Context, key string) bool {
		if x, ok := c.Commands[key]; ok && x.Hand != nil {
			m.TryCatch(m, true, func(m *Message) {
				m.Log("cmd", "%s:%s %s %v %v", c.Name, s.Name, key, arg, m.Meta["option"])

				if args := []string{}; x.Form != nil {
					for i := 0; i < len(arg); i++ {
						if n, ok := x.Form[arg[i]]; ok {

							if n < 0 {
								n += len(arg) - i
							}
							for j := i + 1; j <= i+n; j++ {
								if _, ok := x.Form[arg[j]]; ok {
									n = j - i - 1
								}
							}
							m.Add("option", arg[i], arg[i+1:i+1+n])
							i += n
						} else {
							args = append(args, arg[i])
						}
					}
					arg = args
				}

				m.Hand = true
				x.Hand(m, c, key, arg...)
			})
		}
		return m.Hand
	})

	if !m.Hand {
		m.Log("error", "cmd run error %s", m.Format())
	}
	return m
}

func (m *Message) Confm(key string, args ...interface{}) map[string]interface{} {
	var chain interface{}
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case []interface{}:
		case []string:
			chain, args = arg, args[1:]
		}
	}

	var v interface{}
	if chain == nil {
		v = m.Confv(key)
	} else {
		v = m.Confv(key, chain)
	}

	table, _ := v.([]interface{})
	value, _ := v.(map[string]interface{})
	if len(args) == 0 {
		return value
	}

	switch fun := args[0].(type) {
	case func(map[string]interface{}):
		fun(value)
	case func(string, map[string]interface{}):
		for k, v := range value {
			if val, ok := v.(map[string]interface{}); ok {
				fun(k, val)
			}
		}
	case func(int, map[string]interface{}):
		for i, v := range table {
			if val, ok := v.(map[string]interface{}); ok {
				fun(i, val)
			}
		}
	}
	return value
}
func (m *Message) Confx(key string, args ...interface{}) string {
	value := kit.Select(m.Conf(key), m.Option(key))
	if len(args) == 0 {
		return value
	}

	switch arg := args[0].(type) {
	case []string:
		if len(args) > 1 {
			value, args = kit.Select(value, arg, args[1]), args[1:]
		} else {
			value = kit.Select(value, arg)
		}
	case map[string]interface{}:
		value = kit.Select(value, kit.Format(arg[key]))
	case string:
		value = kit.Select(value, arg)
	case nil:
	default:
		value = kit.Select(value, args[0])
	}

	format := "%s"
	if args = args[1:]; len(args) > 0 {
		format, args = kit.Format(args[0]), args[1:]
	}

	arg := []interface{}{format, value}
	for _, v := range args {
		args = append(args, v)
	}
	return kit.Format(arg...)
}
func (m *Message) Confs(key string, arg ...interface{}) bool {
	return kit.Right(m.Confv(key, arg...))
}
func (m *Message) Confi(key string, arg ...interface{}) int {
	return kit.Int(m.Confv(key, arg...))
}
func (m *Message) Confv(key string, args ...interface{}) interface{} {
	var config *Config
	m.Match(key, false, func(m *Message, s *Context, c *Context, key string) bool {
		if x, ok := c.Configs[key]; ok {
			config = x
			return true
		}
		return false
	})

	if len(args) == 0 {
		if config == nil {
			return nil
		}
		return config.Value
	}

	if config == nil {
		config = &Config{}
		m.target.Configs[key] = config
	}

	switch config.Value.(type) {
	case string:
		config.Value = kit.Format(args...)
	case bool:
		config.Value = kit.Right(args...)
	case int:
		config.Value = kit.Int(args...)
	case nil:
		config.Value = args[0]
	default:
		for i := 0; i < len(args); i += 2 {
			if i < len(args)-1 {
				config.Value = kit.Chain(config.Value, args[i], args[i+1])
			} else {
				return kit.Chain(config.Value, args[i])
			}
		}
	}

	return config.Value
}
func (m *Message) Conf(key string, args ...interface{}) string {
	return kit.Format(m.Confv(key, args...))
}
func (m *Message) Caps(key string, arg ...interface{}) bool {
	if len(arg) > 0 {
		return kit.Right(m.Cap(key, arg...))
	}
	return kit.Right(m.Cap(key))
}
func (m *Message) Capi(key string, arg ...interface{}) int {
	n := kit.Int(m.Cap(key))
	if len(arg) > 0 {
		return kit.Int(m.Cap(key, n+kit.Int(arg...)))
	}
	return n
}
func (m *Message) Cap(key string, arg ...interface{}) string {
	if len(arg) == 0 {
		if val, ok := m.Gdb("cache", "read", key).(string); ok {
			return val
		}
	} else {
		if val, ok := m.Gdb("cache", "write", key, arg[0]).(string); ok {
			return val
		}
	}

	var cache *Cache
	m.Match(key, false, func(m *Message, s *Context, c *Context, key string) bool {
		if x, ok := c.Caches[key]; ok {
			cache = x
			return true
		}
		return false
	})

	if len(arg) == 0 {
		if cache == nil {
			return ""
		}
		if cache.Hand != nil {
			return cache.Hand(m, cache)
		}
		return cache.Value
	}

	if cache == nil {
		cache = &Cache{}
		m.target.Caches[key] = cache
	}

	if cache.Hand != nil {
		cache.Value = cache.Hand(m, cache, kit.Format(arg...))
	} else {
		cache.Value = kit.Format(arg...)
	}
	return cache.Value
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

		return strings.Join(Pulse.Sess("cli").Cmd(arg).Meta["result"], "")
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
	"slice": func(list interface{}, arg ...interface{}) interface{} {
		switch l := list.(type) {
		case string:
			if len(arg) == 0 {
				return l
			}
			if len(arg) == 1 {
				return l[arg[0].(int):]
			}
			if len(arg) == 2 {
				return l[arg[0].(int):arg[1].(int)]
			}
		}

		return ""
	},

	"work": func(m *Message, arg ...interface{}) interface{} {
		switch len(arg) {
		case 0:
			list := map[string]map[string]interface{}{}
			m.Confm("auth", []string{m.Option("sessid"), "ship"}, func(key string, ship map[string]interface{}) {
				if ship["type"] == "bench" {
					if work := m.Confm("auth", key); work != nil {
						list[key] = work
					}
				}
			})
			return list
		}
		return nil
	},

	"unescape": func(str string) interface{} {
		return template.HTML(str)
	},
	"json": func(arg ...interface{}) interface{} {
		if len(arg) == 0 {
			return ""
		}

		b, _ := json.MarshalIndent(arg[0], "", "  ")
		return string(b)
	},
	"so": func(arg ...interface{}) interface{} {
		if len(arg) == 0 {
			return ""
		}

		cli := Pulse.Sess("cli")
		cmd := strings.Join(kit.Trans(arg), " ")
		cli.Cmd("source", cmd)

		result := []string{}
		if len(cli.Meta["append"]) > 0 {
			result = append(result, "<table>")
			result = append(result, "<caption>", cmd, "</caption>")
			cli.Table(func(maps map[string]string, list []string, line int) bool {
				if line == -1 {
					result = append(result, "<tr>")
					for _, v := range list {
						result = append(result, "<th>", v, "</th>")
					}
					result = append(result, "</tr>")
					return true
				}
				result = append(result, "<tr>")
				for _, v := range list {
					result = append(result, "<td>", v, "</td>")
				}
				result = append(result, "</tr>")
				return true
			})
			result = append(result, "</table>")
		} else {
			result = append(result, "<pre><code>")
			result = append(result, fmt.Sprintf("%s", cli.Find("shy", false).Conf("prompt")), cmd, "\n")
			result = append(result, cli.Meta["result"]...)
			result = append(result, "</code></pre>")
		}

		return template.HTML(strings.Join(result, ""))
	},
}
var Pulse = &Message{code: 0, time: time.Now(), source: Index, target: Index, Meta: map[string][]string{}}
var Index = &Context{Name: "ctx", Help: "模块中心", Server: &CTX{},
	Caches: map[string]*Cache{
		"begin_time": &Cache{Name: "begin_time", Value: "", Help: "启动时间"},
		"nserver":    &Cache{Name: "nserver", Value: "0", Help: "服务数量"},
		"ncontext":   &Cache{Name: "ncontext", Value: "0", Help: "模块数量"},
		"nmessage":   &Cache{Name: "nmessage", Value: "1", Help: "消息数量"},
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
		"table_space":   &Config{Name: "table_space", Value: " ", Help: "命令列表帮助"},

		"page_offset": &Config{Name: "page_offset", Value: "0", Help: "列表偏移"},
		"page_limit":  &Config{Name: "page_limit", Value: "10", Help: "列表大小"},

		"time_format": &Config{Name: "time_format", Value: "2006-01-02 15:04:05", Help: "时间格式"},
	},
	Commands: map[string]*Command{
		"help": &Command{Name: "help topic", Help: "帮助", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
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
				m.Echo("More: https://github.com/shylinux/context\n")
				return
			}

			switch arg[0] {
			case "context":
				switch len(arg) {
				case 1:
					keys := []string{}
					values := map[string]*Context{}
					m.Target().root.Travel(m, func(m *Message, i int) bool {
						if _, ok := values[m.Cap("module")]; !ok {
							keys = append(keys, m.Cap("module"))
							values[m.Cap("module")] = m.Target()
						}
						return false
					})

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
							switch help := v.Help.(type) {
							case []string:
								m.Echo("%s: %s\n", k, v.Name)
								for _, v := range help {
									m.Echo("  %s\n", v)
								}
							case string:
								m.Echo("%s: %s\n%s\n", k, v.Name, v.Help)
							}
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

			return
		}},

		"message": &Command{Name: "message [code] [cmd...]", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
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
			return
		}},
		"detail": &Command{Name: "detail [index] [value...]", Help: "查看或添加参数", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
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
			return
		}},
		"option": &Command{Name: "option [all] [key [index] [value...]]", Help: "查看或添加选项", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
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
						m.Add("append", "len", len(msg.Meta[k]))
						m.Add("append", "value", fmt.Sprintf("%v", msg.Meta[k]))
						continue
					}

					if k != arg[0] {
						continue
					}

					if len(arg) > 1 {
						msg.Meta[k] = kit.Array(msg.Meta[k], index, arg[1:])
						m.Echo("%v", msg.Meta[k])
						return
					}

					if index != -100 {
						m.Echo(kit.Array(msg.Meta[k], index)[0])
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
			m.Sort("key", "string").Table()
			return
		}},
		"result": &Command{Name: "result [index] [value...]", Help: "查看或添加返回值", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
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
			return
		}},
		"append": &Command{Name: "append [all] [key [index] [value...]]", Help: "查看或添加附加值", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
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
						msg.Meta[k] = kit.Array(msg.Meta[k], index, arg[1:])
						m.Echo("%v", msg.Meta[k])
						return
					}

					if index != -100 {
						m.Echo(kit.Array(msg.Meta[k], index)[0])
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
			return
		}},
		"session": &Command{Name: "session [all] [key [module]]", Help: "查看或添加会话", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
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
			return
		}},
		"callback": &Command{Name: "callback", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			msg := m.message
			for msg := msg; msg != nil; msg = msg.message {
				m.Add("append", "msg", msg.code)
				m.Add("append", "fun", msg.callback)
			}
			m.Table()
			return
		}},

		"context": &Command{Name: "context [find|search] [root|back|home] [first|last|rand|magic] [module] [cmd|switch|list|spawn|start|close]",
			Help: "查找并操作模块;\n查找方法, find: 精确查找, search: 模糊搜索;\n查找起点, root: 根模块, back: 父模块, home: 本模块;\n过滤结果, first: 取第一个, last: 取最后一个, rand: 随机选择, magic: 智能选择;\n操作方法, cmd: 执行命令, switch: 切换为当前, list: 查看所有子模块, spwan: 创建子模块并初始化, start: 启动模块, close: 结束模块",
			Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
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

						if m.Options("sso_bench") && m.Options("sso_username") &&
							!m.Cmds("aaa.work", m.Option("sso_bench"), "right", m.Option("sso_username"), "componet", "source", "command", arg[0]) {

							m.Log("info", "sso check %v: %v failure", m.Option("sso_componet"), m.Option("sso_command"))
							m.Echo("error: ").Echo("no right [%s: %s %s]", m.Option("sso_componet"), m.Option("sso_command"), arg[0])
							break
						}

						if msg.Cmd(arg); !msg.Hand {
							msg = msg.Sess("cli").Cmd("cmd", arg)
						}
						msg.CopyTo(m)
					case "switch":
						m.target = msg.target
					case "list":
						m.Target().Travel(msg, func(msg *Message, n int) bool {
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
				return
			}},
		"command": &Command{Name: "command [all] [show]|[list [begin [end]] [prefix] test [key val]...]|[add [list_name name] [list_help help] cmd...]|[delete cmd]",
			Help: "查看或修改命令, show: 查看命令;\nlist: 查看列表命令, begin: 起始索引, end: 截止索引, prefix: 过滤前缀, test: 执行命令;\nadd: 添加命令, list_name: 命令别名, list_help: 命令帮助;\ndelete: 删除命令",
			Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
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
					c.BackTrace(m, func(m *Message) bool {
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

						return !all
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

					m.target.Commands[m.Cap("list_count")] = &Command{Name: strings.Join(arg, " "), Help: list_help, Hand: func(cmd *Message, c *Context, key string, args ...string) (e error) {
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
						return
					}}

					if list_name != "" {
						m.target.Commands[list_name] = m.target.Commands[m.Cap("list_count")]
					}
					m.Capi("list_count", 1)
				case "delete":
					c.BackTrace(m, func(m *Message) bool {
						delete(m.target.Commands, arg[0])
						return !all
					})
				}
				return
			}},
		"config": &Command{Name: "config [all] [export key..] [save|load file key...] [list|map arg...] [create map|list|string key name help] [delete key]",
			Help: "配置管理, export: 导出配置, save: 保存配置到文件, load: 从文件加载配置, create: 创建配置, delete: 删除配置",
			Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
				if len(arg) > 2 && arg[2] == "list" {
					chain := strings.Split(arg[1], ".")
					chain = append(chain, "-2")

					for _, val := range arg[3:] {
						m.Confv(arg[0], chain, val)
					}
					return
				}
				if len(arg) > 2 && arg[2] == "map" {
					chain := strings.Split(arg[1], ".")

					for i := 3; i < len(arg)-1; i += 2 {
						m.Confv(arg[0], append(chain, arg[i]), arg[i+1])
					}
					return
				}

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
							return e
						}
						defer f.Close()

						de := json.NewDecoder(f)
						if e = de.Decode(&save); e != nil {
							m.Log("info", "e: %v", e)
						}
					}

					c.BackTrace(m, func(m *Message) bool {
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
						return !all
					})
					m.Sort("key", "str").Table()

					switch action {
					case "save":
						buf, e := json.MarshalIndent(save, "", "  ")
						m.Assert(e)
						m.Sess("nfs").Add("option", "data", string(buf)).Cmd("save", which)
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

				msg := m.Spawn().Put("option", "_cache", value).Cmd("trans", "_cache")
				m.Copy(msg, "append").Copy(msg, "result")
				return
			}},
		"cache": &Command{Name: "cache [all] |key [value]|key = value|key name value help|delete key]",
			Help: "查看、读写、赋值、新建、删除缓存变量",
			Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
				all := false
				if len(arg) > 0 && arg[0] == "all" {
					arg, all = arg[1:], true
				}

				switch len(arg) {
				case 0:
					c.BackTrace(m, func(m *Message) bool {
						for k, v := range m.target.Caches {
							m.Add("append", "key", k)
							m.Add("append", "value", m.Cap(k))
							m.Add("append", "name", v.Name)
						}
						return !all
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
					m.Echo(m.Cap(arg[0], arg[1:]))
					return
				}
				return
			}},

		"trans": &Command{Name: "trans option [type|data|json] limit 10 [index...]", Help: "数据转换", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			value, arg := m.Optionv(arg[0]), arg[1:]

			view := "data"
			if len(arg) > 0 {
				switch arg[0] {
				case "type", "data", "json":
					view, arg = arg[0], arg[1:]
				}
			}

			limit := m.Confi("page_limit")
			if len(arg) > 0 && arg[0] == "limit" {
				limit, arg = kit.Int(arg[1]), arg[2:]
			}

			chain := strings.Join(arg, ".")
			if chain != "" {
				value = kit.Chain(value, chain)
			}

			switch view {
			case "type": // 查看数据类型
				switch value := value.(type) {
				case map[string]interface{}:
					for k, v := range value {
						m.Add("append", "key", k)
						m.Add("append", "type", fmt.Sprintf("%T", v))
					}
					m.Sort("key", "str").Table()
				case []interface{}:
					for k, v := range value {
						m.Add("append", "key", k)
						m.Add("append", "type", fmt.Sprintf("%T", v))
					}
					m.Sort("key", "int").Table()
				case nil:
				default:
					m.Add("append", "key", chain)
					m.Add("append", "type", fmt.Sprintf("%T", value))
					m.Sort("key", "str").Table()
				}
				return
			case "data":
			case "json": // 查看文本数据
				b, e := json.MarshalIndent(value, "", " ")
				m.Assert(e)
				m.Echo(string(b))
				return nil
			}

			switch val := value.(type) {
			case map[string]interface{}:
				for k, v := range val {
					m.Add("append", "key", k)
					switch val := v.(type) {
					case nil:
						m.Add("append", "value", "")
					case string:
						m.Add("append", "value", val)
					case float64:
						m.Add("append", "value", fmt.Sprintf("%d", int(val)))
					default:
						b, _ := json.Marshal(val)
						m.Add("append", "value", fmt.Sprintf("%s", string(b)))
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
				fields := map[string]int{}
				for i, v := range val {
					if i >= limit {
						break
					}
					switch val := v.(type) {
					case map[string]interface{}:
						for k, _ := range val {
							fields[k]++
						}
					}
				}

				if len(fields) > 0 {
					for i, v := range val {
						if i >= limit {
							break
						}
						switch val := v.(type) {
						case map[string]interface{}:
							for k, _ := range fields {
								switch value := val[k].(type) {
								case nil:
									m.Add("append", k, "")
								case string:
									m.Add("append", k, value)
								case float64:
									m.Add("append", k, fmt.Sprintf("%d", int(value)))
								default:
									b, _ := json.Marshal(value)
									m.Add("append", k, fmt.Sprintf("%v", string(b)))
								}
							}
						}
					}
				} else {
					for i, v := range val {
						switch val := v.(type) {
						case nil:
							m.Add("append", "index", i)
							m.Add("append", "value", "")
						case string:
							m.Add("append", "index", i)
							m.Add("append", "value", val)
						case float64:
							m.Add("append", "index", i)
							m.Add("append", "value", fmt.Sprintf("%v", int(val)))
						default:
							m.Add("append", "index", i)
							b, _ := json.Marshal(val)
							m.Add("append", "value", fmt.Sprintf("%v", string(b)))
						}
					}
				}
				m.Table()
			case []string:
				for i, v := range val {
					m.Add("append", "index", i)
					m.Add("append", "value", v)
				}
				m.Table()
			case string:
				m.Echo("%s", val)
			case float64:
				m.Echo("%d", int(val))
			case nil:
			default:
				b, _ := json.Marshal(val)
				m.Echo("%s", string(b))
			}
			return
		}},
		"select": &Command{Name: "select key value field",
			Form: map[string]int{"parse": 2, "hide": 1, "fields": -1, "group": 1, "order": 2, "limit": 1, "offset": 1, "format": -1, "trans_map": -1, "vertical": 0},
			Help: "选取数据", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
				msg := m.Set("result").Spawn()

				// 解析
				if len(m.Meta["append"]) == 0 {
					return
				}
				nrow := len(m.Meta[m.Meta["append"][0]])
				keys := []string{}
				for i := 0; i < nrow; i++ {
					for j := 0; j < len(m.Meta["parse"]); j += 2 {
						var value interface{}
						json.Unmarshal([]byte(m.Meta[m.Meta["parse"][j]][i]), &value)
						if m.Meta["parse"][j+1] != "" {
							value = kit.Chain(value, m.Meta["parse"][j+1])
						}

						switch val := value.(type) {
						case map[string]interface{}:
							for k, _ := range val {
								keys = append(keys, k)
							}
						default:
							keys = append(keys, m.Meta["parse"][j+1])
						}
					}
				}
				for i := 0; i < nrow; i++ {
					for _, k := range keys {
						m.Add("append", k, "")
					}
				}
				for i := 0; i < nrow; i++ {
					for j := 0; j < len(m.Meta["parse"]); j += 2 {
						var value interface{}
						json.Unmarshal([]byte(m.Meta[m.Meta["parse"][j]][i]), &value)
						if m.Meta["parse"][j+1] != "" {
							value = kit.Chain(value, m.Meta["parse"][j+1])
						}

						switch val := value.(type) {
						case map[string]interface{}:
							for k, v := range val {
								switch val := v.(type) {
								case string:
									m.Meta[k][i] = val
								case float64:
									m.Meta[k][i] = fmt.Sprintf("%d", int(val))
								default:
									b, _ := json.Marshal(val)
									m.Meta[k][i] = string(b)
								}
							}
						case string:
							m.Meta[m.Meta["parse"][j+1]][i] = val
						case float64:
							m.Meta[m.Meta["parse"][j+1]][i] = fmt.Sprintf("%d", int(val))
						default:
							b, _ := json.Marshal(val)
							m.Meta[m.Meta["parse"][j+1]][i] = string(b)
						}
					}
				}

				// 隐藏列
				hides := map[string]bool{}
				for _, k := range m.Meta["hide"] {
					hides[k] = true
				}
				for i := 0; i < nrow; i++ {
					if len(arg) == 0 || strings.Contains(m.Meta[arg[0]][i], arg[1]) {
						for _, k := range m.Meta["append"] {
							if hides[k] {
								continue
							}
							msg.Add("append", k, m.Meta[k][i])
						}
					}
				}

				// 选择列
				if m.Option("fields") != "" {
					msg = m.Spawn()
					m.Hand = true
					msg.Copy(m, strings.Split(strings.Join(m.Meta["fields"], " "), " ")...)
					m.Hand = false
					m.Set("append").Copy(msg, "append")
				}

				// 聚合
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

				// 排序
				if m.Has("order") {
					m.Sort(m.Meta["order"][1], m.Option("order"))
				}

				// 分页
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

				// 值转换
				for i := 0; i < len(m.Meta["trans_map"]); i += 3 {
					trans := m.Meta["trans_map"][i:]
					for j := 0; j < len(m.Meta[trans[0]]); j++ {
						if m.Meta[trans[0]][j] == trans[1] {
							m.Meta[trans[0]][j] = trans[2]
						}
					}
				}

				// 格式化
				for i := 0; i < len(m.Meta["format"])-1; i += 2 {
					format := m.Meta["format"]
					for j, v := range m.Meta[format[i]] {
						m.Meta[format[i]][j] = fmt.Sprintf(format[i+1], v)
					}
				}

				// 变换列
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

				// 取单值
				if len(arg) > 2 {
					if len(m.Meta[arg[2]]) > 0 {
						m.Echo(m.Meta[arg[2]][0])
					}
					return
				}

				m.Set("result").Table()
				return
			}},
	},
}

type CTX struct {
}

func (ctx *CTX) Spawn(m *Message, c *Context, arg ...string) Server {
	s := new(CTX)
	return s
}
func (ctx *CTX) Begin(m *Message, arg ...string) Server {
	m.Sess(m.target.Name, m)
	m.target.root = m.target
	m.root = m
	m.Cap("begin_time", m.Time())
	for _, msg := range m.Search("") {
		msg.target.root = m.target
		if msg.target == m.target {
			continue
		}
		msg.target.Begin(msg, arg...)
		m.Sess(msg.target.Name, msg)
	}
	return ctx
}
func (ctx *CTX) Start(m *Message, arg ...string) bool {
	m.Cmd("cli.source", arg)
	return false
}
func (ctx *CTX) Close(m *Message, arg ...string) bool {
	return true
}

func Start(args ...string) bool {
	if len(args) == 0 {
		args = append(args, os.Args[1:]...)
	}

	if Index.Begin(Pulse, args...); Index.Start(Pulse, args...) {
	}
	return false
}
