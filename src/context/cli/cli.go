package cli // {{{
// }}}
import ( // {{{
	"context"

	"bufio"
	"io"

	"fmt"
	"strconv"
	"strings"

	"os"
	"os/exec"
	"time"

	"regexp"
)

// }}}

type CLI struct { // {{{
	out   io.WriteCloser
	bio   *bufio.Reader
	lines []string

	yac    *ctx.Message
	lex    *ctx.Message
	target *ctx.Context
	alias  map[string][]string

	*ctx.Context
}

// }}}

func (cli *CLI) print(str string, arg ...interface{}) bool { // {{{
	if cli.out != nil {
		fmt.Fprintf(cli.out, str, arg...)
		return true
	}
	return false
}

// }}}
func (cli *CLI) check(arg []string) bool {
	if len(arg) < 0 {
		return false
	}

	switch arg[0] {
	case "", "0", "false":
		return false
	}

	return true
}

func (cli *CLI) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}
	c.Caches["skip"] = &ctx.Cache{Name: "跳过执行", Value: cli.Pulse.Cap("skip"), Help: "命令只解析不执行"}

	s := new(CLI)
	s.Context = c
	s.lex = cli.lex
	s.yac = cli.yac
	if m.Has("for") {
		s.lines = append(s.lines, cli.lines[cli.Pulse.Capi("pos")-1:]...)
	} else {
		s.lines = append(s.lines, cli.lines[cli.Pulse.Capi("pos"):]...)
	}
	return s
}

// }}}
func (cli *CLI) Begin(m *ctx.Message, arg ...string) ctx.Server {
	cli.Caches["target"] = &ctx.Cache{Name: "操作目标", Value: cli.Name, Help: "命令操作的目标"}
	cli.Caches["result"] = &ctx.Cache{Name: "执行结果", Value: "", Help: "前一条命令的执行结果"}
	cli.Caches["back"] = &ctx.Cache{Name: "前一条指令", Value: "", Help: "前一条指令"}
	cli.Caches["next"] = &ctx.Cache{Name: "下一条指令", Value: "", Help: "下一条指令"}

	cli.Caches["nline"] = &ctx.Cache{Name: "缓存命令行数", Value: "0", Help: "缓存命令行数"}
	cli.Caches["pos"] = &ctx.Cache{Name: "当前缓存命令", Value: "0", Help: "当前缓存命令"}

	cli.Caches["else"] = &ctx.Cache{Name: "解析选择语句", Value: "false", Help: "解析选择语句"}
	cli.Caches["exit"] = &ctx.Cache{Name: "解析结束", Value: "false", Help: "解析结束模块销毁"}

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
		if len(arg) > 0 && len(arg[0]) > 0 {
			cli, ok := m.Target().Server.(*CLI)
			m.Assert(ok, "模块类型错误")

			yac := m.Find(arg[0], true)
			m.Assert(yac != nil, "词法解析模块不存在")
			if yac.Cap("status") != "start" {
				yac.Target().Start(yac)
				m.Spawn(yac.Target()).Cmd("train", "void", "void", "[\t ]+")

				m.Spawn(yac.Target()).Cmd("train", "key", "key", "[A-Za-z_][A-Za-z_0-9]*")
				m.Spawn(yac.Target()).Cmd("train", "num", "num", "mul{", "[1-9][0-9]*", "0[0-9]+", "0x[0-9]+", "}")
				m.Spawn(yac.Target()).Cmd("train", "str", "str", "mul{", "\"[^\"]*\"", "'[^']*'", "}")

				m.Spawn(yac.Target()).Cmd("train", "tran", "tran", "mul{", "@", "$", "}", "opt{", "[a-zA-Z0-9]+", "}")
				m.Spawn(yac.Target()).Cmd("train", "word", "word", "mul{", "~", "!", "tran", "\"[^\"]*\"", "'[^']*'", "[a-zA-Z0-9_/.]+", "}")

				m.Spawn(yac.Target()).Cmd("train", "op1", "op1", "opt{", "mul{", "-z", "-n", "}", "}", "word")
				m.Spawn(yac.Target()).Cmd("train", "op2", "op2", "op1", "rep{", "mul{", "+", "-", "*", "/", "}", "op1", "}")
				m.Spawn(yac.Target()).Cmd("train", "op1", "op1", "(", "op2", ")")

				m.Spawn(yac.Target()).Cmd("train", "stm", "var", "var", "key", "opt{", "=", "op2", "}")

				m.Spawn(yac.Target()).Cmd("train", "cmd", "cmd", "rep{", "word", "}")
				m.Spawn(yac.Target()).Cmd("train", "tran", "tran", "$", "(", "cmd", ")")

				m.Spawn(yac.Target()).Cmd("train", "line", "line", "opt{", "mul{", "stm", "cmd", "}", "}", "mul{", ";", "\n", "#[^\n]*\n", "}")
			}
			cli.yac = yac
			return arg[0]
		}
		return x.Value

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

	cli.target = cli.Context
	cli.alias = map[string][]string{
		"~": []string{"context"},
		"!": []string{"message"},
		"@": []string{"config"},
		"$": []string{"cache"},
	}

	if cli.Context == Index {
		Pulse = m
	}

	return cli
}

func (cli *CLI) Start(m *ctx.Message, arg ...string) bool { // {{{
	cli.Caches["#"] = &ctx.Cache{Name: "参数个数", Value: fmt.Sprintf("%d", len(arg)), Help: "参数个数"}
	for i, v := range arg {
		cli.Caches[fmt.Sprintf("%d", i)] = &ctx.Cache{Name: "执行参数", Value: v, Help: "执行参数"}
	}

	cli.Context.Exit = make(chan bool)
	if m.Caps("else", false); !m.Has("skip") {
		m.Caps("skip", false)
	} else if m.Capi("skip", 1) == 1 {
		m.Caps("else", true)
	}
	m.Caps("exit", false)

	if m.Has("stdio") {
		cli.Caches["init.shy"] = &ctx.Cache{Name: "启动脚本", Value: "etc/init.shy", Help: "模块启动时自动运行的脚本"}
		cli.Caches["level"] = &ctx.Cache{Name: "模块嵌套层数", Value: "0", Help: "模块嵌套层数"}
		if len(arg) > 0 {
			m.Cap("init.shy", arg[0])
		}
		m.Find("nfs").Cmd("scan", m.Cap("init.shy"))

		m.Cap("next", fmt.Sprintf("source %s\n", m.Cap("init.shy")))
		cli.bio = bufio.NewReader(os.Stdin)
		cli.out = os.Stdout
		m.Conf("yac", "yac")
		m.Cap("stream", "stdout")
	} else if stream, ok := m.Data["file"]; ok {
		if bio, ok := stream.(*bufio.Reader); ok {
			cli.bio = bio
			m.Cap("stream", "bufio")
		} else {
			cli.bio = bufio.NewReader(stream.(io.ReadWriteCloser))
			m.Cap("stream", "file")
		}
	}

	m.Capi("nline", 0, len(cli.lines))
	m.Caps("pos", m.Has("for"))

	m.Log("info", nil, "%p %s pos:%s nline:%s %d", cli.bio, m.Cap("stream"), m.Cap("pos"), m.Cap("nline"), len(cli.lines))

	go m.AssertOne(m, true, func(m *ctx.Message) {
		for !m.Caps("exit") {
			line := m.Cap("next")
			if m.Cap("next", ""); line == "" {
				if cli.bio == nil {
					line = cli.lines[m.Capi("pos", 1)-1]
				} else {
					cli.print(m.Conf("PS1"))
					if l, e := cli.bio.ReadString('\n'); m.Assert(e) {
						line = l
					}
				}
			}

			if line == "\n" && cli.out != nil {
				line = m.Cap("back") + "\n"
				m.Cap("back", "")
			}

			yac := m.Spawn(cli.yac.Target())
			yac.Cmd("parse", "line", "void", line)
		}
	}, func(m *ctx.Message) {
		m.Caps("exit", true)
		m.Spawn(cli.Context).Set("detail", "end").Post(cli.Context)
	})

	m.Deal(func(msg *ctx.Message, arg ...string) bool {
		return !cli.Has("skip") || !m.Caps("skip") || Index.Has(msg.Get("detail"), "command")

	}, func(msg *ctx.Message, arg ...string) bool {
		if cli.Has("skip") && m.Caps("skip") {
			return !m.Caps("exit")
		}

		if !msg.Hand && cli.Owner == ctx.Index.Owner {
			msg.Hand = true
			msg.Log("system", nil, "%v", msg.Meta["detail"])

			msg.Set("result").Set("append")
			c := exec.Command(msg.Meta["detail"][0], msg.Meta["detail"][1:]...)

			if cli.out == os.Stdout {
				c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
				if e := c.Start(); e != nil {
					msg.Echo("error: ")
					msg.Echo("%s\n", e)
				} else if e := c.Wait(); e != nil {
					msg.Echo("error: ")
					msg.Echo("%s\n", e)
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

		if msg.Target().Context() != nil || msg.Target() == ctx.Index {
			cli.target = msg.Target()
		}
		m.Cap("target", cli.target.Name)

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
		if p, ok := cli.Context.Context().Server.(*CLI); ok {
			if p.bio != nil && cli.Context != Index {
				if m.Has("for") {
					cli.lines = cli.lines[1:]
				}
				p.lines = append(p.lines, cli.lines...)
				p.Pulse.Capi("nline", 0, len(p.lines))
			}
			if p.Pulse.Capi("pos", cli.Pulse.Capi("pos")); m.Has("for") {
				p.Pulse.Capi("pos", -1)
			}

			m.Log("info", nil, "%p %s pos:%s nline:%s %d", cli.bio, m.Cap("stream"), m.Cap("pos"), m.Cap("nline"), len(cli.lines))
			m.Log("info", nil, "%p %s pos:%s nline:%s %d", p.bio, p.Pulse.Cap("stream"), p.Pulse.Cap("pos"), p.Pulse.Cap("nline"), len(p.lines))
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
		"skip": &ctx.Cache{Name: "跳过执行", Value: "0", Help: "命令只解析不执行"},
	},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"express": &ctx.Command{Name: "express exp", Help: "表达式运算", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			result := "false"
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
				case "==":
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
		}},
		"op1": &ctx.Command{Name: "op1 word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*CLI); m.Assert(ok) {
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
				}
			}
		}},
		"op2": &ctx.Command{Name: "op2 word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) {
				pre := map[string]int{"+": 1, "-": 1, "*": 2, "/": 2}
				num := []string{arg[0]}
				op := []string{}

				for i := 1; i < len(arg); i += 2 {
					if len(op) > 0 && pre[op[len(op)-1]] >= pre[arg[i]] {
						num[len(op)-1] = m.Spawn(cli.Context).Cmd("express", num[len(op)-1], op[len(op)-1], num[len(op)])
						num = num[:len(num)-1]
						op = op[:len(op)-1]
					}

					num = append(num, arg[i+1])
					op = append(op, arg[i])
				}

				for i := len(op) - 1; i >= 0; i-- {
					num[i] = m.Spawn(cli.Context).Cmd("express", num[i], op[i], num[i+1])
				}

				m.Echo("%s", num[0])
			}
		}},
		"tran": &ctx.Command{Name: "tran word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
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
						m.Echo(m.Cap(arg[1]))
					case "@":
						m.Echo(m.Conf(arg[1]))
					}
				}
			} // }}}
		}},
		"cmd": &ctx.Command{Name: "cmd word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{

				msg := m.Spawn(cli.target)
				if a, ok := cli.alias[arg[0]]; ok {
					msg.Set("detail", a...)
					msg.Meta["detail"] = append(msg.Meta["detail"], arg[1:]...)
				} else {
					msg.Set("detail", arg...)
				}

				msg.Post(cli.Context)
				if m.Hand = false; msg.Hand {
					m.Hand = true
					m.Meta["result"] = msg.Meta["result"]
				}
			} // }}}
		}},
		"line": &ctx.Command{Name: "line word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				arg = arg[:len(arg)-1]

				result := strings.TrimRight(strings.Join(arg, ""), "\n")
				if m.Cap("result", result); len(result) > 0 {
					cli.print(result + "\n")
				}

				if m.Cap("back", ""); cli.bio != nil {
					cli.lines = append(cli.lines, strings.Join(arg, " "))
					m.Capi("nline", 1)
					m.Capi("pos", 1)
				}
			} // }}}
		}},
		"alias": &ctx.Command{Name: "alias [short [long]]|[delete short]", Help: "查看、定义或删除命令别名, short: 命令别名, long: 命令原名, delete: 删除别名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) { // {{{
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
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) { // {{{
				if d, e := time.ParseDuration(arg[0]); m.Assert(e) {
					m.Log("info", nil, "sleep %v", d)
					time.Sleep(d)
					m.Log("info", nil, "sleep %v done", d)
				}
			} // }}}
		}},
		"var": &ctx.Command{Name: "var a [= exp]", Help: "定义变量, a: 变量名, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) { // {{{
				val := ""
				if len(arg) > 3 {
					val = m.Spawn(cli.Context).Cmd(append([]string{"express"}, arg[3:]...)...)
				}
				m.Cap(arg[1], arg[1], val, "临时变量")
			} // }}}
		}},
		"let": &ctx.Command{Name: "let a = exp", Help: "设置变量, a: 变量名, exp: 表达式(a {+|-|*|/|%} b)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) { // {{{
				m.Echo(cli.Pulse.Cap(arg[0], m.Spawn(cli.Context).Cmd(append([]string{"express"}, arg[2:]...)...)))
			} // }}}
		}},
		"source": &ctx.Command{Name: "source file", Help: "运行脚本, file: 脚本文件名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) { // {{{
				if f, e := os.Open(arg[0]); m.Assert(e) {
					m.Put("option", "file", f).Start(fmt.Sprintf("%s%d", key, Pulse.Capi("level", 1)), "脚本文件")
					<-m.Target().Exit
					Pulse.Capi("level", -1)
				}
			} // }}}
		}},
		"return": &ctx.Command{Name: "return result...", Help: "结束脚本, rusult: 返回值", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) { // {{{
				call := cli.Requests[len(cli.Requests)-1]
				call.Set("result", arg...)
				cli.Pulse.Caps("exit", true)
			} // }}}
		}},
		"if": &ctx.Command{Name: "if exp", Help: "条件语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) { // {{{
				if m.Target(m.Source()); (cli.Has("skip") && cli.Pulse.Caps("skip")) || !cli.check(arg) {
					m.Add("option", "skip")
				}

				m.Put("option", "file", cli.bio).Start(fmt.Sprintf("%s%d", key, Pulse.Capi("level", 1)), "条件语句")
				<-m.Target().Exit
				Pulse.Capi("level", -1)
			} // }}}
		}},
		"elif": &ctx.Command{Name: "elif exp", Help: "条件语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) { // {{{
				if cli.Pulse.Caps("skip", !cli.Pulse.Caps("else")) {
					return
				}
				cli.Pulse.Caps("else", cli.Pulse.Caps("skip", !cli.check(arg)))
			} // }}}
		}},
		"else": &ctx.Command{Name: "else", Help: "条件语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) { // {{{
				cli.Pulse.Caps("skip", !cli.Pulse.Caps("else"))
			} // }}}
		}},
		"end": &ctx.Command{Name: "end", Help: "结束语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) { // {{{
				if cli.Pulse.Caps("exit", true); cli.Pulse.Has("for") && !cli.Pulse.Caps("skip") {
					cli.Pulse.Caps("exit", false)
					cli.Pulse.Cap("pos", "0")

				}
				cli.bio = nil
			} // }}}
		}},
		"for": &ctx.Command{Name: "for exp", Help: "循环语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) { // {{{
				if cli.Pulse.Has("for") && cli.Pulse.Capi("pos") == 1 {
					cli.Pulse.Caps("skip", !cli.check(arg))
					return
				}
				if m.Target(m.Source()); (cli.Has("skip") && cli.Pulse.Caps("skip")) || !cli.check(arg) {
					m.Add("option", "skip")
				}
				m.Add("option", "for", cli.Pulse.Cap("back"))
				m.Put("option", "file", cli.bio).Start(fmt.Sprintf("%s%d", key, Pulse.Capi("level", 1)), "循环语句")
				<-m.Target().Exit
				Pulse.Capi("level", -1)
			} // }}}
		}},
		"function": &ctx.Command{Name: "function name", Help: "函数定义, name: 函数名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) { // {{{
				if _, ok := cli.Context.Context().Server.(*CLI); ok {
					m.Target(m.Source().Context())
				} else {
					m.Target(m.Source())
				}

				m.Add("option", "skip").Add("option", "save")
				m.Put("option", "file", cli.bio).Start(arg[0], "函数定义")
				<-m.Target().Exit
			} // }}}
		}},
		"call": &ctx.Command{Name: "call name arg...", Help: "函数调用, name: 函数名, arg: 参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Target(m.Source()) // {{{
			m.BackTrace(func(msg *ctx.Message) bool {
				if fun := msg.Find(arg[0], false); fun != nil {
					fun.Add("detail", arg[0], arg[1:]...).Target().Start(fun)
					<-fun.Target().Exit
					m.Set("result", fun.Meta["result"]...)
					return false
				}
				return true
			}) // }}}
		}},
	},
}

func init() {
	cli := &CLI{}
	cli.Context = Index
	ctx.Index.Register(Index, cli)
}
