package ssh // {{{
// }}}
import ( // {{{
	"contexts"

	"fmt"
	"strings"
	"time"
)

// }}}

type SSH struct {
	nfs *ctx.Context

	*ctx.Message
	*ctx.Context
}

func (ssh *SSH) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{
		"domain": &ctx.Config{Name: "domain", Value: "", Help: "主机数量"},
	}

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
	ssh.Message = m
	ssh.nfs = m.Source()
	m.Cap("stream", m.Source().Name)
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

func Done(m *ctx.Message, lock chan bool) { // {{{
	m.Log("lock", nil, "done before %v", m.Meta["detail"])
	if m.Options("stdio") {
		lock <- true
	}
	m.Log("lock", nil, "done after %v", m.Meta["detail"])
}

// }}}
func Wait(m *ctx.Message, lock chan bool) { // {{{
	m.Log("lock", nil, "wait before %v", m.Meta["detail"])
	if m.Options("stdio") {
		<-lock
	}
	m.Log("lock", nil, "wait after %v", m.Meta["detail"])
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches: map[string]*ctx.Cache{
		"nhost": &ctx.Cache{Name: "主机数量", Value: "0", Help: "主机数量"},
		"route": &ctx.Cache{Name: "route", Value: "com", Help: "主机数量"},
		"count": &ctx.Cache{Name: "count", Value: "3", Help: "主机数量"},
		"share": &ctx.Cache{Name: "share", Value: "root", Help: "主机数量"},
		"level": &ctx.Cache{Name: "level", Value: "root", Help: "主机数量"},

		"domain": &ctx.Cache{Name: "domain", Value: "com", Help: "主机数量"},
	},
	Configs: map[string]*ctx.Config{
		"domain": &ctx.Config{Name: "domain", Value: "com", Help: "主机数量"},

		"route.json": &ctx.Config{Name: "route.json", Value: "var/route.json", Help: "主机数量"},
		"route.png":  &ctx.Config{Name: "route.png", Value: "var/route.png", Help: "主机数量"},
		"type":       &ctx.Config{Name: "type", Value: "terminal", Help: "主机数量"},
		"kind":       &ctx.Config{Name: "kind", Value: "terminal", Help: "主机数量"},
		"name":       &ctx.Config{Name: "name", Value: "vps", Help: "主机数量"},
		"mark":       &ctx.Config{Name: "mark", Value: "com", Help: "主机数量"},
	},
	Commands: map[string]*ctx.Command{
		"listen": &ctx.Command{Name: "listen address protocol", Help: "监听连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*SSH); m.Assert(ok) { // {{{
				m.Find("nfs").Call(func(file *ctx.Message) *ctx.Message {
					sub := file.Spawn(m.Target())
					sub.Start(fmt.Sprintf("host%d", Pulse.Capi("nhost", 1)), "远程主机")
					m.Sessions["ssh"] = sub
					// sub.Sesss("nfs").Cmd("send", "route", sub.Target().Name, m.Cap("route"))
					return sub
				}, m.Meta["detail"])
			}
			// }}}
		}},
		"dial": &ctx.Command{Name: "dial address protocol", Help: "建立连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*SSH); m.Assert(ok) { // {{{
				m.Find("nfs").Call(func(file *ctx.Message) *ctx.Message {
					sub := file.Spawn(m.Target())
					sub.Target().Start(sub)
					m.Sessions["ssh"] = sub

					time.Sleep(time.Second)
					sub.Spawn(sub.Target()).Cmd("pwd", m.Conf("domain"))
					return sub
				}, m.Meta["detail"])
			}
			// }}}
		}},
		"route": &ctx.Command{Name: "route", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Conf("route", arg[0]) // {{{
			m.Cap("route", arg[1]+"."+arg[0])
			info := map[string]string{"route": m.Cap("route")}

			msg := m.Sesss("nfs")
			msg.Put("option", "data", info)
			msg.Cmd("json", m.Conf("route.json"))

			png := m.Sesss("nfs")
			png.Cmd("genqr", m.Conf("route.png"), msg.Result(0))

			m.Back(m)
			// }}}
		}},
		"who": &ctx.Command{Name: "who", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			aaa := m.Sesss("aaa")
			if aaa != nil {
				m.Echo(aaa.Cap("group"))
			}
		}},
		"pwd": &ctx.Command{Name: "pwd", Help: "远程执行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			switch len(arg) {
			case 0:
				m.Echo(m.Cap("domain"))
			case 1:
				if m.Options("nsend") {
					m.Conf("domain", arg[0])
					m.Echo(m.Cap("domain"))
					m.Echo(".")
					m.Echo(m.Conf("domain"))
				} else {
					m.Spawn(m.Target()).CallBack(m.Options("stdio"), func(msg *ctx.Message) *ctx.Message {
						m.Conf("domain", msg.Result(2))
						m.Echo(m.Cap("domain", strings.Join(msg.Meta["result"], "")))
						m.Back(msg)
						return nil
					}, "send", "pwd", arg[0])
				}
			}
		}},
		"send": &ctx.Command{Name: "send [route domain] cmd arg...", Help: "远程执行",
			Formats: map[string]int{"route": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if ssh, ok := m.Target().Server.(*SSH); m.Assert(ok) { // {{{

					target := strings.Split(m.Option("route"), ".")
					if len(target) == 1 && len(target[0]) == 0 {
						if m.Options("nsend") {
							msg := m.Spawn(m.Target())
							msg.Cmd(arg)
							m.Back(msg)
						} else {
							ssh.Message.Sesss("nfs").CallBack(m.Options("stdio"), func(host *ctx.Message) *ctx.Message {
								m.Back(m.Copy(host, "result").Copy(host, "option"))
								return nil
							}, "send", "send", arg)
						}
						return
					}

					miss := true
					m.Travel(c, func(m *ctx.Message) bool {
						if ssh, ok := m.Target().Server.(*SSH); ok && m.Target().Name == target[0] {
							msg := m.Spawn(ssh.nfs)
							msg.Option("route", strings.Join(target[1:], "."))
							msg.Call(func(host *ctx.Message) *ctx.Message {
								return m.Copy(host, "result").Copy(host, "append")
							}, "send", "send", arg)

							miss = false
						}
						return miss
					})

					if ssh, ok := c.Server.(*SSH); m.Assert(ok) && ssh.nfs != nil {
						msg := m.Spawn(ssh.nfs)
						msg.Option("route", m.Option("route"))
						msg.Call(func(host *ctx.Message) *ctx.Message {
							return m.Copy(host, "result").Copy(host, "append")
						}, "send", "send", arg)
					}
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
	},
}

func init() {
	ssh := &SSH{}
	ssh.Context = Index
	ctx.Index.Register(Index, ssh)
}
