package cli

import (
	"bytes"
	"contexts/ctx"
	"encoding/csv"
	"encoding/json"
	"path"
	"toolkit"

	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Frame struct {
	key   string
	run   bool
	pos   int
	index int
	list  []string
}
type CLI struct {
	label map[string]string
	stack []*Frame

	*time.Timer
	Context *ctx.Context
}

func (cli *CLI) schedule(m *ctx.Message) string {
	first, timer := "", int64(1<<50)
	for k, v := range m.Confv("timer").(map[string]interface{}) {
		val := v.(map[string]interface{})
		if val["action_time"].(int64) < timer && !val["done"].(bool) {
			first, timer = k, val["action_time"].(int64)
		}
	}
	cli.Timer.Reset(time.Until(time.Unix(0, timer/int64(m.Confi("time_unit"))*1000000000)))
	return m.Conf("timer_next", first)
}

func (cli *CLI) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{
		"level": &ctx.Cache{Name: "level", Value: "0", Help: "嵌套层级"},
		"parse": &ctx.Cache{Name: "parse(true/false)", Value: "true", Help: "命令解析"},
	}

	return &CLI{Context: c}
}
func (cli *CLI) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return cli
}
func (cli *CLI) Start(m *ctx.Message, arg ...string) bool {
	m.Cap("stream", m.Sess("yac").Call(func(cmd *ctx.Message) *ctx.Message {
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

		if cmd.Cmd(); cmd.Has("return") {
			m.Options("scan_end", true)
			m.Target().Close(m)
		}

		v := cmd.Optionv("ps_target")
		if v != nil {
			m.Optionv("ps_target", v)
		}
		return nil
	}, "scan", arg).Target().Name)

	return false
}
func (cli *CLI) Close(m *ctx.Message, arg ...string) bool {
	switch cli.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

var Index = &ctx.Context{Name: "cli", Help: "管理中心",
	Caches: map[string]*ctx.Cache{
		"nshell": &ctx.Cache{Name: "nshell", Value: "0", Help: "终端数量"},
	},
	Configs: map[string]*ctx.Config{
		"init_shy": &ctx.Config{Name: "init_shy", Value: "etc/init.shy", Help: "启动脚本"},
		"exit_shy": &ctx.Config{Name: "exit_shy", Value: "etc/exit.shy", Help: "启动脚本"},
		"cmd_script": &ctx.Config{Name: "cmd_script", Value: map[string]interface{}{
			"sh": "bash", "shy": "source", "py": "python",
		}, Help: "系统命令超时"},
		"alias": &ctx.Config{Name: "alias", Value: map[string]interface{}{
			"~":  []string{"context"},
			"!":  []string{"message"},
			":":  []string{"command"},
			"::": []string{"command", "list"},

			"pwd":  []string{"nfs.pwd"},
			"path": []string{"nfs.path"},
			"dir":  []string{"nfs.dir"},
			"git":  []string{"nfs.git"},
			"brow": []string{"web.brow"},
		}, Help: "启动脚本"},

		"time_unit":  &ctx.Config{Name: "time_unit", Value: "1000", Help: "时间倍数"},
		"time_close": &ctx.Config{Name: "time_close(open/close)", Value: "open", Help: "时间区间"},

		"source_list": &ctx.Config{Name: "source_list", Value: []interface{}{}, Help: "系统命令超时"},
		"system_env":  &ctx.Config{Name: "system_env", Value: map[string]interface{}{}, Help: "系统命令超时"},
		"cmd_timeout": &ctx.Config{Name: "cmd_timeout", Value: "60s", Help: "系统命令超时"},
		"cmd_combine": &ctx.Config{Name: "cmd_combine", Value: map[string]interface{}{
			"vi":  map[string]interface{}{"active": true},
			"top": map[string]interface{}{"active": true},
			"ls":  map[string]interface{}{"arg": []interface{}{"-l"}},
		}, Help: "系统命令配置, active: 交互方式, cmd: 命令映射, arg: 命令参数, args: 子命令参数, path: 命令目录, env: 环境变量, dir: 工作目录"},

		"timer":      &ctx.Config{Name: "timer", Value: map[string]interface{}{}, Help: "定时器"},
		"timer_next": &ctx.Config{Name: "timer_next", Value: "", Help: "定时器"},
	},
	Commands: map[string]*ctx.Command{
		"source": &ctx.Command{Name: "source [script|stdio|snippet]", Help: "解析脚本, script: 脚本文件, stdio: 命令终端, snippet: 代码片段", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmdy("dir", "", "dir_deep", "dir_reg", ".*\\.(sh|shy|py)$")
				return
			}

			// 解析脚本文件
			if p := m.Cmdx("nfs.path", arg[0]); p != "" && strings.Contains(p, ".") {
				arg[0] = p
				switch path.Ext(p) {
				case "":
				case ".shy":
					m.Option("scan_end", "false")
					m.Start(fmt.Sprintf("shell%d", m.Capi("nshell", 1)), "shell", arg...)
					m.Wait()
				default:
					m.Cmdy("system", m.Conf("cmd_script", strings.TrimPrefix(path.Ext(p), ".")), arg)
				}
				return
			}

			// 解析终端命令
			if arg[0] == "stdio" {
				m.Option("scan_end", "false")
				m.Start("shy", "shell", "stdio", "engine")
				m.Wait()
				return
			}

			// 解析代码片段
			m.Sess("yac").Call(func(msg *ctx.Message) *ctx.Message {
				switch msg.Cmd().Detail(0) {
				case "cmd":
					m.Set("append").Copy(msg, "append")
					m.Set("result").Copy(msg, "result")
				}
				return nil
			}, "parse", "line", "void", strings.Join(arg, " "))
			return
		}},
		"system": &ctx.Command{Name: "system word...", Help: []string{"调用系统命令, word: 命令",
			"cmd_active(true/false): 是否交互", "cmd_timeout: 命令超时", "cmd_env: 环境变量", "cmd_dir: 工作目录"},
			Form: map[string]int{"cmd_active": 1, "cmd_timeout": 1, "cmd_env": 2, "cmd_dir": 1, "cmd_error": 0, "cmd_parse": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				for _, v := range m.Meta["result"] {
					if strings.TrimSpace(v) != "" {
						arg = append(arg, v)
					}
				}

				conf := m.Confm("cmd_combine", arg[0])
				if v, ok := kit.Chain(conf, "cmd").(string); ok {
					arg[0] = m.Parse(v)
				}

				args := []string{arg[0]}
				if list, ok := kit.Chain(conf, "arg").([]interface{}); ok {
					for _, v := range list {
						args = append(args, m.Parse(v))
					}
				}
				args = append(args, arg[1:]...)
				cmd := exec.Command(args[0], args[1:]...)

				if v, ok := kit.Chain(conf, "path").(string); ok {
					cmd.Path = m.Parse(v)
				}
				m.Log("info", "cmd %v %v", cmd.Path, cmd.Args)

				for k, v := range m.Confm("system_env") {
					cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, m.Parse(v)))
				}
				if list, ok := kit.Chain(conf, "env").([]interface{}); ok {
					for k, v := range list {
						cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, m.Parse(v)))
					}
				}
				for i := 0; i < len(m.Meta["cmd_env"])-1; i += 2 {
					cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", m.Meta["cmd_env"][i], m.Parse(m.Meta["cmd_env"][i+1])))
				}
				m.Log("info", "cmd.env %v", cmd.Env)
				for _, v := range os.Environ() {
					cmd.Env = append(cmd.Env, v)
				}

				if m.Options("cmd_dir") {
					cmd.Dir = m.Option("cmd_dir")
				} else if v, ok := kit.Chain(conf, "dir").(string); ok {
					cmd.Dir = m.Parse(v)
				}
				m.Log("info", "cmd.dir %v", cmd.Dir)

				if m.Options("cmd_active") || kit.Right(conf["active"]) {
					cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
					if e := cmd.Start(); e != nil {
						m.Echo("error: ").Echo("%s\n", e)
					} else if e := cmd.Wait(); e != nil {
						m.Echo("error: ").Echo("%s\n", e)
					}
				} else {
					wait := make(chan bool, 1)
					m.GoFunc(m, func(m *ctx.Message) {
						out := bytes.NewBuffer(make([]byte, 0, 1024))
						err := bytes.NewBuffer(make([]byte, 0, 1024))
						cmd.Stdout = out
						cmd.Stderr = err
						if e := cmd.Run(); e != nil {
							m.Echo("error: ").Echo("%s\n", e).Echo(err.String())
						} else {
							switch m.Option("cmd_parse") {
							case "json":
								var data interface{}
								if json.Unmarshal(out.Bytes(), &data) == nil {
									msg := m.Spawn().Put("option", "data", data).Cmd("trans", "data", "")
									m.Copy(msg, "append").Copy(msg, "result")
								} else {
									m.Echo(out.String())
								}

							case "csv":
								data, e := csv.NewReader(out).ReadAll()
								m.Assert(e)
								for i := 1; i < len(data); i++ {
									for j := 0; j < len(data[i]); j++ {
										m.Add("append", data[0][j], data[i][j])
									}
								}
								m.Table()
							default:
								m.Echo(out.String())
							}
						}
						wait <- true
					})

					timeout := m.Conf("cmd_timeout")
					if conf["timeout"] != nil {
						timeout = conf["timeout"].(string)
					}
					if m.Option("timeout") != "" {
						timeout = m.Option("timeout")
					}

					d, e := time.ParseDuration(timeout)
					m.Assert(e)

					select {
					case <-time.After(d):
						cmd.Process.Kill()
						m.Echo("%s: %s timeout", arg[0], m.Conf("cmd_timeout"))
					case <-wait:
					}
				}
				return
			}},

		"alias": &ctx.Command{Name: "alias [short [long...]]|[delete short]|[import module [command [alias]]]",
			Help: "查看、定义或删除命令别名, short: 命令别名, long: 命令原名, delete: 删除别名, import导入模块所有命令",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				switch len(arg) {
				case 0:
					m.Cmdy("ctx.config", "alias")
				case 1:
					m.Cmdy("ctx.config", "alias", arg[0])
				default:
					switch arg[0] {
					case "delete":
						alias := m.Confm("alias")
						m.Echo("delete: %s %v\n", arg[1], alias[arg[1]])
						delete(alias, arg[1])
					case "import":
						msg := m.Find(arg[1], false)
						if msg == nil {
							msg = m.Find(arg[1], true)
						}
						if msg == nil {
							m.Echo("%s not exist", arg[1])
							return
						}

						module := msg.Cap("module")
						for k, _ := range msg.Target().Commands {
							if len(k) > 0 && k[0] == '/' {
								continue
							}

							if len(arg) == 2 {
								m.Confv("alias", k, []string{module + "." + k})
								m.Log("info", "import %s.%s", module, k)
								continue
							}

							if key := k; k == arg[2] {
								if len(arg) > 3 {
									key = arg[3]
								}
								m.Confv("alias", key, []string{module + "." + k})
								m.Log("info", "import %s.%s as %s", module, k, key)
								break
							}
						}
					default:
						m.Confv("alias", arg[0], arg[1:])
						m.Log("info", "%s: %v", arg[0], arg[1:])
					}
				}
				return
			}},
		"cmd": &ctx.Command{Name: "cmd word", Help: "解析命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 解析别名
			detail := []string{}
			if alias, ok := m.Confv("alias", arg[0]).([]string); ok {
				detail, arg = append(detail, alias...), arg[1:]
			}
			detail = append(detail, arg...)

			if m.Options("current_ctx") && detail[0] != "context" {
				detail = append([]string{"context", m.Option("current_ctx")}, detail...)
				m.Option("current_ctx", "")
			}

			// 目标切换
			target := m.Optionv("ps_target")
			if detail[0] != "context" {
				defer func() { m.Optionv("ps_target", target) }()
			}

			// 解析脚本
			msg := m
			for k, v := range m.Confv("cmd_script").(map[string]interface{}) {
				if strings.HasSuffix(detail[0], "."+k) {
					msg = m.Spawn(m.Optionv("ps_target"))
					detail[0] = m.Cmdx("nfs.path", detail[0])
					detail = append([]string{v.(string)}, detail...)
					break
				}
			}

			// 解析路由
			if msg == m {
				if routes := strings.Split(detail[0], "."); len(routes) > 1 {
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
					msg = m.Spawn(m.Optionv("ps_target"))
				}
			}
			msg.Copy(m, "option").Copy(m, "append")

			// 解析命令
			args, rest := []string{}, []string{}
			exports := []map[string]string{}
			exec, execexec := true, false
			for i := 0; i < len(detail); i++ {
				switch detail[i] {
				case "?":
					if !kit.Right(detail[i+1]) {
						return
					}
					i++
				case "??":
					exec = false
					execexec = execexec || kit.Right(detail[i+1])
					i++
				case "<":
					m.Cmdy("nfs.import", detail[i+1])
					i++
				case ">":
					exports = append(exports, map[string]string{"file": detail[i+1]})
					i++
				case ">$":
					if i == len(detail)-2 {
						exports = append(exports, map[string]string{"cache": detail[i+1], "index": "result"})
						i += 1
						break
					}
					exports = append(exports, map[string]string{"cache": detail[i+1], "index": detail[i+2]})
					i += 2
				case ">@":
					if i == len(detail)-2 {
						exports = append(exports, map[string]string{"config": detail[i+1], "index": "result"})
						i += 1
						break
					}
					exports = append(exports, map[string]string{"config": detail[i+1], "index": detail[i+2]})
					i += 2
				case "|":
					detail, rest = detail[:i], detail[i+1:]
				case "%":
					rest = append(rest, "select")
					detail, rest = detail[:i], append(rest, detail[i+1:]...)
				default:
					args = append(args, detail[i])
				}
			}
			if !exec && !execexec {
				return
			}

			// 执行命令
			if msg.Set("detail", args).Cmd(); !msg.Hand {
				msg.Cmd("system", args)
			}
			if msg.Appends("ps_target1") {
				target = msg.Target()
			}

			// 管道命令
			if len(rest) > 0 {
				pipe := m.Spawn().Copy(msg, "option")
				pipe.Copy(msg, "append").Copy(msg, "result").Cmd("cmd", rest)
				msg.Set("append").Copy(pipe, "append")
				msg.Set("result").Copy(pipe, "result")
			}

			// 导出结果
			for _, v := range exports {
				if v["file"] != "" {
					m.Sess("nfs").Copy(msg, "option").Copy(msg, "append").Copy(msg, "result").Cmd("export", v["file"])
				}
				if v["cache"] != "" {
					if v["index"] == "result" {
						m.Cap(v["cache"], strings.Join(msg.Meta["result"], ""))
					} else {
						m.Cap(v["cache"], msg.Append(v["index"]))
					}
				}
				if v["config"] != "" {
					if v["index"] == "result" {
						m.Conf(v["config"], strings.Join(msg.Meta["result"], ""))
					} else {
						m.Conf(v["config"], msg.Append(v["index"]))
					}
				}
			}

			// 返回结果
			m.Optionv("ps_target", msg.Target())
			m.Set("append").Copy(msg, "append")
			m.Set("result").Copy(msg, "result")
			return
		}},
		"str": &ctx.Command{Name: "str word", Help: "解析字符串", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo(arg[0][1 : len(arg[0])-1])
			return
		}},
		"exe": &ctx.Command{Name: "exe $ ( cmd )", Help: "解析嵌套命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch len(arg) {
			case 1:
				m.Echo(arg[0])
			case 2:
				msg := m.Spawn(m.Optionv("ps_target"))
				switch arg[0] {
				case "$":
					m.Echo(msg.Cap(arg[1]))
				case "@":
					value := msg.Option(arg[1])
					if value == "" {
						value = msg.Conf(arg[1])
					}

					m.Echo(value)
				default:
					m.Echo(arg[0]).Echo(arg[1])
				}
			default:
				switch arg[0] {
				case "$", "@":
					m.Result(0, arg[2:len(arg)-1])
				}
			}
			return
		}},
		"val": &ctx.Command{Name: "val exp", Help: "表达式运算", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
				case "+":
					result = arg[1]
				case "-":
					result = arg[1]
					if i, e := strconv.Atoi(arg[1]); e == nil {
						result = fmt.Sprintf("%d", -i)
					}
				}
			case 3:
				v1, e1 := strconv.Atoi(arg[0])
				v2, e2 := strconv.Atoi(arg[2])
				switch arg[1] {
				case ":=":
					if !m.Target().Has(arg[0]) {
						result = m.Cap(arg[0], arg[0], arg[2], "临时变量")
					}
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

			return
		}},
		"exp": &ctx.Command{Name: "exp word", Help: "表达式运算", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) > 0 && arg[0] == "{" {
				msg := m.Spawn()
				for i := 1; i < len(arg); i++ {
					key := arg[i]
					for i += 3; i < len(arg); i++ {
						if arg[i] == "]" {
							break
						}
						msg.Add("append", key, arg[i])
					}
				}
				m.Echo("%d", msg.Code())
				return
			}

			pre := map[string]int{
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
			return
		}},
		"let": &ctx.Command{Name: "let a = exp", Help: "设置变量, a: 变量名, exp: 表达式(a {+|-|*|/|%} b)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch arg[2] {
			case "=":
				m.Cap(arg[1], arg[3])
			case "<-":
				m.Cap(arg[1], m.Cap("last_msg"))
			}
			m.Echo(m.Cap(arg[1]))
			return
		}},
		"var": &ctx.Command{Name: "var a [= exp]", Help: "定义变量, a: 变量名, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if m.Cap(arg[1], arg[1], "", "临时变量"); len(arg) > 1 {
				switch arg[2] {
				case "=":
					m.Cap(arg[1], arg[3])
				case "<-":
					m.Cap(arg[1], m.Cap("last_msg"))
				}
			}
			m.Echo(m.Cap(arg[1]))
			return
		}},
		"expr": &ctx.Command{Name: "expr arg...", Help: "输出表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo("%s", strings.Join(arg[1:], ""))
			return
		}},
		"return": &ctx.Command{Name: "return result...", Help: "结束脚本, result: 返回值", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Add("append", "return", arg[1:])
			return
		}},
		"arguments": &ctx.Command{Name: "arguments", Help: "脚本参数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Set("result", m.Optionv("arguments"))
			return
		}},

		"tmux": &ctx.Command{Name: "tmux buffer", Help: "终端管理, buffer: 查看复制", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch arg[0] {
			case "buffer":
				bufs := strings.Split(m.Spawn().Cmd("system", "tmux", "list-buffers").Result(0), "\n")

				n := 3
				if m.Option("limit") != "" {
					n = m.Optioni("limit")
				}

				for i, b := range bufs {
					if i >= n {
						break
					}
					bs := strings.SplitN(b, ": ", 3)
					if len(bs) > 1 {
						m.Add("append", "buffer", bs[0][:len(bs[0])])
						m.Add("append", "length", bs[1][:len(bs[1])-6])
						m.Add("append", "strings", bs[2][1:len(bs[2])-1])
					}
				}

				if m.Option("index") == "" {
					m.Echo(m.Spawn().Cmd("system", "tmux", "show-buffer").Result(0))
				} else {
					m.Echo(m.Spawn().Cmd("system", "tmux", "show-buffer", "-b", m.Option("index")).Result(0))
				}
			}
			return
		}},
		"sysinfo": &ctx.Command{Name: "sysinfo", Help: "sysinfo", Hand: sysinfo},
		"runtime": &ctx.Command{Name: "runtime", Help: "runtime", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			mem := &runtime.MemStats{}
			runtime.ReadMemStats(mem)
			m.Append("NumGo", runtime.NumGoroutine())
			m.Append("NumGC", mem.NumGC)
			m.Append("other", kit.FmtSize(mem.OtherSys))
			m.Append("stack", kit.FmtSize(mem.StackSys))

			m.Append("heapsys", kit.FmtSize(mem.HeapSys))
			m.Append("heapinuse", kit.FmtSize(mem.HeapInuse))
			m.Append("heapidle", kit.FmtSize(mem.HeapIdle))
			m.Append("heapalloc", kit.FmtSize(mem.HeapAlloc))

			m.Append("lookups", mem.Lookups)
			m.Append("objects", mem.HeapObjects)

			// sys := &syscall.Sysinfo_t{}
			// syscall.Sysinfo(sys)
			//
			// m.Append("total", kit.FmtSize(uint64(sys.Totalram)))
			// m.Append("free", kit.FmtSize(uint64(sys.Freeram)))
			// m.Append("mper", fmt.Sprintf("%d%%", sys.Freeram*100/sys.Totalram))
			//
			m.Table()
			return
		}},
		"windows": &ctx.Command{Name: "windows", Help: "windows", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Append("nclient", strings.Count(m.Spawn().Cmd("system", "tmux", "list-clients").Result(0), "\n"))
			m.Append("nsession", strings.Count(m.Spawn().Cmd("system", "tmux", "list-sessions").Result(0), "\n"))
			m.Append("nwindow", strings.Count(m.Spawn().Cmd("system", "tmux", "list-windows", "-a").Result(0), "\n"))
			m.Append("npane", strings.Count(m.Spawn().Cmd("system", "tmux", "list-panes", "-a").Result(0), "\n"))

			m.Append("nbuf", strings.Count(m.Spawn().Cmd("system", "tmux", "list-buffers").Result(0), "\n"))
			m.Append("ncmd", strings.Count(m.Spawn().Cmd("system", "tmux", "list-commands").Result(0), "\n"))
			m.Append("nkey", strings.Count(m.Spawn().Cmd("system", "tmux", "list-keys").Result(0), "\n"))
			m.Table()
			return
		}},

		"label": &ctx.Command{Name: "label name", Help: "记录当前脚本的位置, name: 位置名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) {
				if cli.label == nil {
					cli.label = map[string]string{}
				}
				cli.label[arg[1]] = m.Option("file_pos")
			}
			return
		}},
		"goto": &ctx.Command{Name: "goto label [exp] condition", Help: "向上跳转到指定位置, label: 跳转位置, condition: 跳转条件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) {
				if pos, ok := cli.label[arg[1]]; ok {
					if !kit.Right(arg[len(arg)-1]) {
						return
					}
					m.Append("file_pos0", pos)
				}
			}
			return
		}},
		"if": &ctx.Command{Name: "if exp", Help: "条件语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) {
				run := m.Caps("parse") && kit.Right(arg[1])
				cli.stack = append(cli.stack, &Frame{pos: m.Optioni("file_pos"), key: key, run: run})
				m.Capi("level", 1)
				m.Caps("parse", run)
			}
			return
		}},
		"else": &ctx.Command{Name: "else", Help: "条件语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) {
				if !m.Caps("parse") {
					m.Caps("parse", true)
				} else {
					if len(cli.stack) == 1 {
						m.Caps("parse", false)
					} else {
						frame := cli.stack[len(cli.stack)-2]
						m.Caps("parse", !frame.run)
					}
				}
			}
			return
		}},
		"end": &ctx.Command{Name: "end", Help: "结束语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) {
				if frame := cli.stack[len(cli.stack)-1]; frame.key == "for" && frame.run {
					m.Append("file_pos0", frame.pos)
					return
				}

				if cli.stack = cli.stack[:len(cli.stack)-1]; m.Capi("level", -1) > 0 {
					m.Caps("parse", cli.stack[len(cli.stack)-1].run)
				} else {
					m.Caps("parse", true)
				}
			}
			return
		}},
		"for": &ctx.Command{Name: "for [[express ;] condition]|[index message meta value]",
			Help: "循环语句, express: 每次循环运行的表达式, condition: 循环条件, index: 索引消息, message: 消息编号, meta: value: ",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				if cli, ok := m.Target().Server.(*CLI); m.Assert(ok) {
					run := m.Caps("parse")
					defer func() { m.Caps("parse", run) }()

					msg := m
					if run {
						if arg[1] == "index" {
							if code, e := strconv.Atoi(arg[2]); m.Assert(e) {
								msg = m.Target().Message().Tree(code)
								run = run && msg != nil && msg.Meta != nil
								switch len(arg) {
								case 4:
									run = run && len(msg.Meta) > 0
								case 5:
									run = run && len(msg.Meta[arg[3]]) > 0
								}
							}
						} else {
							run = run && kit.Right(arg[len(arg)-1])
						}

						if len(cli.stack) > 0 {
							if frame := cli.stack[len(cli.stack)-1]; frame.key == "for" && frame.pos == m.Optioni("file_pos") {
								if arg[1] == "index" {
									frame.index++
									if run = run && len(frame.list) > frame.index; run {
										if len(arg) == 5 {
											arg[3] = arg[4]
										}
										m.Cap(arg[3], frame.list[frame.index])
									}
								}
								frame.run = run
								return
							}
						}
					}

					cli.stack = append(cli.stack, &Frame{pos: m.Optioni("file_pos"), key: key, run: run, index: 0})
					if m.Capi("level", 1); run && arg[1] == "index" {
						frame := cli.stack[len(cli.stack)-1]
						switch len(arg) {
						case 4:
							frame.list = []string{}
							for k, _ := range msg.Meta {
								frame.list = append(frame.list, k)
							}
						case 5:
							frame.list = msg.Meta[arg[3]]
							arg[3] = arg[4]
						}
						m.Cap(arg[3], arg[3], frame.list[0], "临时变量")
					}
				}
				return
			}},

		"sleep": &ctx.Command{Name: "sleep time", Help: "睡眠, time(ns/us/ms/s/m/h): 时间值(纳秒/微秒/毫秒/秒/分钟/小时)", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if d, e := time.ParseDuration(arg[0]); m.Assert(e) {
				m.Log("info", "sleep %v", d)
				time.Sleep(d)
				m.Log("info", "sleep %v done", d)
			}
			return
		}},
		"time": &ctx.Command{Name: "time when [begin|end|yestoday|tommorow|monday|sunday|first|last|new|eve] [offset]",
			Help: "查看时间, when: 输入的时间戳, 剩余参数是时间偏移",
			Form: map[string]int{"time_format": 1, "time_close": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				t, stamp := time.Now(), true
				if len(arg) > 0 {
					if i, e := strconv.ParseInt(arg[0], 10, 64); e == nil {
						t, stamp, arg = time.Unix(int64(i/int64(m.Confi("time_unit"))), 0), false, arg[1:]
					} else if n, e := time.ParseInLocation(m.Confx("time_format"), arg[0], time.Local); e == nil {
						t, arg = n, arg[1:]
					} else {
						for _, v := range []string{"01-02", "2006-01-02", "15:04:05", "15:04"} {
							if n, e := time.ParseInLocation(v, arg[0], time.Local); e == nil {
								t, arg = n, arg[1:]
								break
							}
						}
					}
				}

				if len(arg) > 0 {
					switch arg[0] {
					case "begin":
						d, e := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", t.Hour(), t.Minute(), t.Second()))
						m.Assert(e)
						t, arg = t.Add(-d), arg[1:]
					case "end":
						d, e := time.ParseDuration(fmt.Sprintf("%dh%dm%ds%dns", t.Hour(), t.Minute(), t.Second(), t.Nanosecond()))
						m.Assert(e)
						t, arg = t.Add(time.Duration(24*time.Hour)-d), arg[1:]
						if m.Confx("time_close") == "close" {
							t = t.Add(-time.Second)
						}
					case "yestoday":
						t, arg = t.Add(-time.Duration(24*time.Hour)), arg[1:]
					case "tomorrow":
						t, arg = t.Add(time.Duration(24*time.Hour)), arg[1:]
					case "monday":
						d, e := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", int((t.Weekday()-time.Monday+7)%7)*24+t.Hour(), t.Minute(), t.Second()))
						m.Assert(e)
						t, arg = t.Add(-d), arg[1:]
					case "sunday":
						d, e := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", int((t.Weekday()-time.Monday+7)%7)*24+t.Hour(), t.Minute(), t.Second()))
						m.Assert(e)
						t, arg = t.Add(time.Duration(7*24*time.Hour)-d), arg[1:]
						if m.Confx("time_close") == "close" {
							t = t.Add(-time.Second)
						}
					case "first":
						t, arg = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local), arg[1:]
					case "last":
						month, year := t.Month()+1, t.Year()
						if month >= 13 {
							month, year = 1, year+1
						}
						t, arg = time.Date(year, month, 1, 0, 0, 0, 0, time.Local), arg[1:]
						if m.Confx("time_close") == "close" {
							t = t.Add(-time.Second)
						}
					case "new":
						t, arg = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.Local), arg[1:]
					case "eve":
						t, arg = time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, time.Local), arg[1:]
						if m.Confx("time_close") == "close" {
							t = t.Add(-time.Second)
						}
					case "":
						arg = arg[1:]
					}
				}

				if len(arg) > 0 {
					if d, e := time.ParseDuration(arg[0]); e == nil {
						t, arg = t.Add(d), arg[1:]
					}
				}

				m.Append("datetime", t.Format(m.Confx("time_format")))
				m.Append("timestamp", t.Unix()*int64(m.Confi("time_unit")))

				if stamp {
					m.Echo("%d", t.Unix()*int64(m.Confi("time_unit")))
				} else {
					m.Echo(t.Format(m.Confx("time_format")))
				}
				return
			}},
		"timer": &ctx.Command{Name: "timer [begin time] [repeat] [order time] time cmd", Help: "定时任务", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if cli, ok := c.Server.(*CLI); m.Assert(ok) {
				if len(arg) == 0 { // 查看任务列表
					for k, v := range m.Confv("timer").(map[string]interface{}) {
						val := v.(map[string]interface{})
						m.Add("append", "key", k)
						m.Add("append", "action_time", time.Unix(0, val["action_time"].(int64)/int64(m.Confi("time_unit"))*1000000000).Format(m.Conf("time_format")))
						m.Add("append", "order", val["order"])
						m.Add("append", "time", val["time"])
						m.Add("append", "cmd", val["cmd"])
						m.Add("append", "msg", val["msg"])
						m.Add("append", "results", fmt.Sprintf("%v", val["result"]))
					}
					m.Table()
					return
				}
				switch arg[0] {
				case "stop":
					if timer := m.Confm("timer", arg[1]); timer != nil {
						timer["stop"] = true
					}
					cli.schedule(m)
					return
				case "start":
					if timer := m.Confm("timer", arg[1]); timer != nil {
						timer["stop"] = false
					}
					cli.schedule(m)
					return
				case "delete":
					delete(m.Confm("timer"), arg[1])
					cli.schedule(m)
					return
				}

				now := int64(m.Sess("cli").Cmd("time").Appendi("timestamp"))
				begin := now
				if len(arg) > 0 && arg[0] == "begin" {
					begin, arg = int64(m.Sess("cli").Cmd("time", arg[1]).Appendi("timestamp")), arg[2:]
				}

				repeat := false
				if len(arg) > 0 && arg[0] == "repeat" {
					repeat, arg = true, arg[1:]
				}

				order := ""
				if len(arg) > 0 && arg[0] == "order" {
					order, arg = arg[1], arg[2:]
				}

				action := int64(m.Sess("cli").Cmd("time", begin, order, arg[0]).Appendi("timestamp"))

				// 创建任务
				hash := m.Sess("aaa").Cmd("hash", "timer", arg, "time", "rand").Result(0)
				m.Confv("timer", hash, map[string]interface{}{
					"create_time": now,
					"begin_time":  begin,
					"action_time": action,
					"repeat":      repeat,
					"order":       order,
					"done":        false,
					"stop":        false,
					"time":        arg[0],
					"cmd":         arg[1:],
					"msg":         0,
					"result":      "",
				})

				if cli.Timer == nil { // 创建时间队列
					cli.Timer = time.NewTimer((time.Duration)((action - now) / int64(m.Confi("time_unit")) * 1000000000))
					m.GoLoop(m, func(m *ctx.Message) {
						select {
						case <-cli.Timer.C:
							m.Log("info", "timer %s", m.Conf("timer_next"))
							if m.Conf("timer_next") == "" {
								break
							}

							if timer := m.Confm("timer", m.Conf("timer_next")); timer != nil && !kit.Right(timer["stop"]) {
								m.Log("info", "timer %s %v", m.Conf("timer_next"), timer["cmd"])

								msg := m.Sess("cli").Cmd("source", timer["cmd"])
								timer["result"] = msg.Meta["result"]
								timer["msg"] = msg.Code()

								if timer["repeat"].(bool) {
									timer["action_time"] = int64(m.Sess("cli").Cmd("time", timer["action_time"], timer["order"], timer["time"]).Appendi("timestamp"))
								} else {
									timer["done"] = true
								}
							}

							cli.schedule(m)
						}
					})
				}

				// 调度任务
				cli.schedule(m)
				m.Echo(hash)
			}
			return
		}},
	},
}

func init() {
	cli := &CLI{}
	cli.Context = Index
	ctx.Index.Register(Index, cli)
}
