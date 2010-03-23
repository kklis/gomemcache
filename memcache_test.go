package memcache
 
import (
	"strconv"
	"testing"
)

var key string = "foo"
var value string = "bar"
var flags int = 1

func TestClient(t *testing.T) {
	memc, err := Connect("127.0.0.1", 11211)
	if err != nil {
		t.Error(err.String())
	}
	// clean
	memc.Delete("foo")
	// test add
	err = memc.Add(key, []uint8(value), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	testGet(t, memc, value)
	// test replace
	err = memc.Replace(key, []uint8(value), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	testGet(t, memc, value)
	// test append
	err = memc.Append(key, []uint8(value), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	testGet(t, memc, value + value)
	// test prepend
	err = memc.Prepend(key, []uint8(value), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	testGet(t, memc, value + value + value)
	// test delete
	err = memc.Delete("foo")
	if err != nil {
		t.Error(err.String())
	}
	_, _, err = memc.Get("foo")
	if err == nil {
		t.Error("Data not removed from memcache")
	}
	// test incr
	err = memc.Set(key, []uint8("1234"), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	i, err := memc.Incr(key, 9)
	if err != nil {
		t.Error(err.String())
	}
	if i != 1243 {
		t.Error("Value expexcted: 1243\nValue received: " + strconv.Uitoa64(i))
	}
	// test decr
	i, err = memc.Decr(key, 9)
	if err != nil {
		t.Error(err.String())
	}
	if i != 1234 {
		t.Error("Value expexcted: 1234\nValue received: " + strconv.Uitoa64(i))
	}
}

func testGet(t *testing.T, memc *Memcache, s string) {
	val, fl, err := memc.Get("foo")
	if err != nil {
		t.Error(err.String())
	}
	if string(val) != s {
		t.Error("Value expexcted: " + s + "\nValue received: " + string(val))
	}
	if fl != flags {
		t.Error("Flags expected: " + strconv.Itoa(flags) + "\nFlags received: " + strconv.Itoa(fl))
	}
}

