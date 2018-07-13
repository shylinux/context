package cli // {{{
// }}}
import ( // {{{
	"contexts"

	"fmt"
	"strconv"
	"strings"

	"os"
	"os/exec"
	"time"

	"regexp"
)

// }}}

type CLI struct {
	nfs *ctx.Message
	lex *ctx.Message
	yac *ctx.Message

	target *ctx.Context
	alias  map[string][]string

	*ctx.Context
}

func (cli *CLI) check(arg string) bool { // {{{
	switch arg {
	case "", "0", "false":
		return false
	}

	return true
}

// }}}

func (cli *CLI) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(CLI)
	s.Context = c
	s.lex = cli.lex
	s.yac = cli.yac
	s.nfs = cli.nfs
	s.target = m.Source()
	if len(arg) > 0 {
		s.target = c
	}
	return s
}

// }}}
func (cli *CLI) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	cli.alias = map[string][]string{
		"~": []string{"context"},
		"!": []string{"message"},
		"@": []string{"config"},
		"$": []string{"cache"},
	}

	if len(arg) > 0 && arg[0] == "scan_file" {
		return cli
	}

	cli.Caches["level"] = &ctx.Cache{Name: "嵌套层级", Value: "0", Help: "嵌套层级"}
	cli.Caches["skip"] = &ctx.Cache{Name: "跳过执行", Value: "0", Help: "命令只解析不执行"}
	cli.Caches["else"] = &ctx.Cache{Name: "解析选择语句", Value: "false", Help: "解析选择语句"}
	cli.Caches["loop"] = &ctx.Cache{Name: "解析循环语句", Value: "-2", Help: "解析选择语句"}
	cli.Caches["fork"] = &ctx.Cache{Name: "解析结束", Value: "-2", Help: "解析结束模块销毁"}
	cli.Caches["exit"] = &ctx.Cache{Name: "解析结束", Value: "false", Help: "解析结束模块销毁"}

	cli.Caches["target"] = &ctx.Cache{Name: "操作目标", Value: cli.Name, Help: "命令操作的目标"}
	cli.Caches["result"] = &ctx.Cache{Name: "执行结果", Value: "", Help: "前一条命令的执行结果"}
	cli.Caches["last"] = &ctx.Cache{Name: "前一条消息", Value: "0", Help: "前一条命令的编号"}
	cli.Caches["back"] = &ctx.Cache{Name: "前一条指令", Value: "", Help: "前一条指令"}
	cli.Caches["next"] = &ctx.Cache{Name: "下一条指令", Value: "", Help: "下一条指令"}

	cli.Configs["target"] = &ctx.Config{Name: "词法解析器", Value: "cli", Help: "命令行词法解析器", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
		if len(arg) > 0 && len(arg[0]) > 0 { // {{{
			cli, ok := m.Target().Server.(*CLI)
			m.Assert(ok, "模块类型错误")

			target := m.Find(arg[0])
			cli.target = target.Target()
			return arg[0]
		}
		return x.Value
		// }}}
	}}
	cli.Configs["lex"] = &ctx.Config{Name: "词法解析器", Value: "", Help: "命令行词法解析器", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
		if len(arg) > 0 && len(arg[0]) > 0 { // {{{
			cli, ok := m.Target().Server.(*CLI)
			m.Assert(ok, "模块类型错误")

			lex := m.Find(arg[0], true)
			m.Assert(lex != nil, "词法解析模块不存在")
			if lex.Cap("status") != "start" {
				lex.Target().Start(lex)
				m.Spawn(lex.Target()).Cmd("train", "'[^']*'", "word", "word")
				m.Spawn(lex.Target()).Cmd("train", "\"[^\"]*\"", "word", "word")
				m.Spawn(lex.Target()).Cmd("train", "[^\t \n]+", "word", "word")
				m.Spawn(lex.Target()).Cmd("train", "[\t \n]+", "void", "void")
				m.Spawn(lex.Target()).Cmd("train", "#[^\n]*\n", "help", "void")
			}
			cli.lex = lex
			return arg[0]
		}
		return x.Value
		// }}}
	}}
	cli.Configs["yac"] = &ctx.Config{Name: "词法解析器", Value: "", Help: "命令行词法解析器", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
		if len(arg) > 0 && len(arg[0]) > 0 { // {{{
			cli, ok := m.Target().Server.(*CLI)
			m.Assert(ok, "模块类型错误")

			yac := m.Find(arg[0], true)
			m.Assert(yac != nil, "词法解析模块不存在")
			if yac.Cap("status") != "start" {
				yac.Target().Start(yac)
				yac.Cmd("train", "void", "void", "[\t ]+")

				yac.Cmd("train", "key", "key", "[A-Za-z_][A-Za-z_0-9]*")
				yac.Cmd("train", "num", "num", "mul{", "0", "-?[1-9][0-9]*", "0[0-9]+", "0x[0-9]+", "}")
				yac.Cmd("train", "str", "str", "mul{", "\"[^\"]*\"", "'[^']*'", "}")

				yac.Cmd("train", "tran", "tran", "mul{", "@", "$", "}", "opt{", "[a-zA-Z0-9_]+", "}")
				yac.Cmd("train", "word", "word", "mul{", "~", "!", "=", "tran", "str", "[a-zA-Z0-9_/\\-.:]+", "}")

				yac.Cmd("train", "op1", "op1", "mul{", "$", "@", "}")
				yac.Cmd("train", "op1", "op1", "mul{", "-z", "-n", "}")
				yac.Cmd("train", "op1", "op1", "mul{", "-e", "-f", "-d", "}")
				yac.Cmd("train", "op1", "op1", "mul{", "-", "+", "}")
				yac.Cmd("train", "op2", "op2", "mul{", "+", "-", "*", "/", "}")
				yac.Cmd("train", "op2", "op2", "mul{", ">", ">=", "<", "<=", "=", "!=", "}")

				yac.Cmd("train", "val", "val", "opt{", "op1", "}", "mul{", "num", "key", "str", "tran", "}")
				yac.Cmd("train", "exp", "exp", "val", "rep{", "op2", "val", "}")
				yac.Cmd("train", "val", "val", "(", "exp", ")")

				yac.Cmd("train", "stm", "var", "var", "key", "opt{", "=", "exp", "}")
				yac.Cmd("train", "stm", "let", "let", "key", "mul{", "=", "<-", "}", "exp")
				yac.Cmd("train", "stm", "if", "if", "exp")
				yac.Cmd("train", "stm", "elif", "elif", "exp")
				yac.Cmd("train", "stm", "for", "for", "exp")
				yac.Cmd("train", "stm", "else", "else")
				yac.Cmd("train", "stm", "end", "end")
				yac.Cmd("train", "stm", "function", "function", "rep{", "key", "}")
				yac.Cmd("train", "stm", "return", "return", "rep{", "exp", "}")

				yac.Cmd("train", "cmd", "echo", "rep{", "exp", "}")
				yac.Cmd("train", "cmd", "cmd", "cache", "rep{", "word", "}")
				yac.Cmd("train", "cmd", "cmd", "cache", "key", "rep{", "word", "}")
				yac.Cmd("train", "cmd", "cmd", "cache", "key", "opt{", "=", "exp", "}")
				yac.Cmd("train", "cmd", "cmd", "rep{", "word", "}")
				yac.Cmd("train", "tran", "tran", "$", "(", "cmd", ")")

				yac.Cmd("train", "line", "line", "opt{", "mul{", "stm", "cmd", "}", "}", "mul{", ";", "\n", "#[^\n]*\n", "}")
			}
			cli.yac = yac
			return arg[0]
		}
		return x.Value
		// }}}
	}}
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

			// ps = append(ps, "\033[32m")
			ps = append(ps, cli.target.Name)
			ps = append(ps, "> ")
			// ps = append(ps, "\033[0m> ")

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

	if cli.Context == Index {
		Pulse = m
	}

	return cli
}

// }}}
func (cli *CLI) Start(m *ctx.Message, arg ...string) bool { // {{{
	if len(arg) > 0 && arg[0] == "scan_file" {
		yac := m.Find("yac")
		if yac.Cap("status") != "start" {
			yac.Target().Start(yac)
			yac.Cmd("train", "void", "void", "[\t ]+")

			yac.Cmd("train", "key", "key", "[A-Za-z_][A-Za-z_0-9]*")
			yac.Cmd("train", "num", "num", "mul{", "0", "-?[1-9][0-9]*", "0[0-9]+", "0x[0-9]+", "}")
			yac.Cmd("train", "str", "str", "mul{", "\"[^\"]*\"", "'[^']*'", "}")

			yac.Cmd("train", "tran", "tran", "mul{", "@", "$", "}", "opt{", "[a-zA-Z0-9_]+", "}")
			yac.Cmd("train", "word", "word", "mul{", "~", "!", "=", "tran", "str", "[a-zA-Z0-9_/\\-.:]+", "}")

			yac.Cmd("train", "op1", "op1", "mul{", "$", "@", "}")
			yac.Cmd("train", "op1", "op1", "mul{", "-z", "-n", "}")
			yac.Cmd("train", "op1", "op1", "mul{", "-e", "-f", "-d", "}")
			yac.Cmd("train", "op1", "op1", "mul{", "-", "+", "}")
			yac.Cmd("train", "op2", "op2", "mul{", "+", "-", "*", "/", "}")
			yac.Cmd("train", "op2", "op2", "mul{", ">", ">=", "<", "<=", "=", "!=", "}")

			yac.Cmd("train", "val", "val", "opt{", "op1", "}", "mul{", "num", "key", "str", "tran", "}")
			yac.Cmd("train", "exp", "exp", "val", "rep{", "op2", "val", "}")
			yac.Cmd("train", "val", "val", "(", "exp", ")")

			yac.Cmd("train", "stm", "var", "var", "key", "opt{", "=", "exp", "}")
			yac.Cmd("train", "stm", "let", "let", "key", "mul{", "=", "<-", "}", "exp")
			yac.Cmd("train", "stm", "if", "if", "exp")
			yac.Cmd("train", "stm", "elif", "elif", "exp")
			yac.Cmd("train", "stm", "for", "for", "exp")
			yac.Cmd("train", "stm", "else", "else")
			yac.Cmd("train", "stm", "end", "end")
			yac.Cmd("train", "stm", "function", "function", "rep{", "key", "}")
			yac.Cmd("train", "stm", "return", "return", "rep{", "exp", "}")

			yac.Cmd("train", "cmd", "echo", "rep{", "exp", "}")
			yac.Cmd("train", "cmd", "cmd", "cache", "rep{", "word", "}")
			yac.Cmd("train", "cmd", "cmd", "cache", "key", "rep{", "word", "}")
			yac.Cmd("train", "cmd", "cmd", "cache", "key", "opt{", "=", "exp", "}")
			yac.Cmd("train", "cmd", "cmd", "rep{", "word", "}")
			yac.Cmd("train", "tran", "tran", "$", "(", "cmd", ")")

			yac.Cmd("train", "line", "line", "opt{", "mul{", "stm", "cmd", "}", "}", "mul{", ";", "\n", "#[^\n]*\n", "}")
		}

		m.Option("target", m.Target().Name)
		yac = m.Find("yac")
		yac.Call(func(cmd *ctx.Message) *ctx.Message {
			if cmd.Detail(0) == "scan_end" {
				m.Target().Close(m.Spawn())
				return nil
			}

			cmd.Cmd()
			m.Option("target", cli.target.Name)
			if cmd.Has("return") {
				m.Target().Close(m.Spawn())
				cmd.Append("scan_file", false)
			}
			return nil
		}, "parse", arg[1])
		m.Cap("stream", yac.Target().Name)
		return false
	}

	m.Sesss("cli", m)
	cli.Caches["#"] = &ctx.Cache{Name: "参数个数", Value: fmt.Sprintf("%d", len(arg)), Help: "参数个数"}
	for i, v := range arg {
		cli.Caches[fmt.Sprintf("%d", i)] = &ctx.Cache{Name: "执行参数", Value: v, Help: "执行参数"}
	}
	cli.Caches["current_cli"] = &ctx.Cache{Name: "当前命令中心", Value: fmt.Sprintf("%d", m.Code()), Help: "当前命令中心"}

	if m.Has("level") {
		m.Cap("level", m.Option("level"))
	}
	if m.Has("skip") {
		m.Cap("skip", m.Option("skip"))
		if m.Caps("else", false); m.Capi("skip") == 1 {
			m.Caps("else", true)
		}
	}
	if m.Has("loop") {
		m.Cap("loop", m.Option("loop"))
	}
	if m.Has("fork") {
		m.Cap("fork", m.Option("fork"))
	}

	m.Caps("exit", false)
	cli.Context.Exit = make(chan bool)
	cli.Context.Master(cli.Context)

	if m.Has("stdio") || len(arg) > 0 {
		go func() {
			cli.Caches["init.shy"] = &ctx.Cache{Name: "启动脚本", Value: "etc/init.shy", Help: "模块启动时自动运行的脚本"}
			if m.Conf("yac", "yac"); len(arg) > 0 {
				m.Cap("init.shy", arg[0])
			}

			cli.nfs = m.Sesss("nfs", "nfs")
			if m.Has("stdio") {
				m.Spawn().Cmd("scan_file", m.Cap("stream", m.Cap("init.shy")))
				cli.Context.Exit = make(chan bool)
				m.Find("yac").Call(func(cmd *ctx.Message) *ctx.Message {
					if cmd.Detail(0) == "scan_end" {
						msg := m.Spawn()
						cli.Exit <- true
						m.Target().Close(msg)
						return nil
					}

					cmd.Source(m.Target())
					cmd.Target(m.Target())
					cmd.Cmd()
					return nil
				}, "scan_file", m.Cap("stream", "stdio"))
			} else {
				if _, e := os.Stat(m.Cap("init.shy")); e == nil {
					// m.Spawn().Cmd("scan_file", m.Cap("stream", m.Cap("init.shy")))
					cli.nfs.Cmd("scan", m.Cap("stream", m.Cap("init.shy")))
				}
			}
		}()
		return false
	}

	m.Deal(func(msg *ctx.Message, arg ...string) bool {
		return !m.Caps("skip") || Index.Has(msg.Get("detail"), "command")

	}, func(msg *ctx.Message, arg ...string) bool {
		if m.Caps("skip") {
			return !m.Caps("exit")
		}

		return !m.Caps("exit")
	})

	if cli.Pulse.Has("save") {
		m.Cap("status", "stop")
		cli.Exit <- true
	}

	return !cli.Pulse.Has("save")
}

// }}}
func (cli *CLI) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch cli.Context {
	case m.Target():
		if _, ok := m.Source().Server.(*CLI); ok {
			// p.target = cli.target
		}
	case m.Source():
		if m.Name == "aaa" {
			if !cli.Context.Close(m.Spawn(cli.Context), arg...) {
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
		"time": &ctx.Cache{Name: "time", Value: "0", Help: "所有模块的当前目录", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
			t := time.Now().Unix()
			return fmt.Sprintf("%d", t)
		}},
		"nshell": &ctx.Cache{Name: "nshell", Value: "0", Help: "模块数量"},
	},
	Configs: map[string]*ctx.Config{
		"time_format":   &ctx.Config{Name: "time_format", Value: "2006-01-02 15:04:05", Help: "时间格式"},
		"time_interval": &ctx.Config{Name: "time_interval(open/close)", Value: "open", Help: "时间区间"},
		"cli_name": &ctx.Config{Name: "cli_name", Value: "shell", Help: "时间格式", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				return arg[0]
			}
			return fmt.Sprintf("%s%d", x.Value, m.Capi("nshell", 1))
			// }}}
		}},
		"cli_help": &ctx.Config{Name: "cli_help", Value: "shell", Help: "时间区间"},
	},
	Commands: map[string]*ctx.Command{
		"alias": &ctx.Command{Name: "alias [short [long]]|[delete short]", Help: "查看、定义或删除命令别名, short: 命令别名, long: 命令原名, delete: 删除别名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) && !m.Caps("skip") { // {{{
				switch len(arg) {
				case 0:
					for k, v := range cli.alias {
						m.Echo("%s: %v\n", k, v)
					}
				case 1:
					m.Echo("%s: %v\n", arg[0], cli.alias[arg[0]])
				default:
					if arg[0] == "delete" {
						m.Echo("delete: %s %v\n", arg[1], cli.alias[arg[1]])
						delete(cli.alias, arg[1])
					} else {
						cli.alias[arg[0]] = arg[1:]
						m.Echo("%s: %v\n", arg[0], cli.alias[arg[0]])
					}
				}
			} // }}}
		}},
		"sleep": &ctx.Command{Name: "sleep time", Help: "睡眠, time(ns/us/ms/s/m/h): 时间值(纳秒/微秒/毫秒/秒/分钟/小时)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Source().Server.(*CLI); m.Assert(ok) && !m.Caps("skip") { // {{{
				if d, e := time.ParseDuration(arg[0]); m.Assert(e) {
					m.Log("info", nil, "sleep %v", d)
					time.Sleep(d)
					m.Log("info", nil, "sleep %v done", d)
				}
			} // }}}
		}},
		"time": &ctx.Command{Name: "time [time_format format] [parse when] when [begin|end|yestoday|tommorow|monday|sunday|first|last|origin|last]",
			Form: map[string]int{"parse": 1, "time_format": 1, "time_interval": 1},
			Help: "睡眠, time(ns/us/ms/s/m/h): 时间值(纳秒/微秒/毫秒/秒/分钟/小时)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				t := time.Now() // {{{
				f := m.Confx("time_format")

				if m.Options("parse") {
					n, e := time.ParseInLocation(f, m.Option("parse"), time.Local)
					m.Assert(e)
					t = n
				}

				if len(arg) > 0 {
					if i, e := strconv.Atoi(arg[0]); e == nil {
						t = time.Unix(int64(i/1000), 0)
						arg = arg[1:]
					}
				}

				if len(arg) > 0 {
					switch arg[0] {
					case "begin":
						// t.Add(-((time.Second)t.Second() + (time.Minute)t.Minute() + (time.Hour)t.Hour()))
						d, e := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", t.Hour(), t.Minute(), t.Second()))
						m.Assert(e)
						t = t.Add(-d)
					case "end":
						d, e := time.ParseDuration(fmt.Sprintf("%dh%dm%ds%dns", t.Hour(), t.Minute(), t.Second(), t.Nanosecond()))
						m.Assert(e)
						t = t.Add(time.Duration(24*time.Hour) - d)
						if m.Confx("time_interval") == "close" {
							t = t.Add(-time.Second)
						}
					case "yestoday":
						t = t.Add(-time.Duration(24 * time.Hour))
					case "tomorrow":
						t = t.Add(time.Duration(24 * time.Hour))
					case "monday":
						d, e := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", int((t.Weekday()-time.Monday+7)%7)*24+t.Hour(), t.Minute(), t.Second()))
						m.Assert(e)
						t = t.Add(-d)
					case "sunday":
						d, e := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", int((t.Weekday()-time.Monday+7)%7)*24+t.Hour(), t.Minute(), t.Second()))
						m.Assert(e)
						t = t.Add(time.Duration(7*24*time.Hour) - d)
						if m.Confx("time_interval") == "close" {
							t = t.Add(-time.Second)
						}
					case "first":
						t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
					case "last":
						month, year := t.Month()+1, t.Year()
						if month >= 13 {
							month, year = 1, year+1
						}
						t = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
						if m.Confx("time_interval") == "close" {
							t = t.Add(-time.Second)
						}
					case "origin":
						t = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.Local)
					case "final":
						t = time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, time.Local)
						if m.Confx("time_interval") == "close" {
							t = t.Add(-time.Second)
						}
					}
				}

				if m.Options("parse") || !m.Options("time_format") {
					m.Echo("%d000", t.Unix())
				} else {
					m.Echo(t.Format(f))
				}
				// }}}
			}},
		"express": &ctx.Command{Name: "express exp", Help: "表达式运算", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			result := "false" // {{{
			switch len(arg) {
			case 0:
				result = ""
			case 1:
				result = arg[0]
			case 2:
				switch arg[0] {
				case "-z":
					if arg[1] == "" {
						result = "true"
					}
				case "-n":
					if arg[1] != "" {
						result = "true"
					}

				case "-e":
					if _, e := os.Stat(arg[1]); e == nil {
						result = "true"
					}
				case "-f":
					if info, e := os.Stat(arg[1]); e == nil && !info.IsDir() {
						result = "true"
					}
				case "-d":
					if info, e := os.Stat(arg[1]); e == nil && info.IsDir() {
						result = "true"
					}
				}
			case 3:
				v1, e1 := strconv.Atoi(arg[0])
				v2, e2 := strconv.Atoi(arg[2])
				switch arg[1] {
				case "+":
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%d", v1+v2)
					} else {
						result = arg[0] + arg[2]
					}
				case "-":
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%d", v1-v2)
					} else {
						result = strings.Replace(arg[0], arg[1], "", -1)
					}
				case "*":
					result = arg[0]
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%d", v1*v2)
					}
				case "/":
					result = arg[0]
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%d", v1/v2)
					}
				case "%":
					result = arg[0]
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%d", v1%v2)
					}

				case "<":
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%t", v1 < v2)
					} else {
						result = fmt.Sprintf("%t", arg[0] < arg[2])
					}
				case "<=":
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%t", v1 <= v2)
					} else {
						result = fmt.Sprintf("%t", arg[0] <= arg[2])
					}
				case ">":
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%t", v1 > v2)
					} else {
						result = fmt.Sprintf("%t", arg[0] > arg[2])
					}
				case ">=":
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%t", v1 >= v2)
					} else {
						result = fmt.Sprintf("%t", arg[0] >= arg[2])
					}
				case "=":
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%t", v1 == v2)
					} else {
						result = fmt.Sprintf("%t", arg[0] == arg[2])
					}
				case "!=":
					if e1 == nil && e2 == nil {
						result = fmt.Sprintf("%t", v1 != v2)
					} else {
						result = fmt.Sprintf("%t", arg[0] != arg[2])
					}

				case "~":
					if m, e := regexp.MatchString(arg[2], arg[0]); m && e == nil {
						result = "true"
					}
				case "!~":
					if m, e := regexp.MatchString(arg[2], arg[0]); !m || e != nil {
						result = "true"
					}
				}
			}
			m.Echo(result)
			// }}}
		}},
		"str": &ctx.Command{Name: "str word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*CLI); m.Assert(ok) && !m.Caps("skip") { // {{{
				str := arg[0][1 : len(arg[0])-1]
				str = strings.Replace(str, "\\n", "\n", -1)
				str = strings.Replace(str, "\\t", "\t", -1)
				m.Echo(str)
			} else {
				m.Set("result", arg...)
			}
			// }}}
		}},
		"val": &ctx.Command{Name: "val word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*CLI); m.Assert(ok) && !m.Caps("skip") { // {{{
				if arg[0] == "(" {
					m.Echo(arg[1])
					return
				}

				if len(arg) == 1 {
					m.Echo(arg[0])
					return
				}

				switch arg[0] {
				case "-z":
					if arg[1] == "" {
						m.Echo("true")
					} else {
						m.Echo("false")
					}
				case "-n":
					if arg[1] == "" {
						m.Echo("false")
					} else {
						m.Echo("true")
					}
				case "-e":
					if _, e := os.Stat(arg[1]); e == nil {
						m.Echo("true")
					} else {
						m.Echo("false")
					}
				case "-f":
					if info, e := os.Stat(arg[1]); e == nil && !info.IsDir() {
						m.Echo("true")
					} else {
						m.Echo("false")
					}
				case "-d":
					if info, e := os.Stat(arg[1]); e == nil && info.IsDir() {
						m.Echo("true")
					} else {
						m.Echo("false")
					}
				case "$":
					m.Echo(m.Cap(arg[1]))
				case "@":
					m.Echo(m.Conf(arg[1]))
				}
			} else {
				m.Set("result", arg...)
			}
			// }}}
		}},
		"exp": &ctx.Command{Name: "exp word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*CLI); m.Assert(ok) && !m.Caps("skip") { // {{{
				pre := map[string]int{"+": 1, "-": 1, "*": 2, "/": 2}
				num := []string{arg[0]}
				op := []string{}

				for i := 1; i < len(arg); i += 2 {
					if len(op) > 0 && pre[op[len(op)-1]] >= pre[arg[i]] {
						num[len(op)-1] = m.Cmd("express", num[len(op)-1], op[len(op)-1], num[len(op)]).Get("result")
						num = num[:len(num)-1]
						op = op[:len(op)-1]
					}

					num = append(num, arg[i+1])
					op = append(op, arg[i])
				}

				for i := len(op) - 1; i >= 0; i-- {
					num[i] = m.Spawn(m.Target()).Cmd("express", num[i], op[i], num[i+1]).Get("result")
				}

				m.Echo("%s", num[0])
			} else {
				m.Set("result", arg...)
			}
			// }}}
		}},
		"tran": &ctx.Command{Name: "tran word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) && !m.Caps("skip") { // {{{
				msg := m.Spawn(cli.target)
				switch len(arg) {
				case 1:
					switch arg[0] {
					case "$":
						m.Echo("cache")
					case "@":
						m.Echo("config")
					}
				case 2:
					switch arg[0] {
					case "$":
						m.Echo(msg.Cap(arg[1]))
					case "@":
						m.Echo(msg.Conf(arg[1]))
					}
				default:
					last := len(arg) - 1
					switch arg[0] {
					case "$":
						m.Result(0, arg[2:last])
					case "@":
						m.Result(0, arg[2:last])
					}
				}
			} else {
				m.Set("result", arg...)
			} // }}}
		}},
		"cmd": &ctx.Command{Name: "cmd word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) && !m.Caps("skip") { // {{{
				detail := []string{}

				if a, ok := cli.alias[arg[0]]; ok {
					detail = append(detail, a...)
					detail = append(detail, arg[1:]...)
				} else {
					detail = append(detail, arg...)
				}

				msg := m.Spawn(cli.target)
				msg.Cmd(detail)
				if msg.Target().Context() != nil || msg.Target() == ctx.Index {
					cli.target = msg.Target()
				}

				if !msg.Hand && cli.Owner == ctx.Index.Owner {
					msg.Hand = true
					msg.Log("system", nil, "%v", msg.Meta["detail"])

					msg.Set("result").Set("append")
					c := exec.Command(msg.Meta["detail"][0], msg.Meta["detail"][1:]...)

					if false {
						c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
						if e := c.Start(); e != nil {
							msg.Echo("error: ")
							msg.Echo("%s\n", e)
						} else if e := c.Wait(); e != nil {
							msg.Echo("error: ")
							m.Echo("%s\n", e)
						}
					} else {
						if out, e := c.CombinedOutput(); e != nil {
							msg.Echo("error: ")
							msg.Echo("%s\n", e)
						} else {
							msg.Echo(string(out))
						}
					}
				}

				m.Cap("target", cli.target.Name)
				m.Set("result", msg.Meta["result"]...)
				m.Capi("last", 0, msg.Code())

			} else {
				m.Set("result", arg...)
			}
			// }}}
		}},
		"var": &ctx.Command{Name: "var a [= exp]", Help: "定义变量, a: 变量名, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*CLI); m.Assert(ok) && !m.Caps("skip") { // {{{
				if m.Cap(arg[1], arg[1], "", "临时变量"); len(arg) > 3 {
					switch arg[2] {
					case "=":
						m.Cap(arg[1], arg[3])
					case "<-":
						m.Cap(arg[1], m.Cap("last"))
					}
				}
			} else {
				m.Set("result", arg...)
			} // }}}
		}},
		"let": &ctx.Command{Name: "let a = exp", Help: "设置变量, a: 变量名, exp: 表达式(a {+|-|*|/|%} b)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*CLI); m.Assert(ok) && !m.Caps("skip") { // {{{
				switch arg[2] {
				case "=":
					m.Cap(arg[1], arg[3])
				case "<-":
					m.Cap(arg[1], m.Cap("last"))
				}
			} else {
				m.Set("result", arg...)
			} // }}}
		}},
		"source": &ctx.Command{Name: "source file", Help: "运行脚本, file: 脚本文件名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			cli, ok := m.Source().Server.(*CLI) // {{{
			if !ok {
				cli, ok = m.Target().Server.(*CLI)
			}

			if !m.Caps("skip") {
				msg := m.Spawn(cli)
				msg.Start(fmt.Sprintf("%s_%d_%s", key, msg.Optioni("level", msg.Capi("level")+1), arg[0]), "脚本文件", arg[0])
				// <-msg.Target().Exit
				// m.Copy(msg, "result").Copy(msg, "append")
				// nfs := msg.Sesss("nfs")
				// nfs.Target().Close(nfs)
				//
				// sub, _ := msg.Target().Server.(*CLI)
				// if sub.target != msg.Target() {
				// 	cli.target = sub.target
				// }
			} // }}}
		}},
		"return": &ctx.Command{Name: "return result...", Help: "结束脚本, rusult: 返回值", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*CLI); m.Assert(ok) && !m.Caps("skip") { // {{{
				m.Add("append", "return", arg[1:]...)
			} // }}}
		}},
		"if": &ctx.Command{Name: "if exp", Help: "条件语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				if m.Optioni("skip", 0); m.Caps("skip") || !cli.check(arg[1]) {
					m.Optioni("skip", m.Capi("skip")+1)
				}

				m.Start(fmt.Sprintf("%s%d", key, m.Optioni("level", m.Capi("level")+1)), "条件语句")
			} // }}}
		}},
		"elif": &ctx.Command{Name: "elif exp", Help: "条件语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				if !m.Caps("else") {
					m.Caps("skip", true)
					return
				}

				if m.Caps("skip") {
					cli.nfs.Capi("pos", -1)
					m.Caps("skip", false)
					return
				}

				m.Caps("else", m.Caps("skip", !cli.check(arg[1])))
			} // }}}
		}},
		"else": &ctx.Command{Name: "else", Help: "条件语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				m.Caps("skip", !m.Caps("else"))
			} // }}}
		}},
		"end": &ctx.Command{Name: "end", Help: "结束语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				if m.Capi("fork") != -2 {
					m.Spawn(cli.nfs.Target()).Cmd("copy", cli.Name, m.Cap("fork"), m.Option("pos"))
				}

				if m.Caps("exit", true); !m.Caps("skip") && m.Capi("loop") >= 0 {
					m.Append("back", m.Cap("loop"))
					m.Caps("exit", false)
				} else {
					m.Put("append", "cli", cli.Context.Context())
				}
			} // }}}
		}},
		"for": &ctx.Command{Name: "for exp", Help: "循环语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				if m.Capi("loop") != -2 && m.Capi("loop") == m.Optioni("pos")-1 {
					m.Caps("skip", !cli.check(arg[1]))
					return
				}

				if m.Optioni("skip", 0); m.Caps("skip") || !cli.check(arg[1]) {
					m.Optioni("skip", m.Capi("skip")+1)
				}
				m.Optioni("loop", m.Optioni("pos")-1)
				m.Start(fmt.Sprintf("%s%d", key, m.Optioni("level", m.Capi("level")+1)), "循环语句")
			} // }}}
		}},
		"function": &ctx.Command{Name: "function name", Help: "函数定义, name: 函数名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				m.Optioni("fork", m.Optioni("pos")+1)
				m.Optioni("skip", m.Capi("skip")+1)
				m.Start(arg[1], "循环语句")
			} // }}}
		}},
		"call": &ctx.Command{Name: "call name arg...", Help: "函数调用, name: 函数名, arg: 参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			fun := m.Find("nfs.file1." + arg[0]) // {{{
			fun.Target().Start(fun)              // }}}
		}},
		"target": &ctx.Command{Name: "taget", Help: "函数调用, name: 函数名, arg: 参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) {
				m.Put("append", "target", cli.target)
			}
		}},
		"seed": &ctx.Command{Name: "seed", Help: "函数调用, name: 函数名, arg: 参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			for i := 0; i < 100; i++ {
				m.Echo("%d\n", i)
			}
		}},
		"echo": &ctx.Command{Name: "echo arg...", Help: "函数调用, name: 函数名, arg: 参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Echo("%s", strings.Join(arg, ""))
		}},
		"scan_file": &ctx.Command{
			Name: "scan_file filename [cli_name [cli_help]]",
			Help: "解析脚本, filename: 文件名, cli_name: 模块名, cli_help: 模块帮助",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				m.Start(m.Confx("cli_name", arg, 1), m.Confx("cli_help", arg, 2), key, arg[0])
				<-m.Target().Exit
			}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void",
			Caches: map[string]*ctx.Cache{
				"nserver": &ctx.Cache{},
			},
			Configs: map[string]*ctx.Config{
				"bench.log": &ctx.Config{},
			},
			Commands: map[string]*ctx.Command{
				"cmd": &ctx.Command{},
			},
		},
	},
}

func init() {
	cli := &CLI{}
	cli.Context = Index
	ctx.Index.Register(Index, cli)

	cli.target = Index
}
