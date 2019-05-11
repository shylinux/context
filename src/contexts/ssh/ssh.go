package ssh

import (
	"contexts/ctx"
	"strings"
	"toolkit"
)

type SSH struct {
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
	return ssh
}
func (ssh *SSH) Start(m *ctx.Message, arg ...string) bool {
	return true
}
func (ssh *SSH) Close(m *ctx.Message, arg ...string) bool {
	return false
}

var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches: map[string]*ctx.Cache{
		"nnode": &ctx.Cache{Name: "nnode", Value: "0", Help: "节点数量"},
	},
	Configs: map[string]*ctx.Config{
		"node":  &ctx.Config{Name: "node", Value: map[string]interface{}{}, Help: "节点信息"},
		"cert":  &ctx.Config{Name: "cert", Value: map[string]interface{}{}, Help: "用户信息"},
		"trust": &ctx.Config{Name: "trust", Value: map[string]interface{}{"fresh": false, "user": true, "up": true}, Help: "可信节点"},
		"timer": &ctx.Config{Name: "timer", Value: map[string]interface{}{"interval": "10s", "timer": ""}, Help: "断线重连"},
	},
	Commands: map[string]*ctx.Command{
		"init": &ctx.Command{Name: "init", Help: "启动", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if m.Confs("runtime", "boot.ctx_box") {
				m.Conf("runtime", "node.type", "worker")
				m.Conf("runtime", "node.name", m.Conf("runtime", "boot.pathname"))
			} else {
				m.Conf("runtime", "node.type", "server")
				m.Conf("runtime", "node.name", strings.Replace(strings.TrimSuffix(m.Conf("runtime", "boot.hostname"), ".local"), ".", "_", -1))
			}
			m.Conf("runtime", "node.route", m.Conf("runtime", "node.name"))
			m.Conf("runtime", "user.name", m.Conf("runtime", "boot.USER"))
			m.Conf("runtime", "work.name", m.Conf("runtime", "boot.USER"))
			return
		}},
		"node": &ctx.Command{Name: "node [create|delete [name [type module]]]", Help: "节点", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmdy("ctx.config", "node", "format", "table", "fields", "type", "module", "create_time")
				return
			}

			switch arg[0] {
			case "create": // 添加节点
				m.Log("info", "create node %s %s", arg[1], arg[2])
				m.Confv("node", arg[1], map[string]interface{}{
					"name": arg[1], "type": arg[2], "module": arg[3],
					"create_time": m.Time(),
				})

			case "delete": // 删除节点
				m.Log("info", "delete node %s %s", arg[1], kit.Formats(m.Conf("node", arg[1])))
				delete(m.Confm("node"), arg[1])
			}
			return
		}},
		"cert": &ctx.Command{Name: "cert [node|user|work]", Help: []string{"认证",
			"node [create|check text|trust node]",
			"user [create|check text|share role node...|proxy node|trust node]",
			"work [create node name|check node name]",
		}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Add("append", "key", "node.cert")
				m.Add("append", "route", m.Conf("runtime", "node.route"))
				m.Add("append", "value", m.Conf("runtime", "node.cert"))
				m.Add("append", "key", "user.cert")
				m.Add("append", "route", m.Conf("runtime", "user.route"))
				m.Add("append", "value", m.Conf("runtime", "user.cert"))
				m.Add("append", "key", "work.name")
				m.Add("append", "route", m.Conf("runtime", "work.route"))
				m.Add("append", "value", m.Conf("runtime", "work.name"))
				m.Table()
				return
			}

			switch arg[0] {
			case "node": // 节点认证
				if len(arg) == 1 {
					m.Echo(m.Conf("runtime", "node.cert"))
					break
				}

				switch arg[1] {
				case "create": // 创建证书
					msg := m.Cmd("aaa.rsa", "gen")
					m.Conf("runtime", "node.cert", msg.Append("certificate"))
					m.Conf("runtime", "node.key", msg.Append("private"))
					m.Echo(m.Conf("runtime", "node.cert"))

				case "check": // 数字验签
					if m.Option("node.cert", m.Cmd("aaa.auth", "nodes", m.Option("node.route"), "cert").Append("meta")); !m.Options("node.cert") {
						m.Option("node.cert", m.Cmdx("ssh.remote", m.Option("node.route"), "check", "node"))
						m.Cmd("aaa.auth", "nodes", m.Option("node.route"), "cert", m.Option("node.cert"))
					}

					if !m.Cmds("aaa.rsa", "verify", m.Option("node.cert"), m.Option("node.sign"), arg[2]) {
						m.Log("warn", "node error")
						m.Echo("false")
					} else {
						m.Echo("true")
					}

				case "trust": // 可信节点
					if m.Confs("trust", arg[2]) {
						m.Echo("true")

					} else if m.Confs("trust", "user") && m.Conf("runtime", "user.route") == arg[2] {
						m.Echo("true")

					} else if m.Confs("trust", "up") && strings.HasPrefix(m.Conf("runtime", "node.route"), arg[2]) {
						m.Echo("true")
					}
				}

			case "user": // 用户认证
				if len(arg) == 1 {
					m.Echo(m.Conf("runtime", "user.cert"))
					break
				}

				switch arg[1] {
				case "create": // 创建证书
					msg := m.Cmd("aaa.rsa", "gen")
					m.Conf("runtime", "user.route", m.Conf("runtime", "node.route"))
					m.Conf("runtime", "user.cert", msg.Append("certificate"))
					m.Conf("runtime", "user.key", msg.Append("private"))
					m.Echo(m.Conf("runtime", "user.cert"))

				case "check": // 数字验签
					if m.Option("user.cert", m.Cmd("aaa.auth", "username", m.Option("username"), "cert").Append("meta")); !m.Options("user.cert") {
						m.Option("user.cert", m.Cmd("ssh.remote", m.Option("user.route"), "check", "user").Append("user.cert"))
						m.Cmd("aaa.auth", "username", m.Option("username"), "cert", m.Option("user.cert"))
						m.Cmd("aaa.auth", "username", m.Option("username"), "userrole", "void")
					}

					if !m.Options("user.cert") || !m.Cmds("aaa.rsa", "verify", m.Option("user.cert"), m.Option("user.sign"), arg[2]) {
						m.Log("warn", "user error")
						m.Echo("false")
					} else {
						m.Echo("true")
					}

				case "share": // 共享用户
					for _, route := range arg[3:] {
						user := m.Cmd("ssh.remote", route, "check", "user")
						if m.Cmd("aaa.role", arg[2], "user", user.Append("user.name"), "cert", user.Append("user.cert")); arg[2] == "root" && route != m.Conf("runtime", "node.route") {
							m.Conf("runtime", "user.route", user.Append("user.route"))
							m.Conf("runtime", "user.name", user.Append("user.name"))
							m.Conf("runtime", "user.cert", user.Append("user.cert"))
							m.Conf("runtime", "user.key", "")
						}
					}

				case "proxy": // 代理用户
					if len(arg) == 2 {
						m.Cmdy("aaa.auth", "proxy")
						break
					}
					if !m.Cmds("aaa.auth", "proxy", arg[2], "session") {
						m.Cmdy("aaa.sess", "proxy", "proxy", arg[2])
					}

				case "trust": // 可信代理
					hash := kit.Hashs("rand", m.Option("text.time", m.Time("stamp")), arg[2])
					m.Option("user.sign", m.Cmdx("ssh.remote", m.Option("user.route"), "check", "user", arg[2], hash))
					m.Echo("%s", m.Options("user.sign") && m.Cmds("ssh.check", hash))
				}

			case "work": // 公有认证
				if len(arg) == 1 {
					m.Echo(m.Conf("runtime", "work.name"))
					break
				}

				switch arg[1] {
				case "create": // 创建证书
					if name := m.Cmdx("ssh.remote", arg[2], "check", "work", arg[3], m.Conf("runtime", "user.route"), m.Conf("runtime", "user.cert")); name != "" {
						m.Conf("runtime", "work.route", arg[2])
						m.Conf("runtime", "work.name", name)
						m.Echo(name)
					}

				case "check": // 数字验签
					if m.Option("user.route") != m.Cmdx("ssh.remote", arg[2], "check", "work", arg[3]) {
						m.Log("warn", "work error")
						m.Echo("false")
					} else {
						m.Echo("true")
					}
				}
			}
			return
		}},
		"check": &ctx.Command{Name: "check node|user|work", Help: []string{"验签",
			"node", "user [node text]", "work name [node cert]",
		}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch arg[0] {
			case "node": // 节点验签
				m.Echo(m.Conf("runtime", "node.cert"))

			case "user": // 用户验签
				if len(arg) == 1 {
					m.Append("user.cert", m.Conf("runtime", "user.cert"))
					m.Append("user.name", m.Conf("runtime", "user.name"))
					m.Append("user.route", kit.Select(m.Conf("runtime", "node.route"), m.Conf("runtime", "user.route")))
				} else { // 代理签证
					if arg[1] == m.Conf("runtime", "node.route") || m.Cmds("aaa.auth", "proxy", arg[1], "session") {
						m.Echo(m.Cmdx("aaa.rsa", "sign", m.Conf("runtime", "user.key"), arg[2]))
					}
				}
			case "work": // 工作验签
				if cert := m.Confm("cert", arg[1]); len(arg) == 2 {
					if cert != nil {
						m.Echo("%s", cert["user"])
					}
				} else { // 工作签证
					if cert == nil {
						m.Conf("cert", arg[1], map[string]interface{}{"create_time": m.Time(), "user": arg[2]})
					} else if cert["user"] != arg[2] {
						return // 签证失败
					}
					m.Echo(arg[1])
				}
			}
			return
		}},
		"action": &ctx.Command{Name: "action", Help: "命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			hash := kit.Hashs(
				m.Option("text.rand"),
				m.Option("text.time"),
				m.Option("text.cmd"),
				m.Option("text.route"),
				m.Option("node.route"),
				m.Option("user.route"),
				m.Option("user.name"),
				m.Option("work.name"),
				m.Option("work.route"),
			)

			// 文本验签
			if m.Option("text.hash") != hash {
				m.Echo("text error %s != %s", m.Option("text.hash"), hash)
				m.Log("warn", "text error")
				return
			}

			// 设备验签
			if !m.Cmds("ssh.cert", "node", "check", hash) {
				m.Echo("node error of %s", m.Option("node.route"))
				return
			}

			// 用户验签
			m.Option("username", m.Option("user.name"))
			if !m.Confs("runtime", "user.route") && m.Confs("trust", "fresh") {
				m.Log("warn", "trust fresh %s", m.Option("node.route"))
				m.Option("trust", "fresh")

			} else if m.Cmds("ssh.cert", "node", "trust", m.Option("node.route")) {
				m.Log("warn", "trust node %s", m.Option("node.route"))
				m.Option("trust", "node")

			} else if m.Options("user.route") && m.Cmds("ssh.cert", "node", "trust", m.Option("user.route")) && m.Cmds("ssh.cert", "user", "trust", m.Option("node.route")) {
				m.Log("warn", "trust user %s", m.Option("user.route"))
				m.Option("trust", "user")

			} else if m.Option("username", m.Option("work.name")); m.Options("work.route") && m.Cmds("ssh.cert", "node", "trust", m.Option("work.route")) && m.Cmds("ssh.cert", "work", "check", m.Option("work.route"), m.Option("username")) {
				m.Log("warn", "trust work %s", m.Option("work.route"))
				m.Option("trust", "work")

			} else if m.Option("userrole", "void"); m.Confs("trust", "none") {
				m.Log("warn", "trust none")
				m.Option("trust", "none")

			} else {
				m.Log("warn", "user error of %s", m.Option("user.route"))
				m.Echo("user error")
				return
			}
			m.Log("info", "username: %s", m.Option("username"))

			// 创建会话
			m.Option("sessid", m.Cmdx("aaa.user", "session", "select"))

			// 创建空间
			m.Option("bench", m.Cmdx("aaa.sess", "bench", "select"))

			// 权限检查
			if !m.Cmds("aaa.work", "right", "remote", arg[0]) {
				m.Echo("no right %s %s", "remote", arg[0])
				return
			}

			// 执行命令
			m.Cmdm(arg)
			return
		}},
		"remote": &ctx.Command{Name: "remote auto|dial|listen args...", Help: "连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 设备证书
			if !m.Confs("runtime", "node.cert") || !m.Confs("runtime", "node.key") {
				m.Cmd("ssh.cert", "node", "create")
			}

			switch arg[0] {
			case "auto": // 自动连接
				if m.Cmd("ssh.remote", "dial", "consul", "/shadow"); !m.Confs("runtime", "boot.ctx_box") {
					m.Cmd("ssh.remote", "listen", m.Conf("runtime", "boot.ssh_port"))
					m.Cmd("web.serve", "usr", m.Conf("runtime", "boot.web_port"))
				}

			case "listen": // 监听连接
				m.Call(func(nfs *ctx.Message) *ctx.Message {
					if nfs.Has("node.port") {
						m.Log("info", "node.port %v", nfs.Optionv("node.port"))
						m.Conf("runtime", "node.port", nfs.Optionv("node.port"))
					}
					return nil
				}, "nfs.remote", arg)

			case "redial": // 断线重连
				if !m.Caps("stream") {
					m.Cmdx("remote", "dial", arg[1:])
				}

			case "dial": // 连接主机
				m.Call(func(nfs *ctx.Message) *ctx.Message {
					// 删除重连
					if m.Confs("timer", "timer") {
						m.Conf("timer", "timer", m.Cmdx("cli.timer", "delete", m.Conf("timer", "timer")))
					}

					// 注册设备
					m.Spawn(nfs.Target()).Call(func(node *ctx.Message) *ctx.Message {
						// 添加网关
						m.Cmd("ssh.node", "create", node.Append("node.name"), "master", m.Cap("stream", nfs.Format("target")))

						// 清理回调
						nfs.Free(func(nfs *ctx.Message) bool {
							m.Cmd("aaa.auth", m.Cmdx("aaa.auth", "nodes", node.Append("node.name")), "delete", "node")
							m.Cmd("ssh.node", "delete", node.Append("node.name"))
							m.Cap("stream", "")

							// 断线重连
							m.Conf("timer", "timer", m.Cmdx("cli.timer", "repeat", m.Conf("timer", "interval"), "context", "ssh", "remote", "redial", arg[1:]))
							return true
						})

						// 本机路由
						m.Conf("runtime", "node.route", node.Append("node.route")+"."+node.Result(0))

						// 用户路由
						if m.Confs("runtime", "user.cert") && m.Confs("runtime", "user.key") {
							m.Conf("runtime", "user.route", m.Conf("runtime", "node.route"))

						} else if node.Appends("user.route") && !m.Confs("runtime", "user.route") {
							m.Cmd("ssh.node", "share", "root", node.Append("user.route"))
						}
						return nil
					}, "send", "add", m.Conf("runtime", "node.name"), m.Conf("runtime", "node.type"), m.Conf("runtime", "node.cert"))
					return nil
				}, "nfs.remote", arg)

			case "add":
				// 命名节点
				name := arg[1]
				for node := m.Confm("node", name); node != nil; node = m.Confm("node", name) {
					name = kit.Format("%s_%s", arg[1], m.Capi("nnode", 1))
				}

				// 添加节点
				m.Cmd("ssh.node", "create", name, arg[2], m.Format("source"))

				// 清理回调
				m.Sess("ms_source", false).Free(func(msg *ctx.Message) bool {
					m.Cmd("aaa.auth", m.Cmdx("aaa.auth", "nodes", name), "delete", "node")
					m.Cmd("ssh.node", "delete", name)
					return true
				})

				// 同步信息
				m.Append("node.name", m.Conf("runtime", "node.name"))
				m.Append("node.route", m.Conf("runtime", "node.route"))
				m.Append("user.route", m.Conf("runtime", "user.route"))
				m.Append("user.name", m.Conf("runtime", "user.name"))
				m.Echo(name).Back(m)

			default:
				// 同步异步
				sync := true
				switch arg[0] {
				case "async", "sync":
					sync, arg = arg[0] == "sync", arg[1:]
				}

				// 局域路由
				if arg[0] == m.Conf("runtime", "node.name") || arg[0] == m.Conf("runtime", "node.route") {
					arg[0] = ""
				}
				arg[0] = strings.TrimPrefix(arg[0], m.Conf("runtime", "node.route")+".")
				arg[0] = strings.TrimPrefix(arg[0], m.Conf("runtime", "node.name")+".")

				// 拆分路由
				route, names, arg := arg[0], strings.SplitN(arg[0], ".", 2), arg[1:]
				if len(names) > 1 && names[0] == "" && names[1] != "" {
					names[0], names[1] = names[1], names[0]
				}

				if rest := kit.Select("", names, 1); names[0] != "" {
					// 数字签名
					if !m.Options("remote_code") && arg[0] != "check" {
						hash, meta := kit.Hash("rand",
							m.Option("text.time", m.Time("stamp")),
							m.Option("text.cmd", strings.Join(arg, " ")),
							m.Option("text.route", route),
							m.Option("node.route", m.Conf("runtime", "node.route")),
							m.Option("user.route", kit.Select(m.Conf("runtime", "node.route"), m.Conf("runtime", "user.route"))),
							m.Option("user.name", m.Option("username")),
							m.Option("work.name", m.Conf("runtime", "work.name")),
							m.Option("work.route", m.Conf("runtime", "work.route")),
						)
						m.Option("text.rand", meta[0])
						m.Option("node.sign", m.Cmdx("aaa.rsa", "sign", m.Conf("runtime", "node.key"), m.Option("text.hash", hash)))
					}

					// 查找路由
					ps := []string{}
					if names[0] == "%" || names[0] == "*" { // 广播命令
						m.Confm("node", names[0], func(name string, node map[string]interface{}) {
							if kit.Format(node["type"]) != "master" {
								ps = append(ps, kit.Format(node["module"]))
							}
						})

					} else if m.Confm("node", names[0], func(node map[string]interface{}) { // 单播命令
						ps = append(ps, kit.Format(node["module"]))

					}) == nil && m.Caps("stream") { // 上报命令
						rest = strings.Join(names, ".")
						ps = append(ps, m.Cap("stream"))
					}
					if len(ps) == 0 {
						m.Echo("error: not found %s", names[0]).Back(m)
						return
					}

					// 路由转发
					for _, p := range ps {
						m.Find(p, true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
							return m.CopyFuck(sub, "append").CopyFuck(sub, "result").Echo("\n")
						}, "send", rest, arg)
					}
					return
				}

				// 远程回调
				defer func() { m.Back(m) }()

				// 执行命令
				if arg[0] == "check" { // 数字验签
					m.Cmd(arg)

				} else if m.Options("remote_code") { // 远程调用
					m.Cmd("action", arg)

				} else { // 本地调用
					m.Cmdm(arg)
				}
			}
			return
		}},
	},
}

func init() {
	ssh := &SSH{}
	ssh.Context = Index
	ctx.Index.Register(Index, ssh)
}
