package main

import (
	"contexts/cli"
	"contexts/ctx"
	"contexts/web"
	"toolkit"

	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"sort"

	"fmt"
	"os"
	"strings"
	"time"
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

var Index = &ctx.Context{Name: "weixin", Help: "微信后台",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"chat": &ctx.Config{Name: "chat", Value: map[string]interface{}{
			"appid": "", "appmm": "", "token": "", "site": "https://shylinux.com",
			"auth":   "https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=STATE#wechat_redirect",
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
		"access": &ctx.Command{Name: "access", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Option("format", "object")
			m.Option("temp_expire", "1")
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
			m.Option("temp_expire", "1")
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
			m.Append("remote_ip", m.Option("remote_ip"))
			m.Append("auth2.0", fmt.Sprintf(m.Conf("chat", "auth"), m.Conf("chat", "appid"),
				url.QueryEscape(fmt.Sprintf("%s%s", m.Conf("chat", "site"), m.Option("index_url"))), kit.Select("snsapi_base", m.Option("scope"))))
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
			m.Option("sessid", m.Cmdx("aaa.user", "session", "select"))
			m.Option("bench", m.Cmdx("aaa.sess", "bench", "select"))

			m.Option("current_ctx", kit.Select("chat", m.Magic("bench", "current_ctx")))

			switch data.MsgType {
			case "text":
				m.Echo(web.Merge(m, map[string]interface{}{"path": "chat"}, m.Conf("chat", "site"), "sessid", m.Option("sessid")))
				if !m.Cmds("aaa.auth", "username", m.Option("usernmae"), "data", "chat.default") && m.Option("username") != m.Conf("runtime", "work.name") {
					if m.Cmds("ssh.work", "share", m.Option("username")) {
						m.Cmd("aaa.auth", "username", m.Option("username"), "data", "nickname", "someone")
						m.Cmds("aaa.auth", "username", m.Option("username"), "data", "chat.default", m.Spawn().Cmdx(".ocean", "spawn", "", m.Option("username")+"@"+m.Conf("runtime", "work.name")))
					}
				}
				Marshal(m, "text")
				return
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
	},
}

var Target = &web.WEB{Context: Index}

func main() {
	fmt.Print(cli.Index.Plugin(Index, os.Args[1:]))
}
