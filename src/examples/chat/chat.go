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
		"login":          &ctx.Config{Name: "login", Value: map[string]interface{}{"check": "false"}, Help: "默认组件"},
		"componet_group": &ctx.Config{Name: "component_group", Value: "index", Help: "默认组件"},
		"componet": &ctx.Config{Name: "componet", Value: map[string]interface{}{
			"index": []interface{}{
				map[string]interface{}{"componet_name": "chat", "componet_tmpl": "head", "metas": []interface{}{
					map[string]interface{}{"name": "viewport", "content": "width=device-width, initial-scale=0.7, user-scalable=no"},
				}, "favicon": "favicon.ico", "styles": []interface{}{"example.css", "chat.css"}},
				map[string]interface{}{"componet_name": "header", "componet_tmpl": "fieldset",
					"componet_view": "Header", "componet_init": "initHeader",
					"title": "shylinux 天行健，君子以自强不息",
				},

				map[string]interface{}{"componet_name": "ocean", "componet_tmpl": "fieldset",
					"componet_view": "Ocean", "componet_init": "initOcean",
					"componet_ctx": "web.chat", "componet_cmd": "ocean",
				},
				map[string]interface{}{"componet_name": "steam", "componet_tmpl": "fieldset",
					"componet_view": "Steam", "componet_init": "initSteam",
					"componet_ctx": "web.chat", "componet_cmd": "steam",
				},
				map[string]interface{}{"componet_name": "river", "componet_tmpl": "fieldset",
					"componet_view": "River", "componet_init": "initRiver",
					"componet_ctx": "web.chat", "componet_cmd": "river",
				},
				map[string]interface{}{"componet_name": "storm", "componet_tmpl": "fieldset",
					"componet_view": "Storm", "componet_init": "initStorm",
					"componet_ctx": "web.chat", "componet_cmd": "storm",
				},

				map[string]interface{}{"componet_name": "target", "componet_tmpl": "fieldset",
					"componet_view": "Target", "componet_init": "initTarget",
					"componet_ctx": "web.chat", "componet_cmd": "river",
				},
				map[string]interface{}{"componet_name": "source", "componet_tmpl": "fieldset",
					"componet_view": "Source", "componet_init": "initSource",
					"componet_ctx": "web.chat", "componet_cmd": "storm",
				},
				map[string]interface{}{"componet_name": "action", "componet_tmpl": "fieldset",
					"componet_view": "Action", "componet_init": "initAction",
					"componet_ctx": "web.chat", "componet_cmd": "storm",
				},

				map[string]interface{}{"componet_name": "footer", "componet_tmpl": "fieldset",
					"componet_view": "Footer", "componet_init": "initFooter",
					"title": "shycontext 地势坤，君子以厚德载物",
				},
				map[string]interface{}{"componet_name": "tail", "componet_tmpl": "tail",
					"scripts": []interface{}{"toolkit.js", "context.js", "example.js", "chat.js"},
				},
			},
		}, Help: "组件列表"},

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

		"flow": &ctx.Config{Name: "flow", Value: map[string]interface{}{}, Help: "聊天记录"},
	},
	Commands: map[string]*ctx.Command{
		"ocean": &ctx.Command{Name: "ocean", Help: "海洋", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmdy("ssh.work", "search")
				return
			}

			switch arg[0] {
			case "spawn":
				h := kit.Select(kit.Hashs("uniq"), arg, 1)
				user := map[string]interface{}{}
				for _, v := range arg[3:] {
					u := m.Cmdx("ssh._route", m.Conf("runtime", "work.route"), "_check", "work", v)
					user[v] = map[string]interface{}{
						"user": u,
					}
				}

				m.Conf("flow", h, map[string]interface{}{
					"conf": map[string]interface{}{
						"create_user": m.Option("username"),
						"create_time": m.Time(),
						"name":        kit.Select("what", arg, 2),
						"route":       kit.Select(m.Conf("runtime", "node.route"), m.Option("node.route"), arg[1] != ""),
					},
					"user": user,
					"text": map[string]interface{}{},
					"tool": map[string]interface{}{},
				})
				m.Echo(h)

				m.Option("username", m.Conf("runtime", "user.name"))
				m.Confm("flow", []string{h, "user"}, func(key string, value map[string]interface{}) {
					if kit.Format(value["user"]) != m.Conf("runtime", "node.route") {
						m.Cmd("ssh._route", value["user"], "context", "chat", "ocean", "spawn", h, arg[2])
					}
				})
			}
			return
		}},
		"river": &ctx.Command{Name: "river", Help: "河流", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Confm("flow", func(key string, value map[string]interface{}) {
					m.Add("append", "key", key)
					m.Add("append", "name", kit.Chains(value, "conf.name"))
					m.Add("append", "create_user", kit.Chains(value, "conf.create_user"))
					m.Add("append", "create_time", kit.Chains(value, "conf.create_time"))

					if list, ok := kit.Chain(value, "text.list").([]interface{}); ok {
						m.Add("append", "count", len(list))
					} else {
						m.Add("append", "count", 0)
					}
				})
				m.Table()
				return
			}

			switch arg[0] {
			case "flow":
				if len(arg) == 2 {
					m.Confm("flow", []string{arg[1], "text.list"}, func(index int, value map[string]interface{}) {
						m.Add("append", "index", index)
						m.Add("append", "type", value["type"])
						m.Add("append", "text", value["text"])
					})
					m.Table()
					return
				}

				if m.Conf("flow", []string{arg[1], "conf.route"}) != m.Conf("runtime", "node.route") && len(arg) == 4 {
					m.Cmdy("ssh._route", m.Conf("flow", []string{arg[1], "conf.route"}),
						"context", "chat", "river", "flow", arg[1], arg[2], arg[3])
					m.Log("info", "upstream")
					return
				}

				m.Conf("flow", []string{arg[1], "text.list.-2"}, map[string]interface{}{
					"create_user": m.Option("username"),
					"create_time": m.Time(),
					"type":        arg[2],
					"text":        arg[3],
				})

				count := m.Confi("flow", []string{arg[1], "text.count"}) + 1
				m.Confi("flow", []string{arg[1], "text.count"}, count)
				m.Echo("%d", count)

				m.Option("username", m.Conf("runtime", "user.name"))
				m.Confm("flow", []string{arg[1], "user"}, func(key string, value map[string]interface{}) {
					if kit.Format(value["user"]) != m.Conf("runtime", "node.route") {
						m.Cmd("ssh._route", value["user"], "context", "chat", "river", "flow", arg[1], arg[2], arg[3], "sync")
					}
				})

			case "wave":
				m.Option("username", "shy")
				m.Cmdy("ssh._route", arg[2], "tool", "run", arg[3], arg[4], arg[5:])
			}
			return
		}},
		"storm": &ctx.Command{Name: "storm", Help: "风雨", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 1 {
				m.Confm("flow", []string{arg[0], "tool"}, func(key string, value map[string]interface{}) {
					m.Add("append", "key", key)
					m.Add("append", "count", kit.Len(value["list"]))
				})
				m.Table()
				return
			}
			if len(arg) == 2 {
				m.Confm("flow", []string{arg[0], "tool", arg[1], "list"}, func(index int, tool map[string]interface{}) {
					m.Add("append", "river", arg[0])
					m.Add("append", "storm", arg[1])
					m.Add("append", "action", index)

					m.Add("append", "node", tool["node"])
					m.Add("append", "group", tool["group"])
					m.Add("append", "index", tool["index"])

					m.Option("username", "shy")
					msg := m.Cmd("ssh._route", tool["node"], "tool", tool["group"], tool["index"])

					m.Add("append", "name", msg.Append("name"))
					m.Add("append", "help", msg.Append("help"))
					m.Add("append", "view", msg.Append("view"))
					m.Add("append", "init", msg.Append("init"))
					m.Add("append", "inputs", msg.Append("inputs"))
				})
				m.Table()
				return
			}

			if tool := m.Confm("flow", []string{arg[0], "tool", arg[1], "list", arg[2]}); tool != nil {
				m.Option("username", "shy")
				m.Cmdy("ssh._route", tool["node"], "tool", "run", tool["group"], tool["index"], arg[3:])
				return
			}
			return
		}},
		"steam": &ctx.Command{Name: "steam", Help: "天空", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if m.Conf("flow", []string{arg[0], "conf.route"}) != m.Conf("runtime", "node.route") {
				m.Cmdy("ssh._remote", m.Conf("flow", []string{arg[0], "conf.route"}), "context", "chat", "steam", arg)
				m.Log("info", "upstream")
				return
			}
			if len(arg) == 1 {
				m.Confm("flow", []string{arg[0], "user"}, func(key string, value map[string]interface{}) {
					m.Add("append", "key", key)
					m.Add("append", "user.route", value["user"])
				})
				m.Table()
				return
			}

			switch arg[1] {
			case "spawn":
				list := []interface{}{}
				for i := 3; i < len(arg)-3; i += 4 {
					list = append(list, map[string]interface{}{
						"node": arg[i], "group": arg[i+1], "index": arg[i+2], "name": arg[i+3],
					})
				}

				m.Conf("flow", []string{arg[0], "tool", arg[2]}, map[string]interface{}{
					"create_user": m.Option("username"),
					"create_time": m.Time(),
					"list":        list,
				})

			default:
				m.Option("username", "shy")
				m.Cmdy("ssh._route", m.Conf("flow", []string{arg[0], "user", arg[1], "user"}), "tool")
			}
			return
		}},

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
				access["expire"] = int(msg.Appendi("expires_in")) + now
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
				ticket["expire"] = int(msg.Appendi("expires_in")) + now
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
			if m.Option("username") == "o978M0ff_Y76hFu1FPLif6hFfmsM" {
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
