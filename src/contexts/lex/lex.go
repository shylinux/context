package lex

import (
	"contexts/ctx"
	"fmt"
	"strconv"
	"strings"
)

type Seed struct {
	page int
	hash int
	word string
}
type State struct {
	star bool
	next int
	hash int
}
type Point struct {
	s int
	c byte
}
type LEX struct {
	seed []*Seed
	page map[string]int
	hash map[string]int

	mat   []map[byte]*State
	state map[State]*State
	char  map[byte][]byte

	*ctx.Context
}

func (lex *LEX) index(m *ctx.Message, hash string, h string) int {
	which := lex.page
	if hash == "nhash" {
		which = lex.hash
	}

	if x, e := strconv.Atoi(h); e == nil {
		if hash == "nhash" {
			lex.hash[hash] = x
		}
		m.Assert(hash != "npage" || x < m.Capi("npage"))
		return x
	}

	if x, ok := which[h]; ok {
		return x
	}

	which[h] = m.Capi(hash, 1)
	m.Assert(hash != "npage" || m.Capi("npage") < m.Confi("info", "nlang"))
	return which[h]
}
func (lex *LEX) charset(c byte) []byte {
	if cs, ok := lex.char[c]; ok {
		return cs
	}
	return []byte{c}
}
func (lex *LEX) train(m *ctx.Message, page int, hash int, seed []byte) int {
	m.Log("debug", "%s %s page: %v hash: %v seed: %v", "train", "lex", page, hash, string(seed))

	ss := []int{page}
	cn := make([]bool, m.Confi("info", "ncell"))
	cc := make([]byte, 0, m.Confi("info", "ncell"))
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
					if line == 0 || !m.Confs("info", "compact") {
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
		if s < m.Confi("info", "nlang") || s >= len(lex.mat) {
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

func (lex *LEX) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(LEX)
	s.Context = c
	return s
}
func (lex *LEX) Begin(m *ctx.Message, arg ...string) ctx.Server {
	lex.Caches["nseed"] = &ctx.Cache{Name: "种子数量", Value: "0", Help: "词法模板的数量"}
	lex.Caches["npage"] = &ctx.Cache{Name: "集合数量", Value: "0", Help: "词法集合的数量"}
	lex.Caches["nhash"] = &ctx.Cache{Name: "类型数量", Value: "0", Help: "单词类型的数量"}

	lex.Caches["nline"] = &ctx.Cache{Name: "状态数量", Value: m.Conf("info", "nlang"), Help: "状态机状态的数量"}
	lex.Caches["nnode"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "状态机连接的逻辑数量"}
	lex.Caches["nreal"] = &ctx.Cache{Name: "实点数量", Value: "0", Help: "状态机连接的存储数量"}

	lex.page = map[string]int{"nil": 0}
	lex.hash = map[string]int{"nil": 0}

	lex.mat = make([]map[byte]*State, m.Confi("info", "nlang"))
	lex.state = make(map[State]*State)

	lex.char = map[byte][]byte{
		't': []byte{'\t'},
		'n': []byte{'\n'},
		'b': []byte{'\t', ' '},
		's': []byte{'\t', ' ', '\n'},
		'd': []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'},
		'x': []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'A', 'B', 'C', 'D', 'E', 'F'},
	}

	return lex
}
func (lex *LEX) Start(m *ctx.Message, arg ...string) bool {
	return false
}
func (lex *LEX) Close(m *ctx.Message, arg ...string) bool {
	switch lex.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

var Index = &ctx.Context{Name: "lex", Help: "词法中心",
	Caches: map[string]*ctx.Cache{
		"nmat": &ctx.Cache{Name: "nmat", Value: "0", Help: "nmat"},
	},
	Configs: map[string]*ctx.Config{
		"npage": &ctx.Config{Name: "npage", Value: "1", Help: "npage"},
		"nhash": &ctx.Config{Name: "nhash", Value: "1", Help: "npage"},
		"info":  &ctx.Config{Name: "info", Value: map[string]interface{}{"compact": true, "ncell": 128, "nlang": 64}, Help: "嵌套层级日志的标记"},
	},
	Commands: map[string]*ctx.Command{
		"spawn": &ctx.Command{Name: "spawn", Help: "添加词法规则", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if _, ok := m.Target().Server.(*LEX); m.Assert(ok) {
				m.Start(fmt.Sprintf("matrix%d", m.Capi("nmat", 1)), "matrix")
			}
			return
		}},
		"train": &ctx.Command{Name: "train seed [hash [page]", Help: "添加词法规则", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) {
				page := lex.index(m, "npage", m.Confx("npage", arg, 2))
				hash := lex.index(m, "nhash", m.Confx("nhash", arg, 1))
				if lex.mat[page] == nil {
					lex.mat[page] = map[byte]*State{}
				}
				m.Cap("npage", len(lex.page))
				m.Cap("nhash", len(lex.hash))

				m.Result(0, lex.train(m, page, hash, []byte(arg[0])))
				lex.seed = append(lex.seed, &Seed{page, hash, arg[0]})
				m.Cap("stream", fmt.Sprintf("%d,%s,%s", m.Capi("nseed", 1), m.Cap("npage"), m.Cap("nhash")))
			}
			return
		}},
		"parse": &ctx.Command{Name: "parse line [page]", Help: "解析单词", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) {
				page := lex.index(m, "npage", m.Confx("npage", arg, 1))
				hash, rest, word := lex.parse(m, page, []byte(arg[0]))
				m.Result(0, hash, string(rest), string(word))
			}
			return
		}},
		"show": &ctx.Command{Name: "show seed|page|hash|mat", Help: "查看信息", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) {
				if len(arg) == 0 {
					m.Append("seed", len(lex.seed))
					m.Append("page", len(lex.page))
					m.Append("hash", len(lex.hash))
					m.Append("node", len(lex.state))
					m.Table()
					return
				}
				switch arg[0] {
				case "seed":
					for _, v := range lex.seed {
						m.Add("append", "page", fmt.Sprintf("%d", v.page))
						m.Add("append", "hash", fmt.Sprintf("%d", v.hash))
						m.Add("append", "word", fmt.Sprintf("%s", strings.Replace(strings.Replace(v.word, "\n", "\\n", -1), "\t", "\\t", -1)))
					}
					m.Table()
				case "page":
					for k, v := range lex.page {
						m.Add("append", "page", k)
						m.Add("append", "code", fmt.Sprintf("%d", v))
					}
					m.Sort("code", "int").Table()
				case "hash":
					for k, v := range lex.hash {
						m.Add("append", "hash", k)
						m.Add("append", "code", fmt.Sprintf("%d", v))
					}
					m.Table()
				case "mat":
					for _, v := range lex.mat {
						for j := byte(0); j < byte(m.Confi("info", "ncell")); j++ {
							s := v[j]
							if s == nil {
								m.Add("append", fmt.Sprintf("%c", j), "")
							} else {
								star := 0
								if s.star {
									star = 1
								}
								m.Add("append", fmt.Sprintf("%c", j), fmt.Sprintf("%d,%d,%d", star, s.next, s.hash))
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
	lex := &LEX{}
	lex.Context = Index
	ctx.Index.Register(Index, lex)
}
