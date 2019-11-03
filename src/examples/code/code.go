package code

import (
	"contexts/ctx"
	"contexts/web"
	"path"
	"regexp"
	"strings"
	"time"
	"toolkit"
)

var Index = &ctx.Context{Name: "code", Help: "代码中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"docker": {Name: "docker", Help: "容器", Value: map[string]interface{}{
			"shy": `
FROM {{options . "base"}}

WORKDIR /home/{{options . "user"}}/context
Env ctx_dev {{options . "host"}}

RUN wget -q -O - $ctx_dev/publish/boot.sh | sh -s install

CMD sh bin/boot.sh

`,
		}},
	},
	Commands: map[string]*ctx.Command{
		"zsh": {Name: "zsh dir grep key [split reg fields] [filter reg fields] [order key method] [group keys method] [sort keys method]",
			Form: map[string]int{"split": 2, "filter": 2, "order": 2, "group": 2, "sort": 2},
			Help: "终端", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				p, arg := kit.Select(".", arg[0]), arg[1:]
				switch arg[0] {
				case "init":
					m.Cmd("cli.system", "apk", "update")
					switch arg[1] {
					case "build":
						m.Cmd("cli.system", "apk", "add", "nginx")
						m.Cmd("cli.system", "apk", "add", "redis")
						m.Cmd("cli.system", "apk", "add", "tmux")
						m.Cmd("cli.system", "apk", "add", "zsh")
						m.Cmd("cli.system", "apk", "add", "git")
						m.Cmd("cli.system", "apk", "add", "vim")
						m.Cmd("cli.system", "apk", "add", "build-base")
						m.Cmd("cli.system", "apk", "add", "golang")
						m.Cmd("cli.system", "apk", "add", "mysql")
					}

				case "list":
					m.Cmdy("nfs.dir", p, "time", "size", "path")

				case "find":
					m.Cmdy("cli.system", "find", p, "-name", arg[1])

				case "grep":
					if m.Options("split") {
						re, _ := regexp.Compile(kit.Select("", m.Optionv("split"), 0))
						fields := map[string]bool{}
						for _, v := range strings.Split(kit.Select("", m.Optionv("split"), 1), " ") {
							if v != "" {
								fields[v] = true
							}
						}

						m.Cmd("cli.system", "grep", "-rn", arg[1], p, "cmd_parse", "cut", ":", "3", "path line text").Table(func(index int, line map[string]string) {
							if ls := re.FindAllStringSubmatch(line["text"], -1); len(ls) > 0 {
								m.Push("path", line["path"])
								m.Push("line", line["line"])
								for _, v := range ls {
									if len(fields) == 0 || fields[v[1]] {
										m.Push(v[1], v[2])
									}
								}
							}
						})
						m.Table()
					} else {
						m.Cmdy("cli.system", "grep", "-rn", arg[1], p, "cmd_parse", "cut", ":", "3", "path line text")
					}

					if m.Has("filter") {
						m.Filter(m.Option("filter"))
					}
					if m.Has("order") {
						m.Sort(kit.Select("", m.Optionv("order"), 0), kit.Select("", m.Optionv("order"), 1))
					}
					if m.Has("group") {
						m.Group(kit.Select("sum", m.Optionv("group"), 1), strings.Split(kit.Select("", m.Option("group"), 0), " ")...)
					}
					if m.Has("sort") {
						m.Sort(kit.Select("", m.Optionv("sort"), 0), kit.Select("", m.Optionv("sort"), 1))
					}

				case "tail":
					m.Cmdy("cli.system", "tail", path.Join(p, arg[1]))

				default:
					m.Cmdy("cli.system", arg)
				}
				return
			}},
		"tmux": {Name: "tmux [session [window [pane cmd]]]", Help: "窗口", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			prefix := []string{"cli.system", "tmux"}
			if len(arg) > 1 {
				switch arg[1] {
				case "modify":
					switch arg[2] {
					case "session":
						m.Cmdy(prefix, "rename-session", "-t", arg[0], arg[3])
					case "window":
						m.Cmdy(prefix, "rename-window", "-t", arg[0], arg[3])
					}
					return
				}
			}
			// 查看会话
			if m.Cmdy(prefix, "list-session", "-F", "#{session_id},#{session_name},#{session_windows},#{session_height},#{session_width}",
				"cmd_parse", "cut", ",", "5", "id session windows height width"); len(arg) == 0 {
				return
			}

			// 创建会话
			if arg[0] != "" && !kit.Contains(m.Meta["session"], arg[0]) {
				m.Cmdy(prefix, "new-session", "-ds", arg[0])
			}
			m.Set("append").Set("result")

			// 查看窗口
			if m.Cmdy(prefix, "list-windows", "-t", arg[0], "-F", "#{window_id},#{window_name},#{window_panes},#{window_height},#{window_width}",
				"cmd_parse", "cut", ",", "5", "id window panes height width"); len(arg) == 1 {
				return
			}

			// 创建窗口
			if arg[1] != "" && !kit.Contains(m.Meta["window"], arg[1]) {
				m.Cmdy(prefix, "new-window", "-dt", arg[0], "-n", arg[1])
			}
			m.Set("append").Set("result")

			// 查看面板
			if len(arg) == 2 {
				m.Cmdy(prefix, "list-panes", "-t", arg[0]+":"+arg[1], "-F", "#{pane_id},#{pane_index},#{pane_tty},#{pane_height},#{pane_width}",
					"cmd_parse", "cut", ",", "5", "id pane tty height width")
				return
			}

			// 执行命令
			target := arg[0] + ":" + arg[1] + "." + arg[2]
			if len(arg) > 3 {
				if len(arg) > 5 {
					switch arg[5] {
					case "modify":
						switch arg[6] {
						case "text":
							m.Cmdy(prefix, "set-buffer", "-b", arg[4], arg[7])
						}
						return
					}
				}
				switch arg[3] {
				case "buffer":
					if len(arg) > 5 {
						m.Cmdy(prefix, "set-buffer", "-b", arg[4], arg[5])
					}
					if len(arg) > 4 {
						m.Cmdy(prefix, "show-buffer", "-b", arg[4])
						return
					}
					m.Cmdy(prefix, "list-buffers", "cmd_parse", "cut", ": ", "3", "buffer size text")
					for i, v := range m.Meta["text"] {
						if i < 3 {
							m.Meta["text"][i] = m.Cmdx(prefix, "show-buffer", "-b", m.Meta["buffer"][i])
						} else {
							m.Meta["text"][i] = v[2 : len(v)-1]
						}
					}
					return
				case "layout":
				default:
					m.Cmdy(prefix, "send-keys", "-t", target, strings.Join(arg[3:], " "), "Enter")
					time.Sleep(1 * time.Second)
				}
			}

			// 查看终端
			m.Echo(strings.TrimSpace(m.Cmdx(prefix, "capture-pane", "-pt", target)))
			return
		}},
		"docker": {Name: "docker", Help: "容器", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			prefix := []string{"cli.system", "docker"}
			switch arg[0] {
			case "image":
				prefix = append(prefix, "image")
				pos := kit.Select("shy", arg, 1)
				tag := kit.Select("2.1", arg, 2)

				// 查看镜像
				if m.Cmdy(prefix, "ls", "cmd_parse", "cut", "cmd_headers", "IMAGE ID", "IMAGE_ID"); len(arg) == 1 {
					return
				} else if i := kit.IndexOf(m.Meta["IMAGE_ID"], arg[1]); i > -1 {
					arg, pos, tag = arg[2:], strings.TrimSpace(m.Meta["REPOSITORY"][i]), strings.TrimSpace(m.Meta["TAG"][i])
				} else {
					arg = arg[3:]
				}

				// 拉取镜像
				if len(arg) == 0 {
					m.Cmdy(prefix, "pull", pos+":"+tag)
					return
				}

				switch arg[0] {
				// 启动容器
				case "运行":
					m.Set("append").Set("result")
					m.Cmdy("cli.system", "docker", "run", "-dt", pos+":"+tag)
					return
				// 清理镜像
				case "清理":
					m.Cmd(prefix, "prune", "-f")

				// 删除镜像
				case "delete":
					m.Cmd(prefix, "rm", pos+":"+tag)

				// 创建镜像
				default:
					m.Option("base", pos+":"+tag)
					m.Option("name", arg[0]+":"+kit.Select("2.1", arg, 1))
					m.Option("host", "http://"+m.Conf("runtime", "boot.hostname")+".local:9095")
					m.Option("user", kit.Select("shy", arg, 2))
					m.Option("file", "etc/Dockerfile")

					if f, _, e := kit.Create(m.Option("file")); m.Assert(e) {
						defer f.Close()
						if m.Assert(ctx.ExecuteStr(m, f, m.Conf("docker", arg[0]))) {
							m.Cmdy("cli.system", "docker", "image", "build", "-f", m.Option("file"), "-t", m.Option("name"), ".")
						}
					}
				}

			case "container":
				prefix = append(prefix, "container")
				if len(arg) > 1 {
					switch arg[2] {
					case "进入":
						m.Cmdy("cli.system", "tmux", "new-window", "-dPF", "#{session_name}:#{window_name}.1", "docker exec -it "+arg[1]+" sh")
						return

					case "停止":
						m.Cmd(prefix, "stop", arg[1])

					case "启动":
						m.Cmd(prefix, "start", arg[1])

					case "重启":
						m.Cmd(prefix, "restart", arg[1])

					case "清理":
						m.Cmd(prefix, "prune", "-f")

					case "modify":
						switch arg[3] {
						case "NAMES":
							m.Cmd(prefix, "rename", arg[1], arg[4:])
						}

						// 删除容器
					case "delete":
						m.Cmd(prefix, "rm", arg[1])

					default:
						if len(arg) > 2 {
							m.Cmdy(prefix, "exec", arg[1], arg[2:])
						} else {
							m.Cmdy(prefix, "inspect", arg[1])
						}
						return
					}
				}
				m.Cmdy(prefix, "ls", "-a", "cmd_parse", "cut", "cmd_headers", "CONTAINER ID", "CONTAINER_ID")

			case "network":
				if len(arg) == 1 {
					m.Cmdy("cli.system", "docker", "network", "ls", "cmd_parse", "cut", "cmd_headers", "NETWORK ID", "NETWORK_ID")
					break
				}

				kit.Map(kit.Chain(kit.UnMarshal(m.Cmdx("cli.system", "docker", "network", "inspect", arg[1])), "0.Containers"), "", func(key string, value map[string]interface{}) {
					m.Push("CONTAINER_ID", key[:12])
					m.Push("name", value["Name"])
					m.Push("IPv4", value["IPv4Address"])
					m.Push("IPv6", value["IPV4Address"])
					m.Push("Mac", value["MacAddress"])
				})
				m.Table()

			case "volume":
				if len(arg) == 1 {
					m.Cmdy("cli.system", "docker", "volume", "ls", "cmd_parse", "cut", "cmd_headers", "VOLUME NAME", "VOLUME_NAME")
					break
				}

			default:
				m.Cmdy("cli.system", "docker", arg)
			}
			return
		}},
		"git": {Name: "git", Help: "版本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			prefix := []string{"cli.system", "git"}
			switch arg[0] {
			case "init":
				m.Cmdy(prefix, "config", "alias.s", "status")
			}
			m.Echo("git")
			return
		}},
		"vim": {Name: "vim", Help: "编辑器", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo("vim")
			return
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.WEB{Context: Index})
}
