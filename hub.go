package bee

import (
	"sync"
	"sync/atomic"
)

type Hub interface {
	SetSession(identifier string, s Session)

	GetSession(identifier string) Session

	RemoveSession(identifier string)

	Range(f func(identifier string, s Session))

	Count() int64
}

type hub struct {
	m sync.Map
	c int64
}

func NewHub() Hub {
	var p = &hub{}
	p.m = sync.Map{}
	return p
}

func (this *hub) SetSession(identifier string, s Session) {
	if s != nil {
		this.m.Store(identifier, s)
		atomic.AddInt64(&this.c, 1)
	}
}

func (this *hub) GetSession(identifier string) Session {
	if value, ok := this.m.Load(identifier); ok {
		return value.(Session)
	}
	return nil
}

func (this *hub) RemoveSession(identifier string) {
	if val, ok := this.m.Load(identifier); ok {
		if s, ok := val.(Session); ok {
			s.Conn().Close()
		}
		this.m.Delete(identifier)
		atomic.AddInt64(&this.c, -1)
	}
}

func (this *hub) Range(f func(identifier string, s Session)) {
	this.m.Range(func(key, value interface{}) bool {
		f(key.(string), value.(Session))
		return true
	})
}

func (this *hub) Count() int64 {
	return this.c
}
