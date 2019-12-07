package ctx

import (
	"text/template"

	"bytes"
	"io"
	"path"
	"strings"
	"toolkit"
)

func index(name string, arg ...interface{}) interface{} {
	if len(arg) == 0 {
		return ""
	}

	switch m := arg[0].(type) {
	case *Message:
		if len(arg) == 1 {
			return m.Meta[name]
		}

		switch value := arg[1].(type) {
		case int:
			if 0 <= value && value < len(m.Meta[name]) {
				return m.Meta[name][value]
			}
		case string:
			if len(arg) == 2 {
				if name == "option" {
					return m.Optionv(value)
				} else {
					return m.Append(value)
				}
			}

			switch val := arg[2].(type) {
			case int:
				switch list := m.Optionv(value).(type) {
				case []string:
					if 0 <= val && val < len(list) {
						return list[val]
					}
				case []interface{}:
					if 0 <= val && val < len(list) {
						return list[val]
					}
				}
			}
		}
	case map[string][]string:
		if len(arg) == 1 {
			return m[name]
		}

		switch value := arg[1].(type) {
		case int:
			return m[name][value]
		case string:
			if len(arg) == 2 {
				return m[value]
			}
			switch val := arg[2].(type) {
			case int:
				return m[value][val]
			}
		}
	case []string:
		if len(arg) == 1 {
			return m
		}
		switch value := arg[1].(type) {
		case int:
			return m[value]
		}
	default:
		return m
	}
	return ""
}

var CGI = template.FuncMap{
	"options": func(arg ...interface{}) string {
		switch value := index("option", arg...).(type) {
		case string:
			return value
		case []string:
			return strings.Join(value, "")
		}
		return ""
	},
	"option": func(arg ...interface{}) interface{} {
		return index("option", arg...)
	},
	"conf": func(arg ...interface{}) interface{} {
		switch m := arg[0].(type) {
		case *Message:
			switch c := arg[1].(type) {
			case string:
				if len(arg) == 2 {
					return m.Confv(c)
				}
				if len(arg) == 3 {
					return m.Confv(c, arg[2])
				}
			}
		}
		return nil
	},
	"cmd": func(arg ...interface{}) interface{} {
		switch m := arg[0].(type) {
		case *Message:
			switch c := arg[1].(type) {
			case string:
				return m.Cmd(c, arg[2:])
			}
		}
		return nil
	},
	"appends": func(arg ...interface{}) interface{} {
		switch value := index("append", arg...).(type) {
		case string:
			return value
		case []string:
			return strings.Join(value, "")
		}
		return ""
	},
	"append": func(arg ...interface{}) interface{} {
		switch m := arg[0].(type) {
		case *Message:
			if len(arg) == 1 {
				return m.Meta["append"]
			}
			if len(arg) > 1 {
				switch c := arg[1].(type) {
				case string:
					if len(arg) > 2 {
						switch i := arg[2].(type) {
						case int:
							return kit.Select("", m.Meta[c], i)
						}
					}
					return m.Meta[c]
				}
			}
		}
		return nil
	},
	"trans": func(arg ...interface{}) interface{} {
		switch m := arg[0].(type) {
		case *Message:
			list := [][]string{m.Meta["append"]}
			m.Table(func(index int, value map[string]string) {
				line := []string{}
				for _, k := range m.Meta["append"] {
					line = append(line, value[k])
				}
				list = append(list, line)
			})
			return list
		}
		return nil
	},
	"result": func(arg ...interface{}) interface{} {
		switch m := arg[0].(type) {
		case *Message:
			return m.Meta["result"]
		}
		return nil
	},
	"results": func(arg ...interface{}) interface{} {
		switch m := arg[0].(type) {
		case *Message:
			return strings.Join(m.Meta["result"], "")
		}
		return nil
	},
}

func LocalCGI(m *Message, c *Context) *template.FuncMap {
	cgi := template.FuncMap{
		"format": func(arg ...interface{}) interface{} {
			switch msg := arg[0].(type) {
			case *Message:
				buffer := bytes.NewBuffer([]byte{})
				tmpl := m.Optionv("tmpl").(*template.Template)
				m.Assert(tmpl.ExecuteTemplate(buffer, kit.Select("table", arg, 1), msg))
				return string(buffer.Bytes())
			}
			return nil
		},
	}
	for k, v := range c.Commands {
		if strings.HasPrefix(k, "/") || strings.HasPrefix(k, "_") {
			continue
		}
		func(k string, v *Command) {
			cgi[k] = func(arg ...interface{}) (res interface{}) {
				m.TryCatch(m.Spawn(), true, func(msg *Message) {

					v.Hand(msg, c, k, msg.Form(v, kit.Trans(arg))...)

					buffer := bytes.NewBuffer([]byte{})
					m.Assert(m.Optionv("tmpl").(*template.Template).ExecuteTemplate(buffer,
						kit.Select(kit.Select("code", "table", len(msg.Meta["append"]) > 0), msg.Option("render")), msg))
					res = string(buffer.Bytes())
				})
				return
			}
		}(k, v)
	}
	for k, v := range CGI {
		cgi[k] = v
	}
	return &cgi
}
func ExecuteFile(m *Message, w io.Writer, p string) error {
	tmpl := template.New("render").Funcs(CGI)
	tmpl.ParseGlob(p)
	return tmpl.ExecuteTemplate(w, path.Base(p), m)
}
func ExecuteStr(m *Message, w io.Writer, p string) error {
	tmpl := template.New("render").Funcs(CGI)
	tmpl, _ = tmpl.Parse(p)
	return tmpl.Execute(w, m)
}
func Execute(m *Message, p string) string {
	m.Log("fuck", "waht %v", path.Join(m.Conf("route", "template_dir"), "/*.tmpl"))
	t := template.Must(template.New("render").Funcs(CGI).ParseGlob(path.Join(m.Conf("route", "template_dir"), "/*.tmpl")))
	for _, v := range t.Templates() {
		m.Log("fuck", "waht %v", v.Name())
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	t.ExecuteTemplate(buf, p, m)
	m.Log("fuck", "waht %v", p)
	m.Log("fuck", "waht %v", buf)
	return buf.String()
}
