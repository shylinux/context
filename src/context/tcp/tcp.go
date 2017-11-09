package tcp

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

func (tcp *TCP) Begin(m *ctx.Message) ctx.Server { // {{{
	return tcp
}

// }}}
func (tcp *TCP) Start(m *ctx.Message) bool { // {{{
	if tcp.Conf("address") == "" {
		return true
	}

	l, e := net.Listen("tcp4", tcp.Conf("address"))
	tcp.Assert(e)
	tcp.listener = l

	log.Printf("%s listen(%d): %v", tcp.Name, tcp.Capi("nlisten", 1), l.Addr())
	defer tcp.Capi("nlisten", -1)
	defer log.Println("%s close(%d): %v", tcp.Name, tcp.Capi("nlisten", 0), l.Addr())

	for {
		c, e := l.Accept()
		tcp.Assert(e)
		log.Printf("%s accept(%d): %v<-%v", tcp.Name, tcp.Capi("nclient", 1), c.LocalAddr(), c.RemoteAddr())
		// defer log.Println(tcp.Name, "close:", tcp.Capi("nclient", -1), c.LocalAddr(), "<-", c.RemoteAddr())

		msg := m.Spawn(m.Context, c.RemoteAddr().String()).Put("option", "io", c)
		msg.Cmd("accept", c.RemoteAddr().String(), "tcp")
	}

	return true
}

// }}}
func (tcp *TCP) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"nclient": &ctx.Cache{Name: "nclient", Value: "0", Help: "连接数量"},
		"status":  &ctx.Cache{Name: "status", Value: "stop", Help: "服务状态"},
	}
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

	if c, ok := m.Data["detail"].(net.Conn); ok && m.Context == tcp.Context {
		c.Close()
		delete(m.Data, "detail")
		return true
	}
	return true
}

// }}}

var Index = &ctx.Context{Name: "tcp", Help: "网络连接",
	Caches: map[string]*ctx.Cache{
		"nclient": &ctx.Cache{Name: "nclient", Value: "0", Help: "连接数量"},
		"nlisten": &ctx.Cache{Name: "nlisten", Value: "0", Help: "连接数量"},
	},
	Configs: map[string]*ctx.Config{
		"address": &ctx.Config{Name: "address", Value: "", Help: "监听地址"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address", Help: "监听端口", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				for k, s := range m.Target.Contexts {
					m.Echo("%s %s\n", k, s.Server.(*TCP).listener.Addr().String())
				}
			case 1:
				m.Start(arg...)
			}
			return ""
			// }}}
		}},
		"dial": &ctx.Command{Name: "dial", Help: "建立连接", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			tcp := c.Server.(*TCP) // {{{
			switch len(arg) {
			case 0:
				for i, v := range tcp.Resource {
					conn := v.Data["result"].(net.Conn)
					m.Echo(tcp.Name, "conn: %s %s -> %s\n", i, conn.LocalAddr(), conn.RemoteAddr())
				}
			case 2:
				conn, e := net.Dial("tcp", arg[0])
				c.Assert(e)
				log.Println(tcp.Name, "dial:", conn.LocalAddr(), "->", conn.RemoteAddr())
			}
			return ""
			// }}}
		}},
		"exit": &ctx.Command{Name: "exit", Help: "退出", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			tcp, ok := m.Target.Server.(*TCP) // {{{
			if !ok {
				tcp, ok = m.Context.Server.(*TCP)
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
