package yac

import (
	"contexts/ctx"
	"sort"
	kit "toolkit"

	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Seed struct {
	page int
	hash int
	word []string
}
type Point struct {
	s int
	c byte
}
type State struct {
	star int
	next int
	hash int
}
type YAC struct {
	seed []*Seed
	page map[string]int
	word map[int]string
	hand map[int]string
	hash map[string]int

	state map[State]*State
	mat   []map[byte]*State

	lex *ctx.Message

	label map[string]string
	*ctx.Context
}
type Parser interface {
	Parse(m *ctx.Message, line []byte, page string) (hash int, rest []byte, word []byte)
}

func (yac *YAC) name(page int) string {
	if name, ok := yac.word[page]; ok {
		return name
	}
	return fmt.Sprintf("yac%d", page)
}
func (yac *YAC) index(m *ctx.Message, hash string, h string) int {
	which, names := yac.page, yac.word
	if hash == "nhash" {
		which, names = yac.hash, yac.hand
	}

	if x, ok := which[h]; ok {
		return x
	}

	which[h] = m.Capi(hash, 1)
	names[which[h]] = h
	m.Assert(hash != "npage" || m.Capi("npage") < m.Confi("meta", "nlang"), "语法集合超过上限")
	return which[h]
}
func (yac *YAC) train(m *ctx.Message, page, hash int, word []string, level int) (int, []*Point, []*Point) {
	m.Log("debug", "%s %s\\%d page: %v hash: %v word: %v", "train", strings.Repeat("#", level), level, page, hash, word)

	ss := []int{page}
	sn := make([]bool, m.Capi("nline"))
	points, ends := []*Point{}, []*Point{}

	for i, n, mul := 0, 1, false; i < len(word); i += n {
		if !mul {
			if hash <= 0 && word[i] == "}" {
				return i + 2, points, ends
			}
			ends = ends[:0]
		}

		for _, s := range ss {
			switch word[i] {
			case "opt{", "rep{":
				sn[s] = true
				num, point, end := yac.train(m, s, 0, word[i+1:], level+1)
				n, points = num, append(points, point...)

				for _, x := range end {
					state := &State{}
					*state = *yac.mat[x.s][x.c]
					for i := len(sn); i <= state.next; i++ {
						sn = append(sn, false)
					}
					sn[state.next] = true

					points = append(points, x)
					if word[i] == "rep{" {
						state.star = s
						yac.mat[x.s][x.c] = state
						m.Log("debug", "REP(%d, %d): %v", x.s, x.c, state)
					}
				}
			case "mul{":
				mul, n = true, 1
				goto next
			case "}":
				if mul {
					mul = false
					goto next
				}
				fallthrough
			default:
				x, ok := yac.page[word[i]]
				if !ok {
					if x = kit.Int(yac.lex.Spawn().Cmdx("parse", word[i], yac.name(s))); x == 0 {
						x = kit.Int(yac.lex.Spawn().Cmdx("train", word[i], len(yac.mat[s]), yac.name(s)))
					}
				}

				c := byte(x)
				state := &State{}
				if yac.mat[s][c] != nil {
					*state = *yac.mat[s][c]
				} else {
					m.Capi("nnode", 1)
				}
				if state.next == 0 {
					state.next = m.Capi("nline", 1) - 1
					yac.mat = append(yac.mat, map[byte]*State{})
					for i := 0; i < m.Capi("nline"); i++ {
						yac.mat[state.next][byte(i)] = nil
					}
					sn = append(sn, false)
				}
				sn[state.next] = true
				yac.mat[s][c] = state
				ends = append(ends, &Point{s, c})
				points = append(points, &Point{s, c})
			}
		}
	next:
		if !mul {
			ss = ss[:0]
			for s, b := range sn {
				if sn[s] = false; b {
					ss = append(ss, s)
				}
			}
		}
	}
	for _, s := range ss {
		if s < m.Confi("meta", "nlang") || s >= len(yac.mat) {
			continue
		}
		void := true
		for _, x := range yac.mat[s] {
			if x != nil {
				void = false
				break
			}
		}
		if void {
			last := m.Capi("nline") - 1
			m.Cap("nline", "0")
			m.Log("debug", "DEL: %d-%d", last, m.Capi("nline", s))
			yac.mat = yac.mat[:s]
		}
	}
	for _, s := range ss {
		for _, p := range points {
			state := &State{}
			*state = *yac.mat[p.s][p.c]
			if state.next == s {
				m.Log("debug", "GET(%d, %d): %v", p.s, p.c, state)
				if state.next >= len(yac.mat) {
					state.next = 0
				}
				if hash > 0 {
					state.hash = hash
				}
				yac.mat[p.s][p.c] = state
				m.Log("debug", "SET(%d, %d): %v", p.s, p.c, state)
			}
			if x, ok := yac.state[*state]; !ok {
				yac.state[*state] = yac.mat[p.s][p.c]
				m.Capi("nreal", 1)
			} else {
				yac.mat[p.s][p.c] = x
			}
		}
	}

	m.Log("debug", "%s %s/%d word: %d point: %d end: %d", "train", strings.Repeat("#", level), level, len(word), len(points), len(ends))
	return len(word), points, ends
}
func (yac *YAC) parse(m *ctx.Message, msg *ctx.Message, stack *kit.Stack, page int, void int, line []byte, level int) (rest []byte, word []string, hash int) {
	m.Log("debug", "%s %s\\%d %s(%d): %s", "parse", strings.Repeat("#", level), level, yac.name(page), page, string(line))

	h, r, w := 0, []byte{}, []byte{}
	p, _ := yac.lex.Target().Server.(Parser)

	for star, s := 0, page; s != 0 && len(line) > 0; {
		//解析空白
		if h, r, _ = p.Parse(m, line, yac.name(void)); h == -1 {
			break
		}
		//解析单词
		if h, r, w = p.Parse(m, r, yac.name(s)); h == -1 {
			break
		}

		//解析状态
		state := yac.mat[s][byte(h)]

		//全局语法检查
		if state != nil {
			if hh, _, ww := p.Parse(m, line, "key"); hh == 0 || len(ww) <= len(w) {
				line, word = r, append(word, string(w))
			} else {
				state = nil
			}
		}
		//嵌套语法递归解析
		if state == nil {
			for i := 0; i < m.Confi("meta", "ncell"); i++ {
				if x := yac.mat[s][byte(i)]; i < m.Confi("meta", "nlang") && x != nil {
					if l, w, _ := yac.parse(m, msg, stack, i, void, line, level+1); len(l) != len(line) {
						line, word, state = l, append(word, w...), x
						break
					}
				}
			}
		}

		//语法切换
		if state == nil {
			s, star = star, 0
		} else if s, star, hash = state.next, state.star, state.hash; s == 0 {
			s, star = star, 0
		}
	}

	if hash == 0 {
		word = word[:0]

	} else if !m.Confs("exec", []string{"disable", yac.hand[hash]}) {
		if stack == nil || stack.Peek().Run || m.Confs("exec", []string{"always", yac.hand[hash]}) {
			//执行命令
			cmd := msg.Spawn(m.Optionv("bio.ctx"))
			if cmd.Cmd(yac.hand[hash], word); cmd.Hand {
				word = cmd.Meta["result"]
			}
			//切换模块
			if v := cmd.Optionv("bio.ctx"); v != nil {
				m.Optionv("bio.ctx", v)
			}
		}
	}

	m.Log("debug", "%s %s/%d %s(%d): %v", "parse", strings.Repeat("#", level), level, yac.hand[hash], hash, word)
	return line, word, hash
}

func (yac *YAC) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	return &YAC{Context: c}
}
func (yac *YAC) Begin(m *ctx.Message, arg ...string) ctx.Server {
	yac.Caches["nseed"] = &ctx.Cache{Name: "种子数量", Value: "0", Help: "语法模板的数量"}
	yac.Caches["npage"] = &ctx.Cache{Name: "集合数量", Value: "0", Help: "语法集合的数量"}
	yac.Caches["nhash"] = &ctx.Cache{Name: "类型数量", Value: "0", Help: "语句类型的数量"}

	yac.Caches["nline"] = &ctx.Cache{Name: "状态数量", Value: m.Conf("meta", "nlang"), Help: "状态机状态的数量"}
	yac.Caches["nnode"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "状态机连接的逻辑数量"}
	yac.Caches["nreal"] = &ctx.Cache{Name: "实点数量", Value: "0", Help: "状态机连接的存储数量"}

	yac.Caches["level"] = &ctx.Cache{Name: "level", Value: "0", Help: "嵌套层级"}
	yac.Caches["parse"] = &ctx.Cache{Name: "parse(true/false)", Value: "true", Help: "命令解析"}

	yac.page = map[string]int{"nil": 0}
	yac.word = map[int]string{0: "nil"}
	yac.hash = map[string]int{"nil": 0}
	yac.hand = map[int]string{0: "nil"}

	yac.state = map[State]*State{}
	yac.mat = make([]map[byte]*State, m.Capi("nline"))
	return yac
}
func (yac *YAC) Start(m *ctx.Message, arg ...string) (close bool) {
	return true
}
func (yac *YAC) Close(m *ctx.Message, arg ...string) bool {
	return false
}

var Index = &ctx.Context{Name: "yac", Help: "语法中心",
	Caches: map[string]*ctx.Cache{
		"nshy": &ctx.Cache{Name: "nshy", Value: "0", Help: "引擎数量"},
	},
	Configs: map[string]*ctx.Config{
		"nline": &ctx.Config{Name: "nline", Value: "line", Help: "默认页"},
		"nvoid": &ctx.Config{Name: "nvoid", Value: "void", Help: "默认值"},
		"meta": &ctx.Config{Name: "meta", Value: map[string]interface{}{
			"ncell": 128, "nlang": 64, "compact": true,
			"name": "shy%d", "help": "engine",
		}, Help: "初始参数"},
		"seed": &ctx.Config{Name: "seed", Value: []interface{}{
			map[string]interface{}{"page": "void", "hash": "void", "word": []interface{}{"[\t ]+"}},

			// 数据类型
			map[string]interface{}{"page": "num", "hash": "num", "word": []interface{}{"mul{", "0", "-?[1-9][0-9]*", "0[0-9]+", "0x[0-9]+", "}"}},
			map[string]interface{}{"page": "key", "hash": "key", "word": []interface{}{"[A-Za-z_][A-Za-z_0-9]*"}},
			map[string]interface{}{"page": "str", "hash": "str", "word": []interface{}{"mul{", "\"[^\"]*\"", "'[^']*'", "}"}},
			map[string]interface{}{"page": "exe", "hash": "exe", "word": []interface{}{"mul{", "$", "@", "}", "opt{", "key", "}"}},
			map[string]interface{}{"page": "exe", "hash": "exe", "word": []interface{}{"mul{", "$", "@", "}", "opt{", "num", "}"}},

			// 表达式语句
			map[string]interface{}{"page": "op1", "hash": "op1", "word": []interface{}{"mul{", "-", "+", "}"}},
			map[string]interface{}{"page": "op2", "hash": "op2", "word": []interface{}{"mul{", "~", "!~", "}"}},
			map[string]interface{}{"page": "op2", "hash": "op2", "word": []interface{}{"mul{", "+", "-", "*", "/", "%", "}"}},
			map[string]interface{}{"page": "op2", "hash": "op2", "word": []interface{}{"mul{", "<", "<=", ">", ">=", "==", "!=", "}"}},
			map[string]interface{}{"page": "val", "hash": "val", "word": []interface{}{"opt{", "op1", "}", "mul{", "num", "key", "str", "exe", "}"}},
			map[string]interface{}{"page": "exp", "hash": "exp", "word": []interface{}{"val", "rep{", "op2", "val", "}"}},
			map[string]interface{}{"page": "stm", "hash": "return", "word": []interface{}{"return", "rep{", "exp", "}"}},

			// 命令语句
			map[string]interface{}{"page": "word", "hash": "word", "word": []interface{}{"mul{", "~", "!", "\\?", "\\?\\?", "exe", "str", "[\\-a-zA-Z0-9_:/.%*]+", "=", "<", ">$", ">@", ">", "\\|", "}"}},
			map[string]interface{}{"page": "cmd", "hash": "cmd", "word": []interface{}{"rep{", "word", "}"}},
			map[string]interface{}{"page": "com", "hash": "cmd", "word": []interface{}{"rep{", ";", "cmd", "}"}},
			map[string]interface{}{"page": "com", "hash": "com", "word": []interface{}{"mul{", "#[^\n]*\n?", "\n", "}"}},
			map[string]interface{}{"page": "line", "hash": "line", "word": []interface{}{"opt{", "mul{", "stm", "cmd", "}", "}", "com"}},

			// 复合语句
			map[string]interface{}{"page": "op1", "hash": "op1", "word": []interface{}{"mul{", "!", "}"}},
			map[string]interface{}{"page": "op2", "hash": "op2", "word": []interface{}{"mul{", "&&", "||", "}"}},
			map[string]interface{}{"page": "exe", "hash": "exe", "word": []interface{}{"(", "exp", ")"}},
			map[string]interface{}{"page": "exe", "hash": "exe", "word": []interface{}{"$", "(", "cmd", ")"}},
			map[string]interface{}{"page": "stm", "hash": "var", "word": []interface{}{"var", "key", "opt{", "=", "exp", "}"}},
			map[string]interface{}{"page": "stm", "hash": "let", "word": []interface{}{"let", "key", "opt{", "=", "exp", "}"}},
			map[string]interface{}{"page": "stm", "hash": "let", "word": []interface{}{"let", "key", "=", "\\[", "rep{", "exp", "}", "\\]"}},
			map[string]interface{}{"page": "stm", "hash": "let", "word": []interface{}{"let", "key", "=", "\\{", "rep{", "exp", "}", "\\}"}},
			map[string]interface{}{"page": "stm", "hash": "if", "word": []interface{}{"if", "exp"}},
			map[string]interface{}{"page": "stm", "hash": "for", "word": []interface{}{"for", "rep{", "key", "}"}},
			map[string]interface{}{"page": "stm", "hash": "for", "word": []interface{}{"for", "rep{", "exp", "}"}},
			map[string]interface{}{"page": "stm", "hash": "fun", "word": []interface{}{"fun", "key", "rep{", "exp", "}"}},
			map[string]interface{}{"page": "stm", "hash": "kit", "word": []interface{}{"kit", "rep{", "exp", "}"}},
			map[string]interface{}{"page": "stm", "hash": "else", "word": []interface{}{"else", "opt{", "if", "exp", "}"}},
			map[string]interface{}{"page": "stm", "hash": "end", "word": []interface{}{"end"}},

			// 标签语句
			map[string]interface{}{"page": "stm", "hash": "label", "word": []interface{}{"label", "key"}},
			map[string]interface{}{"page": "stm", "hash": "goto", "word": []interface{}{"goto", "key"}},
			/*

				map[string]interface{}{"page": "op1", "hash": "op1", "word": []interface{}{"mul{", "-z", "-n", "}"}},
				map[string]interface{}{"page": "op1", "hash": "op1", "word": []interface{}{"mul{", "-e", "-f", "-d", "}"}},
				map[string]interface{}{"page": "op2", "hash": "op2", "word": []interface{}{"mul{", ":=", "=", "+=", "}"}},

				map[string]interface{}{"page": "exp", "hash": "exp", "word": []interface{}{"\\{", "rep{", "map", "}", "\\}"}},
				map[string]interface{}{"page": "val", "hash": "val", "word": []interface{}{"opt{", "op1", "}", "(", "exp", ")"}},

				map[string]interface{}{"page": "stm", "hash": "var", "word": []interface{}{"var", "key", "<-"}},
				map[string]interface{}{"page": "stm", "hash": "var", "word": []interface{}{"var", "key", "<-", "opt{", "exe", "}"}},
				map[string]interface{}{"page": "stm", "hash": "let", "word": []interface{}{"let", "key", "<-", "opt{", "exe", "}"}},

				map[string]interface{}{"page": "stm", "hash": "for", "word": []interface{}{"for", "opt{", "exp", ";", "}", "exp"}},
				map[string]interface{}{"page": "stm", "hash": "for", "word": []interface{}{"for", "index", "exp", "opt{", "exp", "}", "exp"}},

			*/

		}, Help: "语法集合的最大数量"},
		"input": &ctx.Config{Name: "input", Value: map[string]interface{}{
			"text":     true,
			"select":   true,
			"button":   true,
			"upfile":   true,
			"textarea": true,
			"exports":  true,
			"feature":  true,
		}, Help: "控件类型"},
		"exec": &ctx.Config{Name: "info", Value: map[string]interface{}{
			"disable": map[string]interface{}{
				"void": true,
				"num":  true,
				"key":  true,
				"op1":  true,
				"op2":  true,
				"word": true,
				"line": true,
			},
			"always": map[string]interface{}{
				"if":   true,
				"else": true,
				"end":  true,
				"for":  true,
			},
		}, Help: "嵌套层级日志的标记"},

		"alias": &ctx.Config{Name: "alias", Value: map[string]interface{}{
			"~":  []string{"context"},
			"!":  []string{"message"},
			":":  []string{"command"},
			"::": []string{"command", "list"},

			"note":     []string{"mdb.note"},
			"pwd":      []string{"nfs.pwd"},
			"path":     []string{"nfs.path"},
			"dir":      []string{"nfs.dir"},
			"brow":     []string{"web.brow"},
			"ifconfig": []string{"tcp.ifconfig"},
		}, Help: "启动脚本"},
	},
	Commands: map[string]*ctx.Command{
		"_init": &ctx.Command{Name: "_init", Help: "添加语法规则, page: 语法集合, hash: 语句类型, word: 语法模板", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) {
				yac.lex = m.Cmd("lex.spawn")
				m.Confm("seed", func(line int, seed map[string]interface{}) {
					m.Spawn().Cmd("train", seed["page"], seed["hash"], seed["word"])
				})
			}
			return
		}},
		"train": &ctx.Command{Name: "train page hash word...", Help: "添加语法规则, page: 语法集合, hash: 语句类型, word: 语法模板", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) {
				page := yac.index(m, "npage", arg[0])
				hash := yac.index(m, "nhash", arg[1])
				if yac.mat[page] == nil {
					yac.mat[page] = map[byte]*State{}
					for i := 0; i < m.Confi("meta", "nlang"); i++ {
						yac.mat[page][byte(i)] = nil
					}
				}
				yac.train(m, page, hash, arg[2:], 1)

				yac.seed = append(yac.seed, &Seed{page, hash, arg[2:]})
				m.Cap("stream", fmt.Sprintf("%d,%s,%s", m.Cap("nseed", len(yac.seed)),
					m.Cap("npage"), m.Cap("nhash", len(yac.hash)-1)))
			}
			return
		}},
		"parse": &ctx.Command{Name: "parse line", Help: "解析语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) {
				stack, _ := m.Optionv("bio.stack").(*kit.Stack)
				m.Optioni("yac.page", yac.page[m.Conf("nline")])
				m.Optioni("yac.void", yac.page[m.Conf("nvoid")])

				_, word, _ := yac.parse(m, m, stack, m.Optioni("yac.page"), m.Optioni("yac.void"), []byte(arg[0]), 1)
				m.Result(word)
			}
			return
		}},
		"show": &ctx.Command{Name: "show seed|page|hash|mat", Help: "查看信息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) {
				if len(arg) == 0 {
					m.Push("seed", len(yac.seed))
					m.Push("page", len(yac.page))
					m.Push("hash", len(yac.hash))
					m.Push("nmat", len(yac.mat))
					m.Push("node", len(yac.state))
					m.Table()
					return
				}

				switch arg[0] {
				case "seed":
					for _, v := range yac.seed {
						m.Push("page", yac.hand[v.page])
						m.Push("word", strings.Replace(strings.Replace(fmt.Sprint(v.word), "\n", "\\n", -1), "\t", "\\t", -1))
						m.Push("hash", yac.word[v.hash])
					}
					m.Sort("page", "int").Table()

				case "page":
					for k, v := range yac.page {
						m.Add("append", "page", k)
						m.Add("append", "code", v)
					}
					m.Sort("code", "int").Table()

				case "hash":
					for k, v := range yac.hash {
						m.Add("append", "hash", k)
						m.Add("append", "code", v)
						m.Add("append", "hand", yac.hand[v])
					}
					m.Sort("code", "int").Table()

				case "node":
					for _, v := range yac.state {
						m.Push("star", v.star)
						m.Push("next", v.next)
						m.Push("hash", v.hash)
					}
					m.Table()

				case "mat":
					for i, v := range yac.mat {
						if i <= m.Capi("npage") {
							m.Push("index", yac.hand[i])
						} else if i < m.Confi("meta", "nlang") {
							continue
						} else {
							m.Push("index", i)
						}

						for j := byte(0); j < byte(m.Confi("meta", "ncell")); j++ {
							c := fmt.Sprintf("%d", j)
							if s := v[j]; s == nil {
								m.Push(c, "")
							} else {
								m.Push(c, fmt.Sprintf("%d,%d,%d", s.star, s.next, s.hash))
							}
						}
					}

					ncol := len(m.Meta["append"])
					nrow := len(m.Meta[m.Meta["append"][0]])
					for i := 0; i < ncol-1; i++ {
						for j := i + 1; j < ncol; j++ {
							same := true
							void := true
							for n := 0; n < nrow; n++ {
								if m.Meta[m.Meta["append"][i]][n] != "" {
									void = false
								}

								if m.Meta[m.Meta["append"][i]][n] != m.Meta[m.Meta["append"][j]][n] {
									same = false
									break
								}
							}

							if same {
								if !void {
									key = m.Meta["append"][i] + "," + m.Meta["append"][j]
									m.Meta[key] = m.Meta[m.Meta["append"][i]]
									m.Meta["append"][i] = key
								}

								for k := j; k < ncol-1; k++ {
									m.Meta["append"][k] = m.Meta["append"][k+1]
								}
								ncol--
								j--
							}
						}
					}
					m.Meta["append"] = m.Meta["append"][:ncol]
					m.Table()
				}
			}
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
				bio := m.Optionv("bio.msg").(*ctx.Message)
				msg := m.Spawn(m.Optionv("bio.ctx"))
				switch arg[0] {
				case "$":
					// 局部变量
					if stack, ok := m.Optionv("bio.stack").(*kit.Stack); ok {
						if v, ok := stack.Hash(arg[1]); ok {
							m.Echo("%v", v)
							break
						}
					}

					// 函数参数
					if i, e := strconv.Atoi(arg[1]); e == nil {
						m.Echo(bio.Detail(i))
						break
					}
					// 函数选项
					m.Echo(kit.Select(msg.Cap(arg[1]), bio.Option(arg[1])))

				case "@":
					// 局部变量
					if stack, ok := m.Optionv("bio.stack").(*kit.Stack); ok {
						if v, ok := stack.Hash(arg[1]); ok {
							m.Echo("%v", v)
							break
						}
					}

					// 函数参数
					if i, e := strconv.Atoi(arg[1]); e == nil {
						m.Echo(bio.Detail(i))
						break
					}
					// 函数配置
					m.Echo(kit.Select(msg.Conf(arg[1]), bio.Option(arg[1])))

				default:
					m.Echo(arg[0]).Echo(arg[1])
				}
			default:
				switch arg[0] {
				case "$", "@":
					m.Result(0, arg[2:len(arg)-1])
				case "(":
					m.Echo(arg[1])
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
				case "!":
					result = kit.Format(!kit.Right(arg[1]))
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
						result = strings.Replace(arg[0], arg[2], "", -1)
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
				case "&&":
					if kit.Right(arg[0]) {
						result = arg[2]
					} else {
						result = arg[0]
					}
				case "||":
					if kit.Right(arg[0]) {
						result = arg[0]
					} else {
						result = arg[2]
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
				"=":  -1,
				"||": 0,
				"==": 1, "~": 1,
				"+": 2, "-": 2,
				"*": 3, "/": 3, "%": 3,
			}
			num, op := []string{arg[0]}, []string{}

			for i := 1; i < len(arg); i += 2 {
				if len(op) > 0 && pre[op[len(op)-1]] >= pre[arg[i]] {
					num[len(op)-1] = m.Cmdx("yac.val", num[len(op)-1], op[len(op)-1], num[len(op)])
					num, op = num[:len(num)-1], op[:len(op)-1]
				}
				num, op = append(num, arg[i+1]), append(op, arg[i])
			}

			for i := len(op) - 1; i >= 0; i-- {
				num[i] = m.Cmdx("yac.val", num[i], op[i], num[i+1])
			}

			m.Echo("%s", num[0])
			return
		}},
		"return": &ctx.Command{Name: "return result...", Help: "结束脚本, result: 返回值", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Appends("bio.end", true)
			m.Result(arg[1:])
			return
		}},
		"com": &ctx.Command{Name: "com", Help: "解析注释", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			return
		}},
		"cmd": &ctx.Command{Name: "cmd word", Help: "解析命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			// 解析别名
			detail := []string{}
			if alias, ok := m.Confv("alias", arg[0]).([]string); ok {
				detail, arg = append(detail, alias...), arg[1:]
			}
			detail = append(detail, arg...)

			// 目标切换
			target := m.Optionv("bio.ctx")
			if detail[0] != "context" {
				defer func() { m.Optionv("bio.ctx", target) }()
			}

			// 解析脚本
			msg := m
			for k, v := range m.Confv("system", "script").(map[string]interface{}) {
				if strings.HasSuffix(detail[0], "."+k) {
					msg = m.Spawn(m.Optionv("bio.ctx"))
					detail[0] = m.Cmdx("nfs.path", detail[0])
					detail = append([]string{v.(string)}, detail...)
					break
				}
			}

			// 解析路由
			if msg == m {
				if routes := strings.Split(detail[0], "."); len(routes) > 1 && !strings.Contains(detail[0], ":") {
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
					msg = m.Spawn(m.Optionv("bio.ctx"))
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
			if msg.Appends("bio.ctx1") {
				target = msg.Target()
			}

			// 管道命令
			if len(rest) > 0 {
				pipe := msg.Spawn()
				pipe.Copy(msg, "append").Copy(msg, "result").Cmd("cmd", rest)
				msg.Set("append").Copy(pipe, "append")
				msg.Set("result").Copy(pipe, "result")
			}

			// 导出结果
			for _, v := range exports {
				if v["file"] != "" {
					m.Sess("nfs").Copy(msg, "option").Copy(msg, "append").Copy(msg, "result").Cmd("export", v["file"])
					msg.Set("result")
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
			m.Optionv("bio.ctx", msg.Target())
			m.Set("append").Copy(msg, "append")
			m.Set("result").Copy(msg, "result")
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

		"var": &ctx.Command{Name: "var a [= exp]", Help: "定义变量, a: 变量名, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if stack, ok := m.Optionv("bio.stack").(*kit.Stack); ok {
				m.Log("stack", "%v = %v", arg[1], arg[3])
				stack.Peek().Hash[arg[1]] = arg[3]
			}
			return
		}},
		"let": &ctx.Command{Name: "let a = exp", Help: "设置变量, a: 变量名, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if stack, ok := m.Optionv("bio.stack").(*kit.Stack); ok {
				switch arg[3] {
				case "[":
					list := []interface{}{}
					for i := 4; i < len(arg)-1; i++ {
						list = append(list, arg[i])
					}
					m.Log("stack", "%v = %v", arg[1], list)
					stack.Hash(arg[1], list)

				case "{":
					list := map[string]interface{}{}
					for i := 4; i < len(arg)-2; i += 2 {
						list[arg[i]] = arg[i+1]
					}
					m.Log("stack", "%v = %v", arg[1], list)
					stack.Hash(arg[1], list)

				default:
					m.Log("stack", "%v = %v", arg[1], arg[3])
					stack.Hash(arg[1], arg[3])
				}
			}
			return
		}},
		"if": &ctx.Command{Name: "if exp", Help: "条件语句, exp: 表达式", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			stack := m.Optionv("bio.stack").(*kit.Stack)
			o := stack.Peek()
			p := stack.Push(arg[0], o.Run && kit.Right(arg[1]), m.Optioni("stack.pos"))
			m.Log("stack", "push %v", p.String("\\"))
			if !o.Run || p.Run {
				p.Done = true
			}
			return
		}},
		"for": &ctx.Command{Name: "for exp | for index val... in list", Help: "循环语句",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				stack := m.Optionv("bio.stack").(*kit.Stack)
				m.Log("stack", "push %v", stack.Push(arg[0], stack.Peek().Run && kit.Right(arg[1]), m.Optioni("stack.pos")).String("\\"))

				val, _ := stack.Hash(arg[len(arg)-1])
				index := kit.Int(stack.FS[len(stack.FS)-2].Hash["_index"])
				switch val := val.(type) {
				case map[string]interface{}:
					list := make([]string, 0, len(val))
					for k, _ := range val {
						list = append(list, k)
					}
					sort.Strings(list)

					if index < len(list) {
						stack.Hash(arg[1], list[index])
					}
					stack.Peek().Run = false
					for i, j := 2, index; i < len(arg)-2 && j < len(list); i, j = i+1, j+1 {
						stack.Peek().Run = true
						stack.Hash(arg[i], val[list[j]])
						stack.FS[len(stack.FS)-2].Hash["_index"] = j + 1
					}

				case []interface{}:
					stack.Hash(arg[1], index)
					stack.Peek().Run = false
					for i, j := 2, index; i < len(arg)-2 && j < len(val); i, j = i+1, j+1 {
						stack.Peek().Run = true
						stack.Hash(arg[i], val[j])
						stack.FS[len(stack.FS)-2].Hash["_index"] = j + 1
					}
				default:
				}
				if !stack.Peek().Run {
					stack.FS[len(stack.FS)-2].Hash["_index"] = 0
				}
				return
			}},
		"fun": &ctx.Command{Name: "fun name help", Help: "小函数", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			p := m.Optionv("bio.stack").(*kit.Stack).Push(arg[0], false, m.Optioni("stack.pos"))
			m.Log("stack", "push %v", p.String("\\"))

			if len(arg) > 2 {
				m.Cmd("kit", "kit", arg[1:])
			}
			self := &ctx.Command{Name: strings.Join(arg[1:], " "), Help: []string{"pwd", "ls"}}
			self.Hand = func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
				m.Goshy(self.Help.([]string), 0, nil, nil)
				m.Log("time", "parse: %v", m.Format("cost"))
				return
			}
			m.Target().Commands[arg[1]] = self
			m.Log("info", "fun: %v %v", arg[1], arg)
			p.Data = self
			return
		}},
		"kit": &ctx.Command{Name: "kit name help [init [view]] [public|protected|private] cmd arg... [input value [key val]...]...", Help: "小功能", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Log("info", "_index: %v", arg)

			args := []interface{}{}
			inputs := []interface{}{}
			exports := []interface{}{}
			feature := map[string]interface{}{}

			init, view, right, cmd := "", "", "", ""
			begin := 3

			switch arg[3] {
			case "private", "protected", "public":
				begin, right, cmd = 5, arg[3], arg[4]
			default:
				switch arg[4] {
				case "private", "protected", "public":
					begin, init, right, cmd = 6, arg[3], arg[4], arg[5]
				default:
					begin, init, view, right, cmd = 7, arg[3], arg[4], arg[5], arg[6]
				}
			}

			if m.Confs("input", cmd) {
				cmd, begin = arg[1], begin-1
			}

			for i := begin; i < len(arg); i++ {
				if !m.Confs("input", arg[i]) {
					args = append(args, arg[i])
					continue
				}
				for j := i; j < len(arg); j++ {
					if j < len(arg)-1 && !m.Confs("input", arg[j+1]) {
						continue
					}
					args := arg[i : j+1]
					if arg[i] == "feature" {
						for k := 2; k < len(args); k++ {
							feature[args[1]] = kit.Merge(feature[args[1]], args[k])
						}

					} else if arg[i] == "exports" {
						for k := 1; k < len(args); k += 1 {
							exports = append(exports, args[k])
						}
					} else {
						input := map[string]interface{}{
							"type":  kit.Select("", args, 0),
							"value": kit.Select("", args, 1),
						}
						for k := 2; k < len(args)-1; k += 2 {
							input[args[k]] = kit.Merge(input[args[k]], args[k+1])
						}
						inputs = append(inputs, input)
					}
					i = j
					break
				}
			}

			if len(inputs) == 0 {
				inputs = []interface{}{
					map[string]interface{}{"type": "text", "name": "arg"},
					map[string]interface{}{"type": "button", "value": "执行"},
				}
			}

			ctx := m.Cap("module")
			if strings.Contains(cmd, ".") {
				cs := strings.Split(cmd, ".")
				ctx = strings.Join(cs[:len(cs)-1], ".")
				cmd = cs[len(cs)-1]
			}

			m.Confv("_index", []interface{}{-2}, map[string]interface{}{
				"name": kit.Select("", arg, 1),
				"help": kit.Select("", arg, 2),
				"view": view,
				"init": init,
				"type": right,

				"ctx":     ctx,
				"cmd":     cmd,
				"args":    args,
				"inputs":  inputs,
				"exports": exports,
				"feature": feature,
			})
			return
		}},
		"else": &ctx.Command{Name: "else", Help: "条件语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			p := m.Optionv("bio.stack").(*kit.Stack).Peek()
			p.Run = !p.Done && !p.Run && (len(arg) == 1 || kit.Right(arg[2]))
			m.Log("stack", "set: %v", p.String("|"))
			if p.Run {
				p.Done = true
			}
			return
		}},
		"end": &ctx.Command{Name: "end", Help: "结束语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			p := m.Optionv("bio.stack").(*kit.Stack).Pop()
			m.Log("stack", "pop: %v", p.String("/"))
			switch p.Key {
			case "for":
				if p.Run {
					m.Appendi("bio.pos0", p.Pos)
				}
			case "fun":
				end := m.Optioni("stack.pos")
				self := p.Data.(*ctx.Command)
				help := []string{}
				for i, v := range m.Optionv("bio.input").([]string) {
					if p.Pos < i && i < end {
						help = append(help, v)
					}
				}
				self.Help = help
			}

			return
		}},

		"label": &ctx.Command{Name: "label name", Help: "记录当前脚本的位置, name: 位置名", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			p := m.Optionv("bio.stack").(*kit.Stack).Peek()
			if p.Label == nil {
				p.Label = map[string]int{}
			}
			m.Log("stack", "%v <= %v", arg[1], m.Optioni("stack.pos")+1)
			p.Label[arg[1]] = m.Optioni("stack.pos") + 1
			return
		}},
		"goto": &ctx.Command{Name: "goto label [exp] condition", Help: "向上跳转到指定位置, label: 跳转位置, condition: 跳转条件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			stack := m.Optionv("bio.stack").(*kit.Stack)
			if i, ok := stack.Label(arg[1]); ok {
				m.Log("stack", "%v => %v", arg[1], i)
				m.Append("bio.pos0", i)
			}
			return
		}},
	},
}

func init() {
	ctx.Index.Register(Index, &YAC{Context: Index})
}
