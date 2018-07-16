package yac // {{{
// }}}
import ( // {{{
	"contexts"
	"fmt"
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
func (yac *YAC) parse(m *ctx.Message, page int, void int, line string, level int) (string, []string) { // {{{
	// m.Log("debug", nil, "%s\\%d %s(%d): %s", m.Conf("label")[0:level], level, yac.name(page), page, line)

	hash, word := 0, []string{}
	for star, s := 0, page; s != 0 && len(line) > 0; {
		//解析空白
		lex := m.Sesss("lex", "lex").Cmd("scan", line, yac.name(void))
		if lex.Result(0) == "-1" {
			break
		}

		//解析单词
		line = lex.Result(1)
		lex = m.Sesss("lex", "lex").Cmd("scan", line, yac.name(s))
		if lex.Result(0) == "-1" {
			break
		}

		//解析状态
		c := byte(lex.Resulti(0))
		state := yac.mat[s][c]

		if state != nil { //全局语法检查
			if key := m.Sesss("lex").Cmd("parse", line, "key"); key.Resulti(0) == 0 || len(key.Result(2)) <= len(lex.Result(2)) {
				line, word = lex.Result(1), append(word, lex.Result(2))
			} else {
				state = nil
			}
		}

		if state == nil { //嵌套语法递归解析
			for i := 0; i < yac.Capi("ncell"); i++ {
				if x := yac.mat[s][byte(i)]; i < m.Capi("nlang") && x != nil {
					if l, w := yac.parse(m, i, void, line, level+1); len(w) > 0 {
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
	} else { //执行命令
		msg := m.Spawn(m.Source()).Add("detail", yac.hand[hash], word)
		if m.Back(msg); msg.Hand { //命令替换
			m.Assert(!msg.Has("return"))
			word = msg.Meta["result"]
		}
	}

	// m.Log("debug", nil, "%s/%d %s(%d): %v", m.Conf("label")[0:level], level, yac.name(page), page, word)
	return line, word
}

// }}}

func (yac *YAC) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	yac.Message = m

	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	if len(arg) > 0 && arg[0] == "parse" {
		return yac
	}

	s := new(YAC)
	s.Context = c
	return s
}

// }}}
func (yac *YAC) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	yac.Message = m

	if len(arg) > 0 && arg[0] == "parse" {
		return yac
	}

	yac.Caches["ncell"] = &ctx.Cache{Name: "词法上限", Value: m.Confx("ncell", arg, 0), Help: "词法集合的最大数量"}
	yac.Caches["nlang"] = &ctx.Cache{Name: "语法上限", Value: m.Confx("nlang", arg, 1), Help: "语法集合的最大数量"}
	yac.Caches["nline"] = &ctx.Cache{Name: "状态数量", Value: m.Confx("nlang", arg, 1), Help: "状态机状态的数量"}
	yac.Caches["nnode"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "状态机连接的逻辑数量"}
	yac.Caches["nreal"] = &ctx.Cache{Name: "实点数量", Value: "0", Help: "状态机连接的存储数量"}

	yac.Caches["nseed"] = &ctx.Cache{Name: "种子数量", Value: "0", Help: "语法模板的数量"}
	yac.Caches["npage"] = &ctx.Cache{Name: "集合数量", Value: "0", Help: "语法集合的数量"}
	yac.Caches["nhash"] = &ctx.Cache{Name: "类型数量", Value: "0", Help: "语句类型的数量"}

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

	if len(arg) > 0 && arg[0] == "parse" {

		var out *ctx.Message
		data := make(chan string, 1)
		next := make(chan bool, 1)

		m.Options("scan_end", false)
		defer func() {
			if e := recover(); e != nil {
				m.Option("scan_end", true)
				next <- true
			}
		}()

		//加载文件
		nfs := m.Find("nfs").Call(func(buf *ctx.Message) *ctx.Message {
			out = buf
			data <- buf.Detail(0) + "; "
			<-next
			return nil
		}, "scan", arg[1], "", "扫描文件")

		m.Find("log").Cmd("silent", yac.Context.Name, "debug", true)

		//解析循环
		for m.Cap("stream", nfs.Target().Name); !m.Options("scan_end"); next <- true {
			_, word := yac.parse(m, m.Optioni("page"), m.Optioni("void"), <-data, 1)
			if len(word) > 0 {
				word = word[:len(word)-1]
				if last := len(word) - 1; last >= 0 && len(word[last]) > 0 && word[last][len(word[last])-1] != '\n' {
					word = append(word, "\n")
				}
			}
			out.Result(0, word)
		}
		return true
	}

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

var Index = &ctx.Context{Name: "yac", Help: "语法中心",
	Caches: map[string]*ctx.Cache{
		"nparse": &ctx.Cache{Name: "nparse", Value: "0", Help: "解析器数量"},
	},
	Configs: map[string]*ctx.Config{
		"ncell": &ctx.Config{Name: "词法上限", Value: "128", Help: "词法集合的最大数量"},
		"nlang": &ctx.Config{Name: "语法上限", Value: "32", Help: "语法集合的最大数量"},
		"name": &ctx.Config{Name: "name", Value: "parse", Help: "模块名", Hand: func(m *ctx.Message, x *ctx.Config, arg ...string) string {
			if len(arg) > 0 { // {{{
				return arg[0]
			}
			return fmt.Sprintf("%s%d", x.Value, m.Capi("nparse", 1))
			// }}}
		}},
		"help":  &ctx.Config{Name: "help", Value: "解析模块", Help: "模块帮助"},
		"line":  &ctx.Config{Name: "line", Value: "line", Help: "默认语法"},
		"void":  &ctx.Config{Name: "void", Value: "void", Help: "默认空白"},
		"label": &ctx.Config{Name: "嵌套标记", Value: "####################", Help: "嵌套层级日志的标记"},
	},
	Commands: map[string]*ctx.Command{
		"init": &ctx.Command{Name: "init [ncell [nlang]]", Help: "初始化语法矩阵", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			if _, ok := m.Target().Server.(*YAC); m.Assert(ok) { // {{{
				s := new(YAC)
				s.Context = m.Target()
				m.Target().Server = s
				m.Target().Begin(m, arg...)
			}
			// }}}
		}},
		"info": &ctx.Command{Name: "info", Help: "查看语法矩阵", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
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
		"parse": &ctx.Command{
			Name: "parse filename [name [help]] [line line] [void void]",
			Help: "解析文件, filename: name:模块名, help:模块帮助, 文件名, line: 默认语法, void: 默认空白",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				if yac, ok := m.Target().Server.(*YAC); m.Assert(ok) { // {{{
					m.Optioni("page", yac.page[m.Confx("line")])
					m.Optioni("void", yac.page[m.Confx("void")])
					m.Start(m.Confx("name", arg, 1), m.Confx("help", arg, 2), key, arg[0])
				}
				// }}}
			}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void", Help: "void",
			Commands: map[string]*ctx.Command{"parse": &ctx.Command{}},
		},
	},
}

func init() {
	yac := &YAC{}
	yac.Context = Index
	ctx.Index.Register(Index, yac)
}
