package cli

import ( // {{{
	"bufio"
	"context"
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
		cli.Check(e)
		line = ls

		if len(cli.ins) > 1 || cli.Conf("slient") != "yes" {
			cli.echo(line)
		}

		if len(line) == 1 {
			return true
		}
	} else {
		line = cli.next
		cli.next = ""

		if cli.Conf("slient") != "yes" {
			cli.echo(line)
			cli.echo("\n")
		}
	}

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
	cli.alias = make(map[string]string, 10)
	cli.target = cli.Context
	return cli.Server
}

// }}}
func (cli *CLI) Start(m *ctx.Message) bool { // {{{
	if detail, ok := m.Data["detail"]; ok {
		io := detail.(io.ReadWriteCloser)
		cli.out = io
		cli.push(io)

		cli.echo("%s\n", cli.Conf("开场白"))

		if f, e := os.Open(cli.Conf("init.sh")); e == nil {
			cli.push(f)
		}

		go cli.Safe(m, func(c *ctx.Context, m *ctx.Message) {
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
	}
	c.Commands = cli.Commands
	c.Messages = make(chan *ctx.Message, 10)

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

var Index = &ctx.Context{Name: "cli", Help: "本地控制",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"开场白":    &ctx.Config{Name: "开场白", Value: "\n~~~  Hello Context & Message World  ~~~\n", Help: "开场白"},
		"结束语":    &ctx.Config{Name: "结束语", Value: "\n~~~  Byebye Context & Message World  ~~~\n", Help: "结束语"},
		"slient": &ctx.Config{Name: "slient", Value: "yes", Help: "屏蔽脚本输出"},

		"PS1": &ctx.Config{Name: "PS1", Value: "etcvpn>", Help: "命令行提示符", Hand: func(c *ctx.Context, arg string) string {
			cli := c.Server.(*CLI) // {{{
			if cli != nil && cli.target != nil {
				arg = cli.target.Name + ">"
			}
			return fmt.Sprintf("%d[%s]\033[32m%s\033[0m ", len(cli.history), time.Now().Format("15:04:05"), arg)
			// }}}
		}},
	},
	Commands: map[string]*ctx.Command{
		"context": &ctx.Command{"context [spawn|find|search name [which]]|root|back|home", "查看上下文", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			cli, ok := c.Server.(*CLI) // {{{
			switch len(arg) {
			case 1:
				cs := []*ctx.Context{msg.Target}
				for i := 0; i < len(cs); i++ {
					if len(cs[i].Contexts) > 0 {
						msg.Echo("%s: ", cs[i].Name)
						for k, v := range cs[i].Contexts {
							cs = append(cs, v)
							msg.Echo("%s, ", k)
						}
						msg.Echo("\n")
					}
				}
			case 2:
				switch arg[1] {
				case "root":
					if ok {
						cli.target = cli.Context.Root
					} else {
						msg.Target = msg.Target.Root
					}
				case "back":
					if ok {
						if cli.Context.Context != nil {
							cli.target = cli.Context.Context
						}
					} else {
						if msg.Target.Context != nil {
							msg.Target = msg.Target.Context
						}
					}
				case "home":
					if ok {
						cli.target = cli.Context
					} else {
						msg.Target = msg.Context
					}
				default:
					// if cs := msg.Target.Find(strings.Split(arg[1], ".")); cs != nil {
					if cs := c.Root.Search(arg[1]); cs != nil && len(cs) > 0 {
						if ok {
							cli.target = cs[0]
						} else {
							msg.Target = cs[0]
						}
					}
				}
			case 3, 4:
				switch arg[1] {
				case "spawn":
					msg.Target.Spawn(msg, arg[2])
				case "find":
					cs := msg.Target.Find(strings.Split(arg[2], "."))
					if cs != nil {
						msg.Echo("%s: %s\n", cs.Name, cs.Help)
						if len(arg) == 4 {
							if ok {
								cli.target = cs
							} else {
								msg.Target = cs
							}
						}
					}
				case "search":
					cs := msg.Target.Search(arg[2])
					for i, v := range cs {
						msg.Echo("[%d] %s: %s\n", i, v.Name, v.Help)
					}

					if len(arg) == 4 {
						n, e := strconv.Atoi(arg[3])
						if 0 <= n && n < len(cs) && e == nil {
							if ok {
								cli.target = cs[n]
							} else {
								msg.Target = cs[n]
							}
						} else {
							msg.Echo("参数错误(0<=n<%s)", len(cs))
						}
					}
				}
			}

			return ""
			// }}}
		}},
		"message": &ctx.Command{"message detail...", "查看上下文", func(c *ctx.Context, m *ctx.Message, arg ...string) string {
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
		"server": &ctx.Command{"server start|stop|switch", "服务启动停止切换", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			s := msg.Target // {{{
			switch len(arg) {
			case 1:
				return "server start"
			case 2:
				switch arg[1] {
				case "start":
					if s != nil {
						go s.Start(msg)
					}
				case "stop":
				case "switch":
				}
			}
			return ""
			// }}}
		}},
		"source": &ctx.Command{"source file", "运行脚本", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			cli := c.Server.(*CLI) // {{{
			switch len(arg) {
			case 2:
				f, e := os.Open(arg[1])
				c.Check(e)
				cli.push(f)
			}

			return ""
			// }}}
		}},
		"alias": &ctx.Command{"alias [short [long]]", "查看日志", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			cli := c.Server.(*CLI) // {{{
			switch len(arg) {
			case 1:
				for k, v := range cli.alias {
					msg.Echo("%s: %s\n", k, v)
				}
			case 3:
				switch arg[1] {
				case "delete":
					delete(cli.alias, arg[2])
				default:
					cli.alias[arg[1]] = arg[2]
					msg.Echo("%s: %s\n", arg[1], cli.alias[arg[1]])
				}
			}
			return ""
			// }}}
		}},
		"history": &ctx.Command{"history number", "查看日志", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			cli := c.Server.(*CLI) // {{{
			switch len(arg) {
			case 1:
				for i, v := range cli.history {
					msg.Echo("%d %s %s\n", i, v["time"], v["cli"])
				}
			case 2:
				n, e := strconv.Atoi(arg[1])
				if e == nil && 0 <= n && n < len(cli.history) {
					return cli.history[n]["cli"]
				}
			}
			return ""
			// }}}
		}},
		"command": &ctx.Command{"command [name [value [help]]]", "查看修改添加配置", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			switch len(arg) { // {{{
			case 1:
				for k, v := range msg.Target.Commands {
					msg.Echo("%s: %s\n", k, v.Help)
				}
			case 2:
				if v, ok := msg.Target.Commands[arg[1]]; ok {
					msg.Echo("%s: %s\n", v.Name, v.Help)
				}
			case 3:
				switch arg[1] {
				case "delete":
					if _, ok := msg.Target.Commands[arg[2]]; ok {
						delete(msg.Target.Commands, arg[2])
					}
				}
			}

			msg.Echo("\n")
			return ""
			// }}}
		}},
		"config": &ctx.Command{"config [name [value [help]]]", "查看修改添加配置", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			switch len(arg) { // {{{
			case 1:
				for k, v := range msg.Target.Configs {
					msg.Echo("%s(%s): %s\n", k, v.Value, v.Help)
				}
			case 2:
				if v, ok := msg.Target.Configs[arg[1]]; ok {
					msg.Echo("%s: %s\n", v.Name, v.Help)
				}
			case 3:
				switch arg[1] {
				case "delete":
					if _, ok := msg.Target.Configs[arg[2]]; ok {
						delete(msg.Target.Configs, arg[2])
					}
				default:
					if _, ok := msg.Target.Configs[arg[1]]; ok {
						msg.Target.Conf(arg[1:]...)
					}
				}
			case 5:
				msg.Target.Conf(arg[1:]...)
			}
			return ""
			// }}}
		}},
		"cache": &ctx.Command{"cache [name [value [help]]]", "查看修改添加配置", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			switch len(arg) { // {{{
			case 1:
				for k, v := range msg.Target.Caches {
					msg.Echo("%s(%s): %s\n", k, v.Value, v.Help)
				}
			case 2:
				if v, ok := msg.Target.Caches[arg[1]]; ok {
					msg.Echo("%s: %s\n", v.Name, v.Help)
				}
			case 3:
				switch arg[1] {
				case "delete":
					if _, ok := msg.Target.Caches[arg[2]]; ok {
						delete(msg.Target.Caches, arg[2])
					}
				default:
					if _, ok := msg.Target.Caches[arg[1]]; ok {
						msg.Echo("%s: %s\n", arg[1], msg.Target.Cap(arg[1:]...))
					}
				}
			case 5:
				msg.Target.Cap(arg[1:]...)
			}
			return ""
			// }}}
		}},
		"exit": &ctx.Command{"exit", "退出", func(c *ctx.Context, m *ctx.Message, arg ...string) string {
			cli, ok := m.Target.Server.(*CLI)
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
		}},
		"remote": &ctx.Command{"remote master|slave listen|dial address protocol", "建立远程连接", func(c *ctx.Context, m *ctx.Message, arg ...string) string {
			switch len(arg) { // {{{
			case 1:
			case 5:
				if arg[1] == "master" {
					if arg[2] == "dial" {
					} else {
					}
				} else {
					if arg[2] == "listen" {
						s := c.Root.Find(strings.Split(arg[4], "."))
						m.Message.Spawn(s, arg[3]).Cmd("listen", arg[3])
					} else {
					}
				}
			}
			return ""
			// }}}
		}},
		"accept": &ctx.Command{"accept address protocl", "建立远程连接", func(c *ctx.Context, m *ctx.Message, arg ...string) string {
			m.Start(arg[1:]...) // {{{
			return ""
			// }}}
		}},
	},
	Messages: make(chan *ctx.Message, 10),
}

func init() {
	cli := &CLI{}
	cli.Context = Index
	ctx.Index.Register(Index, cli)
}
