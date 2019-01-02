package main

import (
	"fmt"
	"math/rand"
	"time"
)

func get_rand() uint64 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return (uint64(r1.Uint32())<<32 + uint64(r1.Uint32()))
}
func main() {
	for i := 0; i < 10; i++ {
		fmt.Println(get_rand())
	}
}
