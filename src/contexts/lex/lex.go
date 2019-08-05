package lex

import (
	"contexts/ctx"
	"toolkit"

	"fmt"
	"strconv"
	"strings"
)

type Seed struct {
	page int
	hash int
	word string
}
type Point struct {
	s int
	c byte
}
type State struct {
	star bool
	next int
	hash int
}

type LEX struct {
	seed []*Seed
	hash map[string]int
	word map[int]string
	hand map[int]string
	page map[string]int

	char  map[byte][]byte
	state map[State]*State
	mat   []map[byte]*State

	*ctx.Context
}

func (lex *LEX) charset(c byte) []byte {
	if cs, ok := lex.char[c]; ok {
		return cs
	}
	return []byte{c}
}
func (lex *LEX) index(m *ctx.Message, hash string, h string) int {
	which, names := lex.hash, lex.word
	if hash == "npage" {
		which, names = lex.page, lex.hand
	}

	if x, e := strconv.Atoi(h); e == nil {
		if hash == "npage" {
			m.Assert(x <= m.Capi("npage"), "语法集合未创建")
		} else {
			lex.hash[h] = x
		}
		return x
	}

	if x, ok := which[h]; ok {
		return x
	}

	which[h] = m.Capi(hash, 1)
	names[which[h]] = h
	m.Assert(hash != "npage" || m.Capi("npage") < m.Confi("meta", "nlang"), "语法集合超过上限")
	return which[h]
}
func (lex *LEX) train(m *ctx.Message, page int, hash int, seed []byte) int {
	m.Log("debug", "%s %s page: %v hash: %v seed: %v", "train", "lex", page, hash, string(seed))

	ss := []int{page}
	cn := make([]bool, m.Confi("meta", "ncell"))
	cc := make([]byte, 0, m.Confi("meta", "ncell"))
	sn := make([]bool, m.Capi("nline"))

	points := []*Point{}

	for p := 0; p < len(seed); p++ {

		switch seed[p] {
		case '[':
			set := true
			if p++; seed[p] == '^' {
				set, p = false, p+1
			}

			for ; seed[p] != ']'; p++ {
				if seed[p] == '\\' {
					p++
					for _, c := range lex.charset(seed[p]) {
						cn[c] = true
					}
					continue
				}

				if seed[p+1] == '-' {
					begin, end := seed[p], seed[p+2]
					if begin > end {
						begin, end = end, begin
					}
					for c := begin; c <= end; c++ {
						cn[c] = true
					}
					p += 2
					continue
				}

				cn[seed[p]] = true
			}

			for c := 0; c < len(cn); c++ {
				if (set && cn[c]) || (!set && !cn[c]) {
					cc = append(cc, byte(c))
				}
				cn[c] = false
			}

		case '.':
			for c := 0; c < len(cn); c++ {
				cc = append(cc, byte(c))
			}

		case '\\':
			p++
			for _, c := range lex.charset(seed[p]) {
				cc = append(cc, c)
			}
		default:
			cc = append(cc, seed[p])
		}

		m.Log("debug", "page: \033[31m%d %v\033[0m", len(ss), ss)
		m.Log("debug", "cell: \033[32m%d %v\033[0m", len(cc), cc)

		flag := '\000'
		if p+1 < len(seed) {
			flag = rune(seed[p+1])
			switch flag {
			case '+', '*', '?':
				p++
			}
		}

		for _, s := range ss {
			line := 0
			for _, c := range cc {

				state := &State{}
				if lex.mat[s][c] != nil {
					*state = *lex.mat[s][c]
				} else {
					m.Capi("nnode", 1)
				}
				m.Log("debug", "GET(%d,%d): %v", s, c, state)

				switch flag {
				case '+':
					state.star = true
				case '*':
					state.star = true
					sn[s] = true
				case '?':
					sn[s] = true
				}

				if state.next == 0 {
					if line == 0 || !m.Confs("meta", "compact") {
						lex.mat = append(lex.mat, make(map[byte]*State))
						line = m.Capi("nline", 1) - 1
						sn = append(sn, false)
					}
					state.next = line
				}
				sn[state.next] = true

				lex.mat[s][c] = state
				points = append(points, &Point{s, c})
				m.Log("debug", "SET(%d,%d): %v(%s,%s)", s, c, state, m.Cap("nnode"), m.Cap("nreal"))
			}
		}

		cc, ss = cc[:0], ss[:0]
		for s, b := range sn {
			if sn[s] = false; b {
				ss = append(ss, s)
			}
		}
	}

	for _, s := range ss {
		if s < m.Confi("meta", "nlang") || s >= len(lex.mat) {
			continue
		}

		if len(lex.mat[s]) == 0 {
			last := m.Capi("nline") - 1
			m.Cap("nline", "0")
			m.Log("debug", "DEL: %d-%d", last, m.Capi("nline", s))
			lex.mat = lex.mat[:s]
		}
	}

	for _, s := range ss {
		for _, p := range points {
			state := &State{}
			*state = *lex.mat[p.s][p.c]

			if state.next == s {
				m.Log("debug", "GET(%d, %d): %v", p.s, p.c, state)
				if state.hash = hash; state.next >= len(lex.mat) {
					state.next = 0
				}
				lex.mat[p.s][p.c] = state
				m.Log("debug", "SET(%d, %d): %v", p.s, p.c, state)
			}

			if x, ok := lex.state[*state]; !ok {
				lex.state[*state] = lex.mat[p.s][p.c]
				m.Capi("nreal", 1)
			} else {
				lex.mat[p.s][p.c] = x
			}
		}
	}

	m.Log("debug", "%s %s npage: %v nhash: %v nseed: %v", "train", "lex", m.Capi("npage"), m.Capi("nhash"), m.Capi("nseed"))
	return hash
}
func (lex *LEX) parse(m *ctx.Message, page int, line []byte) (hash int, rest []byte, word []byte) {
	m.Log("debug", "%s %s page: %v line: %v", "parse", "lex", page, line)

	pos := 0
	for star, s := 0, page; s != 0 && pos < len(line); pos++ {

		c := line[pos]
		if c == '\\' && pos < len(line)-1 { //跳过转义
			pos++
			c = lex.charset(line[pos])[0]
		}
		if c > 127 { //跳过中文
			word = append(word, c)
			continue
		}

		state := lex.mat[s][c]
		if state == nil {
			s, star, pos = star, 0, pos-1
			continue
		}
		m.Log("debug", "GET (%d,%d): %v", s, c, state)

		word = append(word, c)

		if state.star {
			star = s
		} else if x, ok := lex.mat[star][c]; !ok || !x.star {
			star = 0
		}

		if s, hash = state.next, state.hash; s == 0 {
			s, star = star, 0
		}
	}

	if pos == len(line) {
		// hash, pos, word = -1, 0, word[:0]
	} else if hash == 0 {
		pos, word = 0, word[:0]
	}
	rest = line[pos:]

	m.Log("debug", "%s %s hash: %v word: %v rest: %v", "parse", "lex", hash, word, rest)
	return
}
func (lex *LEX) Parse(m *ctx.Message, line []byte, page string) (hash int, rest []byte, word []byte) {
	hash, rest, word = lex.parse(m, lex.index(m, "npage", page), line)
	return hash, rest, word
}

func (lex *LEX) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	return &LEX{Context: c}
}
func (lex *LEX) Begin(m *ctx.Message, arg ...string) ctx.Server {
	lex.Caches["nseed"] = &ctx.Cache{Name: "种子数量", Value: "0", Help: "词法模板的数量"}
	lex.Caches["npage"] = &ctx.Cache{Name: "集合数量", Value: "0", Help: "词法集合的数量"}
	lex.Caches["nhash"] = &ctx.Cache{Name: "类型数量", Value: "0", Help: "词法类型的数量"}

	lex.Caches["nline"] = &ctx.Cache{Name: "状态数量", Value: m.Conf("meta", "nlang"), Help: "状态机状态的数量"}
	lex.Caches["nnode"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "状态机连接的逻辑数量"}
	lex.Caches["nreal"] = &ctx.Cache{Name: "实点数量", Value: "0", Help: "状态机连接的存储数量"}

	lex.page = map[string]int{"nil": 0}
	lex.hash = map[string]int{"nil": 0}
	lex.word = map[int]string{0: "nil"}
	lex.hand = map[int]string{0: "nil"}

	lex.char = map[byte][]byte{
		't': []byte{'\t'},
		'n': []byte{'\n'},
		'b': []byte{'\t', ' '},
		's': []byte{'\t', ' ', '\n'},
		'd': []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'},
		'x': []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'A', 'B', 'C', 'D', 'E', 'F'},
	}
	lex.state = make(map[State]*State)
	lex.mat = make([]map[byte]*State, m.Capi("nline"))

	return lex
}
func (lex *LEX) Start(m *ctx.Message, arg ...string) bool {
	return false
}
func (lex *LEX) Close(m *ctx.Message, arg ...string) bool {
	return true
}

var Index = &ctx.Context{Name: "lex", Help: "词法中心",
	Caches: map[string]*ctx.Cache{
		"nmat": &ctx.Cache{Name: "nmat", Value: "0", Help: "矩阵数量"},
	},
	Configs: map[string]*ctx.Config{
		"npage": &ctx.Config{Name: "npage", Value: "1", Help: "默认页"},
		"nhash": &ctx.Config{Name: "nhash", Value: "1", Help: "默认值"},
		"meta": &ctx.Config{Name: "meta", Value: map[string]interface{}{
			"ncell": 128, "nlang": 64, "compact": true,
			"name": "mat%d", "help": "matrix",
		}, Help: "初始参数"},
	},
	Commands: map[string]*ctx.Command{
		"_init": &ctx.Command{Name: "_init", Help: "默认矩阵", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if _, ok := m.Target().Server.(*LEX); m.Assert(ok) {
				m.Spawn().Cmd("train", "-?[a-zA-Z_0-9:/.]+", "key", "cmd")
				m.Spawn().Cmd("train", "\"[^\"]*\"", "str", "cmd")
				m.Spawn().Cmd("train", "'[^']*'", "str", "cmd")
				m.Spawn().Cmd("train", "#[^\n]*", "com", "cmd")
				m.Spawn().Cmd("train", "[~!@$%()]", "ops", "cmd")
			}
			return
		}},
		"spawn": &ctx.Command{Name: "spawn [help [name]]", Help: "创建矩阵", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if _, ok := m.Target().Server.(*LEX); m.Assert(ok) {
				m.Start(fmt.Sprintf(kit.Select(m.Conf("lex.meta", "name"), arg, 1), m.Capi("nmat", 1)),
					kit.Select(m.Conf("lex.meta", "help"), arg, 0))
			}
			return
		}},
		"train": &ctx.Command{Name: "train seed [hash [page]", Help: "词法训练", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) {
				hash := lex.index(m, "nhash", m.Confx("nhash", arg, 1))
				page := lex.index(m, "npage", m.Confx("npage", arg, 2))
				if lex.mat[page] == nil {
					lex.mat[page] = map[byte]*State{}
				}
				m.Result(0, lex.train(m, page, hash, []byte(arg[0])))

				lex.seed = append(lex.seed, &Seed{page, hash, arg[0]})
				m.Cap("stream", fmt.Sprintf("%s,%s,%s", m.Cap("nseed", len(lex.seed)),
					m.Cap("npage"), m.Cap("nhash", len(lex.hash)-1)))
			}
			return
		}},
		"parse": &ctx.Command{Name: "parse line [page]", Help: "词法解析", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) {
				hash, rest, word := lex.parse(m, lex.index(m, "npage", m.Confx("npage", arg, 1)), []byte(arg[0]))
				m.Result(0, hash, string(rest), string(word))
			}
			return
		}},
		"split": &ctx.Command{Name: "split line [page]", Help: "词法分隔", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) {
				for input := []byte(arg[0]); len(input) > 0; {
					hash, rest, word := lex.parse(m, lex.index(m, "npage", m.Confx("npage", arg, 1)), input)
					m.Log("fuck", "what %v %v %v", hash, rest, word)
					if hash == 0 || len(word) == 0 || len(rest) == len(input) {
						if len(input) > 0 {
							input = input[1:]
						}
						continue
					}

					m.Push("word", string(word))
					m.Push("hash", lex.word[hash])
					m.Push("rest", string(rest))
					input = rest
				}
				m.Table()
			}
			return
		}},
		"show": &ctx.Command{Name: "show seed|page|hash|mat|node", Help: "查看信息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) {
				if len(arg) == 0 {
					m.Push("seed", len(lex.seed))
					m.Push("page", len(lex.page))
					m.Push("hash", len(lex.hash))
					m.Push("nmat", len(lex.mat))
					m.Push("node", len(lex.state))
					m.Table()
					return
				}
				switch arg[0] {
				case "seed":
					for _, v := range lex.seed {
						m.Push("page", lex.hand[v.page])
						m.Push("word", strings.Replace(strings.Replace(v.word, "\n", "\\n", -1), "\t", "\\t", -1))
						m.Push("hash", lex.word[v.hash])
					}
					m.Sort("page", "int").Table()

				case "page":
					for k, v := range lex.page {
						m.Push("page", k)
						m.Push("code", v)
					}
					m.Sort("code", "int").Table()

				case "hash":
					for k, v := range lex.hash {
						m.Push("hash", k)
						m.Push("code", v)
					}
					m.Sort("code", "int").Table()

				case "node":
					for _, v := range lex.state {
						m.Push("star", v.star)
						m.Push("next", v.next)
						m.Push("hash", v.hash)
					}
					m.Table()

				case "mat":
					for i, v := range lex.mat {
						if i <= m.Capi("npage") {
							m.Push("index", lex.hand[i])
						} else if i < m.Confi("meta", "nlang") {
							continue
						} else {
							m.Push("index", i)
						}

						for j := byte(0); j < byte(m.Confi("meta", "ncell")); j++ {
							c := fmt.Sprintf("%c", j)
							switch c {
							case "\n":
								c = "\\n"
							case "\t":
								c = "\\t"
							case " ":
							default:
								if j < 0x20 {
									c = fmt.Sprintf("\\%x", j)
								}
							}

							if s := v[j]; s == nil {
								m.Push(c, "")
							} else {
								star := 0
								if s.star {
									star = 1
								}
								m.Push(c, fmt.Sprintf("%d,%d,%d", star, s.next, s.hash))
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
									key = m.Meta["append"][i] + m.Meta["append"][j]
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
	ctx.Index.Register(Index, &LEX{Context: Index})
}
