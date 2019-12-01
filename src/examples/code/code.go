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
		"prefix": {Name: "prefix", Help: "外部命令", Value: map[string]interface{}{
			"zsh":    []interface{}{"cli.system", "zsh"},
			"tmux":   []interface{}{"cli.system", "tmux"},
			"docker": []interface{}{"cli.system", "docker"},
			"git":    []interface{}{"cli.system", "git"},
			"vim":    []interface{}{"cli.system", "vim"},
		}},
		"package": {Name: "package", Help: "软件包", Value: map[string]interface{}{
			"apk": map[string]interface{}{"update": "update", "install": "add",
				"base": []interface{}{"curl", "bash"},
			},
			"apt": map[string]interface{}{"update": "update", "install": "install -y",
				"base": []interface{}{"wget", "curl"},
			},
			"yum": map[string]interface{}{"update": "update -y", "install": "install",
				"base": []interface{}{"wget"},
			},

			"udpate":  []interface{}{"apk", "update"},
			"install": []interface{}{"apk", "add"},
			"build":   []interface{}{"build-base"},
			"develop": []interface{}{"zsh", "tmux", "git", "vim", "golang"},
			"product": []interface{}{"nginx", "redis", "mysql"},
		}},
		"docker": {Name: "docker", Help: "容器", Value: map[string]interface{}{
			"template": map[string]interface{}{"shy": Dockfile},
			"output":   "etc/Dockerfile",
		}},
		"tmux": {Name: "tmux", Help: "终端", Value: map[string]interface{}{
			"favor": map[string]interface{}{
				"index": []interface{}{
					"ctx_dev ctx_share",
					"curl -s $ctx_dev/publish/auto.sh >auto.sh",
					"source auto.sh",
					"ShyLogin",
				},
			},
		}},
		"git": {Name: "git", Help: "记录", Value: map[string]interface{}{
			"alias": map[string]interface{}{"s": "status", "b": "branch"},
		}},
		"cache": {Name: "cache", Help: "缓存默认的全局配置", Value: map[string]interface{}{
			"store": "var/tmp/hi.csv",
			"limit": 6,
			"least": 3,
		}},

		"zsh": {Name: "zsh", Help: "命令行", Value: map[string]interface{}{
			"history": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "time sid cmd pwd",
				"store":  "var/tmp/zsh/history.csv",
				"limit":  "30",
				"least":  "10",
			}},
			"free": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "time sid type total used free shared buffer available",
				"store":  "var/tmp/zsh/free.csv",
				"limit":  "30",
				"least":  "10",
			}},
			"env": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "sid name value",
			}},
			"ps": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "sid UID PID PPID TTY CMD",
			}},
			"df": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "sid fs size used rest per pos",
			}},
		}},
		"vim": {Name: "vim", Help: "编辑器", Value: map[string]interface{}{
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
				"fields": "buf tag file line",
			}},
			"regs": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "reg word",
			}},
			"marks": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "mark line col file",
			}},
			"tags": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "tag line file",
			}},
			"fixs": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "fix file line word",
			}},
		}},

		"login": {Name: "login", Value: map[string]interface{}{"check": false, "local": true, "expire": "720h", "meta": map[string]interface{}{
			"fields": "time sid type status ship.dream ship.stage pwd pid pane hostname username",
			"script": "usr/script",
		}, "hash": map[string]interface{}{}}, Help: "用户登录"},
		"dream": {Name: "dream", Help: "使命必达", Value: map[string]interface{}{
			"layout": map[string]interface{}{
				"three": []interface{}{
					"rename-window -t $dream:1 source",
					"split-window -t $dream:1.1",
					"split-window -v -t $dream:1.2",
					"select-layout -t $dream:1 main-horizontal",

					"new-window -t $dream:2 -n docker",
					"split-window -t $dream:2.1",
					"split-window -v -t $dream:2.2",
					"select-layout -t $dream:2 main-horizontal",

					"new-window -t $dream:3 -n online",
					"split-window -t $dream:3.1",
					"split-window -v -t $dream:3.2",
					"select-layout -t $dream:3 main-horizontal",
				},
			},
			"topic": map[string]interface{}{
				"index": map[string]interface{}{
					"ship": []interface{}{"tip", "miss.md", "task", "feed"},
					"git":  []interface{}{},
					// "git":    []interface{}{"clone https://github.com/shylinux/context"},
					"layout": []interface{}{"three"}, "tmux": []interface{}{
						"send-keys -t $dream.1.1 pwd",
					},
				},
			},
			"share": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "river dream favor story stage order expire",
			}, "hash": map[string]interface{}{}},
			"favor": map[string]interface{}{"meta": map[string]interface{}{
				"fields": "tab note word file line col",
			}},
		}},
	},
	Commands: map[string]*ctx.Command{
		"dream": {Name: "dream init name [topic]", Help: "使命必达", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			switch arg[0] {
			case "init":
				// 检查会话
				tmux := kit.Trans(m.Confv("prefix", "tmux"))
				if !m.Cmds(tmux, "has-session", "-t", arg[1]) {
					break
				}

				// 下载代码
				home := path.Join(m.Conf("missyou", "path"), arg[1], m.Conf("missyou", "local"))
				topic := kit.Select("index", kit.Select(m.Option("topic"), arg, 2))
				git := kit.Trans(m.Confv("prefix", "git"), "cmd_dir", home)
				m.Confm(cmd, []string{"topic", topic, "git"}, func(index int, value string) {
					value = strings.Replace(value, "$dream", arg[1], -1)
					m.Cmd(git, strings.Split(value, " "))
				})

				// 创建文档
				m.Cmd("cli.system", "mkdir", "-p", path.Join(home, "wiki"))
				m.Cmd("cli.system", "touch", path.Join(home, "wiki/miss.md"))

				// 创建终端
				share := m.Cmdx(cmd, "share", topic)
				m.Cmd(tmux, "set-environment", "-g", "ctx_share", share)
				m.Cmd(tmux, "new-session", "-ds", arg[1], "cmd_dir", home, "cmd_env", "TMUX", "")
				m.Cmd(tmux, "set-environment", "-t", arg[1], "ctx_share", share)
				m.Cmd(tmux, "set-environment", "-t", arg[1], "ctx_dev", os.Getenv("ctx_self"))
				m.Confm(cmd, []string{"layout", m.Conf(cmd, []string{"topic", topic, "layout", "0"})}, func(index int, value string) {
					value = strings.Replace(value, "$dream", arg[1], -1)
					m.Cmd(tmux, strings.Split(value, " "), "cmd_dir", home)
				})
				m.Confm(cmd, []string{"topic", topic, "tmux"}, func(index int, value string) {
					value = strings.Replace(value, "$dream", arg[1], -1)
					m.Cmd(tmux, strings.Split(value, " "), "cmd_dir", home)
				})
				m.Echo(share)

			case "share":
				// 模板参数
				topic := kit.Select("index", kit.Select(m.Option("topic"), arg, 1))
				m.Confm(cmd, []string{"topic", topic, "ship"}, func(index int, value string) {
					if len(arg) < index+3 {
						arg = append(arg, value)
					} else if arg[index+2] == "" {
						arg[index+2] = value
					}
				})
				// 共享链接
				h := kit.ShortKey(m.Confm(cmd, []string{"share", "hash"}), 6)
				m.Confv(cmd, []string{"share", "hash", h}, map[string]interface{}{
					"river": m.Option("river"), "dream": m.Option("dream"),
					"favor": arg[2], "story": arg[3], "stage": arg[4], "order": arg[5],
					"topic": topic, "share": h, "expire": m.Time("10m", "stamp"),
				})
				m.Echo(h)

			case "list":
				m.Confm(cmd, "share.hash", func(key string, value map[string]interface{}) {
					m.Push("key", key).Push(m.Conf(cmd, "share.meta.fields"), value)
				})
				m.Table()

			case "exit":
			}
			return
		}},
		"login": {Name: "login open|init|list|exit|quit", Help: "登录", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {

			switch kit.Select("list", arg, 0) {
			case "open":
			case "init":
				if m.Option("sid") != "" {
					if m.Confs(cmd, []string{"hash", m.Option("sid"), "status"}) {
						m.Conf(cmd, []string{"hash", m.Option("sid"), "status"}, "login")
						m.Echo(m.Option("sid"))
						return
					}
				}

				// 添加终端
				h := kit.ShortKey(m.Confm(cmd, "hash"), 6, m.Option("pid"), m.Option("hostname"), m.Option("username"))
				m.Conf(cmd, []string{"hash", h}, map[string]interface{}{
					"time":     m.Time(),
					"status":   "login",
					"type":     kit.Select("vim", arg, 1),
					"ship":     m.Confv("dream", []string{"share", "hash", m.Option("share")}),
					"pwd":      m.Option("pwd"),
					"pid":      m.Option("pid"),
					"pane":     m.Option("pane"),
					"hostname": m.Option("hostname"),
					"username": m.Option("username"),
				})
				m.Echo(h)

			case "list":
				if len(arg) > 4 {
					sid := kit.Select(m.Option("sid"), arg[3])
					switch arg[4] {
					case "prune":
						// 清理终端
						m.Cmd(".prune", m.Conf("login", []string{"hash", sid, "type"}), sid)
						arg = arg[:2]
					case "modify":
						// 修改终端
						m.Conf(cmd, []string{"hash", sid, arg[5]}, arg[6])
						arg = arg[:2]
					}
				}

				// 终端列表
				if len(arg) == 2 || arg[2] == "" {
					fields := strings.Split(m.Conf(cmd, "meta.fields"), " ")
					m.Confm(cmd, "hash", func(key string, value map[string]interface{}) {
						if arg[1] == "" || arg[1] == kit.Format(value["type"]) {
							value["sid"] = key
							m.Push(fields, value)
						}
					})
					m.Table()
					break
				}

				// 终端绑定
				if len(arg) > 4 && arg[4] != "" && arg[4] != m.Conf(cmd, []string{"hash", arg[2], "ship", "topic"}) && m.Confs(cmd, []string{"hash", arg[2]}) {
					share := m.Cmdx("dream", "share", arg[4:])
					m.Conf(cmd, []string{"hash", arg[2], "ship"},
						m.Confv("dream", []string{"share", "hash", share}))
				}
				if len(arg) > 3 && arg[3] != "" {
					m.Conf(cmd, []string{"hash", arg[2], "ship", "dream"}, arg[3])
				}

				// 终端详情
				m.Option("table.format", "table")
				m.Confm(cmd, []string{"hash", arg[2], "ship"}, func(key string, value string) {
					m.Push("ship."+key, value)
				})
				m.Confm(cmd, []string{"hash", arg[2]}, func(key string, value string) {
					if key != "ship" {
						m.Push(key, value)
					}
				})
				m.Sort("key")

			case "exit":
				// 退出终端
				if m.Confs(cmd, []string{"hash", m.Option("sid")}) {
					m.Conf(cmd, []string{"hash", m.Option("sid"), "status"}, "logout")
					m.Conf(cmd, []string{"hash", m.Option("sid"), "time"}, m.Time())
				}
			case "quit":
			}
			return
		}},
		"favor": {Name: "favor download|upload|file|post|list", Help: "收藏", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			switch arg[0] {
			case "download":
				// 下载文件
				if len(arg) > 1 && arg[1] != "" {
					m.Cmd("/download/", "", arg[1])
					break
				}
				// 下载列表
				m.Cmd("ssh._route", m.Option("dream"), "ssh.data", "show", "file").Table(func(index int, value map[string]string) {
					m.Push("hash", value["hash"])
					m.Push("time", value["upload_time"])
					m.Push("size", kit.FmtSize(int64(kit.Int(value["size"]))))
					m.Push("name", value["name"])
				})
				m.Table().Set("append")

			case "upload":
				// 上传文件
				if m.Cmd("/upload"); m.Options("dream") {
					// 下发文件
					m.Cmd("ssh._route", m.Option("dream"), "web.get", "dev", "/download/"+m.Append("hash"),
						"save", m.Conf("login", "meta.script")+"/"+m.Append("name"))
				}
				m.Echo("code: %s\n", m.Append("code"))
				m.Echo("hash: %s\n", m.Append("hash"))
				m.Echo("time: %s\n", m.Append("time"))
				m.Echo("type: %s\n", m.Append("type"))
				m.Echo("size: %s\n", m.Append("size"))
				m.Set("append")

			case "file":
				// 文件列表
				if len(arg) > 2 && arg[2] != "" {
					m.Cmdy("ssh.data", "show", arg[1:3])
					break
				}
				m.Cmd("ssh.data", "show", arg[1:]).Table(func(index int, value map[string]string) {
					m.Push("id", value["id"])
					m.Push("time", value["upload_time"])
					m.Push("name", value["name"])
					m.Push("size", kit.FmtSize(int64(kit.Int(value["size"]))))
					m.Push("file", fmt.Sprintf(`<a href="/download/%s" target="_blank">%s</a>`, kit.Select(value["hash"], value["code"]), value["name"]))
					m.Push("hash", kit.Short(value["hash"], 6))
				})
				m.Table()

			case "post":
				// 上传记录
				if len(arg) < 3 || arg[2] == "" {
					break
				}
				args := []string{}
				for i, v := range strings.Split(kit.Select("tab note word file line", m.Conf("dream", "favor.meta.fields")), " ") {
					args = append(args, v, kit.Select("", arg, i+2))
				}
				m.Cmdy("ssh.data", "insert", arg[1], args)

			case "list":
				if len(arg) > 3 && arg[3] == "modify" {
					m.Cmd("ssh.data", "update", arg[1], arg[2], arg[4], arg[5])
					arg = arg[:3]
				}
				m.Cmdy("ssh.data", "show", arg[1:])
			}
			return
		}},
		"trend": {Name: "trend post|list", Help: "趋势", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			switch arg[0] {
			case "post":
				m.Cmdy("ssh.data", "insert", arg[1], arg[2:])

			case "list":
				if len(arg) > 3 && arg[3] == "modify" {
					m.Cmd("ssh.data", "update", arg[1], arg[2], arg[4], arg[5])
					arg = arg[:3]
				}
				m.Cmdy("ssh.data", "show", arg[1:])
			}
			return
		}},
		"state": {Name: "state post|list agent type", Help: "状态", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			switch arg[0] {
			case "post":
				data := map[string]interface{}{}
				for i := 3; i < len(arg)-1; i += 2 {
					kit.Chain(data, arg[i], arg[i+1])
				}
				m.Confv(arg[1], []string{arg[2], "hash", m.Option("sid"), "-2"}, data)

			case "list":
				if len(arg) > 6 {
					arg[6] = strings.Join(arg[6:], " ")
				}
				fields := strings.Split(kit.Select(m.Conf(arg[1], arg[2]+".meta.fields"), arg, 6), " ")
				m.Confm(arg[1], []string{arg[2], "hash"}, func(key string, index int, value map[string]interface{}) {
					if arg[3] != "" && key != arg[3] {
						return
					}
					if value["sid"] = key; len(arg) < 6 || arg[4] == "" || strings.HasPrefix(kit.Format(value[arg[4]]), arg[5]) {
						m.Push(fields, value)
					}
				})
			}
			return
		}},
		"prune": {Name: "prune type sid...", Help: "清理", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			ps := arg[1:]
			if len(ps) == 0 {
				m.Confm("login", "hash", func(key string, value map[string]interface{}) {
					if value["type"] == arg[0] && kit.Format(value["status"]) == "logout" {
						ps = append(ps, key)
					}
				})
			}

			for _, p := range ps {
				m.Confm(arg[0], func(key string, value map[string]interface{}) {
					m.Log("info", "prune %v:%v %v:%v", arg[0], key, p, kit.Formats(kit.Chain(value, []string{"hash", p})))
					kit.Chain(value, []string{"hash", p}, "")
				})

				m.Log("info", "prune %v %v:%v", "login", p, kit.Formats(m.Confv("login", []string{"hash", p})))
				m.Confv("login", []string{"hash", p}, "")
			}
			return
		}},

		"/zsh": {Name: "/zsh sid pwd cmd arg", Help: "命令行", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			cmd = strings.TrimPrefix(cmd, "/")
			if f, _, e := m.Optionv("request").(*http.Request).FormFile("sub"); e == nil {
				defer f.Close()
				if b, e := ioutil.ReadAll(f); e == nil {
					m.Option("sub", string(b))
				}
			}
			m.Log("info", "%v %v %v %v", cmd, m.Option("cmd"), m.Option("arg"), m.Option("sub"))
			m.Confm("login", []string{"hash", m.Option("sid"), "ship"}, func(key string, value string) { m.Option(key, value) })

			switch m.Option("cmd") {
			case "help":
				m.Echo(strings.Join(kit.Trans(m.Confv("help", "index")), "\n"))
			case "login":
				m.Cmd("login", "init", cmd)
			case "logout":
				m.Cmd("login", "exit")
			case "upload":
				m.Cmd("favor", "upload")
			case "download":
				m.Cmd("favor", "download", m.Option("arg"))
			case "favor":
				m.Log("info", "river: %v dream: %v favor: %v", m.Option("river"), m.Option("dream"), m.Option("favor"))
				// 添加收藏
				prefix := []string{"ssh._route", m.Option("dream"), "web.code.favor"}
				if m.Options("arg") {
					m.Option("arg", strings.SplitN(strings.TrimSpace(m.Option("arg")), " ", 2)[1])
					m.Cmdy(prefix, "post", m.Option("favor"), m.Option("tab"), m.Option("note"), m.Option("arg"))
					m.Set("append")
					break
				}

				// 生成脚本
				m.Echo("#/bin/sh\n\n")
				m.Cmd(prefix, "list", m.Option("favor"), "", kit.Select("1000", m.Option("limit")), "0", "tab", m.Option("tab")).Table(func(index int, value map[string]string) {
					m.Echo("# %v:%v\n%v\n\n", value["tab"], value["note"], value["word"])
				})

			case "history":
				vs := strings.SplitN(strings.TrimSpace(m.Option("arg")), " ", 2)
				m.Cmd("trend", "post", "zsh.history", "sid", m.Option("sid"),
					"num", vs[0], "cmd", kit.Select("", vs, 1), "pwd", m.Option("pwd"))
				m.Set("append").Set("result")

			case "sync":
				m.Confv(cmd, []string{m.Option("arg"), "hash", m.Option("sid")}, []interface{}{})
				switch m.Option("arg") {
				case "free":
					sub := strings.Replace(m.Option("sub"), "    ", "type", 1)
					m.Split(sub, " ", "7", "type total used free shared buffer available").Table(func(index int, value map[string]string) {
						if index == 1 {
							m.Cmd("trend", "post", "zsh.free", value)
						}
					})
				case "env":
					m.Split(strings.TrimPrefix(m.Option("sub"), "\n"), "=", "2", "name value").Table(func(index int, value map[string]string) {
						m.Cmd("state", "post", cmd, m.Option("arg"), value)
					})
				case "ps":
					m.Split(m.Option("sub"), " ", "8", "UID PID PPID C STIME TTY TIME CMD").Table(func(index int, value map[string]string) {
						if index > 0 {
							m.Cmd("state", "post", cmd, m.Option("arg"), value)
						}
					})
				case "df":
					m.Split(m.Option("sub"), " ", "6", "fs size used rest per pos").Table(func(index int, value map[string]string) {
						if index > 0 {
							m.Cmd("state", "post", cmd, m.Option("arg"), value)
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
				case "init":
					m.Cmd("cli.system", m.Confv("package", "upadte"))
					for _, v := range kit.View(arg[1:], m.Confm("package")) {
						m.Cmd("cli.system", m.Confv("package", "install"), v)
					}

				case "git":
					if s, e := os.Stat(path.Join(p, ".git")); e == nil && s.IsDir() || arg[1] == "init" {
						m.Cmdy(".git", p, arg[1:])
						break
					}

					fallthrough
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
		"tmux": {Name: "tmux [session [window [pane cmd]]]", Help: "窗口", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			prefix := kit.Trans(m.Confv("prefix", "tmux"))
			if len(arg) > 1 {
				switch arg[1] {
				case "cmd":

				case "pane":
					prefix = append(prefix, "list-panes")
					if arg[0] == "" {
						prefix = append(prefix, "-a")
					} else {
						prefix = append(prefix, "-s", "-t", arg[0])
					}
					m.Cmdy(prefix, "cmd_parse", "cut", " ", "8", "pane_name size some lines bytes haha pane_id tag")

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

				case "favor":
					env := m.Cmdx(prefix, "show-environment", "-g") + m.Cmdx(prefix, "show-environment", "-t", arg[0])
					for _, l := range strings.Split(env, "\n") {
						if strings.HasPrefix(l, "ctx_") {
							v := strings.SplitN(l, "=", 2)
							m.Option(v[0], v[1])
						}
					}
					m.Option("ctx_dev", m.Option("ctx_self"))

					m.Confm("tmux", "favor."+kit.Select("index", arg, 4), func(index int, value string) {
						if index == 0 {
							keys := strings.Split(value, " ")
							value = "export"
							for _, k := range keys {
								value += " " + k + "=" + m.Option(k)
							}

						}
						m.Cmdy(prefix, "send-keys", "-t", arg[0], value, "Enter")
						time.Sleep(100 * time.Millisecond)
					})
					m.Echo(strings.TrimSpace(m.Cmdx(prefix, "capture-pane", "-pt", arg[0])))
					return

				case "buffer":
					// 写缓存
					if len(arg) > 5 {
						switch arg[3] {
						case "modify":
							switch arg[4] {
							case "text":
								m.Cmdy(prefix, "set-buffer", "-b", arg[2], arg[5])
							}
						}
					} else if len(arg) > 3 {
						m.Cmd(prefix, "set-buffer", "-b", arg[2], arg[3])
					}

					// 读缓存
					if len(arg) > 2 {
						m.Cmdy(prefix, "show-buffer", "-b", arg[2])
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

				case "select":
					// 切换会话
					if !m.Has("session") {
						m.Cmd(prefix, "switch-client", "-t", arg[0])
						arg = arg[:0]
						break
					}
					m.Cmd(prefix, "switch-client", "-t", m.Option("session"))

					// 切换窗口
					if !m.Has("window") {
						m.Cmd(prefix, "select-window", "-t", m.Option("session")+":"+arg[0])
						arg = []string{m.Option("session")}
						break
					}
					m.Cmd(prefix, "select-window", "-t", m.Option("session")+":"+m.Option("window"))

					// 切换终端
					m.Cmd(prefix, "select-pane", "-t", m.Option("session")+":"+m.Option("window")+"."+arg[0])
					arg = []string{m.Option("session"), m.Option("window")}

				case "modify":
					switch arg[2] {
					case "session":
						// 重命名会话
						m.Cmdy(prefix, "rename-session", "-t", arg[0], arg[3])
						arg = arg[:0]

					case "window":
						// 重命名窗口
						m.Cmdy(prefix, "rename-window", "-t", m.Option("session")+":"+arg[0], arg[3])
						arg = []string{m.Option("session")}

					default:
						return
					}
				case "delete":
					// 删除会话
					if !m.Has("session") {
						m.Cmdy(prefix, "kill-session", "-t", arg[0])
						arg = arg[:0]
						break
					}

					// 删除窗口
					if !m.Has("window") {
						m.Cmdy(prefix, "kill-window", "-t", m.Option("session")+":"+arg[0])
						arg = []string{m.Option("session")}
						break
					}

					// 删除终端
					m.Cmd(prefix, "kill-pane", "-t", m.Option("session")+":"+m.Option("window")+"."+arg[3])
					arg = []string{m.Option("session"), m.Option("window")}
				}
			}

			// 查看会话
			if m.Cmdy(prefix, "list-session", "-F", "#{session_id},#{session_attached},#{session_name},#{session_windows},#{session_height},#{session_width}",
				"cmd_parse", "cut", ",", "6", "id tag session windows height width"); len(arg) == 0 {
				return
			}

			// 创建会话
			if arg[0] != "" && !kit.Contains(m.Meta["session"], arg[0]) {
				m.Cmdy(prefix, "new-session", "-ds", arg[0])
			}
			m.Set("append").Set("result")

			// 查看窗口
			if m.Cmdy(prefix, "list-windows", "-t", arg[0], "-F", "#{window_id},#{window_active},#{window_name},#{window_panes},#{window_height},#{window_width}",
				"cmd_parse", "cut", ",", "6", "id tag window panes height width"); len(arg) == 1 {
				return
			}

			// 创建窗口
			if arg[1] != "" && !kit.Contains(m.Meta["window"], arg[1]) {
				m.Cmdy(prefix, "new-window", "-dt", arg[0], "-n", arg[1])
			}
			m.Set("append").Set("result")

			// 查看面板
			if len(arg) == 2 {
				m.Cmdy(prefix, "list-panes", "-t", arg[0]+":"+arg[1], "-F", "#{pane_id},#{pane_active},#{pane_index},#{pane_tty},#{pane_height},#{pane_width}",
					"cmd_parse", "cut", ",", "6", "id tag pane tty height width")
				return
			}

			// 执行命令
			target := arg[0] + ":" + arg[1] + "." + arg[2]
			if len(arg) > 3 {
				m.Cmdy(prefix, "send-keys", "-t", target, strings.Join(arg[3:], " "), "Enter")
				time.Sleep(1 * time.Second)
			}

			// 查看终端
			m.Echo(strings.TrimSpace(m.Cmdx(prefix, "capture-pane", "-pt", target)))
			return
		}},
		"docker": {Name: "docker image|volume|network|container", Help: "容器", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			prefix := kit.Trans(m.Confv("prefix", "docker"))
			switch arg[0] {
			case "image":
				if prefix = append(prefix, "image"); len(arg) < 3 {
					m.Cmdy(prefix, "ls", "cmd_parse", "cut", "cmd_headers", "IMAGE ID", "IMAGE_ID")
					break
				}

				switch arg[2] {
				case "运行":
					m.Cmdy(prefix[:2], "run", "-dt", m.Option("REPOSITORY")+":"+m.Option("TAG"))
				case "清理":
					m.Cmdy(prefix, "prune", "-f")
				case "delete":
					m.Cmdy(prefix, "rm", m.Option("IMAGE_ID"))
				case "创建":
					m.Option("base", m.Option("REPOSITORY")+":"+m.Option("TAG"))
					app := m.Conf("runtime", "boot.ctx_app")
					m.Option("name", app+":"+m.Time("20060102"))
					m.Option("file", m.Conf("docker", "output"))
					m.Option("user", m.Conf("runtime", "boot.username"))
					m.Option("host", "http://"+m.Conf("runtime", "boot.hostname")+".local"+m.Conf("runtime", "boot.web_port"))

					if f, _, e := kit.Create(m.Option("file")); m.Assert(e) {
						defer f.Close()
						if m.Assert(ctx.ExecuteStr(m, f, m.Conf("docker", "template."+app))) {
							m.Cmdy(prefix, "build", "-f", m.Option("file"), "-t", m.Option("name"), ".")
						}
					}

				default:
					if len(arg) == 3 {
						m.Cmdy(prefix, "pull", arg[1]+":"+arg[2])
						break
					}
				}

			case "volume":
				if prefix = append(prefix, "volume"); len(arg) == 1 {
					m.Cmdy(prefix, "ls", "cmd_parse", "cut", "cmd_headers", "VOLUME NAME", "VOLUME_NAME")
					break
				}

			case "network":
				if prefix = append(prefix, "network"); len(arg) == 1 {
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

			case "container":
				if prefix = append(prefix, "container"); len(arg) > 1 {
					switch arg[2] {
					case "进入":
						m.Cmdy(m.Confv("prefix", "tmux"), "new-window", "-t", "", "-n", m.Option("CONTAINER_NAME"),
							"-PF", "#{session_name}:#{window_name}.1", "docker exec -it "+arg[1]+" sh")
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
						m.Cmdy(prefix, "rm", arg[1])

					default:
						if len(arg) == 2 {
							m.Cmdy(prefix, "inspect", arg[1])
							return
						}
						m.Cmdy(prefix, "exec", arg[1], arg[2:])
						return
					}
				}
				m.Cmdy(prefix, "ls", "-a", "cmd_parse", "cut", "cmd_headers", "CONTAINER ID", "CONTAINER_ID")

			case "command":
				switch arg[3] {
				case "base":
					m.Echo("\n0[%s]$ %s %s\n", time.Now().Format("15:04:05"), arg[2], m.Conf("package", arg[2]+".update"))
					m.Cmdy(prefix, "exec", arg[1], arg[2], strings.Split(m.Conf("package", arg[2]+".update"), " "))
					m.Confm("package", []string{arg[2], arg[3]}, func(index int, value string) {
						m.Echo("\n%d[%s]$ %s %s %s\n", index+1, time.Now().Format("15:04:05"), arg[2], m.Conf("package", arg[2]+".install"), value)
						m.Cmdy(prefix, "exec", arg[1], arg[2], strings.Split(m.Conf("package", arg[2]+".install"), " "), value)
					})
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
				if _, e := os.Stat(path.Join(prefix[len(prefix)-1], ".git")); e != nil {
					m.Cmdy(prefix, "init")
				}
				if len(arg) > 1 {
					m.Cmdy(prefix, "remote", "add", "-f", kit.Select("origin", arg, 2), arg[1])
					m.Cmdy(prefix, "pull", kit.Select("origin", arg, 2), kit.Select("master", arg, 3))
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
			case "ctag":
				if f, p, e := kit.Create("etc/conf/tags"); m.Assert(e) {
					defer f.Close()
					for k, _ := range c.Commands {
						fmt.Fprintf(f, "%s\t%s\t/\"%s\": {Name/\n", k, "../../src/examples/code/code.go", k)
					}
					m.Echo(p)
				}
				return
			}
			return
		}},
		"/vim": {Name: "/vim sid pwd cmd arg sub", Help: "编辑器", Hand: func(m *ctx.Message, c *ctx.Context, cmd string, arg ...string) (e error) {
			cmd = strings.TrimPrefix(cmd, "/")
			if f, _, e := m.Optionv("request").(*http.Request).FormFile("sub"); e == nil {
				defer f.Close()
				if b, e := ioutil.ReadAll(f); e == nil {
					m.Option("sub", string(b))
				}
			}
			m.Log("info", "%v %v %v %v", cmd, m.Option("cmd"), m.Option("arg"), m.Option("sub"))
			m.Confm("login", []string{"hash", m.Option("sid"), "ship"}, func(key string, value string) { m.Option(key, value) })

			switch m.Option("cmd") {
			case "help":
				m.Echo(strings.Join(kit.Trans(m.Confv("help", "index")), "\n"))
			case "login":
				m.Cmd("login", "init", cmd)
			case "logout":
				m.Cmd("login", "exit")
			case "favor":
				m.Log("info", "river: %v dream: %v favor: %v", m.Option("river"), m.Option("dream"), m.Option("favor"))
				prefix := []string{"ssh._route", m.Option("dream"), "web.code.favor"}
				if m.Options("arg") {
					m.Cmd(prefix, "post", m.Option("favor"), m.Option("tab"), m.Option("note"), m.Option("arg"),
						m.Option("buf"), m.Option("line"), m.Option("col"))
					m.Set("append")
					break
				}

				m.Cmd(prefix, "list", m.Option("favor"), "", kit.Select("10", m.Option("limit")), "0", "tab", m.Option("tab")).Table(func(index int, value map[string]string) {
					m.Echo("%v\n", value["tab"]).Echo("%v:%v:%v:(%v): %v\n", value["file"], value["line"], value["col"], value["note"], value["word"])
				})

			case "read", "write":
				m.Cmd("trend", "post", "vim.opens", "sid", m.Option("sid"),
					"action", m.Option("cmd"), "file", m.Option("arg"), "pwd", m.Option("pwd"))
			case "exec":
				m.Cmd("trend", "post", "vim.cmds", "sid", m.Option("sid"),
					"cmd", m.Option("sub"), "file", m.Option("buf"), "pwd", m.Option("pwd"))
			case "insert":
				m.Cmd("trend", "post", "vim.txts", "sid", m.Option("sid"),
					"word", m.Option("sub"), "line", m.Option("row"), "col", m.Option("col"), "file", m.Option("buf"), "pwd", m.Option("pwd"))
			case "tasklet":
				m.Cmd("ssh._route", m.Option("dream"), "web.team.task", "create", "task",
					"3", "add", "action", m.Time(), m.Time("10m"), m.Option("arg"), m.Option("sub"))

			case "sync":
				m.Confv(cmd, []string{m.Option("arg"), "hash", m.Option("sid")}, []interface{}{})
				switch m.Option("arg") {
				case "bufs":
					m.Split(m.Option("sub"), " ", "5", "id tag name some line").Table(func(index int, value map[string]string) {
						m.Cmd("state", "post", cmd, m.Option("arg"),
							"buf", value["id"], "tag", value["tag"], "line", value["line"],
							"file", strings.TrimSuffix(strings.TrimPrefix(value["name"], "\""), "\""))
					})
				case "regs":
					m.Split(strings.TrimPrefix(m.Option("sub"), "--- Registers ---\n"), " ", "2", "name word").Table(func(index int, value map[string]string) {
						m.Cmd("state", "post", cmd, m.Option("arg"),
							"reg", strings.TrimPrefix(value["name"], "\""),
							"word", strings.Replace(strings.Replace(value["word"], "^I", "\t", -1), "^J", "\n", -1))
					})
				case "marks":
					m.Split(m.Option("sub"), " ", "4").Table(func(index int, value map[string]string) {
						m.Cmd("state", "post", cmd, m.Option("arg"),
							"mark", value["mark"], "line", value["line"], "col", value["col"],
							"file", value["file/text"])
					})
				case "tags":
					m.Split(strings.TrimSuffix(m.Option("sub"), ">"), " ", "6").Table(func(index int, value map[string]string) {
						m.Cmd("state", "post", cmd, m.Option("arg"),
							"tag", value["tag"], "line", value["line"], "file", value["in file/text"])
					})
				case "fixs":
					if strings.HasPrefix(m.Option("sub"), "\nError") {
						break
					}
					m.Split(m.Option("sub"), " ", "3", "id file word").Table(func(index int, value map[string]string) {
						vs := strings.Split(kit.Format(value["file"]), ":")
						m.Cmd("state", "post", cmd, m.Option("arg"),
							"fix", value["id"], "file", vs[0], "line", vs[1], "word", value["word"])
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
