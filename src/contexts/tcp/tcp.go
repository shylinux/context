package tcp

import (
	"contexts/ctx"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type TCP struct {
	net.Conn
	net.Listener
	*ctx.Context
}

func (tcp *TCP) Read(b []byte) (n int, e error) {
	m := tcp.Context.Message()
	m.Assert(tcp.Conn != nil)
	n, e = tcp.Conn.Read(b)
	m.Capi("nrecv", n)
	if e != io.EOF {
		m.Assert(e)
	}
	return
}
func (tcp *TCP) Write(b []byte) (n int, e error) {
	m := tcp.Context.Message()
	m.Assert(tcp.Conn != nil)
	n, e = tcp.Conn.Write(b)
	m.Capi("nsend", n)
	if e != io.EOF {
		m.Assert(e)
	}
	return
}

func (tcp *TCP) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{
		"protocol": &ctx.Cache{Name: "protocol(tcp/tcp4/tcp6)", Value: "", Help: "网络协议"},
		"security": &ctx.Cache{Name: "security(true/false)", Value: "", Help: "加密通信"},
		"address":  &ctx.Cache{Name: "address", Value: "", Help: "网络地址"},
		"nrecv":    &ctx.Cache{Name: "nrecv", Value: "0", Help: "接收字节数"},
		"nsend":    &ctx.Cache{Name: "nsend", Value: "0", Help: "发送字节数"},
	}
	c.Configs = map[string]*ctx.Config{}

	s := new(TCP)
	s.Context = c
	return s
}
func (tcp *TCP) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return tcp
}
func (tcp *TCP) Start(m *ctx.Message, arg ...string) bool {
	if arg[1] == "consul" {
		arg[1] = m.Cmdx("web.get", "", arg[2], "temp", "hostport.0")
		if arg[1] == "" {
			return true
		}
		for i := 2; i < len(arg)-1; i++ {
			arg[i] = arg[i+1]
		}
		if len(arg) > 2 {
			arg = arg[:len(arg)-1]
		}
	}
	m.Cap("address", m.Confx("address", arg, 1))
	m.Cap("security", m.Confx("security", arg, 2))
	m.Cap("protocol", m.Confx("protocol", arg, 3))

	switch arg[0] {
	case "dial":
		if m.Caps("security") {
			m.Sess("aaa", m.Sess("aaa").Cmd("login", "cert", m.Cap("certfile"), "key", m.Cap("keyfile"), "tcp"))
			cert, e := tls.LoadX509KeyPair(m.Cap("certfile"), m.Cap("keyfile"))
			m.Assert(e)
			conf := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

			c, e := tls.Dial(m.Cap("protocol"), m.Cap("address"), conf)
			m.Assert(e)
			tcp.Conn = c
		} else {
			c, e := net.Dial(m.Cap("protocol"), m.Cap("address"))
			m.Assert(e)
			tcp.Conn = c
		}

		m.Log("info", "%s dial %s", m.Cap("nclient"),
			m.Cap("stream", fmt.Sprintf("%s->%s", tcp.LocalAddr(), tcp.RemoteAddr())))

		m.Sess("tcp", m)
		m.Option("ms_source", tcp.Context.Name)
		m.Put("option", "io", tcp).Back()
		return false

	case "accept":
		c, e := m.Optionv("io").(net.Conn)
		m.Assert(e)
		tcp.Conn = c

		m.Log("info", "%s accept %s", m.Cap("nclient"),
			m.Cap("stream", fmt.Sprintf("%s<-%s", tcp.LocalAddr(), tcp.RemoteAddr())))

		m.Sess("tcp", m)
		m.Option("ms_source", tcp.Context.Name)
		m.Put("option", "io", tcp).Back()
		return false

	case "listen":
		if m.Caps("security") {
			m.Sess("aaa", m.Sess("aaa").Cmd("login", "cert", m.Cap("certfile"), "key", m.Cap("keyfile"), "tcp"))
			cert, e := tls.LoadX509KeyPair(m.Cap("certfile"), m.Cap("keyfile"))
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

		m.Log("info", "%d listen %v", m.Capi("nlisten"),
			m.Cap("stream", fmt.Sprintf("%s", tcp.Addr())))

		addr := strings.Split(tcp.Addr().String(), ":")
		m.Log("fuck", "what %v", addr)
		m.Back(m.Spawn(m.Source()).Add("option", "hostport", fmt.Sprintf("%s:%s", m.Cmd("tcp.ifconfig", "eth0").Append("ip"), addr[len(addr)-1])))
	}

	for {
		c, e := tcp.Accept()
		m.Assert(e)
		m.Spawn(Index).Put("option", "io", c).Call(func(sub *ctx.Message) *ctx.Message {
			return sub.Spawn(m.Source())
		}, "accept", c.RemoteAddr().String(), m.Cap("security"), m.Cap("protocol"))
	}

	return true
}
func (tcp *TCP) Close(m *ctx.Message, arg ...string) bool {
	switch tcp.Context {
	case m.Target():
		if tcp.Listener != nil {
			m.Log("info", " close %v", m.Cap("stream"))
			tcp.Listener.Close()
			tcp.Listener = nil
		}
		if tcp.Conn != nil {
			m.Log("info", " close %v", m.Cap("stream"))
			tcp.Conn.Close()
			tcp.Conn = nil
		}
	case m.Source():
	}
	return true
}

var Index = &ctx.Context{Name: "tcp", Help: "网络中心",
	Caches: map[string]*ctx.Cache{
		"nlisten": &ctx.Cache{Name: "nlisten", Value: "0", Help: "监听数量"},
		"nclient": &ctx.Cache{Name: "nclient", Value: "0", Help: "连接数量"},
	},
	Configs: map[string]*ctx.Config{
		"":         &ctx.Config{Name: "address", Value: ":9090", Help: "网络地址"},
		"address":  &ctx.Config{Name: "address", Value: ":9090", Help: "网络地址"},
		"security": &ctx.Config{Name: "security(true/false)", Value: "false", Help: "加密通信"},
		"protocol": &ctx.Config{Name: "protocol(tcp/tcp4/tcp6)", Value: "tcp4", Help: "网络协议"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address [security [protocol]]", Help: "网络监听", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Start(fmt.Sprintf("pub%d", m.Capi("nlisten", 1)), "网络监听", m.Meta["detail"]...)
			return
		}},
		"accept": &ctx.Command{Name: "accept address [security [protocol]]", Help: "网络连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Start(fmt.Sprintf("com%d", m.Capi("nclient", 1)), "网络连接", m.Meta["detail"]...)
			return
		}},
		"dial": &ctx.Command{Name: "dial address [security [protocol]]", Help: "网络连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Start(fmt.Sprintf("com%d", m.Capi("nclient", 1)), "网络连接", m.Meta["detail"]...)
			return
		}},
		"send": &ctx.Command{Name: "send message", Help: "发送消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if tcp, ok := m.Target().Server.(*TCP); m.Assert(ok) {
				tcp.Write([]byte(arg[0]))
			}
			return
		}},
		"recv": &ctx.Command{Name: "recv size", Help: "接收消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if tcp, ok := m.Target().Server.(*TCP); m.Assert(ok) {
				n, e := strconv.Atoi(arg[0])
				m.Assert(e)
				buf := make([]byte, n)

				n, _ = tcp.Read(buf)
				m.Echo(string(buf[:n]))
			}
			return
		}},
		"ifconfig": &ctx.Command{Name: "ifconfig [name]", Help: "网络配置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if ifs, e := net.Interfaces(); m.Assert(e) {
				for _, v := range ifs {

					if ips, e := v.Addrs(); m.Assert(e) {
						for _, x := range ips {
							ip := strings.Split(x.String(), "/")

							if !strings.Contains(ip[0], ":") && len(ip) > 0 && len(v.HardwareAddr) > 0 {
								if len(arg) > 0 && !strings.Contains(v.Name, arg[0]) {
									continue
								}
								m.Add("append", "index", v.Index)
								m.Add("append", "name", v.Name)
								m.Add("append", "hard", v.HardwareAddr)
								m.Add("append", "ip", ip[0])
							}
						}
					}
				}
				m.Table()
			}
			return
		}},
	},
}

func init() {
	tcp := &TCP{}
	tcp.Context = Index
	ctx.Index.Register(Index, tcp)
}
