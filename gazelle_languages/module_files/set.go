package module_files

type OrderedSet[T comparable] struct {
	m     map[T]struct{}
	order []T
}

func NewOrderedSet[T comparable]() *OrderedSet[T] {
	return &OrderedSet[T]{
		m:     make(map[T]struct{}),
		order: []T{},
	}
}

func NewOrderedSetFromSlice[T comparable](slice []T) *OrderedSet[T] {
	set := NewOrderedSet[T]()

	for _, entry := range slice {
		set.Add(entry)
	}
	return set
}

func (s *OrderedSet[T]) Add(entry T) {
	if _, exists := s.m[entry]; !exists {
		s.m[entry] = struct{}{}
		s.order = append(s.order, entry)
	}
}

func (s *OrderedSet[T]) Remove(entry T) {
	if _, exists := s.m[entry]; exists {
		delete(s.m, entry)
		// Remove from order slice
		for i, v := range s.order {
			if v == entry {
				s.order = append(s.order[:i], s.order[i+1:]...)
				break
			}
		}
	}
}

func (s *OrderedSet[T]) Contains(entry T) bool {
	_, ok := s.m[entry]
	return ok
}

func (s *OrderedSet[T]) Len() int {
	return len(s.m)
}

func (s *OrderedSet[T]) ToSlice() []T {
	return append([]T(nil), s.order...)
}

func (s *OrderedSet[T]) Range(f func(T)) {
	for _, entry := range s.order {
		f(entry)
	}
}
