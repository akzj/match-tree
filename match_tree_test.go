package match_tree

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMatchTree(t *testing.T) {
	tree := NewMatchTree()

	tree.Insert("a.b.c.d", 1)
	tree.Insert("a.b.c.e", 2)
	tree.Insert("a.b.c.f", 3)
	tree.Insert("a.b.*.f", 4)
	tree.Insert("a.*.*.f", 5)
	tree.Insert("*.*.*.f", 6)

	res := tree.Match("a.b.c.f")
	if res == nil {
		t.Fatalf("match error")
	}
	if !reflect.DeepEqual(res, []interface{}{3, 4, 5, 6}) {
		t.Fatalf("match error %+v", res)
	}

	res = tree.Match("a.b.e.f")
	if res == nil {
		t.Fatalf("match error")
	}
	if !reflect.DeepEqual(res, []interface{}{4, 5, 6}) {
		t.Fatalf("match error %+v", res)
	}
}

func TestMatchTree2(t *testing.T) {
	tree := NewMatchTree()

	tree.Insert("#.e", 2)

	res := tree.Match("c.e")
	if res == nil {
		t.Fatalf("match error")
	}
	fmt.Println(res)
	if !reflect.DeepEqual(res, []interface{}{2}) {
		t.Fatalf("match error %+v", res)
	}
}
