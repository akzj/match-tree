package match_tree

import (
	"strings"
)

type Node struct {
	path    string
	token   string
	pre     *Node
	nextMap map[string]*Node
	values  []interface{}
}
type MatchTree struct {
	root *Node
}

func (node *Node) appendValue(value interface{}) {
	node.values = append(node.values, value)
}
func (node *Node) RangeNext(f func(next *Node) bool) {
	for _, next := range node.nextMap {
		if f(next) == false {
			return
		}
	}
}

func (node *Node) loadOrCreateNext(token string, path string) *Node {
	next, ok := node.nextMap[token]
	if ok {
		return next
	}
	next = newNode(token, path, node)
	node.nextMap[token] = next
	return next
}

func (node *Node) insert(tokens []string, prefix string, value interface{}) {
	if len(tokens) == 0 {
		node.values = append(node.values, value)
		return
	}
	if len(prefix) > 0 {
		prefix += "."
	}
	prefix += tokens[0]
	node.loadOrCreateNext(tokens[0], prefix).insert(tokens[1:], prefix, value)
}

func (node *Node) mathUniq(tokens []string, set map[string][]interface{}) {
	//fmt.Println("tokens", tokens, "node.token", node.token, "node.path", node.path, "node.values", node.values)
	if len(tokens) == 0 {
		if node.values != nil {
			set[node.path] = node.values
		} else {
			var last *Node
			for next := node.findNext("#"); next != nil; next = next.findNext("#") {
				last = next
			}
			if last != nil && last.values != nil {
				set[last.path] = last.values
			}
		}
		return
	}
	if next := node.findNext(tokens[0]); next != nil {
		next.mathUniq(tokens[1:], set)
	}
	if next := node.findNext("*"); next != nil {
		next.mathUniq(tokens[1:], set)
	}
	if next := node.findNext("#"); next != nil {
		next.mathUniq(tokens[1:], set)
	}
	if node.token == "#" {
		if node.nextEmpty() {
			set[node.path] = node.values
		} else {
			node.mathUniq(tokens[1:], set)
		}
	}
}

func (node *Node) match(tokens []string, objs *[]interface{}) {
	if len(tokens) == 0 {
		*objs = append(*objs, node.values...)
		return
	}
	for _, token := range []string{tokens[0], "*", "#"} {
		if next := node.findNext(token); next != nil {
			next.match(tokens[1:], objs)
		}
	}
	if node.token == "#" {
		if node.nextEmpty() {
			*objs = append(*objs, node.values...)
		} else {
			node.match(tokens[1:], objs)
		}
	}
}

func (node *Node) findNext(token string) *Node {
	next, _ := node.nextMap[token]
	return next
}

func (node *Node) nextEmpty() bool {
	return len(node.nextMap) == 0
}

func newNode(token string, path string, pre *Node) *Node {
	return &Node{
		token:   token,
		pre:     pre,
		path:    path,
		nextMap: map[string]*Node{},
	}
}

func NewMatchTree() *MatchTree {
	return &MatchTree{
		root: newNode("", "", nil),
	}
}

func (tree *MatchTree) Insert(key string, value interface{}) {
	//tree.root.insert(tree.split(key), "", value)
	tree.root.insert(strings.Split(key, "."), "", value)
}

func (tree *MatchTree) split(key string) []string {
	var tmp []string
	for _, token := range strings.Split(key, ".") {
		if len(tmp) == 0 {
			tmp = append(tmp, token)
		} else {
			if token == "#" && tmp[len(tmp)-1] == "#" {
				continue
			}
			tmp = append(tmp, token)
		}
	}
	return tmp
}

func (tree *MatchTree) Match(key string) []interface{} {
	var res []interface{}
	tree.root.match(tree.split(key), &res)
	return res
}

func (tree *MatchTree) MatchUniq(key string) []interface{} {
	var set = make(map[string][]interface{})
	tree.root.mathUniq(tree.split(key), set)
	var objs []interface{}
	for _, values := range set {
		objs = append(objs, values...)
	}
	return objs
}

func (tree *MatchTree) MatchTokenUniq(tokens []string, set map[string][]interface{}) {
	tree.root.mathUniq(tokens, set)
}
