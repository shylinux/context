package ssh

import (
	"contexts/ctx"
	"encoding/base64"
	"fmt"
	"os"
	"path"
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
	return true
}

var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches: map[string]*ctx.Cache{
		"nnode": &ctx.Cache{Name: "nnode", Value: "0", Help: "节点数量"},
	},
	Configs: map[string]*ctx.Config{
		"node":           &ctx.Config{Name: "node", Value: map[string]interface{}{}, Help: "主机信息"},
		"current":        &ctx.Config{Name: "current", Value: "", Help: "当前主机"},
		"timer":          &ctx.Config{Name: "timer", Value: "", Help: "断线重连"},
		"timer_interval": &ctx.Config{Name: "timer_interval", Value: "10s", Help: "断线重连"},
	},
	Commands: map[string]*ctx.Command{
		"init": &ctx.Command{Name: "init", Help: "启动", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if m.Confs("runtime", "ctx_box") {
				m.Conf("runtime", "node.type", "worker")
				m.Conf("runtime", "node.name", m.Conf("runtime", "pathname"))
			} else {
				m.Conf("runtime", "node.type", "server")
				m.Conf("runtime", "node.name", strings.Replace(strings.TrimSuffix(m.Conf("runtime", "hostname"), ".local"), ".", "_", -1))
			}
			m.Conf("runtime", "node.route", m.Conf("runtime", "node.name"))
			m.Conf("runtime", "user.name", m.Conf("runtime", "USER"))
			return
		}},
		"remote": &ctx.Command{Name: "remote auto|dial|listen args...", Help: "远程连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmdy("ctx.config", "node")
				return
			}

			// 设备证书
			if !m.Confs("runtime", "node.cert") || !m.Confs("runtime", "node.key") {
				msg := m.Cmd("aaa.rsa", "gen", "common", m.Confv("runtime", "node"))
				m.Conf("runtime", "node.cert", msg.Append("certificate"))
				m.Conf("runtime", "node.key", msg.Append("private"))
			}

			switch arg[0] {
			case "auto": // 自动连接
				if m.Cmd("ssh.remote", "dial", "consul", "/shadow"); !m.Confs("runtime", "ctx_box") {
					m.Cmd("ssh.remote", "listen", m.Conf("runtime", "ssh_port"))
					m.Cmd("web.serve", "usr", m.Conf("runtime", "web_port"))
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
					// 断线重连
					if m.Confs("timer") {
						m.Conf("timer", m.Cmdx("cli.timer", "delete", m.Conf("timer")))
					}

					msg := m.Spawn(nfs.Target())
					if m.Confs("runtime", "user.cert") {
						msg.Option("user.route", kit.Select(m.Conf("runtime", "user.route"), m.Conf("runtime", "user.route")))
					}
					msg.Call(func(node *ctx.Message) *ctx.Message {
						// 添加主机
						m.Confv("node", node.Append("node.name"), map[string]interface{}{
							"module":      m.Cap("stream", nfs.Format("target")),
							"create_time": m.Time(),
							"node": map[string]interface{}{
								"name": node.Append("node.name"),
								"type": "master",
							},
						})

						m.Log("fuck", "what %v", node.Meta)
						// 主机路由
						m.Conf("runtime", "node.route", node.Append("node.route")+"."+node.Result(0))
						if !m.Confs("runtime", "user.route") {
							if m.Confs("runtime", "user.cert") && m.Confs("runtime", "user.key") {
								m.Cmd("ssh.share", "root", m.Conf("runtime", "node.route"))
								m.Log("fuck", "what %v", 123)
								m.Log("fuck", "what %v", node.Meta)
								if !node.Appends("user.route") {
									m.Log("fuck", "what %v", 123)
									m.Cmd("ssh.share", node.Append("node.route"), "root", m.Conf("runtime", "node.route"))
								}

							} else if node.Appends("user.route") {
								m.Cmd("ssh.share", "root", node.Append("user.route"))
							}
						}

						// 默认节点
						if !m.Confs("current") {
							m.Conf("current", node.Append("node.name"))
						}

						// 清理主机
						nfs.Free(func(nfs *ctx.Message) bool {
							m.Conf("timer", m.Cmdx("cli.timer", "repeat", m.Conf("timer_interval"), "context", "ssh", "remote", "redial", arg[1:]))

							m.Log("info", "delete node %s", node.Append("node.name"))
							delete(m.Confm("node"), node.Append("node.name"))
							m.Cap("stream", "")
							return true
						})
						return nil
					}, "send", "recv", "add", m.Conf("runtime", "node.name"), m.Conf("runtime", "node.type"), m.Conf("runtime", "node.cert"))
					return nil
				}, "nfs.remote", arg)

			case "recv":
				switch arg[1] {
				case "add":
					// 节点命名
					name := arg[2]
					for node := m.Confm("node", name); node != nil; node = m.Confm("node", name) {
						name = fmt.Sprintf("%s_%d", arg[2], m.Capi("nnode", 1))
					}

					info := map[string]string{}
					if len(arg) > 4 && m.Cmds("aaa.rsa", "check", m.Conf("runtime", "node.cert"), arg[4]) {
						m.Cmd("aaa.rsa", "info", arg[4]).Table(func(line map[string]string) {
							for k, v := range line {
								info[k] = v
							}
						})
					}
					m.Log("info", "info--- %v", info)

					// 添加节点
					m.Confv("node", name, map[string]interface{}{
						"module":      m.Format("source"),
						"create_time": m.Time(),
						"node": map[string]interface{}{
							"name": name,
							"type": arg[3],
						},
					})

					// 节点路由
					m.Append("user.name", m.Conf("runtime", "user.name"))
					m.Append("user.route", m.Conf("runtime", "user.route"))
					m.Append("node.route", m.Conf("runtime", "node.route"))
					m.Append("node.name", m.Conf("runtime", "node.name"))
					m.Echo(name).Back(m)

					// 默认节点
					if !m.Confs("current") {
						m.Conf("current", name)
					}

					// 清理节点
					m.Sess("ms_source", false).Free(func(msg *ctx.Message) bool {
						m.Log("info", "delete node %s", name)
						delete(m.Confm("node"), name)
						return true
					})
				}

			default:
				// 拆分路由
				if arg[0] == m.Conf("runtime", "node.name") || arg[0] == m.Conf("runtime", "node.route") {
					arg[0] = ""
				}
				arg[0] = strings.TrimPrefix(arg[0], m.Conf("runtime", "node.route")+".")
				route, names, arg := arg[0], strings.SplitN(arg[0], ".", 2), arg[1:]
				if len(names) > 1 && names[0] == "" && names[1] != "" {
					names[0], names[1] = names[1], names[0]
				}

				// 同步异步
				sync := !m.Options("remote_code")
				switch arg[0] {
				case "async", "sync":
					sync, arg = arg[0] == "sync", arg[1:]
				}

				// 路由转发
				if rest := kit.Select("", names, 1); names[0] != "" {
					// 数字签名
					if !m.Options("remote_code") {
						// 用户路由
						m.Option("user.route", kit.Select(m.Conf("runtime", "node.route"), m.Conf("runtime", "user.route")))
						m.Cmd("aaa.auth", "username", m.Option("username"), "session", "login").Table(func(line map[string]string) {
							m.Option("user.route", m.Cmd("aaa.auth", line["key"], "login").Append("meta"))
						})

						// 数据哈希
						hash, meta := kit.Hash("rand",
							m.Option("text.time", m.Time("stamp")),
							m.Option("text.cmd", strings.Join(arg, " ")),
							m.Option("text.route", route),
							m.Option("node.route", m.Conf("runtime", "node.route")),
							m.Option("user.route"),
							m.Option("user.name", m.Option("username")),
						)
						m.Option("text.rand", meta[0])

						// 设备签名
						m.Option("node.sign", m.Cmdx("aaa.rsa", "sign", m.Conf("runtime", "node.key"), m.Option("text.hash", hash)))
						// 用户签名
						if m.Options("user.sign") && m.Confs("runtime", "user.key") {
							m.Option("user.sign", m.Cmdx("aaa.rsa", "sign", m.Conf("runtime", "user.key"), m.Option("text.hash", hash)))
						}
					}

					if names[0] == "*" { // 广播命令
						m.Confm("node", func(name string, node map[string]interface{}) {
							m.Find(kit.Format(node["module"]), true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
								return m.Copy(sub, "append").Copy(sub, "result")
							}, "send", "", arg)
						})

					} else if m.Confm("node", names[0], func(node map[string]interface{}) { // 单播命令
						m.Find(kit.Format(node["module"]), true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
							return m.Copy(sub, "append").Copy(sub, "result")
						}, "send", rest, arg)

					}) == nil { // 回溯命令
						m.Find(m.Cap("stream"), true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
							return m.Copy(sub, "append").Copy(sub, "result")
						}, "send", strings.Join(names, "."), arg)
					}
					return
				}

				// 返回结果
				defer func() { m.Back(m) }()

				// 查看证书
				switch arg[0] {
				case "check":
					switch arg[1] {
					case "node": // 设备证书
						m.Echo(m.Conf("runtime", "node.cert"))
					case "user":
						if len(arg) == 2 { // 用户证书
							m.Append("user.cert", m.Conf("runtime", "user.cert"))
							m.Append("user.name", m.Conf("runtime", "user.name"))
							m.Append("user.route", kit.Select(m.Conf("runtime", "node.route"), m.Conf("runtime", "user.route")))
						} else { // 代理验证
							if arg[2] == m.Conf("runtime", "node.route") || m.Cmds("aaa.auth", "proxy", arg[2], "session") {
								m.Echo(m.Cmdx("aaa.rsa", "sign", m.Conf("runtime", "user.key"), arg[3]))
							}
						}
					}
					return
				}

				if m.Options("remote_code") {
					// 检查数据
					hash, _ := kit.Hash(
						m.Option("text.rand"),
						m.Option("text.time"),
						m.Option("text.cmd"),
						m.Option("text.route"),
						m.Option("node.route"),
						m.Option("user.route"),
						m.Option("user.name"),
					)
					if m.Option("text.hash") != hash {
						m.Log("warning", "text error")
						return
					}

					// 设备证书
					m.Option("node.cert", m.Cmd("aaa.auth", "nodes", m.Option("node.route"), "cert").Append("meta"))
					if !m.Options("node.cert") {
						m.Option("node.cert", m.Spawn().Cmdx("ssh.remote", m.Option("node.route"), "sync", "check", "node"))
						m.Cmd("aaa.auth", "nodes", m.Option("node.route"), "cert", m.Option("node.cert"))
					}

					// 设备验签
					if !m.Cmds("aaa.rsa", "verify", m.Option("node.cert"), m.Option("node.sign"), m.Option("text.hash", hash)) {
						m.Log("warning", "node error")
						return
					}
				} else {
					m.Option("user.name", m.Conf("runtime", "user.name"))
				}

				switch arg[0] {
				case "login": // 用户代理
					if !m.Cmds("aaa.auth", "proxy", m.Option("node.route")) {
						return
					}

					sess := m.Cmd("aaa.auth", "username", m.Option("user.name"), "session", "proxy").Append("key")
					if sess == "" {
						sess = m.Cmdx("aaa.sess", "proxy", "username", m.Option("user.name"))
					}

					m.Cmd("aaa.auth", sess, "proxy", m.Option("node.route"))
					m.Echo(sess)
					return

				case "share": // 设备权限
					// 默认用户
					if !m.Confs("runtime", "user.route") {
						user := m.Spawn().Cmd("ssh.remote", m.Option("user.route"), "sync", "check", "user")
						m.Conf("runtime", "user.route", user.Append("user.route"))
						m.Conf("runtime", "user.name", user.Append("user.name"))
						m.Conf("runtime", "user.cert", user.Append("user.cert"))
						m.Cmd("aaa.auth", "username", user.Append("user.name"), "cert", user.Append("user.cert"))
						m.Cmd("aaa.user", "root", user.Append("user.name"), "what")
						return
					}

					// 共享用户
					if !m.Options("remote_code") || (m.Options("user.sign") && m.Conf("runtime", "user.name") == m.Option("user.name")) {
						if !m.Options("remote_code") || m.Cmds("aaa.rsa", "verify", m.Conf("runtime", "user.cert"), m.Option("user.sign"), m.Option("text.hash")) {
							for _, v := range arg[2:] {
								user := m.Spawn().Cmd("ssh.remote", v, "sync", "check", "user")
								m.Cmd("aaa.auth", "username", user.Append("user.name"), "cert", user.Append("user.cert"))
								m.Cmd("aaa.user", arg[1], user.Append("user.name"), "what")
							}
							return
						}
					}

					// 申请权限
					m.Spawn().Set("option", "remote_code", "").Cmds("ssh.remote", m.Conf("runtime", "user.route"), "sync", "apply", arg[1:])
					return

				case "apply": // 权限申请
					for _, v := range arg[2:] {
						user := m.Spawn().Cmd("ssh.remote", v, "sync", "check", "user")
						m.Cmd("aaa.auth", "username", user.Append("user.name"), "cert", user.Append("user.cert"))

						sess := m.Cmd("aaa.auth", "username", user.Append("user.name"), "session", "apply").Append("key")
						if sess == "" {
							sess = m.Cmdx("aaa.sess", "apply", "username", arg[2])
						}
						m.Cmd("aaa.auth", sess, "apply", m.Option("node.route"))
						m.Cmd("aaa.auth", sess, "share", user.Append("user.route"))
					}
					return
				}

				// 检查会话
				m.Option("sessid", "")
				m.Cmd("aaa.auth", "nodes", m.Option("node.route"), "session").Table(func(line map[string]string) {
					if m.Cmds("aaa.auth", line["key"], "username", m.Option("user.name")) {
						m.Option("sessid", line["key"])
					}
				})

				if m.Options("remote_code") {
					if !m.Options("sessid") {
						// 用户签名
						hash, _ := kit.Hash("rand", m.Option("text.time", m.Time("stamp")), m.Option("node.route"))
						m.Option("user.cert", m.Cmd("aaa.auth", "username", m.Option("user.name"), "cert").Append("meta"))
						m.Option("user.sign", m.Spawn().Cmdx("ssh.remote", m.Option("user.route"), "sync", "check", "user", m.Option("node.route"), hash))

						// 代理验签
						if !m.Options("user.cert") || !m.Options("user.sign") || !m.Cmds("aaa.rsa", "verify", m.Option("user.cert"), m.Option("user.sign"), hash) {
							m.Log("warn", "user error")
							m.Echo("no right of %s", m.Option("text.route"))
							return
						}
						// 创建会话
						m.Option("sessid", m.Cmdx("aaa.sess", "nodes", "username", m.Option("user.name")))
						m.Cmd("aaa.auth", m.Option("sessid"), "nodes", m.Option("node.route"))
					}

					// 创建空间
					if m.Option("bench", m.Cmd("aaa.sess", m.Option("sessid"), "bench").Append("key")); !m.Options("bench") {
						m.Option("bench", m.Cmdx("aaa.work", m.Option("sessid"), "nodes"))
					}
				}

				if !m.Options("remote_code") || m.Cmds("aaa.work", m.Option("bench"), "right", m.Option("user.name"), "remote", arg[0]) {
					m.Option("username", m.Option("user.name"))
					m.Option("current_ctx", kit.Select("ssh", m.Cmdx("aaa.auth", m.Option("bench"), "data", "target")))
					// 执行命令
					msg := m.Find(m.Option("current_ctx")).Cmd(arg).CopyTo(m)
					m.Cmd("aaa.auth", m.Option("bench"), "data", "target", msg.Cap("module"))
				} else {
					m.Echo("no right %s %s", "remote", arg[0])
				}
			}
			return
		}},
		"proxy": &ctx.Command{Name: "proxy [proxy.route]", Help: "代理节点", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmd("aaa.auth", "proxy")
				return
			}
			m.Cmd("aaa.auth", "proxy", arg[0])
			return
		}},
		"login": &ctx.Command{Name: "login client.route", Help: "用户节点", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmd("aaa.auth", "login")
				return
			}

			if !m.Cmds("ssh.remote", arg[0], "login") {
				m.Echo("error: ").Echo("login failure")
				return
			}

			sess := m.Cmd("aaa.auth", "username", m.Option("username"), "session", "login").Append("key")
			if sess == "" {
				sess = m.Cmdx("aaa.sess", "login", "username", m.Option("username"))
			}

			m.Cmd("aaa.auth", sess, "login", arg[0])
			m.Echo(sess)
			return
		}},
		"share": &ctx.Command{Name: "share serve.route role client.route...", Help: "共享权限", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmd("aaa.auth", "apply").Table(func(node map[string]string) {
					m.Cmd("aaa.auth", node["key"], "session", "apply").Table(func(sess map[string]string) {
						m.Cmd("aaa.auth", sess["key"], "username").Table(func(user map[string]string) {
							m.Add("append", "time", sess["create_time"])
							m.Add("append", "user", user["meta"])
							m.Add("append", "node", node["meta"])
						})
					})
				})
				m.Table()
				return
			}

			if len(arg) == 2 {
				m.Option("user.route", arg[1])
				m.Cmd("ssh.remote", "", "share", arg[1:])
				return
			}

			m.Option("user.sign", "yes")
			m.Cmd("ssh.remote", arg[0], "sync", "share", arg[1:])
			return
		}},
		"check": &ctx.Command{Name: "check proxy.route client.route", Help: "权限检查", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			return
		}},
		"sh": &ctx.Command{Name: "sh [[node] name] cmd...", Help: "发送命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Echo(m.Conf("current"))
				return
			}

			if arg[0] == "node" {
				m.Conf("current", arg[1])
				arg = arg[2:]
			} else if m.Confm("node", arg[0]) != nil {
				m.Conf("current", arg[0])
				arg = arg[1:]
			}

			msg := m.Cmd("ssh.remote", m.Conf("current"), arg)
			m.Copy(msg, "append")
			m.Copy(msg, "result")
			return
		}},
		"cp": &ctx.Command{Name: "cp [[node] name] filename", Help: "发送文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Echo(m.Conf("current"))
				return
			}

			if arg[0] == "node" {
				m.Conf("current", arg[1])
				arg = arg[2:]
			} else if m.Confm("node", arg[0]) != nil {
				m.Conf("current", arg[0])
				arg = arg[1:]
			}

			if arg[0] == "save" {
				buf, e := base64.StdEncoding.DecodeString(m.Option("filebuf"))
				m.Assert(e)

				f, e := os.OpenFile(path.Join("tmp", m.Option("filename")), os.O_RDWR|os.O_CREATE, 0666)
				f.WriteAt(buf, int64(m.Optioni("filepos")))
				return e
			}

			p := m.Cmdx("nfs.path", arg[0])
			f, e := os.Open(p)
			m.Assert(e)
			s, e := f.Stat()
			m.Assert(e)

			buf := make([]byte, 1024)

			for i := int64(0); i < s.Size(); i += 1024 {
				n, _ := f.ReadAt(buf, i)
				if n == 0 {
					break
				}

				buf = buf[:n]
				msg := m.Spawn()
				msg.Option("filename", arg[0])
				msg.Option("filesize", s.Size())
				msg.Option("filepos", i)
				msg.Option("filebuf", base64.StdEncoding.EncodeToString(buf))
				msg.Cmd("remote", m.Conf("current"), "cp", "save", arg[0])
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
