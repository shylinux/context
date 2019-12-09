package ctx

import (
	"fmt"
	_ "github.com/shylinux/icebergs"
	"strconv"
	"strings"
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
	Auto func(m *Message, c *Context, key string, arg ...string) (ok bool)
	Hand func(m *Message, c *Context, key string, arg ...string) (e error)
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
type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
}

func (c *Context) Context() *Context {
	return c.context
}
func (c *Context) Message() *Message {
	return c.message
}

type Message struct {
	time time.Time
	code int

	source *Context
	target *Context

	Hand bool
	Meta map[string][]string
	Data map[string]interface{}
	Sync chan bool

	callback func(msg *Message) (sub *Message)
	freeback []func(msg *Message) (done bool)
	Sessions map[string]*Message

	messages []*Message
	message  *Message
	root     *Message
}
type LOGGER interface {
	Log(*Message, string, string, ...interface{})
}
type DEBUG interface {
	Wait(*Message, ...interface{}) interface{}
	Goon(interface{}, ...interface{})
}

func (m *Message) Time(arg ...interface{}) string {
	t := m.time
	if len(arg) > 0 {
		if d, e := time.ParseDuration(arg[0].(string)); e == nil {
			arg, t = arg[1:], t.Add(d)
		}
	}

	str := m.Conf("time", "format")
	if len(arg) > 1 {
		str = fmt.Sprintf(arg[0].(string), arg[1:]...)
	} else if len(arg) > 0 {
		str = fmt.Sprintf("%v", arg[0])
	}

	if str == "stamp" {
		return kit.Format(t.Unix())
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

func (m *Message) Insert(meta string, index int, arg ...interface{}) string {
	if m == nil {
		return ""
	}
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
func (m *Message) Option(key string, arg ...interface{}) string {
	if m == nil {
		return ""
	}
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
func (m *Message) Appendi(key string, arg ...interface{}) int64 {
	i, _ := strconv.ParseInt(m.Append(key, arg...), 10, 64)
	return i
}
func (m *Message) Appends(key string, arg ...interface{}) bool {
	return kit.Right(m.Append(key, arg...))
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

func (m *Message) Push(key interface{}, arg ...interface{}) *Message {
	keys := []string{}
	switch key := key.(type) {
	case string:
		keys = strings.Split(key, " ")
	case []string:
		keys = key
	}

	for _, key := range keys {
		switch m.Option("table.format") {
		case "table":
			m.Add("append", "key", key)
			key = "value"
		}
		switch value := arg[0].(type) {
		case map[string]interface{}:
			m.Add("append", key, kit.Select(" ", kit.Format(kit.Chain(value, key))))
		default:
			m.Add("append", key, arg...)
		}
	}
	return m
}
func (m *Message) Sort(key string, arg ...string) *Message {
	cmp := "str"
	if len(arg) > 0 && arg[0] != "" {
		cmp = arg[0]
	} else {
		cmp = "int"
		for _, v := range m.Meta[key] {
			if _, e := strconv.Atoi(v); e != nil {
				cmp = "str"
			}
		}
	}

	number := map[int]int{}
	table := []map[string]string{}
	m.Table(func(index int, line map[string]string) {
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
	})

	for i := 0; i < len(table)-1; i++ {
		for j := i + 1; j < len(table); j++ {
			result := false
			switch cmp {
			case "", "str":
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
func (m *Message) Split(str string, arg ...string) *Message {
	c := rune(kit.Select(" ", arg, 0)[0])
	lines := strings.Split(str, "\n")

	pos := []int{}
	heads := []string{}
	if h := kit.Select("", arg, 2); h != "" {
		heads = strings.Split(h, " ")
	} else {
		h, lines = lines[0], lines[1:]
		v := kit.Trans(m.Optionv("cmd_headers"))
		for i := 0; i < len(v)-1; i += 2 {
			h = strings.Replace(h, v[i], v[i+1], 1)
		}

		heads = kit.Split(h, c, kit.Int(kit.Select("-1", arg, 1)))
		for _, v := range heads {
			pos = append(pos, strings.Index(h, v))
		}
	}

	for _, l := range lines {
		if len(l) == 0 {
			continue
		}
		if len(pos) > 0 {
			for i, v := range pos {
				if v < len(l) && i == len(pos)-1 {
					m.Add("append", heads[i], strings.TrimSpace(l[v:]))
				} else if v < len(l) && i+1 < len(pos) && pos[i+1] < len(l) {
					m.Add("append", heads[i], strings.TrimSpace(l[v:pos[i+1]]))
				} else {
					m.Add("append", heads[i], "")
				}
			}
			continue
		}
		ls := kit.Split(l, c, len(heads))
		for i, v := range heads {
			m.Add("append", v, kit.Select("", ls, i))
		}
	}
	m.Table()
	return m
}
func (m *Message) Limit(offset, limit int) *Message {
	l := len(m.Meta[m.Meta["append"][0]])
	if offset < 0 {
		offset = 0
	}
	if offset > l {
		offset = l
	}
	if offset+limit > l {
		limit = l - offset
	}
	for _, k := range m.Meta["append"] {
		m.Meta[k] = m.Meta[k][offset : offset+limit]
	}
	return m
}
func (m *Message) Filter(value string) *Message {

	return m
}
func (m *Message) Group(method string, args ...string) *Message {

	nrow := len(m.Meta[m.Meta["append"][0]])

	keys := map[string]bool{}
	for _, v := range args {
		keys[v] = true
	}

	counts := []int{}
	mis := map[int]bool{}
	for i := 0; i < nrow; i++ {
		counts = append(counts, 1)
		if mis[i] {
			continue
		}
	next:
		for j := i + 1; j < nrow; j++ {
			if mis[j] {
				continue
			}
			for key := range keys {
				if m.Meta[key][i] != m.Meta[key][j] {
					continue next
				}
			}
			for _, k := range m.Meta["append"] {
				if !keys[k] {
					switch method {
					case "sum", "avg":
						v1, e1 := strconv.Atoi(m.Meta[k][i])
						v2, e2 := strconv.Atoi(m.Meta[k][j])
						if e1 == nil && e2 == nil {
							m.Meta[k][i] = fmt.Sprintf("%d", v1+v2)
						}
					}
				}
			}
			mis[j] = true
			counts[i]++
		}
	}

	for i := 0; i < nrow; i++ {
		for _, k := range m.Meta["append"] {
			if !keys[k] {
				switch method {
				case "avg":
					if v1, e1 := strconv.Atoi(m.Meta[k][i]); e1 == nil {
						m.Meta[k][i] = strconv.Itoa(v1 / counts[i])
					}
				}
			}
		}
	}
	for i := 0; i < nrow; i++ {
		m.Push("_counts", counts[i])
	}

	for i := 0; i < nrow; i++ {
		if mis[i] {
			for j := i + 1; j < nrow; j++ {
				if !mis[j] {
					for _, k := range m.Meta["append"] {
						m.Meta[k][i] = m.Meta[k][j]
					}
					mis[i], mis[j] = false, true
					break
				}
			}
		}
		if mis[i] {
			for _, k := range m.Meta["append"] {
				m.Meta[k] = m.Meta[k][0:i]
			}
			break
		}
	}
	return m
}
func (m *Message) Table(cbs ...interface{}) *Message {
	if len(m.Meta["append"]) == 0 {
		return m
	}

	// 遍历函数
	if len(cbs) > 0 {
		nrow := len(m.Meta[m.Meta["append"][0]])
		for i := 0; i < nrow; i++ {
			line := map[string]string{}
			for _, k := range m.Meta["append"] {
				line[k] = kit.Select("", m.Meta[k], i)
			}

			switch cb := cbs[0].(type) {
			case func(map[string]string):
				cb(line)
			case func(map[string]string) bool:
				if !cb(line) {
					return m
				}
			case func(int, map[string]string):
				cb(i, line)
			}
		}
		return m
	}

	//计算列宽
	space := kit.Select(m.Conf("table", "space"), m.Option("table.space"))
	depth, width := 0, map[string]int{}
	for _, k := range m.Meta["append"] {
		if len(m.Meta[k]) > depth {
			depth = len(m.Meta[k])
		}
		width[k] = kit.Width(k, len(space))
		for _, v := range m.Meta[k] {
			if kit.Width(v, len(space)) > width[k] {
				width[k] = kit.Width(v, len(space))
			}
		}
	}

	// 回调函数
	rows := kit.Select(m.Conf("table", "row_sep"), m.Option("table.row_sep"))
	cols := kit.Select(m.Conf("table", "col_sep"), m.Option("table.col_sep"))
	compact := kit.Right(kit.Select(m.Conf("table", "compact"), m.Option("table.compact")))
	cb := func(maps map[string]string, lists []string, line int) bool {
		for i, v := range lists {
			if k := m.Meta["append"][i]; compact {
				v = maps[k]
			}

			if m.Echo(v); i < len(lists)-1 {
				m.Echo(cols)
			}
		}
		m.Echo(rows)
		return true
	}

	// 输出表头
	row := map[string]string{}
	wor := []string{}
	for _, k := range m.Meta["append"] {
		row[k], wor = k, append(wor, k+strings.Repeat(space, width[k]-kit.Width(k, len(space))))
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

			row[k], wor = data, append(wor, data+strings.Repeat(space, width[k]-kit.Width(data, len(space))))
		}
		if !cb(row, wor, i) {
			break
		}
	}

	return m
}
func (m *Message) Copy(msg *Message, arg ...string) *Message {
	if msg == nil || m == msg {
		return m
	}

	for i := 0; i < len(arg); i++ {
		meta := arg[i]

		switch meta {
		case "target":
			m.target = msg.target
		case "callback":
			m.callback = msg.callback
		case "detail", "result":
			if len(msg.Meta[meta]) > 0 {
				m.Add(meta, msg.Meta[meta][0], msg.Meta[meta][1:])
			}
		case "option", "append":
			if msg.Meta == nil {
				msg.Meta = map[string][]string{}
			}
			if msg.Meta[meta] == nil {
				break
			}
			if i == len(arg)-1 {
				arg = append(arg, msg.Meta[meta]...)
			}

			for i++; i < len(arg); i++ {
				if v, ok := msg.Data[arg[i]]; ok {
					m.Put(meta, arg[i], v)
				} else if v, ok := msg.Meta[arg[i]]; ok {
					m.Set(meta, arg[i], v) // TODO fuck Add
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

func (m *Message) Cmdp(t time.Duration, head []string, prefix []string, suffix [][]string) *Message {
	if head != nil && len(head) > 0 {
		m.Show(fmt.Sprintf("[%s]...\n", strings.Join(head, " ")))
	}

	for i, v := range suffix {
		m.Show(fmt.Sprintf("%v/%v %v...\n", i+1, len(suffix), v))
		m.CopyFuck(m.Cmd(prefix, v), "append")
		time.Sleep(t)
	}
	m.Show("\n")
	m.Table()
	return m
}
func (m *Message) Cmdy(args ...interface{}) *Message {
	m.Cmd(args...).CopyTo(m)
	return m
}
func (m *Message) Cmdx(args ...interface{}) string {
	msg := m.Cmd(args...)
	if msg.Result(0) == "error: " {
		return msg.Result(1)
	}
	return msg.Result(0)
}
func (m *Message) Cmds(args ...interface{}) bool {
	return kit.Right(m.Cmdx(args...))
}
func (m *Message) Cmd(args ...interface{}) *Message {
	if m == nil {
		return m
	}

	if len(args) > 0 {
		m.Set("detail", kit.Trans(args...))
	}
	key, arg := m.Meta["detail"][0], m.Meta["detail"][1:]
	if key == "_" {
		return m
	}

	msg := m
	if strings.Contains(key, ":") {
		ps := strings.Split(key, ":")
		if ps[0] == "_" {
			ps[0], arg = arg[0], arg[1:]
		}
		msg, key, arg = m.Sess("ssh"), "_route", append([]string{"sync", ps[0], ps[1]}, arg...)
		defer func() { m.Copy(msg, "append").Copy(msg, "result") }()
		m.Hand = true

	} else if strings.Contains(key, ".") {
		arg := strings.Split(key, ".")
		if msg, key = m.Sess(arg[0]), arg[1]; len(arg) == 2 && msg != nil {
			msg.Option("remote_code", "")

		} else if msg, key = m.Find(strings.Join(arg[0:len(arg)-1], "."), true), arg[len(arg)-1]; msg != nil {
			msg.Option("remote_code", "")

		}
	}
	if msg == nil {
		return msg
	}

	msg = msg.Match(key, true, func(msg *Message, s *Context, c *Context, key string) bool {
		msg.Hand = false
		if x, ok := c.Commands[key]; ok && x.Hand != nil {
			msg.TryCatch(msg, true, func(msg *Message) {
				msg.Log("cmd", "%s %s %v %v", c.Name, key, arg, msg.Meta["option"])
				msg.Hand = true
				x.Hand(msg, c, key, msg.Form(x, arg)...)
			})
		}
		return msg.Hand
	})
	return msg
}

func (m *Message) Confm(key string, args ...interface{}) map[string]interface{} {
	random := ""

	var chain interface{}
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case []interface{}:
			chain, args = arg, args[1:]
		case []string:
			chain, args = arg, args[1:]
		case string:
			switch arg {
			case "%", "*":
				random, args = arg, args[1:]
			default:
				chain, args = arg, args[1:]
			}
		}
	}

	var v interface{}
	if chain == nil {
		v = m.Confv(key)
	} else {
		v = m.Confv(key, chain)
	}
	return kit.Map(v, random, args...)
}
func (m *Message) Confx(key string, args ...interface{}) string {
	value := kit.Select(m.Conf(key), m.Option(key))
	if len(args) == 0 {
		return value
	}

	switch arg := args[0].(type) {
	case []string:
		if len(args) > 1 {
			value = kit.Select(value, arg, args[1])
		} else {
			value = kit.Select(value, arg)
		}
		args = args[1:]
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
		arg = append(arg, v)
	}

	return kit.Format(arg...)
}
func (m *Message) Confv(key string, args ...interface{}) interface{} {
	if strings.Contains(key, ".") {
		target := m.target
		defer func() { m.target = target }()

		ps := strings.Split(key, ".")
		if msg := m.Sess(ps[0], false); msg != nil {
			m.target, key = msg.target, ps[1]
		}
	}

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
func (m *Message) Confs(key string, arg ...interface{}) bool {
	return kit.Right(m.Confv(key, arg...))
}
func (m *Message) Confi(key string, arg ...interface{}) int {
	return kit.Int(m.Confv(key, arg...))
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
