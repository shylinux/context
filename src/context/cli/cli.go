package cli // {{{
// }}}
import ( // {{{
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// }}}

type CLI struct {
	in   *os.File
	ins  []*os.File
	bio  *bufio.Reader
	bios []*bufio.Reader
	out  *os.File

	history []map[string]string
	alias   map[string]string
	exit    chan bool
	next    string

	target *ctx.Context
	*ctx.Context
}

func (cli *CLI) push(f *os.File) { // {{{
	if cli.ins == nil || cli.bios == nil {
		cli.ins = make([]*os.File, 0, 3)
		cli.bios = make([]*bufio.Reader, 0, 3)
	}

	cli.in = f
	cli.ins = append(cli.ins, cli.in)
	cli.bio = bufio.NewReader(f)
	cli.bios = append(cli.bios, cli.bio)
}

// }}}
func (cli *CLI) parse() bool { // {{{
	if len(cli.ins) == 1 {
		cli.echo(cli.Conf("PS1"))
	}

	line := ""
	if cli.next == "" {
		l, e := cli.bio.ReadString('\n')
		if e == io.EOF {
			l := len(cli.ins)
			if l == 1 {
				if cli.Conf("slient") != "yes" {
					cli.echo("\n")
					cli.echo(cli.Conf("结束语"))
				}
				cli.echo("\n")
				cli.exit <- true
				return false

			} else {
				cli.ins = cli.ins[:l-1]
				cli.bios = cli.bios[:l-1]
				cli.in = cli.ins[l-2]
				cli.bio = cli.bios[l-2]
				return true
			}
		}
		cli.Check(e)
		line = l
	} else {
		line = cli.next
		if len(cli.ins) == 1 {
			cli.echo(line)
			cli.echo("\n")
		}
	}

	if len(cli.ins) > 1 {
		cli.echo(cli.Conf("PS1"))
		cli.echo(line)
	}
	cli.next = ""

	if len(line) == 1 {
		return true
	}

	line = strings.TrimSpace(line)
	if line[0] == '#' {
		return true
	}
	ls := strings.Split(line, " ")

	msg := &ctx.Message{Wait: make(chan bool)}
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

	cli.Post(msg)

	for _, v := range msg.Meta["result"] {
		cli.echo(v)
	}

	return true
}

// }}}
func (cli *CLI) deal(msg *ctx.Message) bool { // {{{
	defer func() {
		if e := recover(); e != nil {
			msg.Echo("%s", e)
			debug.PrintStack()
			log.Println(e)
		}
	}()

	detail := msg.Meta["detail"]
	switch cli.Conf("mode") {
	case "local":
		if a, ok := cli.alias[detail[0]]; ok {
			detail[0] = a
		}

		if _, ok := cli.Commands[detail[0]]; ok {
			cli.next = cli.Cmd(msg, detail...)
		} else if _, ok := cli.target.Commands[detail[0]]; ok {
			cli.target.Message = msg
			cli.next = cli.target.Cmd(msg, detail...)
		} else {
			cmd := exec.Command(detail[0], detail[1:]...)
			v, e := cmd.CombinedOutput()
			if e != nil {
				msg.Echo("%s\n", e)
			}
			msg.Echo(string(v))
			log.Println(cli.Name, "command:", detail)
		}
	}
	cli.history = append(cli.history, map[string]string{
		"time":  time.Now().Format("15:04:05"),
		"index": fmt.Sprintf("%d", len(cli.history)),
		"cli":   strings.Join(detail, " "),
	})
	return true
}

// }}}
func (cli *CLI) echo(str string, arg ...interface{}) { // {{{
	if len(cli.ins) == 1 || cli.Conf("slient") != "yes" {
		fmt.Fprintf(cli.out, str, arg...)
	}
}

// }}}

func (cli *CLI) Begin() bool { // {{{
	cli.history = make([]map[string]string, 0, 100)
	cli.alias = make(map[string]string, 10)
	cli.exit = make(chan bool)
	cli.target = cli.Context

	if f, e := os.Open(cli.Conf("init.sh")); e == nil {
		cli.push(f)
	}

	if cli.Conf("slient") != "yes" {
		cli.echo("\n")
		cli.echo(cli.Conf("开场白"))
		cli.echo("\n")
	}

	return true
}

// }}}
func (cli *CLI) Start() bool { // {{{
	go func() {
		for cli.parse() {
		}
	}()

	for {
		msg := cli.Get()
		msg.End(cli.deal(msg))
	}

	return true
}

// }}}
func (cli *CLI) Spawn(c *ctx.Context, arg ...string) ctx.Server { // {{{
	s := new(CLI)
	s.Context = c
	return s
}

// }}}

var Index = &ctx.Context{Name: "cli", Help: "本地控制",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"开场白":  &ctx.Config{Name: "开场白", Value: "你好，命令行", Help: "开场白"},
		"结束语":  &ctx.Config{Name: "结束语", Value: "再见，命令行", Help: "结束语"},
		"mode": &ctx.Config{Name: "mode", Value: "local", Help: "命令执行模式"},
		"io": &ctx.Config{Name: "io", Value: "stdout", Help: "输入输出", Hand: func(c *ctx.Context, arg string) string {
			cli := c.Server.(*CLI) // {{{
			cli.out = os.Stdout
			cli.push(os.Stdin)

			return arg
			// }}}
		}},
		"slient":  &ctx.Config{Name: "slient", Value: "yes", Help: "静默启动"},
		"init.sh": &ctx.Config{Name: "init.sh", Value: "etc/hi.sh", Help: "启动脚本"},
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
		"cache": &ctx.Command{"cache [name [value [help]]]", "查看修改添加配置", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			cli := c.Server.(*CLI) // {{{

			switch len(arg) {
			case 1:
				for k, v := range cli.target.Caches {
					msg.Echo("%s(%s): %s\n", k, v.Value, v.Help)
				}
			case 2:
				if v, ok := cli.target.Caches[arg[1]]; ok {
					msg.Echo("%s: %s\n", v.Name, v.Help)
				}
			case 3:
				switch arg[1] {
				case "delete":
					if _, ok := cli.target.Caches[arg[2]]; ok {
						delete(cli.target.Caches, arg[2])
					}
				default:
					if _, ok := cli.target.Caches[arg[1]]; ok {
						msg.Echo("%s: %s\n", arg[1], cli.target.Cap(arg[1:]...))
					}
				}
			case 5:
				cli.target.Cap(arg[1:]...)
			}
			return ""
			// }}}
		}},
		"config": &ctx.Command{"config [name [value [help]]]", "查看修改添加配置", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			cli := c.Server.(*CLI) // {{{

			switch len(arg) {
			case 1:
				for k, v := range cli.target.Configs {
					msg.Echo("%s(%s): %s\n", k, v.Value, v.Help)
				}
			case 2:
				if v, ok := cli.target.Configs[arg[1]]; ok {
					msg.Echo("%s: %s\n", v.Name, v.Help)
				}
			case 3:
				switch arg[1] {
				case "delete":
					if _, ok := cli.target.Configs[arg[2]]; ok {
						delete(cli.target.Configs, arg[2])
					}
				default:
					if _, ok := cli.target.Configs[arg[1]]; ok {
						cli.target.Conf(arg[1:]...)
					}
				}
			case 5:
				cli.target.Conf(arg[1:]...)
			}
			return ""
			// }}}
		}},
		"command": &ctx.Command{"command [name [value [help]]]", "查看修改添加配置", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			cli := c.Server.(*CLI) // {{{

			switch len(arg) {
			case 1:
				for k, v := range cli.target.Commands {
					msg.Echo("%s: %s\n", k, v.Help)
				}
			case 2:
				if v, ok := cli.target.Commands[arg[1]]; ok {
					msg.Echo("%s: %s\n", v.Name, v.Help)
				}
			case 3:
				switch arg[1] {
				case "delete":
					if _, ok := cli.target.Commands[arg[2]]; ok {
						delete(cli.target.Commands, arg[2])
					}
				}
			}

			msg.Echo("\n")
			return ""
			// }}}
		}},
		"source": &ctx.Command{"source file", "运行脚本", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			cli := c.Server.(*CLI) // {{{

			f, e := os.Open(arg[1])
			c.Check(e)
			cli.push(f)

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
		"message": &ctx.Command{"message detail...", "查看上下文", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			msg.Meta["detail"] = arg[1:]
			msg.Target.Post(msg)
			return ""
		}},
		"server": &ctx.Command{"server start|stop|switch", "服务启动停止切换", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			s := msg.Target // {{{
			switch len(arg) {
			case 1:
				return "server start"
			case 2:
				switch arg[1] {
				case "start":
					go s.Start()
				case "stop":
				case "switch":
				}
			}
			return ""
			// }}}
		}},
		"context": &ctx.Command{"context [spawn|fork|find|search name [switch]]|root|back|home", "查看上下文", func(c *ctx.Context, msg *ctx.Message, arg ...string) string {
			cli := c.Server.(*CLI) // {{{

			switch len(arg) {
			case 1:
				cs := []*ctx.Context{cli.target}
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
			case 2, 3, 4, 5:
				switch arg[1] {
				case "spawn":
					msg.Target.Spawn(arg[2])
				case "root":
					cli.target = cli.Context.Root
				case "back":
					if cli.Context.Context != nil {
						cli.target = cli.Context.Context
					}
				case "home":
					cli.target = cli.Context
				case "find":
					cs := c.Root.Find(strings.Split(arg[2], "."))
					if cs != nil {
						msg.Echo("%s: %s\n", cs.Name, cs.Help)
						if len(arg) == 4 {
							cli.target = cs
						}
					}
				case "search":
					cs := c.Root.Search(arg[2])
					for i, v := range cs {
						msg.Echo("[%d] %s: %s\n", i, v.Name, v.Help)
					}

					if len(arg) == 5 {
						n, e := strconv.Atoi(arg[4])
						if 0 <= n && n < len(cs) && e == nil {
							cli.target = cs[0]
						} else {
							msg.Echo("参数错误(0<=n<%s)", len(cs))
						}
					}

					if len(arg) == 4 {
						cli.target = cs[0]
					}
				default:
					cs := c.Root.Find(strings.Split(arg[1], "."))
					if cs != nil {
						msg.Echo("%s: %s\n", cs.Name, cs.Help)
						if cs != nil {
							cli.target = cs
						}
					}
				}
			}

			return ""
			// }}}
		}},
	},
	Messages: make(chan *ctx.Message, 10),
}

func init() {
	self := &CLI{}
	self.Context = Index
	ctx.Index.Register(Index, self)
}
