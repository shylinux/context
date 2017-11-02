package web

import (
	"context"
	"context/cli"
	"net/http"
)

type WEB struct {
	*ctx.Context
}

func (web *WEB) Begin() bool {

	return true
}
func (web *WEB) Start() bool {
	http.Handle("/", http.FileServer(http.Dir(web.Conf("path"))))
	http.ListenAndServe(web.Conf("address"), nil)
	return true
}
func (web *WEB) Spawn(c *ctx.Context, key string) ctx.Server {

	return nil
}
func (web *WEB) Fork(c *ctx.Context, key string) ctx.Server {
	return nil
}

var Index = &ctx.Context{Name: "web", Help: "网页服务",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"path":    &ctx.Config{Name: "path", Value: "srv", Help: "监听地址"},
		"address": &ctx.Config{Name: "address", Value: ":9494", Help: "监听地址"},
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	cli.Index.Register(Index, web)
}
