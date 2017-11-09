package cli // {{{
// }}}
import ( // {{{
	"bufio"
	"context"
	_ "context/tcp"
	_ "context/web"
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
func (cli *CLI) parse() bool { // {{{
	if len(cli.ins) == 1 || cli.Conf("slient") != "yes" {
		cli.echo(cli.Conf("PS1"))
	}

	line := ""
	if cli.next == "" {
		ls, e := cli.bio.ReadString('\n')
		if e == io.EOF {
			l := len(cli.ins)
			if l == 1 {
				cli.echo("\n%s\n", cli.Conf("结束语"))
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

	msg := &ctx.Message{Wait: make(chan bool)}
	msg.Message = cli.Resource[0]
	msg.Context = cli.Context
	msg.Target = cli.target

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

func (cli *CLI) Begin(m *ctx.Message) ctx.Server { // {{{
	cli.history = make([]map[string]string, 0, 100)
	cli.target = cli.Context
	return cli.Server
}

// }}}
func (cli *CLI) Start(m *ctx.Message) bool { // {{{
	cli.Capi("nterm", 1)
	defer cli.Capi("nterm", -1)

	if stream, ok := m.Data["io"]; ok {
		io := stream.(io.ReadWriteCloser)
		cli.out = io
		cli.push(io)

		cli.echo("%s\n", cli.Conf("开场白"))

		if f, e := os.Open(cli.Conf("init.sh")); e == nil {
			cli.push(f)
		}

		defer recover()
		go cli.AssertOne(m, func(c *ctx.Context, m *ctx.Message) {
			for cli.parse() {
			}
		})
	}

	for cli.Deal(func(msg *ctx.Message) bool {
		arg := msg.Meta["detail"]
		if a, ok := cli.alias[arg[0]]; ok {
			arg[0] = a
		}
		return true

	}, func(msg *ctx.Message) bool {
		arg := msg.Meta["detail"]
		cli.history = append(cli.history, map[string]string{
			"time":  time.Now().Format("15:04:05"),
			"index": fmt.Sprintf("%d", len(cli.history)),
			"cli":   strings.Join(arg, " "),
		})

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
	c.Caches = map[string]*ctx.Cache{
		"status": &ctx.Cache{Name: "status", Value: "stop", Help: "服务状态"},
	}
	c.Configs = map[string]*ctx.Config{
		"address":  &ctx.Config{Name: "address", Value: arg[0], Help: "监听地址"},
		"protocol": &ctx.Config{Name: "protocol", Value: arg[1], Help: "监听协议"},
		"init.sh":  &ctx.Config{Name: "init.sh", Value: "", Help: "默认启动脚本"},
	}
	c.Commands = cli.Commands
	c.Messages = make(chan *ctx.Message, 10)

	s := new(CLI)
	s.Context = c
	s.alias = map[string]string{
		"~": "context",
		"!": "history",
		"@": "config",
		"$": "cache",
		"&": "server",
		"*": "message",
	}
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
		"nterm":  &ctx.Cache{Name: "nterm", Value: "0", Help: "终端数量"},
		"status": &ctx.Cache{Name: "status", Value: "stop", Help: "服务状态"},
		"nhistory": &ctx.Cache{Name: "nhistory", Value: "0", Help: "终端数量", Hand: func(c *ctx.Context, x *ctx.Cache, arg ...string) string {
			if cli, ok := c.Server.(*CLI); ok { // {{{
				return fmt.Sprintf("%d", len(cli.history))
			}

			return x.Value
			// }}}
		}},
	},
	Configs: map[string]*ctx.Config{
		"开场白":    &ctx.Config{Name: "开场白", Value: "\n~~~  Hello Context & Message World  ~~~\n", Help: "开场白"},
		"结束语":    &ctx.Config{Name: "结束语", Value: "\n~~~  Byebye Context & Message World  ~~~\n", Help: "结束语"},
		"slient": &ctx.Config{Name: "slient", Value: "yes", Help: "屏蔽脚本输出"},

		"PS1": &ctx.Config{Name: "PS1", Value: "target", Help: "命令行提示符", Hand: func(c *ctx.Context, x *ctx.Config, arg ...string) string {
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
		"context": &ctx.Command{Name: "context [spawn|find|search name [which]]|root|back|home", Help: "查看上下文", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			cli, ok := c.Server.(*CLI) // {{{
			switch len(arg) {
			case 0:
				// cs := []*ctx.Context{m.Target}
				cs := []*ctx.Context{m.Target.Root}
				for i := 0; i < len(cs); i++ {
					if len(cs[i].Contexts) > 0 {
						m.Echo("%s: ", cs[i].Name)
						for k, v := range cs[i].Contexts {
							cs = append(cs, v)
							m.Echo("%s, ", k)
						}
						m.Echo("\n")
					}
				}
			case 1:
				switch arg[0] {
				case "root":
					if ok {
						cli.target = cli.Context.Root
					} else {
						m.Target = m.Target.Root
					}
				case "back":
					if ok {
						if cli.Context.Context != nil {
							cli.target = cli.Context.Context
						}
					} else {
						if m.Target.Context != nil {
							m.Target = m.Target.Context
						}
					}
				case "home":
					if ok {
						cli.target = cli.Context
					} else {
						m.Target = m.Context
					}
				default:
					// if cs := m.Target.Find(strings.Split(arg[1], ".")); cs != nil {
					if cs := c.Root.Search(arg[0]); cs != nil && len(cs) > 0 {
						if ok {
							cli.target = cs[0]
						} else {
							m.Target = cs[0]
						}
					}
				}
			case 2, 3:
				switch arg[0] {
				case "spawn":
					m.Target.Spawn(m, arg[1])
				case "find":
					cs := m.Target.Find(arg[1])
					if cs != nil {
						m.Echo("%s: %s\n", cs.Name, cs.Help)
						if len(arg) == 4 {
							if ok {
								cli.target = cs
							} else {
								m.Target = cs
							}
						}
					}
				case "search":
					cs := m.Target.Search(arg[1])
					for i, v := range cs {
						m.Echo("[%d] %s: %s\n", i, v.Name, v.Help)
					}

					if len(arg) == 3 {
						n, e := strconv.Atoi(arg[2])
						if 0 <= n && n < len(cs) && e == nil {
							if ok {
								cli.target = cs[n]
							} else {
								m.Target = cs[n]
							}
						} else {
							m.Echo("参数错误(0<=n<%s)", len(cs))
						}
					}
				}
			}

			return ""
			// }}}
		}},
		"message": &ctx.Command{Name: "message detail...", Help: "查看上下文", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			// {{{
			ms := []*ctx.Message{ctx.Pulse}
			for i := 0; i < len(ms); i++ {
				m.Echo("%d %s.%s -> %s.%d\n", ms[i].Code, ms[i].Context.Name, ms[i].Name, ms[i].Target.Name, ms[i].Index)
				// m.Echo("%d %s %s.%s -> %s.%d\n", ms[i].Code, ms[i].Time.Format("2006/01/02 15:03:04"), ms[i].Context.Name, ms[i].Name, ms[i].Target.Name, ms[i].Index)
				ms = append(ms, ms[i].Messages...)
			}
			return ""
			// }}}
		}},
		"server": &ctx.Command{Name: "server start|stop|switch", Help: "服务启动停止切换", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			s := m.Target // {{{
			switch len(arg) {
			case 0:
				cs := []*ctx.Context{m.Target.Root}
				for i := 0; i < len(cs); i++ {
					if x, ok := cs[i].Caches["status"]; ok {
						m.Echo("%s(%s): %s\n", cs[i].Name, x.Value, cs[i].Help)
					}

					for _, v := range cs[i].Contexts {
						cs = append(cs, v)
					}
				}
				return "server start"
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
					log.Println("shy log why:", cli.history[n]["cli"])
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
		"command": &ctx.Command{Name: "command [name [value [help]]]", Help: "查看修改添加配置", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				for k, v := range m.Target.Commands {
					m.Echo("%s: %s\n", k, v.Help)
				}
			case 2:
				if v, ok := m.Target.Commands[arg[0]]; ok {
					m.Echo("%s: %s\n", v.Name, v.Help)
				}
			case 3:
				switch arg[0] {
				case "delete":
					if _, ok := m.Target.Commands[arg[1]]; ok {
						delete(m.Target.Commands, arg[1])
					}
				}
			}

			m.Echo("\n")
			return ""
			// }}}
		}},
		"config": &ctx.Command{Name: "config [name [value [help]]]", Help: "查看修改添加配置", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				for k, v := range m.Target.Configs {
					m.Echo("%s(%s): %s\n", k, v.Value, v.Help)
				}
			case 1:
				if v, ok := m.Target.Configs[arg[0]]; ok {
					m.Echo("%s: %s\n", v.Name, v.Help)
				}
			case 2:
				switch arg[0] {
				case "void":
					m.Target.Conf(arg[1], "")
				case "delete":
					if _, ok := m.Target.Configs[arg[1]]; ok {
						delete(m.Target.Configs, arg[1])
					}
				default:
					m.Target.Conf(arg[0], arg[1])
				}
			case 4:
				m.Target.Conf(arg[0], arg[1:]...)
			}
			return ""
			// }}}
		}},
		"cache": &ctx.Command{Name: "cache [name [value [help]]]", Help: "查看修改添加配置", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			switch len(arg) { // {{{
			case 0:
				for k, v := range m.Target.Caches {
					m.Echo("%s(%s): %s\n", k, v.Value, v.Help)
				}
			case 1:
				if v, ok := m.Target.Caches[arg[0]]; ok {
					m.Echo("%s: %s\n", v.Name, v.Help)
				}
			case 2:
				switch arg[0] {
				case "delete":
					if _, ok := m.Target.Caches[arg[1]]; ok {
						delete(m.Target.Caches, arg[1])
					}
				default:
					if _, ok := m.Target.Caches[arg[0]]; ok {
						m.Echo("%s: %s\n", arg[0], m.Target.Cap(arg[0], arg[1:]...))
					}
				}
			case 4:
				m.Target.Cap(arg[0], arg[1:]...)
			}
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
		"accept": &ctx.Command{Name: "accept address protocl", Help: "建立远程连接",
			Options: map[string]string{"io": "读写流"},
			Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
				m.Start(fmt.Sprintf("PTS%d", c.Capi("nterm")), arg[1]) // {{{
				return ""
				// }}}
			}},
		"void": &ctx.Command{Name: "", Help: "", Hand: nil},
	},
	Messages: make(chan *ctx.Message, 10),
}

func init() {
	cli := &CLI{alias: map[string]string{
		"~": "context",
		"!": "history",
		"@": "config",
		"$": "cache",
		"&": "server",
		"*": "message",
	}}
	cli.Context = Index
	ctx.Index.Register(Index, cli)
}
