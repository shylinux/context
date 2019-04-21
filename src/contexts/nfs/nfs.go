package nfs

import (
	"contexts/ctx"
	"toolkit"

	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/skip2/go-qrcode"
	"net/url"
	"regexp"
	"strings"

	"bufio"
	"github.com/nsf/termbox-go"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"time"
)

type NFS struct {
	io  io.ReadWriter
	in  *os.File
	out *os.File

	send chan *ctx.Message
	echo chan *ctx.Message
	hand map[int]*ctx.Message

	*ctx.Context
}

func dir(m *ctx.Message, name string, level int, deep bool, dir_type string, trip int, dir_reg *regexp.Regexp, fields []string, format string) {
	back, e := os.Getwd()
	m.Assert(e)

	if fs, e := ioutil.ReadDir(name); m.Assert(e) {
		for _, f := range fs {
			if f.Name() == "." || f.Name() == ".." {
				continue
			}

			if strings.HasPrefix(f.Name(), ".") && dir_type != "all" {
				continue
			}

			f, e := os.Stat(path.Join(name, f.Name()))
			if e != nil {
				m.Log("info", "%s", e)
				continue
			}

			if !(dir_type == "file" && f.IsDir() || dir_type == "dir" && !f.IsDir()) && (dir_reg == nil || dir_reg.MatchString(f.Name())) {
				for _, field := range fields {
					switch field {
					case "time":
						m.Add("append", "time", f.ModTime().Format(format))
					case "type":
						if m.Assert(e) && f.IsDir() {
							m.Add("append", "type", "dir")
						} else {
							m.Add("append", "type", "file")
						}
					case "full":
						m.Add("append", "full", path.Join(back, name, f.Name()))
					case "path":
						m.Add("append", "path", path.Join(back, name, f.Name())[trip:])
					case "tree":
						if level == 0 {
							m.Add("append", "tree", f.Name())
						} else {
							m.Add("append", "tree", strings.Repeat("| ", level-1)+"|-"+f.Name())
						}
					case "filename":
						if f.IsDir() {
							m.Add("append", "filename", f.Name()+"/")
						} else {
							m.Add("append", "filename", f.Name())
						}
					case "size":
						m.Add("append", "size", f.Size())
					case "line":
						nline := 0
						if f.IsDir() {
							d, e := ioutil.ReadDir(path.Join(name, f.Name()))
							m.Assert(e)
							nline = len(d)
						} else {
							f, e := os.Open(path.Join(name, f.Name()))
							m.Assert(e)
							defer f.Close()

							bio := bufio.NewScanner(f)
							for bio.Scan() {
								bio.Text()
								nline++
							}
						}
						m.Add("append", "line", nline)
					case "hash":
						if f.IsDir() {
							d, e := ioutil.ReadDir(path.Join(name, f.Name()))
							m.Assert(e)
							meta := []string{}
							for _, v := range d {
								meta = append(meta, fmt.Sprintf("%s%d%s", v.Name(), v.Size(), v.ModTime()))
							}
							sort.Strings(meta)

							h := sha1.Sum([]byte(strings.Join(meta, "")))
							m.Add("append", "hash", hex.EncodeToString(h[:]))
							break
						}

						f, e := ioutil.ReadFile(path.Join(name, f.Name()))
						m.Assert(e)
						h := sha1.Sum(f)
						m.Add("append", "hash", hex.EncodeToString(h[:]))
					}
				}
			}
			if f.IsDir() && deep {
				dir(m, path.Join(name, f.Name()), level+1, deep, dir_type, trip, dir_reg, fields, format)
			}
		}
	}
}
func open(m *ctx.Message, name string, arg ...int) (string, *os.File, error) {
	if !path.IsAbs(name) {
		paths := m.Confv("paths").([]interface{})
		for _, v := range paths {
			p := path.Join(v.(string), name)
			if len(arg) > 0 {
				name = p
				break
			}
			if s, e := os.Stat(p); e == nil && !s.IsDir() {
				name = p
				break
			}
		}
	}

	flag := os.O_RDONLY
	if len(arg) > 0 {
		flag = arg[0]
	}

	f, e := os.OpenFile(name, flag, 0660)
	if e == nil {
		m.Log("info", "open %s", name)
		return name, f, e
	}
	m.Log("warn", "%v", e)
	return name, f, e
}

func (nfs *NFS) Read(p []byte) (n int, err error) {
	m := nfs.Context.Message()
	if !m.Caps("termbox") {
		return nfs.in.Read(p)
	}

	m.TryCatch(m, true, func(m *ctx.Message) {
		scroll_count := 0
		scroll_lines := m.Confi("term", "scroll_lines")
		which := m.Capi("ninput")
		what := make([]rune, 0, 1024)
		rest := make([]rune, 0, 1024)
		back := make([]rune, 0, 1024)

		m.Optionv("auto_target", m.Optionv("ps_target"))
		m.Option("auto_cmd", "")
		m.Options("show_shadow", m.Confs("show_shadow"))

		defer func() { m.Option("auto_cmd", "") }()

		frame, table, index, pick := map[string]interface{}{}, []map[string]string{}, 0, 0
		if change, f, t, i := nfs.Auto(what, ":", 0); change {
			frame, table, index, pick = f, t, i, 0
		}

		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventInterrupt:
			case termbox.EventResize:
				nfs.Term(m, "resize")
			case termbox.EventMouse:
				switch ev.Key {
				case termbox.MouseLeft:
					if m.Confs("term", "mouse.resize") {
						nfs.Term(m, "window", ev.MouseX, ev.MouseY)
						nfs.prompt(what).shadow(rest)
					}

				case termbox.MouseMiddle:
				case termbox.MouseRight:
					if m.Confs("term", "mouse.resize") {
						nfs.Term(m, "resize", ev.MouseX, ev.MouseY)
					}
				case termbox.MouseRelease:
				case termbox.MouseWheelUp:
					if scroll_count++; scroll_count > m.Confi("term", "scroll_count") {
						nfs.Term(m, "scroll", -1).Term(m, "flush")
						scroll_count = 0
					}
				case termbox.MouseWheelDown:
					if scroll_count++; scroll_count > m.Confi("term", "scroll_count") {
						nfs.Term(m, "scroll", 1).Term(m, "flush")
						scroll_count = 0
					}
				}
			case termbox.EventError:
			case termbox.EventNone:
			case termbox.EventRaw:
			case termbox.EventKey:

				switch ev.Key {
				case termbox.KeyCtrlP:
					if which--; which < 0 {
						which = m.Capi("ninput") - 1
					}
					if v := m.Conf("input", []interface{}{which, "line"}); v != "" {
						what, rest = append(what[:0], []rune(v)...), rest[:0]
						nfs.prompt(what).shadow(rest)
					}
				case termbox.KeyCtrlN:
					if which++; which >= m.Capi("ninput") {
						which = 0
					}
					if v := m.Conf("input", []interface{}{which, "line"}); v != "" {
						what, rest = append(what[:0], []rune(v)...), rest[:0]
						nfs.prompt(what).shadow(rest)
					}

				case termbox.KeyCtrlA:
					if len(what) > 0 {
						what, rest = append(what, rest...), rest[:0]
						rest, what = append(rest, what...), what[:0]
						nfs.prompt(what).shadow(rest)
					}
				case termbox.KeyCtrlF:
					if len(rest) > 0 {
						pos := len(what) + 1
						what, rest = append(what, rest...), rest[:0]
						rest, what = append(rest, what[pos:]...), what[:pos]
						nfs.prompt(what).shadow(rest)
					}
				case termbox.KeyCtrlB:
					if len(what) > 0 {
						pos := len(what) - 1
						what, rest = append(what, rest...), rest[:0]
						rest, what = append(rest, what[pos:]...), what[:pos]
						nfs.prompt(what).shadow(rest)
					}
				case termbox.KeyCtrlE:
					if len(rest) > 0 {
						what, rest = append(what, rest...), rest[:0]
						nfs.prompt(what).shadow(rest)
					}

				case termbox.KeyCtrlU:
					back = back[:0]
					back, what = append(back, what...), what[:0]
					nfs.prompt(what).shadow(rest)
				case termbox.KeyCtrlD: // termbox.KeyBackspace
					if len(rest) > 0 {
						pos := len(what)
						what, rest = append(what, rest[1:]...), rest[:0]
						rest, what = append(rest, what[pos:]...), what[:pos]
						nfs.prompt(what).shadow(rest)
					}
				case termbox.KeyCtrlH: // termbox.KeyBackspace
					if len(what) > 0 {
						what = what[:len(what)-1]
						nfs.prompt(what).shadow(rest)
					}
				case termbox.KeyCtrlK:
					back = back[:0]
					back, rest = append(back, rest...), rest[:0]
					nfs.prompt(what).shadow(rest)

				case termbox.KeyCtrlW:
					if len(what) > 0 {
						pos := len(what) - 1
						for space := what[pos] == ' '; pos >= 0; pos-- {
							if (space && what[pos] != ' ') || (!space && what[pos] == ' ') {
								break
							}
						}
						back = back[:0]
						back, what = append(back, what[pos+1:]...), what[:pos+1]
						nfs.prompt(what).shadow(rest)
					}
				case termbox.KeyCtrlY:
					what = append(what, back...)
					nfs.prompt(what).shadow(rest)

				case termbox.KeyCtrlR:
					nfs.Term(m, "refresh").Term(m, "flush")
					nfs.prompt(what).shadow(rest)
				case termbox.KeyCtrlL:
					m.Confi("term", "begin_row", m.Capi("noutput"))
					m.Confi("term", "begin_col", 0)
					nfs.Term(m, "clear", "all").Term(m, "flush")
					nfs.prompt(what).shadow(rest)

				case termbox.KeyCtrlT:
					m.Option("scroll", true)
					nfs.Term(m, "scroll", scroll_lines).Term(m, "flush")
					m.Option("scroll", false)
				case termbox.KeyCtrlO:
					m.Option("scroll", true)
					nfs.Term(m, "scroll", -scroll_lines).Term(m, "flush")
					m.Option("scroll", false)

				case termbox.KeyCtrlJ, termbox.KeyEnter:
					what = append(what, '\n')
					n = copy(p, []byte(string(what)))
					return
				case termbox.KeyCtrlQ, termbox.KeyCtrlC:
					nfs.Term(m, "exit")
					n = copy(p, []byte("return\n"))
					return

				case termbox.KeyCtrlV:

					for i := -1; i < 8; i++ {
						nfs.Term(m, "color", i, m.Confi("term", "rest_fg"), fmt.Sprintf("\nhello bg %d", i))
					}

					for i := -1; i < 8; i++ {
						nfs.Term(m, "color", m.Confi("term", "rest_bg"), i, fmt.Sprintf("\nhello fg %d", i))
					}
					nfs.Term(m, "flush")

				case termbox.KeyCtrlG:
				case termbox.KeyCtrlX:
					m.Options("show_shadow", !m.Options("show_shadow"))
				case termbox.KeyCtrlS:

				case termbox.KeyCtrlZ:

				case termbox.KeyTab:
					m.Options("show_shadow", true)
					// if index > len(what) {
					// 	nfs.shadow("", table, frame)
					// } else {
					// 	if lines := kit.Int(frame["lines"]); lines > 0 {
					// 		pick = (pick + 1) % lines
					// 	}
					// 	nfs.shadow(what[index:], table, frame, pick)
					// 	rest = append(rest[:0], []rune(kit.Format(frame["pick"]))[len(what)-index:]...)
					// 	nfs.prompt(what).shadow(rest)
					// 	nfs.shadow(what[index:], table, frame, pick)
					// }
					//
				case termbox.KeySpace:
					what = append(what, ' ')
					nfs.prompt(what).shadow(rest)
					if !m.Options("show_shadow") {
						break
					}

					if index > len(what) {
						nfs.shadow("", table, frame)
					} else {
						m.Option("auto_key", strings.TrimSpace(string(what[index:])))
						if change, f, t, i := nfs.Auto(what, " ", len(what)); change {
							frame, table, index, pick = f, t, i, 0
						}

						if nfs.shadow(what[index:], table, frame); len(table) > 0 {
							rest = append(rest[:0], []rune(table[0][kit.Format(frame["field"])])...)
							nfs.prompt(what).shadow(rest)
							nfs.shadow(what[index:], table, frame)
						}
					}

				default:
					what = append(what, ev.Ch)
					nfs.prompt(what).shadow(rest)
					if !m.Options("show_shadow") {
						break
					}

					if change, f, t, i := nfs.Auto(what, kit.Format(ev.Ch), len(what)); change {
						frame, table, index, pick = f, t, i, 0
					}

					if index > len(what) {
						nfs.shadow("", table, frame)
					} else {
						nfs.shadow(what[index:], table, frame, pick)
						if pos, word := len(what)-index, kit.Format(frame["pick"]); len(table) > 0 && pos < len(word) {
							rest = append(rest[:0], []rune(word)[pos:]...)
						} else {
						}
						nfs.prompt(what).shadow(rest)
						nfs.shadow(what[index:], table, frame, pick)
					}
				}
			}
		}
	})
	return
}
func (nfs *NFS) Auto(what []rune, trigger string, index int) (change bool, frame map[string]interface{}, table []map[string]string, nindex int) {
	m := nfs.Context.Message()

	auto_target := m.Optionv("auto_target").(*ctx.Context)
	auto_cmd := ""
	auto_arg := []string{}

	switch trigger {
	case " ":
		switch m.Conf("term", "help_state") {
		case "context":
			auto_target = auto_target.Sub(m.Option("auto_key"))
			m.Optionv("auto_target", auto_target)
			trigger = ":"
		case "command":
			m.Option("arg_index", index)
			auto_cmd = m.Option("auto_cmd", m.Option("auto_key"))
			trigger = "="
		case "argument":
			auto_cmd = m.Option("auto_cmd")
			auto_arg = strings.Split(strings.TrimSpace(string(what[m.Optioni("arg_index"):])), " ")
			trigger = "="
		}
	}

	auto := m.Confm("auto", trigger)
	if auto == nil {
		return false, nil, nil, 0
	}

	cmd := []interface{}{kit.Select(kit.Format(auto["cmd"]), auto_cmd)}
	if len(auto_arg) == 0 {
		auto_arg = kit.Trans(auto["arg"])
	}
	for _, v := range auto_arg {
		cmd = append(cmd, m.Parse(v))
	}

	table = []map[string]string{}
	m.Spawn(auto_target).Cmd(cmd...).Table(func(maps map[string]string, list []string, line int) bool {
		if line >= 0 {
			fields := []interface{}{}
			for _, v := range auto["fields"].([]interface{}) {
				fields = append(fields, maps[kit.Format(v)])
			}
			maps["format"] = fmt.Sprintf(kit.Format(auto["format"]), fields...)
			table = append(table, maps)
		}
		return true
	})
	m.Conf("term", []interface{}{"help_table", auto["table"]}, table)

	frame = map[string]interface{}{
		"color": auto["color"],
		"table": auto["table"],
		"field": auto["field"],
	}

	if m.Conf("term", []interface{}{"help_index"}, index); index == 0 {
		m.Conf("term", "help_stack", []interface{}{frame})
	} else {
		m.Conf("term", []interface{}{"help_stack", -2}, frame)
	}

	m.Conf("term", "help_next_auto", auto["next_auto"])
	m.Conf("term", "help_state", auto["state"])
	return true, frame, table, index
}
func (nfs *NFS) Term(msg *ctx.Message, action string, args ...interface{}) *NFS {
	m := nfs.Context.Message()
	// m.Log("debug", "%s %v", action, args)

	switch action {
	case "init":
		termbox.Init()
		termbox.SetInputMode(termbox.InputEsc)
		termbox.SetInputMode(termbox.InputMouse)
		m.Cap("termbox", "true")
	}

	width, height := termbox.Size()
	msg.Conf("term", "width", width)
	msg.Conf("term", "height", height)

	left := msg.Confi("term", "left")
	top := msg.Confi("term", "top")
	right := msg.Confi("term", "right")
	bottom := msg.Confi("term", "bottom")

	x := m.Confi("term", "cursor_x")
	y := m.Confi("term", "cursor_y")
	bg := termbox.Attribute(msg.Confi("term", "bgcolor"))
	fg := termbox.Attribute(msg.Confi("term", "fgcolor"))

	begin_row := m.Confi("term", "begin_row")
	begin_col := m.Confi("term", "begin_col")

	switch action {
	case "init":
		m.Conf("term", "left", 0)
		m.Conf("term", "top", 0)
		m.Conf("term", "right", width)
		m.Conf("term", "bottom", height)

	case "exit":
		m.Cap("termbox", "false")
		termbox.Close()

	case "window":
		if len(args) > 1 {
			msg.Conf("term", "left", args[0])
			msg.Conf("term", "top", args[1])
		}
		if len(args) > 3 {
			msg.Conf("term", "right", args[2])
			msg.Conf("term", "bottom", args[3])
		} else {
			msg.Conf("term", "right", width)
			msg.Conf("term", "bottom", height)
		}

	case "resize":
		if len(args) > 1 {
			msg.Conf("term", "right", args[0])
			msg.Conf("term", "bottom", args[1])
			right = msg.Confi("term", "right")
			bottom = msg.Confi("term", "bottom")
		} else {
			msg.Conf("term", "right", right)
			msg.Conf("term", "bottom", bottom)
		}

		fallthrough
	case "clear":
		if len(args) == 0 {
			top = m.Confi("term", "prompt_y")
		} else if kit.Format(args[0]) == "all" {
			// nothing
		}

		for x := left; x < right; x++ {
			for y := top; y < bottom; y++ {
				termbox.SetCell(x, y, ' ', fg, bg)
			}
		}
		m.Conf("term", "cursor_x", left)
		m.Conf("term", "cursor_y", top)
		termbox.SetCursor(left, top)

	case "cursor":
		m.Conf("term", "cursor_x", kit.Format(args[0]))
		m.Conf("term", "cursor_y", kit.Format(args[1]))
		termbox.SetCursor(m.Confi("term", "cursor_x"), m.Confi("term", "cursor_y"))

	case "flush":
		termbox.Flush()

	case "scroll":
		n := 1
		if len(args) > 0 {
			n = kit.Int(args[0])
		}
		m.Options("on_scroll", true)

		// 向下滚动
		for i := begin_row; n > 0 && i < m.Capi("noutput"); i++ {
			line := []rune(m.Conf("output", []interface{}{i, "line"}))

			for j, l := begin_col, left; n > 0; j, l = j+1, l+1 {
				if j >= len(line)-1 {
					begin_row, begin_col = i+1, 0
					n--
					break
				} else if line[j] == '\n' {
					begin_row, begin_col = i, j+1
					n--
				} else if l >= right-1 && m.Confs("term", "wrap") {
					begin_row, begin_col = i, j
					n--
				}
			}
		}

		// 向上滚动
		for i := begin_row; n < 0 && i >= 0; i-- {
			line := []rune(m.Conf("output", []interface{}{i, "line"}))
			if begin_col == 0 {
				i--
				line = []rune(m.Conf("output", []interface{}{i, "line"}))
				begin_col = len(line)
			}

			for j, l := begin_col-1, right-1; n < 0; j, l = j-1, l-1 {
				if j <= 0 {
					begin_row, begin_col = i, 0
					n++
					break
				} else if line[j-1] == '\n' {
					begin_row, begin_col = i, j
					n++
				} else if l < left && m.Confs("term", "wrap") {
					begin_row, begin_col = i, j
					n++
				}
			}
		}

		m.Conf("term", "begin_row", begin_row)
		m.Conf("term", "begin_col", begin_col)

		fallthrough
	case "refresh":

		nfs.Term(m, "clear", "all")
		for i := begin_row; i < m.Capi("noutput"); i++ {
			if line := m.Conf("output", []interface{}{i, "line"}); begin_col < len(line) {
				nfs.Term(m, "print", line[begin_col:])
			}
			begin_col = 0
		}
		// nfs.Term(m, "print", "\n")
		// nfs.Term(m, "print", m.Conf("prompt"))
		m.Options("on_scroll", false)

	case "print":
		list := kit.Format(args...)
		n := strings.Count(list, "\n") + y - bottom

		for _, v := range list {
			if x < right {
				if termbox.SetCell(x, y, v, fg, bg); v > 255 {
					x++
				}
			}

			if x++; v == '\n' || (x >= right && m.Confs("term", "wrap")) {
				x, y = left, y+1
				if y >= bottom {
					if m.Options("on_scroll") {
						break
					}
					if n%bottom > 0 {

						nfs.Term(m, "scroll", n%bottom+1)
						n -= n % bottom
						x = m.Confi("term", "cursor_x")
						y = m.Confi("term", "cursor_y")

					} else if n > 0 {

						nfs.Term(m, "scroll", bottom)
						n -= bottom
						x = m.Confi("term", "cursor_x")
						y = m.Confi("term", "cursor_y")

					}
				}
			}

			if x < right {
				m.Conf("term", "cursor_x", x)
				m.Conf("term", "cursor_y", y)
				termbox.SetCursor(x, y)
			}
		}

		if m.Options("on_scroll") {
			x = 0
			y = y + 1
			m.Conf("term", "cursor_x", x)
			m.Conf("term", "cursor_y", y)
			termbox.SetCursor(x, y)
		}

	case "color":
		msg.Conf("term", "bgcolor", kit.Int(args[0])+1)
		msg.Conf("term", "fgcolor", kit.Int(args[1])+1)
		nfs.Term(m, "print", args[2:]...)
		msg.Conf("term", "fgcolor", fg)
		msg.Conf("term", "bgcolor", bg)

	case "shadow":
		x := m.Confi("term", "cursor_x")
		y := m.Confi("term", "cursor_y")
		nfs.Term(m, "color", args...)
		nfs.Term(m, "cursor", x, y).Term(m, "flush")
	}
	return nfs
}
func (nfs *NFS) prompt(arg ...interface{}) *NFS {
	m := nfs.Context.Message()
	target, _ := m.Optionv("ps_target").(*ctx.Context)
	if target == nil {
		target = nfs.Context
	}

	line := fmt.Sprintf("%d[%s]%s> ", m.Capi("ninput"), time.Now().Format("15:04:05"), target.Name)
	m.Conf("prompt", line)

	line += kit.Format(arg...)
	if m.Caps("termbox") {
		m.Conf("term", "prompt_y", m.Conf("term", "cursor_y"))
		nfs.Term(m, "clear").Term(m, "print", line).Term(m, "flush")
	} else if nfs.out != nil {
		nfs.out.WriteString(line)
	}
	return nfs
}
func (nfs *NFS) shadow(args ...interface{}) *NFS {
	if len(args) == 0 {
		return nfs
	}

	m := nfs.Context.Message()
	x := m.Confi("term", "cursor_x")
	y := m.Confi("term", "cursor_y")
	defer func() { nfs.Term(m, "cursor", x, y).Term(m, "flush") }()

	switch arg := args[0].(type) {
	case []rune:
		if len(args) == 1 {
			nfs.Term(m, "color", m.Confi("term", "rest_bg"), m.Confi("term", "rest_fg"), string(arg))
		} else {
			cmd := strings.Split(string(arg), " ")
			switch table := args[1].(type) {
			case []map[string]string:
				frame := args[2].(map[string]interface{})
				field := kit.Format(frame["field"])
				fg := kit.Int(frame["color"])
				pick := kit.Int(kit.Select("0", args, 3))

				i := 0
				for _, line := range table {
					if strings.Contains(kit.Format(line[field]), cmd[0]) {
						if i == pick {
							frame["pick"] = line[field]
							nfs.Term(m, "color", m.Confi("term", "pick_bg"), m.Confi("term", "pick_fg"), "\n", kit.Format(line["format"]))
						} else {
							nfs.Term(m, "color", 0, fg, "\n", kit.Format(line["format"]))
						}
						i++
						if i > 10 {
							break
						}
					}
				}
				frame["lines"] = i
			}
		}

	}

	return nfs
}
func (nfs *NFS) printf(arg ...interface{}) *NFS {
	m := nfs.Context.Message()

	line := strings.TrimRight(kit.Format(arg...), "\n")
	m.Log("debug", "noutput %s", m.Cap("noutput", m.Capi("noutput")+1))
	m.Confv("output", -2, map[string]interface{}{"time": time.Now().Unix(), "line": line})

	if m.Caps("termbox") {
		nfs.Term(m, "clear").Term(m, "print", line).Term(m, "flush")
		m.Conf("term", "prompt_y", m.Confi("term", "cursor_y")+1)
		m.Conf("term", "cursor_y", m.Confi("term", "cursor_y")+1)
	} else {
		nfs.out.WriteString(line)
		nfs.out.WriteString("\n")
	}
	return nfs
}

func (nfs *NFS) Recv(line string) (field string, value string) {
	m := nfs.Context.Message()

	m.Log("recv", "%d [%s]", len(line), line)
	m.Capi("nread", len(line)+1)

	word := strings.Split(line, ": ")
	field, e := url.QueryUnescape(word[0])
	m.Assert(e)
	if len(word) == 1 {
		return
	}
	if len(word[1]) == 0 {
		return
	}

	value, e = url.QueryUnescape(word[1])
	m.Assert(e)
	return
}
func (nfs *NFS) Send(meta string, arg ...interface{}) *NFS {
	m := nfs.Context.Message()

	line := "\n"
	if meta != "" {
		line = fmt.Sprintf("%s: %s\n", url.QueryEscape(meta), url.QueryEscape(kit.Format(arg[0])))
	}

	n, e := fmt.Fprint(nfs.io, line)
	m.Assert(e)
	m.Capi("nwrite", n)
	m.Log("send", "%d [%s]", len(line), line)

	return nfs
}

func (nfs *NFS) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	if len(arg) > 0 && (arg[0] == "scan" || arg[0] == "open" || arg[0] == "append") {
		c.Caches = map[string]*ctx.Cache{
			"pos":    &ctx.Cache{Name: "pos", Value: "0", Help: "pos"},
			"size":   &ctx.Cache{Name: "size", Value: "0", Help: "size"},
			"nread":  &ctx.Cache{Name: "nread", Value: "0", Help: "nread"},
			"nwrite": &ctx.Cache{Name: "nwrite", Value: "0", Help: "nwrite"},
		}
		c.Configs = map[string]*ctx.Config{}
	} else {
		c.Caches = map[string]*ctx.Cache{
			"nsend":  &ctx.Cache{Name: "消息发送数量", Value: "0", Help: "消息发送数量"},
			"nrecv":  &ctx.Cache{Name: "消息接收数量", Value: "0", Help: "消息接收数量"},
			"nread":  &ctx.Cache{Name: "nread", Value: "0", Help: "nread"},
			"nwrite": &ctx.Cache{Name: "nwrite", Value: "0", Help: "nwrite"},
		}
		c.Configs = map[string]*ctx.Config{}
	}

	s := new(NFS)
	s.Context = c
	return s

}
func (nfs *NFS) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return nfs
}
func (nfs *NFS) Start(m *ctx.Message, arg ...string) bool {

	if len(arg) > 0 && (arg[0] == "open" || arg[0] == "append") {
		nfs.out = m.Optionv("out").(*os.File)
		nfs.in = m.Optionv("in").(*os.File)
		m.Cap("stream", arg[1])

		if s, e := nfs.in.Stat(); m.Assert(e) {
			if m.Capi("size", int(s.Size())); arg[0] == "append" {
				m.Capi("pos", int(s.Size()))
			}
		}
		return false
	}

	if len(arg) > 0 && arg[0] == "scan" {
		// 终端用户
		m.Cmd("aaa.role", "root", "user", m.Option("username", m.Conf("runtime", "boot.USER")))

		// 创建会话
		m.Option("sessid", m.Cmdx("aaa.user", "session", "select"))

		// 创建空间
		m.Option("bench", m.Cmdx("aaa.sess", "bench", "select"))

		// 默认配置
		m.Cap("stream", arg[1])
		nfs.Caches["ninput"] = &ctx.Cache{Value: "0"}
		nfs.Caches["noutput"] = &ctx.Cache{Value: "0"}
		nfs.Caches["termbox"] = &ctx.Cache{Value: "0"}
		nfs.Configs["input"] = &ctx.Config{Value: []interface{}{}}
		nfs.Configs["output"] = &ctx.Config{Value: []interface{}{}}
		nfs.Configs["prompt"] = &ctx.Config{Value: ""}

		// 终端控制
		if nfs.in = m.Optionv("in").(*os.File); m.Has("out") {
			if nfs.out = m.Optionv("out").(*os.File); m.Cap("goos") != "windows" && !m.Options("daemon") {
				nfs.Term(m, "init")
				defer nfs.Term(m, "exit")
			}
			if what := make(chan bool); m.Options("daemon") {
				<-what
			}
		}

		line, bio := "", bufio.NewScanner(nfs)
		for nfs.prompt(); !m.Options("scan_end"); nfs.prompt() {
			for bio.Scan() {
				if text := bio.Text(); text == "" {
					continue
				} else if !strings.HasSuffix(text, "\\") {
					line += text
					break
				} else {
					line += strings.TrimSuffix(text, "\\")
				}
			}
			if line == "" {
				line = "return"
			}

			m.Log("debug", "%s %d %d [%s]", "input", m.Capi("ninput", 1), len(line), line)
			m.Confv("input", -2, map[string]interface{}{"time": time.Now().Unix(), "line": line})

			for i := m.Capi("ninput") - 1; i < m.Capi("ninput"); i++ {
				line = m.Conf("input", []interface{}{i, "line"})

				msg := m.Backs(m.Spawn(m.Source()).Set(
					"detail", line).Set(
					"option", "file_pos", i).Set(
					"option", "username", m.Conf("runtime", "boot.USER")))

				nfs.printf(m.Conf("prompt"), line)
				nfs.printf(msg.Meta["result"])

				if msg.Appends("file_pos0") {
					i = msg.Appendi("file_pos0") - 1
					msg.Append("file_pos0", "")
				}
			}
			line = ""
		}
		return false
	}

	m.Cap("stream", m.Option("ms_source"))
	nfs.io, _ = m.Optionv("io").(io.ReadWriter)
	nfs.send = make(chan *ctx.Message, 10)
	nfs.echo = make(chan *ctx.Message, 10)
	nfs.hand = map[int]*ctx.Message{}

	// 消息发送队列
	m.GoLoop(m, func(m *ctx.Message) {
		msg, code, meta, body := m, 0, "detail", "option"
		select {
		case msg = <-nfs.send: // 发送请求
			code = msg.Code()
			nfs.hand[code] = msg
		case msg = <-nfs.echo: // 发送响应
			code, meta, body = msg.Optioni("remote_code"), "result", "append"
		}

		nfs.Send("code", code)
		for _, v := range msg.Meta[meta] {
			nfs.Send(meta, v)
		}
		for _, k := range msg.Meta[body] {
			for _, v := range msg.Meta[k] {
				nfs.Send(k, v)
			}
		}
		nfs.Send("")
	})

	// 消息接收队列
	msg, code, head, body := m, "0", "result", "append"
	for bio := bufio.NewScanner(nfs.io); bio.Scan(); {

		m.TryCatch(m, true, func(m *ctx.Message) {
			switch field, value := nfs.Recv(bio.Text()); field {
			case "code":
				msg, code = m.Sess("ms_target"), value
				msg.Meta = map[string][]string{}

			case "detail":
				head, body = "detail", "option"
				msg.Add(field, value)

			case "result":
				head, body = "result", "append"
				msg.Add(field, value)

			case "":
				m.Log("recv", "time %v", time.Now().Format(m.Conf("time_format")))
				if head == "detail" { // 接收请求
					msg.Detail(-1, "remote")
					msg.Option("remote_code", code)
					go msg.Call(func(msg *ctx.Message) *ctx.Message {
						nfs.echo <- msg
						return nil
					})
				} else { // 接收响应
					h := nfs.hand[kit.Int(code)]
					h.Copy(msg, "result").Copy(msg, "append")
					go func() {
						h.Back(h)
					}()
				}
				msg, code, head, body = nil, "0", "result", "append"

			default:
				msg.Add(body, field, value)
			}
		}, func(m *ctx.Message) {
			for bio.Scan() {
				if text := bio.Text(); text == "" {
					break
				}
			}
		})
	}

	m.Sess("tcp", false).Close()
	return true
}
func (nfs *NFS) Close(m *ctx.Message, arg ...string) bool {
	switch nfs.Context {
	case m.Target():
		if nfs.in != nil {
			m.Log("info", "close in %s", m.Cap("stream"))
			nfs.in.Close()
			nfs.in = nil
		}
		if nfs.out != nil {
			m.Log("info", "close out %s", m.Cap("stream"))
			nfs.out.Close()
			nfs.out = nil
		}
	case m.Source():
	}
	if nfs.Name == "stdio" {
		return false
	}
	return true
}

var Index = &ctx.Context{Name: "nfs", Help: "存储中心",
	Caches: map[string]*ctx.Cache{
		"nfile": &ctx.Cache{Name: "nfile", Value: "0", Help: "已经打开的文件数量"},
	},
	Configs: map[string]*ctx.Config{
		"term": &ctx.Config{Name: "term", Value: map[string]interface{}{
			"mouse": map[string]interface{}{
				"resize": false,
			},
			"width": 80, "height": "24",

			"left": 0, "top": 0, "right": 80, "bottom": 24,
			"cursor_x": 0, "cursor_y": 0, "fgcolor": -1, "bgcolor": -1,
			"prompt": "", "wrap": "false",
			"scroll_count": "5",
			"scroll_lines": "5",
			"begin_row":    0, "begin_col": 0,

			"shadow":      "hello",
			"show_shadow": "false",

			"rest_fg": "0",
			"rest_bg": "7",
			"pick_fg": "0",
			"pick_bg": "7",
			"pick":    "",

			"help_index":     0,
			"help_state":     "command",
			"help_next_auto": "=",
			"help_stack":     []interface{}{},
			"help_table":     map[string]interface{}{},
		}, Help: "二维码的默认大小"},
		"auto": &ctx.Config{Name: "auto", Value: map[string]interface{}{
			"!": map[string]interface{}{
				"state":     "message",
				"next_auto": ":",
				"color":     2, "cmd": "message",
				"table": "message", "field": "code",
				"format": "%s(%s) %s->%s %s %s", "fields": []interface{}{"code", "time", "source", "target", "details", "options"},
			},
			"~": map[string]interface{}{
				"state": "context", "next_auto": ":",
				"color": 2, "cmd": "context",
				"table": "context", "field": "name",
				"format": "%s(%s) %s %s", "fields": []interface{}{"name", "status", "stream", "help"},
			},
			"": map[string]interface{}{
				"state": "command", "next_auto": "=",
				"color": 3, "cmd": "command",
				"table": "command", "field": "key",
				"format": "%s %s", "fields": []interface{}{"key", "name"},
			},
			":": map[string]interface{}{
				"state": "command", "next_auto": "=",
				"color": 3, "cmd": "command",
				"table": "command", "field": "key",
				"format": "%s %s", "fields": []interface{}{"key", "name"},
			},
			"=": map[string]interface{}{
				"cmd":    "help",
				"format": "%s %s %s ", "fields": []interface{}{"value", "name", "help"},
				"color": 3, "table": "command", "field": "value",
				"state": "argument", "next_auto": "=",
			},
			"@": map[string]interface{}{
				"state": "config", "next_auto": "@",
				"color": 4, "cmd": "config",
				"table": "config", "field": "key",
				"format": "%s(%s) %s", "fields": []interface{}{"key", "value", "name"},
			},
			"$": map[string]interface{}{
				"state": "cache", "next_auto": "$",
				"color": 7, "cmd": "cache",
				"table": "cache", "field": "key",
				"format": "%s(%s) %s", "fields": []interface{}{"key", "value", "name"},
			},
		}, Help: "读取文件的缓存区的大小"},

		"buf_size":   &ctx.Config{Name: "buf_size", Value: "1024", Help: "读取文件的缓存区的大小"},
		"dir_type":   &ctx.Config{Name: "dir_type(file/dir/both/all)", Value: "both", Help: "dir命令输出的文件类型, file: 只输出普通文件, dir: 只输出目录文件, 否则输出所有文件"},
		"dir_fields": &ctx.Config{Name: "dir_fields(time/type/name/size/line/hash)", Value: "time size line filename", Help: "dir命令输出文件名的类型, name: 文件名, tree: 带缩进的文件名, path: 相对路径, full: 绝对路径"},

		"git": &ctx.Config{Name: "git", Value: map[string]interface{}{
			"args":   []interface{}{"-C", "@git_dir"},
			"info":   map[string]interface{}{"cmds": []interface{}{"log", "status", "branch"}},
			"branch": map[string]interface{}{"args": []interface{}{"branch", "-v"}},
			"status": map[string]interface{}{"args": []interface{}{"status", "-sb"}},
			"log":    map[string]interface{}{"args": []interface{}{"log", "-n", "limit", "--reverse", "pretty", "date"}},
			"trans": map[string]interface{}{
				"date":   "--date=format:%m/%d %H:%M",
				"pretty": "--pretty=format:%h %ad %an %s",
				"limit":  "10",
			},
		}, Help: "命令集合"},
		"paths": &ctx.Config{Name: "paths", Value: []interface{}{"var", "usr", "etc", "bin", ""}, Help: "文件路径"},
	},
	Commands: map[string]*ctx.Command{
		"init": &ctx.Command{Name: "init", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Conf("paths", -2, m.Conf("runtime", "boot.ctx_home"))
			m.Conf("paths", -2, m.Conf("runtime", "boot.ctx_root"))
			return
		}},
		"pwd": &ctx.Command{Name: "pwd [all] | [[index] path] ", Help: "工作目录，all: 查看所有, index path: 设置路径, path: 设置当前路径", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) > 0 && arg[0] == "all" {
				m.Cmdy("nfs.config", "paths")
				return
			}

			index := 0
			if len(arg) > 1 {
				index, arg = kit.Int(arg[0]), arg[1:]
			}
			for i, v := range arg {
				m.Log("info", "paths %s %s", index+i, v)
				m.Confv("paths", index+i, v)
			}

			if p := m.Conf("paths", index); path.IsAbs(p) {
				m.Echo("%s", p)
			} else if wd, e := os.Getwd(); m.Assert(e) {
				m.Echo("%s", path.Join(wd, p))
			}
			return
		}},
		"dir": &ctx.Command{Name: "dir path [dir_deep] [dir_type both|file|dir] [dir_reg reg] [dir_sort field order] fields...",
			Help: "查看目录, path: 路径, dir_deep: 递归查询, dir_type: 文件类型, dir_reg: 正则表达式, dir_sort: 排序, fields: 查询字段",
			Form: map[string]int{"dir_deep": 0, "dir_type": 1, "dir_reg": 1, "dir_sort": 2},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if len(arg) == 0 {
					arg = append(arg, "")
				}

				wd, e := os.Getwd()
				m.Assert(e)
				trip := len(wd) + 1

				rg, e := regexp.Compile(m.Option("dir_reg"))

				m.Confm("paths", func(index int, value string) bool {
					p := path.Join(value, m.Option("dir_root"), arg[0])
					if s, e := os.Stat(p); e == nil {
						if s.IsDir() {
							dir(m, p, 0, kit.Right(m.Has("dir_deep")), m.Confx("dir_type"), trip, rg,
								strings.Split(m.Confx("dir_fields", strings.Join(arg[1:], " ")), " "),
								m.Conf("time_format"))
						} else {
							m.Append("directory", p)
						}
						return true
					}
					return false
				})

				if m.Has("dir_sort") {
					m.Sort(m.Meta["dir_sort"][0], m.Meta["dir_sort"][1:]...)
				}

				if len(m.Meta["append"]) == 1 {
					for _, v := range m.Meta[m.Meta["append"][0]] {
						m.Echo(v).Echo(" ")
					}
				} else {
					m.Table()
				}
				return
			}},
		"git": &ctx.Command{Name: "git sum", Help: "版本控制", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) > 0 && arg[0] == "sum" {
				if out, e := exec.Command("git", "log", "--shortstat", "--pretty=commit: %ad", "--date=format:%Y-%m-%d").CombinedOutput(); m.Assert(e) {
					for _, v := range strings.Split(string(out), "commit: ") {
						if l := strings.Split(v, "\n"); len(l) > 2 {
							fs := strings.Split(strings.TrimSpace(l[2]), ", ")
							m.Add("append", "date", l[0])

							if adds := strings.Split(fs[1], " "); len(fs) > 2 {
								dels := strings.Split(fs[2], " ")
								m.Add("append", "adds", adds[0])
								m.Add("append", "dels", dels[0])
							} else if adds[1] == "insertions(+)" {
								m.Add("append", "adds", adds[0])
								m.Add("append", "dels", "0")
							} else {
								m.Add("append", "adds", "0")
								m.Add("append", "dels", adds[0])
							}
						}
					}
					m.Table()
				}
				return
			}

			if len(arg) == 0 {
				m.Cmdy("nfs.config", "git")
				return
			}

			wd, e := os.Getwd()
			m.Assert(e)
			m.Option("git_dir", wd)

			cmds := []string{}
			if v := m.Confv("git", []string{arg[0], "cmds"}); v != nil {
				cmds = append(cmds, kit.Trans(v)...)
			} else {
				cmds = append(cmds, arg[0])
			}

			for _, cmd := range cmds {
				args := append([]string{}, kit.Trans(m.Confv("git", "args"))...)
				if v := m.Confv("git", []string{cmd, "args"}); v != nil {
					args = append(args, kit.Trans(v)...)
				} else {
					args = append(args, cmd)
				}
				args = append(args, arg[1:]...)

				for i, _ := range args {
					args[i] = m.Parse(args[i])
					args[i] = kit.Select(args[i], m.Conf("git", []string{"trans", args[i]}))
				}

				m.Cmd("cli.system", "git", args).Echo("\n\n").CopyTo(m)
			}
			return
		}},

		"temp": &ctx.Command{Name: "temp data", Help: "查找文件路径", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			h, _ := kit.Hash("uniq")
			name := fmt.Sprintf("var/tmp/file/%s", h)

			m.Assert(os.MkdirAll("var/tmp/file/", 0777))
			f, e := os.Create(name)
			m.Assert(e)
			defer f.Close()
			f.Write([]byte(arg[0]))

			m.Echo(name)
			return
		}},
		"trash": &ctx.Command{Name: "trash file", Help: "查找文件路径", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			os.Remove(arg[0])
			return
		}},

		"path": &ctx.Command{Name: "path filename", Help: "查找文件路径", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				return
			}

			m.Confm("paths", func(index int, value string) bool {
				p := path.Join(value, arg[0])
				if _, e := os.Stat(p); e == nil {
					m.Echo(p)
					return true
				}
				return false
			})
			return
		}},
		"load": &ctx.Command{Name: "load file [buf_size [pos]]", Help: "加载文件, buf_size: 加载大小, pos: 加载位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if p, f, e := open(m, arg[0]); e == nil {
				defer f.Close()

				pos := kit.Int(kit.Select("0", arg, 2))
				size := kit.Int(m.Confx("buf_size", arg, 1))
				buf := make([]byte, size)

				if l, e := f.ReadAt(buf, int64(pos)); e == io.EOF || m.Assert(e) {
					m.Log("info", "load %s %d %d", p, l, pos)
					m.Echo(string(buf[:l]))
				}
			}
			return
		}},
		"save": &ctx.Command{Name: "save file string...", Help: "保存文件, file: 保存的文件, string: 保存的内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 1 && m.Has("data") {
				arg = append(arg, m.Option("data"))
			}
			if p, f, e := open(m, kit.Format(arg[0]), os.O_WRONLY|os.O_CREATE|os.O_TRUNC); m.Assert(e) {
				defer f.Close()
				m.Append("directory", p)
				m.Echo(p)

				for _, v := range arg[1:] {
					n, e := fmt.Fprint(f, v)
					m.Assert(e)
					m.Log("info", "save %s %d", p, n)
				}
			}
			return
		}},
		"import": &ctx.Command{Name: "import filename [index]", Help: "导入数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			p, f, e := open(m, arg[0])
			m.Assert(e)
			defer f.Close()

			s, e := f.Stat()
			m.Option("filepath", p)
			m.Option("filename", s.Name())
			m.Option("filesize", s.Size())
			m.Option("filetime", s.ModTime().Format(m.Conf("time_format")))

			switch {
			case strings.HasSuffix(arg[0], ".json"):
				var data interface{}
				de := json.NewDecoder(f)
				de.Decode(&data)

				m.Put("option", "filedata", data).Cmdy("ctx.trans", "filedata", arg[1:]).CopyTo(m)
			case strings.HasSuffix(arg[0], ".csv"):
				r := csv.NewReader(f)

				l, e := r.Read()
				m.Assert(e)
				m.Meta["append"] = l

				for l, e = r.Read(); e != nil; l, e = r.Read() {
					for i, v := range l {
						m.Add("append", m.Meta["append"][i], v)
					}
				}
				m.Table()
			default:
				b, e := ioutil.ReadAll(f)
				m.Assert(e)
				m.Echo(string(b))
			}
			return
		}},
		"export": &ctx.Command{Name: "export filename", Help: "导出数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			tp := false
			if len(arg) > 0 && arg[0] == "time" {
				tp, arg = true, arg[1:]
			}

			p, f, e := open(m, kit.Select(arg[0], m.Format(arg[0]), tp), os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
			m.Assert(e)
			defer f.Close()

			data := m.Optionv(kit.Select("data", arg, 1))
			if len(arg) > 0 && arg[0] == "all" {
				data, arg = m.Meta, arg[1:]
			}

			switch {
			case strings.HasSuffix(arg[0], ".json"):
				if data == nil && len(m.Meta["append"]) > 0 {
					lines := []interface{}{}
					nrow := len(m.Meta[m.Meta["append"][0]])
					for i := 0; i < nrow; i++ {
						line := map[string]interface{}{}
						for _, k := range m.Meta["append"] {
							line[k] = m.Meta[k][i]
						}

						lines = append(lines, line)
						data = lines
					}
				}

				en := json.NewEncoder(f)
				en.SetIndent("", "  ")
				en.Encode(data)
			case strings.HasSuffix(arg[0], ".csv"):
				fields := m.Meta["append"]
				if m.Options("fields") {
					fields = m.Meta["fields"]
				}

				if data == nil && len(m.Meta["append"]) > 0 {
					lines := []interface{}{}
					nrow := len(m.Meta[m.Meta["append"][0]])
					for i := 0; i < nrow; i++ {
						line := []string{}
						for _, k := range fields {
							line = append(line, m.Meta[k][i])
						}
						lines = append(lines, line)
						data = lines
					}
				}

				if data, ok := data.([]interface{}); ok {
					w := csv.NewWriter(f)
					w.Write(fields)
					for _, v := range data {
						w.Write(kit.Trans(v))
					}
					w.Flush()
				}
			case strings.HasSuffix(arg[0], ".png"):
				if data == nil {
					data = kit.Format(arg[1:])
				}

				qr, e := qrcode.New(kit.Format(data), qrcode.Medium)
				m.Assert(e)
				m.Assert(qr.Write(256, f))
			default:
				f.WriteString(kit.Format(m.Meta["result"]))
			}

			m.Set("append").Add("append", "directory", p)
			m.Set("result").Echo(p)
			return
		}},
		"json": &ctx.Command{Name: "json str", Help: "导入数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			var data interface{}
			m.Assert(json.Unmarshal([]byte(arg[0]), &data))
			b, e := json.MarshalIndent(data, "", "  ")
			m.Assert(e)
			m.Echo(string(b))
			return
		}},

		"open": &ctx.Command{Name: "open file", Help: "打开文件, file: 文件名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if m.Has("io") {
			} else if p, f, e := open(m, arg[0], os.O_RDWR|os.O_CREATE); e == nil {
				m.Put("option", "in", f).Put("option", "out", f)
				arg[0] = p
			} else {
				return nil
			}

			m.Start(fmt.Sprintf("file%d", m.Capi("nfile")), fmt.Sprintf("file %s", arg[0]), "open", arg[0])
			m.Append("ps_target1", m.Cap("module"))
			m.Echo(m.Cap("module"))
			return
		}},
		"read": &ctx.Command{Name: "read [buf_size [pos]]", Help: "读取文件, buf_size: 读取大小, pos: 读取位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.in != nil {
				if len(arg) > 1 {
					m.Cap("pos", arg[1])
				}

				buf := make([]byte, kit.Int(m.Confx("buf_size", arg, 0)))
				if n, e := nfs.in.ReadAt(buf, int64(m.Capi("pos"))); e == io.EOF || m.Assert(e) {
					m.Capi("nread", n)
					if m.Capi("pos", n); n == 0 {
						m.Cap("pos", "0")
					}
				}
				m.Echo(string(buf))
			}
			return
		}},
		"write": &ctx.Command{Name: "write string [pos]", Help: "写入文件, string: 写入内容, pos: 写入位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.out != nil {
				if len(arg) > 1 {
					m.Cap("pos", arg[1])
				}

				if len(arg[0]) == 0 {
					m.Assert(nfs.out.Truncate(int64(m.Capi("pos"))))
					m.Cap("size", m.Cap("pos"))
					m.Cap("pos", "0")
				} else {
					n, e := nfs.out.WriteAt([]byte(arg[0]), int64(m.Capi("pos")))
					if m.Capi("nwrite", n); m.Assert(e) && m.Capi("pos", n) > m.Capi("size") {
						m.Cap("size", m.Cap("pos"))
					}
					nfs.out.Sync()
				}

				m.Echo(m.Cap("pos"))
			}
			return
		}},

		"scan": &ctx.Command{Name: "scan file name", Help: "扫描文件, file: 文件名, name: 模块名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if _, ok := m.Target().Server.(*NFS); m.Assert(ok) {
				if help := fmt.Sprintf("scan %s", arg[0]); arg[0] == "stdio" {
					m.Put("option", "in", os.Stdin).Put("option", "out", os.Stdout).Start(arg[0], help, key, arg[0])
				} else if p, f, e := open(m, arg[0]); m.Assert(e) {
					m.Put("option", "in", f).Start(fmt.Sprintf("file%d", m.Capi("nfile", 1)), help, key, p)
				}
			}
			return
		}},
		"printf": &ctx.Command{Name: "printf arg", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) {
				nfs.printf(arg)
			}
			return
		}},
		"prompt": &ctx.Command{Name: "prompt arg", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) {
				nfs.prompt(arg)
			}
			return
		}},
		"term": &ctx.Command{Name: "term action args...", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) {
				nfs.Term(m, arg[0], arg[1:])
			}
			return
		}},
		"action": &ctx.Command{Name: "action cmd", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) {
				msg := m.Cmd("cli.source", arg)
				nfs.printf(msg.Conf("prompt"), arg, "\n")
				nfs.printf(msg.Meta["result"])
			}
			return
		}},

		"remote": &ctx.Command{Name: "remote listen|dial args...", Help: "启动文件服务, args: 参考tcp模块, listen命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if _, ok := m.Target().Server.(*NFS); m.Assert(ok) { //{{{
				m.Sess("tcp").Call(func(sub *ctx.Message) *ctx.Message {
					if sub.Has("node.port") {
						return sub
					}
					sub.Sess("ms_source", sub)
					sub.Sess("ms_target", m.Source())
					sub.Start(fmt.Sprintf("file%d", m.Capi("nfile", 1)), "远程文件")
					return sub
				}, arg)
			}
			return
		}},
		"send": &ctx.Command{Name: "send [file] args...", Help: "连接文件服务, args: 参考tcp模块, dial命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.io != nil {
				nfs.send <- m.Set("detail", arg)
			}
			return
		}},
	},
}

func init() {
	nfs := &NFS{}
	nfs.Context = Index
	ctx.Index.Register(Index, nfs)
}
