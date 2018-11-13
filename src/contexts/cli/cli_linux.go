package cli

import (
	"contexts/ctx"
	"fmt"
	"runtime"
	"syscall"
	"time"
	"toolkit"
)

func sysinfo(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
	sys := &syscall.Sysinfo_t{}
	syscall.Sysinfo(sys)

	d, e := time.ParseDuration(fmt.Sprintf("%ds", sys.Uptime))
	m.Assert(e)
	m.Append("NumCPU", runtime.NumCPU())
	m.Append("uptime", d)
	m.Append("procs", sys.Procs)

	m.Append("total", kit.FmtSize(uint64(sys.Totalram)))
	m.Append("free", kit.FmtSize(uint64(sys.Freeram)))
	m.Append("mper", fmt.Sprintf("%d%%", sys.Freeram*100/sys.Totalram))

	fs := &syscall.Statfs_t{}
	syscall.Statfs("./", fs)
	m.Append("blocks", kit.FmtSize(fs.Blocks*uint64(fs.Bsize)))
	m.Append("bavail", kit.FmtSize(fs.Bavail*uint64(fs.Bsize)))
	m.Append("bper", fmt.Sprintf("%d%%", fs.Bavail*100/fs.Blocks))

	m.Append("files", fs.Files)
	m.Append("ffree", fs.Ffree)
	m.Append("fper", fmt.Sprintf("%d%%", fs.Ffree*100/fs.Files))

	m.Table()
}
