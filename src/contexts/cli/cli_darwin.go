package cli

import (
	"contexts/ctx"
	"fmt"
	"runtime"
	"syscall"
	"toolkit"
)

func sysinfo(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
	m.Append("NumCPU", runtime.NumCPU())

	fs := &syscall.Statfs_t{}
	syscall.Statfs("./", fs)
	m.Append("blocks", kit.FmtSize(fs.Blocks*uint64(fs.Bsize)))
	m.Append("bavail", kit.FmtSize(fs.Bavail*uint64(fs.Bsize)))
	m.Append("bper", fmt.Sprintf("%d%%", fs.Bavail*100/fs.Blocks))
	m.Table()
}
