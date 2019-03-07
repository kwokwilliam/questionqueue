package indexes

import (
	"reflect"
	"testing"
)

func TestAddAndLen(t *testing.T) {
	cases := []struct {
		name           string
		addKeys        []string
		addValues      []int64
		expectedLength int
	}{
		{
			"Simple case",
			[]string{"abcde"},
			[]int64{111},
			1,
		},
		{
			"Two values",
			[]string{"abcde", "cdf"},
			[]int64{111, 23},
			2,
		},
		{
			"Multiple values on one",
			[]string{"abcde", "abcde"},
			[]int64{111, 123},
			2,
		},
		{
			"One on the child value",
			[]string{"abcde", "abcdef"},
			[]int64{111, 123},
			2,
		},
		{
			"Branching",
			[]string{"abcde", "adefda"},
			[]int64{111, 123},
			2,
		},
		{
			"Zero case",
			[]string{},
			[]int64{},
			0,
		},
		{
			"Fancy characters",
			[]string{"𝖙𝖊𝖘𝖙", "𝖜𝖔𝖜", "ɥɔnɯ", "🄵🄰🄽🄲🅈"},
			[]int64{1, 2, 3, 4},
			4,
		},
		{
			"The ultimate emoji test",
			[]string{"💩💩💩💩👀👀👀👀👀💕💕❤❤🙌🙌👍👍💋👏🤞👏👏👏👏😜"},
			[]int64{1},
			1,
		},
		{
			"The ultimate emoji test optimization test",
			[]string{
				"💩💩💩💩👀👀👀👀👀💕💕❤❤🙌🙌👍👍💋👏🤞👏👏👏👏😜",
				"💩💩😘🎉🤦‍♂️🤷‍♂️🌹✌✊🤷‍♂️👀👀👏🤞✊👍🙌❤😂💦💩🤷‍♀️👀😘",
				"👏🤞✊👍🙌❤😂💦💩🤷‍♀️👀😘👍👍🤦‍♂️🤷‍♂️🌹✌✊🤷‍♂️👀🤦‍♂️🤷‍♂️🌹✌✊🤷‍♂️👀",
				"💦💦💩🤷‍♀️👀😘👍♂️🤷‍♂️🌹✌✊🤷‍♂️👀👀💦💩🤷‍♀️👀😘👍😜",
				"💦😘🎉🤦‍♂️🤷‍♂️🌹✌✊",
				"🤷‍💩💩💩👀♀️👀😘👍♂️🤷‍♂️🌹🤞👏👏👏👏😜",
				"🤷‍♂️💩♀️👀😘👍♂️🤷‍♂️🌹👍👍💋👏🤞👏👏👏👏😜",
				"😂♂️💩💩👀👀👀👀👀💕💕♀️👀😘👍♂️🤷‍♂️🌹👏👏👏😜",
				"✌💩💩💩👀👀👀👀👀💕💕❤❤♂️👀👀💦💩🤷‍♀️👀😘👏👏😜"},
			[]int64{1, 2, 3, 4, 5, 6, 7, 8, 9},
			9,
		},
	}

	for _, c := range cases {
		trie := NewTrie()
		for i := range c.addKeys {
			trie.Add(c.addKeys[i], c.addValues[i])
		}
		trieLen := trie.Len()
		if trieLen != c.expectedLength {
			t.Errorf("[%v] Expected lengths not equal. Expected [%v] but got [%v]", c.name, c.expectedLength, trieLen)
		}
	}
}

func TestAddAndFind(t *testing.T) {
	cases := []struct {
		name          string
		addKeys       []string
		addValues     []int64
		prefix        string
		max           int
		expectedSlice []int64
	}{
		{
			"Basic case",
			[]string{"abc"},
			[]int64{1},
			"abc",
			1,
			[]int64{1},
		},
		{
			"Two values on the same short prefix",
			[]string{"abc", "abc"},
			[]int64{1, 2},
			"abc",
			2,
			[]int64{1, 2},
		},
		{
			"Two values on the same prefix path",
			[]string{"abc", "abcde"},
			[]int64{1, 2},
			"abc",
			2,
			[]int64{1, 2},
		},
		{
			"Multiple values on separate paths",
			[]string{"abc", "ace", "adf", "aeg"},
			[]int64{1, 2, 3, 4},
			"a",
			4,
			[]int64{1, 2, 3, 4},
		},
		{
			"Multiple values on separate paths only two max",
			[]string{"abc", "ace", "adf", "aeg"},
			[]int64{1, 2, 3, 4},
			"a",
			2,
			[]int64{1, 2},
		},
		{
			"Multiple values on separate paths try more than max",
			[]string{"abc", "ace", "adf", "aeg"},
			[]int64{1, 2, 3, 4},
			"a",
			6,
			[]int64{1, 2, 3, 4},
		},
		{
			"Empty trie",
			[]string{},
			[]int64{},
			"a",
			6,
			nil,
		},
		{
			"Empty prefix",
			[]string{"abc", "ace", "adf", "aeg"},
			[]int64{1, 2, 3, 4},
			"",
			6,
			nil,
		},
		{
			"Max == 0",
			[]string{"abc", "ace", "adf", "aeg"},
			[]int64{1, 2, 3, 4},
			"a",
			0,
			nil,
		},
		{
			"Cant find path",
			[]string{"abc", "ace", "adf", "aeg"},
			[]int64{1, 2, 3, 4},
			"b",
			6,
			nil,
		},
		{
			"Multiple on the same prefix, but stop at max",
			[]string{"abc", "abc", "abc"},
			[]int64{1, 2, 3},
			"abc",
			2,
			[]int64{1, 2},
		},
		{
			"Real world test",
			[]string{"william", "kwok", "asdfasdf", "misc1", "fdsafdsa", "fdsafdsa"},
			[]int64{1, 1, 1, 2, 2, 2},
			"w",
			20,
			[]int64{1},
		},
	}

	for _, c := range cases {
		trie := NewTrie()
		for i := range c.addKeys {
			trie.Add(c.addKeys[i], c.addValues[i])
		}
		returnSlice := trie.Find(c.prefix, c.max)
		if !reflect.DeepEqual(returnSlice, c.expectedSlice) {
			t.Errorf("[%v] Expected slices not equal. Expected [%v] but got [%v]", c.name, c.expectedSlice, returnSlice)
		}
	}
}

func TestAddAndRemove(t *testing.T) {
	cases := []struct {
		name             string
		addKeys          []string
		addValues        []int64
		removeKey        string
		removeVal        int64
		expecStopRune    rune
		expecChildrenLen int
		expecValLen      int
	}{
		{
			"Simple case",
			[]string{"abc", "abcd"},
			[]int64{1, 2},
			"abcd",
			2,
			[]rune("c")[0],
			0,
			1,
		},
		{
			"Complex case",
			[]string{"abab", "abac", "ace", "abacad", "abacbe", "abacdf"},
			[]int64{1, 2, 3, 4, 5, 6},
			"abacdf",
			6,
			[]rune("c")[0],
			2,
			1,
		},
		{
			"Complex case multiple values in one",
			[]string{"abab", "abac", "ace", "abacad", "abacbe", "abacdf", "abacdf"},
			[]int64{1, 2, 3, 4, 5, 6, 7},
			"abacdf",
			6,
			[]rune("f")[0],
			0,
			1,
		},
		{
			"Remove values in node with children under it",
			[]string{"abab", "abac", "ace", "abacad", "abacbe", "abacdf"},
			[]int64{1, 2, 3, 4, 5, 6},
			"abac",
			2,
			[]rune("c")[0],
			3,
			0,
		},
	}

	for _, c := range cases {
		trie := NewTrie()
		for i := range c.addKeys {
			trie.Add(c.addKeys[i], c.addValues[i])
		}
		trie.Remove(c.removeKey, c.removeVal)

		nodes := []rune(c.removeKey)
		triePointer := trie
		var stopNode rune
		for _, s := range nodes {
			if triePointer.children[s] != nil {
				triePointer = triePointer.children[s]
				stopNode = s
			} else {
				break
			}
		}
		if stopNode != c.expecStopRune {
			t.Errorf("[%v] Unexpected stop rune. Expected [%v] but got [%v]", c.name, stopNode, c.expecStopRune)
		}
		if len(triePointer.children) != c.expecChildrenLen {
			t.Errorf("[%v] Unexpected children length. Expected [%v] but got [%v]", c.name, c.expecChildrenLen, len(triePointer.children))
		}
		if len(triePointer.values) != c.expecValLen {
			t.Errorf("[%v] Unexpected value length. Expected [%v] but got [%v]", c.name, c.expecValLen, len(triePointer.values))
		}
	}
}

func TestIRL(t *testing.T) {
	addKeys := []string{"william", "kwok", "wkwok16", "user2", "name1", "name2"}
	addVals := []int64{1, 1, 1, 2, 2, 2}
	trie := NewTrie()
	for i := range addKeys {
		trie.Add(addKeys[i], addVals[i])
	}
	findFirst := trie.Find("name", 20)
	if len(findFirst) != 2 {
		t.Errorf("error 1")
	}
	trie.RemoveNamesInTrie("name1", "name2", 2)
	if trie.Len() != 4 {
		t.Errorf("error 2")
	}
	trie.AddUserToTrie("name", "name12", 2)
	if trie.Len() != 6 {
		t.Errorf("error 3")
	}
}
