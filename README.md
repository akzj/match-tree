# match-tree

amqp route key match trie tree
* base on trie tree
* support #,* match
* support Copy On Write ,lock free,update tree will no block Match()

```
tree := NewMatchTree()
tree.Insert("#.5.#.#.#", 1)
tree.Insert("#.c.#.5.#", 2)

res := tree.Match("c.c.c.c.5")
fmt.Println(res)
```
```
output :[1,2]
```
