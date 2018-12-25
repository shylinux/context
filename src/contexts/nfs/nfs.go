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
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type NFS struct {
	in     *os.File
	out    *os.File
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

func (nfs *NFS) prompt(arg ...string) string {
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
	if nfs.Caps("windows") || !nfs.Caps("termbox") || nfs.Confs("term_simple") {
		return nfs.in.Read(p)
	}

	buf := make([]rune, 0, 1024)
	rest := make([]rune, 0, 1024)

	back := buf

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
					prefix := string(buf)
					nfs.Message.BackTrace(func(m *ctx.Message) bool {
						for k, _ := range m.Target().Commands {
							if strings.HasPrefix(k, prefix) {
								tab = append(tab, k[len(prefix):])
							}
						}
						return true
					}, nfs.Optionv("ps_target").(*ctx.Context))
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

func (nfs *NFS) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	if len(arg) > 0 && (arg[0] == "scan" || arg[0] == "open" || arg[0] == "append") {
		c.Caches = map[string]*ctx.Cache{
			"pos":    &ctx.Cache{Name: "pos", Value: "0", Help: "pos"},
			"size":   &ctx.Cache{Name: "size", Value: "0", Help: "size"},
			"nread":  &ctx.Cache{Name: "nread", Value: "0", Help: "nread"},
			"nwrite": &ctx.Cache{Name: "nwrite", Value: "0", Help: "nwrite"},
			"nline":  &ctx.Cache{Name: "缓存命令行数", Value: "0", Help: "缓存命令行数"},
		}
		c.Configs = map[string]*ctx.Config{
			"history": &ctx.Config{Name: "history", Value: []interface{}{}, Help: "读取记录"},
		}
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
	nfs.width, nfs.height = 1, 1
	return nfs
}
func (nfs *NFS) Start(m *ctx.Message, arg ...string) bool {
	nfs.Message = m
	if len(arg) > 0 && arg[0] == "scan" {
		nfs.Caches["windows"] = &ctx.Cache{Name: "windows", Value: "false", Help: "termbox"}
		nfs.Caches["termbox"] = &ctx.Cache{Name: "termbox", Value: "false", Help: "termbox"}
		nfs.Caches["cursor_pos"] = &ctx.Cache{Name: "cursor_pos", Value: "1", Help: "termbox"}

		nfs.Configs["color"] = &ctx.Config{Name: "color", Value: "false", Help: "color"}
		nfs.Configs["fgcolor"] = &ctx.Config{Name: "fgcolor", Value: "9", Help: "fgcolor"}
		nfs.Configs["bgcolor"] = &ctx.Config{Name: "bgcolor", Value: "9", Help: "bgcolor"}
		nfs.Configs["pscolor"] = &ctx.Config{Name: "pscolor", Value: "2", Help: "pscolor"}
		nfs.Configs["statusfgcolor"] = &ctx.Config{Name: "statusfgcolor", Value: "1", Help: "pscolor"}
		nfs.Configs["statusbgcolor"] = &ctx.Config{Name: "statusbgcolor", Value: "2", Help: "pscolor"}

		nfs.in = m.Optionv("in").(*os.File)
		bio := bufio.NewScanner(nfs)
		s, e := nfs.in.Stat()
		m.Assert(e)
		m.Capi("size", int(s.Size()))

		if m.Cap("stream", arg[1]) == "stdio" {
			nfs.out = m.Optionv("out").(*os.File)
			if !m.Caps("windows", runtime.GOOS == "windows") {
				termbox.Init()
				defer termbox.Close()
				nfs.width, nfs.height = termbox.Size()
				nfs.Cap("termbox", "true")
				nfs.Conf("color", "true")
			}
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

		line := ""
		for nfs.prompt(); !m.Options("scan_end") && bio.Scan(); nfs.prompt() {
			text := bio.Text()
			m.Capi("nread", len(text)+1)

			if line += text; len(text) > 0 && text[len(text)-1] == '\\' {
				line = line[:len(line)-1]
				continue
			}
			m.Capi("nline", 1)
			m.Confv("history", -2, line)
			history := m.Confv("history").([]interface{})

			for i := len(history) - 1; i < len(history); i++ {
				line = history[i].(string)

				msg := m.Spawn(m.Source()).Set("detail", line)
				msg.Option("file_pos", i)
				m.Back(msg)

				for _, v := range msg.Meta["result"] {
					m.Capi("nwrite", len(v))
					nfs.print(v)
				}
				if msg.Append("file_pos0") != "" {
					i = msg.Appendi("file_pos0") - 1
					msg.Append("file_pos0", "")
				}
			}
			line = ""
		}
		if !m.Options("scan_end") {
			msg := m.Spawn(m.Source()).Set("detail", "return")
			m.Back(msg)
		}
		return true
	}

	if len(arg) > 0 && (arg[0] == "open" || arg[0] == "append") {
		nfs.out = m.Optionv("out").(*os.File)
		nfs.in = m.Optionv("in").(*os.File)
		s, e := nfs.in.Stat()
		m.Assert(e)
		m.Capi("size", int(s.Size()))
		m.Cap("stream", arg[1])
		if arg[0] == "append" {
			m.Capi("pos", int(s.Size()))
		}
		return false
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
			nfs.in.Close()
			nfs.in = nil
		}
		if nfs.out != nil {
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
		"nfile": &ctx.Cache{Name: "nfile", Value: "-1", Help: "已经打开的文件数量"},
	},
	Configs: map[string]*ctx.Config{
		"term_simple": &ctx.Config{Name: "term_simple", Value: "false", Help: "二维码的默认大小"},
		"qr_size":     &ctx.Config{Name: "qr_size", Value: "256", Help: "二维码的默认大小"},

		"pscolor": &ctx.Config{Name: "pscolor", Value: "2", Help: "pscolor"},
		"nfs_name": &ctx.Config{Name: "nfs_name", Value: "file", Help: "默认模块命名", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
			if len(arg) > 0 {
				return arg[0]
			}
			return fmt.Sprintf("%s%d", x.Value, m.Capi("nfile", 1))

		}},
		"nfs_help": &ctx.Config{Name: "nfs_help", Value: "file", Help: "默认模块帮助"},

		"buf_size": &ctx.Config{Name: "buf_size", Value: "1024", Help: "读取文件的缓存区的大小"},
		"dir_conf": &ctx.Config{Name: "dir_conf", Value: map[string]interface{}{
			"dir_root": "usr",
		}, Help: "读取文件的缓存区的大小"},

		"dir_type":   &ctx.Config{Name: "dir_type(file/dir/all)", Value: "all", Help: "dir命令输出的文件类型, file: 只输出普通文件, dir: 只输出目录文件, 否则输出所有文件"},
		"dir_name":   &ctx.Config{Name: "dir_name(name/tree/path/full)", Value: "name", Help: "dir命令输出文件名的类型, name: 文件名, tree: 带缩进的文件名, path: 相对路径, full: 绝对路径"},
		"dir_fields": &ctx.Config{Name: "dir_fields(time/type/name/size/line/hash)", Value: "time size line filename", Help: "dir命令输出文件名的类型, name: 文件名, tree: 带缩进的文件名, path: 相对路径, full: 绝对路径"},

		"git_branch":   &ctx.Config{Name: "git_branch", Value: "--list", Help: "版本控制状态参数"},
		"git_status":   &ctx.Config{Name: "git_status", Value: "-sb", Help: "版本控制状态参数"},
		"git_diff":     &ctx.Config{Name: "git_diff", Value: "--stat", Help: "版本控制状态参数"},
		"git_log":      &ctx.Config{Name: "git_log", Value: "--pretty=%h %an(%ad) %s  --date=format:%m/%d %H:%M  --graph", Help: "版本控制状态参数"},
		"git_log_form": &ctx.Config{Name: "git_log", Value: "stat", Help: "版本控制状态参数"},
		"git_log_skip": &ctx.Config{Name: "git_log", Value: "0", Help: "版本控制状态参数"},
		"git_log_line": &ctx.Config{Name: "git_log", Value: "3", Help: "版本控制状态参数"},
		"git_path":     &ctx.Config{Name: "git_path", Value: ".", Help: "版本控制默认路径"},
		"git_info":     &ctx.Config{Name: "git_info", Value: "branch status diff log", Help: "命令集合"},

		"paths": &ctx.Config{Name: "paths", Value: []interface{}{"var", "usr", "etc", ""}, Help: "文件路径"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen args...", Help: "启动文件服务, args: 参考tcp模块, listen命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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

		}},
		"dial": &ctx.Command{Name: "dial args...", Help: "连接文件服务, args: 参考tcp模块, dial命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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

		}},
		"send": &ctx.Command{Name: "send [file] args...", Help: "连接文件服务, args: 参考tcp模块, dial命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.io != nil {
				m.Remote = make(chan bool, 1)
				nfs.send <- m
				<-m.Remote
			}
		}},

		"scan": &ctx.Command{Name: "scan file name", Help: "扫描文件, file: 文件名, name: 模块名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*NFS); m.Assert(ok) {
				help := fmt.Sprintf("scan %s", arg[0])

				if arg[0] == "stdio" {
					m.Optionv("in", os.Stdin)
					m.Optionv("out", os.Stdout)
					m.Start(arg[0], help, key, arg[0])
					return
				}

				if p, f, e := open(m, arg[0]); m.Assert(e) {
					m.Optionv("in", f)
					m.Start(m.Confx("nfs_name", arg, 1), help, key, p)
				}
			}
		}},
		"prompt": &ctx.Command{Name: "prompt arg", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.out != nil {
				nfs.prompt()
				for _, v := range arg {
					nfs.out.WriteString(v)
					m.Echo(v)
				}
			}
		}},
		"exec": &ctx.Command{Name: "exec cmd", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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
		}},
		"show": &ctx.Command{Name: "show arg", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.out != nil {
				for _, v := range arg {
					nfs.out.WriteString(v)
					m.Echo(v)
				}
			}
		}},

		"open": &ctx.Command{Name: "open file name", Help: "打开文件, file: 文件名, name: 模块名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			file := arg[0]
			if m.Has("io") {
			} else if p, f, e := open(m, file, os.O_RDWR|os.O_CREATE); e == nil {
				m.Put("option", "in", f).Put("option", "out", f)
				file = p
			} else {
				return
			}

			m.Start(m.Confx("nfs_name", arg, 1), fmt.Sprintf("file %s", file), "open", file)
			m.Echo(file)
		}},
		"read": &ctx.Command{Name: "read [buf_size [pos]]", Help: "读取文件, buf_size: 读取大小, pos: 读取位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.in != nil {
				n, e := strconv.Atoi(m.Confx("buf_size", arg, 0))
				m.Assert(e)

				if len(arg) > 1 {
					m.Cap("pos", arg[1])
				}

				buf := make([]byte, n)
				if n, e = nfs.in.ReadAt(buf, int64(m.Capi("pos"))); e != io.EOF {
					m.Assert(e)
				}
				m.Capi("nread", n)
				m.Echo(string(buf))

				if m.Capi("pos", n); n == 0 {
					m.Cap("pos", "0")
				}
			}
		}},
		"write": &ctx.Command{Name: "write string [pos]", Help: "写入文件, string: 写入内容, pos: 写入位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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
		}},

		"load": &ctx.Command{Name: "load file [buf_size [pos]]", Help: "加载文件, buf_size: 加载大小, pos: 加载位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if p, f, e := open(m, arg[0]); m.Assert(e) {
				defer f.Close()

				pos := 0
				if len(arg) > 2 {
					i, e := strconv.Atoi(arg[2])
					m.Assert(e)
					pos = i
				}

				s, e := strconv.Atoi(m.Confx("buf_size", arg, 1))
				m.Assert(e)
				buf := make([]byte, s)

				if l, e := f.ReadAt(buf, int64(pos)); e == io.EOF || m.Assert(e) {
					m.Log("info", "load %s %d %d", p, l, pos)
					m.Echo(string(buf[:l]))
				}
			}
		}},
		"save": &ctx.Command{Name: "save file string...", Help: "保存文件, file: 保存的文件, string: 保存的内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if len(arg) == 1 && m.Has("data") {
				arg = append(arg, m.Option("data"))
			}
			if p, f, e := open(m, arg[0], os.O_WRONLY|os.O_CREATE|os.O_TRUNC); m.Assert(e) {
				defer f.Close()
				m.Append("directory", p)
				m.Echo(p)

				for _, v := range arg[1:] {
					n, e := fmt.Fprint(f, v)
					m.Assert(e)
					m.Log("info", "save %s %d", p, n)
				}
			}
		}},
		"print": &ctx.Command{Name: "print file string...", Help: "输出文件, file: 输出的文件, string: 输出的内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if p, f, e := open(m, arg[0], os.O_WRONLY|os.O_CREATE|os.O_APPEND); m.Assert(e) {
				defer f.Close()

				for _, v := range arg[1:] {
					n, e := fmt.Fprint(f, v)
					m.Assert(e)
					m.Log("info", "print %s %d", p, n)
				}
			}
		}},
		"export": &ctx.Command{Name: "export filename", Help: "导出数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			name := time.Now().Format(arg[0])
			_, f, e := open(m, name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
			m.Assert(e)
			defer f.Close()

			switch {
			case strings.HasSuffix(arg[0], ".json") && len(m.Meta["append"]) > 0:
				data := []interface{}{}

				nrow := len(m.Meta[m.Meta["append"][0]])
				for i := 0; i < nrow; i++ {
					line := map[string]interface{}{}
					for _, k := range m.Meta["append"] {
						line[k] = m.Meta[k][i]
					}
					data = append(data, line)
				}
				en := json.NewEncoder(f)
				en.SetIndent("", "  ")
				en.Encode(data)

			case strings.HasSuffix(arg[0], ".csv") && len(m.Meta["append"]) > 0:
				w := csv.NewWriter(f)

				line := []string{}
				for _, v := range m.Meta["append"] {
					line = append(line, v)
				}
				w.Write(line)

				nrow := len(m.Meta[m.Meta["append"][0]])
				for i := 0; i < nrow; i++ {
					line := []string{}
					for _, k := range m.Meta["append"] {
						line = append(line, m.Meta[k][i])
					}
					w.Write(line)
				}
				w.Flush()
			default:
				for _, v := range m.Meta["result"] {
					f.WriteString(v)
				}
			}
			m.Set("append").Set("result").Add("append", "directory", name).Echo(name)
		}},
		"import": &ctx.Command{Name: "import filename [index]", Help: "导入数据", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			_, f, e := open(m, arg[0])
			m.Assert(e)
			defer f.Close()

			switch {
			case strings.HasSuffix(arg[0], ".json"):
				var data interface{}
				de := json.NewDecoder(f)
				de.Decode(&data)

				msg := m.Spawn().Put("option", "data", data).Cmd("trans", "data", arg[1:])
				m.Copy(msg, "append").Copy(msg, "result")
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
			}
		}},

		"pwd": &ctx.Command{Name: "pwd [all] | [[index] path] ", Help: "工作目录，all: 查看所有, index path: 设置路径, path: 设置当前路径", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if len(arg) > 0 && arg[0] == "all" {
				list := m.Confv("paths").([]interface{})
				for i, v := range list {
					m.Add("append", "index", i)
					m.Add("append", "path", v)
				}
				m.Table()
				return
			} else if len(arg) > 1 {
				m.Log("info", "paths %s %s", arg[0], arg[1])
				m.Confv("paths", arg[0], arg[1])
			} else if len(arg) > 0 {
				m.Log("info", "paths 0 %s", arg[0])
				m.Confv("paths", 0, arg[0])
			}

			p := m.Confv("paths", 0).(string)
			if path.IsAbs(p) {
				m.Echo("%s", p)
				return
			}

			wd, e := os.Getwd()
			m.Assert(e)
			m.Echo("%s", path.Join(wd, p))
		}},
		"path": &ctx.Command{Name: "path file", Help: "查找文件路径", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			for _, v := range m.Confv("paths").([]interface{}) {
				p := path.Join(v.(string), arg[0])
				if _, e := os.Stat(p); e == nil {
					m.Echo(p)
					break
				}
			}
		}},
		"json": &ctx.Command{Name: "json [key value]...", Help: "生成格式化内容, key: 参数名, value: 参数值", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if len(arg) == 1 {
				var data interface{}
				json.Unmarshal([]byte(arg[0]), &data)

				buf, e := json.MarshalIndent(data, "", "  ")
				m.Assert(e)
				m.Echo(string(buf))
				return
			}

			if len(arg) > 1 && arg[0] == "file" {
				var data interface{}
				f, e := os.Open(arg[1])
				m.Assert(e)
				d := json.NewDecoder(f)
				d.Decode(&data)

				buf, e := json.MarshalIndent(data, "", "  ")
				m.Assert(e)
				m.Echo(string(buf))
				return
			}

			data := map[string]interface{}{}
			for _, k := range m.Meta["option"] {
				if v, ok := m.Data[k]; ok {
					data[k] = v
					continue
				}
				data[k] = m.Meta[k]
			}

			for i := 1; i < len(arg)-1; i += 2 {
				data[arg[i]] = arg[i+1]
			}

			buf, e := json.Marshal(data)
			m.Assert(e)
			m.Echo(string(buf))

		}},
		"genqr": &ctx.Command{Name: "genqr [qr_size size] filename string...", Help: "生成二维码图片, qr_size: 图片大小, filename: 文件名, string: 输出内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if size, e := strconv.Atoi(m.Confx("qr_size")); m.Assert(e) {
				p := path.Join(m.Confv("paths", 0).(string), arg[0])
				qrcode.WriteFile(strings.Join(arg[1:], ""), qrcode.Medium, size, p)
				m.Log("info", "genqr %s", p)
				m.Append("directory", p)
			}
		}},

		"dir": &ctx.Command{Name: "dir dir [dir_type both|file|dir] [dir_deep] fields...",
			Help: "查看目录, dir: 目录名, dir_type: 文件类型, dir_deep: 递归查询, fields: 查询字段",
			Form: map[string]int{"dir_reg": 1, "dir_type": 1, "dir_deep": 0, "dir_sort": 2},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				wd, e := os.Getwd()
				m.Assert(e)
				trip := len(wd) + 1

				if len(arg) == 0 {
					arg = append(arg, "")
				}
				dirs := arg[0]
				if m.Options("dir_root") {
					dirs = path.Join(m.Option("dir_root"), dirs)
				}

				rg, e := regexp.Compile(m.Option("dir_reg"))

				for _, v := range m.Confv("paths").([]interface{}) {
					d := path.Join(v.(string), dirs)
					if s, e := os.Stat(d); e == nil {
						if s.IsDir() {
							dir(m, d, 0, ctx.Right(m.Has("dir_deep")), m.Confx("dir_type"), trip, rg,
								strings.Split(m.Confx("dir_fields", strings.Join(arg[1:], " ")), " "),
								m.Conf("time_format"))
						} else {
							m.Append("directory", d)
							return
						}
						break
					}
				}
				if m.Has("dir_sort") {
					m.Sort(m.Meta["dir_sort"][1], m.Meta["dir_sort"][0])
				}

				if len(m.Meta["append"]) == 1 {
					for _, v := range m.Meta[m.Meta["append"][0]] {
						m.Echo(v).Echo(" ")
					}
				} else {
					m.Table()
				}
			}},
		"git": &ctx.Command{
			Name: "git branch|status|diff|log|info arg... [dir path]...",
			Help: "版本控制, branch: 分支管理, status: 查看状态, info: 查看分支与状态, dir: 指定路径",
			Form: map[string]int{"dir": 1, "git_info": 1, "git_log": 1, "git_log_form": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if len(arg) == 0 {
					arg = []string{"info"}
				}
				cmds := []string{arg[0]}
				switch arg[0] {
				case "s":
					arg[0] = "status"
				case "b":
					arg[0] = "branch"
				case "d":
					arg[0] = "diff"
				}
				if arg[0] == "info" {
					cmds = strings.Split(m.Confx("git_info"), " ")
				}
				wd, e := os.Getwd()
				m.Assert(e)
				if !m.Has("dir") {
					m.Option("dir", m.Confx("dir"))
				}
				for _, p := range m.Meta["dir"] {
					if !path.IsAbs(p) {
						p = path.Join(wd, p)
					}
					m.Echo("path: %s\n", p)
					for _, c := range cmds {
						args := []string{}
						switch c {
						case "branch", "status", "diff":
							if c != "status" {
								args = append(args, "--color")
							}
							args = append(args, strings.Split(m.Confx("git_"+c, arg, 1), "  ")...)
							if len(arg) > 2 {
								args = append(args, arg[2:]...)
							}
						case "difftool":
							cmd := exec.Command("git", "difftool", "-y")
							m.Log("info", "cmd: %s %v", "git", "difftool", "-y")
							cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
							if e := cmd.Start(); e != nil {
								m.Echo("error: ")
								m.Echo("%s\n", e)
							} else if e := cmd.Wait(); e != nil {
								m.Echo("error: ")
								m.Echo("%s\n", e)
							}
							continue
						case "csv":
							cmd := exec.Command("git", "log", "--shortstat", "--pretty=commit: %ad", "--date=format:%Y-%m-%d")
							if out, e := cmd.CombinedOutput(); e != nil {
								m.Echo("error: ")
								m.Echo("%s\n", e)
							} else {
								f, e := os.Create(arg[1])
								m.Assert(e)
								defer f.Close()

								type stat struct {
									date string
									adds int
									dels int
								}
								stats := []*stat{}
								list := strings.Split(string(out), "commit: ")
								for _, v := range list {
									l := strings.Split(v, "\n")
									if len(l) > 2 {
										fs := strings.Split(strings.Trim(l[2], " "), ", ")
										stat := &stat{date: l[0]}
										if len(fs) > 2 {
											adds := strings.Split(fs[1], " ")
											dels := strings.Split(fs[2], " ")
											a, e := strconv.Atoi(adds[0])
											m.Assert(e)
											stat.adds = a
											d, e := strconv.Atoi(dels[0])
											m.Assert(e)
											stat.dels = d
										} else {
											adds := strings.Split(fs[1], " ")
											a, e := strconv.Atoi(adds[0])
											m.Assert(e)
											if adds[1] == "insertions(+)" {
												stat.adds = a
											} else {
												stat.dels = a
											}
										}

										stats = append(stats, stat)
									}
								}

								fmt.Fprintf(f, "order,date,adds,dels,sum,top,bottom,last\n")
								l := len(stats)
								for i := 0; i < l/2; i++ {
									stats[i], stats[l-i-1] = stats[l-i-1], stats[i]
								}
								sum := 0
								for i, v := range stats {
									fmt.Fprintf(f, "%d,%s,%d,%d,%d,%d,%d,%d\n", i, v.date, v.adds, v.dels, sum, sum+v.adds, sum-v.dels, sum+v.adds-v.dels)
									sum += v.adds - v.dels
								}
							}
							continue

						case "log":
							args = append(args, "--color")
							args = append(args, strings.Split(m.Confx("git_log"), "  ")...)
							args = append(args, fmt.Sprintf("--%s", m.Confx("git_log_form")))
							args = append(args, m.Confx("git_log_skip", arg, 1, "--skip=%s"))
							args = append(args, m.Confx("git_log_line", arg, 2, "-n %s"))
						default:
							args = append(args, arg[1:]...)
						}

						switch c {
						case "commit":
							m.Find("web.code").Cmd("counter", "ncommit", 1)
						case "push":
							m.Find("web.code").Cmd("counter", "npush", 1)
						}

						m.Log("info", "cmd: %s %v", "git", ctx.Trans("-C", p, c, args))
						msg := m.Sess("cli").Cmd("system", "git", "-C", p, c, args)
						m.Copy(msg, "result").Copy(msg, "append")
						m.Echo("\n")
					}
				}
			}},
	},
}

func init() {
	nfs := &NFS{}
	nfs.Context = Index
	ctx.Index.Register(Index, nfs)
}
