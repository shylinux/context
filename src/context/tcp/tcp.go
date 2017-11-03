package tcp

import (
	"context"
	"log"
	"net"
	"time"
)

type TCP struct {
	listener net.Listener
	*ctx.Context
	*ctx.CTX
}

func (tcp *TCP) Begin() bool {
	return true
}

func (tcp *TCP) Start() bool {
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

		msg := &ctx.Message{Code: tcp.Root.Capi("nmessage", 1), Time: time.Now()}
		msg.Context = tcp.Resource[0].Context
		msg.Index = tcp.Capi("nclient", 1)
		msg.Target = tcp.Context
		msg.Put("result", c)
		msg.Context.Cmd(msg, "accept", c.RemoteAddr().String(), "tcp")
		tcp.Resource = append(tcp.Resource, msg)
	}
	return true
}

func (tcp *TCP) Spawn(c *ctx.Context, arg ...string) ctx.Server {
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

var Index = &ctx.Context{Name: "tcp", Help: "网络连接",
	Caches: map[string]*ctx.Cache{
		"nclient": &ctx.Cache{Name: "nclient", Value: "0", Help: "连接数量"},
		"nlisten": &ctx.Cache{Name: "nlisten", Value: "0", Help: "连接数量"},
	},
	Configs: map[string]*ctx.Config{
		"address": &ctx.Config{Name: "address", Value: "", Help: "监听地址"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{"listen", "监听端口", func(c *ctx.Context, m *ctx.Message, arg ...string) string {
			switch len(arg) {
			case 1:
				for k, s := range c.Contexts {
					x := s.Server.(*TCP)
					m.Echo("%s %s\n", k, x.listener.Addr().String())
				}
			case 2:
				s := c.Spawn(arg[1:]...)
				s.Resource = make([]*ctx.Message, 0, 3)
				s.Resource = append(s.Resource, m)
				m.Target = s
				m.Index = 0
				go s.Start()
			}
			return ""
		}},
		"dial": &ctx.Command{"dial", "建立连接", func(c *ctx.Context, m *ctx.Message, arg ...string) string {
			tcp := c.Server.(*TCP)
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
		}},
	},
}

func init() {
	tcp := &TCP{}
	tcp.CTX = ctx.Ctx
	tcp.Context = Index
	ctx.Index.Register(Index, tcp)
}
