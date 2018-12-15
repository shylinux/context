package team

import (
	"contexts/ctx"
	"contexts/web"
)

var Index = &ctx.Context{Name: "team", Help: "任务中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"/demo": &ctx.Command{Name: "/demo", Help: "demo", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Echo("demo")
		}},
	},
}

func init() {
	team := &web.WEB{}
	team.Context = Index
	web.Index.Register(Index, team)
}
