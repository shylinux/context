package tcp // {{{
// }}}
import ( // {{{
	"context"
	"log"
	"net"
)

// }}}

type TCP struct {
	l     net.Listener
	c     net.Conn
	close bool
	*ctx.Context
}

func (tcp *TCP) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	return tcp
}

// }}}
func (tcp *TCP) Start(m *ctx.Message, arg ...string) bool { // {{{
	if arg[0] == "dial" {
		c, e := net.Dial(m.Conf("protocol"), m.Conf("address"))
		m.Assert(e)
		tcp.c = c

		log.Printf("%s dial(%d): %v->%v", tcp.Name, m.Capi("nclient", 1), c.LocalAddr(), c.RemoteAddr())
		m.Reply(c.LocalAddr().String()).Put("option", "io", c).Cmd("open")
		return true
	}

	l, e := net.Listen(m.Conf("protocol"), m.Conf("address"))
	m.Assert(e)
	tcp.l = l

	log.Printf("%s listen(%d): %v", tcp.Name, m.Capi("nlisten", 1), l.Addr())
	defer m.Capi("nlisten", -1)
	defer log.Println("%s close(%d): %v", tcp.Name, m.Capi("nlisten"), l.Addr())

	for {
		c, e := l.Accept()
		m.Assert(e)
		log.Printf("%s accept(%d): %v<-%v", tcp.Name, m.Capi("nclient", 1), c.LocalAddr(), c.RemoteAddr())
		m.Reply(c.RemoteAddr().String()).Put("option", "io", c).Cmd("open")
	}

	return true
}

// }}}
func (tcp *TCP) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	if len(arg) > 1 {
		switch arg[0] {
		case "listen":
			c.Caches["nclient"] = &ctx.Cache{Name: "nclient", Value: "0", Help: "连接数量"}
			c.Configs["address"] = &ctx.Config{Name: "address", Value: arg[1], Help: "监听地址"}
		case "dial":
			c.Configs["address"] = &ctx.Config{Name: "address", Value: arg[1], Help: "连接地址"}
		}
	}
	if len(arg) > 2 {
		c.Configs["security"] = &ctx.Config{Name: "security(true/false)", Value: "true", Help: "加密通信"}
	}

	s := new(TCP)
	s.Context = c
	return s

}

// }}}
func (tcp *TCP) Exit(m *ctx.Message, arg ...string) bool { // {{{
	switch tcp.Context {
	case m.Source:
		c, ok := m.Data["io"].(net.Conn)
		if !ok {
			c = tcp.c
		}
		if c != nil {
			log.Println(tcp.Name, "close:", c.LocalAddr(), "--", c.RemoteAddr())
			c.Close()
		}

	case m.Target:
		if tcp.l != nil {
			log.Println(tcp.Name, "close:", tcp.l.Addr())
			tcp.l.Close()
		}
		if tcp.c != nil {
			log.Println(tcp.Name, "close:", tcp.c.LocalAddr(), "->", tcp.c.RemoteAddr())
			tcp.c.Close()
		}
	}

	return true
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
		"listen": &ctx.Command{Name: "listen [address [security]]", Help: "监听连接", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				m.Travel(m.Target, func(m *ctx.Message) bool {
					if tcp, ok := m.Target.Server.(*TCP); ok && tcp.l != nil {
						m.Echo("%s %v\n", m.Target.Name, tcp.l.Addr())
					}
					return true
				})
			case 1:
				go m.Start(arg[0], m.Meta["detail"]...)
			}
			return ""
			// }}}
		}},
		"dial": &ctx.Command{Name: "dial [address [security]]", Help: "建立连接", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				m.Travel(m.Target, func(m *ctx.Message) bool {
					if tcp, ok := m.Target.Server.(*TCP); ok && tcp.c != nil {
						m.Echo("%s %v->%v\n", m.Target.Name, tcp.c.LocalAddr(), tcp.c.RemoteAddr())
					}
					return true
				})
			case 1:
				m.Start(arg[0], m.Meta["detail"]...)
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
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{
			Commands: map[string]*ctx.Command{
				"listen": &ctx.Command{},
			},
		},
	},
}

func init() {
	tcp := &TCP{}
	tcp.Context = Index
	ctx.Index.Register(Index, tcp)
}
