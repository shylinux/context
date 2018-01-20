package yac // {{{
// }}}
import ( // {{{
	"context"
	"fmt"
	"strings"
)

// }}}

type Seed struct { // {{{
	page int
	hash int
	word []string
}

// }}}
type State struct { // {{{
	star int
	next int
	hash int
}

// }}}
type YAC struct { // {{{
	seed []*Seed
	page map[string]int
	word map[int]string
	hash map[string]int
	hand map[int]string

	state map[State]*State
	mat   []map[byte]*State

	lex *ctx.Message
	*ctx.Message
	*ctx.Context
}

// }}}

func (yac *YAC) train(page, hash int, word []string) ([]*State, int) { // {{{

	if yac.mat[page] == nil {
		yac.mat[page] = map[byte]*State{}
		for i := 0; i < yac.Capi("nlang"); i++ {
			yac.mat[page][byte(i)] = nil
		}
	}
	sn := make([]bool, yac.Capi("nline"))
	ss := []int{page}

	begin := page
	point := []*State{}
	ends := []*State{}
	state := []*State{}
	mul := false
	skip := len(word)

	for i, n := 0, 1; i < len(word); i += n {
		if !mul {
			if hash <= 0 && word[i] == "}" {
				if skip = i + 2; hash == -1 {
					hash = 0
					break
				}
				return ends, skip
			} else {
				ends = ends[:0]
			}
		}

		for _, s := range ss {
			switch word[i] {
			case "opt{":
				sn[s] = true
				state, n = yac.train(s, 0, word[i+1:])
				for _, x := range state {
					for i := len(sn); i <= x.next; i++ {
						sn = append(sn, false)
					}
					sn[x.next] = true
					point = append(point, x)
				}
			case "rep{":
				sn[s] = true
				state, n = yac.train(s, -1, word[i+1:])
				for _, x := range state {
					x.star = s
					sn[x.next] = true
					point = append(point, x)
					yac.Pulse.Log("info", nil, "END: %v", x)
				}
			case "mul{":
				mul, n = true, 1
				goto next
			case "}":
				if mul {
					mul = false
					goto next
				}
			default:
				c := byte(0)
				if x, ok := yac.page[word[i]]; !ok {
					lex := yac.lex.Spawn(yac.lex.Target())
					lex.Cmd("parse", word[i], fmt.Sprintf("yac%d", s))
					if lex.Gets("result") {
						x = lex.Geti("result")
					} else {
						x = len(yac.mat[s])
						lex = yac.lex.Spawn(yac.lex.Target())
						lex.Cmd("train", word[i], fmt.Sprintf("%d", x), fmt.Sprintf("yac%d", s))
					}
					c = byte(x)
				} else {
					c = byte(x)
				}

				state := yac.mat[s][c]
				yac.Pulse.Log("info", nil, "GET(%d, %d): %v", s, c, state)
				if state == nil {
					state = &State{}
					yac.Pulse.Capi("nnode", 1)
				}

				if state.next == 0 {
					yac.mat = append(yac.mat, map[byte]*State{})
					state.next = yac.Pulse.Capi("nline", 1) - 1
					for i := 0; i < yac.Capi("nlang"); i++ {
						yac.mat[state.next][byte(i)] = nil
					}
					sn = append(sn, false)
				}
				sn[state.next] = true

				if x, ok := yac.state[*state]; !ok {
					yac.Pulse.Capi("nreal", 1)
					yac.state[*state] = state
				} else {
					yac.mat[s][c] = x
				}
				yac.mat[s][c] = state

				yac.Pulse.Log("info", nil, "SET(%d, %d): %v", s, c, state)
				ends = append(ends, state)
				point = append(point, state)
				if s > begin {
					begin = s
				}

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

	for _, n := range ss {
		if n < yac.Pulse.Capi("nlang") || n >= len(yac.mat) {
			continue
		}

		void := true
		for _, x := range yac.mat[n] {
			if x != nil {
				void = false
				break
			}
		}
		if void {
			yac.Pulse.Log("info", nil, "DEL: %d %d", yac.Pulse.Capi("nline"), n)
			yac.Pulse.Capi("nline", 0, n)
			yac.mat = yac.mat[:n]
		}
	}

	for _, n := range ss {
		for _, s := range point {
			if s.next == n {
				yac.Pulse.Log("info", nil, "GET: %v", s)
				if s.next >= len(yac.mat) {
					s.next = 0
				}
				if hash > 0 {
					s.hash = hash
				}
				yac.Pulse.Log("info", nil, "SET: %v", s)
			}
		}

	}

	return ends, skip
}

// }}}
func (yac *YAC) parse(m *ctx.Message, page, void int, line string) ([]string, string) { // {{{

	level := m.Capi("level", 1)
	m.Log("info", nil, "%s\\%d %s(%d):", m.Cap("label")[0:level], level, yac.word[page], page)

	hash, word := 0, []string{}
	for star, s := 0, page; s != 0 && len(line) > 0; {

		lex := yac.lex.Spawn(yac.lex.Target())
		lex.Cmd("parse", line, fmt.Sprintf("yac%d", void))
		line = lex.Meta["result"][2]

		lex = yac.lex.Spawn(yac.lex.Target())
		lex.Cmd("parse", line, fmt.Sprintf("yac%d", s))
		line = lex.Meta["result"][2]

		c := byte(lex.Geti("result"))
		state := yac.mat[s][c]
		if state == nil {

			for i := 0; i < yac.Capi("ncell"); i++ {
				x := yac.mat[s][byte(i)]
				if i >= m.Capi("nlang") || x == nil {
					continue
				}
				m.Log("info", nil, "%s|%d try(%d,%d): %v", m.Cap("label")[0:level], level, s, i, x)
				if w, l := yac.parse(m, i, void, line); l != line {
					m.Log("info", nil, "%s|%d end(%d,%d): %v", m.Cap("label")[0:level], level, s, i, x)
					word = append(word, w...)
					state = x
					line = l
					break
				}
			}
		} else {
			m.Log("info", nil, "%s|%d get(%d,%d): %v \033[31m(%s)\033[0m", m.Cap("label")[0:level], level, s, c, state, lex.Meta["result"][1])
			word = append(word, lex.Meta["result"][1])
		}

		if state == nil {
			s, star = star, 0
			continue
		}

		star, s, hash = state.star, state.next, state.hash
		if s == 0 {
			s, star = star, 0
		}
	}

	if hash == 0 {
		word = word[:0]
	} else {
		msg := m.Spawn(m.Source()).Add("detail", yac.hand[hash], word...)
		if msg.Cmd(); msg.Hand {
			m.Log("info", nil, "%s>%d set(%d): \033[31m%v\033[0m->\033[32m%v\033[0m", m.Cap("label")[0:level], level, hash, word, msg.Meta["result"])
			word = msg.Meta["result"]
		}
	}

	m.Log("info", nil, "%s/%d %s(%d):", m.Cap("label")[0:level], level, yac.hand[hash], hash)
	level = m.Capi("level", -1)
	return word, line
}

// }}}

func (yac *YAC) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(YAC)
	s.Context = c
	return s
}

// }}}
func (yac *YAC) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	yac.Message = m
	if yac.Context == Index {
		Pulse = m
	}

	yac.Caches["ncell"] = &ctx.Cache{Name: "单词上限", Value: "128", Help: "单词上限"}
	yac.Caches["nlang"] = &ctx.Cache{Name: "集合上限", Value: "16", Help: "集合上限"}

	yac.Caches["nseed"] = &ctx.Cache{Name: "种子数量", Value: "0", Help: "种子数量"}
	yac.Caches["npage"] = &ctx.Cache{Name: "集合数量", Value: "0", Help: "集合数量"}
	yac.Caches["nhash"] = &ctx.Cache{Name: "类型数量", Value: "0", Help: "类型数量"}

	yac.Caches["nline"] = &ctx.Cache{Name: "状态数量", Value: "16", Help: "状态数量"}
	yac.Caches["nnode"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "节点数量"}
	yac.Caches["nreal"] = &ctx.Cache{Name: "实点数量", Value: "0", Help: "实点数量"}

	yac.Caches["level"] = &ctx.Cache{Name: "嵌套层级", Value: "0", Help: "嵌套层级"}
	yac.Caches["label"] = &ctx.Cache{Name: "嵌套标记", Value: "####################", Help: "嵌套层级"}

	yac.page = map[string]int{"nil": 0}
	yac.word = map[int]string{0: "nil"}
	yac.hash = map[string]int{"nil": 0}
	yac.hand = map[int]string{0: "nil"}

	yac.mat = make([]map[byte]*State, m.Capi("nlang"))
	yac.state = map[State]*State{}
	return yac
}

// }}}
func (yac *YAC) Start(m *ctx.Message, arg ...string) bool { // {{{
	yac.Message = m
	return false
}

// }}}
func (yac *YAC) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch yac.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

// }}}

var Pulse *ctx.Message
var Index = &ctx.Context{Name: "yac", Help: "语法中心",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"train": &ctx.Command{Name: "train page hash word...", Help: "添加语法规则, page: 语法集合, hash: 语句类型, word: 语法模板", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) { // {{{
				page, ok := yac.page[arg[0]]
				if !ok {
					page = m.Capi("npage", 1)
					yac.page[arg[0]] = page
					yac.word[page] = arg[0]
				}

				hash, ok := yac.hash[arg[1]]
				if !ok {
					hash = m.Capi("nhash", 1)
					yac.hash[arg[1]] = hash
					yac.hand[hash] = arg[1]
				}

				if yac.lex == nil {
					lex := m.Find("lex", true)
					lex.Start("lyacc", "语法词法")
					yac.lex = lex
				}
				yac.seed = append(yac.seed, &Seed{page, hash, arg[2:]})
				yac.Capi("nseed", 1)
				yac.train(page, hash, arg[2:])
			} // }}}
		}},
		"parse": &ctx.Command{Name: "parse page void word...", Help: "解析语句, page: 语法集合, void: 空白语法集合, word: 语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) { // {{{
				page, ok := yac.page[arg[0]]
				m.Assert(ok)
				void, ok := yac.page[arg[1]]
				m.Assert(ok)
				word, rest := yac.parse(m, page, void, strings.Join(arg[2:], " "))
				m.Add("result", rest, word...)
			} // }}}
		}},
	},
}

func init() {
	yac := &YAC{}
	yac.Context = Index
	ctx.Index.Register(Index, yac)
}
