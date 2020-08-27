package match_tree

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

func sortIntObjs(objs []interface{}) []interface{} {
	sort.Slice(objs, func(i, j int) bool {
		return objs[i].(int) < objs[j].(int)
	})
	return objs
}
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
	sortIntObjs(res)
	if !reflect.DeepEqual(res, []interface{}{3, 4, 5, 6}) {
		t.Fatalf("match error %+v", res)
	}

	res = tree.Match("a.b.e.f")
	sortIntObjs(res)
	if res == nil {
		t.Fatalf("match error")
	}
	if !reflect.DeepEqual(res, []interface{}{4, 5, 6}) {
		t.Fatalf("match error %+v", res)
	}
}

func buildRegexp(routingKey string) (*regexp.Regexp, error) {
	routingKey = strings.TrimSpace(routingKey)
	routingParts := strings.Split(routingKey, ".")

	for idx, routingPart := range routingParts {
		if routingPart == "*" {
			routingParts[idx] = "*"
		} else if routingPart == "#" {
			routingParts[idx] = "#"
		} else {
			routingParts[idx] = regexp.QuoteMeta(routingPart)
		}
	}

	routingKey = strings.Join(routingParts, "\\.")
	routingKey = strings.Replace(routingKey, "*", `([^\.]+)`, -1)

	for strings.HasPrefix(routingKey, "#\\.") {
		routingKey = strings.TrimPrefix(routingKey, "#\\.")
		if strings.HasPrefix(routingKey, "#\\.") {
			continue
		}
		routingKey = `(.*\.?)+` + routingKey
	}

	for strings.HasSuffix(routingKey, "\\.#") {
		routingKey = strings.TrimSuffix(routingKey, "\\.#")
		if strings.HasSuffix(routingKey, "\\.#") {
			continue
		}
		routingKey = routingKey + `(.*\.?)+`
	}
	routingKey = strings.Replace(routingKey, "\\.#\\.", `(.*\.?)+`, -1)
	routingKey = strings.Replace(routingKey, "#", `(.*\.?)+`, -1)
	pattern := "^" + routingKey + "$"

	return regexp.Compile(pattern)
}

func objSort(objs []interface{}) {
	sort.Slice(objs, func(i, j int) bool {
		return objs[i].(int) < objs[j].(int)
	})

}
func TestMatchTree2(t *testing.T) {
	type Node struct {
		regexp *regexp.Regexp
		key    string
		obj    interface{}
	}

	var nodes []Node

	appendNodes := func(expr string, obj interface{}) {
		req, err := buildRegexp(expr)
		if err != nil {
			t.Fatalf(err.Error())
		}
		nodes = append(nodes, Node{
			regexp: req,
			key:    expr,
			obj:    obj,
		})
	}
	tree := NewMatchTree()

	rand.Seed(time.Now().Unix())
	for i := 0; i < 2000; i++ {
		var key string
		for i := 0; i < 5; i++ {
			if len(key) != 0 {
				key += "."
			}
			if rand.Int31n(10) == 11 {
				key += "#"
			} else {
				str := uuid.New().String()[:1]
				key += str
			}
		}
		tree.Insert(key, i)
		appendNodes(key, i)
		//fmt.Println(key)
	}

	key := "c.c.c.c.5"
	begin := time.Now()
	count := 1000000
	for i := 0; i < count; i++ {
		objs := tree.Match(key)
		if objs == nil {
			//t.Fatalf("failed")
		} else {
			//objSort(objs)
			//fmt.Println(objs)
		}
	}
	fmt.Println(int(float64(count) / time.Now().Sub(begin).Seconds()))
	fmt.Println("Match use seconds", time.Now().Sub(begin).Seconds())

	begin = time.Now()
	if count > 50000 {
		count = 50000
	}
	for i := 0; i < count; i++ {
		var objs []interface{}
		for _, node := range nodes {
			if node.regexp.MatchString(key) {
				objs = append(objs, node.obj)
			}
		}
		if objs == nil {
			//t.Fatalf("failed")
		} else {
			//objSort(objs)
			//fmt.Println(objs)
		}
	}
	fmt.Println(float64(count) / time.Now().Sub(begin).Seconds())
	fmt.Println("use seconds", time.Now().Sub(begin).Seconds())
}

func TestMatchTree_MatchUniq(t *testing.T) {
	tree := NewMatchTree()

	tree.Insert("#.5.#.#.#", 1)
	tree.Insert("#.c.#.5.#", 2)

	res := tree.Match("c.c.c.c.5")
	if len(res) != 2 {
		t.Fatalf("%+v", res)
	}
}

func TestMatchTree_MatchUniq2(t *testing.T) {
	tree := NewMatchTree()

	tree.Insert("1.2.3.4.5", 1)
	tree.Insert("#.1.#.5.#", 2)
	tree.Insert("*.2.#.4.*", 3)
	tree.Insert("*.*.*.#", 4)
	tree.Insert("*.2.#.5", 5)
	tree.Insert("#.1.#.*", 6)
	tree.Insert("#.1.#.5", 7)
	tree.Insert("#.*", 8)
	tree.Insert("*.#.*", 9)
	tree.Insert("*.#.#.2.#", 10)
	tree.Insert("*.#.*.*.4.*", 11)
	tree.Insert("*.#.*.*.*.*.#", 12)
	tree.Insert("*.#.2.*.*.*.#", 13)
	tree.Insert("#", 14)
	tree.Insert("*.#", 15)
	tree.Insert("*.#.2.*.#", 16)

	res := sortIntObjs(tree.Match("1.2.3.4.5"))
	if len(res) != 16 {
		t.Fatalf("%+v", res)
	}
}

func BenchmarkMatchTree_MatchTokenUniq(b *testing.B) {
	rand.Seed(time.Now().Unix())
	tree := NewMatchTree()
	for i := 0; i < 2000; i++ {
		var key string
		for i := 0; i < 5; i++ {
			if len(key) != 0 {
				key += "."
			}
			if rand.Int31n(50) == 1 {
				key += "#"
			} else {
				str := uuid.New().String()[:1]
				key += str
			}
		}
		tree.Insert(key, i)
	}

	key := "c.c.c.c.5"
	begin := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	b.N = 1000000
	for i := 0; i < b.N; i++ {
		tree.Match(key)
	}
	fmt.Println(int(float64(b.N) / time.Now().Sub(begin).Seconds()))
}

func TestNextToken(t *testing.T) {
	var tokens []string
	for token, remain := nextToken("1.2.3.4.5.6"); len(token) != 0; token, remain = nextToken(remain) {
		tokens = append(tokens, token)
	}
	if reflect.DeepEqual(tokens, []string{"1", "2", "3", "4", "5", "6"}) == false {
		t.Fatalf("%+v", tokens)
	}
}

func TestCopyOnWrite(t *testing.T) {
	tree := NewMatchTree()

	tree.Insert("1.2.3.4.5", 1)
	tree.Insert("1.2.3.4.6", 2)

	if path, objs := treeWalk(tree); len(path) != 2 || len(objs) != 2 {
		t.Fatalf("error")
	}

	clone := tree.Clone()
	clone.Insert("1.2.3.4.5.6.7", 3)
	if path, objs := treeWalk(tree); len(path) != 2 || len(objs) != 2 {
		t.Fatalf("error")
	}
	if path, objs := treeWalk(clone); len(path) != 3 || len(objs) != 3 {
		t.Fatalf("error")
	}
}

func treeWalk(tree *MatchTree) ([]string, []interface{}) {
	var paths []string
	var allObjs []interface{}
	tree.Walk(func(path string, objs []interface{}) bool {
		paths = append(paths, path)
		allObjs = append(allObjs, objs...)
		return true
	})
	return paths, allObjs
}

func TestDelete(t *testing.T) {
	tree := NewMatchTree()

	tree.Insert("1.2.3.4.5", 1)

	tree.Insert("1.2.3.4.6", 2)

	tree2 := tree.Clone()

	tree.Delete("1.2.3.4.5", 1)
	tree.Delete("1.2.3.4.6", 2)

	path, objs := treeWalk(tree)
	if len(path) != 0 || len(objs) != 0 {
		t.Fatalf("delete error")
	}

	path, objs = treeWalk(tree2)
	if len(path) != 2 || len(objs) != 2 {
		t.Fatalf("clone error")
	}
}

func BenchmarkInsertCoW(b *testing.B) {

	rand.Seed(time.Now().Unix())
	tree := NewMatchTree()
	begin := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	b.N = 100000
	for i := 0; i < b.N; i++ {
		var key string
		for i := 0; i < 5; i++ {
			if len(key) != 0 {
				key += "."
			}
			if rand.Int31n(50) == 1 {
				key += "#"
			} else {
				str := uuid.New().String()[:1]
				key += str
			}
		}
		tree.Insert(key, i)
		tree = tree.Clone()
	}
	fmt.Println(int(float64(b.N) / time.Now().Sub(begin).Seconds()))
}
