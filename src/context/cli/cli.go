package cli // {{{
// }}}
import ( // {{{
	"bufio"
	"context"
	// _ "context/tcp"
	// _ "context/web"
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

	history []map[string]string
	alias   map[string]string
	next    string
	exit    bool

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
	if len(cli.ins) == 1 || cli.Conf("slient") != "yes" {
		cli.echo(cli.Conf("PS1"))
	}

	line := ""
	if cli.next == "" {
		ls, e := cli.bio.ReadString('\n')
		if e == io.EOF {
			l := len(cli.ins)
			if l == 1 {
				// cli.echo("\n%s\n", cli.Conf("结束语"))
				return false
				ls = "exit"
				e = nil
			} else {
				cli.ins = cli.ins[:l-1]
				cli.bios = cli.bios[:l-1]
				cli.in = cli.ins[l-2]
				cli.bio = cli.bios[l-2]
				return true
			}
		}
		cli.Assert(e)
		line = ls

		if len(cli.ins) > 1 || cli.Conf("slient") != "yes" {
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

		if cli.Conf("slient") != "yes" {
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
	if len(cli.ins) == 1 || cli.Conf("slient") != "yes" {
		fmt.Fprintf(cli.out, str, arg...)
	}
}

// }}}

func (cli *CLI) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	cli.history = make([]map[string]string, 0, 100)
	cli.alias = map[string]string{
		"~": "context",
		"!": "history",
		"@": "config",
		"$": "cache",
		"&": "server",
		"*": "message",
	}

	cli.target = cli.Context

	cli.Caches["nhistory"] = &ctx.Cache{Name: "历史命令数量", Value: "0", Help: "当前终端已经执行命令的数量", Hand: func(c *ctx.Context, x *ctx.Cache, arg ...string) string {
		x.Value = fmt.Sprintf("%d", len(cli.history))
		return x.Value
	}}

	return cli
}

// }}}
func (cli *CLI) Start(m *ctx.Message, arg ...string) bool { // {{{
	cli.Capi("nterm", 1)
	defer cli.Capi("nterm", -1)

	if cli.Messages == nil {
		cli.Messages = make(chan *ctx.Message, cli.Confi("MessageQueueSize"))
	}
	if len(arg) > 0 {
		cli.Configs["init.sh"] = &ctx.Config{Name: "启动脚本", Value: arg[0], Help: "模块启动时自动运行的脚本"}
	}

	if stream, ok := m.Data["io"]; ok {
		io := stream.(io.ReadWriteCloser)
		cli.out = io
		cli.push(io)

		if f, e := os.Open(cli.Conf("init.sh")); e == nil {
			cli.push(f)
		}

		// cli.echo("%s\n", cli.Conf("hello"))

		go cli.AssertOne(m, true, func(c *ctx.Context, m *ctx.Message) {
			for cli.parse(m) {
			}
		})
	}

	for cli.Deal(func(msg *ctx.Message, arg ...string) bool {
		if a, ok := cli.alias[arg[0]]; ok {
			arg[0] = a
		}
		return true

	}, func(msg *ctx.Message, arg ...string) bool {
		cli.history = append(cli.history, map[string]string{
			"time":  time.Now().Format("15:04:05"),
			"index": fmt.Sprintf("%d", len(cli.history)),
			"cli":   strings.Join(arg, " "),
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

	s := new(CLI)
	s.Context = c
	return s
}

// }}}
func (cli *CLI) Exit(m *ctx.Message, arg ...string) bool { // {{{
	if cli.Context != Index {
		delete(cli.Context.Context.Contexts, cli.Name)
	}
	return true
}

// }}}

var Index = &ctx.Context{Name: "cli", Help: "管理终端",
	Caches: map[string]*ctx.Cache{
		"nterm": &ctx.Cache{Name: "终端数量", Value: "0", Help: "已经运行的终端数量"},
	},
	Configs: map[string]*ctx.Config{
		"slient": &ctx.Config{Name: "屏蔽脚本输出(yes/no)", Value: "yes", Help: "屏蔽脚本输出的信息，yes:屏蔽，no:不屏蔽"},
		// "hello":  &ctx.Config{Name: "开场白", Value: "\n~~~  Hello Context & Message World  ~~~\n", Help: "模块启动时输出的信息"},
		// "byebye": &ctx.Config{Name: "结束语", Value: "\n~~~  Byebye Context & Message World  ~~~\n", Help: "模块停止时输出的信息"},

		"PS1": &ctx.Config{Name: "命令行提示符(target/detail)", Value: "target", Help: "命令行提示符，target:显示当前模块，detail:显示详细信息", Hand: func(c *ctx.Context, x *ctx.Config, arg ...string) string {
			cli, ok := c.Server.(*CLI) // {{{
			if ok && cli.target != nil {
				// c = cli.target
				switch x.Value {
				case "target":
					return fmt.Sprintf("%s[%s]\033[32m%s\033[0m> ", c.Cap("nhistory"), time.Now().Format("15:04:05"), cli.target.Name)
				case "detail":
					return fmt.Sprintf("%s[%s](%s,%s,%s)\033[32m%s\033[0m> ", c.Cap("nhistory"), time.Now().Format("15:04:05"), c.Cap("ncontext"), c.Cap("nmessage"), c.Cap("nserver"), c.Name)
				}

			}

			return fmt.Sprintf("[%s]\033[32m%s\033[0m ", time.Now().Format("15:04:05"), x.Value)
			// }}}
		}},
	},
	Commands: map[string]*ctx.Command{
		"context": &ctx.Command{Name: "context [root|back|home] [[find|search] name] [show|spawn|start|switch][args]", Help: "查找并操作模块，\n查找起点root:根模块、back:父模块、home:本模块，\n查找方法find:路径匹配、search:模糊匹配，\n查找对象name:支持点分和正则，\n操作类型show:显示信息、switch:切换为当前、start:启动模块、spawn:分裂子模块，args:启动参数", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			cli, ok := c.Server.(*CLI) // {{{
			if !ok {
				return ""
			}

			switch len(arg) {
			case 0:
				m.Target.Root.Travel(func(c *ctx.Context) bool {
					if c.Context != nil {
						m.Echo("%s: %s(%s)\n", c.Context.Name, c.Name, c.Help)
					}
					return true
				})
				return ""
			}

			target := m.Target.Root
			method := "search"
			action := "switch"
			which := ""
			args := []string{}

			for len(arg) > 0 {
				switch arg[0] {
				case "root":
					target = m.Target.Root
				case "back":
					if m.Target.Context != nil {
						target = m.Target.Context
					}
				case "home":
					target = m.Target
				case "find", "search":
					method = arg[0]
					which = arg[1]
					arg = arg[1:]
				case "switch", "spawn", "start", "show":
					action = arg[0]
					args = arg[1:]
					arg = arg[:1]
				default:
					which = arg[0]
				}

				arg = arg[1:]
			}

			cs := []*ctx.Context{}

			if which == "" {
				cs = append(cs, target)
			} else {
				switch method {
				case "search":
					if s := target.Search(which); len(s) > 0 {
						cs = append(cs, s...)
					}
				case "find":
					if s := target.Find(which); s != nil {
						cs = append(cs, s)
					}
				}
			}

			for _, v := range cs {
				switch action {
				case "switch":
					cli.target = v
				case "spawn":
					msg := m.Spawn(v)
					v.Spawn(msg, args[0], args[1:]...)
					v.Begin(msg)
				case "start":
					m.Spawn(v).Start(args...)
				case "show":
					m.Echo("%s: %s\n", v.Name, v.Help)
					m.Echo("引用模块：\n")
					for k, v := range v.Session {
						m.Echo("\t%s(%d): %s %s\n", k, v.Code, v.Target.Name, v.Target.Help)
					}
					m.Echo("索引模块：\n")
					for i, v := range v.Resource {
						m.Echo("\t%d(%d): %s %s\n", i, v.Code, v.Context.Name, v.Context.Help)
					}
				}
			}

			return ""
			// }}}
		}},
		"message": &ctx.Command{Name: "message", Help: "查看上下文", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			ms := []*ctx.Message{ctx.Pulse} // {{{
			for i := 0; i < len(ms); i++ {
				m.Echo("%d %s.%s -> %s.%d: %s %v\n", ms[i].Code, ms[i].Context.Name, ms[i].Name, ms[i].Target.Name, ms[i].Index, ms[i].Time.Format("15:04:05"), ms[i].Meta["detail"])
				ms = append(ms, ms[i].Messages...)
			}
			return ""
			// }}}
		}},
		"server": &ctx.Command{Name: "server start|stop|switch", Help: "服务启动停止切换", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			s := m.Target // {{{
			switch len(arg) {
			case 0:
				m.Target.Root.Travel(func(c *ctx.Context) bool {
					if x, ok := c.Caches["status"]; ok {
						m.Echo("%s(%s): %s\n", c.Name, x.Value, c.Help)
					}
					return true
				})

			case 1:
				switch arg[0] {
				case "start":
					if s != nil {
						go s.Start(m)
					}
				case "stop":
				case "switch":
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
				c.Assert(e)
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
		"command": &ctx.Command{Name: "command [all] [name args]", Help: "查看修改添加配置", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			all := false // {{{
			if len(arg) > 0 && arg[0] == "all" {
				arg = arg[1:]
				all = true
			}

			m.Target.BackTrace(func(s *ctx.Context) bool {
				switch len(arg) {
				case 0:
					for k, v := range s.Commands {
						m.Echo("%s: %s\n", k, v.Name)
					}
				case 1:
					if v, ok := s.Commands[arg[0]]; ok {
						m.Echo("%s\n%s\n", v.Name, v.Help)
					}
				default:
					m.Spawn(s).Cmd(arg...)
					return false
				}
				return all
			})
			return ""
			// }}}
		}},
		"config": &ctx.Command{Name: "config [all] [key value|[name value help]]", Help: "查看修改添加配置", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			all := false // {{{
			if len(arg) > 0 && arg[0] == "all" {
				arg = arg[1:]
				all = true
			}

			m.Target.BackTrace(func(s *ctx.Context) bool {
				switch len(arg) {
				case 0:
					for k, v := range s.Configs {
						m.Echo("%s(%s): %s\n", k, v.Value, v.Name)
					}
				case 1:
					if v, ok := s.Configs[arg[0]]; ok {
						m.Echo("%s: %s\n", v.Name, v.Help)
					}
				case 2:
					if s != m.Target {
						return false
					}

					switch arg[0] {
					case "void":
						s.Conf(arg[1], "")
					case "delete":
						if _, ok := s.Configs[arg[1]]; ok {
							delete(s.Configs, arg[1])
						}
					default:
						s.Conf(arg[0], arg[1])
					}
				case 4:
					s.Conf(arg[0], arg[1:]...)
					return false
				}
				return all
			})
			return ""
			// }}}
		}},
		"cache": &ctx.Command{Name: "cache [all] [key value|[name value help]]", Help: "查看修改添加配置", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			all := false // {{{
			if len(arg) > 0 && arg[0] == "all" {
				arg = arg[1:]
				all = true
			}

			m.Target.BackTrace(func(s *ctx.Context) bool {
				switch len(arg) {
				case 0:
					for k, v := range s.Caches {
						m.Echo("%s(%s): %s\n", k, v.Value, v.Name)
					}
				case 1:
					if v, ok := s.Caches[arg[0]]; ok {
						m.Echo("%s: %s\n", v.Name, v.Help)
					}
				case 2:
					if s != m.Target {
						return false
					}

					switch arg[0] {
					case "delete":
						if _, ok := s.Caches[arg[1]]; ok {
							delete(s.Caches, arg[1])
						}
					default:
						if _, ok := s.Caches[arg[0]]; ok {
							m.Echo("%s: %s\n", arg[0], s.Cap(arg[0], arg[1:]...))
						}
					}
				case 4:
					s.Cap(arg[0], arg[1:]...)
					return false
				}

				return all
			})
			return ""
			// }}}
		}},
		"exit": &ctx.Command{Name: "exit", Help: "退出", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			cli, ok := m.Target.Server.(*CLI) // {{{
			if !ok {
				cli, ok = m.Context.Server.(*CLI)
			}
			if ok {
				if !cli.exit {
					m.Echo(c.Conf("结束语"))
					cli.Context.Exit(m)
				}
				cli.exit = true
			}

			return ""
			// }}}
		}},
		"remote": &ctx.Command{Name: "remote master|slave listen|dial address protocol", Help: "建立远程连接", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
			case 4:
				if arg[0] == "master" {
					if arg[1] == "dial" {
					} else {
					}
				} else {
					if arg[1] == "listen" {
						s := c.Root.Find(arg[3])
						m.Message.Spawn(s, arg[2]).Cmd("listen", arg[2])
					} else {
					}
				}
			}
			return ""
			// }}}
		}},
		"open": &ctx.Command{Name: "open address protocl", Help: "建立远程连接",
			Options: map[string]string{"io": "读写流"},
			Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
				if m.Has("io") {
					m.Start(fmt.Sprintf("PTS%d", c.Capi("nterm")), arg[1])
				} else {
					switch arg[1] {
					case "tcp":
					}
				}
				// {{{
				return ""
				// }}}
			}},
		"void": &ctx.Command{Name: "", Help: "", Hand: nil},
	},
	Messages: make(chan *ctx.Message, 10),
}

func init() {
	cli := &CLI{}
	cli.Context = Index
	ctx.Index.Register(Index, cli)
}
