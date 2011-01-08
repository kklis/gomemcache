package main

import (
	"github.com/kklis/gomemcache"
	"fmt"
)

func main() {
	memc, err := memcache.Connect("127.0.0.1", 11211)
	if err != nil {
		panic("Error: " + err.String())
	}
	err = memc.Set("foo", []uint8("bar"), 0, 0)
	if err != nil {
		panic("Error: " + err.String())
	}
	val, fl, _ := memc.Get("foo")
	fmt.Printf("%s %d\n", val, fl)
}
