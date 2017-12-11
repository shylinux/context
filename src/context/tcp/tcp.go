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
		"address":  &ctx.Cache{Name: "address", Value: arg[1], Help: "监听地址"},
	}
	c.Configs = map[string]*ctx.Config{}

	if len(arg) > 2 {
		m.Cap("security", arg[2])
	}

	s := new(TCP)
	s.Context = c
	return s

}

// }}}
func (tcp *TCP) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	return tcp
}

// }}}
func (tcp *TCP) Start(m *ctx.Message, arg ...string) bool { // {{{
	switch arg[0] {
	case "dial":
		c, e := net.Dial(m.Cap("protocol"), m.Cap("address"))
		m.Assert(e)
		tcp.Conn = c

		m.Log("info", nil, "dial(%d) %v->%v", m.Capi("nclient"), c.LocalAddr(), c.RemoteAddr())
		// m.Reply(c.LocalAddr().String()).Put("option", "io", c).Cmd("open")
		return false
	case "accept":
		return false
	}

	l, e := net.Listen(m.Cap("protocol"), m.Cap("address"))
	m.Assert(e)
	tcp.Listener = l

	m.Log("info", nil, "listen(%d) %v", m.Capi("nlisten"), l.Addr())

	for {
		c, e := l.Accept()
		m.Assert(e)

		s, i := m.Target, 0
		m.BackTrace(func(m *ctx.Message) bool {
			s = m.Target
			if i++; i == 2 {
				return false
			}
			return true
		})

		msg := m.Spawn(s)
		msg.Start(fmt.Sprintf("com%d", m.Capi("nclient", 1)), c.RemoteAddr().String(), "accept", c.RemoteAddr().String())
		msg.Log("info", nil, "accept(%d) %v<-%v", m.Capi("nclient"), c.LocalAddr(), c.RemoteAddr())

		if tcp, ok := msg.Target.Server.(*TCP); ok {
			tcp.Conn = c
		}
		rep := m.Reply(c.RemoteAddr().String())
		rep.Source = msg.Target
		rep.Put("option", "io", c).Cmd("open")
	}

	return true
}

// }}}
func (tcp *TCP) Close(m *ctx.Message, arg ...string) bool { // {{{
	if tcp.Listener != nil {
		m.Log("info", nil, "close(%d) %v", m.Capi("nlisten", -1)+1, tcp.Listener.Addr())
		tcp.Listener.Close()
		return true
	}
	if tcp.Conn != nil {
		tcp.Conn.Close()
		return true
	}

	return false
}

// }}}

var Index = &ctx.Context{Name: "tcp", Help: "网络连接",
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
				m.Travel(m.Target, func(m *ctx.Message) bool {
					if tcp, ok := m.Target.Server.(*TCP); ok && tcp.Listener != nil {
						m.Echo("%s %v\n", m.Target.Name, tcp.Addr())
					}
					return true
				})
			default:
				m.Start(fmt.Sprintf("pub%d", m.Capi("nlisten", 1)), arg[0], m.Meta["detail"]...)
			}
			return ""
			// }}}
		}},
		"dial": &ctx.Command{Name: "dial [address [security]]", Help: "建立连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				m.Travel(m.Target, func(m *ctx.Message) bool {
					if tcp, ok := m.Target.Server.(*TCP); ok && tcp.Conn != nil {
						m.Echo("%s %v<->%v\n", m.Target.Name, tcp.LocalAddr(), tcp.RemoteAddr())
					}
					return true
				})
			default:
				m.Start(fmt.Sprintf("com%d", m.Capi("nclient", 1)), arg[0], m.Meta["detail"]...)
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
			if tcp, ok := m.Target.Server.(*TCP); ok && tcp.Conn != nil { // {{{
				size, e := strconv.Atoi(arg[0])
				m.Assert(e)
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

func init() {
	tcp := &TCP{}
	tcp.Context = Index
	ctx.Index.Register(Index, tcp)
}
