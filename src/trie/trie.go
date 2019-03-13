package trie

import (
	"sort"
	"sync"
)

// implement a trie data structure that stores keys of type string and values of type int64
// 1. stores key/value pairs, where the key is a Unicode string and the value is an int64
// (data type for DBMS-assigned primary key values).
// 2. supports keys containing any valid Unicode characters
// 3. key/value pairs remains a distinct set
// 4. has a method that returns the first n values that match a given prefix string
// 5. has a method that removes a given key/value pairs from the trie
// 6. trie protected for concurrent use.
type Node struct {
	key      byte // each character of a username, collectively form the username
	val      int64set
	parent   *Node
	children map[byte]*Node
}

type Trie struct {
	root *Node
	lock sync.RWMutex
}

// generate and return the new pointer of a new trie
func NewTrie() *Trie {
	return &Trie{newNode(), sync.RWMutex{}}
}

func newNode() *Node {
	return &Node{
		key:      0,
		val:      int64set{},
		parent:   nil,
		children: map[byte]*Node{},
	}
}

// add a new key/value pairs into a given tree
func (t *Trie) Add(k string, v int64) {

	t.lock.Lock()
	defer t.lock.Unlock()

	if len(k) == 0 || v < 0 {
		return
	}
	bytes := []byte(k)

	curr := t.root
	for _, b := range bytes {

		if curr.children[b] == nil {
			curr.children[b] = newNode()
		}

		curr.children[b].parent = curr
		curr = curr.children[b]
		curr.key = b
	}

	curr.val.add(v)
}

func (t *Trie) Find(prefix string, i int) []int64 {

	t.lock.Lock()
	defer t.lock.Unlock()

	if t == nil || len(prefix) == 0 || i == 0 {
		return nil
	}

	var result []int64
	// end of prefix
	eop := walk(t.root, []byte(prefix))

	if eop == nil {
		return nil
	}

	collect(eop, &i, &result)
	return result
}

func (t *Trie) Remove(k string, v int64) {

	t.lock.Lock()
	defer t.lock.Unlock()

	//if len(k) == 0 || v < 0 {
	//	return
	//}

	bytes := []byte(k)
	n := walk(t.root, bytes)

	if n == nil {
		return
	}

	if n.val.has(v) {
		n.val.remove(v)
	} else {
		return
	}

	//if n.val.isEmpty() {
	//	n.parent.children[n.key] = nil
	//	n.children = nil
	//}

	if n.val.isEmpty() && len(n.children) == 0 {
		n.children = nil
	}

}

// from a given node, walk down the trie and return the pointer of the node
// where the end of the prefix lies
func walk(n *Node, remains []byte) *Node {
	//if n == nil {
	//	return nil
	//}

	if len(remains) == 0 {
		return n
	}

	b := remains[0]
	_, ok := n.children[b]
	if !ok {
		return nil
	} else {
		return walk(n.children[b], remains[1:])
	}
}

// collect all words below a given `node`, up to `i` words
// results are written to the `result` pointer
func collect(n *Node, i *int, result *[]int64) {
	if n == nil {
		return
	}

	// append existing values
	for v := range n.val {
		if *i == 0 {
			return
		}
		*i--
		*result = append(*result, v)
	}

	// sort keys
	sortedKeys := sortKeys(n.children)
	for _, k := range sortedKeys {
		curr := n.children[k]
		collect(curr, i, result)
	}
}

// sort and return all keys of a given map `m` in a byte slice
func sortKeys(m map[byte]*Node) []byte {
	keys := make([]byte, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}
