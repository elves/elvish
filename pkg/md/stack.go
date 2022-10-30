package md

type stack[T any] []T

func (s *stack[T]) push(v T) {
	*s = append(*s, v)
}

func (s stack[T]) peek() T {
	return s[len(s)-1]
}

func (s *stack[T]) pop() T {
	last := s.peek()
	*s = (*s)[:len(*s)-1]
	return last
}
