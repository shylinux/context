package cli // {{{
// }}}
import ( // {{{
	"contexts"

	"fmt"
	"regexp"
	"strconv"
	"strings"

	"os"
	"os/exec"
	"time"
)

// }}}

type Frame struct {
	key   string
	run   bool
	pos   int
	index int
}

type CLI struct {
	label  map[string]string
	alias  map[string][]string
	target *ctx.Context
	stack  []*Frame

	*ctx.Message
	*ctx.Context
}

func (cli *CLI) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	cli.Message = m
	c.Caches = map[string]*ctx.Cache{
		"level":     &ctx.Cache{Name: "level", Value: "0", Help: "嵌套层级"},
		"parse":     &ctx.Cache{Name: "parse(true/false)", Value: "true", Help: "命令解析"},
		"last_msg":  &ctx.Cache{Name: "last_msg", Value: "0", Help: "前一条消息"},
		"ps_count":  &ctx.Cache{Name: "ps_count", Value: "0", Help: "命令计数"},
		"ps_target": &ctx.Cache{Name: "ps_target", Value: c.Name, Help: "当前模块"},
	}
	c.Configs = map[string]*ctx.Config{
		"ps_time": &ctx.Config{Name: "ps_time", Value: "[15:04:05]", Help: "当前时间", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				return arg[0]
			}
			return time.Now().Format(x.Value)
			// }}}
		}},
		"ps_end": &ctx.Config{Name: "ps_end", Value: "> ", Help: "命令行提示符结尾"},
		"prompt": &ctx.Config{Name: "prompt(ps_target/ps_time)", Value: "ps_count ps_time ps_target ps_end", Help: "命令行提示符, 以空格分隔, 依次显示各种信息", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				return arg[0]
			}

			ps := make([]string, 0, 3)
			for _, v := range strings.Split(x.Value, " ") {
				if m.Conf(v) != "" {
					ps = append(ps, m.Conf(v))
				} else {
					ps = append(ps, m.Cap(v))
				}
			}
			return strings.Join(ps, "")
			// }}}
		}},
	}

	s := new(CLI)
	s.Context = c
	s.target = c
	s.alias = map[string][]string{
		"~": []string{"context"},
		"!": []string{"message"},
		"@": []string{"config"},
		"$": []string{"cache"},
	}

	return s
}

// }}}
func (cli *CLI) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	cli.Message = m
	return cli
}

// }}}
func (cli *CLI) Start(m *ctx.Message, arg ...string) bool { // {{{
	cli.Message = m
	m.Sess("cli", m)
	yac := m.Sess("yac")
	if yac.Cap("status") != "start" {
		yac.Target().Start(yac)
		yac.Cmd("train", "void", "void", "[\t ]+")

		yac.Cmd("train", "key", "key", "[A-Za-z_][A-Za-z_0-9]*")
		yac.Cmd("train", "num", "num", "mul{", "0", "-?[1-9][0-9]*", "0[0-9]+", "0x[0-9]+", "}")
		yac.Cmd("train", "str", "str", "mul{", "\"[^\"]*\"", "'[^']*'", "}")
		yac.Cmd("train", "tran", "tran", "mul{", "@", "$", "}", "opt{", "[a-zA-Z0-9_]+", "}")
		yac.Cmd("train", "tran", "tran", "$", "(", "cache", ")")

		yac.Cmd("train", "op1", "op1", "mul{", "$", "@", "}")
		yac.Cmd("train", "op1", "op1", "mul{", "-z", "-n", "}")
		yac.Cmd("train", "op1", "op1", "mul{", "-e", "-f", "-d", "}")
		yac.Cmd("train", "op1", "op1", "mul{", "-", "+", "}")
		yac.Cmd("train", "op2", "op2", "mul{", "=", "+=", "}")
		yac.Cmd("train", "op2", "op2", "mul{", "+", "-", "*", "/", "%", "}")
		yac.Cmd("train", "op2", "op2", "mul{", ">", ">=", "<", "<=", "==", "!=", "}")

		yac.Cmd("train", "val", "val", "opt{", "op1", "}", "mul{", "num", "key", "str", "tran", "}")
		yac.Cmd("train", "exp", "exp", "val", "rep{", "op2", "val", "}")
		yac.Cmd("train", "val", "val", "(", "exp", ")")
		yac.Cmd("train", "stm", "var", "var", "key")
		yac.Cmd("train", "stm", "var", "var", "key", "=", "exp")
		yac.Cmd("train", "stm", "var", "var", "key", "<-", "exp")
		yac.Cmd("train", "stm", "let", "let", "key", "mul{", "=", "<-", "}", "exp")

		yac.Cmd("train", "stm", "if", "if", "exp")
		yac.Cmd("train", "stm", "else", "else")
		yac.Cmd("train", "stm", "end", "end")

		yac.Cmd("train", "word", "word", "mul{", "~", "!", "=", "tran", "str", "[a-zA-Z0-9_/\\-.:]+", "}")

		yac.Cmd("train", "stm", "elif", "elif", "exp")
		yac.Cmd("train", "stm", "for", "for", "exp")
		yac.Cmd("train", "stm", "for", "for", "exp", ";", "exp")
		yac.Cmd("train", "stm", "for", "for", "index", "word", "word", "word")
		yac.Cmd("train", "stm", "function", "function", "rep{", "key", "}")
		yac.Cmd("train", "stm", "return", "return", "rep{", "exp", "}")

		yac.Cmd("train", "cmd", "goto", "goto", "word", "exp")
		yac.Cmd("train", "cmd", "cmd", "cache", "rep{", "word", "}")
		yac.Cmd("train", "cmd", "cmd", "cache", "key", "rep{", "word", "}")
		yac.Cmd("train", "cmd", "cmd", "cache", "key", "opt{", "=", "exp", "}")
		yac.Cmd("train", "cmd", "cmd", "rep{", "word", "}")
		yac.Cmd("train", "tran", "tran", "$", "(", "cmd", ")")

		yac.Cmd("train", "line", "line", "opt{", "mul{", "stm", "cmd", "}", "}", "mul{", ";", "\n", "#[^\n]*\n", "}")
	}

	m.Options("scan_end", false)
	m.Optionv("ps_target", cli.target)
	m.Option("prompt", m.Conf("prompt"))
	m.Cap("stream", m.Spawn(yac.Target()).Call(func(cmd *ctx.Message) *ctx.Message {
		if !m.Caps("parse") {
			switch cmd.Detail(0) {
			case "if":
				cmd.Set("detail", "if", "false")
			case "else":
			case "end":
			case "for":
			default:
				cmd.Hand = true
				return nil
			}
		}

		if m.Option("prompt", cmd.Cmd().Conf("prompt")); cmd.Has("return") {
			m.Result(0, cmd.Meta["return"])
			m.Options("scan_end", true)
			m.Target().Close(m.Spawn())
		}
		m.Optionv("ps_target", cli.target)
		return nil
	}, "parse", arg[1]).Target().Name)

	if arg[1] == "stdio" {
		msg := m.Spawn().Cmd("source", m.Conf("init.shy"))
		msg.Result(0, msg.Meta["return"])
	}
	return false
}

// }}}
func (cli *CLI) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch cli.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "cli", Help: "管理中心",
	Caches: map[string]*ctx.Cache{
		"nshell": &ctx.Cache{Name: "nshell", Value: "0", Help: "终端数量"},
	},
	Configs: map[string]*ctx.Config{
		"init.shy": &ctx.Config{Name: "init.shy", Value: "etc/init.shy", Help: "启动脚本"},
		"cli_name": &ctx.Config{Name: "cli_name", Value: "shell", Help: "模块命名", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				return arg[0]
			}
			return fmt.Sprintf("%s%d", x.Value, m.Capi("nshell", 1))
			// }}}
		}},
		"cli_help": &ctx.Config{Name: "cli_help", Value: "shell", Help: "模块文档"},

		"time_format":   &ctx.Config{Name: "time_format", Value: "2006-01-02 15:04:05", Help: "时间格式"},
		"time_unit":     &ctx.Config{Name: "time_unit", Value: "1000", Help: "时间倍数"},
		"time_interval": &ctx.Config{Name: "time_interval(open/close)", Value: "open", Help: "时间区间"},
	},
	Commands: map[string]*ctx.Command{
		"source": &ctx.Command{
			Name: "source filename [async [cli_name [cli_help]]",
			Help: "解析脚本, filename: 文件名, async: 异步执行, cli_name: 模块名, cli_help: 模块帮助",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if _, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
					m.Start(m.Confx("cli_name", arg, 2), m.Confx("cli_help", arg, 3), key, arg[0])
					if len(arg) < 2 || arg[1] != "async" {
						m.Wait()
					}
				} // }}}
			}},
		"label": &ctx.Command{Name: "label name", Help: "记录当前脚本的位置, name: 位置名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				if cli.label == nil {
					cli.label = map[string]string{}
				}
				cli.label[arg[0]] = m.Option("file_pos")
			} // }}}
		}},
		"goto": &ctx.Command{Name: "goto label [condition]", Help: "向上跳转到指定位置, label: 跳转位置, condition: 跳转条件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				if pos, ok := cli.label[arg[1]]; ok {
					if len(arg) > 2 && !ctx.Right(arg[2]) {
						return
					}
					m.Append("file_pos0", pos)
				}
			} // }}}
		}},
		"return": &ctx.Command{Name: "return result...", Help: "结束脚本, result: 返回值", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Add("append", "return", arg[1:])
		}},
		"target": &ctx.Command{Name: "target module", Help: "设置当前模块, module: 模块全名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				if len(arg) == 0 {
					m.Echo("%s", m.Cap("ps_target"))
					return
				}
				if msg := m.Find(arg[0]); msg != nil {
					cli.target = msg.Target()
					m.Cap("ps_target", cli.target.Name)
				}
			} // }}}
		}},
		"alias": &ctx.Command{
			Name: "alias [short [long...]]|[delete short]|[import module [command [alias]]]",
			Help: "查看、定义或删除命令别名, short: 命令别名, long: 命令原名, delete: 删除别名, import导入模块所有命令",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
					switch len(arg) {
					case 0:
						for k, v := range cli.alias {
							m.Echo("%s: %v\n", k, v)
						}
					case 1:
						m.Echo("%s: %v\n", arg[0], cli.alias[arg[0]])
					default:
						switch arg[0] {
						case "delete":
							m.Echo("delete: %s %v\n", arg[1], cli.alias[arg[1]])
							delete(cli.alias, arg[1])
						case "import":
							msg := m.Find(arg[1], false)
							if msg == nil {
								msg = m.Find(arg[1], true)
							}
							if msg == nil {
								m.Echo("%s not exist", arg[1])
								return
							}
							m.Log("info", "import %s", arg[1])
							module := msg.Cap("module")
							for k, _ := range msg.Target().Commands {
								if len(arg) == 2 {
									cli.alias[k] = []string{module + "." + k}
									continue
								}
								if key := k; k == arg[2] {
									if len(arg) > 3 {
										key = arg[3]
									}
									cli.alias[key] = []string{module + "." + k}
									break
								}
							}
						default:
							cli.alias[arg[0]] = arg[1:]
							m.Echo("%s: %v\n", arg[0], cli.alias[arg[0]])
							m.Log("info", "%s: %v", arg[0], cli.alias[arg[0]])
						}
					}
				} // }}}
			}},
		"sleep": &ctx.Command{Name: "sleep time", Help: "睡眠, time(ns/us/ms/s/m/h): 时间值(纳秒/微秒/毫秒/秒/分钟/小时)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if d, e := time.ParseDuration(arg[0]); m.Assert(e) { // {{{
				m.Log("info", "sleep %v", d)
				time.Sleep(d)
				m.Log("info", "sleep %v done", d)
			} // }}}
		}},
		"time": &ctx.Command{
			Name: "time [time_format format] [parse when] when [begin|end|yestoday|tommorow|monday|sunday|first|last|origin|last]",
			Form: map[string]int{"time_format": 1, "parse": 1, "time_interval": 1},
			Help: "查看时间, time_format: 输出或解析的时间格式, parse: 输入的时间字符串, when: 输入的时间戳, 其它是时间偏移",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				t := time.Now() // {{{
				if m.Options("parse") {
					n, e := time.ParseInLocation(m.Confx("time_format"), m.Option("parse"), time.Local)
					m.Assert(e)
					t = n
				}

				if len(arg) > 0 {
					if i, e := strconv.Atoi(arg[0]); e == nil {
						m.Option("time_format", m.Conf("time_format"))
						t = time.Unix(int64(i/m.Confi("time_unit")), 0)
						arg = arg[1:]
					} else if n, e := time.ParseInLocation(m.Confx("time_format"), arg[0], time.Local); e == nil {
						m.Option("parse", arg[0])
						arg = arg[1:]
						t = n
					}
				}

				if len(arg) > 0 {
					switch arg[0] {
					case "begin":
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
					m.Echo("%d", t.Unix()*int64(m.Confi("time_unit")))
				} else {
					m.Echo(t.Format(m.Confx("time_format")))
				}
				// }}}
			}},
		"echo": &ctx.Command{Name: "echo arg...", Help: "函数调用, name: 函数名, arg: 参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Echo("%s", strings.Join(arg, ""))
		}},

		"tran": &ctx.Command{Name: "tran word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				msg := m.Spawn(cli.target)
				switch len(arg) {
				case 1:
					m.Echo(arg[0])
				case 2:
					switch arg[0] {
					case "$":
						m.Echo(msg.Cap(arg[1]))
					case "@":
						m.Echo(msg.Conf(arg[1]))
					default:
						m.Echo(arg[0])
						m.Echo(arg[1])
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
			} //}}}
		}},
		"str": &ctx.Command{Name: "str word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Echo(arg[0][1 : len(arg[0])-1])
		}},
		"val": &ctx.Command{Name: "val exp", Help: "表达式运算", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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
				case "=":
					result = m.Cap(arg[0], arg[2])
				case "+=":
					if i, e := strconv.Atoi(m.Cap(arg[0])); e == nil && e2 == nil {
						result = m.Cap(arg[0], fmt.Sprintf("%d", v2+i))
					} else {
						result = m.Cap(arg[0], m.Cap(arg[0])+arg[2])
					}
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
			// }}}
		}},
		"exp": &ctx.Command{Name: "exp word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			pre := map[string]int{ // {{{
				"=": 1,
				"+": 2, "-": 2,
				"*": 3, "/": 3, "%": 3,
			}
			num := []string{arg[0]}
			op := []string{}

			for i := 1; i < len(arg); i += 2 {
				if len(op) > 0 && pre[op[len(op)-1]] >= pre[arg[i]] {
					num[len(op)-1] = m.Spawn().Cmd("val", num[len(op)-1], op[len(op)-1], num[len(op)]).Get("result")
					num = num[:len(num)-1]
					op = op[:len(op)-1]
				}

				num = append(num, arg[i+1])
				op = append(op, arg[i])
			}

			for i := len(op) - 1; i >= 0; i-- {
				num[i] = m.Spawn().Cmd("val", num[i], op[i], num[i+1]).Get("result")
			}

			m.Echo("%s", num[0])
			// }}}
		}},
		"var": &ctx.Command{Name: "var a [= exp]", Help: "定义变量, a: 变量名, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if m.Cap(arg[1], arg[1], "", "临时变量"); len(arg) > 3 { // {{{
				switch arg[2] {
				case "=":
					m.Cap(arg[1], arg[3])
				case "<-":
					m.Cap(arg[1], m.Cap("last_msg"))
				}
			}
			m.Echo(m.Cap(arg[1]))
			// }}}
		}},
		"let": &ctx.Command{Name: "let a = exp", Help: "设置变量, a: 变量名, exp: 表达式(a {+|-|*|/|%} b)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			switch arg[2] { // {{{
			case "=":
				m.Cap(arg[1], arg[3])
			case "<-":
				m.Cap(arg[1], m.Cap("last_msg"))
			}
			m.Echo(m.Cap(arg[1]))
			// }}}
		}},
		"if": &ctx.Command{Name: "if exp", Help: "条件语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				run := m.Caps("parse") && ctx.Right(arg[1])
				cli.stack = append(cli.stack, &Frame{pos: m.Optioni("file_pos"), key: key, run: run})
				m.Capi("level", 1)
				m.Caps("parse", run)
			} // }}}
		}},
		"else": &ctx.Command{Name: "else", Help: "条件语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				if !m.Caps("parse") {
					m.Caps("parse", true)
				} else {
					if len(cli.stack) == 1 {
						m.Caps("parse", false)
					} else {
						frame := cli.stack[len(cli.stack)-2]
						if frame.run {
							m.Caps("parse", false)
						}
					}
				}
			} // }}}
		}},
		"end": &ctx.Command{Name: "end", Help: "结束语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				if frame := cli.stack[len(cli.stack)-1]; frame.key == "for" && frame.run {
					m.Append("file_pos0", frame.pos)
					return
				}

				if cli.stack = cli.stack[:len(cli.stack)-1]; m.Capi("level", -1) > 0 {
					frame := cli.stack[len(cli.stack)-1]
					m.Caps("parse", frame.run)
				} else {
					m.Caps("parse", true)
				}
			} // }}}
		}},
		"for": &ctx.Command{Name: "for [express ;] condition", Help: "循环语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				run := m.Caps("parse")
				defer func() { m.Caps("parse", run) }()

				msg := m
				if run {
					if arg[1] == "index" {
						if code, e := strconv.Atoi(arg[2]); m.Assert(e) {
							msg = cli.Message.Tree(code)
							run = run && msg != nil && msg.Meta != nil && len(msg.Meta[arg[3]]) > 0
						}
					} else if len(arg) > 3 {
						run = run && ctx.Right(arg[3])
					} else {
						run = run && ctx.Right(arg[1])
					}

					if len(cli.stack) > 0 {
						if frame := cli.stack[len(cli.stack)-1]; frame.key == "for" && frame.pos == m.Optioni("file_pos") {
							if arg[1] == "index" {
								frame.index++
								if run = run && len(msg.Meta[arg[3]]) > frame.index; run {
									m.Cap(arg[4], msg.Meta[arg[3]][frame.index])
								}
							}
							frame.run = run
							return
						}
					}
				}

				cli.stack = append(cli.stack, &Frame{pos: m.Optioni("file_pos"), key: key, run: run, index: 0})
				if m.Capi("level", 1); run && arg[1] == "index" {
					m.Cap(arg[4], arg[4], msg.Meta[arg[3]][0], "临时变量")
				}
			} // }}}
		}},
		"cmd": &ctx.Command{Name: "cmd word", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) { // {{{
				detail := []string{}

				if a, ok := cli.alias[arg[0]]; ok {
					detail = append(detail, a...)
					detail = append(detail, arg[1:]...)
				} else {
					detail = append(detail, arg...)
				}

				if detail[0] != "context" {
					target := cli.target
					defer func() {
						cli.target = target
						m.Cap("ps_target", cli.target.Name)
					}()
				}

				routes := strings.Split(detail[0], ".")
				msg := m
				if len(routes) > 1 {

					route := strings.Join(routes[:len(routes)-1], ".")
					if msg = m.Find(route, false); msg == nil {
						msg = m.Find(route, true)
					}

					if msg == nil {
						m.Echo("%s not exist", route)
						return
					}
					detail[0] = routes[len(routes)-1]
				} else {
					msg = m.Spawn(cli.target)

				}

				m.Capi("ps_count", 1)
				m.Capi("last_msg", 0, msg.Code())
				if msg.Cmd(detail); msg.Hand {
					cli.target = msg.Target()
					m.Cap("ps_target", cli.target.Name)
				} else {
					msg.Hand = true
					msg.Log("system", "%v", msg.Meta["detail"])

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
				m.Copy(msg, "result").Copy(msg, "append")
			}
			// }}}
		}},
	},
}

func init() {
	cli := &CLI{}
	cli.Context = Index
	ctx.Index.Register(Index, cli)
}
