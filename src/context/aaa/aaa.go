package aaa // {{{
// }}}
import ( // {{{
	"context"
	"log"
	"time"
)

// }}}

type AAA struct {
	username  string
	password  string
	logintime time.Time
	*ctx.Context
}

func (aaa *AAA) Begin(m *ctx.Message) ctx.Server { // {{{
	return aaa
}

// }}}
func (aaa *AAA) Start(m *ctx.Message) bool { // {{{
	return true
}

// }}}
func (aaa *AAA) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}
	c.Commands = map[string]*ctx.Command{}

	s := new(AAA)
	s.username = arg[0]
	s.password = arg[1]
	s.logintime = time.Now()
	s.Context = c
	return s
}

// }}}
func (aaa *AAA) Exit(m *ctx.Message, arg ...string) bool { // {{{
	return true
}

// }}}
var Index = &ctx.Context{Name: "aaa", Help: "会话管理",
	Caches: map[string]*ctx.Cache{
		"status": &ctx.Cache{Name: "status", Value: "stop", Help: "服务状态"},
		"root":   &ctx.Cache{Name: "root", Value: "root", Help: "初始用户"},
	},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"login": &ctx.Command{Name: "login [username [password]]", Help: "会话管理", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) {
			case 0:
				m.Target.Travel(func(s *ctx.Context) bool {
					aaa := s.Server.(*AAA)
					m.Echo("%s(%s): %s\n", s.Name, aaa.username, aaa.logintime.Format("15:04:05"))
					return true
				})
			case 1:
				m.Target.Travel(func(s *ctx.Context) bool {
					aaa := s.Server.(*AAA)
					if aaa.username == arg[0] {
						m.Echo("%s(%s): %s\n", s.Name, aaa.username, aaa.logintime.Format("15:04:05"))
						return false
					}
					return true
				})
			case 2:
				m.Target.Travel(func(s *ctx.Context) bool {
					aaa := s.Server.(*AAA)
					if aaa.username == arg[0] {
						m.Add("result", arg[0])
						if aaa.password == arg[1] {
							m.Add("result", time.Now().Format("15:04:05"))
						}
						return false
					}
					return true
				})
				if m.Get("result") == arg[0] {
					if len(m.Meta["result"]) == 2 {
						m.Echo("login success\n")
						log.Println("login success")
					} else {
						m.Echo("login error\n")
						log.Println("login error")
					}
				} else {
					m.Start(arg[0], arg[1])
					m.Add("result", arg[0])
				}
			}
			return ""
		}},
	},
}

func init() {
	aaa := &AAA{username: "root", password: "root", logintime: time.Now()}
	aaa.Context = Index
	ctx.Index.Register(Index, aaa)
}
