package ctx

import (
	"html/template"

	"io"
	"path"
	"strings"
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
				return m.Optionv(value)
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
	"option": func(arg ...interface{}) interface{} {
		return index("option", arg...)
	},
	"options": func(arg ...interface{}) string {
		switch value := index("option", arg...).(type) {
		case string:
			return value
		case []string:
			return strings.Join(value, "")
		}
		return ""
	},
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
