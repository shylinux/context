package nfs

import (
	"context"

	"bufio"
	"fmt"
	"github.com/skip2/go-qrcode"
	"io"
	"os"
	"strconv"
	"strings"
)

type NFS struct {
	in  *os.File
	out *os.File
	*ctx.Context
}

func (nfs *NFS) print(str string, arg ...interface{}) bool {
	if nfs.out == nil {
		return false
	}

	fmt.Fprintf(nfs.out, str, arg...)
	return true
}

func (nfs *NFS) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{
		"pos": &ctx.Cache{Name: "读写位置", Value: "0", Help: "读写位置"},
	}
	c.Configs = map[string]*ctx.Config{}

	if info, e := os.Stat(arg[1]); e == nil {
		c.Caches["name"] = &ctx.Cache{Name: "name", Value: info.Name(), Help: "文件名"}
		c.Caches["mode"] = &ctx.Cache{Name: "mode", Value: info.Mode().String(), Help: "文件权限"}
		c.Caches["size"] = &ctx.Cache{Name: "size", Value: fmt.Sprintf("%d", info.Size()), Help: "文件大小"}
		c.Caches["time"] = &ctx.Cache{Name: "time", Value: info.ModTime().Format("15:03:04"), Help: "创建时间"}
	}

	s := new(NFS)
	s.Context = c
	return s

}

func (nfs *NFS) Begin(m *ctx.Message, arg ...string) ctx.Server {
	if nfs.Context == Index {
		Pulse = m
	}
	return nfs
}

func (nfs *NFS) Start(m *ctx.Message, arg ...string) bool {
	if out, ok := m.Data["out"]; ok {
		nfs.out = out.(*os.File)
	}
	if in, ok := m.Data["in"]; ok {
		nfs.in = in.(*os.File)
	}

	m.Log("info", nil, "%d %v", Pulse.Capi("nfile"), arg)
	if m.Cap("stream", arg[1]); arg[0] == "open" {
		return false
	}

	cli := m.Reply()
	cli.Conf("yac", "yac")
	yac := m.Find(cli.Conf("yac"))
	bio := bufio.NewScanner(nfs.in)
	nfs.print("%s", cli.Conf("PS1"))

	for bio.Scan() {
		line := m
		if yac != nil {
			// line = cli.Spawn(yac.Target())
			line = m.Spawn(yac.Target())
		} else {
			line = m.Reply()
		}
		line.Cmd(append([]string{"parse", "line", "void"}, strings.Split(bio.Text()+" \n", " ")...)...)

		if result := strings.TrimRight(strings.Join(line.Meta["result"], ""), "\n"); len(result) > 0 {
			nfs.print("%s", result+"\n")
		}
		nfs.print("%s", cli.Conf("PS1"))
	}
	return true
}

func (nfs *NFS) Close(m *ctx.Message, arg ...string) bool {
	switch nfs.Context {
	case m.Target():
		if nfs.in != nil {
			m.Log("info", nil, "%d close %s", Pulse.Capi("nfile", -1)+1, m.Cap("name"))
			nfs.in.Close()
			nfs.in = nil
		}
	case m.Source():
	}
	return true
}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "nfs", Help: "存储中心",
	Caches: map[string]*ctx.Cache{
		"nfile": &ctx.Cache{Name: "nfile", Value: "0", Help: "已经打开的文件数量"},
	},
	Configs: map[string]*ctx.Config{
		"size": &ctx.Config{Name: "size", Value: "1024", Help: "读取文件的默认大小值"},
	},
	Commands: map[string]*ctx.Command{
		"scan": &ctx.Command{Name: "scan file", Help: "扫描文件, file: 文件名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if arg[0] == "stdio" {
				m.Put("option", "in", os.Stdin).Put("option", "out", os.Stdout)
				m.Start("stdio", "扫描文件", m.Meta["detail"]...)
			} else if f, e := os.Open(arg[0]); m.Assert(e) {
				m.Put("option", "in", f)
				m.Start(fmt.Sprintf("file%d", Pulse.Capi("nfile", 1)), "扫描文件", m.Meta["detail"]...)
			}
			m.Echo(m.Target().Name)
		}},
		"open": &ctx.Command{Name: "open file", Help: "打开文件, file: 文件名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if f, e := os.OpenFile(arg[0], os.O_RDWR|os.O_CREATE, os.ModePerm); m.Assert(e) {
				m.Put("option", "in", f).Put("option", "out", f)
				m.Start(fmt.Sprintf("file%d", Pulse.Capi("nfile", 1)), "打开文件", m.Meta["detail"]...)
			}
			m.Echo(m.Target().Name)
		}},
		"read": &ctx.Command{Name: "read [size [pos]]", Help: "读取文件, size: 读取大小, pos: 读取位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			nfs, ok := m.Target().Server.(*NFS)
			m.Assert(ok)

			var e error
			n := m.Confi("size")
			if len(arg) > 0 {
				n, e = strconv.Atoi(arg[0])
				m.Assert(e)
			}
			if len(arg) > 1 {
				m.Cap("pos", arg[1])
			}

			buf := make([]byte, n)
			if n, e = nfs.in.ReadAt(buf, int64(m.Capi("pos"))); e != io.EOF {
				m.Assert(e)
			}
			m.Echo(string(buf))

			if m.Capi("pos", n); m.Capi("pos") == m.Capi("size") {
				m.Cap("pos", "0")
			}
		}},
		"write": &ctx.Command{Name: "write string [pos]", Help: "写入文件, string: 写入内容, pos: 写入位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			nfs, ok := m.Target().Server.(*NFS)
			if m.Assert(ok); len(arg) > 1 {
				m.Cap("pos", arg[1])
			}

			if len(arg[0]) == 0 {
				m.Assert(nfs.in.Truncate(int64(m.Capi("pos"))))
				m.Cap("size", m.Cap("pos"))
				m.Cap("pos", "0")
			} else {
				n, e := nfs.in.WriteAt([]byte(arg[0]), int64(m.Capi("pos")))
				if m.Assert(e) && m.Capi("pos", n) > m.Capi("size") {
					m.Cap("size", m.Cap("pos"))
				}
			}

			m.Echo(m.Cap("pos"))
		}},
		"load": &ctx.Command{Name: "load file [size]", Help: "写入文件, string: 写入内容, pos: 写入位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			f, e := os.Open(arg[0])
			if e != nil {
				return
			}
			defer f.Close()

			m.Meta = nil
			size := 1024
			if len(arg) > 1 {
				if s, e := strconv.Atoi(arg[1]); e == nil {
					size = s
				}
			}
			buf := make([]byte, size)
			l, e := f.Read(buf)
			m.Echo(string(buf[:l]))
			m.Log("info", nil, "read %d", l)
		}},
		"save": &ctx.Command{Name: "save file string...", Help: "写入文件, string: 写入内容, pos: 写入位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			f, e := os.Create(arg[0])
			m.Assert(e)
			defer f.Close()
			fmt.Fprint(f, strings.Join(arg[1:], ""))
		}},
		"genqr": &ctx.Command{Name: "genqr [size] file string...", Help: "写入文件, string: 写入内容, pos: 写入位置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			size := 256
			if len(arg) > 2 {
				if s, e := strconv.Atoi(arg[0]); e == nil {
					arg = arg[1:]
					size = s
				}
			}
			qrcode.WriteFile(strings.Join(arg[1:], ""), qrcode.Medium, size, arg[0])
		}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void",
			Commands: map[string]*ctx.Command{
				"open":  &ctx.Command{},
				"save":  &ctx.Command{},
				"load":  &ctx.Command{},
				"genqr": &ctx.Command{},
			},
		},
	},
}

func init() {
	nfs := &NFS{}
	nfs.Context = Index
	ctx.Index.Register(Index, nfs)
}
