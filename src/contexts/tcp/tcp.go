package tcp

import (
	"contexts/ctx"
	"toolkit"

	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type TCP struct {
	net.Conn
	net.Listener
	*ctx.Context
}

func (tcp *TCP) parse(m *ctx.Message, arg ...string) ([]string, []string, bool) {
	defer func() {
		if e := recover(); e != nil {
			m.Log("warn", "%v", e)
		}
	}()

	address := []string{}
	if arg[1] == "dev" {
		m.Cmd("web.get", arg[1], arg[2], "temp", "ports", "format", "object", "temp_expire", "10").Table(func(line map[string]string) {
			address = append(address, line["value"])
		})
		if len(address) == 0 {
			return nil, nil, false
		}

		for i := 2; i < len(arg)-1; i++ {
			arg[i] = arg[i+1]
		}
		if len(arg) > 2 {
			arg = arg[:len(arg)-1]
		}
	} else {
		address = append(address, m.Cap("address", m.Confx("address", arg, 1)))
	}
	return address, arg, true
}
func (tcp *TCP) retry(m *ctx.Message, address []string, action func(address string) (net.Conn, error)) net.Conn {
	var count int32
	cs := make(chan net.Conn)

	for i := 0; i < m.Confi("retry", "counts"); i++ {
		for _, p := range address {
			m.Gos(m.Spawn().Add("option", "address", p), func(msg *ctx.Message) {
				m.Log("info", "dial: %v", msg.Option("address"))
				if count >= 1 {
					msg.Log("info", "skip: %v", msg.Option("address"))
				} else if c, e := action(msg.Option("address")); e != nil {
					msg.Log("info", "%s", e)
				} else if atomic.AddInt32(&count, 1) > 1 {
					msg.Log("info", "close: %s", c.LocalAddr())
					c.Close()
				} else {
					cs <- c
				}
			})
		}

		select {
		case c := <-cs:
			return c

		case <-time.After(kit.Duration(m.Conf("retry", "interval"))):
			m.Log("info", "dial %s:%v timeout", m.Cap("protocol"), address)
		}
	}
	return nil
}
func (tcp *TCP) Read(b []byte) (n int, e error) {
	if m := tcp.Context.Message(); m.Assert(tcp.Conn != nil) {
		if n, e = tcp.Conn.Read(b); e == io.EOF || m.Assert(e) {
			m.Capi("nrecv", n)
		}
	}
	return
}
func (tcp *TCP) Write(b []byte) (n int, e error) {
	if m := tcp.Context.Message(); m.Assert(tcp.Conn != nil) {
		if n, e = tcp.Conn.Write(b); e == io.EOF || m.Assert(e) {
			m.Capi("nsend", n)
		}
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
	return &TCP{Context: c}
}
func (tcp *TCP) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return tcp
}
func (tcp *TCP) Start(m *ctx.Message, arg ...string) bool {
	address, arg, ok := tcp.parse(m, arg...)
	if len(address) == 0 || !ok {
		return true
	}
	m.Cap("security", m.Confx("security", arg, 2))
	m.Cap("protocol", m.Confx("protocol", arg, 3))

	switch arg[0] {
	case "dial":
		if m.Caps("security") {
			cert, e := tls.LoadX509KeyPair(m.Cap("certfile"), m.Cap("keyfile"))
			m.Assert(e)
			conf := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

			tcp.Conn = tcp.retry(m, address, func(p string) (net.Conn, error) {
				return tls.Dial(m.Cap("protocol"), p, conf)
			})
		} else {
			tcp.Conn = tcp.retry(m, address, func(p string) (net.Conn, error) {
				return net.DialTimeout(m.Cap("protocol"), p, kit.Duration(m.Conf("retry", "timeout")))
			})
		}

		m.Log("info", "%s connect %s", m.Cap("nclient"),
			m.Cap("stream", fmt.Sprintf("%s->%s", tcp.LocalAddr(), m.Cap("address", tcp.RemoteAddr().String()))))

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

		m.Log("info", "%d listen %v", m.Capi("nlisten"), m.Cap("stream", fmt.Sprintf("%s", tcp.Addr())))

		addr := strings.Split(tcp.Addr().String(), ":")
		ports := []interface{}{}
		if m.Cmd("tcp.ifconfig").Table(func(line map[string]string) {
			ports = append(ports, fmt.Sprintf("%s:%s", line["ip"], addr[len(addr)-1]))
		}); len(ports) == 0 {
			ports = append(ports, fmt.Sprintf("%s:%s", "127.0.0.1", addr[len(addr)-1]))
		}
		m.Back(m.Spawn(m.Source()).Put("option", "node.port", ports))
	default:
		return true
	}

	for {
		if c, e := tcp.Accept(); m.Assert(e) {
			m.Spawn(Index).Put("option", "io", c).Call(func(sub *ctx.Message) *ctx.Message {
				return sub.Spawn(m.Source())
			}, "accept", c.RemoteAddr().String(), m.Cap("security"), m.Cap("protocol"))
		}
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
		"address":  &ctx.Config{Name: "address", Value: ":9090", Help: "网络地址"},
		"security": &ctx.Config{Name: "security(true/false)", Value: "false", Help: "加密通信"},
		"protocol": &ctx.Config{Name: "protocol(tcp/tcp4/tcp6)", Value: "tcp4", Help: "网络协议"},
		"retry": &ctx.Config{Name: "retry", Value: map[string]interface{}{
			"interval": "3s", "counts": 3, "timeout": "10s",
		}, Help: "网络重试"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address [security [protocol]]", Help: "网络监听", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Start(fmt.Sprintf("pub%d", m.Capi("nlisten", 1)), "网络监听", m.Meta["detail"]...)
			return
		}},
		"accept": &ctx.Command{Name: "accept address [security [protocol]]", Help: "网络连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Start(fmt.Sprintf("sub%d", m.Capi("nclient", 1)), "网络连接", m.Meta["detail"]...)
			return
		}},
		"dial": &ctx.Command{Name: "dial address [security [protocol]]", Help: "网络连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Start(fmt.Sprintf("com%d", m.Capi("nclient", 1)), "网络连接", m.Meta["detail"]...)
			return
		}},

		"send": &ctx.Command{Name: "send message", Help: "发送消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if tcp, ok := m.Target().Server.(*TCP); m.Assert(ok) {
				fmt.Fprint(tcp, arg[0])
			}
			return
		}},
		"recv": &ctx.Command{Name: "recv size", Help: "接收消息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if tcp, ok := m.Target().Server.(*TCP); m.Assert(ok) {
				if n, e := strconv.Atoi(arg[0]); m.Assert(e) {
					buf := make([]byte, n)
					n, _ = tcp.Read(buf)
					m.Echo(string(buf[:n]))
				}
			}
			return
		}},

		"ifconfig": &ctx.Command{Name: "ifconfig [name]", Help: "网络配置", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if ifs, e := net.Interfaces(); m.Assert(e) {
				for _, v := range ifs {
					if len(arg) > 0 && !strings.Contains(v.Name, arg[0]) {
						continue
					}
					if ips, e := v.Addrs(); m.Assert(e) {
						for _, x := range ips {
							ip := strings.Split(x.String(), "/")
							if strings.Contains(ip[0], ":") || len(ip) == 0 {
								continue
							}
							if len(v.HardwareAddr.String()) == 0 {
								continue
							}

							m.Push("index", v.Index)
							m.Push("name", v.Name)
							m.Push("ip", ip[0])
							m.Push("mask", ip[1])
							m.Push("hard", v.HardwareAddr.String())
						}
					}
				}
				m.Table()
			}
			return
		}},
		"probe": &ctx.Command{Name: "probe [port]", Help: "端口检测", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				for i := 0; i < 1024; i++ {
					m.Show("port: %v", i)
					if t, e := net.DialTimeout("tcp", fmt.Sprintf(":%d", i), 3*time.Second); e == nil {
						m.Push("port", i)
						t.Close()
					}
				}
				m.Table()
				return
			}
			if t, e := net.DialTimeout("tcp", arg[0], 10*time.Second); e == nil {
				m.Echo("active")
				t.Close()
			}
			return
		}},
	},
}

func init() {
	ctx.Index.Register(Index, &TCP{Context: Index})
}
