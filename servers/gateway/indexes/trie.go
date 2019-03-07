package indexes

import (
	"sort"
	"strings"
	"sync"
)

//Trie implements a trie data structure mapping strings to int64s
//that is safe for concurrent use.
type Trie struct {
	children map[rune]*Trie
	values   int64set
	mx       sync.RWMutex
}

//NewTrie constructs a new Trie.
func NewTrie() *Trie {
	return &Trie{}
}

//Len returns the number of entries in the trie.
func (t *Trie) Len() int {
	t.mx.RLock()
	defer t.mx.RUnlock()
	entryCount := len(t.values)
	for child := range t.children {
		entryCount += t.children[child].Len()
	}
	return entryCount
}

// add is a private helper method that adds a key and value to the trie
func (t *Trie) add(key []rune, value int64) {
	if len(t.children) == 0 {
		t.children = make(map[rune]*Trie)
	}
	if t.children[key[0]] == nil {
		t.children[key[0]] = NewTrie()
	}
	if len(key) == 1 {
		if len(t.children[key[0]].values) == 0 {
			t.children[key[0]].values = make(map[int64]struct{})
		}
		t.children[key[0]].values.add(value)
		return
	}
	t.children[key[0]].add(key[1:len(key)], value)
}

//Add adds a key and value to the trie.
func (t *Trie) Add(key string, value int64) {
	t.mx.Lock()
	defer t.mx.Unlock()
	runes := []rune(key)
	t.add(runes, value)
}

func (t *Trie) findDFS(list *[]int64, max int) {
	// add all current values in node to list (or until hit max)
	values := t.values.all()
	canGet := max - len(*list)
	if len(values) > canGet {
		*list = append(*list, values[0:canGet]...)
		return
	}
	*list = append(*list, values...)
	// if max reached or no children, just return.
	if len(*list) == max || len(t.children) == 0 {
		return
	}

	// sort children
	children := make([]rune, 0, len(t.children))
	for k := range t.children {
		children = append(children, k)
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i] < children[j]
	})

	// for every child, recurse and add to list and check for max
	for _, child := range children {
		t.children[child].findDFS(list, max)
		if len(*list) == max {
			return
		}
	}
	return
}

//Find finds `max` values matching `prefix`. If the trie
//is entirely empty, or the prefix is empty, or max == 0,
//or the prefix is not found, this returns a nil slice.
func (t *Trie) Find(prefix string, max int) []int64 {
	t.mx.RLock()
	defer t.mx.RUnlock()

	if len(t.children) == 0 || prefix == "" || max <= 0 {
		return nil
	}

	// iterate through trie until at end of prefix | O(1)
	prefixRunes := []rune(prefix)
	triePointer := t
	for _, s := range prefixRunes {
		if triePointer.children[s] == nil {
			return nil
		}
		triePointer = triePointer.children[s]
	}
	// create int64 slice
	var returnSlice []int64
	triePointer.findDFS(&returnSlice, max)
	return returnSlice
}

func (t *Trie) remove(key []rune, value int64) {
	if len(key) == 0 {
		t.values.remove(value)
		return
	}
	focusChild := t.children[key[0]]
	focusChild.remove(key[1:], value)
	if len(focusChild.children) == 0 && len(focusChild.values) == 0 {
		delete(t.children, key[0])
	}
}

//Remove removes a key/value pair from the trie
//and trims branches with no values.
func (t *Trie) Remove(key string, value int64) {
	// split key into runes
	t.mx.Lock()
	defer t.mx.Unlock()
	runes := []rune(key)
	t.remove(runes, value)
}

// AddUserToTrie adds a user to the trie
func (t *Trie) AddUserToTrie(username, firstname, lastname string, id int64) {
	t.addSplitToTrie(username, id)
	t.addSplitToTrie(firstname, id)
	t.addSplitToTrie(lastname, id)
}

// RemoveNamesInTrie removes the names in the trie
func (t *Trie) RemoveNamesInTrie(firstname, lastname string, id int64) {
	t.removeSplitToTrie(firstname, id)
	t.removeSplitToTrie(lastname, id)
}

func (t *Trie) addSplitToTrie(s string, id int64) {
	split := strings.Split(s, " ")
	for _, sp := range split {
		t.Add(strings.ToLower(sp), id)
	}
}

func (t *Trie) removeSplitToTrie(s string, id int64) {
	split := strings.Split(s, " ")
	for _, sp := range split {
		t.Remove(strings.ToLower(sp), id)
	}
}
