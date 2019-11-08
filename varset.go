package sec

type varSet map[string]struct{}

func (v varSet) add(name string) {
	v[name] = struct{}{}
}

func (v varSet) del(name string) {
	delete(v, name)
}

func (v varSet) names() (names []string) {
	names = make([]string, 0, len(v))
	for name := range v {
		names = append(names, name)
	}
	return
}
