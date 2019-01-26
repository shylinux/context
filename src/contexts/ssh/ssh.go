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
		"nhost":    &ctx.Cache{Name: "nhost", Value: "0", Help: "主机数量"},
		"hostname": &ctx.Cache{Name: "hostname", Value: "shy", Help: "本机域名"},
	},
	Configs: map[string]*ctx.Config{
		"host":     &ctx.Config{Name: "host", Value: map[string]interface{}{}, Help: "主机信息"},
		"hostport": &ctx.Config{Name: "hostport", Value: "", Help: "主机域名"},
		"hostname": &ctx.Config{Name: "hostname", Value: "com", Help: "主机域名"},
		"current":  &ctx.Config{Name: "current", Value: "", Help: "当前主机"},
		"timer":    &ctx.Config{Name: "timer", Value: "", Help: "当前主机"},
	},
	Commands: map[string]*ctx.Command{
		"remote": &ctx.Command{Name: "remote listen|dial args...", Help: "远程连接", Form: map[string]int{"right": 1, "hostname": 1}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 { // 查看主机
				m.Cmdy("ctx.config", "host")
				return
			}

			switch arg[0] {
			case "redial": // 断线重连
				if !m.Caps("hostname") {
					m.Cmdx("remote", "dial", arg[1:])
				}
			case "listen", "dial":
				m.Call(func(nfs *ctx.Message) *ctx.Message {
					if arg[0] == "listen" && nfs.Options("hostport") { // 监听端口
						m.Conf("hostport", nfs.Option("hostport"))

					} else if arg[0] == "dial" {
						if m.Confs("timer") { // 断线重连
							m.Conf("timer", m.Cmdx("cli.timer", "delete", m.Conf("timer")))
						}

						m.Spawn(nfs.Target()).Call(func(host *ctx.Message) *ctx.Message {
							m.Confv("host", host.Result(1), map[string]interface{}{ // 添加主机
								"module":      nfs.Format("target"),
								"create_time": m.Time(),
								"access_time": m.Time(),
								"username":    m.Option("right"),
								"cm_target":   "ctx.web.code",
							})

							m.Cap("stream", nfs.Format("target"))
							m.Cap("hostname", host.Result(0))
							if !m.Confs("current") {
								m.Conf("current", host.Result(1))
							}

							nfs.Free(func(nfs *ctx.Message) bool { // 连接中断
								m.Conf("timer", m.Cmdx("cli.timer", "repeat", "10s", "context", "ssh", "remote", "redial", arg[1:]))

								m.Log("info", "delete host %s", host.Result(1))
								delete(m.Confm("host"), host.Result(1))
								m.Cap("hostname", "")
								m.Cap("stream", "")
								return true
							})
							return nil
						}, "send", "recv", "add", m.Confx("hostname"))
					}
					return nil
				}, "nfs.remote", arg)
			case "recv":
				switch arg[1] {
				case "add":
					if host := m.Confm("host", arg[2]); host == nil { // 添加主机
						m.Confv("host", arg[2], map[string]interface{}{
							"module":      m.Format("source"),
							"create_time": m.Time(),
							"access_time": m.Time(),
							"username":    m.Option("right"),
							"cm_target":   "ctx.web.code",
						})
					} else if len(arg) > 3 && arg[3] == kit.Format(host["token"]) { // 断线重连
						host["access_time"] = m.Time()
						host["module"] = m.Format("source")
					} else { // 域名冲突
						arg[2] = fmt.Sprintf("%s_%d", arg[2], m.Capi("nhost", 1))
						m.Confv("host", arg[2], map[string]interface{}{
							"module":      m.Format("source"),
							"create_time": m.Time(),
							"access_time": m.Time(),
							"username":    m.Option("right"),
							"cm_target":   "ctx.web.code",
						})
					}

					if !m.Confs("current") {
						m.Conf("current", arg[2])
					}

					m.Echo(arg[2]).Echo(m.Cap("hostname")).Back(m)
					m.Sess("ms_source", false).Free(func(msg *ctx.Message) bool { // 断线清理
						m.Log("info", "delete host %s", arg[2])
						delete(m.Confm("host"), arg[2])
						return true
					})
				}

			default:
				names, arg := strings.SplitN(arg[0], ".", 2), arg[1:]

				if names[0] == "" { // 本地执行
					host := m.Confm("host", m.Option("hostname"))
					user := kit.Format(kit.Chain(host, "username"))

					sessid := m.Cmd("aaa.user", user, "ssh").Append("meta")
					if sessid == "" { // 创建会话
						sessid = m.Cmdx("aaa.sess", "ssh", "ip", "what")
						m.Cmd("aaa.sess", sessid, user, "ppid", "what")
					}

					bench := m.Cmd("aaa.sess", sessid, "bench").Append("meta")
					if bench == "" { // 创建空间
						bench = m.Cmdx("aaa.work", sessid, "ssh")
					}

					if m.Cmds("aaa.work", bench, "right", user, "remote", arg[0]) { // 执行命令
						msg := m.Find(m.Option("current_ctx", kit.Format(host["cm_target"]))).Cmd(arg).CopyTo(m)
						host["cm_target"] = msg.Cap("module")
					} else {
						m.Echo("no right %s %s", "remote", arg[0])
					}

					// 返回结果
					m.Back(m)
					return
				}

				//同步或异步
				sync := !m.Options("remote_code")
				switch arg[0] {
				case "async", "sync":
					sync, arg = arg[0] == "sync", arg[1:]
				}

				rest := kit.Select("", names, 1)
				m.Option("hostname", m.Cap("hostname"))

				if names[0] == "*" { // 广播命令
					m.Confm("host", func(name string, host map[string]interface{}) {
						m.Find(kit.Format(host["module"]), true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
							return m.Copy(sub, "append").Copy(sub, "result")
						}, "send", "", arg)
					})

				} else if m.Confm("host", names[0], func(host map[string]interface{}) { // 单播命令
					m.Find(kit.Format(host["module"]), true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
						return m.Copy(sub, "append").Copy(sub, "result")
					}, "send", rest, arg)

				}) == nil { // 回溯命令
					m.Find(m.Cap("stream"), true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
						return m.Copy(sub, "append").Copy(sub, "result")
					}, "send", strings.Join(names, "."), arg)
				}
			}
			return
		}},
		"sh": &ctx.Command{Name: "sh [[host] name] cmd...", Help: "发送命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Echo(m.Conf("current"))
				return
			}

			if arg[0] == "host" {
				m.Conf("current", arg[1])
				arg = arg[2:]
			} else if m.Confm("host", arg[0]) != nil {
				m.Conf("current", arg[0])
				arg = arg[1:]
			}

			msg := m.Cmd("ssh.remote", m.Conf("current"), arg)
			m.Copy(msg, "append")
			m.Copy(msg, "result")
			return
		}},
		"cp": &ctx.Command{Name: "cp [[host] name] filename", Help: "发送文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Echo(m.Conf("current"))
				return
			}

			if arg[0] == "host" {
				m.Conf("current", arg[1])
				arg = arg[2:]
			} else if m.Confm("host", arg[0]) != nil {
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
