package ctx // {{{
// }}}
import ( // {{{
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

// }}}

func Right(str string) bool { // {{{
	switch str {
	case "", "0", "false", "off", "no", "error: ":
		return false
	}
	return true
}

// }}}
func Trans(arg ...interface{}) []string { // {{{
	ls := []string{}
	for _, v := range arg {
		switch val := v.(type) {
		case string:
			ls = append(ls, val)
		case bool:
			ls = append(ls, fmt.Sprintf("%t", val))
		case int, int8, int16, int32, int64:
			ls = append(ls, fmt.Sprintf("%d", val))
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
		default:
			ls = append(ls, fmt.Sprintf("%v", val))
		}
	}
	return ls
}

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

	contexts map[string]*Context
	context  *Context
	root     *Context

	exit chan bool
	Server
}

func (c *Context) Register(s *Context, x Server) { // {{{
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

// }}}
func (c *Context) Spawn(m *Message, name string, help string) *Context { // {{{
	s := &Context{Name: name, Help: help, root: c.root, context: c}

	if m.target = s; c.Server != nil {
		c.Register(s, c.Server.Spawn(m, s, m.Meta["detail"]...))
	} else {
		c.Register(s, nil)
	}

	return s
}

// }}}
func (c *Context) Begin(m *Message, arg ...string) *Context { // {{{
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

	m.Log("begin", "%d context %v %v", m.root.Capi("ncontext", 1), m.Meta["detail"], m.Meta["option"])
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
func (c *Context) Start(m *Message, arg ...string) bool { // {{{
	if len(arg) > 0 {
		m.Meta["detail"] = arg
	}

	c.requests = append(c.requests, m)
	m.source.sessions = append(m.source.sessions, m)
	if m.Hand = true; m.Cap("status") == "start" {
		return true
	}

	running := make(chan bool)
	go m.TryCatch(m, true, func(m *Message) {
		m.Log(m.Cap("status", "start"), "%d server %v %v", m.root.Capi("nserver", 1), m.Meta["detail"], m.Meta["option"])

		c.exit = make(chan bool, 1)
		if running <- true; c.Server != nil && c.Server.Start(m, m.Meta["detail"]...) {
			c.Close(m, m.Meta["detail"]...)
		}
	})
	return <-running
}

// }}}
func (c *Context) Close(m *Message, arg ...string) bool { // {{{
	m.Log("close", "%d:%d %v", len(c.requests), len(c.sessions), arg)

	if m.target == c {
		for i := len(c.requests) - 1; i >= 0; i-- {
			if msg := c.requests[i]; msg.code == m.code {
				if msg.source == c || c.Server == nil || c.Server.Close(m, arg...) {
					for j := i + 1; j < len(c.requests)-1; j++ {
						c.requests[j] = c.requests[j+1]
					}
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
			if msg.target != c {
				msg.target.Close(msg, arg...)
			}
		}
	}

	if c.context != nil {
		m.Log("close", "%d context %v", m.root.Capi("ncontext", -1)+1, arg)
		delete(c.context.contexts, c.Name)
		m.Wait()
	}
	return true
}

// }}}

func (c *Context) Context() *Context { // {{{
	return c.context
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

type Message struct {
	code int
	time time.Time

	source *Context
	target *Context

	Hand bool
	Meta map[string][]string
	Data map[string]interface{}

	callback func(msg *Message) (sub *Message)
	sessions map[string]*Message

	messages []*Message
	message  *Message
	root     *Message
}

func (m *Message) Code() int { // {{{
	return m.code
}

// }}}
func (m *Message) Message() *Message { // {{{
	return m.message
}

// }}}
func (m *Message) Source() *Context { // {{{
	return m.source
}

// }}}
func (m *Message) Target() *Context { // {{{
	return m.target
}

// }}}
func (m *Message) Format() string { // {{{
	return fmt.Sprintf("%d(%s->%s): %s %v %v", m.code, m.source.Name, m.target.Name, m.time.Format("15:04:05"), m.Meta["detail"], m.Meta["option"])
}

// }}}
func (m *Message) Tree(code int) *Message { // {{{
	ms := []*Message{m}
	for i := 0; i < len(ms); i++ {
		if ms[i].Code() == code {
			return ms[i]
		}
		ms = append(ms, ms[i].messages...)
	}
	return nil
}

// }}}
func (m *Message) Copy(msg *Message, meta string, arg ...string) *Message { // {{{
	switch meta {
	case "target":
		m.target = msg.target
	case "callback":
		m.callback = msg.callback
	case "session":
		if len(arg) == 0 {
			for k, v := range msg.sessions {
				m.sessions[k] = v
			}
		} else {
			for _, k := range arg {
				m.sessions[k] = msg.sessions[k]
			}
		}
	case "detail", "result":
		m.Set(meta, msg.Meta[meta]...)
	case "option", "append":
		if len(arg) == 0 {
			arg = msg.Meta[meta]
		}

		for _, k := range arg {
			if v, ok := msg.Data[k]; ok {
				m.Put(meta, k, v)
			}
			if v, ok := msg.Meta[k]; ok {
				m.Set(meta, k).Add(meta, k, v)
			}
		}
	}

	return m
}

// }}}

func (m *Message) Log(action string, str string, arg ...interface{}) *Message { // {{{
	if !m.Options("log") {
		return m
	}

	if l := m.Sess("log"); l != nil {
		l.Options("log", false)
		l.Cmd("log", action, fmt.Sprintf(str, arg...))
	}
	return m
}

// }}}
func (m *Message) Assert(e interface{}, msg ...string) bool { // {{{
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

	panic(m.Set("result", "error: ", fmt.Sprintln(e), "\n"))
}

// }}}
func (m *Message) TryCatch(msg *Message, safe bool, hand ...func(msg *Message)) *Message { // {{{
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

// }}}

func (m *Message) BackTrace(hand func(m *Message) bool, c ...*Context) { // {{{
	target := m.target
	if len(c) > 0 {
		target = c[0]
	}
	for s := target; s != nil; s = s.context {
		if m.target = s; !hand(m) {
			break
		}
	}
	m.target = target
}

// }}}
func (m *Message) Travel(hand func(m *Message, i int) bool, c ...*Context) { // {{{
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

// }}}

func (m *Message) Spawn(arg ...*Context) *Message { // {{{
	c := m.target
	if len(arg) > 0 {
		c = arg[0]
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
			m.Log("find", "not find %s", v)
			return nil
		}
	}
	m.Log("find", "find %s", name)
	return m.Spawn(target)
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

	return ms
}

// }}}
func (m *Message) Sess(key string, arg ...interface{}) *Message { // {{{
	spawn := true
	if _, ok := m.sessions[key]; !ok && len(arg) > 0 {
		if m.sessions == nil {
			m.sessions = make(map[string]*Message)
		}

		switch value := arg[0].(type) {
		case *Message:
			m.sessions[key] = value
			return m.sessions[key]
		case *Context:
			m.sessions[key] = m.Spawn(value)
			return m.sessions[key]
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
				m.sessions[key] = m.Find(value, root)
			case "search":
				m.sessions[key] = m.Search(value, root)[0]
			}
			return m.sessions[key]
		case bool:
			spawn = value
		}
	}

	for msg := m; msg != nil; msg = msg.message {
		if x, ok := msg.sessions[key]; ok {
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
	m.callback = cb
	m.Cmd(arg...)
	return m
}

// }}}
func (m *Message) Back(msg *Message) *Message { // {{{
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

// }}}
func (m *Message) CallBack(sync bool, cb func(msg *Message) (sub *Message), arg ...interface{}) *Message { // {{{
	if !sync {
		return m.Call(cb, arg...)
	}

	wait := make(chan bool)

	go m.Call(func(sub *Message) *Message {
		msg := cb(sub)
		m.Log("lock", "before done %v", arg)
		wait <- true
		m.Log("lock", "after done %v", arg)
		return msg
	}, arg...)

	m.Log("lock", "before wait %v", arg)
	<-wait
	m.Log("lock", "after wait %v", arg)
	return m
}

// }}}

func (m *Message) Add(meta string, key string, value ...interface{}) *Message { // {{{
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

// }}}
func (m *Message) Set(meta string, arg ...string) *Message { // {{{
	switch meta {
	case "detail", "result":
		delete(m.Meta, meta)
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

// }}}
func (m *Message) Put(meta string, key string, value interface{}) *Message { // {{{
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

// }}}
func (m *Message) Has(key string) bool { // {{{
	if _, ok := m.Data[key]; ok {
		return true
	}
	if _, ok := m.Meta[key]; ok {
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
	if str = fmt.Sprintf(str, arg...); m.Options("terminal_color") {
		str = fmt.Sprintf("\033[%dm%s\033[0m", color, str)
	}
	return m.Add("result", str)
}

// }}}
func (m *Message) Table(cb func(maps map[string]string, list []string, line int) (goon bool)) *Message { // {{{
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
			case "time":
				ti, e := time.ParseInLocation(m.Conf("time_layout"), table[i][key], time.Local)
				m.Assert(e)
				tj, e := time.ParseInLocation(m.Conf("time_layout"), table[j][key], time.Local)
				m.Assert(e)
				if tj.Before(ti) {
					result = true
				}
			case "time_r":
				ti, e := time.ParseInLocation(m.Conf("time_layout"), table[i][key], time.Local)
				m.Assert(e)
				tj, e := time.ParseInLocation(m.Conf("time_layout"), table[j][key], time.Local)
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

	str := Trans(arg...)

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
func (m *Message) Detail(arg ...interface{}) string { // {{{
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

// }}}
func (m *Message) Detaili(arg ...interface{}) int { // {{{
	i, e := strconv.Atoi(m.Detail(arg...))
	m.Assert(e)
	return i
}

// }}}
func (m *Message) Details(arg ...interface{}) bool { // {{{
	return Right(m.Detail(arg...))
}

// }}}
func (m *Message) Result(arg ...interface{}) string { // {{{
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

// }}}
func (m *Message) Resulti(arg ...interface{}) int { // {{{
	i, e := strconv.Atoi(m.Result(arg...))
	m.Assert(e)
	return i
}

// }}}
func (m *Message) Results(arg ...interface{}) bool { // {{{
	return Right(m.Result(arg...))
}

// }}}

func (m *Message) Option(key string, arg ...interface{}) string { // {{{
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

// }}}
func (m *Message) Optioni(key string, arg ...interface{}) int { // {{{
	i, e := strconv.Atoi(m.Option(key, arg...))
	m.Assert(e)
	return i
}

// }}}
func (m *Message) Options(key string, arg ...interface{}) bool { // {{{
	return Right(m.Option(key, arg...))
}

// }}}
func (m *Message) Optionv(key string, arg ...interface{}) interface{} { // {{{
	if len(arg) > 0 {
		m.Put("option", key, arg[0])
	}

	for msg := m; msg != nil; msg = msg.message {
		if !msg.Has(key) {
			continue
		}
		for _, k := range msg.Meta["option"] {
			if k == key {
				return msg.Data[key]
			}
		}
	}
	return nil
}

// }}}
func (m *Message) Optionx(key string, format string) interface{} { // {{{
	if value := m.Option(key); value != "" {
		return fmt.Sprintf(format, value)
	}
	return ""
}

// }}}
func (m *Message) Append(key string, arg ...interface{}) string { // {{{
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

// }}}
func (m *Message) Appendi(key string, arg ...interface{}) int { // {{{
	i, e := strconv.Atoi(m.Append(key, arg...))
	m.Assert(e)
	return i
}

// }}}
func (m *Message) Appends(key string, arg ...interface{}) bool { // {{{
	return Right(m.Append(key, arg...))
}

// }}}
func (m *Message) Appendv(key string, arg ...interface{}) interface{} { // {{{
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
				return ms[i].Data[key]
			}
		}
	}
	return nil
}

// }}}

func (m *Message) Wait() bool { // {{{
	if m.target.exit != nil {
		return <-m.target.exit
	}
	return true
}

// }}}
func (m *Message) Start(name string, help string, arg ...string) bool { // {{{
	return m.Set("detail", arg...).target.Spawn(m, name, help).Begin(m).Start(m)
}

// }}}
func (m *Message) Cmd(args ...interface{}) *Message { // {{{
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

					x.Hand(m, s, key, arg...)
					m.Hand = true
				})
				return m
			}
		}
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
		value = m.Conf(key)
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

// }}}
func (m *Message) Confs(key string, arg ...bool) bool { // {{{
	if len(arg) > 0 {
		if arg[0] {
			m.Conf(key, "1")
		} else {
			m.Conf(key, "0")
		}
	}

	return Right(m.Conf(key))
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

	for _, c := range []*Context{m.target, m.source} {
		for s := c; s != nil; s = s.context {
			if x, ok := s.Configs[key]; ok {
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
		if m.target.Configs == nil {
			m.target.Configs = make(map[string]*Config)
		}

		m.target.Configs[key] = &Config{Name: arg[0], Value: arg[1], Help: arg[2], Hand: hand}
		m.Log("conf", "%s %v", key, arg)
		return m.Conf(key, arg[1])
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

	return Right(m.Cap(key))
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

// }}}

var CGI = template.FuncMap{
	"meta": func(arg ...interface{}) string { // {{{
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
				m.Sess(which, arg[2:]...)
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
					for k, _ := range m.sessions {
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
					for k, _ := range m.sessions {
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
			return m.Detail(arg[1:]...)
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
			return m.Result(arg[1:]...)
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

var Pulse = &Message{code: 0, time: time.Now(), source: Index, target: Index, Meta: map[string][]string{}}

var Index = &Context{Name: "ctx", Help: "模块中心",
	Caches: map[string]*Cache{
		"nserver":  &Cache{Name: "nserver", Value: "0", Help: "服务数量"},
		"ncontext": &Cache{Name: "ncontext", Value: "0", Help: "模块数量"},
		"nmessage": &Cache{Name: "nmessage", Value: "0", Help: "消息数量"},
	},
	Configs: map[string]*Config{
		"debug": &Config{Name: "debug(on/off)", Value: "on", Help: "调试模式，on:打印，off:不打印)"},

		"search_method": &Config{Name: "search_method(find/search)", Value: "search", Help: "搜索方法, find: 模块名精确匹配, search: 模块名或帮助信息模糊匹配"},
		"search_choice": &Config{Name: "search_choice(first/last/rand/magic)", Value: "magic", Help: "搜索匹配, first: 匹配第一个模块, last: 匹配最后一个模块, rand: 随机选择, magic: 加权选择"},
		"search_action": &Config{Name: "search_action(list/switch)", Value: "switch", Help: "搜索操作, list: 输出模块列表, switch: 模块切换"},
		"search_root":   &Config{Name: "search_root(true/false)", Value: "true", Help: "搜索起点, true: 根模块, false: 当前模块"},

		"detail_index": &Config{Name: "detail_index", Value: "0", Help: "参数的索引"},
		"result_index": &Config{Name: "result_index", Value: "-2", Help: "返回值的索引"},

		"list_help": &Config{Name: "list_help", Value: "list command", Help: "命令列表帮助"},
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
		"message": &Command{Name: "message [code] [all]|[cmd...]", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.Sess("cli", false) // {{{
			if len(arg) > 0 {
				if code, e := strconv.Atoi(arg[0]); e == nil {
					ms := []*Message{m.root}
					for i := 0; i < len(ms); i++ {
						ms = append(ms, ms[i].messages...)
						if ms[i].code == code {
							msg, arg = ms[i], arg[1:]
							break
						}
					}
				}
			}

			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			if len(arg) == 0 {
				m.Echo("%s\n", msg.Format())
				if len(msg.Meta["option"]) > 0 {
					m.Color(31, "option: %v\n", msg.Meta["option"])
					for _, k := range msg.Meta["option"] {
						if v, ok := msg.Data[k]; ok {
							m.Echo(" %s: %v\n", k, v)
						} else {
							m.Echo(" %s: %v\n", k, msg.Meta[k])
						}
					}
				}

				m.Color(31, "result: %v\n", msg.Meta["result"])
				if len(msg.Meta["append"]) > 0 {
					m.Color(31, "append: %v\n", msg.Meta["append"])
					for _, k := range msg.Meta["append"] {
						if v, ok := msg.Data[k]; ok {
							m.Echo(" %s: %v\n", k, v)
						} else {
							m.Echo(" %s: %v\n", k, msg.Meta[k])
						}
					}
				}

				if msg.message != nil {
					m.Color(31, "message:\n")
					m.Echo(" %s\n", msg.message.Format())
				}

				if len(msg.messages) > 0 {
					m.Color(31, "messages:\n")
					for i, v := range msg.messages {
						if !all {
							switch v.target.Name {
							case "log", "yac", "lex":
								continue
							}
						}
						m.Echo(" %d %s\n", i, v.Format())
					}
				}

				if len(msg.sessions) > 0 {
					m.Color(31, "sessions:\n")
					for k, v := range msg.sessions {
						m.Echo(" %s %s\n", k, v.Format())
					}
				}
				return
			}

			switch arg[0] {
			case "message":
				for msg := m; msg != nil; msg = msg.message {
					m.Echo("%d: %s\n", msg.code, msg.Format())
				}
			default:
				sub := msg.Spawn().Cmd(arg)
				m.Copy(sub, "result").Copy(sub, "append")
			}

			// }}}
		}},
		"detail": &Command{Name: "detail [index] [value...]", Help: "查看或添加参数", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.message // {{{
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
			msg := m.message // {{{
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
			msg := m.message // {{{
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
			msg := m.message // {{{
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

			switch arg[0] {
			case "ncol":
				if msg.Meta != nil && len(msg.Meta["append"]) > 0 {
					m.Echo("%d", len(msg.Meta["append"]))
				} else {
					m.Echo("0")
				}
			case "nrow":
				if msg.Meta != nil && len(msg.Meta["append"]) > 0 {
					m.Echo("%d", len(msg.Meta[msg.Meta["append"][0]]))
				} else {
					m.Echo("0")
				}
			}

			m.Echo("%s", msg.Append(arg[0], arg[1:]))
			// }}}
		}},
		"session": &Command{Name: "session [key [cmd...]]", Help: "查看或添加会话", Hand: func(m *Message, c *Context, key string, arg ...string) {
			msg := m.message // {{{
			if len(arg) == 0 {
				for msg = msg; msg != nil; msg = msg.message {
					for k, v := range msg.sessions {
						m.Echo("%s: %d(%s->%s) %v %v\n", k, v.code, v.source.Name, v.target.Name, v.Meta["detail"], v.Meta["option"])
					}
				}
				return
			}

			var sub *Message
			for m := msg; m != nil; m = m.message {
				for k, v := range m.sessions {
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
			msg.sessions[arg[0]] = cmd
			// }}}
		}},
		"callback": &Command{Name: "callback index", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) {
			index := 0
			for msg := m.Sess("cli", false); msg != nil; msg = msg.message { // {{{
				index--
				switch len(arg) {
				case 0:
					if msg.callback == nil {
						return
					}

					m.Echo("%d: %v\n", index, msg.callback)
				case 1:
					if msg.callback == nil {
						return
					}

					if i, e := strconv.Atoi(arg[0]); e == nil && i == index {
						m.Echo("%v", msg.callback)
					}
				default:
					if i, e := strconv.Atoi(arg[0]); e == nil && i == index {
						msg.callback = func(msg *Message) *Message {
							return msg
						}
						m.Echo("%v", msg.callback)
					}
					return
				}
				continue

				if len(arg) == 0 {
					m.Echo("msg(%s->%s): %d(%s) %v\n", msg.source.Name, msg.target.Name, msg.code, msg.time.Format("15:04:05"), msg.Meta["detail"])
					if msg.callback != nil {
						m.Echo("  hand: %v\n", msg.callback)
					}
				} else {
					switch arg[0] {
					case "del":
						msg.message.callback = nil
					case "add":
						msg.message.callback = func(msg *Message) *Message {
							msg.Log("info", "callback default")
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
						method, arg = arg[0], arg[1:]
					}
				}

				root := m.Confs("search_root")
				if len(arg) > 0 {
					switch arg[0] {
					case "root":
						root, arg = true, arg[1:]
					case "home":
						root, arg = false, arg[1:]
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
						choice := m.Conf("search_choice")
						switch arg[0] {
						case "magic", "rand", "first", "last":
							choice, arg = arg[0], arg[1:]
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
							arg = arg[1:]
						}
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
						switch which {
						case "cache":
							keys := []string{}
							for k, _ := range msg.target.Caches {
								keys = append(keys, k)
							}
							sort.Strings(keys)
							for _, k := range keys {
								v := msg.target.Caches[k]
								m.Add("append", "key", k)
								m.Add("append", "name", v.Name)
								m.Add("append", "value", v.Value)
								m.Add("append", "help", v.Help)

							}
						case "config":
							keys := []string{}
							for k, _ := range msg.target.Configs {
								keys = append(keys, k)
							}
							sort.Strings(keys)
							for _, k := range keys {
								v := msg.target.Configs[k]
								m.Add("append", "key", k)
								m.Add("append", "name", v.Name)
								m.Add("append", "value", v.Value)
								m.Add("append", "help", v.Help)
							}
						case "command":
							keys := []string{}
							for k, _ := range msg.target.Commands {
								keys = append(keys, k)
							}
							sort.Strings(keys)
							for _, k := range keys {
								v := msg.target.Commands[k]
								m.Add("append", "key", k)
								m.Add("append", "name", v.Name)
								m.Add("append", "help", v.Help)
							}
						case "module":
							m.Travel(func(msg *Message, i int) bool {
								m.Add("append", "name", msg.target.Name)
								m.Add("append", "help", msg.target.Help)
								m.Add("append", "module", msg.Cap("module"))
								m.Add("append", "status", msg.Cap("status"))
								m.Add("append", "stream", msg.Cap("stream"))
								return true
							}, msg.target)
						case "domain":
							m.Find("ssh", true).Travel(func(msg *Message, i int) bool {
								m.Add("append", "name", msg.target.Name)
								m.Add("append", "help", msg.target.Help)
								m.Add("append", "domain", msg.Cap("domain")+"."+msg.Conf("domains"))
								return true
							})
						default:
							msg.Travel(func(msg *Message, i int) bool {
								target := msg.target
								m.Echo("%s(", target.Name)

								if target.context != nil {
									m.Echo("%s", target.context.Name)
								}
								m.Echo(":")

								msg.target = target

								m.Echo("%s(%s) ", msg.Cap("status"), msg.Cap("stream"))
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
					m.Travel(func(m *Message, i int) bool {
						if m.Cap("status") == "start" {
							m.Echo("%s(%s): %s\n", m.target.Name, m.Cap("stream"), m.target.Help)
						}
						return true
					}, m.target.root)

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
			Name: "command [all|add cmd arg...|list [begin [end]]|test [begin [end]]|delete cmd]",
			Help: "查看或修改命令",
			Form: map[string]int{"condition": -1, "list_help": 1},
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
						Help: m.Confx("list_help"),
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
							m.Echo("%d(%s): %s\n", i, c.Help, c.Name)
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
				aaa := m.Sess("aaa")
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
}

func Start(args ...string) {
	Index.root = Index
	Pulse.root = Pulse

	for _, m := range Pulse.Search("") {
		m.target.root = Index
		m.target.Begin(m)
	}

	Pulse.Sess("nfs", "nfs")
	Pulse.Sess("lex", "lex")
	Pulse.Sess("yac", "yac")
	Pulse.Sess("cli", "cli")
	Pulse.Sess("aaa", "aaa")
	Pulse.Sess("log", "log")

	if len(args) > 0 {
		Pulse.Sess("cli", false).Conf("init.shy", args[0])
		args = args[1:]
	}
	if len(args) > 0 {
		Pulse.Sess("log", false).Conf("bench.log", args[0])
		args = args[1:]
	}

	Pulse.Options("log", true)
	log := Pulse.Sess("log", false)
	log.target.Start(log)

	Pulse.Options("terminal_color", true)
	Pulse.Sess("cli", false).Cmd("source", "stdio", "async").Wait()
}
