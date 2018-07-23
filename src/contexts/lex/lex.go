package lex // {{{
// }}}
import ( // {{{
	"contexts"
	"fmt"
	"strconv"
)

// }}}

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

	*ctx.Message
	*ctx.Context
}

func (lex *LEX) index(hash string, h string) int { // {{{
	which := lex.page
	if hash == "nhash" {
		which = lex.hash
	}

	if x, e := strconv.Atoi(h); e == nil {
		lex.Assert(hash != "npage" || x < lex.Capi("npage"))
		return x
	}

	if x, ok := which[h]; ok {
		return x
	}

	which[h] = lex.Capi(hash, 1)
	lex.Assert(hash != "npage" || lex.Capi("npage") < lex.Capi("nlang"))
	return which[h]
}

// }}}
func (lex *LEX) charset(c byte) []byte { // {{{
	if cs, ok := lex.char[c]; ok {
		return cs
	}
	return []byte{c}
}

// }}}
func (lex *LEX) train(page int, hash int, seed []byte) int { // {{{

	ss := []int{page}
	cn := make([]bool, lex.Capi("ncell"))
	cc := make([]byte, 0, lex.Capi("ncell"))
	sn := make([]bool, lex.Capi("nline"))

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

		lex.Log("debug", nil, "page: \033[31m%d %v\033[0m", len(ss), ss)
		lex.Log("debug", nil, "cell: \033[32m%d %v\033[0m", len(cc), cc)

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
					lex.Capi("nnode", 1)
				}
				lex.Log("debug", nil, "GET(%d,%d): %v", s, c, state)

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
					if line == 0 || !lex.Confs("compact") {
						lex.mat = append(lex.mat, make(map[byte]*State))
						line = lex.Capi("nline", 1) - 1
						sn = append(sn, false)
					}
					state.next = line
				}
				sn[state.next] = true

				lex.mat[s][c] = state
				points = append(points, &Point{s, c})
				lex.Log("debug", nil, "SET(%d,%d): %v(%s,%s)", s, c, state, lex.Cap("nnode"), lex.Cap("nreal"))
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
		if s < lex.Capi("nlang") || s >= len(lex.mat) {
			continue
		}

		if len(lex.mat[s]) == 0 {
			lex.Log("debug", nil, "DEL: %d-%d", lex.Capi("nline")-1, lex.Capi("nline", 0, s))
			lex.mat = lex.mat[:s]
		}
	}

	for _, s := range ss {
		for _, p := range points {
			state := &State{}
			*state = *lex.mat[p.s][p.c]

			if state.next == s {
				lex.Log("debug", nil, "GET(%d, %d): %v", p.s, p.c, state)
				if state.hash = hash; state.next >= len(lex.mat) {
					state.next = 0
				}
				lex.mat[p.s][p.c] = state
				lex.Log("debug", nil, "SET(%d, %d): %v", p.s, p.c, state)
			}

			if x, ok := lex.state[*state]; !ok {
				lex.state[*state] = lex.mat[p.s][p.c]
				lex.Capi("nreal", 1)
			} else {
				lex.mat[p.s][p.c] = x
			}
		}
	}

	return hash
}

// }}}
func (lex *LEX) parse(m *ctx.Message, page int, line []byte) (hash int, rest []byte, word []byte) { // {{{

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
		lex.Log("debug", nil, "(%d,%d): %v", s, c, state)
		if state == nil {
			s, star, pos = star, 0, pos-1
			continue
		}

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
		hash, pos, word = -1, 0, word[:0]
	} else if hash == 0 {
		pos, word = 0, word[:0]
	}
	rest = line[pos:]
	return
}

// }}}

func (lex *LEX) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	lex.Message = m
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(LEX)
	s.Context = c
	return s
}

// }}}
func (lex *LEX) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	lex.Message = m

	lex.Caches["ncell"] = &ctx.Cache{Name: "字符上限", Value: "128", Help: "字符集合的最大数量"}
	lex.Caches["nlang"] = &ctx.Cache{Name: "词法上限", Value: "64", Help: "词法集合的最大数量"}

	lex.Caches["nseed"] = &ctx.Cache{Name: "种子数量", Value: "0", Help: "词法模板的数量"}
	lex.Caches["npage"] = &ctx.Cache{Name: "集合数量", Value: "0", Help: "词法集合的数量"}
	lex.Caches["nhash"] = &ctx.Cache{Name: "类型数量", Value: "0", Help: "单词类型的数量"}

	lex.Caches["nline"] = &ctx.Cache{Name: "状态数量", Value: "64", Help: "状态机状态的数量"}
	lex.Caches["nnode"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "状态机连接的逻辑数量"}
	lex.Caches["nreal"] = &ctx.Cache{Name: "实点数量", Value: "0", Help: "状态机连接的存储数量"}

	lex.Configs["compact"] = &ctx.Config{Name: "紧凑模式", Value: "true", Help: "词法状态的共用"}

	if len(arg) > 0 {
		if _, e := strconv.Atoi(arg[0]); lex.Assert(e) {
			lex.Cap("nlang", arg[0])
			lex.Cap("nline", arg[0])
		}
	}

	lex.page = map[string]int{"nil": 0}
	lex.hash = map[string]int{"nil": 0}

	lex.mat = make([]map[byte]*State, lex.Capi("nlang"))
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

// }}}
func (lex *LEX) Start(m *ctx.Message, arg ...string) bool { // {{{
	lex.Message = m
	return false
}

// }}}
func (lex *LEX) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch lex.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

// }}}

var Index = &ctx.Context{Name: "lex", Help: "词法中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"train": &ctx.Command{Name: "train seed [hash [page]", Help: "添加词法规则", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) { // {{{
				page, hash := 1, 1
				if len(arg) > 2 {
					page = lex.index("npage", arg[2])
					m.Assert(page < m.Capi("nlang"), "词法集合过多")
				}
				if len(arg) > 1 {
					hash = lex.index("nhash", arg[1])
				}

				if lex.mat[page] == nil {
					lex.mat[page] = map[byte]*State{}
				}

				m.Result(0, lex.train(page, hash, []byte(arg[0])))
				lex.seed = append(lex.seed, &Seed{page, hash, arg[0]})
				lex.Cap("stream", fmt.Sprintf("%d,%s,%s", lex.Capi("nseed", 1), lex.Cap("npage"), lex.Cap("nhash")))
			} // }}}
		}},
		"parse": &ctx.Command{Name: "parse line [page]", Help: "解析单词", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) { // {{{
				page := 1
				if len(arg) > 1 {
					page = lex.index("npage", arg[1])
				}

				hash, rest, word := lex.parse(m, page, []byte(arg[0]))
				m.Result(0, hash, string(rest), string(word))
			} // }}}
		}},
		"split": &ctx.Command{Name: "split line page void help", Help: "分割语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) { // {{{
				page := 1
				if len(arg) > 1 {
					page = lex.index("npage", arg[1])
				}

				void := 2
				if len(arg) > 2 {
					void = lex.index("npage", arg[2])
				}

				help := 2
				if len(arg) > 3 {
					help = lex.index("npage", arg[3])
				}

				rest := []byte(arg[0])
				_, _, rest = lex.parse(m, help, []byte(rest))
				_, _, rest = lex.parse(m, void, []byte(rest))
				hash, word, rest := lex.parse(m, page, []byte(rest))
				m.Add("result", fmt.Sprintf("%d", hash), string(word), string(rest))
			} // }}}
		}},
		"info": &ctx.Command{Name: "info", Help: "显示缓存", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) { // {{{
				for i, v := range lex.seed {
					m.Echo("seed: %d %v\n", i, v)
				}
				for i, v := range lex.page {
					m.Echo("page: %s %d\n", i, v)
				}
				for i, v := range lex.hash {
					m.Echo("hash: %s %d\n", i, v)
				}
				for i, v := range lex.state {
					m.Echo("node: %v %v\n", i, v)
				}
				for i, v := range lex.mat {
					for k, v := range v {
						m.Echo("node: %v %v %v\n", i, k, v)
					}
				}
			} // }}}
		}},
		"check": &ctx.Command{Name: "check page void word...", Help: "解析语句, page: 语法集合, void: 空白语法集合, word: 语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) { // {{{
				set := map[*State]bool{}
				nreal := 0
				for _, v := range lex.state {
					nreal++
					set[v] = true
				}

				nnode := 0
				for i, v := range lex.mat {
					for j, x := range v {
						if x == nil && int(j) < m.Capi("nlang") {
							continue
						}
						nnode++

						if _, ok := set[x]; !ok {
							m.Log("fuck", nil, "not in %d %d %v %p", i, j, x, x)
						}
					}
				}
				m.Log("fuck", nil, "node: %d real: %d", nnode, nreal)
			} // }}}
		}},
	},
}

func init() {
	lex := &LEX{}
	lex.Context = Index
	ctx.Index.Register(Index, lex)
}
