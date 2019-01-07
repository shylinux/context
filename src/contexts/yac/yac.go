package yac

import (
	"contexts/ctx"
	"fmt"
	"strconv"
	"strings"
)

type Seed struct {
	page int
	hash int
	word []string
}
type State struct {
	next int
	star int
	hash int
}
type Point struct {
	s int
	c byte
}
type YAC struct {
	seed []*Seed
	page map[string]int
	word map[int]string
	hash map[string]int
	hand map[int]string

	mat   []map[byte]*State
	state map[State]*State

	lex *ctx.Message
	*ctx.Context
}

func (yac *YAC) name(page int) string {
	if name, ok := yac.word[page]; ok {
		return name
	}
	return fmt.Sprintf("yac%d", page)
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
					if x = yac.lex.Spawn().Cmd("parse", word[i], yac.name(s)).Resulti(0); x == 0 {
						x = yac.lex.Spawn().Cmd("train", word[i], len(yac.mat[s]), yac.name(s)).Resulti(0)
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
		if s < m.Confi("info", "nlang") || s >= len(yac.mat) {
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
func (yac *YAC) parse(m *ctx.Message, out *ctx.Message, page int, void int, line string, level int) (string, []string, int) {
	m.Log("debug", "%s %s\\%d %s(%d): %s", "parse", strings.Repeat("#", level), level, yac.name(page), page, line)

	hash, word := 0, []string{}
	for star, s := 0, page; s != 0 && len(line) > 0; {
		//解析空白
		lex := yac.lex.Spawn()
		if lex.Cmd("parse", line, yac.name(void)); lex.Result(0) == "-1" {
			break
		}
		//解析单词
		line = lex.Result(1)
		lex = yac.lex.Spawn()
		if lex.Cmd("parse", line, yac.name(s)); lex.Result(0) == "-1" {
			break
		}
		//解析状态
		result := append([]string{}, lex.Meta["result"]...)
		i, _ := strconv.Atoi(result[0])
		c := byte(i)
		state := yac.mat[s][c]
		if state != nil { //全局语法检查
			if key := yac.lex.Spawn().Cmd("parse", line, "key"); key.Resulti(0) == 0 || len(key.Result(2)) <= len(result[2]) {
				line, word = result[1], append(word, result[2])
			} else {
				state = nil
			}
		}
		if state == nil { //嵌套语法递归解析
			for i := 0; i < m.Confi("info", "ncell"); i++ {
				if x := yac.mat[s][byte(i)]; i < m.Confi("info", "nlang") && x != nil {
					if l, w, _ := yac.parse(m, out, i, void, line, level+1); l != line {
						line, word = l, append(word, w...)
						state = x
						break
					}
				}
			}
		}
		if state == nil { //语法切换
			s, star = star, 0
			continue
		}
		if s, star, hash = state.next, state.star, state.hash; s == 0 { //状态切换
			s, star = star, 0
		}
	}
	if hash == 0 {
		word = word[:0]
	} else if out != nil { //执行命令
		msg := out.Spawn(m.Source()).Add("detail", yac.hand[hash], word)
		if m.Back(msg); msg.Hand { //命令替换
			m.Assert(!msg.Has("return"))
			word = msg.Meta["result"]
		}
	}

	m.Log("debug", "%s %s/%d %s(%d): %v", "parse", strings.Repeat("#", level), level, yac.name(page), page, word)
	return line, word, hash
}

func (yac *YAC) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}
	if len(arg) > 0 && arg[0] == "scan" {
		return yac
	}

	s := new(YAC)
	s.Context = c
	return s
}
func (yac *YAC) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return yac
}
func (yac *YAC) Start(m *ctx.Message, arg ...string) (close bool) {
	if len(arg) > 0 && arg[0] == "scan" {
		m.Sess("nfs").Call(func(input *ctx.Message) *ctx.Message {
			_, word, _ := yac.parse(m, input, m.Optioni("page"), m.Optioni("void"), input.Detail(0)+"\n", 1)
			input.Result(0, word)
			return nil
		}, "scan", arg[1:])
		return false
	}
	return true
}
func (yac *YAC) Close(m *ctx.Message, arg ...string) bool {
	switch yac.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

var Index = &ctx.Context{Name: "yac", Help: "语法中心",
	Caches: map[string]*ctx.Cache{
		"nparse": &ctx.Cache{Name: "nparse", Value: "0", Help: "解析器数量"},
	},
	Configs: map[string]*ctx.Config{
		"seed": &ctx.Config{Name: "seed", Value: []interface{}{
			map[string]interface{}{"page": "void", "hash": "void", "word": []interface{}{"[\t ]+"}},

			map[string]interface{}{"page": "key", "hash": "key", "word": []interface{}{"[A-Za-z_][A-Za-z_0-9]*"}},
			map[string]interface{}{"page": "num", "hash": "num", "word": []interface{}{"mul{", "0", "-?[1-9][0-9]*", "0[0-9]+", "0x[0-9]+", "}"}},
			map[string]interface{}{"page": "str", "hash": "str", "word": []interface{}{"mul{", "\"[^\"]*\"", "'[^']*'", "}"}},
			map[string]interface{}{"page": "exe", "hash": "exe", "word": []interface{}{"mul{", "$", "@", "}", "key"}},

			map[string]interface{}{"page": "op1", "hash": "op1", "word": []interface{}{"mul{", "-z", "-n", "}"}},
			map[string]interface{}{"page": "op1", "hash": "op1", "word": []interface{}{"mul{", "-e", "-f", "-d", "}"}},
			map[string]interface{}{"page": "op1", "hash": "op1", "word": []interface{}{"mul{", "-", "+", "}"}},
			map[string]interface{}{"page": "op2", "hash": "op2", "word": []interface{}{"mul{", ":=", "=", "+=", "}"}},
			map[string]interface{}{"page": "op2", "hash": "op2", "word": []interface{}{"mul{", "+", "-", "*", "/", "%", "}"}},
			map[string]interface{}{"page": "op2", "hash": "op2", "word": []interface{}{"mul{", "<", "<=", ">", ">=", "==", "!=", "}"}},
			map[string]interface{}{"page": "op2", "hash": "op2", "word": []interface{}{"mul{", "~", "!~", "}"}},

			map[string]interface{}{"page": "val", "hash": "val", "word": []interface{}{"opt{", "op1", "}", "mul{", "num", "key", "str", "exe", "}"}},
			map[string]interface{}{"page": "exp", "hash": "exp", "word": []interface{}{"val", "rep{", "op2", "val", "}"}},
			map[string]interface{}{"page": "map", "hash": "map", "word": []interface{}{"key", ":", "\\[", "rep{", "key", "}", "\\]"}},
			map[string]interface{}{"page": "exp", "hash": "exp", "word": []interface{}{"\\{", "rep{", "map", "}", "\\}"}},
			map[string]interface{}{"page": "val", "hash": "val", "word": []interface{}{"opt{", "op1", "}", "(", "exp", ")"}},

			map[string]interface{}{"page": "stm", "hash": "var", "word": []interface{}{"var", "key", "opt{", "=", "exp", "}"}},
			map[string]interface{}{"page": "stm", "hash": "let", "word": []interface{}{"let", "key", "opt{", "=", "exp", "}"}},
			map[string]interface{}{"page": "stm", "hash": "var", "word": []interface{}{"var", "key", "<-"}},
			map[string]interface{}{"page": "stm", "hash": "var", "word": []interface{}{"var", "key", "<-", "opt{", "exe", "}"}},
			map[string]interface{}{"page": "stm", "hash": "let", "word": []interface{}{"let", "key", "<-", "opt{", "exe", "}"}},

			map[string]interface{}{"page": "stm", "hash": "if", "word": []interface{}{"if", "exp"}},
			map[string]interface{}{"page": "stm", "hash": "else", "word": []interface{}{"else"}},
			map[string]interface{}{"page": "stm", "hash": "end", "word": []interface{}{"end"}},
			map[string]interface{}{"page": "stm", "hash": "for", "word": []interface{}{"for", "opt{", "exp", ";", "}", "exp"}},
			map[string]interface{}{"page": "stm", "hash": "for", "word": []interface{}{"for", "index", "exp", "opt{", "exp", "}", "exp"}},
			map[string]interface{}{"page": "stm", "hash": "label", "word": []interface{}{"label", "exp"}},
			map[string]interface{}{"page": "stm", "hash": "goto", "word": []interface{}{"goto", "exp", "opt{", "exp", "}", "exp"}},

			map[string]interface{}{"page": "stm", "hash": "expr", "word": []interface{}{"expr", "rep{", "exp", "}"}},
			map[string]interface{}{"page": "stm", "hash": "return", "word": []interface{}{"return", "rep{", "exp", "}"}},

			map[string]interface{}{"page": "word", "hash": "word", "word": []interface{}{"mul{", "~", "!", "=", "\\?\\?", "\\?", "<", ">$", ">@", ">", "\\|", "%", "exe", "str", "[a-zA-Z0-9_/\\-.:]+", "}"}},
			map[string]interface{}{"page": "cmd", "hash": "cmd", "word": []interface{}{"rep{", "word", "}"}},
			map[string]interface{}{"page": "exe", "hash": "exe", "word": []interface{}{"$", "(", "cmd", ")"}},

			map[string]interface{}{"page": "line", "hash": "line", "word": []interface{}{"opt{", "mul{", "stm", "cmd", "}", "}", "mul{", ";", "\n", "#[^\n]*\n", "}"}},
		}, Help: "语法集合的最大数量"},
		"info": &ctx.Config{Name: "info", Value: map[string]interface{}{"ncell": 128, "nlang": 64}, Help: "嵌套层级日志的标记"},
	},
	Commands: map[string]*ctx.Command{
		"init": &ctx.Command{Name: "init", Help: "添加语法规则, page: 语法集合, hash: 语句类型, word: 语法模板", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) {
				yac.Caches["nline"] = &ctx.Cache{Name: "状态数量", Value: "64", Help: "状态机状态的数量"}
				yac.Caches["nnode"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "状态机连接的逻辑数量"}
				yac.Caches["nreal"] = &ctx.Cache{Name: "实点数量", Value: "0", Help: "状态机连接的存储数量"}

				yac.Caches["nseed"] = &ctx.Cache{Name: "种子数量", Value: "0", Help: "语法模板的数量"}
				yac.Caches["npage"] = &ctx.Cache{Name: "集合数量", Value: "0", Help: "语法集合的数量"}
				yac.Caches["nhash"] = &ctx.Cache{Name: "类型数量", Value: "0", Help: "语句类型的数量"}

				yac.page = map[string]int{"nil": 0}
				yac.word = map[int]string{0: "nil"}
				yac.hash = map[string]int{"nil": 0}
				yac.hand = map[int]string{0: "nil"}

				yac.mat = make([]map[byte]*State, m.Confi("info", "nlang"))
				yac.state = map[State]*State{}

				m.Confm("seed", func(line int, seed map[string]interface{}) {
					m.Spawn().Cmd("train", seed["page"], seed["hash"], seed["word"])
				})
			}
			return
		}},
		"train": &ctx.Command{Name: "train page hash word...", Help: "添加语法规则, page: 语法集合, hash: 语句类型, word: 语法模板", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) {
				page, ok := yac.page[arg[0]]
				if !ok {
					page = m.Capi("npage", 1)
					yac.page[arg[0]] = page
					yac.word[page] = arg[0]
					m.Assert(page < m.Confi("info", "nlang"), "语法集合过多")

					yac.mat[page] = map[byte]*State{}
					for i := 0; i < m.Confi("info", "nlang"); i++ {
						yac.mat[page][byte(i)] = nil
					}
				}

				hash, ok := yac.hash[arg[1]]
				if !ok {
					hash = m.Capi("nhash", 1)
					yac.hash[arg[1]] = hash
					yac.hand[hash] = arg[1]
				}

				if yac.lex == nil {
					yac.lex = m.Cmd("lex.spawn")
				}

				yac.train(m, page, hash, arg[2:], 1)
				yac.seed = append(yac.seed, &Seed{page, hash, arg[2:]})
				m.Cap("stream", fmt.Sprintf("%d,%s,%s", m.Capi("nseed", 1), m.Cap("npage"), m.Cap("nhash")))
			}
			return
		}},
		"parse": &ctx.Command{Name: "parse page void word...", Help: "解析语句, page: 初始语法, void: 空白语法, word: 解析语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) {
				str, word, hash := yac.parse(m, m, m.Optioni("page", yac.page[arg[0]]), m.Optioni("void", yac.page[arg[1]]), arg[2], 1)
				m.Result(str, yac.hand[hash], word)
			}
			return
		}},
		"scan": &ctx.Command{Name: "scan filename", Help: "解析文件", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) {
				m.Optioni("page", yac.page["line"])
				m.Optioni("void", yac.page["void"])
				if len(arg) > 0 {
					m.Start(fmt.Sprintf("parse%d", m.Capi("nparse", 1)), "parse", key, arg[0])
				} else {
					m.Start(fmt.Sprintf("parse%d", m.Capi("nparse", 1)), "parse")
				}
			}
			return
		}},
		"show": &ctx.Command{Name: "show seed|page|hash|mat", Help: "查看信息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) {
				if len(arg) == 0 {
					m.Append("seed", len(yac.seed))
					m.Append("page", len(yac.page))
					m.Append("hash", len(yac.hash))
					m.Append("node", len(yac.state))
					m.Table()
					return
				}

				switch arg[0] {
				case "seed":
					for _, v := range yac.seed {
						m.Add("append", "page", fmt.Sprintf("%d", v.page))
						m.Add("append", "hash", fmt.Sprintf("%d", v.hash))
						m.Add("append", "word", fmt.Sprintf("%s", strings.Replace(strings.Replace(strings.Join(v.word, " "), "\n", "\\n", -1), "\t", "\\t", -1)))
					}
					m.Table()
				case "page":
					for k, v := range yac.page {
						m.Add("append", "page", k)
						m.Add("append", "code", fmt.Sprintf("%d", v))
					}
					m.Sort("code", "int").Table()
				case "hash":
					for k, v := range yac.hash {
						m.Add("append", "hash", k)
						m.Add("append", "code", fmt.Sprintf("%d", v))
						m.Add("append", "hand", yac.hand[v])
					}
					m.Sort("code", "int").Table()
				case "mat":
					for _, v := range yac.mat {
						for j := byte(0); j < byte(m.Confi("info", "ncell")); j++ {
							s := v[j]
							if s == nil {
								m.Add("append", fmt.Sprintf("%d", j), "")
							} else {
								m.Add("append", fmt.Sprintf("%d", j), fmt.Sprintf("%d,%d,%d", s.star, s.next, s.hash))
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
	},
}

func init() {
	yac := &YAC{}
	yac.Context = Index
	ctx.Index.Register(Index, yac)
}
