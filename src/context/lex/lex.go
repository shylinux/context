package lex // {{{
// }}}
import ( // {{{
	"context"
	"fmt"
	"strconv"
)

// }}}

type State struct {
	star bool
	next int
	hash int
}

type Seed struct {
	page int
	hash int
	word string
}

type LEX struct {
	page int
	cell int

	seed []*Seed

	mat []map[byte]*State
	M   *ctx.Message
	*ctx.Context
}

func (lex *LEX) train(page int, hash int, seed []byte) { // {{{

	s := []int{page}
	c := make([]byte, 0, lex.cell)
	cn := make([]bool, lex.cell)
	sn := make([]bool, len(lex.mat))
	ends := make([]*State, 0, len(seed))

	for p := 0; p < len(seed); p++ {
		switch seed[p] {
		case '[':
			p++
			set := true
			if seed[p] == '^' {
				set = false
				p++
			}

			for ; seed[p] != ']'; p++ {
				if seed[p] == '\\' {
					p++
					cn[seed[p]] = true
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
			for i := 0; i < lex.cell; i++ {
				c = append(c, byte(i))
			}
		case '\\':
			p++
			fallthrough
		default:
			c = append(c, seed[p])
		}

		lex.M.Log("debug", "page: %v", s)
		lex.M.Log("debug", "cell: %v", c)

		flag := '\000'
		if p+1 < len(seed) {
			flag = rune(seed[p+1])
			switch flag {
			case '+', '*', '?':
				p++
			}
		}

		for i := 0; i < len(s); i++ {
			line := 0
			for j := byte(0); int(j) < len(c); j++ {
				si := s[i]

				begin, end := j, j+1

				if false && flag == '+' {
					if lex.mat[si][c[j]] == nil {
						lex.mat[si][c[j]] = new(State)
					}
					state := lex.mat[si][c[j]]
					lex.M.Log("debug", "GET(%d,%d) state:%v", si, c[j], state)

					if state.next == 0 {
						sn = append(sn, false)
						state.next = len(lex.mat)
						lex.mat = append(lex.mat, make(map[byte]*State))
					}
					if p == len(seed)-1 {
						state.hash = hash
					}
					ends = append(ends, state)
					lex.M.Log("debug", "SET(%d,%d) state:%v", si, c[j], state)

					si = state.next
					begin, end = byte(0), byte(len(c))
				}

				next := true

				for j := begin; j < end; j++ {

					if lex.mat[si][c[j]] == nil {
						lex.mat[si][c[j]] = new(State)
					}
					state := lex.mat[si][c[j]]
					lex.M.Log("debug", "GET(%d,%d) state:%v", si, c[j], state)

					switch flag {
					case '+', '*':
						state.star = true
						fallthrough
					case '?':
						sn[si] = true
						if p < len(seed)-1 {
							break
						}

						for _, s := range ends {
							if s.next == si && s.hash == 0 {
								lex.M.Log("debug", "GET() state:%v", s)
								s.hash = hash
								lex.M.Log("debug", "END() state:%v", s)
							}
						}
						fallthrough
					case '\000':
						next = false
					}

					if next {
						if state.next == 0 {
							if line == 0 {
								sn = append(sn, false)
								line = len(lex.mat)
								lex.mat = append(lex.mat, make(map[byte]*State))
							}
							state.next = line
						}
						sn[state.next] = true
					} else {
						state.hash = hash
					}
					ends = append(ends, state)
					lex.M.Log("debug", "SET(%d,%d) state:%v", si, c[j], state)
				}
			}
		}

		c = c[:0]
		s = s[:0]
		for i := 0; i < len(sn); i++ {
			if sn[i] {
				s = append(s, i)
			}
			sn[i] = false
		}
	}
}

// }}}
func (lex *LEX) parse(page int, line []byte) (word []byte, hash int, rest []byte) { // {{{

	s := page
	star := 0

	begin, end := 0, 0

	for i := 0; s != 0 && i < len(line); i++ {
		c := line[i]
		if c == '\\' && i < len(line)-1 {
			c = 'a'
			end++
			i++
		}

		state := lex.mat[s][c]
		lex.M.Log("debug", "(%d,%d): %v", s, c, state)
		if state == nil && star != 0 {
			s, star = star, 0
			state = lex.mat[s][c]
			lex.M.Log("debug", "(%d,%d): %v", s, c, state)
		}
		if state == nil {
			break
		}

		if state, ok := lex.mat[star][c]; star == 0 || !ok || state == nil || !state.star {
			star = 0
		}

		end++
		hash = state.hash
		if state.star {
			star = s
		}

		s = state.next
		if s == 0 {
			s, star = star, 0
		}
	}

	if hash == 0 {
		begin, end = 0, 0
	}

	word = line[begin:end]
	rest = line[end:]
	lex.M.Log("debug", "%d %v %v", hash, word, rest)

	return
}

// }}}

func (lex *LEX) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	lex.Configs["page"] = &ctx.Config{Name: "词法集合", Value: "16", Help: "词法集合"}
	lex.Configs["cell"] = &ctx.Config{Name: "字符集合", Value: "128", Help: "字符集合"}

	if len(arg) > 0 {
		lex.Configs["page"].Value = arg[0]
	}
	if len(arg) > 1 {
		lex.Configs["cell"].Value = arg[1]
	}

	return lex
}

// }}}
func (lex *LEX) Start(m *ctx.Message, arg ...string) bool { // {{{
	lex.M = m

	lex.page = m.Confi("page")
	lex.cell = m.Confi("cell")

	lex.mat = make([]map[byte]*State, m.Confi("page"))

	for i := 0; i < len(lex.mat); i++ {
		lex.mat[i] = make(map[byte]*State)
	}
	lex.seed = make([]*Seed, 0, 10)

	return true
}

// }}}
func (lex *LEX) Spawn(c *ctx.Context, m *ctx.Message, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(LEX)
	s.Context = c
	return s
}

// }}}
func (lex *LEX) Exit(m *ctx.Message, arg ...string) bool { // {{{
	return true
}

// }}}

var Index = &ctx.Context{Name: "lex", Help: "词法解析",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"page": &ctx.Config{Name: "词法集合", Value: "16", Help: "词法集合"},
		"cell": &ctx.Config{Name: "字符集合", Value: "128", Help: "字符集合"},
	},
	Commands: map[string]*ctx.Command{
		"train": &ctx.Command{Name: "train seed [hash [page]", Help: "添加词法规则", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			lex, ok := m.Target.Server.(*LEX) // {{{
			if !ok {
				return ""
			}
			hash := 1
			if len(arg) > 1 {
				hash, _ = strconv.Atoi(arg[1])
			}
			page := 1
			if len(arg) > 2 {
				page, _ = strconv.Atoi(arg[2])
			}
			lex.train(page, hash, []byte(arg[0]))
			lex.seed = append(lex.seed, &Seed{page, hash, arg[0]})

			return ""
			// }}}
		}},
		"parse": &ctx.Command{Name: "parse line [page]", Help: "解析单词", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			lex, ok := m.Target.Server.(*LEX) // {{{
			if !ok {
				return ""
			}

			page := 1
			if len(arg) > 1 {
				page, _ = strconv.Atoi(arg[1])
			}

			word, hash, rest := lex.parse(page, []byte(arg[0]))
			m.Echo(string(word))
			m.Echo(fmt.Sprintf("%d", hash))
			m.Echo(string(rest))
			m.Log("debug", "%s %d %s", string(word), hash, string(rest))
			return ""
			// }}}
		}},
		"split": &ctx.Command{Name: "split line [page1 [page2]]", Help: "分割语句", Hand: func(c *ctx.Context, m *ctx.Message, key string, arg ...string) string {
			lex, ok := m.Target.Server.(*LEX) // {{{
			if !ok {
				return ""
			}

			line := arg[0]
			page1 := 1
			page2 := 2
			if len(arg) > 1 {
				page1, _ = strconv.Atoi(arg[1])
			}
			if len(arg) > 2 {
				page2, _ = strconv.Atoi(arg[2])
			}

			for len(line) > 0 {
				word, hash, rest := lex.parse(page1, []byte(line))
				m.Log("debug", "\033[31mvoid [%s]\033[0m\n", string(word))
				line = string(rest)

				word, hash, rest = lex.parse(page2, []byte(line))
				m.Log("debug", "\033[31mword [%s]\033[0m\n", string(word))
				if hash == 0 {
					break
				}
				m.Echo(string(word))
				line = string(rest)
			}
			return ""
			// }}}
		}},
	},
}

func init() {
	lex := &LEX{}
	lex.Context = Index
	ctx.Index.Register(Index, lex)
}
