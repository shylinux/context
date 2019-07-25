package kit

import (
	"fmt"
	"strings"
)

type Frame struct {
	Key string
	Run bool
	Pos int

	deep  int
	Done  bool
	Data  interface{}
	Hash  map[string]interface{}
	Label map[string]int
}

func (f *Frame) String(meta string) string {
	return fmt.Sprintf("%s%s%d %s %t", strings.Repeat("#", f.deep), meta, f.deep, f.Key, f.Run)
}

var bottom = &Frame{}

type Stack struct {
	Target interface{}
	fs     []*Frame
}

func (s *Stack) Push(key string, run bool, pos int) *Frame {
	s.fs = append(s.fs, &Frame{Key: key, Run: run, Pos: pos, deep: len(s.fs), Hash: map[string]interface{}{}})
	return s.fs[len(s.fs)-1]
}
func (s *Stack) Peek() *Frame {
	if len(s.fs) == 0 {
		return bottom
	}
	return s.fs[len(s.fs)-1]
}
func (s *Stack) Pop() *Frame {
	if len(s.fs) == 0 {
		return bottom
	}
	f := s.fs[len(s.fs)-1]
	s.fs = s.fs[:len(s.fs)-1]
	return f
}
func (s *Stack) Hash(key string, val ...interface{}) (interface{}, bool) {
	for i := len(s.fs) - 1; i >= 0; i-- {
		if v, ok := s.fs[i].Hash[key]; ok {
			if len(val) > 0 {
				s.fs[i].Hash[key] = val[0]
			}
			return v, ok
		}
	}

	if len(val) > 0 {
		s.fs[len(s.fs)-1].Hash[key] = val[0]
		return val[0], true
	}
	return nil, false
}
func (s *Stack) Label(key string) (int, bool) {
	for i := len(s.fs) - 1; i >= 0; i-- {
		if v, ok := s.fs[i].Label[key]; ok {
			s.fs = s.fs[:i+1]
			return v, ok
		}
	}
	return -1, false
}
