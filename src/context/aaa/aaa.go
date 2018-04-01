package aaa // {{{
// }}}
import ( // {{{
	"context"

	"crypto/md5"
	"encoding/hex"
	"math/rand"

	"fmt"
	"strconv"
	"time"
)

// }}}

type AAA struct {
	sessions map[string]*ctx.Context
	*ctx.Context
}

func (aaa *AAA) Session(meta string) string { // {{{
	bs := md5.Sum([]byte(fmt.Sprintln("%d%d%s", time.Now().Unix(), rand.Int(), meta)))
	sessid := hex.EncodeToString(bs[:])
	return sessid
}

// }}}

func (aaa *AAA) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(AAA)
	s.Context = c
	return s
}

// }}}
func (aaa *AAA) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	aaa.Context.Master(nil)
	aaa.Caches["group"] = &ctx.Cache{Name: "用户组", Value: "", Help: "用户组"}
	aaa.Caches["username"] = &ctx.Cache{Name: "用户名", Value: "", Help: "用户名"}
	aaa.Caches["password"] = &ctx.Cache{Name: "用户密码", Value: "", Help: "用户密码，加密存储", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
		if len(arg) > 0 {
			bs := md5.Sum([]byte(fmt.Sprintln("用户密码:%s", arg[0])))
			m.Assert(x.Value == "" || x.Value == hex.EncodeToString(bs[:]), "密码错误")
			m.Cap("expire", fmt.Sprintf("%d", time.Now().Unix()+int64(Pulse.Confi("expire"))))
			return hex.EncodeToString(bs[:])
		}
		return x.Value
	}}

	aaa.Caches["sessid"] = &ctx.Cache{Name: "会话令牌", Value: "", Help: "用户的会话标识"}
	aaa.Caches["expire"] = &ctx.Cache{Name: "会话超时", Value: "", Help: "用户的会话标识"}
	aaa.Caches["time"] = &ctx.Cache{Name: "登录时间", Value: fmt.Sprintf("%d", time.Now().Unix()), Help: "用户登录时间", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
		if len(arg) > 0 {
			return arg[0]
		}

		n, e := strconv.Atoi(x.Value)
		m.Assert(e)
		return time.Unix(int64(n), 0).Format("15:03:04")
	}}

	if m.Target() == Index {
		Pulse = m
	}
	return aaa
}

// }}}
func (aaa *AAA) Start(m *ctx.Message, arg ...string) bool { // {{{
	if len(arg) > 1 && m.Cap("sessid") == "" {
		m.Cap("group", arg[0])
		m.Cap("username", arg[1])
		m.Cap("stream", m.Cap("username"))
		m.Cap("sessid", aaa.Session(arg[1]))
		Pulse.Capi("nuser", 1)
		aaa.Owner = aaa.Context
		aaa.Group = arg[0]
	}

	m.Log("info", m.Source(), "%s login %s %s", Pulse.Cap("nuser"), m.Cap("group"), m.Cap("username"))
	return false
}

// }}}
func (aaa *AAA) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch aaa.Context {
	case m.Target():
		root := Pulse.Target().Server.(*AAA)
		delete(root.sessions, m.Cap("sessid"))
		m.Log("info", nil, "%d logout %s", Pulse.Capi("nuser", -1)+1, m.Cap("username"))
	case m.Source():
	}

	return true
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "aaa", Help: "认证中心",
	Caches: map[string]*ctx.Cache{
		"nuser": &ctx.Cache{Name: "用户数量", Value: "0", Help: "用户数量"},
	},
	Configs: map[string]*ctx.Config{
		"rootname": &ctx.Config{Name: "根用户名", Value: "root", Help: "根用户名"},
		"expire":   &ctx.Config{Name: "会话超时(s)", Value: "120", Help: "会话超时"},
	},
	Commands: map[string]*ctx.Command{
		"login": &ctx.Command{Name: "login [sessid]|[[group] username password]]", Help: "用户登录", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Target(c) // {{{
			aaa := c.Server.(*AAA)

			switch len(arg) {
			case 0:
				m.Travel(c, func(m *ctx.Message) bool {
					m.Echo("%s(%s): %s\n", m.Target().Name, m.Cap("group"), m.Cap("time"))
					if int64(m.Capi("expire")) < time.Now().Unix() {
						m.Target().Close(m)
					}
					return true
				})
			case 1:
				s, ok := aaa.sessions[arg[0]]
				m.Assert(ok, "会话失败")
				m.Target(s)
				m.Assert(int64(m.Capi("expire")) > time.Now().Unix(), "会话失败")

				m.Source().Group, m.Source().Owner = m.Cap("group"), m.Target()
				m.Log("info", m.Source(), "logon %s %s", m.Cap("username"), m.Cap("group"))
				m.Echo(m.Cap("username"))
			case 2, 3:
				group, username, password := arg[0], arg[0], arg[1]
				if len(arg) == 3 {
					username, password = arg[1], arg[2]
				}

				msg := m
				if username == Pulse.Conf("rootname") {
					msg = Pulse.Spawn(Pulse.Target())
					msg.Set("detail", group, username).Target().Start(msg)
				} else if msg = Pulse.Find(username, false); msg == nil {
					m.Start(username, "认证用户", group, username)
					msg = m
				} else {
					m.Target(msg.Target())
				}

				msg.Cap("password", password)
				m.Source().Group, m.Source().Owner = msg.Cap("group"), msg.Target()
				aaa.sessions[m.Cap("sessid")] = msg.Target()
				m.Echo(msg.Cap("sessid"))
			} // }}}
		}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void", Commands: map[string]*ctx.Command{"login": &ctx.Command{}}},
	},
}

func init() {
	aaa := &AAA{}
	aaa.Context = Index
	ctx.Index.Register(Index, aaa)

	aaa.sessions = make(map[string]*ctx.Context)
}
