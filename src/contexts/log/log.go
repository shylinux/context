package log // {{{
// }}}
import ( // {{{
	"contexts"
	"fmt"
	"os"
	"strings"
	"time"
)

// }}}

type LOG struct {
	nfs *ctx.Message
	out *os.File

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
	log.Configs["flag_color"] = &ctx.Config{Name: "flag_color", Value: "true", Help: "模块日志输出颜色"}
	log.Configs["flag_time"] = &ctx.Config{Name: "flag_time", Value: "2006/01/02 15:04:05 ", Help: "模块日志输出颜色"}
	log.Configs["bench.log"] = &ctx.Config{Name: "bench.log", Value: "var/bench.log", Help: "模块日志输出的文件"}
	return log
}

// }}}
func (log *LOG) Start(m *ctx.Message, arg ...string) bool { // {{{
	if f, e := os.Stat(m.Confx("bench.log", arg, 0)); e == nil && !f.IsDir() {
		log.nfs = m.Sess("nfs").Cmd("append", m.Confx("bench.log", arg, 0), "", "日志文件")
		log.out = log.nfs.Optionv("out").(*os.File)
		fmt.Fprintln(log.out, "\n\n")
	}
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
	Caches: map[string]*ctx.Cache{
		"nlog": &ctx.Cache{Name: "nlog", Value: "0", Help: "日志屏蔽类型"},
	},
	Configs: map[string]*ctx.Config{
		"silent": &ctx.Config{Name: "silent", Value: map[string]interface{}{"cb": true, "find": true}, Help: "日志屏蔽类型"},
		"module": &ctx.Config{Name: "module", Value: map[string]interface{}{
			"log": map[string]interface{}{"cmd": true},
			"lex": map[string]interface{}{"cmd": true, "debug": true},
			"yac": map[string]interface{}{"cmd": true, "debug": true},
		}, Help: "日志屏蔽模块"},
		"color": &ctx.Config{Name: "color", Value: map[string]interface{}{
			"debug": 0, "error": 31, "check": 31,
			"cmd": 32, "conf": 33,
			"search": 35, "find": 35, "cb": 35, "lock": 35,
			"begin": 36, "start": 36, "close": 36,
		}, Help: "日志输出颜色"},
		"log_name": &ctx.Config{Name: "log_name", Value: "dump", Help: "日志屏蔽类型"},
	},
	Commands: map[string]*ctx.Command{
		"open": &ctx.Command{Name: "open file", Help: "打开日志文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
		}},
		"log": &ctx.Command{Name: "log level string...", Help: "输出日志, level: 日志类型, string: 日志内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if log, ok := m.Target().Server.(*LOG); m.Assert(ok) && log.out != nil { // {{{
				if m.Confs("silent", arg[0]) {
					return
				}

				msg, ok := m.Optionv("msg").(*ctx.Message)
				if !ok {
					msg = m
				}
				if m.Confs("module", fmt.Sprintf("%s.%s", msg.Target().Name, arg[0])) {
					return
				}

				color := 0
				if m.Confs("flag_color") && m.Confs("color", arg[0]) {
					color = m.Confi("color", arg[0])
				}

				date := time.Now().Format(m.Conf("flag_time"))
				action := fmt.Sprintf("%d %s(%s->%s)", msg.Code(), arg[0], msg.Source().Name, msg.Target().Name)
				cmd := strings.Join(arg[1:], "")

				if color > 0 {
					log.out.WriteString(fmt.Sprintf("%s\033[%dm%s %s\033[0m\n", date, color, action, cmd))
				} else {
					log.out.WriteString(fmt.Sprintf("%s%s %s\n", date, action, cmd))
				}
			} // }}}
		}},
	},
}

func init() {
	log := &LOG{}
	log.Context = Index
	ctx.Index.Register(Index, log)
}
