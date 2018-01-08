package lex

import (
	"context"
	"fmt"
)

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
	seed []*Seed
	page map[string]int
	hash map[string]int

	state map[State]*State
	mat   []map[byte]*State

	*ctx.Message
	*ctx.Context
}

func (lex *LEX) train(seed []byte, arg ...string) {
	cell, page, hash := 128, 1, 1
	if len(arg) > 0 {
		if x, ok := lex.hash[arg[0]]; ok {
			hash = x
		} else {
			hash = lex.Capi("nhash", 1)
			lex.hash[arg[0]] = hash
		}
	}
	if len(arg) > 1 {
		if x, ok := lex.page[arg[1]]; ok {
			page = x
		} else {
			lex.mat = append(lex.mat, make(map[byte]*State))
			page = lex.Capi("nline", 1)
			lex.page[arg[1]] = page
			lex.Capi("npage", 1)
		}
	}
	lex.Log("debug", nil, "%d %d %v", page, hash, seed)
	lex.seed = append(lex.seed, &Seed{page, hash, string(seed)})
	lex.Capi("nseed", 1)
	lex.Cap("stream", fmt.Sprintf("%s,%s,%s", lex.Cap("nseed"), lex.Cap("npage"), lex.Cap("nhash")))

	s := []int{page}
	c := make([]byte, 0, cell)
	sn := make([]bool, len(lex.mat))
	cn := make([]bool, cell)

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
			for i := 0; i < cell; i++ {
				c = append(c, byte(i))
			}
		case '\\':
			p++
			fallthrough
		default:
			c = append(c, seed[p])
		}

		lex.Log("debug", nil, "page: \033[31m%v\033[0m", s)
		lex.Log("debug", nil, "cell: \033[32m%v\033[0m", c)

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
				if state == nil {
					state = new(State)
					lex.Capi("nnode", 1)
				}
				lex.Log("debug", nil, "GET(%d,%d): %v", s[i], c[j], state)

				switch flag {
				case '+':
					state.star = true
				case '*':
					state.star = true
					fallthrough
				case '?':
					if sn[s[i]] = true; p == len(seed)-1 {
						for _, n := range ends {
							if n.next == s[i] && n.hash == 0 {
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
						if line == 0 {
							line = len(lex.mat)
							sn = append(sn, false)
							lex.mat = append(lex.mat, make(map[byte]*State))
							lex.Capi("nline", 1)
						}
						state.next = line
					}
					sn[state.next] = true
				}

				if s, ok := lex.state[*state]; ok {
					state = s
				}

				lex.state[*state] = state
				lex.mat[s[i]][c[j]] = state

				lex.Log("debug", nil, "SET(%d,%d): %v", s[i], c[j], state)
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
}

func (lex *LEX) parse(line []byte, arg ...string) (word []byte, hash int, rest []byte) {
	page, begin, end := 1, 0, 0
	if len(arg) > 0 {
		if x, ok := lex.page[arg[0]]; ok {
			page = x
		} else {
			return line, 0, nil
		}
	}

	for star, s, i := 0, page, 0; s != 0 && i < len(line); i++ {

		c := line[i]
		if c == '\\' && i < len(line)-1 {
			c = 'a'
			end++
			i++
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

		if state, ok := lex.mat[star][c]; star == 0 || !ok || state == nil || !state.star {
			star = 0
		}
		if end++; state.star {
			star = s
		}
		if s, hash = state.next, state.hash; s == 0 {
			s, star = star, 0
		}
	}

	if hash == 0 {
		begin, end = 0, 0
	}

	word, rest = line[begin:end], line[end:]
	lex.Log("debug", nil, "\033[31m[%v]\033[0m %d [%v]", string(word), hash, string(rest))
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
	lex.Message = m
	lex.Caches["nseed"] = &ctx.Cache{Name: "种子数量", Value: "0", Help: "种子数量"}
	lex.Caches["npage"] = &ctx.Cache{Name: "集合数量", Value: "1", Help: "集合数量"}
	lex.Caches["nhash"] = &ctx.Cache{Name: "类型数量", Value: "1", Help: "类型数量"}

	lex.Caches["nline"] = &ctx.Cache{Name: "状态数量", Value: "1", Help: "状态数量"}
	lex.Caches["nnode"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "节点数量"}
	lex.Caches["npush"] = &ctx.Cache{Name: "节点数量", Value: "0", Help: "节点数量", Hand: func(m *ctx.Message, x *ctx.Cache, arg ...string) string {
		return fmt.Sprintf("%d", len(m.Target().Server.(*LEX).state))
	}}

	return lex
}

func (lex *LEX) Start(m *ctx.Message, arg ...string) bool {
	lex.seed = make([]*Seed, 0, 10)
	lex.page = map[string]int{"nil": 0}
	lex.hash = map[string]int{"nil": 0}

	lex.state = make(map[State]*State)
	lex.mat = make([]map[byte]*State, 2, 10)
	for i := 0; i < len(lex.mat); i++ {
		lex.mat[i] = make(map[byte]*State)
	}

	lex.Message = m
	return false
}

func (lex *LEX) Close(m *ctx.Message, arg ...string) bool {
	switch lex.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

var Index = &ctx.Context{Name: "lex", Help: "词法解析",
	Caches:  map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{
		"train": &ctx.Command{Name: "train seed [hash [page]", Help: "添加词法规则", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			lex, ok := m.Target().Server.(*LEX)
			m.Assert(ok, "模块类型错误")
			m.Assert(len(arg) > 0, "参数错误")

			lex.train([]byte(arg[0]), arg[1:]...)
		}},
		"parse": &ctx.Command{Name: "parse line [page]", Help: "解析单词", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			lex, ok := m.Target().Server.(*LEX)
			m.Assert(ok, "模块类型错误")
			m.Assert(len(arg) > 0, "参数错误")

			word, hash, rest := lex.parse([]byte(arg[0]), arg[1:]...)
			m.Add("result", string(word), fmt.Sprintf("%d", hash), string(rest))
		}},
		"split": &ctx.Command{Name: "split line page1 [page2]", Help: "分割语句", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			lex, ok := m.Target().Server.(*LEX)
			m.Assert(ok, "模块类型错误")
			m.Assert(len(arg) > 1, "参数错误")

			for line := arg[0]; len(line) > 0; {
				word, hash, rest := lex.parse([]byte(line), arg[1:]...)
				line = string(rest)
				word, hash, rest = lex.parse([]byte(line), arg[2:]...)
				line = string(rest)

				if hash == 0 {
					break
				}

				if len(word) > 1 {
					switch word[0] {
					case '"', '\'':
						word = word[1 : len(word)-1]
					}
				}

				m.Add("result", string(word))
			}

		}},
		"cache": &ctx.Command{Name: "cache", Help: "显示缓存", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			lex, ok := m.Target().Server.(*LEX)
			m.Assert(ok, "模块类型错误")
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
		}},
	},
	Index: map[string]*ctx.Context{
		"void": &ctx.Context{Name: "void",
			Commands: map[string]*ctx.Command{"split": &ctx.Command{}},
		},
	},
}

func init() {
	lex := &LEX{}
	lex.Context = Index
	ctx.Index.Register(Index, lex)
}
