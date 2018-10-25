package mall

import (
	"contexts/ctx"
	"contexts/web"
)

var Index = &ctx.Context{Name: "mall", Help: "交易中心",
	Caches:   map[string]*ctx.Cache{},
	Configs:  map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{},
}

func init() {
	mall := &web.WEB{}
	mall.Context = Index
	web.Index.Register(Index, mall)
}
