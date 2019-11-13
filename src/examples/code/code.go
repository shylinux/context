package code

import (
	"contexts/ctx"
	"contexts/web"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"
	"toolkit"
)

var Dockfile = `
FROM {{options . "base"}}

WORKDIR /home/{{options . "user"}}/context
Env ctx_dev {{options . "host"}}

RUN wget -q -O - $ctx_dev/publish/boot.sh | sh -s install

CMD sh bin/boot.sh

`

var Index = &ctx.Context{Name: "code", Help: "代码中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"login": {Name: "login", Value: map[string]interface{}{"check": false, "local": true, "expire": "720h"}, Help: "用户登录"},
		"prefix": {Name: "prefix", Help: "外部命令", Value: map[string]interface{}{
			"zsh":    []interface{}{"cli.system", "zsh"},
			"tmux":   []interface{}{"cli.system", "tmux"},
			"docker": []interface{}{"cli.system", "docker"},
			"git":    []interface{}{"cli.system", "git"},
			"vim":    []interface{}{"cli.system", "vim"},
		}},
		"package": {Name: "package", Help: "软件包", Value: map[string]interface{}{
			"udpate":  []interface{}{"apk", "update"},
			"install": []interface{}{"apk", "add"},
			"build":   []interface{}{"build-base"},
			"develop": []interface{}{"zsh", "tmux", "git", "vim", "golang"},
			"product": []interface{}{"nginx", "redis", "mysql"},
		}},
		"docker": {Name: "docker", Help: "容器", Value: map[string]interface{}{
			"shy": Dockfile,
		}},
		"git": {Name: "git", Help: "记录", Value: map[string]interface{}{
			"alias": map[string]interface{}{"s": "status", "b": "branch"},
		}},
		"zsh": {Name: "vim", Help: "记录", Value: map[string]interface{}{
			"terminal": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "time sid status pwd pid pane hostname username",
			}},
			"history": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "time sid cmd pwd",
				"store":  "var/tmp/zsh.csv",
				"limit":  "30",
				"least":  "10",
			}},
			"free": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "time sid type total used free shared buffer available",
				"store":  "var/tmp/vim/opens.csv",
				"limit":  "30",
				"least":  "10",
			}},
			"df": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "fs size used rest per pos",
			}},
			"ps": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "PID TIME COMMAND",
			}},
			"env": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "sid name value",
			}},
		}},
		"vim": {Name: "vim", Help: "编辑器", Value: map[string]interface{}{
			"editor": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "time sid status pwd pid pane hostname username",
			}},
			"favor": map[string]interface{}{},
			"opens": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "time sid action file pwd",
				"store":  "var/tmp/vim/opens.csv",
				"limit":  "30",
				"least":  "10",
			}},
			"cmds": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "time sid cmd file pwd",
				"store":  "var/tmp/vim/cmds.csv",
				"limit":  "30",
				"least":  "10",
			}},
			"txts": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "time sid word line col file pwd",
				"store":  "var/tmp/vim/txts.csv",
				"limit":  "30",
				"least":  "10",
			}},

			"bufs": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "sid buf tag file line",
			}},
			"regs": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "sid reg word",
			}},
			"marks": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "sid mark line col file",
			}},
			"tags": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "sid tag line file",
			}},
			"fixs": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "sid fix file line word",
			}},
		}},
		"cache": {Name: "cache", Help: "缓存默认的全局配置", Value: map[string]interface{}{
			"store": "var/tmp/hi.csv",
			"limit": 6,
			"least": 3,
		}},
	},
	Commands: map[string]*ctx.Command{
		"/zsh": {Name: "/zsh sid pwd cmd arg", Help: "终端", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			cmd = strings.TrimPrefix(cmd, "/")
			if f, _, e := m.Optionv("request").(*http.Request).FormFile("sub"); e == nil {
				defer f.Close()
				if b, e := ioutil.ReadAll(f); e == nil {
					m.Option("sub", string(b))
				}
			}
			m.Log("info", "%v %v %v %v", cmd, m.Option("cmd"), m.Option("arg"), m.Option("sub"))

			switch m.Option("cmd") {
			case "login":
				name := kit.Hashs(m.Option("pid"), m.Option("hostname"), m.Option("username"))
				m.Conf(cmd, []string{"terminal", "hash", name}, map[string]interface{}{
					"time":     m.Time(),
					"status":   "login",
					"sid":      name,
					"pwd":      m.Option("pwd"),
					"pid":      m.Option("pid"),
					"pane":     m.Option("pane"),
					"hostname": m.Option("hostname"),
					"username": m.Option("username"),
				})
				m.Echo(name)
			case "logout":
				name := m.Option("sid")
				m.Conf(cmd, []string{"terminal", "hash", name, "time"}, m.Time())
				m.Conf(cmd, []string{"terminal", "hash", name, "status"}, "logout")

			case "historys":
				vs := strings.SplitN(strings.TrimSpace(m.Option("arg")), " ", 2)
				m.Grow(cmd, "history", map[string]interface{}{
					"time":  m.Time(),
					"sid":   m.Option("sid"),
					"index": vs[0],
					"cmd":   kit.Select("", vs, 1),
					"pwd":   m.Option("pwd"),
				})
			case "history":
				switch path.Base(m.Option("SHELL")) {
				case "zsh":
					m.Option("arg", strings.SplitN(m.Option("arg"), ";", 2)[1])
				}
				m.Grow(cmd, "history", map[string]interface{}{
					"time": m.Time(),
					"sid":  m.Option("sid"),
					"cmd":  m.Option("arg"),
					"pwd":  m.Option("pwd"),
				})
			case "sync":
				m.Conf(cmd, []string{m.Option("arg"), "hash", m.Option("sid")})
				switch m.Option("arg") {
				case "df":
					m.Split(m.Option("sub"), " ", "6", "fs size used rest per pos").Table(func(index int, value map[string]string) {
						if index > 0 {
							m.Confv(cmd, []string{m.Option("arg"), "hash", m.Option("sid"), "-2"}, map[string]interface{}{
								"fs":   value["fs"],
								"size": value["size"],
								"used": value["used"],
								"rest": value["rest"],
								"per":  value["per"],
								"pos":  value["pos"],
							})
						}
					})
				case "ps":
					m.Split(m.Option("sub"), " ").Table(func(index int, value map[string]string) {
						m.Confv(cmd, []string{m.Option("arg"), "hash", m.Option("sid"), "-2"}, map[string]interface{}{
							"PID":     value["PID"],
							"TIME":    value["TIME"],
							"COMMAND": value["COMMAND"],
						})
					})
				case "env":
					m.Log("fuck", "what %v fuck", m.Option("sub"))
					m.Split(strings.TrimPrefix(m.Option("sub"), "\n"), "=", "2", "name value").Table(func(index int, value map[string]string) {
						m.Log("fuck", "what %v fuck", value)
						m.Confv(cmd, []string{m.Option("arg"), "hash", m.Option("sid"), "-2"}, map[string]interface{}{
							"name":  value["name"],
							"value": value["value"],
						})
					})
				case "free":
					sub := strings.Replace(m.Option("sub"), "    ", "type", 1)
					m.Split(sub, " ", "7", "type total used free shared buffer available").Table(func(index int, value map[string]string) {
						if index > 0 {
							m.Confv(cmd, []string{m.Option("arg"), "list", "-2"}, map[string]interface{}{
								"time":      m.Time(),
								"sid":       m.Option("sid"),
								"type":      value["type"],
								"total":     value["total"],
								"used":      value["used"],
								"free":      value["free"],
								"shared":    value["shared"],
								"buffer":    value["buffer"],
								"available": value["available"],
							})
						}
					})
				}
				m.Set("append").Set("result")
			}
			return
		}},
		"zsh": {Name: "zsh dir grep key [split reg fields] [filter reg fields] [order key method] [group keys method] [sort keys method]",
			Form: map[string]int{"split": 2, "filter": 2, "order": 2, "group": 2, "sort": 2, "limit": 2},
			Help: "终端", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
				if len(arg) == 0 {
					arg = append(arg, "")
				}
				p, arg := kit.Select(".", arg[0]), arg[1:]
				switch arg[0] {
				case "prune":
					ps := []string{}
					m.Confm(cmd, "terminal.hash", func(key string, value map[string]interface{}) {
						if kit.Format(value["status"]) == "logout" {
							ps = append(ps, key)
						}
					})
					for _, v := range ps {
						for _, k := range []string{"terminal"} {
							m.Log("info", "prune %v %v %v %v", cmd, k, v, kit.Formats(m.Conf(cmd, []string{k, "hash", v})))
							m.Confv(cmd, []string{k, "hash", v}, "")
						}
					}
					fallthrough
				case "terminal":
					fields := strings.Split(kit.Select(m.Conf(cmd, arg[0]+".meta.fields"), arg, 1), " ")
					m.Confm(cmd, arg[0]+".hash", func(key string, value map[string]interface{}) {
						m.Push(fields, kit.Shortm(value, "times", "files", "sids"))
					})
					if m.Appends("times") {
						m.Sort("times", "time_r")
					}
					if m.Appends("time") {
						m.Sort("time", "time_r")
					}
					m.Table()
				case "history":
					if len(arg) > 3 {
						arg[3] = strings.Join(arg[3:], " ")
					}
					m.Option("cache.limit", kit.Select("10", arg, 1))
					m.Option("cache.offset", kit.Select("0", arg, 2))
					fields := strings.Split(kit.Select(m.Conf(cmd, arg[0]+".meta.fields"), arg, 3), " ")
					m.Grows(cmd, arg[0], func(meta map[string]interface{}, index int, value map[string]interface{}) {
						m.Push(fields, kit.Shortm(value, "times", "files", "sids"))
					})
					if m.Appends("index") {
						m.Sort("index", "int_r")
					} else if m.Appends("time") {
						m.Sort("time", "time_r")
					} else if m.Appends("times") {
						m.Sort("times", "time_r")
					}
					m.Table()
				case "env", "ps", "df":
					if len(arg) > 3 {
						arg[3] = strings.Join(arg[3:], " ")
					}
					fields := strings.Split(kit.Select(m.Conf(cmd, arg[0]+".meta.fields"), arg, 3), " ")
					m.Confm(cmd, []string{arg[0], "hash"}, func(key string, index int, value map[string]interface{}) {
						if value["sid"] = key; len(arg) == 1 || arg[1] == "" || strings.HasPrefix(kit.Format(value[arg[1]]), arg[2]) {
							m.Push(fields, kit.Shortm(value, "times", "files", "sids"))
						}
					})
					if arg[0] == "df" {
						m.Sort("size", "int_r")
					}
					m.Table()
				case "free":
					if len(arg) > 3 {
						arg[3] = strings.Join(arg[3:], " ")
					}
					m.Option("cache.limit", kit.Select("10", arg, 1))
					m.Option("cache.offset", kit.Select("0", arg, 2))
					fields := strings.Split(kit.Select(m.Conf(cmd, arg[0]+".meta.fields"), arg, 3), " ")
					m.Grows(cmd, arg[0], func(meta map[string]interface{}, index int, value map[string]interface{}) {
						m.Push(fields, kit.Shortm(value, "times", "files", "sids"))
					})

				case "init":
					m.Cmd("cli.system", m.Confv("package", "upadte"))
					for _, v := range kit.View(arg[1:], m.Confm("package")) {
						m.Cmd("cli.system", m.Confv("package", "install"), v)
					}

				case "list":
					m.Cmdy("nfs.dir", p, "time", "size", "path").Sort("time", "time_r").Table()

				case "find":
					m.Cmdy("cli.system", "find", p, "-name", arg[1], "cmd_parse", "cut", "", "1", "path")

				case "tail":
					m.Cmdy("cli.system", "tail", path.Join(p, arg[1]))

				case "grep":
					s, _ := os.Stat(p)
					prefix := []string{"cli.system", "grep", "-rn", arg[1], p, "cmd_parse", "cut", ":", kit.Select("2", "3", s.IsDir()), kit.Select("line text", "path line text", s.IsDir())}
					if m.Options("split") {
						re, _ := regexp.Compile(kit.Select("", m.Optionv("split"), 0))
						fields := map[string]bool{}
						for _, v := range strings.Split(kit.Select("", m.Optionv("split"), 1), " ") {
							if v != "" {
								fields[v] = true
							}
						}

						m.Cmd(prefix).Table(func(index int, line map[string]string) {
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
						m.Cmdy(prefix)
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
					if m.Has("limit") {
						m.Limit(kit.Int(kit.Select("0", m.Optionv("limit"), 0)), kit.Int(kit.Select("10", m.Optionv("limit"), 1)))
					}

				default:
					m.Cmdy("cli.system", arg)
				}
				return
			}},
		"tmux": {Name: "tmux [session [window [pane cmd]]]", Help: "窗口", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			prefix := kit.Trans(m.Confv("prefix", "tmux"))
			// 修改信息
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
				// 修改缓存
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
				// 操作缓存
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
				// 面板列表
				case "pane":
					m.Cmdy(prefix, "list-panes", "-a", "cmd_parse", "cut", " ", "8", "pane_name size some lines bytes haha pane_id tag")
					m.Meta["append"] = []string{"pane_id", "pane_name", "size", "lines", "bytes", "tag"}
					m.Table(func(index int, value map[string]string) {
						m.Meta["pane_name"][index] = strings.TrimSuffix(value["pane_name"], ":")
						m.Meta["pane_id"][index] = strings.TrimPrefix(value["pane_id"], "%")
						m.Meta["lines"][index] = strings.TrimSuffix(value["lines"], ",")
						m.Meta["bytes"][index] = kit.FmtSize(kit.Int64(value["bytes"]))
					})
					m.Sort("pane_name")
					m.Table()
					return
				// 运行命令
				case "run":
					arg = arg[1:]
					fallthrough
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
			prefix := kit.Trans(m.Confv("prefix", "docker"))
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
					m.Cmdy(prefix[:2], "run", "-dt", pos+":"+tag)
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
							m.Cmdy(prefix, "build", "-f", m.Option("file"), "-t", m.Option("name"), ".")
						}
					}
				}

			case "container":
				prefix = append(prefix, "container")
				if len(arg) > 1 {
					switch arg[2] {
					case "进入":
						m.Cmdy(m.Confv("prefix", "tmux"), "new-window", "-dPF", "#{session_name}:#{window_name}.1", "docker exec -it "+arg[1]+" sh")
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
				prefix = append(prefix, "network")
				if len(arg) == 1 {
					m.Cmdy(prefix, "ls", "cmd_parse", "cut", "cmd_headers", "NETWORK ID", "NETWORK_ID")
					break
				}

				kit.Map(kit.Chain(kit.UnMarshal(m.Cmdx(prefix, "inspect", arg[1])), "0.Containers"), "", func(key string, value map[string]interface{}) {
					m.Push("CONTAINER_ID", key[:12])
					m.Push("name", value["Name"])
					m.Push("IPv4", value["IPv4Address"])
					m.Push("IPv6", value["IPV4Address"])
					m.Push("Mac", value["MacAddress"])
				})
				m.Table()

			case "volume":
				if len(arg) == 1 {
					m.Cmdy(prefix, "volume", "ls", "cmd_parse", "cut", "cmd_headers", "VOLUME NAME", "VOLUME_NAME")
					break
				}

			default:
				m.Cmdy(prefix, arg)
			}
			return
		}},
		"git": {Name: "git init|diff|status|commit|branch|remote|pull|push|sum", Help: "版本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			prefix, arg := append(kit.Trans(m.Confv("prefix", "git")), "cmd_dir", kit.Select(".", arg[0])), arg[1:]

			switch arg[0] {
			case "init":
				if s, e := os.Stat(path.Join(prefix[len(prefix)-1], ".git")); e == nil && s.IsDir() {
					if len(arg) > 1 {
						m.Cmdy(prefix, "remote", "add", "-f", kit.Select("origin", arg, 2), arg[1])
					}
				} else if len(arg) > 1 {
					m.Cmdy(prefix, "clone", arg[1], ".")
				} else {
					m.Cmdy(prefix, "init")
				}

				m.Confm("git", "alias", func(key string, value string) {
					m.Cmdy(prefix, "config", "alias."+key, value)
				})

			case "diff":
				m.Cmdy(prefix, "diff")
			case "status":
				m.Cmdy(prefix, "status", "-sb", "cmd_parse", "cut", " ", "2", "tags file")
			case "commit":
				if len(arg) > 1 && m.Cmdy(prefix, "commit", "-am", arg[1]).Result() == "" {
					break
				}
				m.Cmdy(prefix, "log", "--stat", "-n", "3")
			case "branch":
				if len(arg) > 1 {
					m.Cmd(prefix, "branch", arg[1])
					m.Cmd(prefix, "checkout", arg[1])
				}
				for _, v := range strings.Split(m.Cmdx(prefix, "branch", "-v"), "\n") {
					if len(v) > 0 {
						m.Push("tags", v[:2])
						vs := strings.SplitN(strings.TrimSpace(v[2:]), " ", 2)
						m.Push("branch", vs[0])
						vs = strings.SplitN(strings.TrimSpace(vs[1]), " ", 2)
						m.Push("hash", vs[0])
						m.Push("note", strings.TrimSpace(vs[1]))
					}
				}
				m.Table()
			case "remote":
				m.Cmdy(prefix, "remote", "-v", "cmd_parse", "cut", " ", "3", "remote url tag")

			case "push":
				m.Cmdy(prefix, "push")
			case "sum":
				total := false
				if len(arg) > 1 && arg[1] == "total" {
					total, arg = true, arg[1:]
				}

				args := []string{"log", "--shortstat", "--pretty=commit: %ad %n%s", "--date=iso", "--reverse"}
				if len(arg) > 1 {
					args = append(args, kit.Select("-n", "--since", strings.Contains(arg[1], "-")))
					if strings.Contains(arg[1], "-") && !strings.Contains(arg[1], ":") {
						arg[1] = arg[1] + " 00:00:00"
					}
					args = append(args, arg[1:]...)
				} else {
					args = append(args, "-n", "30")
				}

				var total_day time.Duration
				count, count_add, count_del := 0, 0, 0
				if out, e := exec.Command("git", args...).CombinedOutput(); e == nil {
					for i, v := range strings.Split(string(out), "commit: ") {
						if i > 0 {
							l := strings.Split(v, "\n")
							hs := strings.Split(l[0], " ")

							add, del := "0", "0"
							if len(l) > 3 {
								fs := strings.Split(strings.TrimSpace(l[3]), ", ")
								if adds := strings.Split(fs[1], " "); len(fs) > 2 {
									dels := strings.Split(fs[2], " ")
									add = adds[0]
									del = dels[0]
								} else if adds[1] == "insertions(+)" {
									add = adds[0]
								} else {
									del = adds[0]
								}
							}

							if total {
								if count++; i == 1 {
									if t, e := time.Parse("2006-01-02", hs[0]); e == nil {
										total_day = time.Now().Sub(t)
										m.Append("from", hs[0])
									}
								}
								count_add += kit.Int(add)
								count_del += kit.Int(del)
								continue
							}

							m.Push("date", hs[0])
							m.Push("adds", add)
							m.Push("dels", del)
							m.Push("rest", kit.Int(add)-kit.Int(del))
							m.Push("note", l[1])
							m.Push("hour", strings.Split(hs[1], ":")[0])
							m.Push("time", hs[1])
						}
					}
					if total {
						m.Append("days", int(total_day.Hours())/24)
						m.Append("commit", count)
						m.Append("adds", count_add)
						m.Append("dels", count_del)
						m.Append("rest", count_add-count_del)
					}
					m.Table()
				} else {
					m.Log("warn", "%v", string(out))
				}

			default:
				m.Cmdy(prefix, arg)
			}
			return
		}},
		"tags": {Name: "tags", Help: "代码索引", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy("cli.system", "gotags", "-f", kit.Select("tags", arg, 1), "-R", kit.Select("src", arg, 0))
			return
		}},
		"vim": {Name: "vim editor|prune|opens|cmds|txts|bufs|regs|marks|tags|fixs", Help: "编辑器", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			switch arg[0] {
			case "favor":

			case "ctag":
				if f, p, e := kit.Create("etc/conf/tags"); m.Assert(e) {
					defer f.Close()
					for k, _ := range c.Commands {
						fmt.Fprintf(f, "%s\t%s\t/\"%s\": {Name/\n", k, "../../src/examples/code/code.go", k)
					}
					m.Echo(p)
				}
				return

			case "prune":
				ps := []string{}
				m.Confm(cmd, "editor.hash", func(key string, value map[string]interface{}) {
					if kit.Format(value["status"]) == "logout" {
						ps = append(ps, key)
					}
				})
				for _, v := range ps {
					for _, k := range []string{"editor", "bufs", "regs", "marks", "tags", "fixs"} {
						m.Log("info", "prune %v %v %v %v", cmd, k, v, kit.Formats(m.Conf(cmd, []string{k, "hash", v})))
						m.Confv(cmd, []string{k, "hash", v}, "")
					}
				}
				fallthrough
			case "editor":
				if len(arg) > 2 && arg[2] != "" {
					m.Conf(cmd, []string{arg[0], "hash", arg[1], "table"}, arg[2])
					m.Conf(cmd, []string{arg[0], "hash", arg[1], "river"}, m.Option("river"))
				}
				if len(arg) > 1 && arg[1] != "" {
					m.Option("table.format", "table")
					m.Confm(cmd, []string{arg[0], "hash", arg[1]}, func(key string, value string) {
						m.Push(key, value)
					})
					m.Sort("key")
					break
				}
				if len(arg) > 3 {
					arg[3] = strings.Join(arg[3:], " ")
				}
				fields := strings.Split(kit.Select(m.Conf(cmd, arg[0]+".meta.fields"), arg, 1), " ")
				m.Confm(cmd, arg[0]+".hash", func(key string, value map[string]interface{}) {
					m.Push(fields, kit.Shortm(value, "times", "files", "sids"))
				})

			case "opens", "cmds", "txts":
				if len(arg) > 3 {
					arg[3] = strings.Join(arg[3:], " ")
				}
				m.Option("cache.limit", kit.Select("10", arg, 1))
				m.Option("cache.offset", kit.Select("0", arg, 2))
				fields := strings.Split(kit.Select(m.Conf(cmd, arg[0]+".meta.fields"), arg, 3), " ")
				m.Grows(cmd, arg[0], func(meta map[string]interface{}, index int, value map[string]interface{}) {
					m.Push(fields, kit.Shortm(value, "times", "files", "sids"))
				})

			case "bufs", "regs", "marks", "tags", "fixs":
				if len(arg) > 3 {
					arg[3] = strings.Join(arg[3:], " ")
				}
				fields := strings.Split(kit.Select(m.Conf(cmd, arg[0]+".meta.fields"), arg, 3), " ")
				m.Confm(cmd, []string{arg[0], "hash"}, func(key string, index int, value map[string]interface{}) {
					if value["sid"] = key; len(arg) == 1 || arg[1] == "" || strings.HasPrefix(kit.Format(value[arg[1]]), arg[2]) {
						m.Push(fields, kit.Shortm(value, "times", "files", "sids"))
					}
				})
			}
			if m.Appends("times") {
				m.Sort("times", "time_r")
			}
			if m.Appends("time") {
				m.Sort("time", "time_r")
			}
			m.Table()
			return
		}},
		"/vim": {Name: "/vim sid pwd cmd arg sub", Help: "编辑器", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			cmd = strings.TrimPrefix(cmd, "/")
			m.Option("arg", strings.Replace(m.Option("arg"), "XXXXXsingleXXXXX", "'", -1))
			m.Option("sub", strings.Replace(m.Option("sub"), "XXXXXsingleXXXXX", "'", -1))
			m.Log("info", "%v %v %v %v", cmd, m.Option("cmd"), m.Option("arg"), m.Option("sub"))

			switch m.Option("cmd") {
			case "login":
				name := kit.Hashs(m.Option("pid"), m.Option("hostname"), m.Option("username"))
				m.Conf(cmd, []string{"editor", "hash", name}, map[string]interface{}{
					"time":     m.Time(),
					"status":   "login",
					"sid":      name,
					"pwd":      m.Option("pwd"),
					"pid":      m.Option("pid"),
					"pane":     m.Option("pane"),
					"hostname": m.Option("hostname"),
					"username": m.Option("username"),
				})
				m.Echo(name)
			case "logout":
				m.Conf(cmd, []string{"editor", "hash", m.Option("sid"), "time"}, m.Time())
				m.Conf(cmd, []string{"editor", "hash", m.Option("sid"), "status"}, "logout")

			case "favors":
				m.Option("river", m.Conf(cmd, []string{"editor", "hash", m.Option("sid"), "river"}))
				table := m.Conf(cmd, []string{"editor", "hash", m.Option("sid"), "table"})
				m.Log("info", "favor: %v table: %v", m.Option("river"), table)

				data := map[string][]string{}
				m.Cmd("ssh.data", "show", table).Table(func(index int, value map[string]string) {
					data[value["tab"]] = append(data[value["tab"]],
						fmt.Sprintf("%v:%v:0:(%v): %v\n", value["file"], value["line"], value["note"], value["word"]))
				})

				for k, v := range data {
					m.Push("tab", k)
					m.Push("fix", strings.Join(v, "\n"))
				}
				return
			case "favor":
				m.Option("river", m.Conf(cmd, []string{"editor", "hash", m.Option("sid"), "river"}))
				table := m.Conf(cmd, []string{"editor", "hash", m.Option("sid"), "table"})
				m.Log("info", "favor: %v table: %v", m.Option("river"), table)

				if m.Options("line") {
					m.Cmd("ssh.data", "insert", table,
						"tab", m.Option("tab"),
						"note", m.Option("note"), "word", m.Option("arg"),
						"file", m.Option("buf"), "line", m.Option("line"), "col", m.Option("col"),
					)
					return
				}
				m.Cmd("ssh.data", "show", table).Table(func(index int, value map[string]string) {
					m.Echo("%v:%v:0:(%v): %v\n", value["file"], value["line"], value["note"], value["word"])
				}).Set("append")
				return

			case "read", "write":
				m.Grow(cmd, "opens", map[string]interface{}{
					"time":   m.Time(),
					"sid":    m.Option("sid"),
					"action": m.Option("cmd"),
					"file":   m.Option("arg"),
					"pwd":    m.Option("pwd"),
				})
			case "exec":
				m.Grow(cmd, "cmds", map[string]interface{}{
					"time": m.Time(),
					"sid":  m.Option("sid"),
					"cmd":  m.Option("arg"),
					"file": m.Option("buf"),
					"pwd":  m.Option("pwd"),
				})
			case "insert":
				m.Grow(cmd, "txts", map[string]interface{}{
					"time": m.Time(),
					"sid":  m.Option("sid"),
					"word": m.Option("arg"),
					"line": m.Option("row"),
					"col":  m.Option("col"),
					"file": m.Option("buf"),
					"pwd":  m.Option("pwd"),
				})

			case "sync":
				m.Conf(cmd, []string{m.Option("arg"), "hash", m.Option("sid")}, "")
				switch m.Option("arg") {
				case "bufs":
					m.Split(strings.TrimSpace(m.Option("sub")), " ", "5", "id tag name some line").Table(func(index int, value map[string]string) {
						m.Confv(cmd, []string{m.Option("arg"), "hash", m.Option("sid"), "-2"}, map[string]interface{}{
							"buf":  value["id"],
							"tag":  value["tag"],
							"file": strings.TrimSuffix(strings.TrimPrefix(value["name"], "\""), "\""),
							"line": value["line"],
						})
					})
				case "regs":
					m.Split(strings.TrimPrefix(m.Option("sub"), "\n--- Registers ---\n"), " ", "2", "name word").Table(func(index int, value map[string]string) {
						m.Confv(cmd, []string{m.Option("arg"), "hash", m.Option("sid"), "-2"}, map[string]interface{}{
							"word": strings.Replace(strings.Replace(value["word"], "^I", "\t", -1), "^J", "\n", -1),
							"reg":  strings.TrimPrefix(value["name"], "\""),
						})
					})
				case "marks":
					m.Split(strings.TrimPrefix(m.Option("sub"), "\n"), " ", "4").Table(func(index int, value map[string]string) {
						m.Confv(cmd, []string{m.Option("arg"), "hash", m.Option("sid"), "-2"}, map[string]interface{}{
							"mark": value["mark"],
							"line": value["line"],
							"col":  value["col"],
							"file": value["file/text"],
						})
					})
				case "tags":
					m.Split(strings.TrimPrefix(m.Option("sub"), "\n"), " ", "6").Table(func(index int, value map[string]string) {
						m.Confv(cmd, []string{m.Option("arg"), "hash", m.Option("sid"), "-2"}, map[string]interface{}{
							"tag":  value["tag"],
							"line": value["line"],
							"file": value["in file/text"],
						})
					})
				case "fixs":
					if strings.HasPrefix(m.Option("sub"), "\nError") {
						break
					}
					m.Split(strings.TrimPrefix(m.Option("sub"), "\n"), " ", "3", "id file word").Table(func(index int, value map[string]string) {
						vs := strings.Split(kit.Format(value["file"]), ":")
						m.Confv(cmd, []string{m.Option("arg"), "hash", m.Option("sid"), "-2"}, map[string]interface{}{
							"fix":  value["id"],
							"file": vs[0],
							"line": vs[1],
							"word": value["word"],
						})
					})
				}
				m.Set("append").Set("result")
			}
			return
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.WEB{Context: Index})
}
