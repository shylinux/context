package tcp // {{{
// }}}
import ( // {{{
	"context"
	"log"
	"net"
)

// }}}

type TCP struct {
	listener net.Listener
	*ctx.Context
}

func (tcp *TCP) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	tcp.Caches["nclient"] = &ctx.Cache{Name: "nclient", Value: "0", Help: "连接数量"}
	return tcp
}

// }}}
func (tcp *TCP) Start(m *ctx.Message, arg ...string) bool { // {{{
	if m.Conf("address") == "" {
		return true
	}

	l, e := net.Listen("tcp4", m.Conf("address"))
	m.Assert(e)
	tcp.listener = l

	log.Printf("%s listen(%d): %v", tcp.Name, m.Capi("nlisten", 1), l.Addr())
	defer m.Capi("nlisten", -1)
	defer log.Println("%s close(%d): %v", tcp.Name, m.Capi("nlisten", 0), l.Addr())

	for {
		c, e := l.Accept()
		m.Assert(e)
		log.Printf("%s accept(%d): %v<-%v", tcp.Name, m.Capi("nclient", 1), c.LocalAddr(), c.RemoteAddr())
		// defer log.Println(tcp.Name, "close:", m.Capi("nclient", -1), c.LocalAddr(), "<-", c.RemoteAddr())

		msg := m.Spawn(m.Source, c.RemoteAddr().String()).Put("option", "io", c)
		msg.Cmd("open", c.RemoteAddr().String(), "tcp")
	}

	return true
}

// }}}
func (tcp *TCP) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{
		"address": &ctx.Config{Name: "address", Value: arg[0], Help: "监听地址"},
	}

	s := new(TCP)
	s.Context = c
	return s

}

// }}}
func (tcp *TCP) Exit(m *ctx.Message, arg ...string) bool { // {{{

	if c, ok := m.Data["result"].(net.Conn); ok && m.Target == tcp.Context {
		c.Close()
		delete(m.Data, "result")
		return true
	}

	if c, ok := m.Data["detail"].(net.Conn); ok && m.Source == tcp.Context {
		c.Close()
		delete(m.Data, "detail")
		return true
	}
	return true
}

// }}}

var Index = &ctx.Context{Name: "tcp", Help: "网络连接",
	Caches: map[string]*ctx.Cache{
		"nlisten": &ctx.Cache{Name: "nlisten", Value: "0", Help: "连接数量"},
	},
	Configs: map[string]*ctx.Config{
		"address": &ctx.Config{Name: "address", Value: "", Help: "监听地址"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address", Help: "监听连接", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				m.Target.Travel(func(c *ctx.Context) bool {
					m.Echo("%s %s\n", c.Name, c.Server.(*TCP).listener.Addr().String())
					return true
				})
			case 1:
				go m.Start(arg[0], arg[0])
			}
			return ""
			// }}}
		}},
		"dial": &ctx.Command{Name: "dial", Help: "建立连接", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			tcp := c.Server.(*TCP) // {{{
			switch len(arg) {
			case 0:
				for i, v := range tcp.Requests {
					conn := v.Data["result"].(net.Conn)
					m.Echo(tcp.Name, "conn: %s %s -> %s\n", i, conn.LocalAddr(), conn.RemoteAddr())
				}
			case 2:
				conn, e := net.Dial("tcp", arg[0])
				m.Assert(e)
				log.Println(tcp.Name, "dial:", conn.LocalAddr(), "->", conn.RemoteAddr())
			}
			return ""
			// }}}
		}},
		"exit": &ctx.Command{Name: "exit", Help: "退出", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			tcp, ok := m.Target.Server.(*TCP) // {{{
			if !ok {
				tcp, ok = m.Source.Server.(*TCP)
			}
			if ok {
				tcp.Context.Exit(m)
			}

			return ""
			// }}}
		}},
	},
}

func init() {
	tcp := &TCP{}
	tcp.Context = Index
	ctx.Index.Register(Index, tcp)
}
