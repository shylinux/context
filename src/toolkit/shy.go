package kit

import (
	"fmt"
	"strings"
)

type Frame struct {
	Key  string
	Run  bool
	Pos  int
	deep int
	// list  []string
	Done bool
	Data interface{}
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
	s.fs = append(s.fs, &Frame{Key: key, Run: run, Pos: pos, deep: len(s.fs)})
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
