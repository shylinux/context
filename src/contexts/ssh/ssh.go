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
		"hostname": &ctx.Config{Name: "hostname", Value: "com", Help: "主机域名"},
		"current":  &ctx.Config{Name: "current", Value: "", Help: "当前主机"},
		"timer":    &ctx.Config{Name: "timer", Value: "", Help: "当前主机"},
	},
	Commands: map[string]*ctx.Command{
		"remote": &ctx.Command{Name: "remote listen|dial args...", Help: "远程连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmdy("ctx.config", "host")
				return
			}

			switch arg[0] {
			case "redial":
				if !m.Caps("hostname") {
					m.Cmdx("remote", "dial", arg[1:])
				}
			case "listen", "dial":
				m.Call(func(nfs *ctx.Message) *ctx.Message {
					if arg[0] == "dial" {
						if m.Confs("timer") {
							m.Conf("timer", m.Cmdx("cli.timer", "delete", m.Conf("timer")))
						}
						m.Spawn(nfs.Target()).Call(func(cmd *ctx.Message) *ctx.Message {
							m.Cap("stream", nfs.Format("target"))
							m.Cap("hostname", cmd.Result(0))

							m.Confv("host", cmd.Result(1), map[string]interface{}{
								"module":      nfs.Format("target"),
								"create_time": m.Time(),
								"access_time": m.Time(),
							})
							if !m.Confs("current") {
								m.Conf("current", cmd.Result(1))
							}

							nfs.Free(func(nfs *ctx.Message) bool {
								m.Conf("timer", m.Cmdx("cli.timer", "repeat", "10s", "context", "ssh", "remote", "redial", arg[1:]))

								m.Log("info", "delete host %s", cmd.Result(1))
								delete(m.Confm("host"), cmd.Result(1))
								m.Cap("hostname", "")
								m.Cap("stream", "")
								return true
							})
							return nil
						}, "send", "recv", "add", m.Conf("hostname"))
					}
					return nil
				}, "nfs.remote", arg)

			case "recv":
				switch arg[1] {
				case "add":
					if host := m.Confm("host", arg[2]); host == nil {
						m.Confv("host", arg[2], map[string]interface{}{
							"module":      m.Format("source"),
							"create_time": m.Time(),
							"access_time": m.Time(),
						})
					} else if len(arg) > 3 && arg[3] == kit.Format(host["token"]) {
						host["access_time"] = m.Time()
						host["module"] = m.Format("source")
					} else {
						arg[2] = fmt.Sprintf("%s_%d", arg[2], m.Capi("nhost", 1))
						m.Confv("host", arg[2], map[string]interface{}{
							"module":      m.Format("source"),
							"create_time": m.Time(),
							"access_time": m.Time(),
						})
					}
					if !m.Confs("current") {
						m.Conf("current", arg[2])
					}

					m.Echo(arg[2]).Echo(m.Cap("hostname")).Back(m)
					m.Sess("ms_source", false).Free(func(msg *ctx.Message) bool {
						m.Log("info", "delete host %s", arg[2])
						delete(m.Confm("host"), arg[2])
						return true
					})
				}

			default:
				names := strings.SplitN(arg[0], ".", 2)
				if names[0] == "" { // 本地执行
					host := m.Confm("host", m.Option("hostname"))
					m.Option("current_ctx", kit.Format(host["cm_target"]))

					msg := m.Find(kit.Format(host["cm_target"])).Cmd(arg[1:])
					m.Copy(msg, "append").Copy(msg, "result")
					host["cm_target"] = msg.Cap("module")

					m.Back(m)
					return
				}

				m.Option("hostname", m.Cap("hostname"))
				sync := !m.Options("remote_code") //同步或异步
				if arg[1] == "async" {
					sync, arg = false, arg[2:]
				} else if arg[1] == "sync" {
					sync, arg = true, arg[2:]
				} else {
					arg = arg[1:]
				}

				rest := kit.Select("", names, 1)
				if names[0] == "*" {
					m.Confm("host", func(name string, host map[string]interface{}) {
						m.Find(kit.Format(host["module"]), true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
							return m.Copy(sub, "append").Copy(sub, "result")
						}, "send", "", arg)
					})

				} else if m.Confm("host", names[0], func(host map[string]interface{}) {
					m.Find(kit.Format(host["module"]), true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
						return m.Copy(sub, "append").Copy(sub, "result")
					}, "send", rest, arg)
					m.Log("fuck", "m %v", m.Meta)

				}) == nil {
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
