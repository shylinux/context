package chat

import (
	"contexts/ctx"
	"contexts/web"
	"examples/chat"
	"toolkit"

	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func get(m *ctx.Message, arg ...string) *ctx.Message {
	m.Option("temp_expire", -1)
	m.Option("format", "object")
	m.Cmdy("web.get", "feishu", arg, "temp", "data")
	return m
}
func post(m *ctx.Message, arg ...string) *ctx.Message {
	m.Option("temp_expire", -1)
	m.Option("format", "object")
	m.Cmdy("web.get", "method", "POST", "feishu", arg,
		"content_type", "application/json",
		"temp", "data",
	)
	return m
}
func parse(m *ctx.Message) {
	data := m.Optionv("content_data")
	if data == nil {
		json.NewDecoder(m.Optionv("request").(*http.Request).Body).Decode(&data)
		m.Optionv("content_data", data)

		switch d := data.(type) {
		case map[string]interface{}:
			for k, v := range d {
				switch d := v.(type) {
				case map[string]interface{}:
					for k, v := range d {
						m.Add("option", k, v)
					}
				default:
					for _, v := range kit.Trans(v) {
						m.Add("option", "msg."+k, v)
					}
				}
			}
		}
	}
	if kit.Map(kit.Chain(data, "action.value"), "", func(key string, value string) {
		m.Add("option", key, value)
	}) != nil {
		m.Option("msg.type", "event_click")
	}
	m.Log("info", "msg: %v", kit.Formats(data))
}

var Index = &ctx.Context{Name: "feishu", Help: "飞书",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"app":  &ctx.Config{Name: "app", Value: map[string]interface{}{}, Help: "服务配置"},
		"user": &ctx.Config{Name: "user", Value: map[string]interface{}{}, Help: "服务配置"},
	},
	Commands: map[string]*ctx.Command{
		"app": &ctx.Command{Name: "app login|token bot", Help: "应用", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Confm("app", func(key string, value map[string]interface{}) {
					m.Push("key", key)
					m.Push("id", value["id"])
				})
				m.Table()
				return
			}

			switch arg[0] {
			case "login":
				m.Confv("app", arg[1], map[string]interface{}{"id": arg[2], "mm": arg[3]})
			case "token":
				if now := time.Now().Unix(); !m.Confs("app", []string{arg[1], "token"}) || int64(m.Confi("app", []string{arg[1], "expire"})) < now {
					post(m, "auth/v3/tenant_access_token/internal/", "app_id", m.Conf("app", []string{arg[1], "id"}),
						"app_secret", m.Conf("app", []string{arg[1], "mm"}))
					m.Conf("app", []string{arg[1], "token"}, m.Append("tenant_access_token"))
					m.Conf("app", []string{arg[1], "expire"}, int64(m.Appendi("expire"))+now)
					m.Set("result")
				}
				m.Echo(m.Conf("app", []string{arg[1], "token"}))
			}
			return
		}},
		"ship": &ctx.Command{Name: "ship", Help: "组织", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			data := kit.UnMarshal(m.Cmdx("web.get", "feishu", "contact/v1/scope/get/",
				"headers", "Authorization", "Bearer "+m.Cmdx(".app", "token", "bot"),
			))
			kit.Map(kit.Chain(data, "data.authed_departments"), "", func(index int, value string) {
				m.Push("type", "ship")
				m.Push("value", value)
			})
			kit.Map(kit.Chain(data, "data.authed_open_ids"), "", func(index int, value string) {
				m.Push("type", "user")
				m.Push("value", value)
			})
			m.Table()
			return
		}},
		"group": &ctx.Command{Name: "group", Help: "群组", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				kit.Map(kit.Chain(kit.UnMarshal(m.Cmdx("web.get", "feishu", "chat/v4/list", "headers", "Authorization", "Bearer "+m.Cmdx(".app", "token", "bot"))),
					"data.groups"), "", func(index int, value map[string]interface{}) {
					m.Push("key", value["chat_id"])
					m.Push("name", value["name"])
					m.Push("avatar", value["avatar"])
				})
				m.Table()
			}
			kit.Map(kit.Chain(kit.UnMarshal(m.Cmdx("web.get", "feishu", "chat/v4/info", "headers", "Authorization", "Bearer "+m.Cmdx(".app", "token", "bot"))),
				"data.members"), "", func(index int, value map[string]interface{}) {
				m.Push("key", value["open_id"])
			})
			m.Table()
			return
		}},
		"user": &ctx.Command{Name: "user code|email|mobile", Help: "用户", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch arg[0] {
			case "code":
				post(m, "/connect/qrconnect/oauth2/access_token/",
					"app_secret", m.Conf("app", []string{"bot", "mm"}), "app_id", m.Conf("app", []string{"bot", "id"}),
					"grant_type", "authorization_code", "code", arg[1],
				)
				msg := get(m.Spawn(), "/connect/qrconnect/oauth2/user_info/",
					"headers", "Authorization", "Bearer "+m.Append("access_token"),
				)
				m.Confv("user", m.Append("open_id"), map[string]interface{}{
					"name":          m.Append("name"),
					"en_name":       m.Append("en_name"),
					"avatar_url":    m.Append("avatar_url"),
					"access_token":  m.Append("access_token"),
					"token_type":    m.Append("token_type"),
					"expire":        int64(m.Appendi("expire")) + time.Now().Unix(),
					"refresh_token": m.Append("refresh_token"),
					"tenant_key":    m.Append("tenant_key"),

					"email":    msg.Append("Email"),
					"mobile":   msg.Append("Mobile"),
					"status":   msg.Append("status"),
					"employee": msg.Append("EmployeeID"),
				})

			default:
				us := []string{}
				ps := []string{}
				for i := 0; i < len(arg); i++ {
					us = append(us, kit.Select("mobiles", "emails", strings.Contains(arg[i], "@")), arg[i])
					ps = append(ps, kit.Select("mobile", "email", strings.Contains(arg[i], "@"))+"_users")
				}

				data := kit.UnMarshal(m.Cmdx("web.get", "feishu", "user/v1/batch_get_id", us,
					"headers", "Authorization", "Bearer "+m.Cmdx(".app", "token", "bot")))

				for i, v := range ps {
					m.Append(arg[i], kit.Chain(data, []string{"data", v, arg[i], "0", "open_id"}))
				}
				m.Table()
			}
			return
		}},
		"send": &ctx.Command{Name: "send [chat_id|open_id|user_id|email] who [menu] [title] text", Help: "消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			var form = map[string]interface{}{"content": map[string]interface{}{}}
			switch arg[0] {
			case "chat_id", "open_id", "user_id", "email":
				form[arg[0]], arg = arg[1], arg[2:]
			default:
				form["chat_id"], arg = arg[0], arg[1:]
			}

			switch arg[0] {
			case "menu":
				elements := []interface{}{}
				elements = append(elements, map[string]interface{}{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "plain_text",
						"content": arg[2],
					},
				})

				actions := []interface{}{}
				for i := 3; i < len(arg); i++ {
					button := map[string]interface{}{
						"tag": "button", "text": map[string]interface{}{
							"tag": "plain_text", "content": arg[i],
						},
						"type": "default",
					}

					switch arg[i+1] {
					case "confirm":
						button[arg[i+1]], i = map[string]interface{}{
							"title": map[string]interface{}{"tag": "lark_md", "content": arg[i+2]},
							"text":  map[string]interface{}{"tag": "lark_md", "content": arg[i+3]},
						}, i+3
					case "value":
						button[arg[i+1]], i = map[string]interface{}{
							arg[i+2]: arg[i+3],
						}, i+3
					case "url":
						button[arg[i+1]], i = arg[i+2], i+2
					default:
						button["value"], i = map[string]interface{}{
							arg[i+1]: arg[i+2],
						}, i+2
					}

					actions = append(actions, button)
				}
				elements = append(elements, map[string]interface{}{
					"tag": "action", "actions": actions,
				})

				kit.Chain(form, "msg_type", "interactive")
				kit.Chain(form, "card", map[string]interface{}{
					"config": map[string]interface{}{
						"wide_screen_mode": true,
						// "title": map[string]interface{}{
						// 	"tag": "lark_md", "content": arg[1],
						// },
					},
					"header": map[string]interface{}{
						"title": map[string]interface{}{
							"tag": "lark_md", "content": arg[1],
						},
					},
					"elements": elements,
				})
			default:
				switch len(arg) {
				case 0:
				case 1:
					kit.Chain(form, "msg_type", "text")
					kit.Chain(form, "content.text", arg[0])
				default:
					content := []interface{}{}
					line := []interface{}{}
					for _, v := range arg[1:] {
						if v == "\n" {
							content, line = append(content, line), []interface{}{}
							continue
						}
						line = append(line, map[string]interface{}{
							"tag": "text", "text": v,
						})
					}
					content = append(content, line)

					kit.Chain(form, "msg_type", "post")
					kit.Chain(form, "content.post", map[string]interface{}{
						"zh_cn": map[string]interface{}{
							"title":   arg[0],
							"content": content,
						},
					})
				}

			}

			m.Cmdy("web.get", "method", "POST", "feishu", "message/v4/send/",
				"headers", "Authorization", "Bearer "+m.Cmdx(".app", "token", "bot"),
				"content_data", kit.Formats(form), "content_type", "application/json",
				"temp", "data", "data.message_id",
			)
			return
		}},
		"/msg": &ctx.Command{Name: "/msg", Help: "消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			parse(m)

			switch m.Option("msg.type") {
			case "url_verification":
				m.Echo(kit.Format(map[string]interface{}{"challenge": m.Option("challenge")}))

			case "event_callback":
				switch m.Option("type") {
				case "chat_disband":
				case "p2p_chat_create":
					if m.Options("open_chat_id") {
						m.Cmdy(".send", m.Option("open_chat_id"), "我们做朋友吧~")
					}
				case "add_bot":
					if m.Options("open_chat_id") {
						m.Cmdy(".send", m.Option("open_chat_id"), "我来也~")
					}
				default:
					if m.Options("open_chat_id") {
						m.Option("username", m.Option("open_id"))
						m.Option("sessid", m.Cmdx("aaa.user", "session", "select"))
						m.Cmd("ssh._check", "work", "create", m.Option("username"))
						m.Goshy([]string{m.Option("text_without_at_bot")}, 0, nil, func(msg *ctx.Message) {
							m.Cmdy(".send", m.Option("open_chat_id"), kit.Select("你好", strings.Join(msg.Meta["result"], "")))
						})
					}
				}
			case "event_click":
				m.Echo(kit.Format(map[string]interface{}{
					"header": map[string]interface{}{
						"title": map[string]interface{}{
							"tag": "lark_md", "content": "haha",
						},
					},
					"elements": []interface{}{
						map[string]interface{}{
							"tag": "action",
							"actions": []interface{}{
								map[string]interface{}{
									"tag":  "button",
									"type": "default",
									"text": map[string]interface{}{
										"tag":     "plain_text",
										"content": m.Time(),
									},
									"value": map[string]interface{}{
										"hi": "hello",
									},
								},
							},
						},
					},
				}))
			}
			return
		}},
		"/sso": &ctx.Command{Name: "/sso", Help: "消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if m.Options("code") {
				m.Option("username", m.Cmd(".user", "code", m.Option("code")).Append("open_id"))
				m.Option("sessid", m.Cmdx("aaa.user", "session", "select"))
				m.Cmd("ssh._check", "work", "create", m.Option("username"))

				w := m.Optionv("response").(http.ResponseWriter)
				r := m.Optionv("request").(*http.Request)
				web.Cookie(m, w, r)
				m.Append("redirect", m.Cmdx("web.spide", "serve", "merge", m.Option("index_path")), "code", "")
				return
			}

			if !m.Options("sessid") || !m.Options("username") {
				m.Append("redirect", m.Cmdx("web.spide", "feishu", "merge", "/connect/qrconnect/page/sso/",
					"redirect_uri", m.Cmdx("web.spide", "serve", "merge", m.Option("index_path")),
					"app_id", m.Conf("app", "bot.id"), "state", "ok"))
				return
			}
			m.Cmd("/render")
			return
		}},
	},
}

func init() {
	chat.Index.Register(Index, &web.WEB{Context: Index})
}
