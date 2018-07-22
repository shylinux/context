package nfs // {{{
// }}}
import ( // {{{
	"contexts"
	"encoding/json"
	"errors"
	"github.com/nsf/termbox-go"
	"github.com/skip2/go-qrcode"

	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"unicode"
)

// }}}

var FileNotExist = errors.New("file not exist")

type NFS struct {
	in            *os.File
	out           *os.File
	history       []string
	pages         []string
	width, height int

	paths []string

	io io.ReadWriteCloser
	*bufio.Reader
	*bufio.Writer
	send   map[int]*ctx.Message
	target *ctx.Context
	cli    *ctx.Message

	*ctx.Message
	*ctx.Context
}

func (nfs *NFS) open(name string) (*os.File, error) { // {{{
	if path.IsAbs(name) {
		nfs.Log("info", nil, "open %s", name)
		return os.Open(name)
	}
	for i := len(nfs.paths) - 1; i >= 0; i-- {
		if f, e := os.Open(path.Join(nfs.paths[i], name)); e == nil {
			nfs.Log("info", nil, "open %s", path.Join(nfs.paths[i], name))
			return f, e
		}
	}
	nfs.Log("info", nil, "open %s", name)
	return os.Open(name)
}

// }}}
func dir(m *ctx.Message, name string, level int) { // {{{
	back, e := os.Getwd()
	m.Assert(e)
	os.Chdir(name)
	defer os.Chdir(back)

	if fs, e := ioutil.ReadDir("."); m.Assert(e) {
		for _, f := range fs {
			if f.Name()[0] == '.' {
				continue
			}

			if f.IsDir() {
				if m.Has("dirs") {
					m.Optioni("dirs", m.Optioni("dirs")+1)
				}
			} else {
				if m.Has("files") {
					m.Optioni("files", m.Optioni("files")+1)
				}
			}

			if m.Has("sizes") {
				m.Optioni("sizes", m.Optioni("sizes")+int(f.Size()))
			}

			line := 0
			if m.Has("lines") {
				if !f.IsDir() {
					f, e := os.Open(path.Join(back, name, f.Name()))
					m.Assert(e)
					defer f.Close()
					bio := bufio.NewScanner(f)
					for bio.Scan() {
						bio.Text()
						line++
					}
					m.Optioni("lines", m.Optioni("lines")+line)
				}
			}

			filename := ""
			switch m.Confx("dir_name") {
			case "name":
				filename = f.Name()
			case "tree":
				filename = strings.Repeat("  ", level) + f.Name()
			default:
				filename = path.Join(back, name, f.Name())
			}

			if !(m.Confx("dir_type") == "file" && f.IsDir() ||
				m.Confx("dir_type") == "dir" && !f.IsDir()) {
				m.Add("append", "filename", filename)
				m.Add("append", "dir", f.IsDir())
				m.Add("append", "size", f.Size())
				m.Add("append", "line", line)
				m.Add("append", "time", f.ModTime().Format("2006-01-02 15:04:05"))
			}

			if f.IsDir() {
				dir(m, f.Name(), level+1)
			}
		}
	}
}

// }}}
func (nfs *NFS) insert(rest []rune, letters []rune) []rune { // {{{
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

// }}}

func (nfs *NFS) escape(form string, args ...interface{}) *NFS { // {{{
	if !nfs.Caps("windows") {
		fmt.Fprintf(nfs.out, "\033[%s", fmt.Sprintf(form, args...))
	}
	return nfs
}

// }}}
func (nfs *NFS) color(str string, attr ...int) *NFS { // {{{
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

// }}}
func (nfs *NFS) print(str string) bool { // {{{
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

// }}}
func (nfs *NFS) prompt(arg ...string) string { // {{{
	line, rest := "", ""
	if len(arg) > 0 {
		line = arg[0]
	}
	if len(arg) > 1 {
		rest = arg[1]
	}

	if !nfs.Caps("windows") && len(nfs.pages) > 0 {
		for i := (len(nfs.pages[len(nfs.pages)-1]) - 1) / (nfs.width); i > 0; i-- {
			nfs.escape("2K").escape("A")
		}
		nfs.escape("2K").escape("G").escape("?25h")
	}

	ps := nfs.Option("prompt")
	if len(nfs.pages) > 0 {
		nfs.pages = nfs.pages[:len(nfs.pages)-1]
	}
	nfs.pages = append(nfs.pages, ps+line+rest+"\n")

	if nfs.color(ps, nfs.Confi("pscolor")).color(line).color(rest); len(rest) > 0 {
		nfs.escape("%dD", len(rest))
	}
	return ps
}

// }}}

func (nfs *NFS) zone(buf []string, top, height int) (row, col int) { // {{{
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

// }}}
func (nfs *NFS) page(buf []string, row, col, top, height int, status bool) { // {{{
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

// }}}
func (nfs *NFS) View(buf []string, top int, height int) { // {{{

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

// }}}
func (nfs *NFS) Read(p []byte) (n int, err error) { // {{{
	if nfs.Caps("windows") || !nfs.Caps("termbox") {
		return nfs.in.Read(p)
	}

	buf := make([]rune, 0, 1024)
	rest := make([]rune, 0, 1024)

	back := buf

	his := len(nfs.history)

	tab := []string{}
	tabi := 0

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				termbox.Close()
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
				for i := 0; i < len(nfs.history); i++ {
					his = (his + len(nfs.history) - 1) % len(nfs.history)
					if strings.HasPrefix(nfs.history[his], string(buf)) {
						rest = rest[:0]
						rest = append(rest, []rune(nfs.history[his][len(buf):])...)
						break
					}
				}

			case termbox.KeyCtrlN:
				for i := 0; i < len(nfs.history); i++ {
					his = (his + len(nfs.history) + 1) % len(nfs.history)
					if strings.HasPrefix(nfs.history[his], string(buf)) {
						rest = rest[:0]
						rest = append(rest, []rune(nfs.history[his][len(buf):])...)
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
					back = back[:0]
					back = append(back, rest...)
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
					msg := nfs.Message.Spawn(nfs.cli.Target())
					target := msg.Cmd("target").Data["target"].(*ctx.Context)
					msg.Spawn(target).BackTrace(func(msg *ctx.Message) bool {
						for k, _ := range msg.Target().Commands {
							if strings.HasPrefix(k, prefix) {
								tab = append(tab, k[len(prefix):])
							}
						}
						return true
					})
				}

				if tabi >= 0 && tabi < len(tab) {
					rest = rest[:0]
					rest = append(rest, []rune(tab[tabi])...)
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

// }}}

func (nfs *NFS) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	nfs.Message = m
	if len(arg) > 0 && (arg[0] == "scan" || arg[0] == "open") {
		c.Caches = map[string]*ctx.Cache{
			"pos":    &ctx.Cache{Name: "pos", Value: "0", Help: "pos"},
			"size":   &ctx.Cache{Name: "size", Value: "0", Help: "size"},
			"nread":  &ctx.Cache{Name: "nread", Value: "0", Help: "nread"},
			"nwrite": &ctx.Cache{Name: "nwrite", Value: "0", Help: "nwrite"},
			"nline":  &ctx.Cache{Name: "缓存命令行数", Value: "0", Help: "缓存命令行数"},
		}
		c.Configs = map[string]*ctx.Config{}
	} else {
		c.Caches = map[string]*ctx.Cache{
			"nbytes": &ctx.Cache{Name: "消息发送字节", Value: "0", Help: "消息发送字节"},
			"nsend":  &ctx.Cache{Name: "消息发送数量", Value: "0", Help: "消息发送数量"},
			"nrecv":  &ctx.Cache{Name: "消息接收数量", Value: "0", Help: "消息接收数量"},
			"target": &ctx.Cache{Name: "消息接收模块", Value: "ssh", Help: "消息接收模块"},
			"result": &ctx.Cache{Name: "前一条指令执行结果", Value: "", Help: "前一条指令执行结果"},
			"sessid": &ctx.Cache{Name: "会话令牌", Value: "", Help: "会话令牌"},
		}
		c.Configs = map[string]*ctx.Config{}
	}

	s := new(NFS)
	s.Context = c
	return s

}

// }}}
func (nfs *NFS) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	nfs.Message = m
	nfs.width, nfs.height = 1, 1
	return nfs
}

// }}}
func (nfs *NFS) Start(m *ctx.Message, arg ...string) bool { // {{{
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
			termbox.Init()
			defer termbox.Close()
			nfs.width, nfs.height = termbox.Size()
			nfs.Cap("termbox", "true")
			nfs.Conf("color", "true")
			nfs.out = m.Optionv("out").(*os.File)
		}

		line := ""
		for nfs.prompt(); !m.Options("scan_end") && bio.Scan(); nfs.prompt() {
			text := bio.Text()
			m.Capi("nread", len(text)+1)

			if line += text; len(text) > 0 && text[len(text)-1] == '\\' {
				line = line[:len(line)-1]
				continue
			}
			nfs.history = append(nfs.history, line)
			m.Capi("nline", 1)

			for i := len(nfs.history) - 1; i < len(nfs.history); i++ {
				line = nfs.history[i]

				msg := m.Spawn(m.Source()).Set("detail", line)
				msg.Option("file_pos", i)
				m.Back(msg)

				for _, v := range msg.Meta["result"] {
					m.Capi("nwrite", len(v))
					nfs.print(v)
				}
				if msg.Append("file_pos0") != "" {
					i = msg.Appendi("file_pos0")
					i--
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

	if len(arg) > 0 && arg[0] == "open" {
		nfs.out = m.Optionv("out").(*os.File)
		nfs.in = m.Optionv("in").(*os.File)
		s, e := nfs.in.Stat()
		m.Assert(e)
		m.Capi("size", int(s.Size()))
		m.Cap("stream", arg[1])
		return false
	}

	m.Target().Sessions["nfs"] = m
	m.Sessions["nfs"] = m

	nfs.Message = m
	if socket, ok := m.Data["io"]; ok {
		m.Cap("stream", m.Source().Name)
		// m.Sesss("aaa", "aaa").Cmd("login", "demo", "demo")
		m.Options("stdio", false)

		nfs.io = socket.(io.ReadWriteCloser)
		nfs.Reader = bufio.NewReader(nfs.io)
		nfs.Writer = bufio.NewWriter(nfs.io)

		nfs.send = make(map[int]*ctx.Message)
		nfs.target = m.Target()
		if target, ok := m.Data["target"]; ok {
			nfs.target = target.(*ctx.Context)
		}

		var msg *ctx.Message

		nfs.Caches["target"] = &ctx.Cache{Name: "target", Value: "", Help: "文件名"}

		nsend := ""

		for {
			line, e := nfs.Reader.ReadString('\n')
			if msg == nil {
				msg = m.Sesss("ssh")
				m.Cap("target", msg.Target().Name)
			}

			if e == io.EOF {
				msg.Cmd("close")
			}
			m.Assert(e)

			if line = strings.TrimSpace(line); len(line) > 0 {
				ls := strings.SplitN(line, ":", 2)
				ls[0] = strings.TrimSpace(ls[0])
				ls[1], e = url.QueryUnescape(strings.TrimSpace(ls[1]))
				m.Assert(e)

				switch ls[0] {
				case "detail":
					msg.Add("detail", ls[1])
				case "result":
					msg.Add("result", ls[1])
				case "nsend":
					nsend = ls[1]
				default:
					msg.Add("option", ls[0], ls[1])
				}
				continue
			}

			if msg.Has("detail") {
				msg.Log("info", nil, "%d recv", m.Capi("nrecv", 1))
				msg.Log("info", nil, "detail: %v", msg.Meta["detail"])
				msg.Log("info", nil, "option: %v", msg.Meta["option"])
				msg.Options("stdio", false)
				msg.Option("nsend", nsend)

				func() {
					cmd := msg
					nsends := nsend
					cmd.Call(func(sub *ctx.Message) *ctx.Message {
						for _, v := range sub.Meta["result"] {
							_, e := fmt.Fprintf(nfs.Writer, "result: %s\n", url.QueryEscape(v))
							sub.Assert(e)
						}

						sub.Append("nsend", nsends)
						for _, k := range sub.Meta["append"] {
							for _, v := range sub.Meta[k] {
								_, e := fmt.Fprintf(nfs.Writer, "%s: %s\n", k, url.QueryEscape(v))
								sub.Assert(e)
							}
						}

						sub.Log("info", nil, "%d recv", sub.Optioni("nsend"))
						sub.Log("info", nil, "result: %v", sub.Meta["result"])
						sub.Log("info", nil, "append: %v", sub.Meta["append"])

						_, e := fmt.Fprintf(nfs.Writer, "\n")
						sub.Assert(e)
						e = nfs.Writer.Flush()
						sub.Assert(e)

						if sub.Has("io") {
							if f, ok := sub.Data["io"].(io.ReadCloser); ok {
								io.Copy(nfs.Writer, f)
								nfs.Writer.Flush()
								f.Close()
							}
						}
						return sub
					})
				}()

			} else {
				msg.Meta["append"] = msg.Meta["option"]
				delete(msg.Meta, "option")

				msg.Log("info", nil, "%s send", nsend)
				msg.Log("info", nil, "result: %v", msg.Meta["result"])
				msg.Log("info", nil, "append: %v", msg.Meta["append"])

				n, e := strconv.Atoi(nsend)
				m.Assert(e)
				send := nfs.send[n]
				send.Copy(msg, "result")
				send.Copy(msg, "append")

				if send.Has("io") {
					if f, ok := send.Data["io"].(io.WriteCloser); ok {
						io.CopyN(f, nfs.Reader, int64(send.Appendi("size")))
						f.Close()
					}
				}

				send.Back(send)
			}

			nsend = ""
			msg = nil
		}
		return true
	}

	return false
}

// }}}
func (nfs *NFS) Close(m *ctx.Message, arg ...string) bool { // {{{
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
		if nfs.io != nil {
			nfs.io.Close()
			nfs.io = nil
		}
	case m.Source():
		m.Source().Close(m.Spawn(m.Source()))
	}
	if m.Target() == Index {
		return false
	}
	return true
}

// }}}

var Index = &ctx.Context{Name: "nfs", Help: "存储中心",
	Caches: map[string]*ctx.Cache{
		"nfile": &ctx.Cache{Name: "nfile", Value: "-1", Help: "已经打开的文件数量"},
	},
	Configs: map[string]*ctx.Config{
		"nfs_name": &ctx.Config{Name: "nfs_name", Value: "file", Help: "默认模块命名", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				return arg[0]
			}
			return fmt.Sprintf("%s%d", x.Value, m.Capi("nfile", 1))
			// }}}
		}},
		"nfs_help": &ctx.Config{Name: "nfs_help", Value: "file", Help: "默认模块帮助"},

		"buf_size": &ctx.Config{Name: "buf_size", Value: "1024", Help: "读取文件的缓存区的大小"},
		"qr_size":  &ctx.Config{Name: "qr_size", Value: "256", Help: "生成二维码的图片的大小"},

		"dir_name":   &ctx.Config{Name: "dir_name(name/tree/path/full)", Value: "name", Help: "dir命令输出文件名的类型, name: 文件名, tree: 带缩进的文件名, path: 相对路径, full: 绝对路径"},
		"dir_info":   &ctx.Config{Name: "dir_info(sizes/lines/files/dirs)", Value: "sizes lines files dirs", Help: "dir命令输出目录的统计信息, info: 输出统计信息, 否则输出"},
		"dir_type":   &ctx.Config{Name: "dir_type(file/dir)", Value: "file", Help: "dir命令输出的文件类型, file: 只输出普通文件, dir: 只输出目录文件, 否则输出所有文件"},
		"sort_field": &ctx.Config{Name: "sort_field", Value: "line", Help: "表格排序字段"},
		"sort_order": &ctx.Config{Name: "sort_order(int/int_r/string/string_r/time/time_r)", Value: "int", Help: "表格排序类型"},

		"git_branch":   &ctx.Config{Name: "git_branch", Value: "--list", Help: "版本控制状态参数"},
		"git_status":   &ctx.Config{Name: "git_status", Value: "-sb", Help: "版本控制状态参数"},
		"git_diff":     &ctx.Config{Name: "git_diff", Value: "--stat", Help: "版本控制状态参数"},
		"git_log":      &ctx.Config{Name: "git_log", Value: "--pretty=%h %an(%ad) %s  --date=format:%m/%d %H:%M  --graph", Help: "版本控制状态参数"},
		"git_log_form": &ctx.Config{Name: "git_log", Value: "stat", Help: "版本控制状态参数"},
		"git_log_skip": &ctx.Config{Name: "git_log", Value: "0", Help: "版本控制状态参数"},
		"git_log_line": &ctx.Config{Name: "git_log", Value: "5", Help: "版本控制状态参数"},
		"git_path":     &ctx.Config{Name: "git_path", Value: ".", Help: "版本控制默认路径"},
		"git_info":     &ctx.Config{Name: "git_info", Value: "branch status diff", Help: "命令集合"},
	},
	Commands: map[string]*ctx.Command{
		"paths": &ctx.Command{
			Name: "paths [add path]|[del index]|[set index path]|[index]",
			Help: "设置文件搜索路径, add: 添加目录, del: 删除目录, set: 修改目录，index: 目录序号, path: 目录名",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) { // {{{
					if len(arg) == 0 {
						for i, v := range nfs.paths {
							m.Echo("%d: %s\n", i, v)
						}
						return
					}
					switch arg[0] {
					case "add":
						nfs.paths = append(nfs.paths, arg[1])
					case "del":
						if i, e := strconv.Atoi(arg[1]); e == nil && i < len(nfs.paths) {
							for ; i < len(nfs.paths)-1; i++ {
								nfs.paths[i] = nfs.paths[i+1]
							}
							nfs.paths = nfs.paths[:len(nfs.paths)-1]
						}
					case "set":
						if i, e := strconv.Atoi(arg[1]); e == nil && i < len(nfs.paths) {
							nfs.paths[i] = arg[2]
							m.Echo("%d: %s\n", i, nfs.paths[i])
						}
					default:
						if i, e := strconv.Atoi(arg[0]); e == nil && i < len(nfs.paths) {
							m.Echo("%d: %s\n", i, nfs.paths[i])
						}
					}
				} // }}}
			}},
		"scan": &ctx.Command{
			Name: "scan filename [nfs_name [nfs_help]]",
			Help: "扫描文件, filename: 文件名, nfs_name: 模块名, nfs_help: 模块帮助",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) { // {{{
					if arg[0] == "stdio" {
						m.Optionv("in", os.Stdin)
						m.Optionv("out", os.Stdout)
					} else {
						if f, e := nfs.open(arg[0]); m.Assert(e) {
							m.Optionv("in", f)
						}
					}

					m.Start(m.Confx("nfs_name", arg, 1), m.Confx("nfs_help", arg, 2), key, arg[0])
				} // }}}
			}},
		"history": &ctx.Command{
			Name: "history [save|load filename [lines [pos]]] [find|search key]",
			Help: "扫描记录, save: 保存记录, load: 加载记录, filename: 文件名, lines: 加载或保存记录数量, pos: 加载或保存的起始位置, find: 查找记录, search: 搜索记录, key: 查找或搜索的参数",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) { // {{{
					if len(arg) == 0 {
						for i, v := range nfs.history {
							m.Echo("%d: %s\n", i, v)
						}
						return
					}
					switch arg[0] {
					case "load":
						f, e := nfs.open(arg[1])
						m.Assert(e)
						defer f.Close()

						pos, lines := 0, -1
						if len(arg) > 3 {
							i, e := strconv.Atoi(arg[3])
							m.Assert(e)
							pos = i

						}
						if len(arg) > 2 {
							i, e := strconv.Atoi(arg[2])
							m.Assert(e)
							lines = i
						}

						bio := bufio.NewScanner(f)
						for i := 0; bio.Scan(); i++ {
							if i < pos {
								continue
							}
							if lines != -1 && (i-pos) >= lines {
								break
							}
							nfs.history = append(nfs.history, bio.Text())
						}
						m.Capi("nline", 0, len(nfs.history))
					case "save":
						f, e := os.Create(arg[1])
						m.Assert(e)
						defer f.Close()

						pos, lines := 0, -1
						if len(arg) > 3 {
							i, e := strconv.Atoi(arg[3])
							m.Assert(e)
							pos = i

						}
						if len(arg) > 2 {
							i, e := strconv.Atoi(arg[2])
							m.Assert(e)
							lines = i
						}

						for i, v := range nfs.history {
							if i < pos {
								continue
							}
							if lines != -1 && (i-pos) >= lines {
								break
							}
							fmt.Fprintln(f, v)
						}
					case "find":
						for i, v := range nfs.history {
							if strings.HasPrefix(v, arg[1]) {
								m.Echo("%d: %s\n", i, v)
							}
						}
					case "search":
					default:
						if i, e := strconv.Atoi(arg[0]); e == nil && i < len(nfs.history) {
							m.Echo(nfs.history[i])
						}
					}
				} // }}}
			}},
		"open": &ctx.Command{
			Name: "open filename [nfs_name [nfs_help]]",
			Help: "打开文件, filename: 文件名, nfs_name: 模块名, nfs_help: 模块帮助",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if m.Has("io") { // {{{
				} else if f, e := os.OpenFile(arg[0], os.O_RDWR|os.O_CREATE, os.ModePerm); m.Assert(e) {
					m.Put("option", "in", f).Put("option", "out", f)
				}
				m.Start(m.Confx("nfs_name", arg, 1), m.Confx("nfs_help", arg, 2), "open", arg[0])
				m.Echo(m.Target().Name)
				// }}}
			}},
		"read": &ctx.Command{Name: "read [buf_size [pos]]", Help: "读取文件, buf_size: 读取大小, pos: 读取位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.in != nil { // {{{
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
			} // }}}
		}},
		"write": &ctx.Command{Name: "write string [pos]", Help: "写入文件, string: 写入内容, pos: 写入位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) && nfs.out != nil { // {{{
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
			} // }}}
		}},
		"load": &ctx.Command{Name: "load file [buf_size [pos]]", Help: "加载文件, buf_size: 加载大小, pos: 加载位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if f, e := os.Open(arg[0]); m.Assert(e) { // {{{
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
					m.Log("info", nil, "read %d", l)
					m.Echo(string(buf[:l]))
				}
			} // }}}
		}},
		"save": &ctx.Command{Name: "save file string...", Help: "保存文件, file: 保存的文件, string: 保存的内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if f, e := os.Create(arg[0]); m.Assert(e) { // {{{
				defer f.Close()

				for _, v := range arg[1:] {
					fmt.Fprint(f, v)
				}
			} // }}}
		}},
		"print": &ctx.Command{Name: "print file string...", Help: "输出文件, file: 输出的文件, string: 输出的内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if f, e := os.OpenFile(arg[0], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); m.Assert(e) { // {{{
				defer f.Close()

				for _, v := range arg[1:] {
					fmt.Fprint(f, v)
				}
				fmt.Fprint(f, "\n")
			} // }}}
		}},
		"genqr": &ctx.Command{Name: "genqr [qr_size size] filename string...", Help: "生成二维码图片, qr_size: 图片大小, filename: 文件名, string: 输出内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if size, e := strconv.Atoi(m.Confx("qr_size")); m.Assert(e) { // {{{
				qrcode.WriteFile(strings.Join(arg[1:], ""), qrcode.Medium, size, arg[0])
			} // }}}
		}},
		"json": &ctx.Command{Name: "json [key value]...", Help: "生成格式化内容, key: 参数名, value: 参数值", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			data := map[string]interface{}{} // {{{
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
			// }}}
		}},
		"pwd": &ctx.Command{Name: "pwd", Help: "查看当前路径", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			wd, e := os.Getwd() // {{{
			m.Assert(e)
			m.Echo(wd) // }}}
		}},
		"dir": &ctx.Command{
			Name: "dir dir [dir_info info] [dir_name name|tree|path|full] [dir_type file|dir] [sort_field name] [sort_order type]",
			Help: "查看目录, dir: 目录名, dir_info: 显示统计信息, dir_name: 文件名类型, dir_type: 文件类型, sort_field: 排序字段, sort_order: 排序类型",
			Form: map[string]int{"dir_info": 1, "dir_name": 1, "dir_type": 1, "sort_field": 1, "sort_order": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				d := "." // {{{
				if len(arg) > 0 {
					d = arg[0]
				}
				trip := 0
				if m.Confx("dir_name") == "path" {
					wd, e := os.Getwd()
					m.Assert(e)
					trip = len(wd) + 1
				}

				info := strings.Split(m.Confx("dir_info"), " ")
				for _, v := range info {
					m.Option(v, 0)
				}

				dir(m, d, 0)
				m.Sort(m.Confx("sort_field"), m.Confx("sort_order"))
				m.Table(func(maps map[string]string, list []string, line int) bool {
					for i, v := range list {
						key := m.Meta["append"][i]
						switch key {
						case "filename":
							if trip > 0 {
								v = v[trip:]
							}
						case "dir":
							continue
						}
						m.Echo("%s\t", v)
					}
					m.Echo("\n")
					return true
				})
				for _, v := range info {
					m.Echo("%s: %s\n", v, m.Option(v))
				}
				// }}}
			}},
		"git": &ctx.Command{
			Name: "git branch|status|diff|log|info arg... [git_path path]...",
			Help: "版本控制, branch: 分支管理, status: 查看状态, info: 查看分支与状态, git_path: 指定路径",
			Form: map[string]int{"git_path": 1, "git_info": 1, "git_log": 1, "git_log_form": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if len(arg) == 0 { // {{{
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
				if !m.Has("git_path") {
					m.Option("git_path", m.Conf("git_path"))
				}
				for _, p := range m.Meta["git_path"] {
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
							m.Log("info", nil, "cmd: %s %v", "git", "difftool", "-y")
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

						m.Log("info", nil, "cmd: %s %v", "git", ctx.Trans("-C", p, c, args))
						cmd := exec.Command("git", ctx.Trans("-C", p, c, args)...)
						if out, e := cmd.CombinedOutput(); e != nil {
							m.Echo("error: ")
							m.Echo("%s\n", e)
						} else {
							if m.Echo(string(out)); len(out) > 0 {
								m.Echo("\n")
							}
						}
					}
				} // }}}
			}},

		"listen": &ctx.Command{Name: "listen args...", Help: "启动文件服务, args: 参考tcp模块, listen命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*NFS); m.Assert(ok) { //{{{
				m.Find("tcp").Call(func(com *ctx.Message) *ctx.Message {
					sub := com.Spawn(m.Target())
					sub.Put("option", "target", m.Source())
					sub.Put("option", "io", com.Data["io"])
					sub.Start(fmt.Sprintf("file%d", m.Capi("nfile", 1)), "打开文件")
					return sub
				}, m.Meta["detail"])
			}
			// }}}
		}},
		"dial": &ctx.Command{Name: "dial args...", Help: "连接文件服务, args: 参考tcp模块, dial命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*NFS); m.Assert(ok) { //{{{
				m.Find("tcp").Call(func(com *ctx.Message) *ctx.Message {
					sub := com.Spawn(m.Target())
					sub.Put("option", "target", m.Source())
					sub.Put("option", "io", com.Data["io"])
					sub.Start(fmt.Sprintf("file%d", m.Capi("nfile", 1)), "打开文件")
					return sub
				}, m.Meta["detail"])
			}
			// }}}
		}},
		"send": &ctx.Command{Name: "send [file] args...", Help: "连接文件服务, args: 参考tcp模块, dial命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) { // {{{
				if m.Has("nrecv") {
					if len(arg) > 1 && arg[0] == "file" {
						info, e := os.Stat(arg[1])
						m.Assert(e)
						m.Append("name", info.Name())
						m.Append("size", info.Size())
						m.Append("time", info.ModTime())
						m.Append("mode", info.Mode())

						f, e := os.Open(arg[1])
						m.Assert(e)
						m.Put("append", "io", f)
					}

				} else {
					nfs.send[m.Optioni("nsend", m.Capi("nsend", 1))] = m

					if len(arg) > 1 && arg[0] == "file" {
						info, e := os.Stat(arg[1])
						m.Assert(e)
						m.Option("name", info.Name())
						m.Option("size", info.Size())
						m.Option("time", info.ModTime())
						m.Option("mode", info.Mode())

						n, e := fmt.Fprintf(nfs.Writer, "detail: recv\n")
						m.Capi("nbytes", n)
						m.Assert(e)
					}
					for _, v := range arg {
						n, e := fmt.Fprintf(nfs.Writer, "detail: %v\n", v)
						m.Capi("nbytes", n)
						m.Assert(e)
					}

					for _, k := range m.Meta["option"] {
						if k == "args" {
							continue
						}
						for _, v := range m.Meta[k] {
							n, e := fmt.Fprintf(nfs.Writer, "%s: %s\n", k, v)
							m.Capi("nbytes", n)
							m.Assert(e)
						}
					}
					m.Log("info", nil, "%d send", m.Optioni("nsend"))
					m.Log("info", nil, "detail: %v", m.Meta["detail"])
					m.Log("info", nil, "option: %v", m.Meta["option"])

					n, e := fmt.Fprintf(nfs.Writer, "\n")
					m.Capi("nbytes", n)
					m.Assert(e)
					nfs.Writer.Flush()

					if len(arg) > 1 && arg[0] == "file" {
						f, e := os.Open(arg[1])
						m.Assert(e)
						defer f.Close()
						_, e = io.Copy(nfs.Writer, f)
					}
				}
			} // }}}
		}},
		"recv": &ctx.Command{Name: "recv [file] args...", Help: "连接文件服务, args: 参考tcp模块, dial命令的参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if nfs, ok := m.Target().Server.(*NFS); m.Assert(ok) { // {{{
				if m.Has("nrecv") {
					if len(arg) > 1 && arg[0] == "file" {
						f, e := os.Create(arg[1])
						m.Assert(e)
						defer f.Close()
						io.CopyN(f, nfs.Reader, int64(m.Optioni("size")))
					}

					return
				}

				nfs.send[m.Optioni("nrecv", m.Capi("nsend", 1))] = m

				if len(arg) > 1 && arg[0] == "file" {
					f, e := os.Create(arg[1])
					m.Assert(e)
					m.Put("option", "io", f)

					fmt.Fprintf(nfs.Writer, "detail: send\n")
				}

				for _, v := range arg {
					fmt.Fprintf(nfs.Writer, "detail: %v\n", v)
				}

				for _, k := range m.Meta["option"] {
					if k == "args" {
						continue
					}
					for _, v := range m.Meta[k] {
						fmt.Fprintf(nfs.Writer, "%s: %s\n", k, v)
					}
				}

				fmt.Fprintf(nfs.Writer, "\n")
				nfs.Writer.Flush()
			} // }}}
		}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void",
			Commands: map[string]*ctx.Command{
				"scan":  &ctx.Command{},
				"open":  &ctx.Command{},
				"save":  &ctx.Command{},
				"load":  &ctx.Command{},
				"genqr": &ctx.Command{},
				"write": &ctx.Command{},
			},
		},
	},
}

func init() {
	nfs := &NFS{}
	nfs.Context = Index
	ctx.Index.Register(Index, nfs)
}
