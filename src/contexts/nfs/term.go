package nfs

import (
	"github.com/nsf/termbox-go"

	"contexts/ctx"
	"toolkit"

	"fmt"
	"strings"
	"time"
)

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

		m.Option("bio.cmd", "")
		m.Options("bio.shadow", m.Confs("show_shadow"))

		defer func() { m.Option("bio.cmd", "") }()

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
					m.Options("bio.shadow", !m.Options("bio.shadow"))
				case termbox.KeyCtrlS:

				case termbox.KeyCtrlZ:

				case termbox.KeyTab:
					m.Options("bio.shadow", true)
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
					if !m.Options("bio.shadow") {
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
					if !m.Options("bio.shadow") {
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
	return
	m := nfs.Context.Message()

	auto_target := m.Optionv("bio.ctx").(*ctx.Context)
	auto_cmd := ""
	auto_arg := []string{}

	switch trigger {
	case " ":
		switch m.Conf("term", "help_state") {
		case "context":
			// auto_target = auto_target.Sub(m.Option("auto_key"))
			m.Optionv("bio.ctx", auto_target)
			trigger = ":"
		case "command":
			m.Option("arg_index", index)
			auto_cmd = m.Option("bio.cmd", m.Option("auto_key"))
			trigger = "="
		case "argument":
			auto_cmd = m.Option("bio.cmd")
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
	m.Spawn(auto_target).Cmd(cmd...).Table(func(line int, maps map[string]string) {
		fields := []interface{}{}
		for _, v := range auto["fields"].([]interface{}) {
			fields = append(fields, maps[kit.Format(v)])
		}
		maps["format"] = fmt.Sprintf(kit.Format(auto["format"]), fields...)
		table = append(table, maps)
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
	case "exit":
		if m.Caps("termbox") {
			termbox.Close()
		}
		m.Caps("termbox", false)
		return nfs

	case "init":
		defer func() {
			if e := recover(); e != nil {
				m.Log("warn", "term init %s", e)
			}
		}()
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
	target, _ := m.Optionv("bio.ctx").(*ctx.Context)
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
func (nfs *NFS) print(arg ...string) *NFS {
	m := nfs.Context.Message()

	line := strings.TrimRight(strings.Join(arg, ""), "\n")
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
func (nfs *NFS) Show(arg ...interface{}) bool {
	nfs.prompt(arg...)
	return true
}
