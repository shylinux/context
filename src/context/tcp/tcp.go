package tcp // {{{
// }}}
import ( // {{{
	"context"

	"fmt"
	"net"
	"strconv"
)

// }}}

type TCP struct {
	net.Conn
	net.Listener
	*ctx.Context
}

func (tcp *TCP) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"protocol": &ctx.Cache{Name: "protocol(tcp/tcp4/tcp6)", Value: m.Conf("protocol"), Help: "监听地址"},
		"security": &ctx.Cache{Name: "security(true/false)", Value: m.Conf("security"), Help: "加密通信"},
		"address":  &ctx.Cache{Name: "address", Value: "", Help: "监听地址"},
	}
	c.Configs = map[string]*ctx.Config{}

	s := new(TCP)
	s.Context = c
	return s

}

// }}}
func (tcp *TCP) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	if m.Target == Index {
		Pulse = m
	}

	return tcp
}

// }}}
func (tcp *TCP) Start(m *ctx.Message, arg ...string) bool { // {{{
	if len(arg) > 1 {
		m.Cap("address", arg[1])
	}
	if len(arg) > 2 {
		m.Cap("security", arg[2])
	}

	switch arg[0] {
	case "dial":
		c, e := net.Dial(m.Cap("protocol"), m.Cap("address"))
		m.Assert(e)
		tcp.Conn = c
		m.Log("info", nil, "dial(%d) %v->%v", m.Capi("nclient", 1), tcp.LocalAddr(), tcp.RemoteAddr())
		m.Cap("stream", fmt.Sprintf("%s->%s", tcp.LocalAddr(), tcp.RemoteAddr()))

		// m.Reply(c.LocalAddr().String()).Put("option", "io", c).Cmd("open")
		return false
	case "accept":
		c, e := m.Data["io"].(net.Conn)
		m.Assert(e)
		tcp.Conn = c
		m.Log("info", nil, "accept(%d) %v<-%v", m.Capi("nclient", 1), tcp.LocalAddr(), tcp.RemoteAddr())
		m.Cap("stream", fmt.Sprintf("%s<-%s", tcp.LocalAddr(), tcp.RemoteAddr()))

		s, e := m.Data["source"].(*ctx.Context)
		m.Assert(e)
		msg := m.Spawn(s).Put("option", "io", c)
		msg.Cmd("open")
		msg.Cap("stream", tcp.RemoteAddr().String())

		if tcp.Sessions == nil {
			tcp.Sessions = make(map[string]*ctx.Message)
		}
		tcp.Sessions["open"] = msg
		msg.Name = "open"

		// m.Reply(c.RemoteAddr().String())
		return false
	}

	l, e := net.Listen(m.Cap("protocol"), m.Cap("address"))
	m.Assert(e)
	tcp.Listener = l
	m.Log("info", nil, "listen(%d) %v", m.Capi("nlisten", 1), l.Addr())
	m.Cap("stream", fmt.Sprintf("%s", l.Addr()))

	for {
		c, e := l.Accept()
		m.Assert(e)
		m.Spawn(Index).Put("option", "io", c).Put("option", "source", m.Source).Start(fmt.Sprintf("com%d", m.Capi("nclient", 1)), "网络连接", "accept", c.RemoteAddr().String())
	}

	return true
}

// }}}
func (tcp *TCP) Close(m *ctx.Message, arg ...string) bool { // {{{
	if tcp.Context == Index {
		return false
	}

	switch tcp.Context {
	case m.Target:
	case m.Source:
		if tcp.Listener != nil {
			return false
		}

	}

	if tcp.Listener != nil {
		m.Log("info", nil, "close(%d) %v", Pulse.Capi("nlisten", -1)+1, tcp.Listener.Addr())
		tcp.Listener.Close()
	}
	if tcp.Conn != nil {
		m.Log("info", nil, "close %v", tcp.Conn.LocalAddr())
		tcp.Conn.Close()
	}
	return true
}

// }}}

var Index = &ctx.Context{Name: "tcp", Help: "网络中心",
	Caches: map[string]*ctx.Cache{
		"nlisten": &ctx.Cache{Name: "nlisten", Value: "0", Help: "监听数量"},
		"nclient": &ctx.Cache{Name: "nclient", Value: "0", Help: "连接数量"},
	},
	Configs: map[string]*ctx.Config{
		"protocol": &ctx.Config{Name: "protocol(tcp/tcp4/tcp6)", Value: "tcp4", Help: "连接协议"},
		"security": &ctx.Config{Name: "security(true/false)", Value: "false", Help: "加密通信"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen [address [security]]", Help: "监听连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				m.Travel(nil, func(m *ctx.Message) bool {
					if tcp, ok := m.Target.Server.(*TCP); ok && tcp.Listener != nil {
						m.Echo("%s %v\n", m.Target.Name, tcp.Addr())
					}
					return true
				})
			default:
				m.Start(fmt.Sprintf("pub%d", m.Capi("nlisten")+1), "网络监听", m.Meta["detail"]...)
			}
			return ""
			// }}}
		}},
		"dial": &ctx.Command{Name: "dial [address [security]]", Help: "建立连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				m.Travel(nil, func(m *ctx.Message) bool {
					if tcp, ok := m.Target.Server.(*TCP); ok && tcp.Conn != nil {
						m.Echo("%s %v<->%v\n", m.Target.Name, tcp.LocalAddr(), tcp.RemoteAddr())
					}
					return true
				})
			default:
				m.Start(fmt.Sprintf("com%d", m.Capi("nclient")+1), "网络连接", m.Meta["detail"]...)
			}
			return ""
			// }}}
		}},
		"send": &ctx.Command{Name: "send message", Help: "发送消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			if tcp, ok := m.Target.Server.(*TCP); ok && tcp.Conn != nil { // {{{
				tcp.Conn.Write([]byte(arg[0]))
			}
			return ""
			// }}}
		}},
		"recv": &ctx.Command{Name: "recv size", Help: "接收消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			size, e := strconv.Atoi(arg[0])
			m.Assert(e)
			if tcp, ok := m.Target.Server.(*TCP); ok && tcp.Conn != nil { // {{{
				buf := make([]byte, size)
				tcp.Conn.Read(buf)
				return string(buf)
			}
			return ""
			// }}}
		}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{
			Commands: map[string]*ctx.Command{
				"listen": &ctx.Command{},
				"dial":   &ctx.Command{},
			},
		},
	},
}

var Pulse *ctx.Message

func init() {
	tcp := &TCP{}
	tcp.Context = Index
	ctx.Index.Register(Index, tcp)
}
