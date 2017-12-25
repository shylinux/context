package nfs

import (
	"context"

	"fmt"
	"github.com/skip2/go-qrcode"
	"io"
	"os"
	"strconv"
	"strings"
)

type NFS struct {
	file *os.File
	*ctx.Context
}

func (nfs *NFS) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	file, e := os.OpenFile(arg[0], os.O_RDWR|os.O_CREATE, os.ModePerm)
	m.Assert(e)
	info, e := os.Stat(arg[0])
	m.Assert(e)

	c.Caches = map[string]*ctx.Cache{
		"name": &ctx.Cache{Name: "name", Value: info.Name(), Help: "文件名"},
		"mode": &ctx.Cache{Name: "mode", Value: info.Mode().String(), Help: "文件权限"},
		"time": &ctx.Cache{Name: "time", Value: info.ModTime().Format("15:03:04"), Help: "创建时间"},
		"size": &ctx.Cache{Name: "size", Value: fmt.Sprintf("%d", info.Size()), Help: "文件大小"},
		"pos":  &ctx.Cache{Name: "pos", Value: "0", Help: "读写位置"},
	}
	c.Configs = map[string]*ctx.Config{}

	s := new(NFS)
	s.Context = c
	s.file = file
	return s

}

func (nfs *NFS) Begin(m *ctx.Message, arg ...string) ctx.Server {
	if nfs.Context == Index {
		Pulse = m
	}
	return nfs
}

func (nfs *NFS) Start(m *ctx.Message, arg ...string) bool {
	m.Log("info", nil, "%d open %s", Pulse.Capi("nfile"), m.Cap("name"))
	m.Cap("stream", m.Cap("name"))
	return false
}

func (nfs *NFS) Close(m *ctx.Message, arg ...string) bool {
	switch nfs.Context {
	case m.Target():
		if nfs.file != nil {
			m.Log("info", nil, "%d close %s", Pulse.Capi("nfile", -1)+1, m.Cap("name"))
			nfs.file.Close()
			nfs.file = nil
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
		"open": &ctx.Command{Name: "open file", Help: "打开文件, file: 文件名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Start(fmt.Sprintf("file%d", Pulse.Capi("nfile", 1)), "打开文件", arg...)
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
			if n, e = nfs.file.ReadAt(buf, int64(m.Capi("pos"))); e != io.EOF {
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
				m.Assert(nfs.file.Truncate(int64(m.Capi("pos"))))
				m.Cap("size", m.Cap("pos"))
				m.Cap("pos", "0")
			} else {
				n, e := nfs.file.WriteAt([]byte(arg[0]), int64(m.Capi("pos")))
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

			size := 1024
			if len(arg) > 1 {
				if s, e := strconv.Atoi(arg[1]); e == nil {
					size = s
				}
			}

			buf := make([]byte, size)
			f.Read(buf)
			m.Echo(string(buf))
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
			m.Log("fuck", nil, "%v %v", arg[0], arg[1:])
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
