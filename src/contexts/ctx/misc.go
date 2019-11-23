package ctx

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
	"toolkit"
)

func (m *Message) Log(action string, str string, arg ...interface{}) *Message {
	if action == "error" {
		kit.Log("error", fmt.Sprintf("chain: %s", m.Format("chain")))
		kit.Log("error", fmt.Sprintf("%s %s %s", m.Format(), action, fmt.Sprintf(str, arg...)))
		kit.Log("error", fmt.Sprintf("stack: %s", m.Format("stack")))
	}

	// if m.Options("log.disable") {
	// 	return m
	// }

	if l := m.Sess("log", false); l != nil {
		if log, ok := l.target.Server.(LOGGER); ok {
			if action == "error" {
				log.Log(m, "error", "chain: %s", m.Format("chain"))
			}
			if log.Log(m, action, str, arg...); action == "error" {
				log.Log(m, "error", "stack: %s", m.Format("stack"))
			}
			return m
		}
	} else {
		fmt.Fprintf(os.Stderr, str, arg...)
	}

	return m
}
func (m *Message) Gdb(arg ...interface{}) interface{} {
	// if !m.Options("log.enable") {
	// 	return ""
	// }

	if g := m.Sess("gdb", false); g != nil {
		if gdb, ok := g.target.Server.(DEBUG); ok {
			return gdb.Wait(m, arg...)
		}
	}
	return ""
}
func (m *Message) Show(str string, args ...interface{}) *Message {
	res := fmt.Sprintf(str, args...)

	if m.Option("bio.modal") == "action" {
		fmt.Fprintf(os.Stderr, res)
	} else if kit.STDIO != nil {
		kit.STDIO.Show(res)
	} else {
		m.Log("info", "show: %v", res)
	}
	return m
}
func (m *Message) Format(arg ...interface{}) string {
	if len(arg) == 0 {
		arg = append(arg, "time", "ship")
	}

	meta := []string{}
	for _, v := range arg {
		switch kit.Format(v) {
		case "cost":
			meta = append(meta, kit.FmtTime(time.Now().Sub(m.time).Nanoseconds()))

		case "summary":
			msg := arg[1].(*Message)
			ms := make([]*Message, 0, 1024)
			ms = append(ms, msg.message, msg)

			for i := 0; i < len(ms); i++ {
				msg := ms[i]
				if m.Add("append", "index", i); msg == nil {
					m.Add("append", "message", "")
					m.Add("append", "time", "")
					m.Add("append", "code", "")
					m.Add("append", "source", "")
					m.Add("append", "target", "")
					m.Add("append", "details", "")
					m.Add("append", "options", "")
					continue
				}

				if msg.message != nil {
					m.Add("append", "message", msg.message.code)
				} else {
					m.Add("append", "message", "")
				}
				m.Add("append", "time", msg.time.Format("15:04:05"))
				m.Add("append", "code", msg.code)
				m.Add("append", "source", msg.source.Name)
				m.Add("append", "target", msg.target.Name)
				m.Add("append", "details", fmt.Sprintf("%v", msg.Meta["detail"]))
				m.Add("append", "options", fmt.Sprintf("%v", msg.Meta["option"]))

				if i == 0 {
					continue
				}

				if len(ms) < 30 && len(arg) > 2 && arg[2] == "deep" {
					ms = append(ms, ms[i].messages...)
				}
			}
			m.Table()
		case "time":
			meta = append(meta, m.Time())
		case "mill":
			meta = append(meta, fmt.Sprintf("%d", int64(m.time.UnixNano())/1000%1000000))
		case "code":
			meta = append(meta, kit.Format(m.code))
		case "ship":
			meta = append(meta, fmt.Sprintf("%s:%d(%s->%s)", m.Option("ctx.routine"), m.code, m.source.Name, m.target.Name))
		case "source":
			target := m.target
			m.target = m.source
			meta = append(meta, m.Cap("module"))
			m.target = target
		case "target":
			meta = append(meta, m.Cap("module"))

		case "detail":
			meta = append(meta, fmt.Sprintf("%v", m.Meta["detail"]))
		case "option":
			meta = append(meta, fmt.Sprintf("%v", m.Meta["option"]))
		case "append":
			meta = append(meta, fmt.Sprintf("%v", m.Meta["append"]))
		case "result":
			meta = append(meta, fmt.Sprintf("%v", m.Meta["result"]))

		case "full":
		case "chain":
			ms := []*Message{}
			if v == "full" {
				ms = append(ms, m)
			} else {
				for msg := m; msg != nil; msg = msg.message {
					ms = append(ms, msg)
				}
			}

			meta = append(meta, "\n")
			for i := len(ms) - 1; i >= 0; i-- {
				msg := ms[i]

				meta = append(meta, fmt.Sprintf("%s", msg.Format("time", "ship")))
				if len(msg.Meta["detail"]) > 0 {
					meta = append(meta, fmt.Sprintf("detail:%d %v", len(msg.Meta["detail"]), msg.Meta["detail"]))
				}

				if len(msg.Meta["option"]) > 0 {
					meta = append(meta, fmt.Sprintf("option:%d %v\n", len(msg.Meta["option"]), msg.Meta["option"]))
					for _, k := range msg.Meta["option"] {
						if v, ok := msg.Meta[k]; ok {
							meta = append(meta, fmt.Sprintf("    %s: %d %v\n", k, len(v), v))
						}
					}
				} else {
					meta = append(meta, "\n")
				}

				if len(msg.Meta["append"]) > 0 {
					meta = append(meta, fmt.Sprintf("  append:%d %v\n", len(msg.Meta["append"]), msg.Meta["append"]))
					for _, k := range msg.Meta["append"] {
						if v, ok := msg.Meta[k]; ok {
							meta = append(meta, fmt.Sprintf("    %s: %d %v\n", k, len(v), v))
						}
					}
				}
				if len(msg.Meta["result"]) > 0 {
					meta = append(meta, fmt.Sprintf("  result:%d %v\n", len(msg.Meta["result"]), msg.Meta["result"]))
				}
			}
		case "stack":
			pc := make([]uintptr, 100)
			pc = pc[:runtime.Callers(5, pc)]
			frames := runtime.CallersFrames(pc)

			for {
				frame, more := frames.Next()
				file := strings.Split(frame.File, "/")
				name := strings.Split(frame.Function, "/")
				meta = append(meta, fmt.Sprintf("\n%s:%d\t%s", file[len(file)-1], frame.Line, name[len(name)-1]))
				if !more {
					break
				}
			}

		default:
			meta = append(meta, kit.Format(v))
		}
	}
	return strings.Join(meta, " ")
}
func (m *Message) Short(arg ...string) {
	for _, k := range arg {
		if v := m.Option(k); v != "" && len(v) != 32 {
			m.Option(k, m.Cmdx("aaa.short", v))
		}
	}
}
func (m *Message) Err(str string, arg ...interface{}) {
	m.Echo("").Echo("error: ").Echo(str, arg...)
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
func (m *Message) Find(name string, root ...bool) *Message {
	if name == "" {
		return m.Spawn()
	}
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
			return nil
		}
	}

	if len(root) > 1 && root[1] {
		m.target = target
		return m
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
func (c *Context) Travel(m *Message, hand func(m *Message, n int) (stop bool)) *Context {
	if c == nil {
		return nil
	}
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
			for _, k := range arg {
				delete(m.Data, kit.Format(k))
				delete(m.Meta, kit.Format(k))
			}
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
		if index >= 0 && index < len(meta) {
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
	msg.Copy(m, "append").Copy(m, "result")
	return m
}
func (m *Message) CopyFuck(msg *Message, arg ...string) *Message {
	if m == msg {
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
					m.Add(meta, arg[i], v) // TODO fuck Add
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
func (m *Message) ToHTML(style string) string {
	cmd := strings.Join(m.Meta["detail"], " ")
	result := []string{}
	if len(m.Meta["append"]) > 0 {
		result = append(result, fmt.Sprintf("<table class='%s'>", style))
		result = append(result, "<caption>", cmd, "</caption>")
		m.Table(func(line int, maps map[string]string) {
			if line == 0 {
				result = append(result, "<tr>")
				for _, v := range m.Meta["append"] {
					result = append(result, "<th>", v, "</th>")
				}
				result = append(result, "</tr>")
				return
			}
			result = append(result, "<tr>")
			for _, k := range m.Meta["append"] {
				result = append(result, "<td>", maps[k], "</td>")
			}
			result = append(result, "</tr>")
		})
		result = append(result, "</table>")
	} else {
		result = append(result, "<pre><code>")
		result = append(result, fmt.Sprintf("%s", m.Find("shy", false).Conf("prompt")), cmd, "\n")
		result = append(result, m.Meta["result"]...)
		result = append(result, "</code></pre>")
	}
	return strings.Join(result, "")
}

func (m *Message) Grow(key string, args interface{}, data interface{}) interface{} {
	cache := m.Confm(key, args)
	if args == nil {
		cache = m.Confm(key)
	}
	if cache == nil {
		cache = map[string]interface{}{}
	}
	meta, ok := cache["meta"].(map[string]interface{})
	if !ok {
		meta = map[string]interface{}{}
	}
	list, _ := cache["list"].([]interface{})

	list = append(list, data)
	if len(list) > kit.Int(kit.Select(m.Conf("cache", "limit"), meta["limit"])) {
		offset := kit.Int(meta["offset"])
		least := kit.Int(kit.Select(m.Conf("cache", "least"), meta["least"]))

		name := kit.Select(m.Option("cache.store"), meta["store"])
		f, e := os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if e != nil {
			f, _, e = kit.Create(name)
		}
		defer f.Close()
		s, e := f.Stat()
		m.Assert(e)

		keys := []string{}
		w := csv.NewWriter(f)
		if s.Size() == 0 {
			for k := range list[0].(map[string]interface{}) {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			w.Write(keys)
			w.Flush()
			s, e = f.Stat()
		} else {
			r := csv.NewReader(f)
			keys, e = r.Read()
		}

		count := len(list) - least
		record, _ := meta["record"].([]interface{})
		meta["record"] = append(record, map[string]interface{}{
			"time":     m.Time(),
			"offset":   offset,
			"position": s.Size(),
			"count":    count,
			"file":     name,
		})

		for i, v := range list {
			if i >= count {
				break
			}

			val := v.(map[string]interface{})

			values := []string{}
			for _, k := range keys {
				values = append(values, kit.Format(val[k]))
			}
			w.Write(values)

			if i < least {
				list[i] = list[count+i]
			}
		}

		m.Log("info", "save %s offset %v+%v", name, offset, count)
		meta["offset"] = offset + count
		list = list[:least]
		w.Flush()
	}
	cache["meta"] = meta
	cache["list"] = list
	if args == nil {
		m.Conf(key, cache)
	} else {
		m.Conf(key, args, cache)
	}
	return list
}
func (m *Message) Grows(key string, args interface{}, cb interface{}) map[string]interface{} {
	cache := m.Confm(key, args)
	if args == nil {
		cache = m.Confm(key)
	}
	if cache == nil {
		return nil
	}
	meta, ok := cache["meta"].(map[string]interface{})
	if !ok {
		return nil
	}
	list, ok := cache["list"].([]interface{})
	if !ok {
		return nil
	}

	offend := kit.Int(kit.Select("0", m.Option("cache.offend")))
	limit := kit.Int(kit.Select("10", m.Option("cache.limit")))
	match := kit.Select("", m.Option("cache.match"))
	value := kit.Select("", m.Option("cache.value"))
	current := kit.Int(meta["offset"])
	end := current + len(list) - offend
	begin := end - limit

	data := make([]interface{}, 0, limit)
	m.Log("info", "read %v-%v from %v-%v", begin, end, current, current+len(list))
	if begin < current {
		store, _ := meta["record"].([]interface{})
		for s := len(store) - 1; s > -1; s-- {
			item, _ := store[s].(map[string]interface{})
			line := kit.Int(item["offset"])
			m.Log("info", "check history %v %v %v", s, line, item)
			if begin < line && s > 0 {
				continue
			}

			for ; s < len(store); s++ {
				if begin >= end {
					break
				}
				item, _ := store[s].(map[string]interface{})
				if line+kit.Int(item["count"]) < begin {
					continue
				}

				name := kit.Format(item["file"])
				pos := kit.Int(item["position"])
				line := kit.Int(item["offset"])
				m.Log("info", "load history %v %v %v", s, line, item)
				if f, e := os.Open(name); m.Assert(e) {
					defer f.Close()
					r := csv.NewReader(f)
					heads, _ := r.Read()
					m.Log("info", "load head %v", heads)

					f.Seek(int64(pos), os.SEEK_SET)
					r = csv.NewReader(f)
					for i := line; i < end; i++ {
						lines, e := r.Read()
						if e != nil {
							break
						}

						if i >= begin {
							item := map[string]interface{}{}
							for i := range heads {
								item[heads[i]] = lines[i]
							}
							m.Log("info", "load line %v %v %v", i, len(data), item)
							if match == "" || strings.Contains(kit.Format(item[match]), value) {
								data = append(data, item)
							}
							begin = i + 1
						} else {
							m.Log("info", "skip line %v", i)
						}
					}
				}
			}
			break
		}
	}

	if begin < current {
		begin = current
	}
	m.Log("info", "cache %v-%v", begin-current, end-current)
	for i := begin - current; i < end-current; i++ {
		if match == "" || strings.Contains(kit.Format(kit.Chain(list[i], match)), value) {
			data = append(data, list[i])
		}
	}
	return kit.Map(map[string]interface{}{"meta": meta, "list": data}, "", cb)
}
