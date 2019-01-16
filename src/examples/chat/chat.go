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
		"chat_msg":      &ctx.Config{Name: "chat_msg", Value: []interface{}{}, Help: "聊天记录"},
		"default":       &ctx.Config{Name: "default", Value: "", Help: "聊天记录"},
		"weather_site":  &ctx.Config{Name: "weather_site", Value: "http://weather.sina.com.cn", Help: "聊天记录"},
		"calendar_site": &ctx.Config{Name: "calendar_site", Value: "http://tools.2345.com/rili.htm", Help: "聊天记录"},
		"topic_site":    &ctx.Config{Name: "topic_site", Value: "https://s.weibo.com/top/summary?cate=realtimehot", Help: "聊天记录"},
		"pedia_site":    &ctx.Config{Name: "pedia_site", Value: "https://zh.wikipedia.org/wiki", Help: "聊天记录"},
		"baike_site":    &ctx.Config{Name: "baike_site", Value: "https://baike.baidu.com/item", Help: "聊天记录"},
		"sinas_site":    &ctx.Config{Name: "sinas_site", Value: "http://www.sina.com.cn/mid/search.shtml?range=all&c=news&q=%s&from=home&ie=utf-8", Help: "聊天记录"},
		"zhihu_site":    &ctx.Config{Name: "zhihu_site", Value: "https://www.zhihu.com/search?type=content&q=%s", Help: "聊天记录"},
		"toutiao_site":  &ctx.Config{Name: "toutiao_site", Value: "https://www.toutiao.com/search/?keyword=%s", Help: "聊天记录"},
	},
	Commands: map[string]*ctx.Command{
		"/chat": &ctx.Command{Name: "user", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
			return
		}},
		"talk": &ctx.Command{Name: "talk", Help: "talk", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if m.Confs("default") {
				m.Echo(m.Conf("default"))
			}
			return
		}},
		"weather": &ctx.Command{Name: "weather where field", Help: "weather", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			where := "beijing"
			if len(arg) > 0 {
				where, arg = arg[0], arg[1:]
			}

			msg := m.Spawn().Cmd("get", fmt.Sprintf("%s/%s", m.Conf("weather_site"), where),
				"parse", "div.blk_fc_c0_i",
				"sub_parse", "date", "p.wt_fc_c0_i_date", "text",
				"sub_parse", "day", "p.wt_fc_c0_i_day", "text",
				"sub_parse", "weather", "p.wt_fc_c0_i_icons.clearfix img", "title",
				"sub_parse", "temp", "p.wt_fc_c0_i_temp", "text",
				"sub_parse", "wind", "p.wt_fc_c0_i_tip", "text",
				"sub_parse", "pm", "ul.wt_fc_c0_i_level li.l", "text",
				"sub_parse", "env", "ul.wt_fc_c0_i_level li.r", "text",
			)

			m.Copy(msg, "append").Copy(msg, "result")

			if len(arg) == 0 {
				arg = append(arg, "temp")
			}

			switch arg[0] {
			case "all":
			case "temp":
				m.Cmd("select", "fields", "date day weather temp")
			case "wind":
				m.Cmd("select", "fields", "date day weather wind")
			case "env":
				m.Cmd("select", "fields", "date day weather pm env")
			default:
				m.Cmd("select", "date", arg[0], "vertical")
			}
			return
		}},
		"calendar": &ctx.Command{Name: "calendar", Help: "calendar", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			msg := m.Spawn().Cmd("get", m.Conf("calendar_site"),
				"parse", "div.almanac-hd")
			m.Copy(msg, "append").Copy(msg, "result")
			return
		}},
		"topic": &ctx.Command{Name: "topic", Help: "topic", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			limit := "10"
			if len(arg) > 0 {
				limit, arg = arg[0], arg[1:]
			}

			msg := m.Spawn().Cmd("get", m.Conf("topic_site"),
				"parse", "table tr",
				"sub_parse", "mark", "td.td-03", "text",
				"sub_parse", "count", "td.td-02 span", "text",
				"sub_parse", "rank", "td.td-01", "text",
				"sub_parse", "topic", "td.td-02 a", "text",
			)

			m.Copy(msg, "append").Copy(msg, "result")
			m.Cmd("select", "limit", limit)
			return
		}},
		"pedia": &ctx.Command{Name: "pedia", Help: "pedia", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			msg := m.Spawn().Cmd("get", fmt.Sprintf("%s/%s", m.Conf("pedia_site"), arg[0]),
				"parse", "div.mw-parser-output>p,div.mw-parser-output>ul",
				"sub_parse", "content", "", "text",
			)
			arg = arg[1:]

			offset := "0"
			if len(arg) > 0 {
				offset, arg = arg[0], arg[1:]
			}

			limit := "3"
			if len(arg) > 0 {
				limit, arg = arg[0], arg[1:]
			}

			m.Copy(msg, "append").Copy(msg, "result")
			m.Cmd("select", "limit", limit, "offset", offset)
			return
		}},
		"baike": &ctx.Command{Name: "baike", Help: "baike", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			msg := m.Spawn().Cmd("get", fmt.Sprintf("%s/%s", m.Conf("baike_site"), arg[0]),
				"parse", "div.mw-body",
				"sub_parse", "content", "p", "text",
			)
			arg = arg[1:]

			offset := "0"
			if len(arg) > 0 {
				offset, arg = arg[0], arg[1:]
			}

			limit := "3"
			if len(arg) > 0 {
				limit, arg = arg[0], arg[1:]
			}

			m.Copy(msg, "append").Copy(msg, "result")
			m.Cmd("select", "limit", limit, "offset", offset)
			return
		}},
		"sinas": &ctx.Command{Name: "sinas", Help: "sinas", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			msg := m.Spawn().Cmd("get", fmt.Sprintf(m.Conf("sinas_site"), arg[0]),
				"parse", "div.box-result.clearfix",
				"sub_parse", "title", "h2", "text",
			)
			arg = arg[1:]

			offset := "0"
			if len(arg) > 0 {
				offset, arg = arg[0], arg[1:]
			}

			limit := "3"
			if len(arg) > 0 {
				limit, arg = arg[0], arg[1:]
			}

			m.Copy(msg, "append").Copy(msg, "result")
			m.Cmd("select", "limit", limit, "offset", offset)
			return
		}},
		"zhihu": &ctx.Command{Name: "zhihu", Help: "zhihu", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			msg := m.Spawn().Cmd("get", fmt.Sprintf(m.Conf("zhihu_site"), arg[0]),
				"parse", "div.SearchMain div.Card.SearchResult-Card",
				"sub_parse", "title", "", "text",
			)
			arg = arg[1:]

			offset := "0"
			if len(arg) > 0 {
				offset, arg = arg[0], arg[1:]
			}

			limit := "3"
			if len(arg) > 0 {
				limit, arg = arg[0], arg[1:]
			}

			m.Copy(msg, "append").Copy(msg, "result")
			m.Cmd("select", "limit", limit, "offset", offset)
			return
		}},

		"toutiao": &ctx.Command{Name: "toutiao", Help: "toutiao", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			msg := m.Spawn().Cmd("get", fmt.Sprintf(m.Conf("toutiao_site"), arg[0]),
				"parse", "div.articleCard",
				"sub_parse", "title", "", "text",
			)
			arg = arg[1:]

			offset := "0"
			if len(arg) > 0 {
				offset, arg = arg[0], arg[1:]
			}

			limit := "3"
			if len(arg) > 0 {
				limit, arg = arg[0], arg[1:]
			}

			m.Copy(msg, "append").Copy(msg, "result")
			m.Cmd("select", "limit", limit, "offset", offset)
			return
		}},
	},
}

func init() {
	chat := &web.WEB{}
	chat.Context = Index
	web.Index.Register(Index, chat)
}
