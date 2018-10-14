package ssh

import (
	"contexts"
	"fmt"
	"strings"
)

type SSH struct {
	peer map[string]*ctx.Message
	*ctx.Context
}

func (ssh *SSH) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(SSH)
	s.Context = c
	return s
}
func (ssh *SSH) Begin(m *ctx.Message, arg ...string) ctx.Server {
	ssh.Caches["hostname"] = &ctx.Cache{Name: "hostname", Value: "", Help: "主机数量"}
	return ssh
}
func (ssh *SSH) Start(m *ctx.Message, arg ...string) bool {
	m.Cap("stream", m.Source().Name)
	return false
}
func (ssh *SSH) Close(m *ctx.Message, arg ...string) bool {
	return false
}

var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches: map[string]*ctx.Cache{
		"nhost":  &ctx.Cache{Name: "主机数量", Value: "0", Help: "主机数量"},
		"domain": &ctx.Cache{Name: "domain", Value: "", Help: "主机域名"},
	},
	Configs: map[string]*ctx.Config{
		"hostname": &ctx.Config{Name: "hostname", Value: "com", Help: "主机数量"},

		"domain.json": &ctx.Config{Name: "domain.json", Value: "var/domain.json", Help: "主机数量"},
		"domain.png":  &ctx.Config{Name: "domain.png", Value: "var/domain.png", Help: "主机数量"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address [security [protocol]]", Help: "网络监听", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Sess("nfs").Call(func(sub *ctx.Message) *ctx.Message {
				sub.Start(fmt.Sprintf("host%d", m.Capi("nhost", 1)), "远程主机")
				sub.Spawn().Cmd("pwd", "")
				return sub
			}, m.Meta["detail"])
			if !m.Caps("domain") {
				m.Cap("domain", m.Cap("hostname", m.Conf("hostname")))
			}
			// m.Spawn(m.Target()).Cmd("save")
		}},
		"dial": &ctx.Command{Name: "dial address [security [protocol]]", Help: "网络连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Sess("nfs").CallBack(true, func(sub *ctx.Message) *ctx.Message {
				sub.Target().Start(sub)
				return sub
			}, m.Meta["detail"])
		}},
		"send": &ctx.Command{Name: "send [domain str] cmd arg...", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if ssh, ok := m.Target().Server.(*SSH); m.Assert(ok) {
				origin, domain := "", ""
				if len(arg) > 1 && arg[0] == "domain" {
					origin, arg = arg[1], arg[2:]
					if d := strings.TrimPrefix(origin, m.Cap("domain")); len(d) > 0 && d[0] == '.' {
						domain = d[1:]
					} else if d == "" {
						domain = d
					} else {
						domain = origin
					}

					if domain == "" { //本地执行
						msg := m.Spawn().Cmd(arg)
						m.Copy(msg, "result").Copy(msg, "append")
						return
					}
				} else {
					if m.Has("send_code") { //本地执行
						msg := m.Spawn().Cmd(arg)
						m.Copy(msg, "result").Copy(msg, "append")
					} else { //对端执行
						msg := m.Spawn(ssh.Message().Source())
						msg.Cmd("send", arg)
						m.Copy(msg, "result").Copy(msg, "append")
					}
					return
				}

				match := false
				host := strings.SplitN(domain, ".", 2)
				m.Travel(func(m *ctx.Message, i int) bool {
					if i == 0 {
						return true
					}
					if m.Cap("hostname") == host[0] || "*" == host[0] {
						ssh, ok := m.Target().Server.(*SSH)
						m.Assert(ok)
						msg := m.Spawn(ssh.Message().Source())

						if len(host) > 1 {
							msg.Cmd("send", "domain", host[1], arg)
						} else {
							msg.Cmd("send", arg)
						}
						m.Copy(msg, "result").Copy(msg, "append")

						if !match {
							match = !m.Appends("domain_miss")
						}
						return host[0] == "*"
					}
					return true
				}, c)

				if match {
					return
				}
				if m.Target() == c && m.Has("send_code") {
					m.Appends("domain_miss", true)
					return
				}
				if m.Cap("domain") == m.Conf("hostname") {
					m.Appends("domain_miss", true)
					return
				}

				// 向上路由
				msg := m.Spawn(c.Message().Source())
				msg.Cmd("send", "domain", origin, arg)
				m.Copy(msg, "result").Copy(msg, "append")
			}
		}},
		"pwd": &ctx.Command{Name: "pwd [hostname]", Help: "主机域名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if len(arg) == 0 {
				m.Echo(m.Cap("domain"))
				return
			}

			if m.Options("send_code") {
				if m.Target() == c {
					msg := m.Spawn().Cmd("send", "pwd", m.Confx("hostname", arg, 0))
					m.Cap("hostname", msg.Result(0))
					m.Cap("domain", msg.Result(1))
				} else {
					hostname := arg[0]
					m.Travel(func(m *ctx.Message, line int) bool {
						if hostname == m.Cap("hostname") {
							hostname += m.Cap("nhost")
							return false
						}
						return true
					}, c)
					m.Echo(m.Cap("hostname", hostname))
					m.Echo("%s.%s", m.Cap("domain"), m.Cap("hostname"))
				}
				return
			}

			if m.Target() == c {
				m.Conf("hostname", arg[0])
				msg := m.Spawn().Cmd("send", "pwd", arg[0])
				m.Cap("hostname", msg.Result(0))
				m.Cap("domain", msg.Result(1))
			} else {
				m.Spawn().Cmd("send", "pwd", arg[0])
			}
			m.Echo(m.Cap("domain"))
		}},
		"hello": &ctx.Command{Name: "hello request", Help: "加密请求", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			aaa := m.Target().Message().Sess("aaa", false)
			for _, k := range m.Meta["seal"] {
				for i, v := range m.Meta[k] {
					m.Meta[k][i] = m.Spawn(aaa).Cmd("deal", v).Result(0)
				}
			}
			for _, k := range m.Meta["encrypt"] {
				for i, v := range m.Meta[k] {
					m.Meta[k][i] = m.Spawn(aaa).Cmd("decrypt", v).Result(0)
				}
			}

			if len(arg) == 0 {
				if !m.Has("mi") {
					cert := aaa.Spawn().Cmd("certificate")
					m.Echo(cert.Result(0))
				} else {
					msg := m.Sess("aaa").Cmd("login", m.Option("mi"), m.Option("mi"))
					m.Echo(msg.Result(0))
					msg.Sess("aaa").Cmd("newcipher", m.Option("mi"))
				}
				return
			}

			msg := m.Spawn().Copy(m, "option").Cmd(arg)
			m.Copy(msg, "result").Copy(msg, "append")

		}},
		"shake": &ctx.Command{
			Name: "shake [domain host] cmd... [seal option...][encrypt option...]",
			Help: "加密通信",
			Form: map[string]int{"seal": -1, "encrypt": -1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if ssh, ok := m.Target().Server.(*SSH); m.Assert(ok) {
					if len(arg) == 0 {
						for k, v := range ssh.peer {
							m.Echo("%s: %s\n", k, v.Cap("stream"))
						}
						return
					}

					peer := "peer"
					args := []string{}
					if len(arg) > 1 && arg[0] == "domain" {
						args = append(args, "domain", arg[1])
						peer, arg = arg[1], arg[2:]
					}
					if ssh.peer == nil {
						ssh.peer = map[string]*ctx.Message{}
					}
					user, ok := ssh.peer[peer]
					if !ok {
						user = m.Sess("aaa").Cmd("login", "cert", m.Spawn().Cmd("send", args, "hello"), peer)
						ssh.peer[peer] = user
						mi := user.Cap("sessid")

						remote := m.Spawn().Add("option", mi, m.Spawn(user).Cmd("seal", mi)).Add("option", "seal", mi).Cmd("send", args, "hello")
						m.Spawn(user).Cmd("newcipher", mi)
						user.Cap("remote", "remote", remote.Result(0), "远程会话")
						user.Cap("remote_mi", "remote_mi", mi, "远程密钥")
					}

					msg := m.Spawn(ssh.Message().Source()).Copy(m, "option")
					msg.Option("hello", "world")
					msg.Option("world", "hello")
					for _, k := range msg.Meta["seal"] {
						for i, v := range msg.Meta[k] {
							msg.Meta[k][i] = msg.Spawn(user).Cmd("seal", v).Result(0)
						}
					}
					for _, k := range msg.Meta["encrypt"] {
						for i, v := range msg.Meta[k] {
							msg.Meta[k][i] = msg.Spawn(user).Cmd("encrypt", v).Result(0)
						}
					}
					msg.Detail("send", args, "hello", arg)
					ssh.Message().Back(msg)
					m.Copy(msg, "result").Copy(msg, "append")
				}
			}},
		"save": &ctx.Command{Name: "save", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			json := m.Sess("nfs")
			json.Put("option", "data", map[string]string{"domain": m.Cap("domain")})
			json.Cmd("json", m.Conf("domain.json"))
			m.Sess("nfs").Cmd("genqr", m.Conf("domain.png"), json.Result(0))

		}},
	},
}

func init() {
	ssh := &SSH{}
	ssh.Context = Index
	ctx.Index.Register(Index, ssh)
}
