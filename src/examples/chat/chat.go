package chat

import (
	"contexts/ctx"
	"contexts/web"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var Index = &ctx.Context{Name: "chat", Help: "会议中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"chat_msg": &ctx.Config{Name: "chat_msg", Value: []interface{}{}, Help: "聊天记录"},
	},
	Commands: map[string]*ctx.Command{
		"/chat": &ctx.Command{Name: "user", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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
			m.Confv("chat_msg", "-1", data)
		}},
	},
}

func init() {
	chat := &web.WEB{}
	chat.Context = Index
	web.Index.Register(Index, chat)
}
