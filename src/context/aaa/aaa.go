package aaa // {{{
// }}}
import ( // {{{
	"context"
	_ "context/cli"

	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"time"

	"fmt"
	"log"
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

func (aaa *AAA) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	return aaa
}

// }}}
func (aaa *AAA) Start(m *ctx.Message, arg ...string) bool { // {{{
	return true
}

// }}}
func (aaa *AAA) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{
		"username": &ctx.Cache{Name: "用户名", Value: arg[0], Help: "显示已经启动运行模块的数量"},
		"password": &ctx.Cache{Name: "密码", Value: "", Help: "用户密码，加密存储", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
			if len(arg) > 0 { // {{{
				if arg[0] == "" {
					return ""
				}

				if x.Value == "" {
					bs := md5.Sum([]byte(fmt.Sprintln("用户密码:%s", arg[0])))
					return hex.EncodeToString(bs[:])
				} else {
					bs := md5.Sum([]byte(fmt.Sprintln("用户密码:%s", arg[0])))
					if x.Value != hex.EncodeToString(bs[:]) {
						log.Println(m.Target.Name, "login in:", arg[0], "密码错误")
						panic("密码错误")
					}
				}
			}
			return x.Value
			// }}}
		}},
		"group":  &ctx.Cache{Name: "群组", Value: arg[0], Help: "用户所属群组"},
		"sessid": &ctx.Cache{Name: "会话标识", Value: aaa.session(arg[0]), Help: "用户的会话标识"},
		"time":   &ctx.Cache{Name: "登录时间", Value: fmt.Sprintf("%d", time.Now().Unix()), Help: "用户登录时间"},
	}

	if len(arg) > 2 {
		c.Caches["password"].Value = arg[1]
		c.Caches["group"].Value = arg[2]
	} else if len(arg) > 1 {
		m.Cap("password", arg[1])
	}

	s := new(AAA)
	s.Context = c
	return s
}

// }}}
func (aaa *AAA) Exit(m *ctx.Message, arg ...string) bool { // {{{
	return true
}

// }}}

var Index = &ctx.Context{Name: "aaa", Help: "认证中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"rootname": &ctx.Config{Name: "根用户的名称", Value: "root", Help: "系统根据此名确定是否超级用户"},
	},
	Commands: map[string]*ctx.Command{
		"login": &ctx.Command{Name: "login [sessid]|[username password [group]]]", Help: "", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			aaa := c.Server.(*AAA) // {{{
			switch len(arg) {
			case 0:
				m.Travel(m.Target, func(m *ctx.Message) bool {
					if m.Target.Name == "aaa" {
						return true
					}
					m.Echo("%s %s %s\n", m.Target.Name, m.Cap("group"), m.Cap("sessid"))
					return true
				})
			case 1:
				target := m.Target
				if s, ok := aaa.sessions[arg[0]]; ok {
					m.Target = s
					m.Source.Owner = s
					log.Println(aaa.Name, "login on:", aaa.sessions)
					return m.Cap("username")
				}
				m.Target = target
			case 2:
				s := m.Target.Find(arg[0])
				if s != nil {
					old := m.Source
					defer func() { m.Source = old }()
					m.Source = s

					m.Target = s

					m.Cap("password", arg[1])
					log.Println(aaa.Name, "login in:", arg[0])
					old.Owner = s

				} else {
					m.Start(arg[0], arg...)
					s = m.Target

					aaa.sessions[m.Cap("sessid")] = s
					log.Println(aaa.Name, "login up:", arg[0])
				}

				m.Target.Owner = s

				m.Source.Owner = ctx.Index.Owner
				if arg[0] == m.Conf("rootname") {
					ctx.Index.Owner = s
					m.Travel(m.Target.Root, func(m *ctx.Message) bool {
						if m.Target.Owner == nil {
							m.Target.Owner = s
						}
						return true
					})
				}
				m.Source.Owner = s
				m.Source.Group = m.Cap("group")

				return m.Cap("sessid")
			case 3:
				m.Start(arg[0], arg...)
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
