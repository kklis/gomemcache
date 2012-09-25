package main

import (
	"fmt"
	"github.com/kklis/gomemcache"
)

func main() {
	memc, err := gomemcache.Connect("127.0.0.1", 11211)
	if err != nil {
		panic(err)
	}
	err = memc.Set("foo", []uint8("bar"), 0, 0)
	if err != nil {
		panic(err)
	}
	val, fl, _ := memc.Get("foo")
	fmt.Printf("%s %d\n", val, fl)
}
