package tcp

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

type TCP struct {
	listener net.Listener
	conn     net.Conn

	target *ctx.Context
	*ctx.Context
}

func (tcp *TCP) Begin() bool {
	return true
}

func (tcp *TCP) Start() bool {
	log.Println(tcp.Name, "start:")
	if tcp.Conf("address") == "" {
		log.Println(tcp.Name, "start:")
		for {
			msg := tcp.Get()
			arg := msg.Meta["detail"]
			switch arg[0] {
			case "listen":
				s := tcp.Context.Spawn(arg[1])
				s.Conf("addresss", "address", arg[1], "监听地址")
				go s.Start()
			}
			if msg.Wait != nil {
				msg.Wait <- true
			}
			msg.End(true)
		}
		return true
	} else {
		if tcp.Conf("remote") == "" {
			l, e := net.Listen("tcp", tcp.Conf("address"))
			tcp.Check(e)
			tcp.listener = l

			for {
				c, e := l.Accept()
				tcp.Check(e)

				s := tcp.Context.Spawn(c.RemoteAddr().String())
				s.Conf("remote", "remote", c.RemoteAddr().String(), "连接地址")
				x := s.Server.(*TCP)
				x.conn = c
				go s.Start()
			}
			return true
		} else {
			for {
				fmt.Fprintln(tcp.conn, "hello context world!\n")
				time.Sleep(3 * time.Second)
			}
		}

	}
}

func (tcp *TCP) Spawn(c *ctx.Context, key string) ctx.Server {
	s := new(TCP)
	s.Context = c
	return s
}

func (tcp *TCP) Fork(c *ctx.Context, key string) ctx.Server {
	s := new(TCP)
	s.Context = c
	return s
}

var Index = &ctx.Context{Name: "tcp", Help: "网络连接",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"address": &ctx.Config{Name: "address", Value: "", Help: "监听地址"},
		"remote":  &ctx.Config{Name: "remote", Value: "", Help: "远程地址"},
	},
	Commands: map[string]*ctx.Command{},
}

func init() {
	tcp := &TCP{}
	tcp.Context = Index
	ctx.Index.Register(Index, tcp)
}
