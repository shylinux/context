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
	recv chan *ctx.Message
	hand map[int]*ctx.Message

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

func (nfs *NFS) Term(msg *ctx.Message, action string, args ...interface{}) *NFS {
	m := nfs.Context.Message()
	m.Log("term", "%s %v", action, args)

	switch action {
	case "init":
		termbox.Init()
		termbox.SetInputMode(termbox.InputEsc)
		termbox.SetInputMode(termbox.InputMouse)
		m.Cap("termbox", "true")
	}

	width, height := termbox.Size()
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
	end_row := m.Confi("term", "end_row")
	end_col := m.Confi("term", "end_col")

	switch action {
	case "init":
		m.Conf("term", "width", width)
		m.Conf("term", "height", height)
		m.Conf("term", "left", 0)
		m.Conf("term", "top", 0)
		m.Conf("term", "right", width)
		m.Conf("term", "bottom", height)
	case "exit":
		m.Cap("termbox", "false")
		termbox.Close()
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

		if m.Caps("termbox") {
			for x := left; x < right; x++ {
				for y := top; y < bottom; y++ {
					termbox.SetCell(x, y, ' ', fg, bg)
				}
			}
			m.Conf("term", "cursor_x", left)
			m.Conf("term", "cursor_y", top)
			termbox.SetCursor(left, top)
			termbox.Flush()
		}
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
		fallthrough
	case "cursor":
		m.Conf("term", "cursor_x", kit.Format(args[0]))
		m.Conf("term", "cursor_y", kit.Format(args[1]))
		termbox.SetCursor(m.Confi("term", "cursor_x"), m.Confi("term", "cursor_y"))
		fallthrough
	case "flush":
		termbox.Flush()
	case "delete":
		if m.Caps("termbox") {
			limit := left
			if len(args) > 0 {
				limit = x - kit.Int(args[0])
			}

			for x = x; x >= limit; x-- {
				termbox.SetCell(x, y, ' ', fg, bg)
			}
			m.Conf("term", "cursor_x", x+1)
			m.Conf("term", "cursor_y", y)
			termbox.SetCursor(x+1, y)
			termbox.Flush()
		}
	case "scroll":
		n := 1
		if len(args) > 0 {
			n = kit.Int(args[0])
		}
		m.Log("term", "<<<< scroll page (%v, %v) (%v, %v)", begin_row, begin_col, end_row, end_col)

		// 向下滚动
		for i := begin_row; n > 0 && i < end_row; i++ {
			line := []rune(m.Conf("output", []interface{}{i, "line"}))

			for j, l := begin_col, left; n > 0; j, l = j+1, l+1 {
				if j >= len(line)-1 {
					begin_row, begin_col = i+1, 0
					n--
					break
				} else if l >= right-1 && !m.Confs("term", "wrap") {
					begin_row, begin_col = i, j
					n--

				} else if line[j] == '\n' || (l >= right-1 && m.Confs("term", "wrap")) {
					begin_row, begin_col = i, j+1
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
				} else if line[j-1] == '\n' || l <= left {
					begin_row, begin_col = i, j
					n++
				}
			}
		}

		m.Log("term", ">>>> scroll page (%v, %v) (%v, %v)", begin_row, begin_col, end_row, end_col)
		m.Conf("term", "begin_row", begin_row)
		m.Conf("term", "begin_col", begin_col)

		fallthrough
	case "refresh":
		m.Log("term", "<<<< refresh page (%v, %v) (%v, %v)", begin_row, begin_col, end_row, end_col)

		noutput := m.Capi("noutput")

		nfs.Term(m, "clear", "all")
		for i := begin_row; i < noutput; i++ {
			line := m.Conf("output", []interface{}{i, "line"})
			m.Conf("term", "end_row", i)
			m.Conf("term", "end_col", 0)
			if begin_col < len(line) {
				nfs.Term(m, "print", line[begin_col:])
			}
		}

		end_row = m.Confi("term", "end_row")
		end_col = m.Confi("term", "end_col")
		m.Log("term", ">>>> refresh page (%v, %v) (%v, %v)", begin_row, begin_col, end_row, end_col)
		nfs.prompt()

	case "print":
		m.Log("term", "page (%v, %v) (%v, %v)", begin_row, begin_col, end_row, end_col)
		if m.Caps("termbox") {
			for _, v := range kit.Format(args...) {
				if x < right && y < bottom {
					termbox.SetCell(x, y, v, fg, bg)
					end_col++
				}

				if v > 255 {
					x++
				}
				if x++; v == '\n' || (x >= right && m.Confs("term", "wrap")) {
					x, y = left, y+1
				}

				if x < right && y < bottom {
					m.Conf("term", "cursor_x", x)
					m.Conf("term", "cursor_y", y)
					termbox.SetCursor(x, y)
				}

				if y >= bottom {
					break
				}
			}
			termbox.Flush()
		}
		m.Conf("term", "end_col", end_col)
		m.Log("term", "page (%v, %v) (%v, %v)", begin_row, begin_col, end_row, end_col)
	}
	return nfs
}
func (nfs *NFS) Read(p []byte) (n int, err error) {
	m := nfs.Context.Message()
	if !m.Caps("termbox") {
		return nfs.in.Read(p)
	}

	what := make([]rune, 0, 1024)
	which := m.Capi("ninput")

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventInterrupt:
		case termbox.EventResize:
			nfs.Term(m, "resize")
		case termbox.EventMouse:
			switch ev.Key {
			case termbox.MouseLeft:
				nfs.Term(m, "window", ev.MouseX, ev.MouseY)
				nfs.prompt()
			case termbox.MouseMiddle:
			case termbox.MouseRight:
				nfs.Term(m, "resize", ev.MouseX, ev.MouseY)
			case termbox.MouseRelease:
			case termbox.MouseWheelUp:
				nfs.Term(m, "scroll", -ev.N)
			case termbox.MouseWheelDown:
				nfs.Term(m, "scroll", ev.N)
			}
		case termbox.EventError:
		case termbox.EventNone:
		case termbox.EventRaw:
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlP: // termbox.KeyBackspace
				if which--; which < 0 {
					which = m.Capi("ninput") - 1
				}
				if v := m.Conf("input", []interface{}{which, "line"}); v != "" {
					what = []rune(v)
					m.Log("term", "what %v %v", which, what)
					nfs.prompt(what)
				}

			case termbox.KeyCtrlN: // termbox.KeyBackspace
				if which++; which >= m.Capi("ninput") {
					which = 0
				}
				if v := m.Conf("input", []interface{}{which, "line"}); v != "" {
					what = []rune(v)
					m.Log("term", "what %v %v", which, what)
					nfs.prompt(what)
				}

			case termbox.KeyCtrlH: // termbox.KeyBackspace
				if len(what) > 0 {
					what = what[:len(what)-1]
				}
				nfs.prompt(what)

			case termbox.KeyCtrlU:
				what = what[:0]
				nfs.prompt(what)

			case termbox.KeyCtrlL:
				nfs.Term(m, "window", m.Confi("term", "left"), m.Confi("term", "top"))
				nfs.prompt(what)

			case termbox.KeyCtrlJ, termbox.KeyEnter:
				what = append(what, '\n')
				n = copy(p, []byte(string(what)))
				return

			case termbox.KeyCtrlC:
				nfs.Term(m, "exit")
				return

			case termbox.KeyCtrlE:
				nfs.Term(m, "scroll")
			case termbox.KeyCtrlY:
				nfs.Term(m, "scroll", -1)

			case termbox.KeyTab:
				what = append(what, '\t')
				nfs.Term(m, "print", "\t")
			case termbox.KeySpace:
				what = append(what, ' ')
				nfs.Term(m, "print", " ")
			default:
				what = append(what, ev.Ch)
				nfs.Term(m, "print", string(ev.Ch))
			}
		}
	}
}
func (nfs *NFS) printf(arg ...interface{}) *NFS {
	m := nfs.Context.Message()

	line := kit.Format(arg...)
	m.Log("term", "noutput %s", m.Cap("noutput", m.Capi("noutput")+1))
	m.Confv("output", -2, map[string]interface{}{"time": time.Now().Unix(), "line": kit.Format(arg...)})

	if m.Caps("termbox") {
		nfs.Term(m, "clear").Term(m, "print", line)
		m.Conf("term", "prompt_y", m.Conf("term", "cursor_y"))
		m.Conf("term", "end_row", m.Confi("term", "end_row")+1)
	} else {
		nfs.out.WriteString(line)
	}
	return nfs
}
func (nfs *NFS) prompt(arg ...interface{}) *NFS {
	m := nfs.Context.Message()

	target, _ := m.Optionv("ps_target").(*ctx.Context)
	line := fmt.Sprintf("%d[%s]%s> ", m.Capi("ninput"), time.Now().Format("15:04:05"), target.Name)
	m.Conf("prompt", line)

	line += kit.Format(arg...)
	if m.Caps("termbox") {
		m.Conf("term", "prompt_y", m.Conf("term", "cursor_y"))
		nfs.Term(m, "clear").Term(m, "print", line)
	} else {
		nfs.out.WriteString(line)
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
		nfs.Caches["termbox"] = &ctx.Cache{Value: "0"}
		nfs.Configs["input"] = &ctx.Config{Value: []interface{}{}}
		nfs.Configs["output"] = &ctx.Config{Value: []interface{}{}}
		nfs.Configs["prompt"] = &ctx.Config{Value: ""}

		if nfs.in = m.Optionv("in").(*os.File); m.Has("out") {
			if nfs.out = m.Optionv("out").(*os.File); m.Cap("goos") != "windows" {
				nfs.Term(m, "init")
				defer nfs.Term(m, "exit")
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
			m.Log("debug", "%s %d %d [%s]", "input", m.Capi("ninput", 1), len(line), line)
			m.Confv("input", -2, map[string]interface{}{"time": time.Now().Unix(), "line": line})

			for i := m.Capi("ninput") - 1; i < m.Capi("ninput"); i++ {
				line = m.Conf("input", []interface{}{i, "line"})

				msg := m.Backs(m.Spawn(m.Source()).Set("detail", line).Set("option", "file_pos", i))

				nfs.printf(m.Conf("prompt"), line, "\n")
				nfs.printf(msg.Meta["result"])

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

var Index = &ctx.Context{Name: "nfs", Help: "存储中心",
	Caches: map[string]*ctx.Cache{
		"nfile": &ctx.Cache{Name: "nfile", Value: "0", Help: "已经打开的文件数量"},
	},
	Configs: map[string]*ctx.Config{
		"term": &ctx.Config{Name: "term", Value: map[string]interface{}{
			"width": 80, "height": "24",

			"left": 0, "top": 0, "right": 80, "bottom": 24,
			"cursor_x": 0, "cursor_y": 0, "fgcolor": 0, "bgcolor": 0,
			"prompt": "", "wrap": "false",
			"begin_row": 0, "begin_col": 0, "end_row": 0, "end_col": 0,
		}, Help: "二维码的默认大小"},

		"buf_size":   &ctx.Config{Name: "buf_size", Value: "1024", Help: "读取文件的缓存区的大小"},
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
		"action": &ctx.Command{Name: "action cmd", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) {
				msg := m.Cmd("cli.source", arg)
				nfs.printf(msg.Conf("prompt"), arg, "\n")
				nfs.printf(msg.Meta["result"])
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
