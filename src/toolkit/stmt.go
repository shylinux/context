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
}

func (f *Frame) String(meta string) string {
	return fmt.Sprintf("%s%s%d %s %t", strings.Repeat("#", f.deep), meta, f.deep, f.Key, f.Run)
}

type Stack struct {
	fs []*Frame
}

func (s *Stack) Pop() *Frame {
	f := s.fs[len(s.fs)-1]
	s.fs = s.fs[:len(s.fs)-1]
	return f
}
func (s *Stack) Push(key string, run bool, pos int) *Frame {
	s.fs = append(s.fs, &Frame{key, run, pos, len(s.fs)})
	return s.fs[len(s.fs)-1]
}
func (s *Stack) Peek() *Frame {
	return s.fs[len(s.fs)-1]
}
