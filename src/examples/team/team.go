package team

import (
	"contexts/ctx"
	"contexts/web"
	"fmt"
	"toolkit"
)

var Index = &ctx.Context{Name: "team", Help: "团队中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"task": {Name: "task create table level class status begin_time close_time target detail arg...", Help: "任务", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch arg[0] {
			case "progress":
				if len(arg) > 3 && arg[1] != "" {
					switch arg[3] {
					case "prepare", "action", "cancel", "finish":
						time := "close_time"
						switch arg[3] {
						case "prepare", "action":
							time = "begin_time"
						case "cancel", "finish":
							time = "close_time"
						default:
							time = "update_time"
						}

						// 更新任务
						m.Cmd("ssh.data", "update", arg[1], arg[2], "status", arg[3], time, m.Time())
						arg = []string{arg[0], m.Option("table")}
					}
				}
				// 任务进度
				m.Option("cache.limit", kit.Select("30", arg, 2))
				m.Option("cache.offend", kit.Select("0", arg, 3))
				m.Meta["append"] = []string{"prepare", "action", "cancel", "finish"}
				m.Cmd("ssh.data", "show", arg[1]).Table(func(index int, value map[string]string) {
					m.Push(value["status"],
						fmt.Sprintf("<span data-id='%s' title='%s'>%s</span>", value["id"], value["detail"], value["target"]))
				})
				m.Table()

			case "create":
				// 创建任务
				if len(arg) > 7 {
					if len(arg) < 9 {
						arg = append(arg, "")
					}
					m.Cmdy("ssh.data", "insert", arg[1], "level", arg[2], "class", arg[3],
						"status", arg[4], "begin_time", arg[5], "close_time", arg[6],
						"target", arg[7], "detail", arg[8], arg[9:])
					break
				}

				arg = []string{arg[1]}
				fallthrough
			default:
				// 更新任务
				if len(arg) > 1 && arg[1] == "modify" {
					m.Cmdy("ssh.data", "update", m.Option("table"), arg[0], arg[2], arg[3])
					break
				}

				// 查看任务
				if len(arg) > 0 && arg[0] == "" {
					arg = arg[:0]
				}
				m.Cmdy("ssh.data", "show", arg, "fields", "id", "level", "class", "status", "target", "begin_time", "close_time")
			}
			return
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.WEB{Context: Index})
}
