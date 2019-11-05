package team

import (
	"contexts/ctx"
	"contexts/web"
)

var Index = &ctx.Context{Name: "team", Help: "团队中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"task": {Name: "task table title content", Help: "任务", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 1 {
				m.Cmdy("ssh.data", "show", arg[0], "fields", "id", "title", "content")
				return
			}
			m.Cmdy("ssh.data", "insert", arg[0], "title", arg[1], "content", arg[2], arg[3:])
			return
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.WEB{Context: Index})
}
