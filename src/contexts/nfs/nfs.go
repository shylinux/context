package nfs

import (
	"bufio"
	"contexts/ctx"
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/skip2/go-qrcode"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"

	// "runtime"
	"sort"
	"strings"
	"time"
	"toolkit"
	"unicode"
)

type NFS struct {
	in  *os.File
	out *os.File

	pages  []string
	width  int
	height int

	io io.ReadWriter

	send chan *ctx.Message
	recv chan *ctx.Message
	hand map[int]*ctx.Message

	*bufio.Reader
	*bufio.Writer
	target *ctx.Context
	cli    *ctx.Message

	*ctx.Message
	*ctx.Context
}

func dir(m *ctx.Message, name string, level int, deep bool, dir_type string, trip int, dir_reg *regexp.Regexp, fields []string, format string) {
	back, e := os.Getwd()
	m.Assert(e)
	os.Chdir(name)
	defer os.Chdir(back)

	if fs, e := ioutil.ReadDir("."); m.Assert(e) {
		for _, f := range fs {
			if f.Name() == "." || f.Name() == ".." {
				continue
			}

			if strings.HasPrefix(f.Name(), ".") && dir_type != "both" {
				continue
			}

			f, _ := os.Stat(f.Name())
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
							d, e := ioutil.ReadDir(f.Name())
							m.Assert(e)
							nline = len(d)
						} else {
							f, e := os.Open(f.Name())
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
							d, e := ioutil.ReadDir(f.Name())
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

						f, e := ioutil.ReadFile(f.Name())
						m.Assert(e)
						h := sha1.Sum(f)
						m.Add("append", "hash", hex.EncodeToString(h[:]))
					}
				}
			}
			if f.IsDir() && deep {
				dir(m, f.Name(), level+1, deep, dir_type, trip, dir_reg, fields, format)
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

	m.Log("info", "open %s", name)
	f, e := os.OpenFile(name, flag, 0660)
	return name, f, e
}

func (nfs *NFS) insert(rest []rune, letters []rune) []rune {
	n := len(rest)
	l := len(letters)
	rest = append(rest, letters...)
	for i := n - 1; i >= 0; i-- {
		rest[i+l] = rest[i]
	}
	for i := 0; i < l; i++ {
		rest[i] = letters[i]
	}
	return rest
}
func (nfs *NFS) escape(form string, args ...interface{}) *NFS {
	if !nfs.Caps("windows") {
		fmt.Fprintf(nfs.out, "\033[%s", fmt.Sprintf(form, args...))
	}
	return nfs
}
func (nfs *NFS) color(str string, attr ...int) *NFS {
	if !nfs.Confs("color") {
		fmt.Fprint(nfs.out, str)
		return nfs
	}

	fg := nfs.Confi("fgcolor")
	if len(attr) > 0 {
		fg = attr[0]
	}

	bg := nfs.Confi("bgcolor")
	if len(attr) > 1 {
		bg = attr[1]
	}

	for i := 2; i < len(attr); i++ {
		nfs.escape("%dm", attr[i])
	}

	nfs.escape("4%dm", bg).escape("3%dm", fg)
	fmt.Fprint(nfs.out, str)
	nfs.escape("0m")
	return nfs
}
func (nfs *NFS) print(str string) bool {
	nfs.out.WriteString(str)
	return true
	ls := strings.Split(str, "\n")
	for i, l := range ls {
		rest := ""
		if len(nfs.pages) > 0 && !strings.HasSuffix(nfs.pages[len(nfs.pages)-1], "\n") {
			rest = nfs.pages[len(nfs.pages)-1]
			nfs.pages = nfs.pages[:len(nfs.pages)-1]
		}

		if rest += l; i < len(ls)-1 {
			rest += "\n"
		}
		nfs.pages = append(nfs.pages, rest)
		if nfs.Capi("cursor_pos") < nfs.height {
			nfs.Capi("cursor_pos", 1)
		}
	}

	switch {
	case nfs.out != nil:
		nfs.color(str)
	case nfs.io != nil:
		fmt.Fprint(nfs.io, str)
	default:
		return false
	}

	return true
}

func (nfs *NFS) zone(buf []string, top, height int) (row, col int) {
	row, col = len(buf)-1, 0
	for i := nfs.Capi("cursor_pos"); i > top-1; {
		if i -= len(buf[row]) / nfs.width; len(buf[row])%nfs.width > 0 {
			i--
		}
		if i < top-1 {
			col -= (i - (top - 1)) * nfs.width
		} else if i > (top-1) && row > 0 {
			row--
		}
	}
	return
}
func (nfs *NFS) page(buf []string, row, col, top, height int, status bool) {
	nfs.escape("2J").escape("H")
	begin := row

	for i := 0; i < height-1; i++ {
		if row >= len(buf) {
			nfs.color("~\n")
			continue
		}

		if len(buf[row])-col > nfs.width {
			nfs.color(buf[row][col : col+nfs.width])
			col += nfs.width
			continue
		}

		nfs.color(buf[row][col:])
		col = 0
		row++
	}

	if status {
		nfs.escape("E").color(fmt.Sprintf("pages: %d/%d", begin, len(nfs.pages)), nfs.Confi("statusfgcolor"), nfs.Confi("statusbgcolor"))
	}
}
func (nfs *NFS) View(buf []string, top int, height int) {

	row, col := nfs.zone(buf, top, height)
	nfs.page(buf, row, col, top, height, true)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				return
			default:
				switch ev.Ch {
				case 'f':
					for i := 0; i < height; i++ {
						if len(buf[row][col:]) > nfs.width {
							col += nfs.width
						} else {
							if col = 0; row < len(buf)-nfs.height {
								row++
							}
						}
					}
				case 'b':
					for i := 0; i < height; i++ {
						if col -= nfs.width; col < 0 {
							if col = 0; row > 0 {
								row--
							}
						}
					}
				case 'j':
					if len(buf[row][col:]) > nfs.width {
						col += nfs.width
					} else {
						if col = 0; row < len(buf)-nfs.height {
							row++
						}
					}
				case 'k':
					if col -= nfs.width; col < 0 {
						if col = 0; row > 0 {
							row--
						}
					}
				case 'q':
					return
				}
				nfs.page(buf, row, col, top, height, true)
			}
		}
	}
}

func (nfs *NFS) Read(p []byte) (n int, err error) {
	m := nfs.Context.Message()
	if !m.Caps("termbox") {
		return nfs.in.Read(p)
	}

	what := make([]rune, 0, 1024)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
			case termbox.KeyCtrlJ:
				b := []byte(string(what))
				n = len(b)
				copy(p, b)
				return
			default:
				m.Log("bench", "event %v", ev.Ch)
				what = append(what, ev.Ch)
			}
		default:
		}
	}

	return

	back := make([]rune, 0, 1024)
	rest := make([]rune, 0, 1024)
	buf := back

	history := nfs.Context.Message().Confv("history").([]interface{})
	his := len(history)

	tab := []string{}
	tabi := 0

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				termbox.Close()
				nfs.out = nil
				b := []byte("return\n")
				n = len(b)
				copy(p, b)
				return

				os.Exit(1)

			case termbox.KeyCtrlV:
				nfs.View(nfs.pages, 1, nfs.height)
				row, col := nfs.zone(nfs.pages, 1, nfs.height)
				nfs.page(nfs.pages, row, col, 1, nfs.Capi("cursor_pos"), false)

			case termbox.KeyCtrlL:
				nfs.escape("2J").escape("H")
				nfs.Cap("cursor_pos", "1")

			case termbox.KeyCtrlJ, termbox.KeyCtrlM:
				buf = append(buf, rest...)
				buf = append(buf, '\n')
				nfs.color("\n")

				b := []byte(string(buf))
				n = len(b)
				copy(p, b)
				return

			case termbox.KeyCtrlP:
				for i := 0; i < len(history); i++ {
					his = (his + len(history) - 1) % len(history)
					if strings.HasPrefix(history[his].(string), string(buf)) {
						rest = rest[:0]
						rest = append(rest, []rune(history[his].(string)[len(buf):])...)
						break
					}
				}

			case termbox.KeyCtrlN:
				for i := 0; i < len(history); i++ {
					his = (his + len(history) + 1) % len(history)
					if strings.HasPrefix(history[his].(string), string(buf)) {
						rest = rest[:0]
						rest = append(rest, []rune(history[his].(string)[len(buf):])...)
						break
					}
				}

			case termbox.KeyCtrlA:
				if len(buf) == 0 {
					continue
				}
				rest = nfs.insert(rest, buf)
				buf = buf[:0]

			case termbox.KeyCtrlE:
				if len(rest) == 0 {
					continue
				}
				buf = append(buf, rest...)
				rest = rest[:0]

			case termbox.KeyCtrlB:
				if len(buf) == 0 {
					continue
				}
				rest = nfs.insert(rest, []rune{buf[len(buf)-1]})
				buf = buf[:len(buf)-1]

			case termbox.KeyCtrlF:
				if len(rest) == 0 {
					continue
				}
				buf = append(buf, rest[0])
				rest = rest[1:]

			case termbox.KeyCtrlW:
				if len(buf) > 0 {
					c := buf[len(buf)-1]
					for len(buf) > 0 && unicode.IsSpace(c) && unicode.IsSpace(buf[len(buf)-1]) {
						buf = buf[:len(buf)-1]
					}

					for len(buf) > 0 && unicode.IsPunct(c) && unicode.IsPunct(buf[len(buf)-1]) {
						buf = buf[:len(buf)-1]
					}

					for len(buf) > 0 && unicode.IsLetter(c) && unicode.IsLetter(buf[len(buf)-1]) {
						buf = buf[:len(buf)-1]
					}

					for len(buf) > 0 && unicode.IsDigit(c) && unicode.IsDigit(buf[len(buf)-1]) {
						buf = buf[:len(buf)-1]
					}
				}
			case termbox.KeyCtrlH:
				if len(buf) == 0 {
					continue
				}
				buf = buf[:len(buf)-1]

			case termbox.KeyCtrlD:
				if len(rest) == 0 {
					continue
				}
				rest = rest[1:]

			case termbox.KeyCtrlU:
				if len(buf) > 0 {
					back = back[:0]
					back = append(back, buf...)
				}

				tab = tab[:0]

				buf = buf[:0]

			case termbox.KeyCtrlK:
				if len(rest) > 0 {
					back = append([]rune{}, rest...)
				}
				rest = rest[:0]

			case termbox.KeyCtrlY:
				buf = append(buf, back...)

			case termbox.KeyCtrlT:
				if l := len(buf); l > 1 {
					buf[l-1], buf[l-2] = buf[l-2], buf[l-1]
				}

			case termbox.KeyCtrlI:
				if len(tab) == 0 {
					tabi = 0
					// prefix := string(buf)
					// nfs.Message.BackTrace(func(m *ctx.Message) bool {
					// 	for k, _ := range m.Target().Commands {
					// 		if strings.HasPrefix(k, prefix) {
					// 			tab = append(tab, k[len(prefix):])
					// 		}
					// 	}
					// 	return true
					// }, nfs.Optionv("ps_target").(*ctx.Context))
				}

				if tabi >= 0 && tabi < len(tab) {
					rest = append(rest[:0], []rune(tab[tabi])...)
					tabi = (tabi + 1) % len(tab)
				}

			case termbox.KeySpace:
				tab = tab[:0]
				buf = append(buf, ' ')

				if len(rest) == 0 {
					nfs.color(" ")
					continue
				}

			default:
				tab = tab[:0]
				buf = append(buf, ev.Ch)
				if len(rest) == 0 {
					nfs.color(string(ev.Ch))
				}
			}
			nfs.prompt(string(buf), string(rest))
		}
	}
	return
}
func (nfs *NFS) prompt(arg ...string) string {
	m := nfs.Context.Message()
	target, _ := m.Optionv("ps_target").(*ctx.Context)
	nfs.out.WriteString(fmt.Sprintf("%d[%s]%s> ", m.Capi("ninput"), time.Now().Format("15:04:05"), target.Name))
	return ""
	return "> "

	ps := nfs.Option("prompt")
	if nfs.Caps("windows") {
		nfs.color(ps)
		return ps
	}
	line, rest := "", ""
	if len(arg) > 0 {
		line = arg[0]
	}
	if len(arg) > 1 {
		rest = arg[1]
	}

	if !nfs.Caps("windows") && len(nfs.pages) > 0 && nfs.width > 0 {
		for i := (len(nfs.pages[len(nfs.pages)-1]) - 1) / (nfs.width); i > 0; i-- {
			nfs.escape("2K").escape("A")
		}
		nfs.escape("2K").escape("G").escape("?25h")
	}

	if len(nfs.pages) > 0 {
		nfs.pages = nfs.pages[:len(nfs.pages)-1]
	}
	nfs.pages = append(nfs.pages, ps+line+rest+"\n")

	if nfs.color(ps, nfs.Confi("pscolor")).color(line).color(rest); len(rest) > 0 {
		nfs.escape("%dD", len(rest))
	}
	return ps
}
func (nfs *NFS) printf(arg ...interface{}) *NFS {
	for _, v := range arg {
		if nfs.io != nil {
			fmt.Fprint(nfs.io, kit.Format(v))
		} else if nfs.out != nil {
			nfs.out.WriteString(kit.Format(v))
		}
	}

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
	nfs.Message = m
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
		m.Cap("stream", arg[1])
		nfs.Caches["ninput"] = &ctx.Cache{Value: "0"}
		nfs.Caches["noutput"] = &ctx.Cache{Value: "0"}
		nfs.Configs["input"] = &ctx.Config{Value: []interface{}{}}
		nfs.Configs["output"] = &ctx.Config{Value: []interface{}{}}

		if nfs.in = m.Optionv("in").(*os.File); m.Has("out") {
			nfs.out = m.Optionv("out").(*os.File)
			if m.Cap("goos") != "windows" {
				termbox.Init()
				defer termbox.Close()
				nfs.Caches["termbox"] = &ctx.Cache{Value: "true"}
			}
		}

		line, bio := "", bufio.NewScanner(nfs)
		for nfs.prompt(); !m.Options("scan_end"); nfs.prompt() {
			for bio.Scan() {
				if line = line + bio.Text(); !strings.HasSuffix(line, "\\") {
					break
				}
				line = strings.TrimSuffix(line, "\\")
			}
			m.Confv("input", -2, map[string]interface{}{"time": time.Now().Unix(), "line": line})
			m.Log("debug", "%s %d %d [%s]", "input", m.Capi("ninput", 1), len(line), line)

			for i := m.Capi("ninput") - 1; i < m.Capi("ninput"); i++ {
				line = m.Conf("input", []interface{}{i, "line"})

				msg := m.Backs(m.Spawn(m.Source()).Set("detail", line).Set("option", "file_pos", i))

				lines := strings.Split(strings.Join(msg.Meta["result"], ""), "\n")
				for j := len(lines) - 1; j > 0; j-- {
					if strings.TrimSpace(lines[j]) != "" {
						break
					}
					lines = lines[:j]
				}
				for _, line := range lines {
					m.Confv("output", -2, map[string]interface{}{"time": time.Now().Unix(), "line": line})
					m.Log("debug", "%s %d %d [%s]", "output", m.Capi("noutput", 1), len(line), line)
					nfs.printf(line).printf("\n")
				}

				if msg.Appends("file_pos0") {
					i = msg.Appendi("file_pos0") - 1
					msg.Append("file_pos0", "")
				}
			}
			line = ""
		}

		if !m.Options("scan_end") {
			m.Backs(m.Spawn(m.Source()).Set("detail", "return"))
		}
		return true
	}

	nfs.Message = m
	if len(arg) > 0 && arg[0] == "scan" {
		// nfs.Caches["windows"] = &ctx.Cache{Name: "windows", Value: "false", Help: "termbox"}
		// nfs.Caches["termbox"] = &ctx.Cache{Name: "termbox", Value: "false", Help: "termbox"}
		// nfs.Caches["cursor_pos"] = &ctx.Cache{Name: "cursor_pos", Value: "1", Help: "termbox"}
		//
		// nfs.Configs["color"] = &ctx.Config{Name: "color", Value: "false", Help: "color"}
		// nfs.Configs["fgcolor"] = &ctx.Config{Name: "fgcolor", Value: "9", Help: "fgcolor"}
		// nfs.Configs["bgcolor"] = &ctx.Config{Name: "bgcolor", Value: "9", Help: "bgcolor"}
		// nfs.Configs["pscolor"] = &ctx.Config{Name: "pscolor", Value: "2", Help: "pscolor"}
		// nfs.Configs["statusfgcolor"] = &ctx.Config{Name: "statusfgcolor", Value: "1", Help: "pscolor"}
		// nfs.Configs["statusbgcolor"] = &ctx.Config{Name: "statusbgcolor", Value: "2", Help: "pscolor"}
		//

		// nfs.in = m.Optionv("in").(*os.File)
		// bio := bufio.NewScanner(nfs)
		//
		// s, e := nfs.in.Stat()
		// m.Assert(e)
		// m.Capi("size", int(s.Size()))

		if m.Cap("stream", arg[1]) == "stdio" {
			nfs.out = m.Optionv("out").(*os.File)
			nfs.width, nfs.height = 1, 1
			// if !m.Caps("windows", runtime.GOOS == "windows") {
			// 	termbox.Init()
			// 	defer termbox.Close()
			// 	nfs.width, nfs.height = termbox.Size()
			// 	nfs.Cap("termbox", "true")
			// 	nfs.Conf("color", "true")
			// }
			// if !m.Options("init.shy") {
			//
			// 	for _, v := range []string{
			// 		// "say you are so pretty",
			// 		"context web serve ./ :9094",
			// 	} {
			// 		m.Back(m.Spawn(m.Source()).Set("detail", v))
			// 	}
			// 	for _, v := range []string{
			// 		"say you are so pretty",
			// 		"context web brow 'http://localhost:9094'",
			// 	} {
			// 		nfs.history = append(nfs.history, v)
			// 		m.Capi("nline", 1)
			// 	}
			// 	for _, v := range []string{
			// 		"say you are so pretty\n",
			// 		"your can brow 'http://localhost:9094'\n",
			// 		"press \"brow\" then press Enter\n",
			// 	} {
			// 		nfs.print(fmt.Sprintf(v))
			// 	}
			// }
		}
	}

	m.Cap("stream", m.Option("stream"))
	nfs.io = m.Optionv("io").(io.ReadWriter)
	nfs.hand = map[int]*ctx.Message{}
	nfs.send = make(chan *ctx.Message, 10)
	nfs.recv = make(chan *ctx.Message, 10)

	go func() { //发送消息队列
		for {
			select {
			case msg := <-nfs.send:
				head, body := "detail", "option"
				if msg.Hand {
					head, body = "result", "append"
					send_code := msg.Option("send_code")
					msg.Append("send_code1", send_code)
					m.Log("info", "%s recv: %v %v", msg.Option("recv_code"), msg.Meta[head], msg.Meta[body])
				} else {
					msg.Option("send_code", m.Capi("nsend", 1))
					m.Log("info", "%d send: %v %v", m.Capi("nsend"), msg.Meta[head], msg.Meta[body])
					nfs.hand[m.Capi("nsend")] = msg
				}

				for _, v := range msg.Meta[head] {
					n, e := fmt.Fprintf(nfs.io, "%s: %s\n", head, url.QueryEscape(v))
					m.Assert(e)
					m.Capi("nwrite", n)
				}
				for _, k := range msg.Meta[body] {
					for _, v := range msg.Meta[k] {
						n, e := fmt.Fprintf(nfs.io, "%s: %s\n", url.QueryEscape(k), url.QueryEscape(v))
						m.Assert(e)
						m.Capi("nwrite", n)
					}
				}

				n, e := fmt.Fprintf(nfs.io, "\n")
				m.Assert(e)
				m.Capi("nwrite", n)
			}
		}
	}()

	go func() { //接收消息队列
		var e error
		var msg *ctx.Message
		head, body := "", ""

		for bio := bufio.NewScanner(nfs.io); bio.Scan(); {
			if msg == nil {
				msg = m.Sess("target")
			}
			if msg.Meta == nil {
				msg.Meta = map[string][]string{}
			}
			line := bio.Text()
			m.Log("recv", "(%s) %s", head, line)
			m.Capi("nread", len(line)+1)
			if len(line) == 0 {
				if head == "detail" {
					m.Log("info", "%d recv: %v %v %v", m.Capi("nrecv", 1), msg.Meta[head], msg.Meta[body], msg.Meta)
					msg.Option("recv_code", m.Cap("nrecv"))
					nfs.recv <- msg
				} else {
					m.Log("info", "%d send: %v %v %v", msg.Appendi("send_code1"), msg.Meta[head], msg.Meta[body], msg.Meta)
					h := nfs.hand[msg.Appendi("send_code1")]
					h.Copy(msg, "result").Copy(msg, "append")
					h.Remote <- true
				}
				msg, head, body = nil, "", "append"
				continue
			}

			word := strings.Split(line, ": ")
			word[0], e = url.QueryUnescape(word[0])
			m.Assert(e)
			word[1], e = url.QueryUnescape(word[1])
			m.Assert(e)
			switch word[0] {
			case "detail":
				head, body = "detail", "option"
				msg.Add(word[0], word[1])
			case "result":
				head, body = "result", "append"
				msg.Add(word[0], word[1])
			default:
				msg.Add(body, word[0], word[1])
			}
		}
	}()

	for {
		select {
		case msg := <-nfs.recv:
			nfs.send <- msg.Cmd()
		}
	}

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
	return true
}

var FileNotExist = errors.New("file not exist")
var Index = &ctx.Context{Name: "nfs", Help: "存储中心",
	Caches: map[string]*ctx.Cache{
		"nfile": &ctx.Cache{Name: "nfile", Value: "0", Help: "已经打开的文件数量"},
	},
	Configs: map[string]*ctx.Config{
		"term_simple": &ctx.Config{Name: "term_simple", Value: "false", Help: "二维码的默认大小"},
		"qr_size":     &ctx.Config{Name: "qr_size", Value: "256", Help: "二维码的默认大小"},

		"pscolor": &ctx.Config{Name: "pscolor", Value: "2", Help: "pscolor"},

		"buf_size": &ctx.Config{Name: "buf_size", Value: "1024", Help: "读取文件的缓存区的大小"},
		"dir_conf": &ctx.Config{Name: "dir_conf", Value: map[string]interface{}{
			"dir_root": "usr",
		}, Help: "读取文件的缓存区的大小"},

		"dir_type":   &ctx.Config{Name: "dir_type(file/dir/all)", Value: "all", Help: "dir命令输出的文件类型, file: 只输出普通文件, dir: 只输出目录文件, 否则输出所有文件"},
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
		"paths": &ctx.Config{Name: "paths", Value: []interface{}{"var", "usr", "etc", ""}, Help: "文件路径"},
	},
	Commands: map[string]*ctx.Command{
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
					p := path.Join(value, m.Option("dir_root"), kit.Select("", arg))
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

		"path": &ctx.Command{Name: "path file", Help: "查找文件路径", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
			if p, f, e := open(m, arg[0]); m.Assert(e) {
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
			if p, f, e := open(m, m.Format(arg[0]), os.O_WRONLY|os.O_CREATE|os.O_TRUNC); m.Assert(e) {
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

			m.Option("hi", "hello world")
			m.Option("he", "hello", "world")

			m.Append("he", "hello", "world")
			m.Append("hi", "nice", "job")

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
					m.Put("option", "in", f).Start(fmt.Sprintf("file%s", m.Capi("nfile")), help, key, p)
				}
			}
			return
		}},
		"prompt": &ctx.Command{Name: "prompt arg", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.out != nil {
				nfs.prompt()
				for _, v := range arg {
					nfs.printf(v)
					m.Echo(v)
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
		"exec": &ctx.Command{Name: "exec cmd", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.out != nil {
				nfs.prompt()
				for _, v := range arg {
					nfs.out.WriteString(v)
				}
				nfs.out.WriteString("\n")

				msg := m.Find("cli.shell1").Cmd("source", arg)
				for _, v := range msg.Meta["result"] {
					nfs.out.WriteString(v)
					m.Echo(v)
				}
				nfs.out.WriteString("\n")
			}
			return
		}},

		"listen": &ctx.Command{Name: "listen args...", Help: "启动文件服务, args: 参考tcp模块, listen命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if _, ok := m.Target().Server.(*NFS); m.Assert(ok) { //{{{
				m.Sess("tcp").Call(func(sub *ctx.Message) *ctx.Message {
					sub.Start(fmt.Sprintf("file%d", m.Capi("nfile", 1)), "远程文件")
					return sub.Sess("target", m.Source()).Call(func(sub1 *ctx.Message) *ctx.Message {
						nfs, _ := sub.Target().Server.(*NFS)
						sub1.Remote = make(chan bool, 1)
						nfs.send <- sub1
						<-sub1.Remote
						return nil
					})
				}, m.Meta["detail"])
			}
			return
		}},
		"dial": &ctx.Command{Name: "dial args...", Help: "连接文件服务, args: 参考tcp模块, dial命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if _, ok := m.Target().Server.(*NFS); m.Assert(ok) {
				m.Sess("tcp").Call(func(sub *ctx.Message) *ctx.Message {
					sub.Start(fmt.Sprintf("file%d", m.Capi("nfile", 1)), "远程文件")
					return sub.Sess("target", m.Source()).Call(func(sub1 *ctx.Message) *ctx.Message {
						nfs, _ := sub.Target().Server.(*NFS)
						sub1.Remote = make(chan bool, 1)
						nfs.send <- sub1
						<-sub1.Remote
						return nil
					})
				}, m.Meta["detail"])
			}

			return
		}},
		"send": &ctx.Command{Name: "send [file] args...", Help: "连接文件服务, args: 参考tcp模块, dial命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.io != nil {
				m.Remote = make(chan bool, 1)
				nfs.send <- m
				<-m.Remote
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
