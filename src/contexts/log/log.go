package log // {{{
// }}}
import ( // {{{
	"contexts"
	"fmt"
	Log "log"
	"os"
	"strconv"
	"strings"
	"time"
)

// }}}

type LOG struct {
	module map[string]map[string]bool
	silent map[string]bool
	color  map[string]int
	*Log.Logger

	*ctx.Message
	*ctx.Context
}

func (log *LOG) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(LOG)
	s.Context = c
	return s
}

// }}}
func (log *LOG) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	log.Message = m

	log.Configs["flag_date"] = &ctx.Config{Name: "输出日期", Value: "true", Help: "模块日志输出消息日期"}
	log.Configs["flag_time"] = &ctx.Config{Name: "输出时间", Value: "true", Help: "模块日志输出消息时间"}
	log.Configs["flag_color"] = &ctx.Config{Name: "输出颜色", Value: "true", Help: "模块日志输出颜色"}
	log.Configs["flag_code"] = &ctx.Config{Name: "输出序号", Value: "true", Help: "模块日志输出消息的编号"}
	log.Configs["flag_action"] = &ctx.Config{Name: "输出类型", Value: "true", Help: "模块日志类型"}
	log.Configs["flag_name"] = &ctx.Config{Name: "输出名称", Value: "true", Help: "模块日志输出消息源模块与消息目的模块"}

	log.Configs["bench.log"] = &ctx.Config{Name: "日志文件", Value: "var/bench.log", Help: "模块日志输出的文件", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
		if len(arg) > 0 {
			if m.Sess("nfs") == nil {
				os.Create(arg[0])
				m.Sess("nfs", "nfs").Cmd("open", arg[0], "", "日志文件")
			}
			return arg[0]
		}
		return x.Value
	}}

	return log
}

// }}}
func (log *LOG) Start(m *ctx.Message, arg ...string) bool { // {{{
	log.Message = m
	return false
}

// }}}
func (log *LOG) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch log.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "log", Help: "日志中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"silent": &ctx.Command{Name: "silent [[module] level state]", Help: "查看或设置日志开关, module: 模块名, level: 日志类型, state(true/false): 是否打印日志", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if log, ok := m.Target().Server.(*LOG); m.Assert(ok) { // {{{
				switch len(arg) {
				case 2:
					if len(arg) > 1 {
						log.silent[arg[0]] = ctx.Right(arg[1])
					}
				case 3:
					if log.module[arg[0]] == nil {
						log.module[arg[0]] = map[string]bool{}
					}
					log.module[arg[0]][arg[1]] = ctx.Right(arg[2])
				}

				for k, v := range log.silent {
					m.Echo("%s: %t\n", k, v)
				}
				for k, v := range log.module {
					for i, x := range v {
						m.Echo("%s(%s): %t\n", k, i, x)
					}
				}
			} // }}}
		}},
		"color": &ctx.Command{Name: "color [level color]", Help: "查看或设置日志颜色, level: 日志类型, color: 文字颜色", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if log, ok := m.Target().Server.(*LOG); m.Assert(ok) { // {{{
				if len(arg) > 1 {
					c, e := strconv.Atoi(arg[1])
					m.Assert(e)
					log.color[arg[0]] = c
				}

				for k, v := range log.color {
					m.Echo("\033[%dm%s: %d\033[0m\n", v, k, v)
				}
			} // }}}
		}},
		"log": &ctx.Command{Name: "log level string...", Help: "输出日志, level: 日志类型, string: 日志内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if log, ok := m.Target().Server.(*LOG); m.Assert(ok) { // {{{
				if s, ok := log.silent[arg[0]]; ok && s == true {
					return
				}

				msg := m.Message()
				if x, ok := m.Data["msg"]; ok {
					if msg, ok = x.(*ctx.Message); !ok {
						msg = m.Message()
					}
				}
				if s, ok := log.module[msg.Target().Name]; ok {
					if x, ok := s[arg[0]]; ok && x {
						return
					}
				}

				date := ""
				if m.Confs("flag_date") {
					date += time.Now().Format("2006/01/02 ")
				}
				if m.Confs("flag_time") {
					date += time.Now().Format("15:04:05 ")
				}

				color := 0
				if m.Confs("flag_color") {
					if c, ok := log.color[arg[0]]; ok {
						color = c
					}
				}

				code := ""
				if m.Confs("flag_code") {
					code = fmt.Sprintf("%d ", msg.Code())
				}

				action := ""
				if m.Confs("flag_action") {
					action = fmt.Sprintf("%s", arg[0])

					if m.Confs("flag_name") {
						action = fmt.Sprintf("%s(%s->%s)", action, msg.Source().Name, msg.Target().Name)
						if msg.Name != "" {
							action = fmt.Sprintf("%s(%s:%s->%s.%d)", action, msg.Source().Name, msg.Name, msg.Target().Name, m.Index)
						}
					}
				}

				cmd := strings.Join(arg[1:], "")

				if nfs := m.Sess("nfs"); nfs != nil {
					if nfs.Options("log", false); color > 0 {
						nfs.Cmd("write", fmt.Sprintf("%s\033[%dm%s%s %s\033[0m\n", date, color, code, action, cmd))
					} else {
						nfs.Cmd("write", fmt.Sprintf("%s%s%s %s\n", date, code, action, cmd))
					}
				}
			} // }}}
		}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void", Help: "void",
			Configs: map[string]*ctx.Config{
				"flag_code":   &ctx.Config{},
				"flag_action": &ctx.Config{},
				"flag_name":   &ctx.Config{},
				"flag_color":  &ctx.Config{},
				"flag_time":   &ctx.Config{},
				"flag_date":   &ctx.Config{},
			},
			Commands: map[string]*ctx.Command{"log": &ctx.Command{}},
		},
	},
}

func init() {
	log := &LOG{}
	log.Context = Index
	ctx.Index.Register(Index, log)

	log.color = map[string]int{
		"error":  31,
		"check":  31,
		"cmd":    32,
		"conf":   33,
		"search": 35,
		"find":   35,
		"cb":     35,
		"lock":   35,
		"spawn":  35,
		"begin":  36,
		"start":  36,
		"close":  36,
		"debug":  0,
	}
	log.silent = map[string]bool{
		// "lock": true,
	}
	log.module = map[string]map[string]bool{
		"log": {"cmd": true},
		"lex": {"cmd": true, "debug": true},
		"yac": {"cmd": true, "debug": true},
	}
}
