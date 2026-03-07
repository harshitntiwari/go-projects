package main

import (
	"cache/internal"
	"fmt"
)

func main() {
	c := internal.NewCache[string, int](3)

	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)

	fmt.Println(c.Get("a"))

	c.Put("d", 4)

	fmt.Println(c.Get("b"))
}