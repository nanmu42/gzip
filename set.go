package gzip

// Set stores distinct strings.
// Original source:
// https://github.com/caddyserver/caddy/blob/7fa90f08aee0861187236b2fbea16b4fa69c5a28/caddyhttp/gzip/requestfilter.go#L76-L105
type Set map[string]struct{}

// Add adds an element to the set.
func (s Set) Add(value string) {
	s[value] = struct{}{}
}

// Remove removes an element from the set.
func (s Set) Remove(value string) {
	delete(s, value)
}

// Contains check if the set contains value.
func (s Set) Contains(value string) bool {
	_, ok := s[value]
	return ok
}

// ContainsFunc is similar to Contains. It iterates all the
// elements in the set and passes each to f. It returns true
// on the first call to f that returns true and false otherwise.
func (s Set) ContainsFunc(f func(string) bool) bool {
	for k := range s {
		if f(k) {
			return true
		}
	}
	return false
}
