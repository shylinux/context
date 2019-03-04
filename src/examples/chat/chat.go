package chat

import (
	"contexts/ctx"
	"contexts/web"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"toolkit"
)

func Marshal(m *ctx.Message, meta string) string {
	b, e := xml.Marshal(struct {
		CreateTime   int64
		FromUserName string
		ToUserName   string
		MsgType      string
		Content      string
		XMLName      xml.Name `xml:"xml"`
	}{
		time.Now().Unix(),
		m.Option("selfname"), m.Option("username"),
		meta, strings.Join(m.Meta["result"], ""), xml.Name{},
	})
	m.Assert(e)
	m.Set("append").Set("result").Echo(string(b))
	return string(b)
}

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

		"chat": &ctx.Config{Name: "chat", Value: map[string]interface{}{
			"appid": "", "appmm": "", "token": "", "site": "https://shylinux.com",
			"access": map[string]interface{}{"token": "", "expire": 0, "url": "/cgi-bin/token?grant_type=client_credential"},
			"ticket": map[string]interface{}{"value": "", "expire": 0, "url": "/cgi-bin/ticket/getticket?type=jsapi"},
		}, Help: "聊天记录"},
		"mp": &ctx.Config{Name: "chat", Value: map[string]interface{}{
			"appid": "", "appmm": "", "token": "", "site": "https://shylinux.com",
			"auth":         "/sns/jscode2session?grant_type=authorization_code",
			"tool_path":    "/Applications/wechatwebdevtools.app/Contents/MacOS/cli",
			"project_path": "/Users/shaoying/context/usr/client/mp",
		}, Help: "聊天记录"},
	},
	Commands: map[string]*ctx.Command{
		"/chat": &ctx.Command{Name: "user", Help: "应用示例", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 信息验证
			nonce := []string{m.Option("timestamp"), m.Option("nonce"), m.Conf("chat", "token")}
			sort.Strings(nonce)
			h := sha1.Sum([]byte(strings.Join(nonce, "")))
			if hex.EncodeToString(h[:]) == m.Option("signature") {
				// m.Echo(m.Option("echostr"))
			} else {
				return
			}

			// 解析数据
			var data struct {
				MsgId        int64
				CreateTime   int64
				ToUserName   string
				FromUserName string
				MsgType      string
				Content      string
			}
			r := m.Optionv("request").(*http.Request)
			m.Assert(xml.NewDecoder(r.Body).Decode(&data))
			m.Option("username", data.FromUserName)
			m.Option("selfname", data.ToUserName)

			// 创建会话
			if m.Option("sessid", m.Cmd("aaa.user", m.Option("username", data.FromUserName), "chat").Append("key")) == "" {
				m.Cmd("aaa.sess", m.Option("sessid", m.Cmdx("aaa.sess", "chat", "ip", "what")), m.Option("username"), "ppid", "what")
			}

			// 创建空间
			if m.Option("bench", m.Cmd("aaa.sess", m.Option("sessid"), "bench").Append("key")) == "" {
				m.Option("bench", m.Cmdx("aaa.work", m.Option("sessid"), "chat"))
			}
			m.Option("current_ctx", kit.Select("chat", m.Magic("bench", "current_ctx")))

			switch data.MsgType {
			case "text":
				// 执行命令
				cmd := strings.Split(data.Content, " ")
				if !m.Cmds("aaa.work", m.Option("bench"), "right", data.FromUserName, "chat", cmd[0]) {
					m.Echo("no right %s %s", "chat", cmd[0])
				} else if m.Cmdy("cli.source", data.Content); m.Appends("redirect") {
				}
				Marshal(m, "text")
			}
			return
		}},
		"access": &ctx.Command{Name: "access", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Option("format", "object")
			now := kit.Int(time.Now().Unix())

			access := m.Confm("chat", "access")
			if kit.Int(access["expire"]) < now {
				msg := m.Cmd("web.get", "wexin", access["url"], "appid", m.Conf("chat", "appid"), "secret", m.Conf("chat", "appmm"), "temp", "data")
				access["token"] = msg.Append("access_token")
				access["expire"] = msg.Appendi("expires_in") + now
			}
			m.Echo("%v", access["token"])
			return
		}},
		"ticket": &ctx.Command{Name: "ticket", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Option("format", "object")
			now := kit.Int(time.Now().Unix())

			ticket := m.Confm("chat", "ticket")
			if kit.Int(ticket["expire"]) < now {
				msg := m.Cmd("web.get", "wexin", ticket["url"], "access_token", m.Cmdx(".access"), "temp", "data")
				ticket["value"] = msg.Append("ticket")
				ticket["expire"] = msg.Appendi("expires_in") + now
			}
			m.Echo("%v", ticket["value"])
			return
		}},
		"js_token": &ctx.Command{Name: "js_token", Help: "zhihu", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			nonce := []string{
				"jsapi_ticket=" + m.Cmdx(".ticket"),
				"noncestr=" + m.Append("nonce", "what"),
				"timestamp=" + m.Append("timestamp", kit.Int(time.Now())),
				"url=" + m.Append("url", m.Conf("chat", "site")+m.Option("index_url")),
			}
			sort.Strings(nonce)
			h := sha1.Sum([]byte(strings.Join(nonce, "&")))

			m.Append("signature", hex.EncodeToString(h[:]))
			m.Append("appid", m.Conf("chat", "appid"))
			return
		}},
		"share": &ctx.Command{Name: "share", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo("%s?bench=%s&sessid=%s", m.Conf("chat", "site"), m.Option("bench"), m.Option("sessid"))
			return
		}},
		"check": &ctx.Command{Name: "check", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			sort.Strings(arg)
			h := sha1.Sum([]byte(strings.Join(arg, "")))
			if hex.EncodeToString(h[:]) == m.Option("signature") {
				m.Echo("true")
			}
			return
		}},

		"/mp": &ctx.Command{Name: "/mp", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 用户登录
			if m.Options("code") {
				m.Option("format", "object")
				msg := m.Cmd("web.get", "wexin", m.Conf("mp", "auth"), "js_code", m.Option("code"), "appid", m.Conf("mp", "appid"), "secret", m.Conf("mp", "appmm"), "parse", "json", "temp", "data")

				// 创建会话
				if !m.Options("sessid") {
					m.Cmd("aaa.sess", m.Option("sessid", m.Cmdx("aaa.sess", "mp", "ip", "what")), msg.Append("openid"), "ppid", "what")
					defer func() {
						m.Set("result").Echo(m.Option("sessid"))
					}()
				}

				m.Magic("session", "user.openid", msg.Append("openid"))
				m.Magic("session", "user.expires_in", kit.Int(msg.Append("expires_in"), time.Now()))
				m.Magic("session", "user.session_key", msg.Append("session_key"))
			}

			// 用户信息
			if m.Options("userInfo") && m.Options("rawData") {
				h := sha1.Sum([]byte(strings.Join([]string{m.Option("rawData"), kit.Format(m.Magic("session", "user.session_key"))}, "")))
				if hex.EncodeToString(h[:]) == m.Option("signature") {
					var info interface{}
					json.Unmarshal([]byte(m.Option("userInfo")), &info)
					m.Log("info", "user %v %v", m.Option("sessid"), info)

					m.Magic("session", "user.info", info)
					m.Magic("session", "user.encryptedData", m.Option("encryptedData"))
					m.Magic("session", "user.iv", m.Option("iv"))
				}
			}

			if m.Option("username", m.Magic("session", "user.openid")) == "" || m.Option("cmd") == "" {
				return
			}

			if m.Option("username") == "o978M0XIrcmco28CU1UbPgNxIL78" {
				m.Option("username", "shy")
			}

			// 创建空间
			if !m.Options("bench") && m.Option("bench", m.Cmd("aaa.sess", m.Option("sessid"), "bench").Append("key")) == "" {
				m.Option("bench", m.Cmdx("aaa.work", m.Option("sessid"), "mp"))
			}
			m.Option("current_ctx", kit.Select("chat", m.Magic("bench", "current_ctx")))

			// 执行命令
			cmd := kit.Trans(m.Optionv("cmd"))
			if !m.Cmds("aaa.work", m.Option("bench"), "right", m.Option("username"), "mp", cmd[0]) {
				m.Echo("no right %s %s", "chat", cmd[0])
			} else if m.Cmdy(cmd); m.Appends("redirect") {
			}
			return
		}},
		"mp": &ctx.Command{Name: "mp", Help: "talk", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy("cli.system", m.Conf("mp", "tool_path"), arg, m.Conf("mp", "project_path"), "cmd_active", "true")
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
