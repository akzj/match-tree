# match_tree

amqp route key match tree


```
tree := NewMatchTree()
tree.Insert("#.5.#.#.#", 1)
tree.Insert("#.c.#.5.#", 2)

res := tree.MatchUniq("c.c.c.c.5")
fmt.Println(res)
```
```
output :[1,2]
```
