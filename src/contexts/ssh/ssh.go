package ssh // {{{
// }}}
import ( // {{{
	"bufio"
	"contexts"
	"fmt"
	"net"
	"strings"
)

// }}}

type SSH struct {
	send map[string]*ctx.Message
	*bufio.Writer
	*bufio.Reader
	net.Conn

	*ctx.Context
}

func (ssh *SSH) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(SSH)
	s.Context = c
	return s
}

// }}}
func (ssh *SSH) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	ssh.Context.Master(nil)
	if ssh.Context == Index {
		Pulse = m
	}
	return ssh
}

// }}}
func (ssh *SSH) Start(m *ctx.Message, arg ...string) bool { // {{{
	return false
}

// }}}
func (ssh *SSH) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch ssh.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

// }}}

func Done(m *ctx.Message, lock chan bool) {
	m.Log("lock", nil, "done before %v", m.Meta["detail"])
	if m.Options("stdio") {
		lock <- true
	}
	m.Log("lock", nil, "done after %v", m.Meta["detail"])
}

func Wait(m *ctx.Message, lock chan bool) {
	m.Log("lock", nil, "wait before %v", m.Meta["detail"])
	if m.Options("stdio") {
		<-lock
	}
	m.Log("lock", nil, "wait after %v", m.Meta["detail"])
}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches: map[string]*ctx.Cache{
		"nhost": &ctx.Cache{Name: "主机数量", Value: "0", Help: "主机数量"},
		"route": &ctx.Cache{Name: "route", Value: "com", Help: "主机数量"},
		"count": &ctx.Cache{Name: "count", Value: "3", Help: "主机数量"},
		"share": &ctx.Cache{Name: "share", Value: "root", Help: "主机数量"},
		"level": &ctx.Cache{Name: "level", Value: "root", Help: "主机数量"},
	},
	Configs: map[string]*ctx.Config{
		"route":      &ctx.Config{Name: "route", Value: "com", Help: "主机数量"},
		"route.json": &ctx.Config{Name: "route.json", Value: "var/route.json", Help: "主机数量"},
		"route.png":  &ctx.Config{Name: "route.png", Value: "var/route.png", Help: "主机数量"},
		"type":       &ctx.Config{Name: "type", Value: "terminal", Help: "主机数量"},
		"kind":       &ctx.Config{Name: "kind", Value: "terminal", Help: "主机数量"},
		"name":       &ctx.Config{Name: "name", Value: "vps", Help: "主机数量"},
		"mark":       &ctx.Config{Name: "mark", Value: "com", Help: "主机数量"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address protocol", Help: "监听连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Find("nfs") // {{{
			msg.Call(func(file *ctx.Message) *ctx.Message {
				sub := file.Spawn(m.Target())
				sub.Start(fmt.Sprintf("host%d", Pulse.Capi("nhost", 1)), "远程主机")
				sub.Cap("stream", file.Target().Name)

				sub.Sess("file", "nfs."+file.Target().Name)
				sub.Sess("file").Cmd("send", "route", sub.Target().Name, msg.Cap("route"))
				return sub
			}, m.Meta["detail"])
			// }}}
		}},
		"dial": &ctx.Command{Name: "dial address protocol", Help: "建立连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Sess("file", "nfs") // {{{
			msg.Call(func(file *ctx.Message) *ctx.Message {
				m.Cap("stream", file.Target().Name)
				return m
			}, m.Meta["detail"])
			// }}}
		}},
		"route": &ctx.Command{Name: "route", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Conf("route", arg[0]) // {{{
			m.Cap("route", arg[1]+"."+arg[0])
			info := map[string]string{"route": m.Cap("route")}

			msg := m.Sess("file")
			msg.Put("option", "data", info)
			msg.Cmd("json", m.Conf("route.json"))

			png := m.Sess("file")
			png.Cmd("genqr", m.Conf("route.png"), msg.Result(0))

			m.Back(m)
			// }}}
		}},
		"pwd": &ctx.Command{Name: "pwd", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Echo(m.Cap("route")) // {{{
			m.Back(m)
			// }}}
		}},
		"send": &ctx.Command{Name: "send route cmd arg...", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*SSH); m.Assert(ok) { // {{{
				lock := make(chan bool)
				if len(arg[0]) == 0 {
					msg := m.Spawn(m.Target()).Cmd(arg[1:])
					m.Copy(msg, "result")
					m.Copy(msg, "append")
					m.Back(m)
					return
				}

				miss := true
				self := true
				target := strings.Split(arg[0], ".")
				m.Travel(m.Target(), func(m *ctx.Message) bool {
					if self {
						self = false
						return true
					}

					if m.Target().Name == target[0] {
						msg := m.Sess("file")
						msg.Call(func(host *ctx.Message) *ctx.Message {
							m.Copy(host, "result")
							m.Copy(host, "append")
							Done(m, lock)
							return m
						}, "send", "send", strings.Join(target[1:], "."), arg[1:])

						miss = false
					}
					return miss
				})

				if miss {
					if target[0] == m.Conf("route") {
						m.Spawn(m.Target()).Call(func(host *ctx.Message) *ctx.Message {
							m.Copy(host, "result")
							m.Copy(host, "append")
							Done(m, lock)
							return m
						}, "send", strings.Join(target[1:], "."), arg[1:])
					} else if m.Cap("route") != "ssh" {
						msg := m.Sess("file")
						msg.Call(func(host *ctx.Message) *ctx.Message {
							m.Copy(host, "result")
							m.Copy(host, "append")
							m.Back(m)
							Done(m, lock)
							return nil
						}, "send", "send", arg)
					} else {
						m.Back(m)
						return
					}
				}
				Wait(m, lock)
			}
			// }}}
		}},
		"good": &ctx.Command{Name: "good", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Append("share", m.Cap("share")) // {{{
			m.Append("level", m.Cap("level"))
			m.Append("type", m.Conf("type"))
			m.Append("value", m.Cap("route"))
			m.Append("kind", m.Conf("kind"))
			m.Append("name", m.Conf("name"))
			m.Append("mark", m.Conf("mark"))
			m.Append("count", m.Cap("count"))
			m.Back(m)
			// }}}
		}},
		"register": &ctx.Command{Name: "remote detail...", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			ssh, ok := m.Target().Server.(*SSH)
			m.Assert(ok)

			m.Capi("nsend", 1)
			m.Add("option", "nrecv", m.Cap("nsend"))
			ssh.send[m.Cap("nsend")] = m

			for _, v := range arg {
				fmt.Fprintf(ssh.Writer, "detail: %v\n", v)
			}
			for _, k := range m.Meta["option"] {
				if k == "args" {
					continue
				}
				for _, v := range m.Meta[k] {
					fmt.Fprintf(ssh.Writer, "%s: %s\n", k, v)
				}
			}
			fmt.Fprintf(ssh.Writer, "\n")
			ssh.Writer.Flush()

			m.Recv = make(chan bool)
			<-m.Recv
		}},
	},
}

func init() {
	ssh := &SSH{}
	ssh.Context = Index
	ctx.Index.Register(Index, ssh)
}
