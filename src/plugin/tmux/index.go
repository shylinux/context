package main

import (
	"contexts/cli"
	"contexts/ctx"
	"toolkit"

	"fmt"
	"os"
	"strings"
	"time"
)

var Index = &ctx.Context{Name: "tmux", Help: "终端管理",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"mux": &ctx.Config{Name: "mux", Value: map[string]interface{}{
			"cmd_timeout": "100ms",
			"view": map[string]interface{}{
				"session": []interface{}{
					"session_id",
					"session_name",
					"session_windows",
					"session_height",
					"session_width",
					"session_created_string",
				},
				"window": []interface{}{
					"window_id",
					"window_name",
					"window_panes",
					"window_height",
					"window_width",
				},
				"pane": []interface{}{
					"pane_id",
					"pane_index",
					"pane_tty",
					"pane_height",
					"pane_width",
				},
			},
			"bind": map[string]interface{}{
				"0": map[string]interface{}{},
				"1": map[string]interface{}{
					"x": []interface{}{"kill-session"},
				},
				"2": map[string]interface{}{
					"x": []interface{}{"kill-window"},
					"s": []interface{}{"swap-window", "-s"},
					"e": []interface{}{"rename-window"},
				},
				"3": map[string]interface{}{
					"x": []interface{}{"kill-pane"},
					"b": []interface{}{"break-pane"},
					"h": []interface{}{"split-window", "-h"},
					"v": []interface{}{"split-window", "-v"},

					"r": []interface{}{"send-keys"},
					"p": []interface{}{"pipe-pane"},
					"g": []interface{}{"capture-pane", "-p"},

					"s":  []interface{}{"swap-pane", "-d", "-s"},
					"mh": []interface{}{"move-pane", "-h", "-s"},
					"mv": []interface{}{"move-pane", "-v", "-s"},

					"H": []interface{}{"resize-pane", "-L"},
					"L": []interface{}{"resize-pane", "-R"},
					"J": []interface{}{"resize-pane", "-D"},
					"K": []interface{}{"resize-pane", "-U"},
					"Z": []interface{}{"resize-pane", "-Z"},
				},
			},
		}, Help: "文档管理"},
	},
	Commands: map[string]*ctx.Command{
		"mux": &ctx.Command{Name: "mux [session [window [pane]]] args...", Help: "终端管理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 会话列表
			if len(arg) == 0 {
				view := kit.View([]string{"session"}, m.Confm("mux", "view"))
				for _, row := range strings.Split(strings.TrimSpace(m.Cmdx("cli.system", "tmux", "list-sessions", "-F", fmt.Sprintf("#{%s}", strings.Join(view, "},#{")))), "\n") {
					for j, col := range strings.Split(row, ",") {
						m.Add("append", view[j], col)
					}
				}
				for i, k := range m.Meta["append"] {
					if strings.HasPrefix(k, "session_") {
						x := strings.TrimPrefix(k, "session_")
						m.Meta["append"][i] = x
						m.Meta[x] = m.Meta[k]
					}
				}
				m.Table()
				return
			}
			if v := m.Confv("mux", []string{"bind", "0", arg[0]}); v != nil {
				m.Cmdy("cli.system", "tmux", v, arg[1:])
				return
			}

			//窗口列表
			if len(arg) == 1 {
				view := kit.View([]string{"window"}, m.Confm("mux", "view"))
				for _, row := range strings.Split(strings.TrimSpace(m.Cmdx("cli.system", "tmux", "list-windows", "-t", arg[0], "-F", fmt.Sprintf("#{%s}", strings.Join(view, "},#{")))), "\n") {
					for j, col := range strings.Split(row, ",") {
						m.Add("append", view[j], col)
					}
				}
				for i, k := range m.Meta["append"] {
					if strings.HasPrefix(k, "window_") {
						x := strings.TrimPrefix(k, "window_")
						m.Meta["append"][i] = x
						m.Meta[x] = m.Meta[k]
					}
				}
				m.Table()
				return
			}

			switch arg[1] {
			// 创建会话
			case "create":
				m.Cmdy("cli.system", "tmux", "new-session", "-s", arg[0], arg[2:], "-d", "cmd_env", "TMUX", "")
				return
			// 检查会话
			case "exist":
				m.Cmdy("cli.system", "tmux", "has-session", "-t", arg[0])
				return
			// 会话操作
			default:
				if v := m.Confv("mux", []string{"bind", "1", arg[1]}); v != nil {
					m.Cmdy("cli.system", "tmux", v, "-t", arg[0], arg[2:])
					return
				}
			}

			target := fmt.Sprintf("%s:%s", arg[0], arg[1])
			// 面板列表
			if len(arg) == 2 {
				view := kit.View([]string{"pane"}, m.Confm("mux", "view"))
				for _, row := range strings.Split(strings.TrimSpace(m.Cmdx("cli.system", "tmux", "list-panes", "-t", target, "-F", fmt.Sprintf("#{%s}", strings.Join(view, "},#{")))), "\n") {
					for j, col := range strings.Split(row, ",") {
						m.Add("append", view[j], col)
					}
				}
				for i, k := range m.Meta["append"] {
					if strings.HasPrefix(k, "pane_") {
						x := strings.TrimPrefix(k, "pane_")
						m.Meta["append"][i] = x
						m.Meta[x] = m.Meta[k]
					}
				}
				m.Table()
				return
			}

			switch arg[2] {
			// 创建窗口
			case "create":
				m.Cmdy("cli.system", "tmux", "new-window", "-t", arg[0], "-n", arg[1], arg[3:])
				return
			// 窗口操作
			default:
				if v := m.Confv("mux", []string{"bind", "2", arg[2]}); v != nil {
					m.Cmdy("cli.system", "tmux", v, arg[3:], "-t", target)
					return
				}
			}

			// 面板内容
			if target = fmt.Sprintf("%s:%s.%s", arg[0], arg[1], arg[2]); len(arg) == 3 {
				m.Cmdy("cli.system", "tmux", "capture-pane", "-t", target, "-p")
				return
			}

			switch arg[3] {
			case "split":
				ls := strings.Split(m.Cmdx("cli.system", "tmux", "capture-pane", "-t", target, "-p"), "\n")
				for i := 1; i < len(ls)-1; i++ {
					for j, v := range strings.Split(ls[i], " ") {
						m.Push(kit.Format(j), v)
					}
				}
				m.Table()

			case "run":
				m.Cmd("cli.system", "tmux", "send-keys", "-t", target, "clear", "Enter")
				time.Sleep(kit.Duration(m.Conf("mux", "cmd_timeout")))
				prompt := strings.TrimSpace(m.Cmdx("cli.system", "tmux", "capture-pane", "-t", target, "-p"))
				m.Log("info", "wait for prompt %v", prompt)

				m.Cmd("cli.system", "tmux", "send-keys", "-t", target, strings.Join(arg[4:], " "), "Enter")
				for i := 0; i < 1000; i++ {
					time.Sleep(kit.Duration(m.Conf("mux", "cmd_timeout")))
					list := strings.Split(m.Cmdx("cli.system", "tmux", "capture-pane", "-t", target, "-p"), "\n")
					m.Log("info", "current %v", list)
					for j := len(list) - 1; j >= 0; j-- {
						if list[j] != "" {
							if list[j] == prompt {
								i = 1000
							}
							break
						}
					}
				}
				return
			}

			if v := m.Confv("mux", []string{"bind", "3", arg[3]}); v != nil {
				switch arg[3] {
				case "r":
					m.Cmd("cli.system", "tmux", "send-keys", "-t", target, "clear", "Enter")
					m.Cmd("cli.system", "tmux", "clear-history", "-t", target)
					time.Sleep(kit.Duration("10ms"))

					m.Cmd("cli.system", "tmux", "send-keys", "-t", target, strings.Join(arg[4:], " "), "Enter")
					time.Sleep(kit.Duration(m.Conf("mux", "cmd_timeout")))
					m.Cmdy("cli.system", "tmux", "capture-pane", "-t", target, "-p", "-S", "-1000")
				case "p":
					m.Cmdy("cli.system", "tmux", "pipe-pane", "-t", target, arg[4:])
				default:
					m.Cmdy("cli.system", "tmux", v, arg[4:], "-t", target)
				}
				return
			}

			m.Cmdy("cli.system", "tmux", "send-keys", "-t", target, arg[3:])
			return
		}},
		"buf": &ctx.Command{Name: "buf", Help: "缓存管理", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) > 0 {
				msg := m.Cmd(".buf")
				arg[0] = msg.Meta[msg.Meta["append"][0]][0]
			}

			switch len(arg) {
			case 0:
				m.Cmdy("cli.system", "tmux", "list-buffer", "cmd_parse", "cut", 3, ":", "cur bytes text")

			case 2:
				m.Cmdy("cli.system", "tmux", "set-buffer", "-b", arg[0], arg[1])
				fallthrough
			case 1:
				m.Cmdy("cli.system", "tmux", "show-buffer", "-b", arg[0])
			}
			return
		}},
	},
}

func main() {
	fmt.Print(cli.Index.Plugin(Index, os.Args[1:]))
}
