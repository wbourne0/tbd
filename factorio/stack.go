package factorio

var _nextId int

func nextId() int {
	_nextId++

	return _nextId
}

type stackItem struct {
	prev *stackItem
	next *stackItem
	ent  Entity
	id   int
}

func (s *stackItem) remove() {
	s.prev.next = s.next
	if s.next != nil {
		s.next.prev = s.prev
	}
}

// note that for this to be safe it needs to be called in reverse the order that it was removed.
// (last removed -> first added back)
func (s *stackItem) reinstate() {
	s.prev.next = s
	if s.next != nil {
		s.next.prev = s
	}
}

func (s *stackItem) iterate(fn func(it *stackItem) (shouldBreak bool)) {
	for i := s.next; i != nil; i = i.next {
		if fn(i) {
			break
		}
	}
}

func (s *stackItem) add(e Entity) *stackItem {
	n := stackItem{
		ent:  e,
		prev: s,
		id:   nextId(),
	}
	s.next = &n

	return &n
}
