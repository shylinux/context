package team

import (
	"contexts/ctx"
	"contexts/web"
	"fmt"
)

var Index = &ctx.Context{Name: "team", Help: "团队中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"task": {Name: "task table index level status begin_time close_time target detail", Help: "任务", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch arg[0] {
			case "progress":
				if len(arg) > 2 {
					switch arg[2] {
					case "prepare", "action", "cancel", "finish":
						m.Cmd("ssh.data", "update", m.Option("table"), arg[1], "status", arg[2])
						arg = []string{arg[0], m.Option("table")}
					}
				}
				m.Meta["append"] = []string{"prepare", "action", "cancel", "finish"}
				m.Cmd("ssh.data", "show", arg[1:]).Table(func(index int, value map[string]string) {
					m.Push(value["status"],
						fmt.Sprintf("<span data-id='%s' title='%s'>%s</span>", value["id"], value["detail"], value["target"]))
				})
				m.Table()

			default:
				if len(arg) > 1 && arg[1] == "modify" {
					m.Cmdy("ssh.data", "update", m.Option("table"), m.Option("index"), arg[2], arg[3])
					return
				}

				if len(arg) < 8 {
					if len(arg) > 2 {
						arg = arg[:2]
					}
					if len(arg) > 1 && arg[1] == "" {
						arg = arg[:1]
					}
					if len(arg) > 0 && arg[0] == "" {
						arg = arg[:0]
					}
					m.Cmdy("ssh.data", "show", arg, "fields", "id", "status", "level", "target", "begin_time", "close_time")
					return
				}
				m.Cmdy("ssh.data", "insert", arg[0],
					"level", arg[2], "status", arg[3],
					"begin_time", arg[4], "close_time", arg[5],
					"target", arg[6], "detail", arg[7], arg[8:],
				)
			}
			return
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.WEB{Context: Index})
}
