package match_tree

import "strings"

type Node struct {
	token  string
	pre    *Node
	next   map[string]*Node
	values []interface{}
}
type MatchTree struct {
	root *Node
}

func (node *Node) appendValue(value interface{}) {
	node.values = append(node.values, value)
}
func (node *Node) RangeNext(f func(next *Node) bool) {
	for _, next := range node.next {
		if f(next) == false {
			return
		}
	}
}

func (node *Node) loadOrCreateNext(token string) *Node {
	next, ok := node.next[token]
	if ok {
		return next
	}
	next = newNode(token, node)
	node.next[token] = next
	return next
}

func (node *Node) insert(tokens []string, value interface{}) {
	if len(tokens) == 0 {
		node.values = append(node.values, value)
		return
	}
	node.loadOrCreateNext(tokens[0]).insert(tokens[1:], value)
}

func (node *Node) match(tokens []string) []interface{} {
	if len(tokens) == 0 {
		return node.values
	}
	var values []interface{}
	if next := node.findNext(tokens[0]); next != nil {
		if res := next.match(tokens[1:]); res != nil {
			values = append(values, res...)
		}
	}
	if next := node.findNext("*"); next != nil {
		if res := next.match(tokens[1:]); res != nil {
			values = append(values, res...)
		}
	}
	if next := node.findNext("#"); next != nil {
		if res := next.match(tokens[1:]); res != nil {
			values = append(values, res...)
		}
	}
	return values
}

func (node *Node) findNext(token string) *Node {
	next, _ := node.next[token]
	return next
}

func newNode(token string, pre *Node) *Node {
	return &Node{
		token: token,
		pre:   pre,
		next:  map[string]*Node{},
	}
}

func NewMatchTree() *MatchTree {
	return &MatchTree{
		root: newNode("", nil),
	}
}

func (tree *MatchTree) Insert(key string, value interface{}) {
	tree.root.insert(strings.Split(key, "."), value)
}

func (tree *MatchTree) Match(key string) []interface{} {
	return tree.root.match(strings.Split(key, "."))
}

