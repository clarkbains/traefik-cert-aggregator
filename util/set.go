package util

type Set[T comparable] struct {
	data map[T]bool
}

func NewSet[T comparable]() *Set[T] {
	s := Set[T]{}
	s.data = make(map[T]bool)
	return &s
}

func NewSetFromArray[T comparable](items []T) *Set[T] {
	s := NewSet[T]()
	for _, val := range items {
		s.Add(val)
	}
	return s
}

func (s *Set[T]) Add(item T) {
	s.data[item] = true
}

func (s *Set[T]) Contains(item T) bool {
	_, b := s.data[item]
	return b
}

func (s Set[T]) GetItems() []T {
	keys := make([]T, len(s.data))
	i := 0
	for k := range s.data {
		keys[i] = k
		i++
	}

	return keys
}
