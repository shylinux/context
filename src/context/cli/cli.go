package cli // {{{
// }}}
import ( // {{{
	"context"

	"bufio"
	"io"

	"fmt"
	"strconv"
	"strings"
	"unicode"

	"os"
	"os/exec"
	"time"

	"regexp"
)

// }}}

type CLI struct {
	out   io.WriteCloser
	bio   *bufio.Reader
	lines []string

	lex    *ctx.Message
	target *ctx.Context
	alias  map[string][]string

	*ctx.Context
}

func (cli *CLI) print(str string, arg ...interface{}) bool { // {{{
	if cli.out != nil {
		fmt.Fprintf(cli.out, str, arg...)
		return true
	}
	return false
}

// }}}
func (cli *CLI) parse(m *ctx.Message) (cmd []string) { // {{{

	line := m.Cap("next")
	if m.Cap("next", ""); line == "" {

		if cli.bio == nil {
			line = cli.lines[m.Capi("pos")]
			m.Capi("pos", 1)
		} else {
			cli.print(m.Conf("PS1"))
			l, e := cli.bio.ReadString('\n')
			m.Assert(e)
			line = l
		}
	}

	if line = strings.TrimSpace(line); len(line) == 0 && cli.out != nil {
		line = m.Cap("back")
		m.Cap("back", "")
	}
	if len(line) == 0 || line[0] == '#' {
		return nil
	}

	ls := []string{}
	if cli.lex == nil {
		ls = strings.Split(line, " ")
		cs := []string{}
		for i := 0; i < len(ls); i++ {
			if ls[i] = strings.TrimSpace(ls[i]); ls[i] == "" {
				continue
			}
			if ls[i][0] == '#' {
				break
			}
			cs = append(cs, ls[i])
		}
		ls = cs
	} else {
		lex := m.Spawn(cli.lex.Target())
		m.Assert(lex.Cmd("split", line, "void"))
		ls = lex.Meta["result"]
	}

	if !cli.Has("skip") || !cli.Pulse.Caps("skip") {
		ls = cli.expand(ls)
	}

	if m.Cap("back", line); cli.bio != nil {
		cli.lines = append(cli.lines, line)
		m.Capi("nline", 1)
		m.Capi("pos", 1)
	}

	return ls
}

// }}}
func (cli *CLI) expand(ls []string) []string { // {{{

	cs := []string{}
	for i := 0; i < len(ls); i++ {
		if len(ls[i]) > 0 {
			if r := rune(ls[i][0]); r == '$' || r == '_' || (!unicode.IsNumber(r) && !unicode.IsLetter(r)) {
				if c, ok := cli.alias[string(r)]; ok {
					if i == 0 {
						ns := []string{}
						ns = append(ns, c...)
						if ls[0] = ls[i][1:]; len(ls[0]) > 0 {
							ns = append(ns, ls...)
						} else {
							ns = append(ns, ls[1:]...)
						}
						ls = ns
					} else if len(ls[i]) > 1 {
						key := ls[i][1:]

						if r == rune(key[0]) {
							ls[i] = key
						} else {
							if cli.Context.Has(key, c[0]) {
								switch c[0] {
								case "config":
									ls[i] = cli.Pulse.Conf(key)
								case "cache":
									ls[i] = cli.Pulse.Cap(key)
								}
							} else {
								msg := cli.Pulse.Spawn(cli.target)
								if msg.Exec(c[0], key) != "error: " {
									ls[i] = msg.Get("result")
								}
							}
						}
					}
				}
			}

			if c, ok := cli.alias[ls[i]]; ok && i == 0 {
				ns := []string{}
				ns = append(ns, c...)
				ns = append(ns, ls[1:]...)
			}
		}

		cs = append(cs, ls[i])
	}

	return cs
}

// }}}
func (cli *CLI) express(arg []string) string { // {{{

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
			result = arg[1]

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
			result = arg[0] + arg[2]
			if e1 == nil && e2 == nil {
				result = fmt.Sprintf("%d", v1+v2)
			}
		case "-":
			result = arg[0]
			if e1 == nil && e2 == nil {
				result = fmt.Sprintf("%d", v1-v2)
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
			if arg[0] == arg[2] {
				result = "true"
			}
		case "!=":
			if arg[0] != arg[2] {
				result = "true"
			}
		case "~":
			if m, e := regexp.MatchString(arg[1], arg[0]); m && e == nil {
				result = "true"
			}
		case "!~":
			if m, e := regexp.MatchString(arg[1], arg[0]); !m || e != nil {
				result = "true"
			}

		}
	}

	cli.Pulse.Log("info", nil, "result: %v", result)
	return result
}

// }}}
func (cli *CLI) check(arg []string) bool { // {{{
	switch cli.express(arg) {
	case "", "0", "false":
		return false
	}
	return true
}

// }}}

func (cli *CLI) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}
	c.Caches["skip"] = &ctx.Cache{Name: "跳过执行", Value: cli.Pulse.Cap("skip"), Help: "命令只解析不执行"}

	s := new(CLI)
	s.Context = c
	s.lex = cli.lex
	if m.Has("for") {
		s.lines = append(s.lines, cli.lines[cli.Pulse.Capi("pos")-1:]...)
	} else {
		s.lines = append(s.lines, cli.lines[cli.Pulse.Capi("pos"):]...)
	}
	return s
}

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
				m.Spawn(lex.Target()).Cmd("train", "'[^']*'")
				m.Spawn(lex.Target()).Cmd("train", "\"[^\"]*\"")
				m.Spawn(lex.Target()).Cmd("train", "[^ \t\n]+")
				m.Spawn(lex.Target()).Cmd("train", "[ \n\t]+", "void", "void")
				m.Spawn(lex.Target()).Cmd("train", "#[^\n]*\n", "void", "void")
			}
			cli.lex = lex
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

func (cli *CLI) Start(m *ctx.Message, arg ...string) bool {
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

		m.Cap("next", "source "+m.Cap("init.shy"))
		cli.bio = bufio.NewReader(os.Stdin)
		cli.out = os.Stdout
		m.Conf("lex", "lex")
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
			if cmd := cli.parse(m); cmd != nil {
				m.Spawn(cli.target).Set("detail", cmd...).Post(cli.Context)
			}
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

		if !msg.Has("result") && cli.Owner == ctx.Index.Owner {
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

		result := strings.TrimRight(strings.Join(msg.Meta["result"], ""), "\n")
		if m.Cap("result", result); len(result) > 0 {
			cli.print(result + "\n")
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

func (cli *CLI) Close(m *ctx.Message, arg ...string) bool {
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

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "cli", Help: "管理中心",
	Caches: map[string]*ctx.Cache{
		"skip": &ctx.Cache{Name: "跳过执行", Value: "0", Help: "命令只解析不执行"},
	},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"alias": &ctx.Command{Name: "alias [short [long]]|[delete short]", Help: "查看、定义或删除命令别名, short: 命令别名, long: 命令原名, delete: 删除别名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) {
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
			}
		}},
		"sleep": &ctx.Command{Name: "sleep time", Help: "睡眠, time(ns/us/ms/s/m/h): 时间值(纳秒/微秒/毫秒/秒/分钟/小时)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) {
				if d, e := time.ParseDuration(arg[0]); m.Assert(e) {
					m.Log("info", nil, "sleep %v", d)
					time.Sleep(d)
					m.Log("info", nil, "sleep %v done", d)
				}
			}
		}},
		"var": &ctx.Command{Name: "var a [= exp]", Help: "定义变量, a: 变量名, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) {
				val := ""
				if len(arg) > 2 {
					val = cli.express(arg[2:])
				}
				cli.Pulse.Cap(arg[0], arg[0], val, "临时变量")
			}
		}},
		"let": &ctx.Command{Name: "let a = exp", Help: "设置变量, a: 变量名, exp: 表达式(a {+|-|*|/|%} b)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) {
				m.Echo(cli.Pulse.Cap(arg[0], cli.express(arg[2:])))
			}
		}},
		"source": &ctx.Command{Name: "source file", Help: "运行脚本, file: 脚本文件名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) {
				if f, e := os.Open(arg[0]); m.Assert(e) {
					m.Put("option", "file", f).Start(fmt.Sprintf("%s%d", key, Pulse.Capi("level", 1)), "脚本文件")
					<-m.Target().Exit
					Pulse.Capi("level", -1)
				}
			}
		}},
		"return": &ctx.Command{Name: "return result...", Help: "结束脚本, rusult: 返回值", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) && (!cli.Has("skip") || !cli.Pulse.Caps("skip")) {
				call := cli.Requests[len(cli.Requests)-1]
				call.Set("result", arg...)
				cli.Pulse.Caps("exit", true)
			}
		}},
		"if": &ctx.Command{Name: "if exp", Help: "条件语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) {
				if m.Target(m.Source()); (cli.Has("skip") && cli.Pulse.Caps("skip")) || !cli.check(arg) {
					m.Add("option", "skip")
				}

				m.Put("option", "file", cli.bio).Start(fmt.Sprintf("%s%d", key, Pulse.Capi("level", 1)), "条件语句")
				<-m.Target().Exit
				Pulse.Capi("level", -1)
			}
		}},
		"elif": &ctx.Command{Name: "elif exp", Help: "条件语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) {
				if cli.Pulse.Caps("skip", !cli.Pulse.Caps("else")) {
					return
				}
				cli.Pulse.Caps("else", cli.Pulse.Caps("skip", !cli.check(arg)))
			}
		}},
		"else": &ctx.Command{Name: "else", Help: "条件语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) {
				cli.Pulse.Caps("skip", !cli.Pulse.Caps("else"))
			}
		}},
		"end": &ctx.Command{Name: "end", Help: "结束语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) {
				if cli.Pulse.Caps("exit", true); cli.Pulse.Has("for") && !cli.Pulse.Caps("skip") {
					cli.Pulse.Caps("exit", false)
					cli.Pulse.Cap("pos", "0")

				}
				cli.bio = nil
			}
		}},
		"for": &ctx.Command{Name: "for exp", Help: "循环语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) {
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
			}
		}},
		"function": &ctx.Command{Name: "function name", Help: "函数定义, name: 函数名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Source().Server.(*CLI); m.Assert(ok) {
				if _, ok := cli.Context.Context().Server.(*CLI); ok {
					m.Target(m.Source().Context())
				} else {
					m.Target(m.Source())
				}

				m.Add("option", "skip").Add("option", "save")
				m.Put("option", "file", cli.bio).Start(arg[0], "函数定义")
				<-m.Target().Exit
			}
		}},
		"call": &ctx.Command{Name: "call name arg...", Help: "函数调用, name: 函数名, arg: 参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Target(m.Source())
			m.BackTrace(func(msg *ctx.Message) bool {
				if fun := msg.Find(arg[0], false); fun != nil {
					fun.Add("detail", arg[0], arg[1:]...).Target().Start(fun)
					<-fun.Target().Exit
					m.Set("result", fun.Meta["result"]...)
					return false
				}
				return true
			})
		}},
	},
}

func init() {
	cli := &CLI{}
	cli.Context = Index
	ctx.Index.Register(Index, cli)
}
