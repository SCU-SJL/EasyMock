package easymock

import "sync"

type StringSet struct {
	mu sync.RWMutex
	set map[string]struct{}
}

func CreateStringSet(list []string) *StringSet {
	set := make(map[string]struct{})
	for _, s := range list {
		set[s] = struct{}{}
	}
	return &StringSet{
		mu:  sync.RWMutex{},
		set: set,
	}
}

func (ss *StringSet) Add(s string) {
	ss.mu.Lock()
	ss.set[s] = struct{}{}
	ss.mu.Unlock()
}

func (ss *StringSet) Remove(s string) {
	ss.mu.Lock()
	delete(ss.set, s)
	ss.mu.Unlock()
}

func (ss *StringSet) Contains(s string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	_, ok := ss.set[s]
	return ok
}
