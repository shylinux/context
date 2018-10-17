package jira

import (
	"contexts/ctx"
	"contexts/web"
)

var Index = &ctx.Context{Name: "jira", Help: "任务中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"/demo": &ctx.Command{Name: "/demo", Help: "demo", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Echo("demo")
		}},
	},
}

func init() {
	jira := &web.WEB{}
	jira.Context = Index
	web.Index.Register(Index, jira)
}
