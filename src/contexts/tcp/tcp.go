package tcp // {{{
// }}}
import ( // {{{
	"contexts"

	"crypto/tls"
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
		"protocol": &ctx.Cache{Name: "网络协议(tcp/tcp4/tcp6)", Value: m.Conf("protocol"), Help: "网络协议"},
		"security": &ctx.Cache{Name: "加密通信(true/false)", Value: m.Conf("security"), Help: "加密通信"},
		"address":  &ctx.Cache{Name: "网络地址", Value: "", Help: "网络地址"},
	}
	c.Configs = map[string]*ctx.Config{}

	s := new(TCP)
	s.Context = c
	return s
}

// }}}
func (tcp *TCP) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	tcp.Context.Master(nil)
	if tcp.Context == Index {
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
	if len(arg) > 3 {
		m.Cap("protocol", arg[3])
	}

	switch arg[0] {
	case "dial":
		if m.Caps("security") {
			cert, e := tls.LoadX509KeyPair(m.Conf("cert"), m.Conf("key"))
			m.Assert(e)
			conf := &tls.Config{Certificates: []tls.Certificate{cert}}

			c, e := tls.Dial(m.Cap("protocol"), m.Cap("address"), conf)
			m.Assert(e)
			tcp.Conn = c
		} else {
			c, e := net.Dial(m.Cap("protocol"), m.Cap("address"))
			m.Assert(e)
			tcp.Conn = c
		}

		m.Log("info", nil, "%s dial %s", Pulse.Cap("nclient"),
			m.Append("stream", m.Cap("stream", fmt.Sprintf("%s->%s", tcp.LocalAddr(), tcp.RemoteAddr()))))
		m.Put("append", "io", tcp.Conn).Back(m)
		return false
	case "accept":
		c, e := m.Data["io"].(net.Conn)
		m.Assert(e)
		tcp.Conn = c

		m.Log("info", nil, "%s accept %s", Pulse.Cap("nclient"),
			m.Append("stream", m.Cap("stream", fmt.Sprintf("%s<-%s", tcp.LocalAddr(), tcp.RemoteAddr()))))
		m.Put("append", "io", tcp.Conn).Back(m)
		return false
	default:
		if m.Cap("security") != "false" {
			cert, e := tls.LoadX509KeyPair(m.Conf("cert"), m.Conf("key"))
			m.Assert(e)
			conf := &tls.Config{Certificates: []tls.Certificate{cert}}

			l, e := tls.Listen(m.Cap("protocol"), m.Cap("address"), conf)
			m.Assert(e)
			tcp.Listener = l
		} else {
			l, e := net.Listen(m.Cap("protocol"), m.Cap("address"))
			m.Assert(e)
			tcp.Listener = l
		}

		m.Log("info", nil, "%d listen %v", Pulse.Capi("nlisten"), m.Cap("stream", fmt.Sprintf("%s", tcp.Addr())))
	}

	for {
		c, e := tcp.Accept()
		m.Assert(e)
		msg := m.Spawn(Index).Put("option", "io", c).Put("option", "source", m.Source())
		msg.Call(func(com *ctx.Message) *ctx.Message {
			return com
		}, "accept", c.RemoteAddr().String(), m.Cap("security"), m.Cap("protocol"))
	}

	return true
}

// }}}
func (tcp *TCP) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch tcp.Context {
	case m.Target():
		if tcp.Listener != nil {
			m.Log("info", nil, "%d close %v", Pulse.Capi("nlisten", -1)+1, m.Cap("stream"))
			tcp.Listener.Close()
			tcp.Listener = nil
		}
		if tcp.Conn != nil {
			m.Log("info", nil, "%d close %v", Pulse.Capi("nclient", -1)+1, m.Cap("stream"))
			tcp.Conn.Close()
			tcp.Conn = nil
		}
	case m.Source():
		if tcp.Conn != nil {
			msg := m.Spawn(tcp.Context)
			if msg.Master(tcp.Context); !tcp.Context.Close(msg, arg...) {
				return false
			}
		}
	}

	return true
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "tcp", Help: "网络中心",
	Caches: map[string]*ctx.Cache{
		"nlisten": &ctx.Cache{Name: "监听数量", Value: "0", Help: "监听数量"},
		"nclient": &ctx.Cache{Name: "连接数量", Value: "0", Help: "连接数量"},
	},
	Configs: map[string]*ctx.Config{
		"security": &ctx.Config{Name: "加密通信(true/false)", Value: "false", Help: "加密通信"},
		"protocol": &ctx.Config{Name: "网络协议(tcp/tcp4/tcp6)", Value: "tcp4", Help: "网络协议"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address [security [protocol]]", Help: "网络监听",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				m.Start(fmt.Sprintf("pub%d", Pulse.Capi("nlisten", 1)), "网络监听", m.Meta["detail"]...)
			}},
		"accept": &ctx.Command{Name: "accept address [security [protocol]]", Help: "网络连接",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				m.Start(fmt.Sprintf("com%d", Pulse.Capi("nclient", 1)), "网络连接", m.Meta["detail"]...)
			}},
		"dial": &ctx.Command{Name: "dial address [security [protocol]]", Help: "网络连接",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				m.Start(fmt.Sprintf("com%d", Pulse.Capi("nclient", 1)), "网络连接", m.Meta["detail"]...)
			}},
		"send": &ctx.Command{Name: "send message", Help: "发送消息",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				tcp, ok := m.Target().Server.(*TCP)
				m.Assert(ok && tcp.Conn != nil)
				tcp.Conn.Write([]byte(arg[0]))
			}},
		"recv": &ctx.Command{Name: "recv size", Help: "接收消息",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				tcp, ok := m.Target().Server.(*TCP)
				m.Assert(ok && tcp.Conn != nil)
				size, e := strconv.Atoi(arg[0])
				m.Assert(e)

				buf := make([]byte, size)
				tcp.Conn.Read(buf)
				m.Echo(string(buf))
			}},
		"ifconfig": &ctx.Command{Name: "ifconfig", Help: "接收消息",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				ifs, e := net.Interfaces()
				m.Assert(e)
				for _, v := range ifs {
					ips, e := v.Addrs()
					m.Assert(e)
					for _, x := range ips {
						m.Echo("%d %s %v %v\n", v.Index, v.Name, v.HardwareAddr, x.String())
					}
				}
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
