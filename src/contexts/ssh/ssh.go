package ssh // {{{
// }}}
import ( // {{{
	"bufio"
	"contexts"
	"fmt"
	"net"
	"net/url"
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
	c.Caches = map[string]*ctx.Cache{
		"nsend":  &ctx.Cache{Name: "消息发送数量", Value: "0", Help: "消息发送数量"},
		"nrecv":  &ctx.Cache{Name: "消息接收数量", Value: "0", Help: "消息接收数量"},
		"target": &ctx.Cache{Name: "消息接收模块", Value: "ssh", Help: "消息接收模块"},
		"result": &ctx.Cache{Name: "前一条指令执行结果", Value: "", Help: "前一条指令执行结果"},
		"sessid": &ctx.Cache{Name: "会话令牌", Value: "", Help: "会话令牌"},
	}
	c.Configs = map[string]*ctx.Config{}

	s := new(SSH)
	s.Context = c
	return s
}

// }}}
func (ssh *SSH) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	if ssh.Context == Index {
		Pulse = m
	}
	return ssh
}

// }}}
func (ssh *SSH) Start(m *ctx.Message, arg ...string) bool { // {{{
	return false

	ssh.Group = ""
	ssh.Owner = nil
	ssh.Conn = m.Data["io"].(net.Conn)
	ssh.Reader = bufio.NewReader(ssh.Conn)
	ssh.Writer = bufio.NewWriter(ssh.Conn)
	ssh.send = make(map[string]*ctx.Message)
	m.Log("info", nil, "%s connect %v <-> %v", Pulse.Cap("nhost"), ssh.Conn.LocalAddr(), ssh.Conn.RemoteAddr())

	target, msg := m.Target(), m.Spawn(m.Target())

	for {
		line, e := ssh.Reader.ReadString('\n')
		m.Assert(e)

		if line = strings.TrimSpace(line); len(line) == 0 {
			if msg.Log("info", nil, "remote: %v", msg.Meta["option"]); msg.Has("detail") {
				msg.Log("info", nil, "%d exec: %v", m.Capi("nrecv", 1), msg.Meta["detail"])

				msg.Cmd(msg.Meta["detail"])
				target = msg.Target()
				m.Cap("target", target.Name)

				for _, v := range msg.Meta["result"] {
					fmt.Fprintf(ssh.Writer, "result: %s\n", url.QueryEscape(v))
				}

				fmt.Fprintf(ssh.Writer, "nsend: %s\n", msg.Get("nrecv"))
				for _, k := range msg.Meta["append"] {
					for _, v := range msg.Meta[k] {
						fmt.Fprintf(ssh.Writer, "%s: %s\n", k, v)
					}
				}
				fmt.Fprintf(ssh.Writer, "\n")
				ssh.Writer.Flush()
			} else {
				msg.Log("info", nil, "%s echo: %v", msg.Get("nsend"), msg.Meta["result"])

				m.Cap("result", msg.Get("result"))
				msg.Meta["append"] = msg.Meta["option"]
				send := ssh.send[msg.Get("nsend")]
				send.Meta = msg.Meta
				send.Recv <- true
			}
			msg = m.Spawn(target)
			continue
		}

		ls := strings.SplitN(line, ":", 2)
		ls[0] = strings.TrimSpace(ls[0])
		ls[1], e = url.QueryUnescape(strings.TrimSpace(ls[1]))
		m.Assert(e)
		msg.Add("option", ls[0], ls[1])
	}
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
			msg := m.Sess("file", "nfs")
			msg.Call(func(ok bool) (done bool, up bool) {
				if ok {
					sub := msg.Spawn(m.Target())
					sub.Start(fmt.Sprintf("host%d", Pulse.Capi("nhost", 1)), "打开文件", sub.Meta["detail"]...)
					sub.Cap("stream", msg.Target().Name)
					sub.Target().Sessions["file"] = msg
					sub.Echo(sub.Target().Name)
					sub.Spawn(sub.Target()).Cmd("send", "context", "ssh")
					sub.Spawn(sub.Target()).Cmd("send", "route", sub.Target().Name, msg.Cap("route"))
				}
				return false, true
			}, false).Cmd(m.Meta["detail"])

		}},
		"dial": &ctx.Command{Name: "dial address protocol", Help: "建立连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Sess("file", "nfs")
			msg.Call(func(ok bool) (done bool, up bool) {
				if ok {
					m.Sess("file").Cmd("send", "context", "ssh")
					m.Cap("stream", msg.Target().Name)
					return true, true
				}
				return false, false
			}, false).Cmd(m.Meta["detail"])

		}},
		"send": &ctx.Command{Name: "send arg...", Help: "打开连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Sess("file")
			msg.Copy(m, "detail").Call(func(ok bool) (done bool, up bool) {
				return ok, ok
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
						msg.Call(func(ok bool) (done bool, up bool) {
							m.Copy(msg, "result")
							return ok, ok
						}, false).Cmd("send", "search", strings.Join(target[1:], "."), arg[1:])

						miss = false
					}
					return miss
				})

				if miss {
					msg := m.Spawn(c)
					msg.Call(func(ok bool) (done bool, up bool) {
						m.Copy(msg, "result")
						return ok, ok
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
