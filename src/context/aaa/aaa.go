package aaa // {{{
// }}}
import ( // {{{
	"context"
	_ "context/cli"

	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"

	"fmt"
)

// }}}

type AAA struct {
	sessions map[string]*ctx.Context
	*ctx.Context
}

func (aaa *AAA) session(meta string) string { // {{{
	bs := md5.Sum([]byte(fmt.Sprintln("%d%d%s", time.Now().Unix(), rand.Int(), meta)))
	sessid := hex.EncodeToString(bs[:])
	return sessid
}

// }}}

func (aaa *AAA) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"sessid": &ctx.Cache{Name: "会话标识", Value: "", Help: "用户的会话标识"},
	}
	c.Configs = map[string]*ctx.Config{}

	s := new(AAA)
	s.Context = c
	return s
}

// }}}
func (aaa *AAA) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	aaa.Caches["group"] = &ctx.Cache{Name: "用户组", Value: m.Conf("rootname"), Help: "用户组"}
	aaa.Caches["username"] = &ctx.Cache{Name: "用户名", Value: m.Conf("rootname"), Help: "用户名"}
	aaa.Caches["password"] = &ctx.Cache{Name: "密码", Value: "", Help: "用户密码，加密存储", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
		if len(arg) > 0 { // {{{
			if x.Value == "" {
				bs := md5.Sum([]byte(fmt.Sprintln("用户密码:%s", arg[0])))
				return hex.EncodeToString(bs[:])
			} else {
				bs := md5.Sum([]byte(fmt.Sprintln("用户密码:%s", arg[0])))
				m.Assert(x.Value == hex.EncodeToString(bs[:]), "密码错误")
			}
		}
		return x.Value
		// }}}
	}}
	aaa.Caches["time"] = &ctx.Cache{Name: "登录时间", Value: fmt.Sprintf("%d", time.Now().Unix()), Help: "用户登录时间", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
		if len(arg) > 0 { // {{{
			return x.Value
		}

		n, e := strconv.Atoi(x.Value)
		m.Assert(e)

		return time.Unix(int64(n), 0).Format("15:03:04")
		// }}}
	}}

	if len(arg) > 0 {
		m.Cap("username", arg[0])
		m.Cap("group", arg[0])
	}
	if len(arg) > 1 {
		m.Cap("group", arg[1])
	}
	m.Capi("nuser", 1)

	return aaa
}

// }}}
func (aaa *AAA) Start(m *ctx.Message, arg ...string) bool { // {{{
	return false
}

// }}}
func (aaa *AAA) Close(m *ctx.Message, arg ...string) bool { // {{{
	m.Master = Index
	if m.Cap("username") != m.Conf("rootname") {
		return true
	}
	return false
}

// }}}

var Index = &ctx.Context{Name: "aaa", Help: "认证中心",
	Caches: map[string]*ctx.Cache{
		"nuser": &ctx.Cache{Name: "用户数量", Value: "0", Help: "用户数量"},
	},
	Configs: map[string]*ctx.Config{
		"rootname": &ctx.Config{Name: "根用户名", Value: "root", Help: "根用户名"},
	},
	Commands: map[string]*ctx.Command{
		"login": &ctx.Command{Name: "login [sessid]|[[group] username password]]", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			m.Master = m.Target // {{{
			aaa := m.Target.Server.(*AAA)

			switch len(arg) {
			case 0:
				m.Travel(m.Target, func(m *ctx.Message) bool {
					m.Echo("%s(%s): %s\n", m.Target.Name, m.Cap("group"), m.Cap("time"))
					return true
				})
			case 1:
				target := m.Target
				if s, ok := aaa.sessions[arg[0]]; ok {
					m.Target = s
					m.Source.Owner = s
					m.Log("info", "%s: logon %s", aaa.Name, m.Cap("username"))
					return m.Cap("username")
				}
				m.Target = target
			case 2:
				if arg[0] == m.Conf("rootname") {
					m.Cap("password", arg[1])
					m.Source.Owner = aaa.Context
					m.Travel(m.Target.Root, func(m *ctx.Message) bool {
						if m.Target.Owner == nil {
							m.Target.Owner = aaa.Context
						}
						return true
					})
					return ""
				}

				source := m.Source
				if msg := m.Find(arg[0]); msg == nil {
					m.Start(arg[0], arg[0], arg[0])
					m.Cap("sessid", aaa.session(arg[0]))
					m.Cap("time", fmt.Sprintf("%d", time.Now().Unix()))
				} else {
					m.Target = msg.Target
				}

				m.Cap("password", arg[1])
				m.Log("info", "%s: login", m.Target.Name)
				aaa.sessions[m.Cap("sessid")] = m.Target

				m.Target.Owner = m.Target
				source.Owner = m.Target
				source.Group = m.Cap("group")

				m.Log("info", "%s: login", source.Name)
				return m.Cap("sessid")

			case 3:
				m.Start(arg[0], arg[0], arg[0])
			}
			return ""
			// }}}
		}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void",
			Commands: map[string]*ctx.Command{
				"login": &ctx.Command{},
			},
		},
	},
}

func init() {
	aaa := &AAA{}
	aaa.Context = Index
	ctx.Index.Register(Index, aaa)

	aaa.sessions = make(map[string]*ctx.Context)
}
