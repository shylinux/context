package ssh // {{{
// }}}
import ( // {{{
	"bufio"
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// }}}

type SSH struct {
	send map[string]*ctx.Message
	*bufio.Reader
	net.Conn
	*ctx.Context
}

func (ssh *SSH) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"nsend": &ctx.Cache{Name: "nsend", Value: "0", Help: "消息发送数量"},
	}
	c.Configs = map[string]*ctx.Config{}
	c.Commands = map[string]*ctx.Command{}

	s := new(SSH)
	s.Context = c

	s.send = make(map[string]*ctx.Message)
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
	ssh.Owner = nil
	ssh.Conn = m.Data["io"].(net.Conn)
	ssh.Reader = bufio.NewReader(ssh.Conn)
	m.Log("info", nil, "%d remote %v", 0, ssh.Conn.RemoteAddr())

	target := m.Target
	msg := m.Spawn(target)

	for {
		line, e := ssh.Reader.ReadString('\n')
		m.Assert(e)

		if line = strings.TrimSpace(line); len(line) == 0 {
			if msg.Has("detail") {
				msg.Log("info", nil, "remote: %v", msg.Meta["detail"])
				msg.Log("info", nil, "remote: %v", msg.Meta["option"])
				msg.Cmd(msg.Meta["detail"]...)
				target = msg.Target

				fmt.Fprintf(ssh.Conn, "result: ")
				for _, v := range msg.Meta["result"] {
					fmt.Fprintf(ssh.Conn, "%s", url.QueryEscape(v))
				}
				fmt.Fprintf(ssh.Conn, "\n")

				msg.Meta["append"] = append(msg.Meta["append"], "nsend")
				msg.Add("append", "nsend", msg.Get("nsend"))
				for _, k := range msg.Meta["append"] {
					for _, v := range msg.Meta[k] {
						fmt.Fprintf(ssh.Conn, "%s: %s\n", k, v)
					}
				}
				fmt.Fprintf(ssh.Conn, "\n")
			} else if msg.Has("result") {
				msg.Log("info", nil, "remote: %v", msg.Meta["result"])
				msg.Log("info", nil, "remote: %v", msg.Meta["append"])
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
	return false
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches: map[string]*ctx.Cache{
		"nhost": &ctx.Cache{Name: "nhost", Value: "0", Help: "主机数量"},
	},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address", Help: "监听连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			tcp := m.Find("tcp", true) // {{{
			tcp.Cmd(m.Meta["detail"]...)
			return ""
			// }}}
		}},
		"dial": &ctx.Command{Name: "dial address", Help: "建立连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			tcp := m.Find("tcp", true) // {{{
			tcp.Cmd(m.Meta["detail"]...)
			return ""
			// }}}
		}},
		"open": &ctx.Command{Name: "open", Help: "打开连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			m.Start("host"+Pulse.Cap("nhost"), "主机连接") // {{{
			return ""
			// }}}
		}},
		"remote": &ctx.Command{Name: "remote detail...", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			ssh, ok := m.Target.Server.(*SSH) // {{{
			m.Assert(ok)
			m.Capi("nsend", 1)
			m.Recv = make(chan bool)
			m.Add("option", "nsend", m.Cap("nsend"))
			ssh.send[m.Cap("nsend")] = m

			for _, v := range arg {
				fmt.Fprintf(ssh.Conn, "detail: %v\n", v)
			}
			for _, k := range m.Meta["option"] {
				for _, v := range m.Meta[k] {
					fmt.Fprintf(ssh.Conn, "%s: %s\n", k, v)
				}
			}
			fmt.Fprintf(ssh.Conn, "\n")
			<-m.Recv
			return ""
			// }}}
		}},
	},
}

func init() {
	ssh := &SSH{}
	ssh.Context = Index
	ctx.Index.Register(Index, ssh)
}
