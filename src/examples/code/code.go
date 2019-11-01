package code

import (
	"contexts/ctx"
	"contexts/web"
)

var Index = &ctx.Context{Name: "code", Help: "代码中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"zsh": &ctx.Command{Name: "zsh", Help: "终端", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo("zsh")
			return
		}},
		"tmux": &ctx.Command{Name: "tmux", Help: "窗口", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo("git")
			return
		}},
		"docker": &ctx.Command{Name: "docker", Help: "容器", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo("docker")
			return
		}},
		"git": &ctx.Command{Name: "git", Help: "版本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo("git")
			return
		}},
		"vim": &ctx.Command{Name: "vim", Help: "编辑器", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo("vim")
			return
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.WEB{Context: Index})
}
