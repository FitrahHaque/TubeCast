package rss

// Global Variables
var STATION_BASE string = "../stations"
var StationNames Set[string]

type Set[T comparable] struct {
	set map[T]struct{}
}

// provide key, a comparable type to create a set
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		set: make(map[T]struct{}),
	}
}

func (s *Set[T]) Add(item T) {
	s.set[item] = struct{}{}
}

func (s *Set[T]) Remove(item T) {
	delete(s.set, item)
}

func (s *Set[T]) Has(item T) bool {
	_, ok := s.set[item]
	return ok
}
