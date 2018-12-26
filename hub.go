package bee

import (
	"sync"
	"sync/atomic"
)

// --------------------------------------------------------------------------------
type Hub interface {
	AddSession(s Session)

	GetSession(identifier, tag string) Session

	GetSessions(identifier string) []Session

	GetAllSessions() []Session

	RemoveSession(s Session)

	RemoveSessions(identifier string)

	Count() int64
}

// --------------------------------------------------------------------------------
type hub struct {
	mu sync.RWMutex
	m  map[string]map[string]Session
	c  int64
}

func NewHub() Hub {
	var h = &hub{}
	h.m = make(map[string]map[string]Session)
	return h
}

func (this *hub) AddSession(s Session) {
	if s != nil {
		this.mu.Lock()
		defer this.mu.Unlock()

		atomic.AddInt64(&this.c, 1)
		var sm = this.m[s.Identifier()]

		if sm == nil {
			sm = make(map[string]Session)
			this.m[s.Identifier()] = sm
		}
		sm[s.Tag()] = s
	}
}

func (this *hub) GetSession(identifier, tag string) Session {
	this.mu.RLock()
	defer this.mu.RUnlock()

	var sm = this.m[identifier]
	if sm != nil {
		var c = sm[tag]
		return c
	}
	return nil
}

func (this *hub) GetSessions(identifier string) []Session {
	this.mu.RLock()
	defer this.mu.RUnlock()

	var sm = this.m[identifier]
	if sm != nil {
		var sl = make([]Session, 0, len(sm))
		for _, s := range sm {
			sl = append(sl, s)
		}
		return sl
	}
	return nil
}

func (this *hub) GetAllSessions() []Session {
	this.mu.RLock()
	defer this.mu.RUnlock()

	var sl = make([]Session, 0, len(this.m))
	for _, cm := range this.m {
		for _, c := range cm {
			sl = append(sl, c)
		}
	}
	return sl
}

func (this *hub) RemoveSession(s Session) {
	if s != nil {
		this.mu.Lock()
		defer this.mu.Unlock()

		var sm = this.m[s.Identifier()]
		if sm != nil {
			var c = sm[s.Tag()]
			if c != nil {
				delete(sm, c.Tag())
				atomic.AddInt64(&this.c, -1)
			}
		}
	}
}

func (this *hub) RemoveSessions(identifier string) {
	this.mu.Lock()
	defer this.mu.Unlock()

	var sm = this.m[identifier]
	if sm != nil {
		delete(this.m, identifier)
		atomic.AddInt64(&this.c, -int64(len(sm)))
	}
}

func (this *hub) Count() int64 {
	return atomic.LoadInt64(&this.c)
}
