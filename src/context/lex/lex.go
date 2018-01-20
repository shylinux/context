package lex // {{{
// }}}
import ( // {{{
	"context"
	"fmt"
	"strconv"
)

// }}}

type Seed struct { // {{{
	page int
	hash int
	word string
}

// }}}
type State struct { // {{{
	star bool
	next int
	hash int
}

// }}}
type LEX struct { // {{{
	seed []*Seed
	page map[string]int
	hash map[string]int

	state map[State]*State
	mat   []map[byte]*State

	*ctx.Message
	*ctx.Context
}

// }}}

func (lex *LEX) index(hash string, h string) int { // {{{
	if x, e := strconv.Atoi(h); e == nil {
		return x
	}

	which := lex.page
	switch hash {
	case "npage":
		which = lex.page
	case "nhash":
		which = lex.hash
	}

	if x, ok := which[h]; ok {
		return x
	}

	x := lex.Capi(hash, 1)
	which[h] = x
	return x
}

// }}}
func (lex *LEX) charset(c byte) []byte { // {{{
	switch c {
	case 't':
		return []byte{'\t'}
	case 'n':
		return []byte{'\n'}
	case 's':
		return []byte{'\t', ' ', '\n'}
	case 'd':
		return []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	case 'x':
		return []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'A', 'B', 'C', 'D', 'E', 'F'}
	}
	return []byte{c}
}

// }}}
func (lex *LEX) train(page int, hash int, seed []byte) int { // {{{

	cn := make([]bool, lex.Capi("ncell"))
	c := make([]byte, 0, lex.Capi("ncell"))
	sn := make([]bool, lex.Capi("nline"))
	s := []int{page}

	ends := []*State{}

	for p, set := 0, true; p < len(seed); p++ {

		switch seed[p] {
		case '[':
			if p++; seed[p] == '^' {
				set, p = false, p+1
			}

			for ; seed[p] != ']'; p++ {
				if seed[p] == '\\' {
					p++
					for _, s := range lex.charset(seed[p]) {
						cn[s] = true
					}
					continue
				}

				if seed[p+1] == '-' {
					begin, end := seed[p], seed[p+2]
					if begin > end {
						begin, end = end, begin
					}
					for i := begin; i <= end; i++ {
						cn[i] = true
					}
					p += 2
					continue
				}

				cn[seed[p]] = true
			}

			for i := 0; i < len(cn); i++ {
				if (set && cn[i]) || (!set && !cn[i]) {
					c = append(c, byte(i))
				}
				cn[i] = false
			}

		case '.':
			for i := 0; i < len(cn); i++ {
				c = append(c, byte(i))
			}

		case '\\':
			p++
			for _, s := range lex.charset(seed[p]) {
				c = append(c, s)
			}
		default:
			c = append(c, seed[p])
		}

		lex.Log("debug", nil, "page: \033[31m%d %v\033[0m", len(s), s)
		lex.Log("debug", nil, "cell: \033[32m%d %v\033[0m", len(c), c)

		flag := '\000'
		if p+1 < len(seed) {
			flag = rune(seed[p+1])
			switch flag {
			case '+', '*', '?':
				p++
			}
		}

		for i := 0; i < len(s); i++ {
			for line, j := 0, byte(0); int(j) < len(c); j++ {
				state := lex.mat[s[i]][c[j]]
				lex.Log("debug", nil, "GET(%d,%d): %v", s[i], c[j], state)
				if state == nil {
					state = &State{}
					lex.Capi("nnode", 1)
				}

				switch flag {
				case '+':
					state.star = true
				case '*':
					state.star = true
					fallthrough
				case '?':
					if sn[s[i]] = true; p == len(seed)-1 {
						for _, n := range ends {
							if n.next == s[i] {
								lex.Log("debug", nil, "GET() state:%v", n)
								n.hash = hash
								lex.Log("debug", nil, "END() state:%v", n)
							}
						}
					}
				}

				if p == len(seed)-1 {
					state.hash = hash
				} else {
					if state.next == 0 {
						if line == 0 || !lex.Caps("compact") {
							lex.mat = append(lex.mat, make(map[byte]*State))
							line = lex.Capi("nline", 1) - 1
							sn = append(sn, false)
						}
						state.next = line
					}
					sn[state.next] = true
				}

				if s, ok := lex.state[*state]; !ok {
					lex.state[*state] = state
					lex.Capi("nreal", 1)
				} else {
					state = s
				}

				lex.mat[s[i]][c[j]] = state

				lex.Log("debug", nil, "SET(%d,%d): %v(%s,%s)", s[i], c[j], state, lex.Cap("nnode"), lex.Cap("nreal"))
				ends = append(ends, state)
			}
		}

		c, s = c[:0], s[:0]
		for i := 0; i < len(sn); i++ {
			if sn[i] {
				s = append(s, i)
			}
			sn[i] = false
		}
	}
	return hash
}

// }}}
func (lex *LEX) parse(page int, line []byte) (hash int, word []byte, rest []byte) { // {{{

	pos := 0

	for star, s := 0, page; s != 0 && pos < len(line); pos++ {

		c := line[pos]
		if c == '\\' && pos < len(line)-1 {
			pos++
			c = lex.charset(line[pos])[0]
		}

		state := lex.mat[s][c]
		lex.Log("debug", nil, "(%d,%d): %v", s, c, state)
		if state == nil && star != 0 {
			s, star = star, 0
			state = lex.mat[s][c]
			lex.Log("debug", nil, "(%d,%d): %v", s, c, state)
		}
		if state == nil {
			break
		}

		if state, ok := lex.mat[star][c]; !ok || state == nil || !state.star {
			star = 0
		}
		if state.star {
			star = s
		}

		word = append(word, c)
		if s, hash = state.next, state.hash; s == 0 {
			s, star = star, 0
		}
	}

	if hash == 0 {
		pos, word = 0, word[:0]
	}
	rest = line[pos:]

	lex.Log("debug", nil, "\033[31m[%v]\033[0m %d [%v]", string(word), hash, string(rest))
	return
}

// }}}

func (lex *LEX) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(LEX)
	s.Context = c
	return s
}

// }}}
func (lex *LEX) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	if lex.Context == Index {
		Pulse = m
	}

	lex.Caches["ncell"] = &ctx.Cache{Name: "字符上限", Value: "128", Help: "字符上限"}
	lex.Caches["nlang"] = &ctx.Cache{Name: "集合上限", Value: "16", Help: "集合上限"}

	lex.Caches["nseed"] = &ctx.Cache{Name: "种子数量", Value: "0", Help: "种子数量"}
	lex.Caches["npage"] = &ctx.Cache{Name: "集合数量", Value: "0", Help: "集合数量"}
	lex.Caches["nhash"] = &ctx.Cache{Name: "类型数量", Value: "0", Help: "类型数量"}

	lex.Caches["nline"] = &ctx.Cache{Name: "状态数量", Value: "16", Help: "状态数量"}
	lex.Caches["nnode"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "节点数量"}
	lex.Caches["nreal"] = &ctx.Cache{Name: "实点数量", Value: "0", Help: "实点数量"}

	lex.Caches["compact"] = &ctx.Cache{Name: "紧凑模式", Value: "true", Help: "实点数量"}
	lex.Caches["debug"] = &ctx.Cache{Name: "调试模式", Value: "false", Help: "调试模式"}
	return lex
}

// }}}
func (lex *LEX) Start(m *ctx.Message, arg ...string) bool { // {{{
	lex.Message = m

	lex.seed = make([]*Seed, 0, 9)
	lex.page = map[string]int{"nil": 0}
	lex.hash = map[string]int{"nil": 0}

	lex.mat = make([]map[byte]*State, lex.Capi("nlang"))
	lex.state = make(map[State]*State)
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

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "lex", Help: "词法中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"train": &ctx.Command{Name: "train seed [hash [page]", Help: "添加词法规则", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) { // {{{
				page, hash := 1, 1
				if len(arg) > 2 {
					page = lex.index("npage", arg[2])
				}
				if len(arg) > 1 {
					hash = lex.index("nhash", arg[1])
				}

				if lex.mat[page] == nil {
					lex.mat[page] = map[byte]*State{}
				}

				lex.seed = append(lex.seed, &Seed{page, hash, string(arg[0])})
				lex.Log("debug", nil, "%d %d %d %v", page, hash, lex.Capi("nseed", 1), arg[0])
				lex.Cap("stream", fmt.Sprintf("%s,%s,%s", lex.Cap("nseed"), lex.Cap("npage"), lex.Cap("nhash")))

				m.Echo("%d", lex.train(page, hash, []byte(arg[0])))
			} // }}}
		}},
		"parse": &ctx.Command{Name: "parse line [page]", Help: "解析单词", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if lex, ok := m.Target().Server.(*LEX); m.Assert(ok) { // {{{
				page := 1
				if len(arg) > 1 {
					page = lex.index("npage", arg[1])
				}

				hash, word, rest := lex.parse(page, []byte(arg[0]))
				m.Add("result", fmt.Sprintf("%d", hash), string(word), string(rest))
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
				_, _, rest = lex.parse(help, []byte(rest))
				_, _, rest = lex.parse(void, []byte(rest))
				hash, word, rest := lex.parse(page, []byte(rest))
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
	},
}

func init() {
	lex := &LEX{}
	lex.Context = Index
	ctx.Index.Register(Index, lex)
}
