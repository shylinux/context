package ssh // {{{
// }}}
import ( // {{{
	"bufio"
	"contexts"
	"fmt"
	"net"
	"strings"
)

// }}}

type SSH struct {
	send map[string]*ctx.Message
	*bufio.Writer
	*bufio.Reader
	net.Conn

	*ctx.Context
}

func (ssh *SSH) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(SSH)
	s.Context = c
	return s
}

// }}}
func (ssh *SSH) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	ssh.Context.Master(nil)
	if ssh.Context == Index {
		Pulse = m
	}
	return ssh
}

// }}}
func (ssh *SSH) Start(m *ctx.Message, arg ...string) bool { // {{{
	return false
}

// }}}
func (ssh *SSH) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch ssh.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches: map[string]*ctx.Cache{
		"nhost": &ctx.Cache{Name: "主机数量", Value: "0", Help: "主机数量"},
		"route": &ctx.Cache{Name: "route", Value: "ssh", Help: "主机数量"},
	},
	Configs: map[string]*ctx.Config{
		"route": &ctx.Config{Name: "route", Value: "com", Help: "主机数量"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address protocol", Help: "监听连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Find("nfs")
			msg.Call(func(ok bool, file *ctx.Message) (bool, *ctx.Message) {
				if ok {
					sub := msg.Spawn(m.Target())
					sub.Start(fmt.Sprintf("host%d", Pulse.Capi("nhost", 1)), "远程主机")

					sub.Cap("stream", file.Target().Name)
					sub.Sess("file", "nfs."+file.Target().Name)
					sub.Sess("file").Cmd("send", "context", "ssh")
					sub.Sess("file").Cmd("send", "route", sub.Target().Name, msg.Cap("route"))
					return false, sub
				}
				return false, nil
			}, false).Cmd(m.Meta["detail"])

		}},
		"dial": &ctx.Command{Name: "dial address protocol", Help: "建立连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Sess("file", "nfs")
			msg.Call(func(ok bool, file *ctx.Message) (bool, *ctx.Message) {
				if ok {
					m.Cap("stream", msg.Target().Name)
					m.Sess("file").Cmd("send", "context", "ssh")
					return true, m
				}
				return false, nil
			}, false).Cmd(m.Meta["detail"])

		}},
		"send": &ctx.Command{Name: "send arg...", Help: "打开连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Sess("file")
			msg.Copy(m, "detail").Call(func(ok bool, file *ctx.Message) (bool, *ctx.Message) {
				return ok, file
			}, false).Cmd()
			m.Copy(msg, "result")
		}},
		"pwd": &ctx.Command{Name: "pwd", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Echo(m.Cap("route"))
		}},
		"route": &ctx.Command{Name: "route", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Conf("route", arg[0])
			m.Cap("route", arg[1]+"."+arg[0])
		}},
		"search": &ctx.Command{Name: "search route cmd arg...", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*SSH); m.Assert(ok) {
				if len(arg[0]) == 0 {
					msg := m.Spawn(m.Target()).Cmd(arg[1:])
					m.Copy(msg, "result")
					return
				}

				miss := true
				target := strings.Split(arg[0], ".")
				m.Travel(m.Target(), func(m *ctx.Message) bool {
					if m.Target().Name == target[0] {
						msg := m.Spawn(m.Target())
						msg.Call(func(ok bool, host *ctx.Message) (bool, *ctx.Message) {
							m.Copy(msg, "result")
							return ok, host
						}, false).Cmd("send", "search", strings.Join(target[1:], "."), arg[1:])

						miss = false
					}
					return miss
				})

				if miss {
					msg := m.Spawn(c)
					msg.Call(func(ok bool, host *ctx.Message) (bool, *ctx.Message) {
						m.Copy(msg, "result")
						return ok, host
					}, false).Cmd("send", "search", arg)
				}
			}
		}},
		"dispatch": &ctx.Command{Name: "dispatch route cmd arg...", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Travel(m.Target(), func(m *ctx.Message) bool {

				msg := m.Spawn(m.Target())
				msg.Cmd("send", arg)
				return true
			})
		}},
		"register": &ctx.Command{Name: "remote detail...", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			ssh, ok := m.Target().Server.(*SSH)
			m.Assert(ok)

			m.Capi("nsend", 1)
			m.Add("option", "nrecv", m.Cap("nsend"))
			ssh.send[m.Cap("nsend")] = m

			for _, v := range arg {
				fmt.Fprintf(ssh.Writer, "detail: %v\n", v)
			}
			for _, k := range m.Meta["option"] {
				if k == "args" {
					continue
				}
				for _, v := range m.Meta[k] {
					fmt.Fprintf(ssh.Writer, "%s: %s\n", k, v)
				}
			}
			fmt.Fprintf(ssh.Writer, "\n")
			ssh.Writer.Flush()

			m.Recv = make(chan bool)
			<-m.Recv
		}},
	},
}

func init() {
	ssh := &SSH{}
	ssh.Context = Index
	ctx.Index.Register(Index, ssh)
}
