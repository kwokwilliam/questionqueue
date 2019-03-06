package trie

import (
	"log"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"
)

type pair struct {
	Key string
	Val int64
}

func TestTrie_Add(t *testing.T) {

	cases := []struct {
		name     string
		pairs    []pair
		query    string
		limit    int
		expected []int64
	}{
		{
			name: "Single ascii pair",
			pairs: []pair{
				{"a", 0}},
			query:    "a",
			limit:    1,
			expected: []int64{0},
		},
		{
			name: "Multiple ascii pairs, high limit",
			pairs: []pair{
				{"aaaaaa", 0},
				{"ab", 1},
				{"abc", 2},
				{"bbbb", 3},
				{"abcccc", 4},
				{"a", 5}},
			query:    "a",
			limit:    6,
			expected: []int64{0, 1, 2, 4, 5},
		},
		{
			name: "Multiple ascii pairs, low limit",
			pairs: []pair{
				{"aaaaaa", 0},
				{"ab", 1},
				{"abc", 2},
				{"bbbb", 3},
				{"abcccc", 4},
				{"a", 5}},
			query:    "a",
			limit:    1,
			expected: []int64{5},
		},
		{
			name: "Single unicode pair",
			pairs: []pair{
				{"😀😁😂", 0}},
			query:    "😀",
			limit:    1,
			expected: []int64{0},
		},
		{
			name: "Multiple unicode pairs, high limit",
			pairs: []pair{
				{"😀😁😂", 0},
				{"😀😁", 1},
				{"🙀😿😾", 2},
				{"中国", 3},
				{"中", 4},
				{"日本語", 5}},
			query:    "😀",
			limit:    10,
			expected: []int64{0, 1},
		},
		{
			name: "Multiple unicode pairs, low limit",
			pairs: []pair{
				{"😀😁😂", 0},
				{"😀😁", 1},
				{"🙀😿😾", 2},
				{"中国", 3},
				{"中", 4},
				{"日本語", 5}},
			query:    "中",
			limit:    1,
			expected: []int64{4},
		},
		{
			name: "Not found",
			pairs: []pair{
				{"😀😁😂", 0},
				{"😀😁", 1},
				{"🙀😿😾", 2},
				{"中国", 3},
				{"中", 4},
				{"日本語", 5}},
			query:    "a",
			limit:    1,
			expected: []int64{},
		},
		{
			name: "Multi-values",
			pairs: []pair{
				{"a", 0},
				{"a", 1},
				{"a", 2},
				{"aa", 3},
				{"aaa", 4},
				{"aaa", 5},
				{"a", 11},
				{"a", 1},},
			query:    "a",
			limit:    10,
			expected: []int64{0, 1, 2, 3, 4, 5, 11},
		},

		{
			name: "Empty query",
			pairs: []pair{
				{"a", 0},
				{"a", 1},
				{"a", 2},
				{"aa", 3},
				{"aaa", 4},
				{"aaa", 5},
				{"a", 11},
				{"a", 1},},
			query:    "",
			limit:    10,
			expected: []int64{},
		},
	}

	for _, c := range cases {
		testTrie := NewTrie()
		for _, p := range c.pairs {
			testTrie.Add(p.Key, p.Val)
		}

		result := testTrie.Find(c.query, c.limit)

		// int64set does not guarantee order of insertion, sort then compare
		sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
		if !reflect.DeepEqual(result, c.expected) && len(c.expected) > 0 {
			t.Errorf("case %s expects %v got %v", c.name, c.expected, result)
		}
	}
}

func TestTrie_Remove(t *testing.T) {
	cases := []struct {
		name        string
		insertPairs []pair
		removePairs []pair
		query       string
		limit       int
		expected    []int64
	}{
		{
			name: "Single ascii pairs",
			insertPairs: []pair{
				{"a", 0},
			},
			removePairs: []pair{
				{"a", 0},
			},
			query:    "a",
			limit:    1,
			expected: []int64{},
		},

		{
			name: "Multiple ascii pairs",
			insertPairs: []pair{
				{"1😀😁😂", 0},
				{"1😀😁", 1},
				{"1🙀😿😾", 2},
				{"1中国", 3},
				{"1中", 4},
				{"1日本語", 5},
			},
			removePairs: []pair{
				{"a", 0},
				{"aa", 1},
				{"aa", 2},
				{"1日本語", 5},
			},
			query:    "1",
			limit:    10,
			expected: []int64{0, 1, 2, 3, 4},
		},

		{
			name: "Multiple ascii pairs",
			insertPairs: []pair{
				{"a", 0},
				{"a", 1},
				{"aa", 2},
				{"aaa", 3},
				{"abc", 4},
			},
			removePairs: []pair{
				{"a", 0},
				{"aa", 1},
				{"aa", 2},
			},
			query:    "a",
			limit:    10,
			expected: []int64{1, 3, 4},
		},

		{
			name: "Empty trie",
			insertPairs: []pair{
			},
			removePairs: []pair{
				{"a", 0},
				{"aa", 1},
				{"aa", 2},
			},
			query:    "a",
			limit:    10,
			expected: []int64{},
		},

		{
			name: "Empty query",
			insertPairs: []pair{
			},
			removePairs: []pair{
				{"a", 0},
				{"aa", 1},
				{"aa", 2},
			},
			query:    "",
			limit:    10,
			expected: []int64{},
		},
	}

	for _, c := range cases {
		testTrie := NewTrie()

		// insertions
		for _, p := range c.insertPairs {
			testTrie.Add(p.Key, p.Val)
		}

		// deletions
		for _, p := range c.removePairs {
			testTrie.Remove(p.Key, p.Val)
		}

		result := testTrie.Find(c.query, c.limit)

		// int64set does not guarantee order of insertion, sort then compare
		sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
		if !reflect.DeepEqual(result, c.expected) && len(c.expected) > 0 {
			t.Errorf("case %s expects %v got %v", c.name, c.expected, result)
		}
	}
}

func BenchmarkTrie(b *testing.B) {
	stressLoad := 1000
	start := time.Now()
	trie := NewTrie()

	for i := 0; i < stressLoad; i++ {
		s := generateRandomString(2000)
		//log.Printf("%v: inserting %v\n", i, s)
		trie.Add(s, int64(i))
	}

	log.Println("inserted", stressLoad, "-", time.Since(start))
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func generateRandomString(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
