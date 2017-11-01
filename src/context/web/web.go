package web

import (
	"context"
	"net/http"
)

type WEB struct {
	*ctx.Context
}

func (web *WEB) Begin() bool {

	return true
}
func (web *WEB) Start() bool {
	if web.Cap("status") == "start" {
		return true
	}
	web.Cap("status", "start")
	defer web.Cap("status", "stop")

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
	Caches: map[string]*ctx.Cache{
		"status": &ctx.Cache{"status", "stop", "服务运行状态", nil},
	},
	Configs: map[string]*ctx.Config{
		"path":    &ctx.Config{"path", "srv", "监听地址", nil},
		"address": &ctx.Config{"address", ":9494", "监听地址", nil},
	},
}

func init() {
	web := &WEB{}
	web.Context = Index
	ctx.Index.Register(Index, web)
}
