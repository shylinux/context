package ctx

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"toolkit"
)

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
	"cmd": func(m *Message, args ...interface{}) *Message {
		if len(args) == 0 {
			return m
		}

		return m.Sess("cli").Put("option", "bench", "").Cmd("source", args)
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
	"parse": func(m *Message, arg ...interface{}) interface{} {
		switch len(arg) {
		case 1:
			return m.Parse(kit.Format(arg[0]))
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
