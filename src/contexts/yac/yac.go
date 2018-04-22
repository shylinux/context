package yac // {{{
// }}}
import ( // {{{
	"contexts"
	"fmt"
	"strconv"
	"strings"
)

// }}}

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

	*ctx.Message
	*ctx.Context
}

func (yac *YAC) name(page int) string { // {{{
	if name, ok := yac.word[page]; ok {
		return name
	}
	return fmt.Sprintf("yac%d", page)
}

// }}}
func (yac *YAC) train(m *ctx.Message, page, hash int, word []string) (int, []*Point, []*Point) { // {{{

	ss := []int{page}
	sn := make([]bool, yac.Capi("nline"))

	points := []*Point{}
	ends := []*Point{}

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
				num, point, end := yac.train(m, s, 0, word[i+1:])
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
						yac.Log("debug", nil, "REP(%d, %d): %v", x.s, x.c, state)
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
					if x = yac.Sess("lex").Cmd("parse", word[i], yac.name(s)).Resulti(0); x == 0 {
						x = yac.Sess("lex").Cmd("train", word[i], len(yac.mat[s]), yac.name(s)).Resulti(0)
					}
				}
				c := byte(x)

				state := &State{}
				if yac.mat[s][c] != nil {
					*state = *yac.mat[s][c]
				} else {
					yac.Capi("nnode", 1)
				}
				yac.Log("debug", nil, "GET(%d, %d): %v \033[31m(%s)\033[0m", s, c, state, word[i])

				if state.next == 0 {
					state.next = yac.Capi("nline", 1) - 1
					yac.mat = append(yac.mat, map[byte]*State{})
					for i := 0; i < yac.Capi("nlang"); i++ {
						yac.mat[state.next][byte(i)] = nil
					}
					sn = append(sn, false)
				}
				sn[state.next] = true
				yac.mat[s][c] = state

				ends = append(ends, &Point{s, c})
				points = append(points, &Point{s, c})
				yac.Log("debug", nil, "SET(%d, %d): %v", s, c, state)
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
		if s < yac.Capi("nlang") || s >= len(yac.mat) {
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
			yac.Log("debug", nil, "DEL: %d-%d", yac.Capi("nline")-1, yac.Capi("nline", 0, s))
			yac.mat = yac.mat[:s]
		}
	}

	for _, s := range ss {
		for _, p := range points {
			state := &State{}
			*state = *yac.mat[p.s][p.c]

			if state.next == s {
				yac.Log("debug", nil, "GET(%d, %d): %v", p.s, p.c, state)
				if state.next >= len(yac.mat) {
					state.next = 0
				}
				if hash > 0 {
					state.hash = hash
				}
				yac.mat[p.s][p.c] = state
				yac.Log("debug", nil, "SET(%d, %d): %v", p.s, p.c, state)
			}

			if x, ok := yac.state[*state]; !ok {
				yac.state[*state] = yac.mat[p.s][p.c]
				yac.Capi("nreal", 1)
			} else {
				yac.mat[p.s][p.c] = x
			}
		}
	}

	return len(word), points, ends
}

// }}}
func (yac *YAC) parse(m *ctx.Message, cli *ctx.Context, page, void int, line string) (*ctx.Context, string, []string) { // {{{

	level := m.Capi("level", 1)
	yac.Log("debug", nil, fmt.Sprintf("%s\\%d %s(%d):", m.Cap("label")[0:level], level, yac.word[page], page))

	hash, word := 0, []string{}
	for star, s := 0, page; s != 0 && len(line) > 0; {

		line = yac.Sess("lex").Cmd("parse", line, yac.name(void)).Result(1)
		lex := yac.Sess("lex").Cmd("parse", line, yac.name(s))

		c := byte(lex.Resulti(0))
		state := yac.mat[s][c]

		if state != nil {
			if key := yac.Sess("lex").Cmd("parse", line, "key"); key.Resulti(0) == 0 || len(key.Result(2)) <= len(lex.Result(2)) {
				m.Log("debug", nil, "%s|%d get(%d,%d): %v \033[31m(%s)\033[0m", m.Cap("label")[0:level], level, s, c, state, lex.Result(2))
				line, word = lex.Result(1), append(word, lex.Result(2))
			} else {
				state = nil
			}
		}

		if state == nil {
			for i := 0; i < yac.Capi("ncell"); i++ {
				if x := yac.mat[s][byte(i)]; i < m.Capi("nlang") && x != nil {
					m.Log("debug", nil, "%s|%d try(%d,%d): %v", m.Cap("label")[0:level], level, s, i, x)

					if c, l, w := yac.parse(m, cli, i, void, line); l != line {
						m.Log("debug", nil, "%s|%d get(%d,%d): %v", m.Cap("label")[0:level], level, s, i, x)
						line, word = l, append(word, w...)

						cli, state = c, x
						break
					}
				}
			}
		}

		if state == nil {
			s, star = star, 0
			continue
		}

		if s, star, hash = state.next, state.star, state.hash; s == 0 {
			s, star = star, 0
		}
	}

	if hash == 0 {
		word = word[:0]
	} else {
		if msg := m.Spawn(cli).Cmd(yac.hand[hash], word); msg.Hand {
			m.Log("debug", nil, "%s>%d set(%d): \033[31m%v\033[0m->\033[32m%v\033[0m",
				m.Cap("label")[0:level], level, hash, word, msg.Meta["result"])
			word = msg.Meta["result"]

			m.Copy(msg, "append", "back", "return")
			if cli = msg.Target(); msg.Has("cli") {
				cli = msg.Data["cli"].(*ctx.Context)
			}
		}
	}

	m.Log("debug", nil, "%s/%d %s(%d):", m.Cap("label")[0:level], level, yac.hand[hash], hash)
	m.Capi("level", -1)
	return cli, line, word
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
	if yac.Message = m; yac.Context == Index {
		Pulse = m
	}
	yac.Context.Master(nil)

	yac.Caches["ncell"] = &ctx.Cache{Name: "词法上限", Value: "128", Help: "词法集合的最大数量"}
	yac.Caches["nlang"] = &ctx.Cache{Name: "语法上限", Value: "32", Help: "语法集合的最大数量"}

	yac.Caches["nseed"] = &ctx.Cache{Name: "种子数量", Value: "0", Help: "语法模板的数量"}
	yac.Caches["npage"] = &ctx.Cache{Name: "集合数量", Value: "0", Help: "语法集合的数量"}
	yac.Caches["nhash"] = &ctx.Cache{Name: "类型数量", Value: "0", Help: "语句类型的数量"}

	yac.Caches["nline"] = &ctx.Cache{Name: "状态数量", Value: "32", Help: "状态机状态的数量"}
	yac.Caches["nnode"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "状态机连接的逻辑数量"}
	yac.Caches["nreal"] = &ctx.Cache{Name: "实点数量", Value: "0", Help: "状态机连接的存储数量"}

	yac.Caches["level"] = &ctx.Cache{Name: "嵌套层级", Value: "0", Help: "语法解析嵌套层级"}
	yac.Caches["label"] = &ctx.Cache{Name: "嵌套标记", Value: "####################", Help: "嵌套层级日志的标记"}

	if len(arg) > 0 {
		if _, e := strconv.Atoi(arg[0]); yac.Assert(e) {
			yac.Cap("nlang", arg[0])
			yac.Cap("nline", arg[0])
		}
	}

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

					m.Assert(page < m.Capi("nlang"), "语法集合过多")
					yac.mat[page] = map[byte]*State{}
					for i := 0; i < yac.Capi("nlang"); i++ {
						yac.mat[page][byte(i)] = nil
					}
				}

				hash, ok := yac.hash[arg[1]]
				if !ok {
					hash = m.Capi("nhash", 1)
					yac.hash[arg[1]] = hash
					yac.hand[hash] = arg[1]
				}

				if m.Sess("lex") == nil {
					lex := m.Sess("lex", "lex")
					if lex.Cap("status") == "start" {
						lex.Start(yac.Context.Name+"lex", "语法词法")
					} else {
						lex.Target().Start(lex)
					}
				}

				yac.train(m, page, hash, arg[2:])
				yac.seed = append(yac.seed, &Seed{page, hash, arg[2:]})
				yac.Cap("stream", fmt.Sprintf("%d,%s,%s", yac.Capi("nseed", 1), yac.Cap("npage"), yac.Cap("nhash")))
			}
			// }}}
		}},
		"parse": &ctx.Command{Name: "parse page void word...", Help: "解析语句, page: 语法集合, void: 空白语法集合, word: 语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) { // {{{
				page, ok := yac.page[arg[0]]
				m.Assert(ok)
				void, ok := yac.page[arg[1]]
				m.Assert(ok)

				if cli, ok := m.Data["cli"].(*ctx.Context); m.Assert(ok) {
					cli, rest, word := yac.parse(m, cli, page, void, strings.Join(arg[2:], " "))
					m.Data["cli"] = cli
					m.Result(0, rest, word)
				}
			}
			// }}}
		}},
		"info": &ctx.Command{Name: "info", Help: "显示缓存", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) { // {{{
				for i, v := range yac.seed {
					m.Echo("seed: %d %v\n", i, v)
				}
				for i, v := range yac.page {
					m.Echo("page: %s %d\n", i, v)
				}
				for i, v := range yac.hash {
					m.Echo("hash: %s %d\n", i, v)
				}
				for i, v := range yac.state {
					m.Echo("node: %v %v\n", i, v)
				}
				for i, v := range yac.mat {
					for k, v := range v {
						if v != nil {
							m.Echo("node: %s(%d,%d): %v\n", yac.name(i), i, k, v)
						}
					}
				}
			}
			// }}}
		}},
		"check": &ctx.Command{Name: "check page void word...", Help: "解析语句, page: 语法集合, void: 空白语法集合, word: 语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) { // {{{
				set := map[*State]bool{}
				nreal := 0
				for _, v := range yac.state {
					nreal++
					set[v] = true
				}

				nnode := 0
				for i, v := range yac.mat {
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
			}
			// }}}
		}},
	},
}

func init() {
	yac := &YAC{}
	yac.Context = Index
	ctx.Index.Register(Index, yac)
}
