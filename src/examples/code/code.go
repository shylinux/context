package code

import (
	"contexts/ctx"
	"contexts/web"
)

type CODE struct {
	web.WEB
}

var Index = &ctx.Context{Name: "code", Help: "代码中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"/demo": &ctx.Command{Name: "/demo", Help: "demo", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Echo("demo")
		}},
	},
}

func init() {
	code := &CODE{}
	code.Context = Index
	web.Index.Register(Index, code)
}
