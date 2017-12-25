package cli // {{{
// }}}
import ( // {{{
	"context"

	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// }}}

type CLI struct {
	bio *bufio.Reader
	out io.WriteCloser

	alias map[string]string
	back  string
	exit  bool

	lex    *ctx.Message
	target *ctx.Context
	wait   *ctx.Context

	*ctx.Message
	*ctx.Context
}

func (cli *CLI) echo(str string, arg ...interface{}) bool {
	if cli.out != nil {
		fmt.Fprintf(cli.out, str, arg...)
		return true
	}
	return false
}

func (cli *CLI) parse(m *ctx.Message) bool {
	line := m.Cap("next")
	if line == "" {
		cli.echo(m.Conf("PS1"))
		l, e := cli.bio.ReadString('\n')
		m.Assert(e)
		line = l
	}
	m.Cap("next", "")

	if line = strings.TrimSpace(line); len(line) == 0 {
		line, cli.back = cli.back, ""
	}
	if len(line) == 0 || line[0] == '#' {
		return true
	}

	ls := []string{}
	if cli.lex == nil {
		ls = strings.Split(line, " ")
	} else {
		lex := m.Spawn(cli.lex.Target())
		m.Assert(lex.Cmd("split", line, "void"))
		ls = lex.Meta["result"]
	}

	msg := m.Spawn(cli.target)

	for i := 0; i < len(ls); i++ {
		if ls[i] = strings.TrimSpace(ls[i]); ls[i] == "" {
			continue
		}
		if ls[i][0] == '#' {
			break
		}

		if r := rune(ls[i][0]); r == '$' || r == '_' || (!unicode.IsNumber(r) && !unicode.IsLetter(r)) {
			if c, ok := cli.alias[string(r)]; ok {
				if i == 0 {
					if msg.Add("detail", c); len(ls[i]) == 1 {
						continue
					}
					ls[i] = ls[i][1:]
				} else if len(ls[i]) > 1 {
					msg := m.Spawn(cli.target)
					m.Assert(msg.Exec(c, ls[i][1:]))
					ls[i] = msg.Get("result")
				}
			}
		}

		msg.Add("detail", ls[i])
	}

	msg.Wait = make(chan bool)
	msg.Post(cli.Context.Master())

	if result := strings.TrimRight(strings.Join(msg.Meta["result"], ""), "\n"); len(result) > 0 {
		cli.echo(m.Cap("result", result) + "\n")
	}
	m.Cap("target", cli.target.Name)
	m.Cap("back", line)
	m.Log("fuck", nil, "over")
	if cli.wait != nil {
		msg.Log("fuck", nil, "wait 1")
		<-cli.wait.Exit
		cli.wait = nil
		msg.Log("fuck", nil, "wait 2")
	} else {
		cli.target = msg.Target()
	}
	return !cli.exit
}

func (cli *CLI) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(CLI)
	s.Context = c
	return s
}

func (cli *CLI) Begin(m *ctx.Message, arg ...string) ctx.Server {
	cli.Caches["target"] = &ctx.Cache{Name: "操作目标", Value: "", Help: "操作目标"}
	cli.Caches["result"] = &ctx.Cache{Name: "前一条指令执行结果", Value: "", Help: "前一条指令执行结果"}
	cli.Caches["back"] = &ctx.Cache{Name: "前一条指令", Value: "", Help: "前一条指令"}
	cli.Caches["next"] = &ctx.Cache{Name: "下一条指令", Value: "", Help: "下一条指令"}

	cli.Configs["PS1"] = &ctx.Config{Name: "命令行提示符(target/detail)", Value: "target", Help: "命令行提示符，target:显示当前模块，detail:显示详细信息", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
		if len(arg) > 0 { // {{{
			return arg[0]
		}

		ps := make([]string, 0, 3)

		if cli, ok := m.Target().Server.(*CLI); ok && cli.target != nil {
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
			cli, ok := m.Target().Server.(*CLI)
			m.Assert(ok, "模块类型错误")

			lex := m.Find(arg[0], true)
			m.Assert(lex != nil, "词法解析模块不存在")
			if lex.Cap("status") != "start" {
				lex.Target().Start(lex)
				m.Spawn(lex.Target()).Cmd("train", "'[^']*'")
				m.Spawn(lex.Target()).Cmd("train", "\"[^\"]*\"")
				m.Spawn(lex.Target()).Cmd("train", "[^ \t\n]+")
				m.Spawn(lex.Target()).Cmd("train", "[ \n\t]+", "void", "void")
				m.Spawn(lex.Target()).Cmd("train", "#[^\n]*\n", "void", "void")
			}
			cli.lex = lex
		}
		return x.Value
		// }}}
	}}

	if len(arg) > 0 {
		cli.Configs["init.sh"] = &ctx.Config{Name: "启动脚本", Value: arg[0], Help: "模块启动时自动运行的脚本"}
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

func (cli *CLI) Start(m *ctx.Message, arg ...string) bool {

	if stream, ok := m.Data["io"]; ok {
		io := stream.(io.ReadWriteCloser)
		cli.bio = bufio.NewReader(io)
		cli.out = io
		cli.Context.Master(cli.Context)

		if msg := m.Find("aaa", true); msg != nil {
			cli.echo("username>")
			username, e := cli.bio.ReadString('\n')
			msg.Assert(e)
			username = strings.TrimSpace(username)

			cli.echo("password>")
			password, e := cli.bio.ReadString('\n')
			msg.Assert(e)
			password = strings.TrimSpace(password)

			msg.Name = "aaa"
			msg.Wait = make(chan bool)
			if msg.Cmd("login", username, password) == "" {
				cli.echo("登录失败")
				io.Close()
				return true
			}

			if cli.Sessions == nil {
				cli.Sessions = make(map[string]*ctx.Message)
			}
			cli.Sessions["aaa"] = msg
		}
	} else if stream, ok := m.Data["bio"]; ok {
		cli.Context.Exit = make(chan bool)
		cli.bio = stream.(*bufio.Reader)
		m.AssertOne(m, true, func(m *ctx.Message) {
			for cli.parse(m) {
			}
		})
		m.Log("fuck", nil, "done 1")
		cli.Context.Exit <- true
		m.Log("fuck", nil, "done 2")
		return true

	} else if stream, ok := m.Data["file"]; ok {
		cli.Context.Exit = make(chan bool)
		io := stream.(io.ReadWriteCloser)
		cli.bio = bufio.NewReader(io)
		m.AssertOne(m, true, func(m *ctx.Message) {
			for cli.parse(m) {
			}
		})
		m.Log("fuck", nil, "done 1")
		cli.Context.Exit <- true
		m.Log("fuck", nil, "done 2")
		return true
	} else if len(arg) > 0 {
		cli.Context.Master(cli.Context)
		cli.bio = bufio.NewReader(os.Stdin)
		cli.out = os.Stdout
		m.Cap("stream", "stdout")
	}

	if cli.bio != nil {
		go m.AssertOne(m, true, func(m *ctx.Message) {
			m.Cap("next", "source "+m.Conf("init.sh"))
			for cli.parse(m) {
			}
		})
	}

	m.Capi("nterm", 1)
	defer m.Capi("nterm", -1)

	m.Deal(nil, func(msg *ctx.Message, arg ...string) bool {
		if msg.Get("result") == "error: " {
			if msg.Get("detail") != "login" {
				msg.Log("system", nil, "%v", msg.Meta["detail"])

				msg.Set("result")
				msg.Set("append")
				c := exec.Command(msg.Meta["detail"][0], msg.Meta["detail"][1:]...)

				if cli.out == os.Stdout {
					c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
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
		}

		cli.target = msg.Target()
		return cli.exit == false
	})

	return true
}

func (cli *CLI) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch cli.Context {
	case m.Target():
		if cli.Context == Index {
			return false
		}
	case m.Source():
		if m.Name == "aaa" {
			msg := m.Spawn(cli.Context)
			msg.Master(cli.Context)
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
			cli := m.Master().Master().Server.(*CLI)
			f, e := os.Open(arg[0])
			m.Assert(e)
			m.Put("option", "file", f).Start(arg[0], "脚本文件", key)
			m.Log("fuck", nil, "source")
			cli.wait = m.Target()
		}},
		"return": &ctx.Command{Name: "return", Help: "运行脚本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			cli := m.Master().Master().Server.(*CLI)
			cli.exit = true
		}},
		"sleep": &ctx.Command{Name: "sleep time", Help: "运行脚本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			t, e := strconv.Atoi(arg[0])
			m.Assert(e)
			m.Log("info", nil, "sleep %ds", t)
			time.Sleep(time.Second * time.Duration(t))
			m.Log("info", nil, "sleep %ds done", t)
		}},
		"if": &ctx.Command{Name: "if a [ = | != ] b", Help: "运行脚本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			cli := m.Master().Master().Server.(*CLI)
			m.Put("option", "bio", cli.bio).Start(strings.Join(arg, " "), "条件语句", key)
			cli.wait = m.Target()
		}},
		"end": &ctx.Command{Name: "end", Help: "运行脚本", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			cli := m.Master().Master().Server.(*CLI)
			cli.exit = true
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
				action := "dial" // {{{
				if m.Has("listen") {
					action = "listen"
				}

				msg := m.Find(m.Get("args"), true)

				if m.Has("master") {
					msg.Template = msg.Spawn(msg.Source()).Add("option", "master")
				}
				msg.Cmd(action, m.Get(action))

			}}, // }}}
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
				m.Assert(m.Target() != c, "模块是主控模块")

				msg := m.Spawn(c)
				msg.Start(fmt.Sprintf("PTS%d", cli.Capi("nterm")), arg[0], arg[1:]...)
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
