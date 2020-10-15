package core

type stackFrame struct {
	Name string
	Args []Any
	Vars map[string]Any
}

type stack struct {
	maxDepth int
	frames   []stackFrame
}

func (s *stack) Push(f stackFrame) {
	if s.maxDepth > 0 && len(s.frames) >= s.maxDepth {
		// stack has exactly maxDepth number of frames and allocating
		// another frame would violate the limit.
		panic("maximum stack depth reached")
	}
	s.frames = append(s.frames, f)
}

func (s *stack) Pop() (f *stackFrame) {
	if len(s.frames) == 0 {
		panic("pop from empty stack")
	}
	f, s.frames = &s.frames[len(s.frames)-1], s.frames[:len(s.frames)-1]
	return f
}

func (s *stack) Size() int { return len(s.frames) }
