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
	for i := 0; i < 20000; i++ {
		var key string
		for i := 0; i < 5; i++ {
			if len(key) != 0 {
				key += "."
			}
			if rand.Int31n(10) == 0 {
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
	count := 1
	for i := 0; i < count; i++ {
		objs := tree.MatchUniq(key)
		if objs == nil {
			t.Fatalf("failed")
		}
		objSort(objs)
		fmt.Println(objs)
	}
	fmt.Println(float64(count) / time.Now().Sub(begin).Seconds())

	begin = time.Now()
	for i := 0; i < count; i++ {
		var objs []interface{}
		for _, node := range nodes {
			if node.regexp.MatchString(key) {
				objs = append(objs, node.obj)
			}
		}
		if objs == nil {
			t.Fatalf("failed")
		} else {
			objSort(objs)
			fmt.Println(objs)
		}
	}
	fmt.Println(float64(count) / time.Now().Sub(begin).Seconds())
}

func TestMatchTree_MatchUniq(t *testing.T) {
	tree := NewMatchTree()

	tree.Insert("#.5.#.#.#", 1)
	tree.Insert("#.c.#.5.#", 2)

	res := tree.MatchUniq("c.c.c.c.5")
	if len(res) == 0 {
		t.Fatalf("failed")
	} else {
		fmt.Println(res)
	}
}
