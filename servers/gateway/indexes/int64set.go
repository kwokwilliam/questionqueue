package indexes

import "sort"

//int64set is a set of int64 values
type int64set map[int64]struct{}

//add adds a value to the set and returns
//true if the value didn't already exist in the set.
func (s int64set) add(value int64) bool {
	_, ok := s[value]
	if ok {
		return false
	}
	s[value] = struct{}{}
	return true
}

//remove removes a value from the set and returns
//true if that value was in the set, false otherwise.
func (s int64set) remove(value int64) bool {
	_, ok := s[value]
	if !ok {
		return false
	}
	delete(s, value)
	return true
}

//has returns true if value is in the set,
//or false if it is not in the set.
func (s int64set) has(value int64) bool {
	_, ok := s[value]
	return ok
}

//all returns all values in the set as a slice.
//The returned slice will always be non-nil, but
//the order will be random. Use sort.Slice to
//sort the slice if necessary.
func (s int64set) all() []int64 {
	keys := make([]int64, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}
