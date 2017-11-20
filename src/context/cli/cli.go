package cli // {{{
// }}}
import ( // {{{
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
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

	target *ctx.Context
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
func (cli *CLI) parse(m *ctx.Message) bool { // {{{
	if len(cli.ins) == 1 && cli.Owner == nil {
		if aaa := cli.Root.Find("aaa"); aaa != nil {

			username := ""
			fmt.Fprintf(cli.out, "username>")
			fmt.Fscanln(cli.in, &username)

			password := ""
			fmt.Fprintf(cli.out, "password>")
			fmt.Fscanln(cli.in, &password)

			msg := m.Spawn(aaa, "username")

			if msg.Cmd("login", username, password) == "" {
				fmt.Fprintln(cli.out, "登录失败")
				m.Cmd("exit")
				cli.out.Close()
				cli.in.Close()
				return false
			}

			m.Cap("username", msg.Cap("username"))
		}
	}

	if len(cli.ins) == 1 || m.Conf("slient") != "yes" {
		cli.echo(m.Conf("PS1"))
	}

	line := ""
	if cli.next == "" {
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
			// cli.echo("\n%s\n", cli.Conf("结束语"))
			return false
		}
		m.Assert(e)
		line = ls

		if len(cli.ins) > 1 && m.Conf("slient") != "yes" {
			cli.echo(line)
		}

		if len(line) == 1 {
			if len(cli.ins) == 1 {
				line = cli.history[len(cli.history)-1]["cli"]
			} else {
				return true
			}
		}
	} else {
		line = cli.next
		cli.next = ""

		if m.Conf("slient") != "yes" {
			cli.echo(line)
			cli.echo("\n")
		}
	}

back:
	line = strings.TrimSpace(line)
	if line[0] == '#' {
		return true
	}
	ls := strings.Split(line, " ")

	msg := m.Spawn(cli.target)
	msg.Wait = make(chan bool)

	r := rune(ls[0][0])
	if !unicode.IsNumber(r) || !unicode.IsLetter(r) || r == '$' || r == '_' {
		if _, ok := cli.alias[string(r)]; ok {
			msg.Add("detail", ls[0][:1])
			if len(ls[0]) > 1 {
				ls[0] = ls[0][1:]
			} else {
				if len(ls) > 1 {
					ls = ls[1:]
				} else {
					ls = nil
				}
			}
		}
	}

	for i := 0; i < len(ls); i++ {
		ls[i] = strings.TrimSpace(ls[i])

		if ls[i][0] == '#' {
			break
		}
		if ls[i] != "" {
			msg.Add("detail", ls[i])
		}
	}

	ls = msg.Meta["detail"]
	if n, e := strconv.Atoi(ls[0]); e == nil && 0 <= n && n < len(cli.history) && ls[0] != cli.history[n]["cli"] {
		line = cli.history[n]["cli"]
		msg.Meta["detail"] = nil
		goto back
	}

	msg.Post(cli.Context)

	for _, v := range msg.Meta["result"] {
		cli.echo(v)
	}

	return true
}

// }}}
func (cli *CLI) echo(str string, arg ...interface{}) { // {{{
	// if len(cli.ins) == 1 || m.Conf("slient") != "yes" {
	fmt.Fprintf(cli.out, str, arg...)
	// }
}

// }}}

func (cli *CLI) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	cli.Caches["username"] = &ctx.Cache{Name: "登录用户", Value: "", Help: "登录用户名"}
	cli.Caches["nhistory"] = &ctx.Cache{Name: "历史命令数量", Value: "0", Help: "当前终端已经执行命令的数量", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
		x.Value = fmt.Sprintf("%d", len(cli.history))
		return x.Value
	}}

	cli.Configs["slient"] = &ctx.Config{Name: "屏蔽脚本输出(yes/no)", Value: "yes", Help: "屏蔽脚本输出的信息，yes:屏蔽，no:不屏蔽"}
	cli.Configs["default"] = &ctx.Config{Name: "默认的搜索起点(root/back/home)", Value: "root", Help: "模块搜索的默认起点，root:从根模块，back:从父模块，home:从当前模块"}
	cli.Configs["PS1"] = &ctx.Config{Name: "命令行提示符(target/detail)", Value: "target", Help: "命令行提示符，target:显示当前模块，detail:显示详细信息", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
		cli, ok := m.Target.Server.(*CLI) // {{{
		if ok && cli.target != nil {
			// c = cli.target
			switch x.Value {
			case "target":
				return fmt.Sprintf("%s[%s]\033[32m%s\033[0m> ", m.Cap("nhistory"), time.Now().Format("15:04:05"), cli.target.Name)
			case "detail":
				return fmt.Sprintf("%s[%s](%s,%s,%s)\033[32m%s\033[0m> ", m.Cap("nhistory"), time.Now().Format("15:04:05"), m.Cap("ncontext"), m.Cap("nmessage"), m.Cap("nserver"), m.Target.Name)
			}

		}

		return fmt.Sprintf("[%s]\033[32m%s\033[0m ", time.Now().Format("15:04:05"), x.Value)
		// }}}
	}}

	if len(arg) > 0 {
		cli.Configs["init.sh"] = &ctx.Config{Name: "启动脚本", Value: arg[0], Help: "模块启动时自动运行的脚本"}
	}

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
			log.Println(cli.Name, "master terminal:")
			if cli.bufs == nil {
				cli.bufs = make([][]byte, 0, 10)
			}
			for {
				b := make([]byte, 128)
				n, e := cli.bio.Read(b)
				log.Println(cli.Name, "read:", n)
				m.Assert(e)
				cli.bufs = append(cli.bufs, b)
			}
			return true
		} else {
			log.Println(cli.Name, "slaver terminal:")

			if f, e := os.Open(m.Conf("init.sh")); e == nil {
				cli.push(f)
			}

			go m.AssertOne(m, true, func(m *ctx.Message) {
				for cli.parse(m) {
				}
			})
		}
	}

	for m.Deal(func(msg *ctx.Message, arg ...string) bool {
		if a, ok := cli.alias[arg[0]]; ok {
			arg[0] = a
		}
		return true

	}, func(msg *ctx.Message, arg ...string) bool {
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
func (cli *CLI) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	c.Owner = nil
	s := new(CLI)
	s.Context = c
	return s
}

// }}}
func (cli *CLI) Exit(m *ctx.Message, arg ...string) bool { // {{{
	switch cli.Context {
	case m.Source:
		return false

	case m.Target:
		log.Println(cli.Name, "release:")
	}

	return true
}

// }}}

var Index = &ctx.Context{Name: "cli", Help: "管理终端",
	Caches: map[string]*ctx.Cache{
		"nterm": &ctx.Cache{Name: "终端数量", Value: "0", Help: "已经运行的终端数量"},
	},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"context": &ctx.Command{Name: "context [root|back|home] [[find|search] name] [show|spawn|start|switch][args]", Help: "查找并操作模块，\n查找起点root:根模块、back:父模块、home:本模块，\n查找方法find:路径匹配、search:模糊匹配，\n查找对象name:支持点分和正则，\n操作类型show:显示信息、switch:切换为当前、start:启动模块、spawn:分裂子模块，args:启动参数",
			Formats: map[string]int{"root": 0, "back": 0, "home": 0, "find": 1, "search": 1, "show": 0, "switch": 0, "start": -1, "spawn": -1},
			Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
				cli, ok := m.Source.Server.(*CLI) // {{{
				if !ok {
					cli, ok = c.Server.(*CLI)
					if !ok {
						return ""
					}
				}

				switch len(arg) {
				case 0:
					m.Travel(m.Target.Root, func(m *ctx.Message) bool {
						if m.Target.Context != nil {
							target := m.Target
							m.Target = m.Target.Owner
							if m.Target != nil && m.Check(m.Target, "caches", "username") && m.Check(m.Target, "caches", "group") {
								m.Echo("%s: %s(%s) %s %s\n", target.Context.Name, target.Name, target.Help, m.Cap("username"), m.Cap("group"))
							} else {
								m.Echo("%s: %s(%s)\n", target.Context.Name, target.Name, target.Help)
							}
							m.Target = target
						}
						return true
					})
					return ""
				}

				target := m.Target.Root
				if m.Has("home") {
					target = m.Target
				}
				if m.Has("root") {
					target = m.Target.Root
				}
				if m.Has("back") && target.Context != nil {
					target = m.Target.Context
				}

				cs := []*ctx.Context{}
				switch {
				case m.Has("search"):
					if s := m.Search(target, m.Get("search")); len(s) > 0 {
						cs = append(cs, s...)
					}
				case m.Has("find"):
					if s := target.Find(m.Get("find")); s != nil {
						cs = append(cs, s)
					}
				case m.Has("args"):
					if s := m.Search(target, m.Get("args")); len(s) > 0 {
						cs = append(cs, s...)
					}
				default:
					cs = append(cs, target)
				}

				for _, v := range cs {
					// if !m.Source.Check(v) {
					// 	continue
					// }
					//
					switch {
					case m.Has("start"):
						args := m.Meta["start"]
						m.Message.Spawn(v, args[0]).Start(arg[0], args[1:]...)
					case m.Has("spawn"):
						args := m.Meta["spawn"]
						msg := m.Spawn(v)
						v.Spawn(msg, args[0]).Begin(msg)
						cli.target = msg.Target
					case m.Has("switch"):
						cli.target = v
					case m.Has("show"):
						m.Echo("%s: %s\n", v.Name, v.Help)
						m.Echo("模块资源：\n")
						for i, v := range v.Requests {
							m.Echo("\t%d(%d): <- %s %s\n", i, v.Code, v.Source.Name, v.Source.Help)
						}
						m.Echo("模块引用：\n")
						for k, v := range v.Sessions {
							m.Echo("\t%s(%d): -> %s %s\n", k, v.Code, v.Target.Name, v.Target.Help)
						}
					default:
						cli.target = v
					}
				}

				return ""
				// }}}
			}},
		"source": &ctx.Command{Name: "source file", Help: "运行脚本", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
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
		"alias": &ctx.Command{Name: "alias [short [long]]", Help: "查看日志", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
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
		"history": &ctx.Command{Name: "history number", Help: "查看日志", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
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
			Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
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

				s := c.Root.Find(m.Get("args"))
				action := "dial"
				if m.Has("listen") {
					action = "listen"
				}

				msg := m.Spawn(s)
				if m.Has("master") {
					msg.Template = msg.Spawn(msg.Source).Add("option", "master")
				}
				msg.Cmd(action, m.Get(action))

				return ""
			}},
		// }}}
		"open": &ctx.Command{Name: "open [master|slaver] [script [log]]", Help: "建立远程连接",
			Options: map[string]string{"master": "主控终端", "slaver": "被控终端", "args": "启动参数", "io": "读写流"},
			Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
				go m.Start(fmt.Sprintf("PTS%d", m.Capi("nterm")), m.Meta["args"]...) // {{{
				return ""
				// }}}
			}},
		"void": &ctx.Command{Name: "", Help: "", Hand: nil},
	},
	Messages: make(chan *ctx.Message, 10),
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void",
			Commands: map[string]*ctx.Command{
				"open": &ctx.Command{},
			},
		},
	},
}

func init() {
	cli := &CLI{}
	cli.Context = Index
	ctx.Index.Register(Index, cli)
}
