package cli // {{{
// }}}
import ( // {{{
	"context"

	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// }}}

type CLI struct {
	out  io.WriteCloser
	in   io.ReadCloser
	ins  []io.ReadCloser
	bio  *bufio.Reader
	bios []*bufio.Reader
	bufs [][]byte

	history []map[string]string
	alias   map[string]string
	next    string
	exit    bool
	login   *ctx.Context
	lex     *ctx.Message

	temp   *ctx.Context
	target *ctx.Context
	m      *ctx.Message
	*ctx.Context
}

func (cli *CLI) push(f io.ReadCloser) { // {{{
	if cli.ins == nil || cli.bios == nil {
		cli.ins = make([]io.ReadCloser, 0, 3)
		cli.bios = make([]*bufio.Reader, 0, 3)
	}

	cli.in = f
	cli.ins = append(cli.ins, cli.in)
	cli.bio = bufio.NewReader(f)
	cli.bios = append(cli.bios, cli.bio)
}

// }}}
func (cli *CLI) echo(str string, arg ...interface{}) { // {{{
	if len(cli.ins) == 1 || cli.m.Conf("slient") != "yes" {
		fmt.Fprintf(cli.out, str, arg...)
	}
}

// }}}
func (cli *CLI) parse(m *ctx.Message) bool { // {{{
	line := ""
	if cli.next == "" {
		cli.echo(m.Conf("PS1"))
		ls, e := cli.bio.ReadString('\n')
		if e == io.EOF {
			l := len(cli.ins)
			if l > 1 {
				cli.ins = cli.ins[:l-1]
				cli.bios = cli.bios[:l-1]
				cli.in = cli.ins[l-2]
				cli.bio = cli.bios[l-2]
				return true
			}
			return false
		}
		m.Assert(e)
		line = ls

		if len(cli.ins) > 1 {
			cli.echo(line)
			cli.echo("\n")
		}

		if len(line) == 1 {
			if len(cli.ins) > 1 {
				return true
			}
			line = cli.history[len(cli.history)-1]["cli"]
		}
	} else {
		line, cli.next = cli.next, ""

		if m.Conf("slient") != "yes" {
			cli.echo(m.Conf("PS1"))
			cli.echo(line)
			cli.echo("\n")
		}
	}

	line = strings.TrimSpace(line)
	if len(line) == 0 || line[0] == '#' {
		return true
	}

	ls := []string{}
	if cli.lex == nil {
		ls = strings.Split(line, " ")
	} else {
		lex := m.Spawn(cli.lex.Target)
		m.Assert(lex.Cmd("split", line, "void"))

		ls = lex.Meta["result"]
	}

	if len(ls) == 0 {
		return true
	}

	msg := m.Spawn(cli.target)
	if cli.temp != nil {
		msg.Target, cli.temp = cli.temp, nil
	}

	for i := 0; i < len(ls); i++ {
		ls[i] = strings.TrimSpace(ls[i])
		if ls[i] == "" {
			continue
		}
		if ls[i][0] == '#' {
			break
		}

		if cli.lex != nil && len(ls[i]) > 1 {
			switch ls[i][0] {
			case '"', '\'':
				ls[i] = ls[i][1 : len(ls[i])-1]
			}
		}

		if r := rune(ls[i][0]); r == '$' || r == '_' || (!unicode.IsNumber(r) && !unicode.IsLetter(r)) {
			if c, ok := cli.alias[string(r)]; ok {
				if msg.Add("detail", c); len(ls[i]) > 1 {
					ls[i] = ls[i][1:]
				} else {
					continue
				}
			}
		}

		msg.Add("detail", ls[i])
	}

	msg.Wait = make(chan bool)
	msg.Post(cli.Context)

	m.Capi("nhistory", 1)
	cli.echo(strings.Join(msg.Meta["result"], ""))

	return true
}

// }}}

func (cli *CLI) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	cli.Owner = nil

	s := new(CLI)
	s.Context = c
	return s
}

// }}}
func (cli *CLI) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	cli.Caches["username"] = &ctx.Cache{Name: "登录用户", Value: "", Help: "登录用户名"}
	cli.Caches["nhistory"] = &ctx.Cache{Name: "历史命令数量", Value: "0", Help: "当前终端已经执行命令的数量"}

	cli.Configs["slient"] = &ctx.Config{Name: "屏蔽脚本输出(yes/no)", Value: "yes", Help: "屏蔽脚本输出的信息，yes:屏蔽，no:不屏蔽"}
	cli.Configs["default"] = &ctx.Config{Name: "默认的搜索起点(root/back/home)", Value: "root", Help: "模块搜索的默认起点，root:从根模块，back:从父模块，home:从当前模块"}
	cli.Configs["PS1"] = &ctx.Config{Name: "命令行提示符(target/detail)", Value: "target", Help: "命令行提示符，target:显示当前模块，detail:显示详细信息", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
		if len(arg) > 0 { // {{{
			return arg[0]
		}

		ps := make([]string, 0, 3)

		if cli, ok := m.Target.Server.(*CLI); ok && cli.target != nil {
			ps = append(ps, m.Cap("nhistory"))
			ps = append(ps, "[")
			ps = append(ps, time.Now().Format("15:04:05"))
			ps = append(ps, "]")

			switch x.Value {
			case "detail":
				ps = append(ps, "(")
				ps = append(ps, m.Cap("ncontext"))
				ps = append(ps, ",")
				ps = append(ps, m.Cap("nmessage"))
				ps = append(ps, ",")
				ps = append(ps, m.Cap("nserver"))
				ps = append(ps, ")")
			case "target":
			}

			ps = append(ps, "\033[32m")
			ps = append(ps, cli.target.Name)
			ps = append(ps, "\033[0m> ")

		} else {
			ps = append(ps, "[")
			ps = append(ps, time.Now().Format("15:04:05"))
			ps = append(ps, "]")

			ps = append(ps, "\033[32m")
			ps = append(ps, x.Value)
			ps = append(ps, "\033[0m> ")
		}

		return strings.Join(ps, "")
		// }}}
	}}
	cli.Configs["lex"] = &ctx.Config{Name: "词法解析器", Value: "", Help: "命令行词法解析器", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
		if len(arg) > 0 { // {{{
			cli, ok := m.Target.Server.(*CLI)
			m.Assert(ok, "模块类型错误")

			cli.lex = m.Find(arg[0], m.Target.Root)
			m.Assert(cli.lex != nil, "词法解析模块不存在")

			cli.lex.Cmd("train", "[ \n\t]+", "void", "void")
			cli.lex.Cmd("train", "#[^\n]*\n", "void", "void")
		}
		return x.Value
		// }}}
	}}

	if len(arg) > 0 {
		cli.Configs["init.sh"] = &ctx.Config{Name: "启动脚本", Value: arg[0], Help: "模块启动时自动运行的脚本"}
	}

	cli.m = m
	cli.Context.Master = cli.Context

	cli.target = cli.Context
	cli.history = make([]map[string]string, 0, 100)
	cli.alias = map[string]string{
		"~": "context",
		"!": "history",
		"@": "config",
		"$": "cache",
		"&": "server",
		"*": "message",
		":": "command",
	}

	return cli
}

// }}}
func (cli *CLI) Start(m *ctx.Message, arg ...string) bool { // {{{
	m.Capi("nterm", 1)
	defer m.Capi("nterm", -1)

	if stream, ok := m.Data["io"]; ok {
		io := stream.(io.ReadWriteCloser)
		cli.out = io
		cli.push(io)

		if m.Has("master") {
			m.Log("info", "%s: master terminal", cli.Name)
			if cli.bufs == nil {
				cli.bufs = make([][]byte, 0, 10)
			}
			for {
				b := make([]byte, 128)
				n, e := cli.bio.Read(b)
				m.Log("info", "%s: read %d", cli.Name, n)
				m.Assert(e)
				cli.bufs = append(cli.bufs, b)
			}
			return true
		} else {
			if cli.Owner == nil {
				if msg := m.Find("aaa", m.Target.Root); msg != nil {
					username := ""
					cli.echo("username>")
					fmt.Fscanln(cli.in, &username)

					password := ""
					cli.echo("password>")
					fmt.Fscanln(cli.in, &password)

					if msg.Cmd("login", username, password) == "" {
						cli.echo("登录失败")
						m.Cmd("exit")
						cli.out.Close()
						cli.in.Close()
						return false
					}

					m.Cap("username", msg.Cap("username"))
				}
			}

			m.Log("info", "%s: slaver terminal", cli.Name)
			m.Log("info", "%s: open %s", cli.Name, m.Conf("init.sh"))
			if f, e := os.Open(m.Conf("init.sh")); e == nil {
				cli.push(f)
			}

			go m.AssertOne(m, true, func(m *ctx.Message) {
				for cli.parse(m) {
				}
			})
		}
	}

	for m.Deal(nil, func(msg *ctx.Message, arg ...string) bool {
		cli.history = append(cli.history, map[string]string{
			"time":  time.Now().Format("15:04:05"),
			"index": fmt.Sprintf("%d", len(cli.history)),
			"cli":   strings.Join(msg.Meta["detail"], " "),
		})

		if len(arg) > 0 {
			// cli.next = arg[0]
			// arg[0] = ""
		}

		if cli.exit == true {
			return false
		}
		return true
	}) {
	}

	return true
}

// }}}
func (cli *CLI) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch cli.Context {
	case m.Source:
		return false

	case m.Target:
		m.Log("exit", "%s: release", cli.Name)
	}

	return false
}

// }}}

var Index = &ctx.Context{Name: "cli", Help: "管理终端",
	Caches: map[string]*ctx.Cache{
		"nterm": &ctx.Cache{Name: "终端数量", Value: "0", Help: "已经运行的终端数量"},
	},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"context": &ctx.Command{Name: "context [root|back|home] [[find|search] name] [list|show|spawn|start|switch|close][args]", Help: "查找并操作模块，\n查找起点root:根模块、back:父模块、home:本模块，\n查找方法find:路径匹配、search:模糊匹配，\n查找对象name:支持点分和正则，\n操作类型show:显示信息、switch:切换为当前、start:启动模块、spawn:分裂子模块，args:启动参数",
			Formats: map[string]int{"root": 0, "back": 0, "home": 0, "find": 1, "search": 1, "list": 0, "show": 0, "close": 0, "switch": 0, "start": 0, "spawn": 0},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
				cli, ok := m.Source.Server.(*CLI) // {{{
				if !ok {
					cli, ok = c.Server.(*CLI)
					if !ok {
						return ""
					}
				}

				target := m.Target
				switch cli.m.Conf("default") {
				case "home":
					target = m.Target
				case "root":
					target = m.Target.Root
				case "back":
					if target.Context != nil {
						target = m.Target.Context
					}
				}
				if m.Has("home") {
					target = m.Target
				}
				if m.Has("root") {
					target = m.Target.Root
				}
				if m.Has("back") && target.Context != nil {
					target = m.Target.Context
				}

				ms := []*ctx.Message{}
				switch {
				case m.Has("search"):
					if s := m.Search(m.Get("search"), target); len(s) > 0 {
						ms = append(ms, s...)
					}
				case m.Has("find"):
					if msg := m.Find(m.Get("find"), target); msg != nil {
						ms = append(ms, msg)
					}
				case m.Has("args"):
					if s := m.Search(m.Get("args"), target); len(s) > 0 {
						ms = append(ms, s...)
						arg = arg[1:]
						break
					}
					fallthrough
				default:
					ms = append(ms, m.Spawn(target))
				}

				for _, v := range ms {
					switch {
					case m.Has("spawn"):
						v.Target.Spawn(v, arg[0], arg[1]).Begin(v)
						cli.target = v.Target
					case m.Has("start"):
						v.Set("detail", arg...).Target.Start(v)
						cli.target = v.Target
					case m.Has("switch"):
						cli.target = v.Target
					case m.Has("close"):
						v.Target.Close(v)
					case m.Has("show"):
						m.Echo("%s(%s): %s\n", v.Target.Name, v.Target.Owner.Name, v.Target.Help)
						if len(v.Target.Requests) > 0 {
							m.Echo("模块资源：\n")
							for i, v := range v.Target.Requests {
								m.Echo("  %d(%d): <- %s %s\n", i, v.Code, v.Source.Name, v.Meta["detail"])
								for i, v := range v.Messages {
									m.Echo("    %d(%d): -> %s %s\n", i, v.Code, v.Source.Name, v.Meta["detail"])
								}
							}
						}
						if len(v.Target.Sessions) > 0 {
							m.Echo("模块引用：\n")
							for k, v := range v.Target.Sessions {
								m.Echo("  %s(%d): -> %s %v\n", k, v.Code, v.Target.Name, v.Meta["detail"])
							}
						}
					case m.Has("list") || len(m.Meta["detail"]) == 1 || (len(arg) == 0 && cli.target == v.Target):
						if len(m.Meta["detail"]) == 1 {
							v.Target = cli.target
						}
						m.Travel(v.Target, func(msg *ctx.Message) bool {
							if msg.Target.Context != nil {
								target := msg.Target
								m.Echo("%s: %s(%s)", target.Context.Name, target.Name, target.Help)

								msg.Target = msg.Target.Owner
								if msg.Target != nil && msg.Check(msg.Target, "caches", "username") && msg.Check(msg.Target, "caches", "group") {
									m.Echo(" %s %s", msg.Cap("username"), msg.Cap("group"))
								}

								m.Echo("\n")
								msg.Target = target
							}
							return true
						})
					case len(arg) > 0:
						cli.next = strings.Join(arg, " ")
						cli.temp = v.Target
					default:
						cli.target = v.Target
						return ""
					}
				}
				return ""
				// }}}
			}},
		"source": &ctx.Command{Name: "source file", Help: "运行脚本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			cli := c.Server.(*CLI) // {{{
			switch len(arg) {
			case 1:
				f, e := os.Open(arg[0])
				m.Assert(e)
				cli.push(f)
			}

			return ""
			// }}}
		}},
		"return": &ctx.Command{Name: "return", Help: "运行脚本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			cli := c.Server.(*CLI) // {{{
			cli.bio.Discard(cli.bio.Buffered())
			return ""
			// }}}
		}},
		"alias": &ctx.Command{Name: "alias [short [long]]", Help: "查看日志", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			cli := c.Server.(*CLI) // {{{
			switch len(arg) {
			case 0:
				for k, v := range cli.alias {
					m.Echo("%s: %s\n", k, v)
				}
			case 2:
				switch arg[0] {
				case "delete":
					delete(cli.alias, arg[1])
				default:
					cli.alias[arg[0]] = arg[1]
					m.Echo("%s: %s\n", arg[0], cli.alias[arg[0]])
				}
			default:
				cli.alias[arg[0]] = strings.Join(arg[1:], " ")
				m.Echo("%s: %s\n", arg[0], cli.alias[arg[0]])
			}
			return ""
			// }}}
		}},
		"history": &ctx.Command{Name: "history number", Help: "查看日志", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
			cli := c.Server.(*CLI) // {{{
			switch len(arg) {
			case 0:
				for i, v := range cli.history {
					m.Echo("%d %s %s\n", i, v["time"], v["cli"])
				}
			case 1:
				n, e := strconv.Atoi(arg[0])
				if e == nil && 0 <= n && n < len(cli.history) {
					cli.next = cli.history[n]["cli"]
				}
			default:
				n, e := strconv.Atoi(arg[0])
				if e == nil && 0 <= n && n < len(cli.history) {
					cli.history[n]["cli"] = strings.Join(arg[1:], " ")
				}
			}
			return ""
			// }}}
		}},
		"remote": &ctx.Command{Name: "remote [send args...]|[[master|slaver] listen|dial address protocol]", Help: "建立远程连接",
			Formats: map[string]int{"send": -1, "master": 0, "slaver": 0, "listen": 1, "dial": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
				if m.Has("send") { // {{{
					cli := m.Target.Server.(*CLI)

					cli.out.Write([]byte(strings.Join(m.Meta["args"], " ") + "\n"))
					m.Echo("~~~remote~~~\n")
					time.Sleep(100 * time.Millisecond)
					for _, b := range cli.bufs {
						m.Echo("%s", string(b))
					}
					cli.bufs = cli.bufs[0:0]
					m.Echo("\n~~~remote~~~\n")

					return ""
				}

				action := "dial"
				if m.Has("listen") {
					action = "listen"
				}
				msg := m.Find(m.Get("args"), m.Target.Root)

				if m.Has("master") {
					msg.Template = msg.Spawn(msg.Source).Add("option", "master")
				}
				msg.Cmd(action, m.Get(action))

				return ""
			}},
		// }}}
		"open": &ctx.Command{Name: "open [master|slaver] [script [log]]", Help: "建立远程连接",
			Options: map[string]string{"master": "主控终端", "slaver": "被控终端", "args": "启动参数", "io": "读写流"},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) string {
				m.Start(fmt.Sprintf("PTS%d", m.Capi("nterm")), "管理终端", arg...) // {{{
				return ""
				// }}}
			}},
	},
	Messages: make(chan *ctx.Message, 10),
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void",
			Commands: map[string]*ctx.Command{
				"context": &ctx.Command{},
				"open":    &ctx.Command{},
			},
		},
	},
}

func init() {
	cli := &CLI{}
	cli.Context = Index
	ctx.Index.Register(Index, cli)
}
