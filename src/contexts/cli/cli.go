package cli

import (
	"contexts/ctx"
	"toolkit"

	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"plugin"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type CLI struct {
	*time.Timer
	Context *ctx.Context
}

func (cli *CLI) schedule(m *ctx.Message) string {
	first, timer := "", int64(1<<50)
	for k, v := range m.Confv("timer", "list").(map[string]interface{}) {
		val := v.(map[string]interface{})
		if val["action_time"].(int64) < timer && !val["done"].(bool) {
			first, timer = k, val["action_time"].(int64)
		}
	}
	cli.Timer.Reset(time.Until(time.Unix(0, timer/int64(m.Confi("time", "unit"))*1000000000)))
	return m.Conf("timer", "next", first)
}

func (cli *CLI) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Configs["_index"] = &ctx.Config{Name: "_index", Value: []interface{}{}, Help: "_index"}
	return &CLI{Context: c}
}
func (cli *CLI) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return cli
}
func (cli *CLI) Start(m *ctx.Message, arg ...string) bool {
	return false
}
func (cli *CLI) Close(m *ctx.Message, arg ...string) bool {
	return true
}

var Index = &ctx.Context{Name: "cli", Help: "管理中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"runtime": &ctx.Config{Name: "runtime", Value: map[string]interface{}{
			"init": []interface{}{
				"ctx_log", "ctx_app", "ctx_bin",
				"ctx_ups", "ctx_box", "ctx_dev",
				"ctx_cas",
				"ctx_root", "ctx_home",
				"ctx_type",
				"web_port", "ssh_port",
			},
			"boot": map[string]interface{}{
				"web_port": ":9095",
				"ssh_port": ":9090",
				"username": "shy",
			},
		}, Help: "运行环境, host, init, boot, node, user, work"},
		"system": &ctx.Config{Name: "system", Value: map[string]interface{}{
			"timeout": "60s",
			"env":     map[string]interface{}{},
			"shell": map[string]interface{}{
				"sh":  map[string]interface{}{"cmd": "bash"},
				"py":  map[string]interface{}{"cmd": "python"},
				"vi":  map[string]interface{}{"active": true},
				"top": map[string]interface{}{"active": true},
				"ls":  map[string]interface{}{"arg": []interface{}{"-l"}},
			},
			"script": map[string]interface{}{
				"sh": "bash", "shy": "source", "py": "python",
				"init": "etc/init.shy", "exit": "etc/exit.shy",
			},
		}, Help: "系统环境, shell: path, cmd, arg, dir, env, active, daemon; "},
		"daemon": &ctx.Config{Name: "daemon", Value: map[string]interface{}{}, Help: "守护任务"},
		"timer": &ctx.Config{Name: "timer", Value: map[string]interface{}{
			"list": map[string]interface{}{}, "next": "",
		}, Help: "定时器"},

		"project": &ctx.Config{Name: "project", Value: map[string]interface{}{
			"github":  "https://github.com/shylinux/context",
			"goproxy": "https://goproxy.cn",
			"plugin": map[string]interface{}{
				"path": "src/plugin", "list": Template,
			}, "script": map[string]interface{}{
				"path": "usr/script",
			}, "trash": map[string]interface{}{
				"path": "usr/trash",
			},
		}, Help: "项目管理"},
		"compile": &ctx.Config{Name: "compile", Value: map[string]interface{}{
			"name": "bench", "bench": "src/extend/bench.go", "list": []interface{}{
				map[string]interface{}{"os": "linux", "cpu": "arm"},
				map[string]interface{}{"os": "linux", "cpu": "386"},
				map[string]interface{}{"os": "linux", "cpu": "amd64"},
				map[string]interface{}{"os": "windows", "cpu": "386"},
				map[string]interface{}{"os": "windows", "cpu": "amd64"},
				map[string]interface{}{"os": "darwin", "cpu": "amd64"},
			}, "tmp": "var/tmp/go", "dep": []interface{}{
				"github.com/nfs/termbox-go",
				"github.com/go-sql-driver/mysql",
				"github.com/redigo/redis",
				"github.com/gomarkdown/markdown",
				"github.com/skip2/go-qrcode",
				"gopkg.in//gomail.v2",
			},
		}, Help: "源码编译"},
		"publish": &ctx.Config{Name: "publish", Value: map[string]interface{}{
			"path": "usr/publish", "list": map[string]interface{}{
				"boot_sh": "bin/boot.sh",
				"zone_sh": "bin/zone.sh",
				"user_sh": "bin/user.sh",
				"node_sh": "bin/node.sh",

				"init_shy":   "etc/init.shy",
				"common_shy": "etc/common.shy",
				"exit_shy":   "etc/exit.shy",

				"template_tar_gz": "usr/template",
				"librarys_tar_gz": "usr/librarys",
			},
			"script": map[string]interface{}{
				"worker": "usr/publish/hello/local.shy",
				"server": "usr/publish/docker/local.shy",
			},
		}, Help: "版本发布"},
		"upgrade": &ctx.Config{Name: "upgrade", Value: map[string]interface{}{
			"install": []interface{}{"context", "tmux", "mind", "love"},
			"system":  []interface{}{"boot.sh", "zone.sh", "user.sh", "node.sh", "init.shy", "common.shy", "exit.shy"},
			"portal":  []interface{}{"template.tar.gz", "librarys.tar.gz"},
			"script":  []interface{}{"test.php"},
			"list": map[string]interface{}{
				"bench": "bin/bench.new",

				"boot_sh": "bin/boot.sh",
				"zone_sh": "bin/zone.sh",
				"user_sh": "bin/user.sh",
				"node_sh": "bin/node.sh",

				"init_shy":   "etc/init.shy",
				"common_shy": "etc/common.shy",
				"exit_shy":   "etc/exit.shy",

				"template_tar_gz": "usr/template.tar.gz",
				"librarys_tar_gz": "usr/librarys.tar.gz",
			},
		}, Help: "服务升级"},
		"missyou": &ctx.Config{Name: "missyou", Value: map[string]interface{}{
			"path": "usr/local/work",
		}, Help: "任务管理"},
	},
	Commands: map[string]*ctx.Command{
		"_init": &ctx.Command{Name: "_init", Help: "环境初始化", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Conf("runtime", "host.GOARCH", runtime.GOARCH)
			m.Conf("runtime", "host.GOOS", runtime.GOOS)
			m.Conf("runtime", "host.pid", os.Getpid())

			m.Confm("runtime", "init", func(index int, key string) {
				if value := os.Getenv(key); value != "" {
					m.Conf("runtime", "boot."+key, kit.Select("", value, value != "-"))
				}
			})

			if name, e := os.Hostname(); e == nil {
				m.Conf("runtime", "boot.hostname", kit.Select(name, os.Getenv("HOSTNAME")))
			}
			if user, e := user.Current(); e == nil {
				ns := strings.Split(user.Username, "\\")
				name := ns[len(ns)-1]
				m.Conf("runtime", "boot.username", kit.Select(name, os.Getenv("USER")))
			}
			if name, e := os.Getwd(); e == nil {
				_, file := path.Split(kit.Select(name, os.Getenv("PWD")))
				ns := strings.Split(file, "\\")
				m.Conf("runtime", "boot.pathname", ns[len(ns)-1])
				m.Conf("runtime", "boot.ctx_path", name)
			}
			return
		}},
		"runtime": &ctx.Command{Name: "runtime [host|boot|node|user|work [name [value]]]", Help: "运行环境", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmdy("ctx.config", "runtime")
				return
			}

			switch arg[0] {
			case "system":
				mem := &runtime.MemStats{}
				runtime.ReadMemStats(mem)
				m.Append("NumCPU", runtime.NumCPU())
				m.Append("NumGo", runtime.NumGoroutine())
				m.Append("NumGC", mem.NumGC)
				m.Append("other", kit.FmtSize(int64(mem.OtherSys)))
				m.Append("stack", kit.FmtSize(int64(mem.StackSys)))
				m.Append("heapsys", kit.FmtSize(int64(mem.HeapSys)))
				m.Append("heapidle", kit.FmtSize(int64(mem.HeapIdle)))
				m.Append("heapinuse", kit.FmtSize(int64(mem.HeapInuse)))
				m.Append("heapalloc", kit.FmtSize(int64(mem.HeapAlloc)))
				m.Append("objects", mem.HeapObjects)
				m.Append("lookups", mem.Lookups)
				m.Table()

			default:
				if len(arg) == 1 {
					m.Cmdy("ctx.config", "runtime", arg[0])
					return
				}

				m.Conf("runtime", arg[0], arg[1])
				if arg[0] == "node.route" && m.Confs("runtime", "work.serve") {
					m.Conf("runtime", "work.route", arg[1])
					return
				}
				m.Echo(arg[1])
			}
			return
		}},
		"system": &ctx.Command{Name: "system word...", Help: []string{"调用系统命令, word: 命令",
			"cmd_timeout: 命令超时",
			"cmd_active(true/false): 是否交互",
			"cmd_daemon(true/false): 是否守护",
			"cmd_dir: 工作目录",
			"cmd_env: 环境变量",
			"cmd_log: 输出日志",
			"cmd_temp: 缓存结果",
			"cmd_parse: 解析结果",
			"cmd_error: 输出错误",
		}, Form: map[string]int{
			"cmd_timeout": 1,
			"cmd_active":  1,
			"cmd_daemon":  1,
			"cmd_dir":     1,
			"cmd_env":     2,
			"cmd_log":     1,
			"cmd_temp":    -1,
			"cmd_parse":   3,
			"cmd_error":   0,
			"cmd_select":  -1,
			"app_log":     1,
		}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 管道参数
			for _, v := range m.Meta["result"] {
				if strings.TrimSpace(v) != "" {
					arg = append(arg, v)
				}
			}

			// 命令配置
			conf := m.Confm("system", []string{"shell", arg[0]})
			if as := strings.Split(arg[0], "."); conf == nil && len(as) > 0 {
				if conf = m.Confm("system", []string{"shell", as[len(as)-1]}); conf != nil {
					arg = append([]string{kit.Format(kit.Chain(conf, "cmd"))}, arg...)
				}
			}

			// 命令替换
			args := []string{kit.Select(arg[0], kit.Format(kit.Chain(conf, "cmd")))}
			if list, ok := kit.Chain(conf, "arg").([]interface{}); ok {
				for _, v := range list {
					args = append(args, m.Parse(v))
				}
			}

			// 命令解析
			args = append(args, arg[1:]...)
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Path = kit.Select(cmd.Path, kit.Format(kit.Chain(conf, "path")))
			if cmd.Path != "" || len(cmd.Args) > 0 {
				m.Log("info", "cmd %v %v", cmd.Path, cmd.Args)
			}

			// 工作目录
			cmd.Dir = kit.Select(kit.Chains(conf, "dir"), m.Option("cmd_dir"))
			if cmd.Dir != "" {
				m.Log("info", "dir %v", cmd.Dir)
			}

			// 环境变量
			m.Confm("system", "env", func(key string, value string) {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, m.Parse(value)))
			})
			kit.Structm(kit.Chain(conf, "env"), func(key string, value string) {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, m.Parse(value)))
			})
			for i := 0; i < len(m.Meta["cmd_env"])-1; i += 2 {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", m.Meta["cmd_env"][i], m.Parse(m.Meta["cmd_env"][i+1])))
			}
			if len(cmd.Env) > 0 {
				m.Log("info", "env %v", cmd.Env)
			}

			// 交互命令
			if m.Options("cmd_active") || kit.Right(conf["active"]) {
				cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
				if e := cmd.Start(); e != nil {
					m.Echo("error: ").Echo("%s\n", e)
				} else if e := cmd.Wait(); e != nil {
					m.Echo("error: ").Echo("%s\n", e)
				}
				return
			}

			// 守护命令
			if m.Options("cmd_daemon") || kit.Right(conf["daemon"]) {
				// 创建日志
				if m.Options("cmd_log") {
					log := fmt.Sprintf("var/log/%s.log", m.Option("cmd_log"))
					if e := os.Rename(log, fmt.Sprintf("var/log/%s_%s.log", m.Option("cmd_log"), m.Time("2006_0102_1504"))); e != nil {
						m.Log("info", "mv %s error %s", log, e)
					}
					if l, e := os.Create(log); m.Assert(e) {
						cmd.Stdout = l
					}

					err := fmt.Sprintf("var/log/%s.err", m.Option("cmd_log"))
					if e := os.Rename(err, fmt.Sprintf("var/log/%s_%s.err", m.Option("cmd_log"), m.Time("2006_0102_1504"))); e != nil {
						m.Log("info", "mv %s error %s", err, e)
					}
					if l, e := os.Create(err); m.Assert(e) {
						cmd.Stderr = l
					}
				}

				// 守护列表
				h, _ := kit.Hash("uniq")
				m.Conf("daemon", h, map[string]interface{}{
					"create_time": m.Time(), "log": kit.Select(m.Option("cmd_log"), m.Option("app_log")), "sub": cmd,
				})
				m.Echo(h)

				// 执行命令
				m.Gos(m, func(m *ctx.Message) {
					if e := cmd.Start(); e != nil {
						m.Log("warn", "%v", e).Echo("error: ").Echo("%s\n", e)
					} else if e := cmd.Wait(); e != nil {
						m.Log("warn", "%v", e).Echo("error: ").Echo("%s\n", e)
					}
					m.Conf("daemon", []string{h, "finish_time"}, time.Now().Format(m.Conf("time", "format")))
				})
				return e
			}

			// 管道命令
			wait := make(chan bool, 1)
			m.Gos(m, func(m *ctx.Message) {
				defer func() { wait <- true }()

				out := bytes.NewBuffer(make([]byte, 0, 1024))
				err := bytes.NewBuffer(make([]byte, 0, 1024))
				cmd.Stdout = out
				cmd.Stderr = err

				// 运行命令
				if e := cmd.Run(); e != nil {
					m.Echo("error: ").Echo(kit.Select(e.Error(), err.String()))
					return
				}

				// 输出错误
				if m.Has("cmd_error") {
					m.Echo(err.String())
				}

				// 缓存结果
				if m.Options("cmd_temp") {
					m.Put("option", "data", out.String()).Cmdy("mdb.temp", "script", strings.Join(arg, " "), "data", "data", m.Meta["cmd_temp"])
					return
				}

				// 解析结果
				switch m.Option("cmd_parse") {
				case "format":
					var data interface{}
					if json.Unmarshal(out.Bytes(), &data) == nil {
						if b, e := json.MarshalIndent(data, "", "  "); e == nil {
							m.Echo(string(b))
							break
						}
					}
					m.Echo(out.String())
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

				case "cli":
					read := csv.NewReader(out)
					read.Comma = ' '
					read.TrimLeadingSpace = true
					read.FieldsPerRecord = 4
					data, e := read.ReadAll()
					m.Assert(e)
					for i := 1; i < len(data); i++ {
						for j := 0; j < len(data[i]); j++ {
							m.Add("append", data[0][j], data[i][j])
						}
					}
					m.Table()

				case "cut":
					bio := bufio.NewScanner(out)
					bio.Scan()
					c := byte(kit.Select(" ", m.Optionv("cmd_parse"), 2)[0])
					for heads := kit.Split(bio.Text(), c, kit.Int(kit.Select("-1", m.Optionv("cmd_parse"), 1))); bio.Scan(); {
						for i, v := range kit.Split(bio.Text(), c, len(heads)) {
							m.Add("append", heads[i], v)
						}
					}
					m.Table()

				default:
					var data interface{}
					if json.Unmarshal(out.Bytes(), &data) == nil {
						if b, e := json.MarshalIndent(data, "", "  "); e == nil {
							m.Echo(string(b))
							break
						}
					}
					m.Echo(out.String())
				}
				if m.Has("cmd_select") {
					m.Cmd("select", m.Meta["cmd_select"])
				}
			})

			// 命令超时
			timeout := kit.Duration(kit.Select(m.Conf("system", "timeout"), kit.Select(kit.Chains(conf, "timeout"), m.Option("cmd_timeout"))))
			select {
			case <-time.After(timeout):
				m.Log("warn", "%s: %v timeout", arg[0], timeout)
				m.Echo("%s: %v timeout", arg[0], timeout)
				cmd.Process.Kill()
			case <-wait:
			}
			return
		}},
		"daemon": &ctx.Command{Name: "daemon", Help: "守护任务", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			pid := ""
			if len(arg) > 0 && m.Confs("daemon", arg[0]) {
				pid, arg = arg[0], arg[1:]
			}

			if len(arg) == 0 {
				m.Confm("daemon", func(key string, info map[string]interface{}) {
					if pid != "" && key != pid {
						return
					}

					m.Add("append", "key", key)
					m.Add("append", "log", info["log"])
					m.Add("append", "create_time", info["create_time"])
					m.Add("append", "finish_time", info["finish_time"])

					if cmd, ok := info["sub"].(*exec.Cmd); ok {
						info["pid"] = cmd.Process.Pid
						info["cmd"] = kit.Select(cmd.Args[0], cmd.Args, 1)
						if cmd.ProcessState != nil {
							info["str"] = cmd.ProcessState.String()
						}
					}
					m.Add("append", "pid", kit.Format(info["pid"]))
					m.Add("append", "cmd", kit.Format(info["cmd"]))
					m.Add("append", "str", kit.Format(info["str"]))
				})
				m.Table()
				return
			}

			if pid != "" {
				if cmd, ok := m.Confm("daemon", pid)["sub"].(*exec.Cmd); ok {
					switch arg[0] {
					case "stop":
						m.Echo("%s", cmd.Process.Signal(os.Interrupt))
					default:
						m.Echo("%v", cmd)
					}
				}
				return
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
		"timer": &ctx.Command{Name: "timer [begin time] [repeat] [order time] time cmd", Help: "定时任务", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if cli, ok := c.Server.(*CLI); m.Assert(ok) {
				// 定时列表
				if len(arg) == 0 {
					m.Confm("timer", "list", func(key string, timer map[string]interface{}) {
						m.Add("append", "key", key)
						m.Add("append", "action_time", time.Unix(0, timer["action_time"].(int64)/int64(m.Confi("time", "unit"))*1000000000).Format(m.Conf("time", "format")))
						m.Add("append", "order", timer["order"])
						m.Add("append", "time", timer["time"])
						m.Add("append", "cmd", timer["cmd"])
						m.Add("append", "msg", timer["msg"])
						m.Add("append", "results", kit.Format(timer["result"]))
					})
					m.Table()
					return
				}

				switch arg[0] {
				case "stop":
					if timer := m.Confm("timer", "list", arg[1]); timer != nil {
						timer["stop"] = true
					}
					cli.schedule(m)
					return
				case "start":
					if timer := m.Confm("timer", "list", arg[1]); timer != nil {
						timer["stop"] = false
					}
					cli.schedule(m)
					return
				case "delete":
					delete(m.Confm("timer", "list"), arg[1])
					cli.schedule(m)
					return
				}

				now := m.Sess("cli").Cmd("time").Appendi("timestamp")
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
				m.Confv("timer", []string{"list", hash}, map[string]interface{}{
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
					cli.Timer = time.NewTimer((time.Duration)((action - now) / int64(m.Confi("time", "unit")) * 1000000000))
					m.GoLoop(m, func(m *ctx.Message) {
						select {
						case <-cli.Timer.C:
							if m.Conf("timer", "next") == "" {
								break
							}

							if timer := m.Confm("timer", []string{"list", m.Conf("timer", "next")}); timer != nil && !kit.Right(timer["stop"]) {
								m.Log("info", "timer %s %v", m.Conf("timer", "next"), timer["cmd"])

								msg := m.Cmd("nfs.source", timer["cmd"])
								timer["result"] = msg.Meta["result"]
								timer["msg"] = msg.Code()

								if timer["repeat"].(bool) {
									timer["action_time"] = int64(m.Sess("cli").Cmd("time", msg.Time(), timer["order"], timer["time"]).Appendi("timestamp"))
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
		"time": &ctx.Command{Name: "time when [begin|end|yestoday|tommorow|monday|sunday|first|last|new|eve] [offset]",
			Help: "查看时间, when: 输入的时间戳, 剩余参数是时间偏移",
			Form: map[string]int{"time_format": 1, "time_close": 1},
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				format := kit.Select(m.Conf("time", "format"), m.Option("time_format"))
				t, stamp := time.Now(), true
				if len(arg) > 0 {
					if arg[0] == "show" {
						stamp = false
					} else if i, e := strconv.ParseInt(arg[0], 10, 64); e == nil {
						t, stamp, arg = time.Unix(int64(i/int64(m.Confi("time", "unit"))), 0), false, arg[1:]
					} else if n, e := time.ParseInLocation(format, arg[0], time.Local); e == nil {
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
						if kit.Select(m.Conf("time", "close"), m.Option("time_close")) == "close" {
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
						if kit.Select(m.Conf("time", "close"), m.Option("time_close")) == "close" {
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
						if kit.Select(m.Conf("time", "close"), m.Option("time_close")) == "close" {
							t = t.Add(-time.Second)
						}
					case "new":
						t, arg = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.Local), arg[1:]
					case "eve":
						t, arg = time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, time.Local), arg[1:]
						if kit.Select(m.Conf("time", "close"), m.Option("time_close")) == "close" {
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

				m.Append("datetime", t.Format(format))
				m.Append("timestamp", t.Unix()*int64(m.Confi("time", "unit")))

				if stamp {
					m.Echo("%d", t.Unix()*int64(m.Confi("time", "unit")))
				} else {
					m.Echo(t.Format(format))
				}
				return
			}},
		"proc": &ctx.Command{Name: "proc", Help: "进程管理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Cmdy("cli.system", "ps", kit.Select("ax", arg, 0), "cmd_parse", "cut")
			if len(arg) > 1 {
				m.Cmd("select", "reg", "COMMAND", arg[1])
			}
			return
		}},
		"quit": &ctx.Command{Name: "quit code", Help: "停止服务", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Conf("runtime", "boot.count", m.Confi("runtime", "boot.count")+1)
			code := kit.Select("0", arg, 0)
			switch code {
			case "0":
				m.Cmd("nfs.source", m.Conf("system", "script.exit"))
				m.Echo("quit")

			case "1":
				if m.Option("bio.modal") != "action" {
					m.Cmd("nfs.source", m.Conf("system", "script.exit"))
					m.Echo("restart")
				}

			case "2":
				m.Echo("term")
			}

			m.Append("time", m.Time())
			m.Append("code", code)
			m.Echo(", wait 1s\n").Table()
			m.Gos(m, func(m *ctx.Message) {
				m.Cmd("ctx._exit")
				time.Sleep(time.Second)
				os.Exit(kit.Int(code))
			})
			return
		}},
		"_exit": &ctx.Command{Name: "_exit", Help: "退出命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Confm("daemon", func(key string, info map[string]interface{}) {
				m.Cmd("cli.daemon", key, "stop")
			})
			return
		}},

		"project": &ctx.Command{Name: "project init|stat|trend|submit|review|plugin [args...]", Help: "项目管理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch arg[0] {
			case "init":
				if _, e := os.Stat(".git"); e == nil {
					// 更新代码
					m.Cmdp(0, []string{"git update"}, []string{"cli.system", "git"}, [][]string{
						[]string{"stash"}, []string{"pull"}, []string{"stash", "pop"},
					})

					// 代码状态
					m.Cmdy("cli.system", "git", "status", "-sb", "cmd_parse", "cut")
					return e
				}

				// 创建项目
				m.Cmdp(0, []string{"git init"}, []string{"cli.system", "git"}, [][]string{
					[]string{"init"}, []string{"remote", "add", kit.Select("origin", arg, 1), kit.Select(m.Conf("project", "github"), arg, 2)},
					[]string{"stash"}, []string{"pull"}, []string{"checkout", "-f", "master"}, []string{"stash", "pop"},
				})

				// 下载依赖
				list := [][]string{}
				m.Confm("compile", "dep", func(index int, value string) {
					list = append(list, []string{value})
				})
				m.Cmdp(0, []string{"go build"}, []string{"cli.system", "go", "get"}, list)

			case "stats":
				// 代码统计
				m.Cmdy("nfs.dir", kit.Select("src", arg, 1), "dir_deep", "dir_type", "file", "dir_sort", "line", "int_r", "dir_select", "group", "")

			case "stat":
				// 代码统计
				m.Cmdy("nfs.dir", "src", "dir_deep", "dir_type", "file", "dir_sort", "line", "int_r")

			case "trend":
				// 提交记录
				m.Cmdy("nfs.git", "sum", "-n", kit.Select("20", arg, 1))

			case "submit":
				// 提交代码
				if len(arg) > 1 {
					m.Cmdp(0, []string{"git submit"}, []string{"cli.system", "git"}, [][]string{
						[]string{"commit", "-am", arg[1]}, []string{"push"},
					})
				}
				// 提交记录
				m.Cmdy("nfs.git", "logs", "table.limit", "3")

			case "review":
				// 代码修改
				m.Cmdy("cli.system", "git", "diff")

			case "plugin":
				// 查看插件
				if arg = arg[1:]; len(arg) == 0 {
					m.Cmdy("nfs.dir", m.Conf("project", "plugin.path"), "time", "line", "name")
					break
				}
				fallthrough
			default:
				// 创建插件
				p := path.Join(m.Conf("project", "plugin.path"), arg[0])
				if _, e := os.Stat(p); os.IsNotExist(e) && m.Assert(os.MkdirAll(p, 0777)) {
					m.Confm("project", "plugin.list", func(index int, value map[string]interface{}) {
						ioutil.WriteFile(path.Join(p, kit.Format(value["name"])), []byte(kit.Format(value["text"])), 0666)
					})
				}
				// 插件列表
				m.Cmdy("nfs.dir", p, "time", "line", "hashs", "path")
			}
			return
		}},
		"compile": &ctx.Command{Name: "compile all|self|linux|windows|darwin|restart|plugin", Help: "项目编译", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch arg[0] {
			case "all":
				// 所有版本
				list := [][]string{}
				m.Confm("compile", "list", func(index int, value map[string]interface{}) {
					list = append(list, []string{kit.Format(value["os"]), kit.Format(value["cpu"])})
				})
				m.Cmdp(0, []string{"go build"}, []string{"cli.compile"}, list)

			case "self":
				// 编译版本
				m.Cmd("cli.version", "create")

				// 编译项目
				if m.Cmdy("cli.compile", ""); m.Has("bin") {
					target := path.Join(kit.Select(os.Getenv("GOBIN"), ""), m.Conf("compile", "name"))
					os.Remove(target)
					m.Append("bin", m.Cmdx("nfs.copy", target, m.Append("bin")))
					os.Chmod(target, 0777)
					m.Cmd("cli.quit", 1)
				}

			case "restart":
				// 重启项目
				m.Cmdy("cli.quit", "1")

			case "plugin":
				// 插件列表
				if arg = arg[1:]; len(arg) == 0 {
					m.Cmdy("nfs.dir", m.Conf("publish", "path"), "dir_deep", "dir_reg", ".*\\.so", "time", "size", "hashs", "path")
					break
				}
				fallthrough
			default:
				// 编译插件
				p, q, o := path.Join(m.Conf("project", "plugin.path"), arg[0], "index.go"), "", []string{}
				if _, e := os.Stat(p); e == nil {
					q = path.Join(m.Conf("publish", "path"), arg[0], "index.so")
					o = append(o, "-buildmode=plugin")
					arg = arg[1:]
				}

				// 目标系统
				goos := kit.Select(m.Conf("runtime", "host.GOOS"), arg, 0)
				arch := kit.Select(m.Conf("runtime", "host.GOARCH"), arg, 1)

				// 编译环境
				tmp := path.Join(kit.Pwd(), m.Conf("compile", "tmp"))
				os.MkdirAll(m.Conf("compile", "tmp"), 0777)
				env := []string{
					"cmd_env", "GOOS", goos, "cmd_env", "GOARCH", arch,
					"cmd_env", "GOTMPDIR", tmp, "cmd_env", "GOCACHE", tmp,
					"cmd_env", "GOPATH", m.Conf("runtime", "boot.ctx_home") + string(os.PathListSeparator) + os.Getenv("GOPATH"),
					"cmd_env", "PATH", os.Getenv("PATH"),
				}

				// 编译目标
				if q == "" {
					q = path.Join(m.Conf("publish", "path"), strings.Join([]string{m.Conf("compile", "name"), goos, arch}, "."))
					p = m.Cmdx("nfs.path", m.Conf("compile", "bench"))
				}

				// 编译项目
				if m.Cmdy("cli.system", env, "go", "build", o, "-o", q, p); m.Result(0) == "" {
					m.Append("time", m.Time())
					m.Append("cost", m.Format("cost"))
					m.Append("hash", m.Cmdx("nfs.hash", q)[:8])
					m.Append("bin", q)
					m.Table()
				}
			}
			return
		}},
		"publish": &ctx.Command{Name: "publish [args...]", Help: "项目发布", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Confm("publish", "list", func(key string, value string) {
					arg = append(arg, strings.Replace(key, "_", ".", -1))
				})
			}

			// 发布插件
			p := path.Join(m.Conf("project", "plugin.path"), arg[0])
			q := path.Join(m.Conf("publish", "path"), arg[0])
			if _, e := os.Stat(p); e == nil && m.Assert(os.MkdirAll(q, 0777)) {
				m.Confm("project", "plugin.list", func(index int, value map[string]interface{}) {
					m.Cmd("nfs.copy", path.Join(q, kit.Format(value["name"])), path.Join(p, kit.Format(value["name"])))
				})
				for _, v := range arg[1:] {
					m.Cmd("nfs.copy", path.Join(q, v), path.Join(p, v))
				}
				m.Cmd("cli.system", "tar", "-zcf", q+".tar.gz", "-C", m.Conf("publish", "path"), arg[0])
				m.Cmdy("nfs.dir", q, "time", "size", "hashs", "path", "dir_sort", "path", "str")
				return e
			}

			// 发布文件
			for _, key := range arg {
				p := m.Cmdx("nfs.path", m.Conf("publish", []string{"list", kit.Key(key)}))
				q := path.Join(m.Conf("publish", "path"), key)
				if s, e := os.Stat(p); e == nil {
					if s.IsDir() {
						m.Cmd("cli.system", "tar", "-zcf", q, "-C", path.Dir(p), path.Base(p))
					} else {
						m.Cmd("nfs.copy", q, p)
					}

					m.Push("time", s.ModTime())
					m.Push("size", s.Size())
					m.Push("hash", kit.Hashs(p)[:8])
					m.Push("file", q)
				}
			}
			m.Sort("file").Table()
			return
		}},
		"upgrade": &ctx.Command{Name: "upgrade project|bench|system|portal|plugin|restart|script", Help: "服务升级", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmdy("ctx.config", "upgrade")
				return
			}

			switch arg[0] {
			case "install":
				m.Cmd("cli.upgrade", "system")
				m.Cmd("cli.upgrade", "portal")
				m.Confm("upgrade", "install", func(index int, value string) {
					m.Cmd("cli.upgrade", "package", value)
				})

			case "project":
				m.Cmd("cli.project", "init")
				m.Cmd("cli.compile", "all")
				m.Cmd("cli.publish")

			case "script":
				// 脚本列表
				if len(arg) == 1 {
					m.Cmdy("nfs.dir", m.Conf("project", "script.path"), "time", "line", "hashs", "path")
					break
				}

				// 局部脚本
				miss := ""
				if len(arg) > 2 {
					miss, arg = arg[1], arg[1:]
				}
				// 下载脚本
				for _, v := range arg[1:] {
					m.Cmdy("web.get", "dev", fmt.Sprintf("publish/%s", v), "upgrade", "script",
						"missyou", miss, "save", path.Join(m.Conf("project", "script.path"), v))
				}

			case "restart":
				m.Cmdy("cli.quit", "1")

			case "package":
				name := arg[1] + ".tar.gz"
				p := path.Join(m.Conf("publish", "path"), name)

				m.Cmd("web.get", "dev", fmt.Sprintf("publish/%s", name), "save", p,
					"GOARCH", m.Conf("runtime", "host.GOARCH"), "GOOS", m.Conf("runtime", "host.GOOS"))

				m.Cmd("cli.system", "tar", "-xvf", p, "-C", path.Dir(p))

			case "plugin":
				// 模块列表
				if arg = arg[1:]; len(arg) == 0 {
					m.Cmdy("cli.context")
					break
				}

				// 加载模块
				if p, e := plugin.Open(path.Join(m.Conf("publish", "path"), arg[0], "index.so")); e == nil {
					if s, e := p.Lookup("Index"); m.Assert(e) {
						m.Spawn(c.Register(*(s.(**ctx.Context)), nil, arg[0]).Begin(m, arg[1:]...)).Cmd("_init", arg[1:])
					}
				}

				// 查找模块
				msg := m.Find(arg[0], false)
				if msg == nil {
					m.Log("info", "not find %s", arg[0])
					m.Start(arg[0], "shy")
					msg = m
				}

				// 查找脚本
				p := m.Cmdx("nfs.path", path.Join(msg.Conf("project", "plugin.path"), arg[0], "index.shy"))
				if p == "" {
					p = m.Cmdx("nfs.hash", m.Cmdx("web.get", "dev", fmt.Sprintf("publish/%s", arg[0]),
						"GOARCH", m.Conf("runtime", "host.GOARCH"),
						"GOOS", m.Conf("runtime", "host.GOOS"),
						"upgrade", "plugin"))
				}

				// 加载脚本
				if p != "" {
					msg.Target().Configs["_index"] = &ctx.Config{Name: "_index", Value: []interface{}{}}
					msg.Optionv("bio.ctx", msg.Target())
					msg.Cmdy("nfs.source", p)
					msg.Confv("ssh.componet", arg[0], msg.Confv("_index"))
				}

				// 组件列表
				m.Confm("ssh.componet", arg[0], func(index int, value map[string]interface{}) {
					m.Push("index", index)
					m.Push("name", value["componet_name"])
					m.Push("help", value["componet_help"])
					m.Push("ctx", value["componet_ctx"])
					m.Push("cmd", value["componet_cmd"])
				})
				m.Table()

			default:
				restart := false
				for _, link := range kit.View([]string{arg[0]}, m.Confm("upgrade")) {
					file := kit.Select(link, m.Conf("upgrade", []string{"list", strings.Replace(link, ".", "_", -1)}))

					// 下载文件
					m.Cmd("web.get", "dev", fmt.Sprintf("publish/%s", link),
						"GOARCH", m.Conf("runtime", "host.GOARCH"),
						"GOOS", m.Conf("runtime", "host.GOOS"),
						"upgrade", arg[0], "save", file)

					// 执行文件
					if strings.HasPrefix(file, "bin/") {
						if m.Cmd("cli.system", "chmod", "a+x", file); link == "bench" {
							m.Cmd("cli.system", "mv", "bin/bench", fmt.Sprintf("bin/bench_%s", m.Time("20060102_150405")))
							m.Cmd("cli.system", "mv", "bin/bench.new", "bin/bench")
							file = "bin/bench"
						}
						restart = true
					}

					// 输出信息
					m.Push("time", m.Time())
					m.Push("cost", m.Format("cost"))
					m.Push("hash", kit.Hashs(file)[:8])
					m.Push("file", file)

					// 压缩文件
					if strings.HasSuffix(link, ".tar.gz") {
						m.Cmd("cli.system", "tar", "-xvf", file, "-C", path.Dir(file))
						os.Remove(file)
					}
				}

				// 重启服务
				if m.Table(); restart {
					m.Cmd("cli.quit", 1)
				}
			}
			return
		}},
		"missyou": &ctx.Command{Name: "missyou [topic] [name [action]]", Help: "任务管理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 任务主题
			topic := "hello"
			if len(arg) > 0 && (arg[0] == "" || m.Cmds("nfs.path", path.Join(m.Conf("cli.project", "plugin.path"), arg[0]))) {
				topic, arg = arg[0], arg[1:]
			}

			// 任务列表
			if len(arg) == 0 {
				m.Cmd("nfs.dir", m.Conf("missyou", "path"), "time", "name").Table(func(value map[string]string) {
					name := strings.TrimSuffix(value["name"], "/")
					m.Push("create_time", value["time"])
					m.Push("you", name)
					m.Push("status", kit.Select("stop", "start", m.Confs("nfs.node", name)))
				})
				m.Sort("you", "str_r").Sort("status").Table()
				return
			}

			// 任务命名
			if !strings.Contains(arg[0], "-") {
				arg[0] = m.Time("20060102-") + arg[0]
			}

			// 任务管理
			if m.Confs("ssh.node", arg[0]) {
				switch kit.Select("", arg, 1) {
				case "stop":
					m.Cmdy("ssh._route", arg[0], "context", "cli", "quit", 0)
				default:
					m.Echo(arg[0])
				}
				return
			}

			// 启动目录
			p := path.Join(m.Conf("missyou", "path"), arg[0])
			m.Assert(os.MkdirAll(p, 0777))

			// 启动参数
			args := []string{
				"daemon",
				"cmd_dir", p,
				"cmd_daemon", "true",
				"cmd_env", "PATH", os.Getenv("PATH"),
				"cmd_env", "ctx_type", kit.Select(topic, arg, 1),
				"cmd_env", "ctx_home", m.Conf("runtime", "boot.ctx_home"),
				"cmd_env", "ctx_ups", fmt.Sprintf("127.0.0.1%s", m.Conf("runtime", "boot.ssh_port")),
				"cmd_env", "ctx_box", fmt.Sprintf("http://127.0.0.1%s", m.Conf("runtime", "boot.web_port")),
				"cmd_env", "ctx_bin", m.Conf("runtime", "boot.ctx_bin"),
			}

			// 启动服务
			m.Cmdy("cli.system", path.Join(m.Conf("runtime", "boot.ctx_home"), "bin/node.sh"), "start", args)
			return
		}},
		"version": &ctx.Command{Name: "version", Help: "", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				types := reflect.TypeOf(version)
				value := reflect.ValueOf(version)
				for i := 0; i < types.NumField(); i++ {
					key := types.Field(i)
					val := value.Field(i)
					m.Add("append", "name", key.Name)
					m.Add("append", "type", key.Type.Name())
					m.Add("append", "value", fmt.Sprintf("%v", val))
				}

				m.Table()
				return
			}
			m.Cmd("nfs.save", path.Join(m.Conf("runtime", "boot.ctx_home"), "src/contexts/cli/version.go"), fmt.Sprintf(`package cli
var version = struct {
	time string
	host string
	self int
}{
	"%s", "%s", %d,
}
`, m.Time(), m.Conf("runtime", "node.route"), version.self+1))

			m.Append("directory", "")
			return
		}},
	},
}

func init() {
	ctx.Index.Register(Index, &CLI{Context: Index})
}
