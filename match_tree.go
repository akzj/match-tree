package match_tree

import (
	"reflect"
	"strings"
)

type Node struct {
	copyOnWrite *copyOnWrite
	path        string
	token       string
	nextMap     map[string]*Node
	values      []interface{}
}

type copyOnWrite struct {
	_ int
}

func (copyOnWrite *copyOnWrite) mutableNode(node *Node) *Node {
	if copyOnWrite == node.copyOnWrite {
		return node
	}
	copyNode := &Node{
		copyOnWrite: copyOnWrite,
		path:        node.path,
		token:       node.token,
		nextMap:     make(map[string]*Node, len(node.nextMap)),
	}
	for k, v := range node.nextMap {
		copyNode.nextMap[k] = v
	}
	if node.values != nil {
		copyNode.values = append(copyNode.values, node.values...)
	}
	return copyNode
}

type MatchTree struct {
	copyOnWrite *copyOnWrite
	root        *Node
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
		if next.copyOnWrite != node.copyOnWrite {
			next = node.copyOnWrite.mutableNode(next)
			node.nextMap[token] = next
		}
		return next
	}
	next = newNode(token, path, node.copyOnWrite)
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

func (node *Node) match(token, remain string, set map[string][]interface{}) {
	//fmt.Println("token", token, "remain", remain, "node.token", node.token, "node.path", node.path, "node.values", node.values)
	if len(token) == 0 {
		if node.values != nil {
			set[node.path] = node.values
		}
		var last *Node
		for next := node.findNext("#"); next != nil; next = next.findNext("#") {
			last = next
		}
		if last != nil && last.values != nil {
			set[last.path] = last.values
		}
		return
	}
	if next := node.findNext(token); next != nil {
		token, remain := nextToken(remain)
		next.match(token, remain, set)
	}
	if next := node.findNext("*"); next != nil {
		token, remain := nextToken(remain)
		next.match(token, remain, set)
	}
	if next := node.findNext("#"); next != nil {
		next.match(token, remain, set)
	}
	if node.token == "#" {
		if node.nextEmpty() {
			set[node.path] = node.values
		} else {
			token, remain = nextToken(remain)
			node.match(token, remain, set)
		}
	}
}

func (node *Node) findNext(token string) *Node {
	next, _ := node.nextMap[token]
	return next
}

func (node *Node) mutableNext(token string) *Node {
	next, _ := node.nextMap[token]
	if next == nil {
		return nil
	}
	if next.copyOnWrite != node.copyOnWrite {
		next = node.copyOnWrite.mutableNode(next)
		node.nextMap[token] = next
	}
	return next
}

func (node *Node) nextEmpty() bool {
	return len(node.nextMap) == 0
}

func (node *Node) Walk(f func(path string, objs []interface{}) bool) bool {
	for _, next := range node.nextMap {
		if next.values != nil {
			if f(next.path, next.values) == false {
				return false
			}
		}
		if next.Walk(f) == false {
			return false
		}
	}
	return true
}

func newNode(token string, path string, copyOnWrite *copyOnWrite) *Node {
	return &Node{
		token:       token,
		copyOnWrite: copyOnWrite,
		path:        path,
		nextMap:     map[string]*Node{},
	}
}

func NewMatchTree() *MatchTree {
	copyOnWrite := new(copyOnWrite)
	return &MatchTree{
		copyOnWrite: copyOnWrite,
		root:        newNode("", "", copyOnWrite),
	}
}

func (tree *MatchTree) Insert(key string, value interface{}) *MatchTree {
	root := tree.copyOnWrite.mutableNode(tree.root)
	root.insert(strings.Split(key, "."), "", value)
	tree.root = root
	return tree
}

func (tree *MatchTree) Match(key string) []interface{} {
	var set = make(map[string][]interface{})
	token, remain := nextToken(key)
	tree.root.match(token, remain, set)
	var objs []interface{}
	for _, values := range set {
		objs = append(objs, values...)
	}
	return objs
}

func (tree *MatchTree) Clone() *MatchTree {
	var clone = NewMatchTree()
	copyOnWrite := *clone.copyOnWrite
	tree.copyOnWrite = &copyOnWrite
	clone.root = tree.root
	return clone
}

func (tree *MatchTree) Walk(f func(path string, objs []interface{}) bool) {
	tree.root.Walk(f)
}

func objRE(first, second interface{}) bool {
	if reflect.DeepEqual(first, second) {
		return true
	}
	return first == second
}

func (tree *MatchTree) Delete(key string, obj interface{}) {
	root := tree.copyOnWrite.mutableNode(tree.root)
	node := root
	var stack = []*Node{node}
	for token, remain := nextToken(key); len(token) != 0; token, remain = nextToken(remain) {
		node = node.mutableNext(token)
		if node == nil {
			return
		}
		stack = append(stack, node)
	}
	if node.path != key {
		return
	}
	var find = false
	for i, val := range node.values {
		if objRE(val, obj) {
			node.values = append(node.values[:i], node.values[i+1:]...)
			find = true
			if len(node.values) == 0 {
				node.values = nil
			}
			break
		}
	}
	if find == false {
		return
	}
	defer func() {
		tree.root = root
	}()
	if node.values == nil {
		return
	}
	stackPop := func() *Node {
		if len(stack) == 0 {
			return nil
		}
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return node
	}
	_ = stackPop()
	for pre := stackPop(); pre != nil; pre = stackPop() {
		delete(pre.nextMap, node.token)
		if len(pre.nextMap) == 0 && pre.values == nil {
			node = pre
		} else {
			break
		}
	}
}

func nextToken(str string) (string, string) {
	pos := strings.IndexByte(str, '.')
	if pos == -1 {
		return str, ""
	}
	return str[:pos], str[pos+1:]
}
