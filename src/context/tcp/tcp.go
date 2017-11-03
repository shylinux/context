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

func (tcp *TCP) Begin() ctx.Server { // {{{
	return tcp
}

// }}}
func (tcp *TCP) Start(m *ctx.Message) bool { // {{{
	if tcp.Conf("address") == "" {
		return true
	}

	l, e := net.Listen("tcp", tcp.Conf("address"))
	tcp.Check(e)
	tcp.listener = l
	tcp.Capi("nlisten", 1)
	log.Println(tcp.Name, "listen:", l.Addr())

	for {
		c, e := l.Accept()
		log.Println(tcp.Name, "accept:", c.LocalAddr(), "<-", c.RemoteAddr())
		tcp.Check(e)

		m := m.Spawn(m.Context, c.RemoteAddr().String(), 0)
		m.Add("detail", "accept", c.RemoteAddr().String(), "tcp").Put("detail", c).Cmd()
	}

	return true
}

// }}}
func (tcp *TCP) Spawn(c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"nclient": &ctx.Cache{Name: "nclient", Value: "0", Help: "连接数量"},
	}
	c.Configs = map[string]*ctx.Config{
		"address": &ctx.Config{Name: "address", Value: arg[0], Help: "监听地址"},
	}

	s := new(TCP)
	s.Context = c
	return s

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
		"listen": &ctx.Command{"listen address", "监听端口", func(c *ctx.Context, m *ctx.Message, arg ...string) string {
			switch len(arg) { // {{{
			case 1:
				for k, s := range m.Target.Contexts {
					m.Echo("%s %s\n", k, s.Server.(*TCP).listener.Addr().String())
				}
			case 2:
				m.Start(arg[1:]...)
			}
			return ""
			// }}}
		}},
		"dial": &ctx.Command{"dial", "建立连接", func(c *ctx.Context, m *ctx.Message, arg ...string) string {
			tcp := c.Server.(*TCP) // {{{
			switch len(arg) {
			case 1:
				for i, v := range tcp.Resource {
					conn := v.Data["result"].(net.Conn)
					m.Echo(tcp.Name, "conn: %s %s -> %s\n", i, conn.LocalAddr(), conn.RemoteAddr())
				}
			case 2:
				conn, e := net.Dial("tcp", arg[1])
				c.Check(e)
				log.Println(tcp.Name, "dial:", conn.LocalAddr(), "->", conn.RemoteAddr())
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
