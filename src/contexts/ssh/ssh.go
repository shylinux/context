package ssh

import (
	"contexts"
	"fmt"
	"strings"
	"time"
)

type SSH struct {
	nfs  *ctx.Context
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
	if ssh.Context == Index {
		Pulse = m
	}
	return ssh
}
func (ssh *SSH) Start(m *ctx.Message, arg ...string) bool {
	ssh.nfs = m.Source()
	m.Cap("stream", m.Source().Name)
	return false
}
func (ssh *SSH) Close(m *ctx.Message, arg ...string) bool {
	return false
	switch ssh.Context {
	case m.Target():
	case m.Source():
	}
	if m.Target() == Index {
		go func() {
			m.Target().Begin(m)
			m.Sess("nfs", "nfs")
			for !m.Caps("stream") {
				time.Sleep(time.Second * time.Duration(m.Confi("interval")))
				go ssh.Message().Spawn(m.Target()).Copy(ssh.Message(), "detail").Cmd()
				time.Sleep(time.Second * time.Duration(m.Confi("interval")))
			}
		}()
		return false
	}
	return true
}
func Done(m *ctx.Message, lock chan bool) {
	m.Log("lock", "done before %v", m.Meta["detail"])
	if m.Options("stdio") {
		lock <- true
	}
	m.Log("lock", "done after %v", m.Meta["detail"])
}
func Wait(m *ctx.Message, lock chan bool) {
	m.Log("lock", "wait before %v", m.Meta["detail"])
	if m.Options("stdio") {
		<-lock
	}
	m.Log("lock", "wait after %v", m.Meta["detail"])
}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches: map[string]*ctx.Cache{
		"nhost":  &ctx.Cache{Name: "主机数量", Value: "0", Help: "主机数量"},
		"domain": &ctx.Cache{Name: "domain", Value: "", Help: "主机域名"},

		"route": &ctx.Cache{Name: "route", Value: "com", Help: "主机数量"},
		"count": &ctx.Cache{Name: "count", Value: "3", Help: "主机数量"},
		"share": &ctx.Cache{Name: "share", Value: "root", Help: "主机数量"},
		"level": &ctx.Cache{Name: "level", Value: "root", Help: "主机数量"},
	},
	Configs: map[string]*ctx.Config{
		"hostname": &ctx.Config{Name: "hostname", Value: "com", Help: "主机数量"},

		"interval":    &ctx.Config{Name: "interval", Value: "3", Help: "主机数量"},
		"domain.json": &ctx.Config{Name: "domain.json", Value: "var/domain.json", Help: "主机数量"},
		"domain.png":  &ctx.Config{Name: "domain.png", Value: "var/domain.png", Help: "主机数量"},

		"mdb": &ctx.Config{Name: "mdb", Value: "mdb.chat", Help: "主机数量"},
		"uid": &ctx.Config{Name: "uid", Value: "", Help: "主机数量"},

		"type": &ctx.Config{Name: "type", Value: "terminal", Help: "主机数量"},
		"kind": &ctx.Config{Name: "kind", Value: "terminal", Help: "主机数量"},
		"name": &ctx.Config{Name: "name", Value: "vps", Help: "主机数量"},
		"mark": &ctx.Config{Name: "mark", Value: "com", Help: "主机数量"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address [security [protocol]]", Help: "网络监听", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*SSH); m.Assert(ok) {
				m.Sess("nfs").Call(func(sub *ctx.Message) *ctx.Message {
					sub.Start(fmt.Sprintf("host%d", Pulse.Capi("nhost", 1)), "远程主机")
					// sub.Spawn().Cmd("pwd", "init")
					return sub
				}, m.Meta["detail"])
				if !m.Caps("domain") {
					m.Cap("domain", m.Cap("hostname", m.Conf("hostname")))
				}
				// m.Spawn(m.Target()).Cmd("save")
			}
		}},
		"dial": &ctx.Command{Name: "dial address [security [protocol]]", Help: "网络连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*SSH); m.Assert(ok) {
				m.Sess("nfs").CallBack(true, func(sub *ctx.Message) *ctx.Message {
					sub.Target().Start(sub)
					return sub
				}, m.Meta["detail"])
			}
		}},
		"send": &ctx.Command{Name: "send [domain str] cmd arg...", Help: "远程执行",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if ssh, ok := m.Target().Server.(*SSH); m.Assert(ok) {
					domain := ""
					if len(arg) > 1 && arg[0] == "domain" {
						domain, arg = arg[1], arg[2:]
						domain = strings.TrimPrefix(strings.TrimPrefix(domain, m.Cap("domain")), ".")
					}

					if m.Has("send_code") {
						msg := m.Spawn().Cmd(arg)
						m.Copy(msg, "result").Copy(msg, "append")
					} else {
						msg := m.Spawn(ssh.Message().Source())
						msg.Cmd("send", arg)
						m.Copy(msg, "result").Copy(msg, "append")
						return
					}

					return

					if domain != "" {
						domain_miss := true
						host := strings.SplitN(domain, ".", 2)
						m.Travel(func(m *ctx.Message, i int) bool {
							if i > 0 {
								if m.Cap("hostname") == host[0] {
									ssh, ok := m.Target().Server.(*SSH)
									m.Assert(ok)

									msg := m.Spawn(ssh.Message().Source()).Copy(m, "option")
									if len(host) > 1 {
										msg.Options("downflow", true)
										msg.Detail("send", "domain", host[1], arg)
									} else {
										msg.Detail(arg)
									}

									if ssh.Message().Back(msg); !msg.Appends("domain_miss") {
										m.Copy(msg, "result").Copy(msg, "append")
										domain_miss = false
										return false
									}
								}
								return false
							}
							return true
						}, c)

						if domain_miss && !m.Options("downflow") && m.Cap("domain") != m.Conf("domain") {
							ssh, ok := c.Server.(*SSH)
							m.Assert(ok)

							msg := m.Spawn(ssh.Message().Source()).Copy(m, "option")
							msg.Detail("send", "domain", domain, arg)

							if ssh.Message().Back(msg); !msg.Appends("domain_miss") {
								m.Copy(msg, "result").Copy(msg, "append")
								domain_miss = false
							}
						}

						m.Appends("domain_miss", domain_miss)
						return
					}

					if m.Options("send_code") || m.Cap("status") != "start" {
						msg := m.Spawn().Cmd(arg)
						m.Copy(msg, "result").Copy(msg, "append")
					} else {
						msg := m.Spawn(ssh.Message().Source())
						msg.Copy(m, "option").Detail(arg)
						ssh.Message().Back(msg)
						m.Copy(msg, "result").Copy(msg, "append")
					}
					return
				}
			}},
		"pwd": &ctx.Command{Name: "pwd", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if len(arg) > 0 {
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
					return
				}
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
		"close": &ctx.Command{Name: "close", Help: "连接断开", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Target().Close(m)
		}},
		"list": &ctx.Command{Name: "list", Help: "连接断开", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			domain := m.Cap("domain")
			m.Travel(func(m *ctx.Message, i int) bool {
				if m.Confs("domains") {
					m.Echo("%s: %s.%s\n", m.Target().Name, domain, m.Conf("domains"))
				}
				return true
			}, c)
		}},
		"save": &ctx.Command{Name: "save", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			json := m.Sess("nfs")
			json.Put("option", "data", map[string]string{"domain": m.Cap("domain")})
			json.Cmd("json", m.Conf("domain.json"))
			m.Sess("nfs").Cmd("genqr", m.Conf("domain.png"), json.Result(0))

		}},
		"who": &ctx.Command{Name: "who", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			aaa := m.Sess("aaa")
			if aaa != nil {
				m.Echo(aaa.Cap("group"))
			}

		}},
		"good": &ctx.Command{Name: "good context|command|config|cache args", Help: "设备注册", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Append("share", m.Cap("share"))
				m.Append("level", m.Cap("level"))
				m.Append("type", m.Conf("type"))
				m.Append("value", m.Cap("domain"))
				m.Append("kind", m.Conf("kind"))
				m.Append("name", m.Cap("domain"))
				m.Append("mark", m.Conf("mark"))
				m.Append("count", m.Cap("count"))
				m.Back(m)
				return
			}
			cmds := m.Option("cmds")

			if arg[0] == "context" {
				if len(arg) > 1 {
					cmds = arg[1]
				}

				m.Travel(func(msg *ctx.Message, i int) bool {
					current := msg.Target()
					if _, ok := current.Index[cmds]; ok {

					} else if cmds != "" && cmds != "root" {
						return true
					}

					m.Add("append", "name", current.Name)
					m.Add("append", "help", current.Help)
					return true
				}, ctx.Index)
				return
			}

			if len(arg) > 2 {
				cmds = arg[2]
			}
			current := m.Sess(arg[1], arg[1], "search").Target()
			if x, ok := current.Index[cmds]; ok {
				current = x
			} else if cmds != "" && cmds != "root" {
				return
			}

			switch arg[0] {
			case "command":

				for k, x := range current.Commands {
					m.Add("append", "key", k)
					m.Add("append", "name", x.Name)
					m.Add("append", "help", x.Help)
				}
			case "config":
				for k, x := range current.Configs {
					m.Add("append", "key", k)
					m.Add("append", "name", x.Name)
					m.Add("append", "value", x.Value)
					m.Add("append", "help", x.Help)
				}
			case "cache":
				for k, x := range current.Caches {
					m.Add("append", "key", k)
					m.Add("append", "name", x.Name)
					m.Add("append", "value", x.Value)
					m.Add("append", "help", x.Help)
				}
			}
		}},
	},
}

func init() {
	ssh := &SSH{}
	ssh.Context = Index
	ctx.Index.Register(Index, ssh)
}
