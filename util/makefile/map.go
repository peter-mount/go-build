package makefile

type Map struct {
	parent Builder
	m      map[string]Builder
}

type MapHandler func(Builder) Builder

func NewMap(parent Builder) Map {
	return Map{
		parent: parent,
		m:      make(map[string]Builder),
	}
}

func (m *Map) Get(k string) Builder {
	return m.m[k]
}

func (m *Map) Contains(k string) bool {
	return m.m[k] != nil
}

func (m *Map) Add(k string, f MapHandler) Builder {
	n := f(m.parent)
	m.m[k] = n
	return n
}

func (m *Map) AddIfNotPresent(k string, f MapHandler) Builder {
	if !m.Contains(k) {
		return m.Add(k, f)
	}
	return m.Get(k)
}
