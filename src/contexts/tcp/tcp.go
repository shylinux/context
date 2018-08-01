package tcp // {{{
// }}}
import ( // {{{
	"contexts"

	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
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
		"nrecv":    &ctx.Cache{Name: "nrecv", Value: "0", Help: "网络地址"},
		"nsend":    &ctx.Cache{Name: "nsend", Value: "0", Help: "网络地址"},
	}
	c.Configs = map[string]*ctx.Config{}

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
	m.Cap("address", m.Confx("address", arg, 1))
	m.Cap("security", m.Confx("security", arg, 2))
	m.Cap("protocol", m.Confx("protocol", arg, 3))

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

		m.Log("info", "%s dial %s", m.Cap("nclient"),
			m.Option("stream", m.Cap("stream", fmt.Sprintf("%s->%s", tcp.LocalAddr(), tcp.RemoteAddr()))))
		m.Put("option", "io", tcp.Conn).Back(m)
		return false
	case "accept":
		c, e := m.Data["io"].(net.Conn)
		m.Assert(e)
		tcp.Conn = c

		m.Log("info", "%s accept %s", m.Cap("nclient"),
			m.Option("stream", m.Cap("stream", fmt.Sprintf("%s<-%s", tcp.LocalAddr(), tcp.RemoteAddr()))))
		m.Put("option", "io", tcp.Conn).Back(m)
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

		m.Log("info", "%d listen %v", m.Capi("nlisten"), m.Cap("stream", fmt.Sprintf("%s", tcp.Addr())))
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
	return false
	switch tcp.Context {
	case m.Target():
		if tcp.Listener != nil {
			m.Log("info", "%d close %v", m.Capi("nlisten", -1)+1, m.Cap("stream"))
			tcp.Listener.Close()
			tcp.Listener = nil
		}
		if tcp.Conn != nil {
			m.Log("info", "%d close %v", m.Capi("nclient", -1)+1, m.Cap("stream"))
			tcp.Conn.Close()
			tcp.Conn = nil
		}
	case m.Source():
		if tcp.Conn != nil {
			msg := m.Spawn(tcp.Context)
			if !tcp.Context.Close(msg, arg...) {
				return false
			}
		}
	}
	if m.Target() == Index {
		return false
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
		"address":  &ctx.Config{Name: "address", Value: ":9090", Help: "加密通信"},
		"security": &ctx.Config{Name: "security(true/false)", Value: "false", Help: "加密通信"},
		"protocol": &ctx.Config{Name: "protocol(tcp/tcp4/tcp6)", Value: "tcp4", Help: "网络协议"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address [security [protocol]]", Help: "网络监听", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Start(fmt.Sprintf("pub%d", m.Capi("nlisten", 1)), "网络监听", m.Meta["detail"]...)
		}},
		"accept": &ctx.Command{Name: "accept address [security [protocol]]", Help: "网络连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Start(fmt.Sprintf("com%d", m.Capi("nclient", 1)), "网络连接", m.Meta["detail"]...)
		}},
		"dial": &ctx.Command{Name: "dial address [security [protocol]]", Help: "网络连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Start(fmt.Sprintf("com%d", m.Capi("nclient", 1)), "网络连接", m.Meta["detail"]...)
		}},
		"send": &ctx.Command{Name: "send message", Help: "发送消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if tcp, ok := m.Target().Server.(*TCP); m.Assert(ok) && tcp.Conn != nil { // {{{
				tcp.Conn.Write([]byte(arg[0]))
				m.Capi("nsend", len(arg[0]))
			} // }}}
		}},
		"recv": &ctx.Command{Name: "recv size", Help: "接收消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if tcp, ok := m.Target().Server.(*TCP); m.Assert(ok) && tcp.Conn != nil { // {{{
				size, e := strconv.Atoi(arg[0])
				m.Assert(e)

				buf := make([]byte, size)
				n, e := tcp.Conn.Read(buf)
				m.Assert(e)
				buf = buf[:n]

				m.Echo(string(buf))
				m.Capi("nrecv", n)
			} // }}}
		}},
		"ifconfig": &ctx.Command{Name: "ifconfig", Help: "网络配置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			ifs, e := net.Interfaces() // {{{
			m.Assert(e)
			for _, v := range ifs {
				ips, e := v.Addrs()
				m.Assert(e)
				for _, x := range ips {
					ip := x.String()
					if !strings.Contains(ip, ":") && len(ip) > 0 && len(v.HardwareAddr) > 0 {
						m.Add("append", "index", v.Index)
						m.Add("append", "name", v.Name)
						m.Add("append", "hard", v.HardwareAddr)
						m.Add("append", "ip", ip)
					}
				}
			}
			m.Table()
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
