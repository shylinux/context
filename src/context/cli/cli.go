package cli // {{{
// }}}
import ( // {{{
	"context"

	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

// }}}

type CLI struct {
	in   io.ReadCloser
	ins  []io.ReadCloser
	bio  *bufio.Reader
	bios []*bufio.Reader
	bufs [][]byte

	out io.WriteCloser

	alias map[string]string
	back  string
	next  string
	exit  bool

	lex    *ctx.Message
	target *ctx.Context

	*ctx.Message
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
	if len(cli.ins) == 1 || cli.Conf("slient") != "yes" {
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
			line, cli.back = cli.back, ""
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

	for i := 0; i < len(ls); i++ {
		ls[i] = strings.TrimSpace(ls[i])
		if ls[i] == "" {
			continue
		}
		if ls[i][0] == '#' {
			break
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

		if cli.lex != nil && len(ls[i]) > 1 {
			switch ls[i][0] {
			case '"', '\'':
				ls[i] = ls[i][1 : len(ls[i])-1]
			}
		}

		msg.Add("detail", ls[i])
	}

	msg.Wait = make(chan bool)
	msg.Post(cli.Context)

	cli.echo(strings.Join(msg.Meta["result"], ""))
	cli.target = msg.Target
	cli.back = line
	return true
}

// }}}

func (cli *CLI) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(CLI)
	s.Context = c
	return s
}

// }}}
func (cli *CLI) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	cli.Configs["slient"] = &ctx.Config{Name: "屏蔽脚本输出(yes/no)", Value: "yes", Help: "屏蔽脚本输出的信息，yes:屏蔽，no:不屏蔽"}
	cli.Configs["PS1"] = &ctx.Config{Name: "命令行提示符(target/detail)", Value: "target", Help: "命令行提示符，target:显示当前模块，detail:显示详细信息", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
		if len(arg) > 0 { // {{{
			return arg[0]
		}

		ps := make([]string, 0, 3)

		if cli, ok := m.Target.Server.(*CLI); ok && cli.target != nil {
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

			lex := m.Find(arg[0], true)
			m.Assert(lex != nil, "词法解析模块不存在")
			if lex.Cap("status") != "start" {
				lex.Target.Start(lex)
				m.Spawn(lex.Target).Cmd("train", "'[^']*'")
				m.Spawn(lex.Target).Cmd("train", "\"[^\"]*\"")
				m.Spawn(lex.Target).Cmd("train", "[^ \t\n]+")
				m.Spawn(lex.Target).Cmd("train", "[ \n\t]+", "void", "void")
				m.Spawn(lex.Target).Cmd("train", "#[^\n]*\n", "void", "void")
			}
			cli.lex = lex
		}
		return x.Value
		// }}}
	}}

	if len(arg) > 0 {
		cli.Configs["init.sh"] = &ctx.Config{Name: "启动脚本", Value: arg[0], Help: "模块启动时自动运行的脚本"}
	}

	cli.Context.Master = cli.Context
	if cli.Context != Index {
		cli.Owner = nil
	}
	if cli.Context == Index {
		Pulse = m
	}

	cli.target = cli.Context
	cli.alias = map[string]string{
		"~": "context",
		"!": "command",
		"@": "config",
		"$": "cache",
		"&": "server",
		"*": "message",
	}

	return cli
}

// }}}
func (cli *CLI) Start(m *ctx.Message, arg ...string) bool { // {{{
	cli.Message = m

	if stream, ok := m.Data["io"]; ok {
		io := stream.(io.ReadWriteCloser)
		cli.out = io
		cli.push(io)

		if m.Has("master") {
			m.Log("info", nil, "master terminal")
			if cli.bufs == nil {
				cli.bufs = make([][]byte, 0, 10)
			}
			for {
				b := make([]byte, 128)
				n, e := cli.bio.Read(b)
				m.Log("info", nil, "read %d", n)
				m.Assert(e)
				cli.bufs = append(cli.bufs, b)
			}
			return true
		} else {
			if cli.Owner == nil {
				if msg := m.Find("aaa", true); msg != nil {
					username := ""
					cli.echo("username>")
					fmt.Fscanln(cli.in, &username)

					password := ""
					cli.echo("password>")
					fmt.Fscanln(cli.in, &password)

					msg.Name = "aaa"
					msg.Wait = make(chan bool)
					if msg.Cmd("login", username, password) == "" {
						cli.echo("登录失败")
						m.Cmd("exit")
						cli.out.Close()
						cli.in.Close()
						return false
					}

					if cli.Sessions == nil {
						cli.Sessions = make(map[string]*ctx.Message)
					}
					cli.Sessions["aaa"] = msg
				}
			} else {
				m.Cap("stream", "stdout")
			}

			m.Log("info", nil, "slaver terminal")
			m.Log("info", nil, "open %s", m.Conf("init.sh"))
			if f, e := os.Open(m.Conf("init.sh")); e == nil {
				cli.push(f)
			}

			go m.AssertOne(m, true, func(m *ctx.Message) {
				for cli.parse(m) {
				}
			})
		}
	}

	m.Capi("nterm", 1)
	defer m.Capi("nterm", -1)

	m.Deal(nil, func(msg *ctx.Message, arg ...string) bool {
		if msg.Get("result") == "error: " {
			msg.Log("system", nil, "%v", msg.Meta["detail"])

			msg.Set("result")
			msg.Set("append")
			c := exec.Command(msg.Meta["detail"][0], msg.Meta["detail"][1:]...)

			if len(cli.ins) == 1 && cli.Context == Index {
				c.Stdin, c.Stdout, c.Stderr = cli.in, cli.out, cli.out
				msg.Assert(c.Start())
				msg.Assert(c.Wait())
			} else {
				if out, e := c.CombinedOutput(); e == nil {
					msg.Echo(string(out))
				} else {
					msg.Echo("error: ")
					msg.Echo("%s\n", e)
				}
			}
		}

		cli.target = msg.Target
		return cli.exit == false
	})

	return true
}

// }}}
func (cli *CLI) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch cli.Context {
	case m.Target:
		if cli.Context == Index {
			return false
		}

		if len(cli.Context.Requests) == 0 {
			m.Log("info", nil, "%s close %v", Pulse.Cap("nterm"), arg)
		}
	case m.Source:
		if m.Name == "aaa" {
			msg := m.Spawn(cli.Context)
			msg.Master = cli.Context
			if !cli.Context.Close(msg, arg...) {
				return false
			}
		}
	}

	return true
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "cli", Help: "管理中心",
	Caches: map[string]*ctx.Cache{
		"nterm": &ctx.Cache{Name: "终端数量", Value: "0", Help: "已经运行的终端数量"},
	},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"source": &ctx.Command{Name: "source file", Help: "运行脚本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			cli := c.Server.(*CLI) // {{{
			switch len(arg) {
			case 1:
				f, e := os.Open(arg[0])
				m.Assert(e)
				cli.push(f)
			}

			// }}}
		}},
		"return": &ctx.Command{Name: "return", Help: "运行脚本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			cli := c.Server.(*CLI) // {{{
			cli.bio.Discard(cli.bio.Buffered())
			// }}}
		}},
		"alias": &ctx.Command{Name: "alias [short [long]]", Help: "查看日志", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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
			// }}}
		}},
		"remote": &ctx.Command{Name: "remote [send args...]|[[master|slaver] listen|dial address protocol]", Help: "建立远程连接",
			Formats: map[string]int{"send": -1, "master": 0, "slaver": 0, "listen": 1, "dial": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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

					return
				}

				action := "dial"
				if m.Has("listen") {
					action = "listen"
				}

				msg := m.Find(m.Get("args"), true)

				if m.Has("master") {
					msg.Template = msg.Spawn(msg.Source).Add("option", "master")
				}
				msg.Cmd(action, m.Get(action))

			}},
		// }}}
		"open": &ctx.Command{Name: "open [master|slaver] [script [log]]", Help: "建立远程连接",
			Options: map[string]string{"master": "主控终端", "slaver": "被控终端", "args": "启动参数", "io": "读写流"},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				m.Start(fmt.Sprintf("PTS%d", m.Capi("nterm")), "管理终端", "void.sh") // {{{
				// }}}
			}},
		"master": &ctx.Command{Name: "open [master|slaver] [script [log]]", Help: "建立远程连接",
			Options: map[string]string{"master": "主控终端", "slaver": "被控终端", "args": "启动参数", "io": "读写流"},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				cli, ok := c.Server.(*CLI) // {{{
				m.Assert(ok, "模块类型错误")
				m.Assert(m.Target != c, "模块是主控模块")

				msg := m.Spawn(c)
				msg.Start(fmt.Sprintf("PTS%d", cli.Capi("nterm")), arg[0], arg[1:]...)
				m.Target.Master = msg.Target
				// }}}
			}},
	},
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
