package lark

import (
	"contexts/ctx"
	"contexts/web"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var Index = &ctx.Context{Name: "lark", Help: "会议中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"lark_msg": &ctx.Config{Name: "lark_msg", Value: []interface{}{}, Help: "聊天记录"},
	},
	Commands: map[string]*ctx.Command{
		"/lark": &ctx.Command{Name: "user", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			r := m.Optionv("request").(*http.Request)
			w := m.Optionv("response").(http.ResponseWriter)

			data := map[string]interface{}{}
			switch r.Header.Get("Content-Type") {
			case "application/json":
				b, e := ioutil.ReadAll(r.Body)
				e = json.Unmarshal(b, &data)
				m.Assert(e)
			}

			if _, ok := data["challenge"]; ok {
				w.Header().Set("Content-Type", "application/javascript")
				fmt.Fprintf(w, "{\"challenge\": \"%s\"}", data["challenge"])
				return
			}
			m.Confv("lark_msg", "-1", data)
		}},
	},
}

func init() {
	lark := &web.WEB{}
	lark.Context = Index
	web.Index.Register(Index, lark)
}
