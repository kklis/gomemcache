package memcache
 
import (
	"strconv"
	"strings"
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
	err = memc.Add(key, strings.Bytes(value), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	testGet(t, memc, value)
	err = memc.Replace(key, strings.Bytes(value), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	testGet(t, memc, value)
	err = memc.Append(key, strings.Bytes(value), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	testGet(t, memc, value + value)
	err = memc.Prepend(key, strings.Bytes(value), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	testGet(t, memc, value + value + value)
	err = memc.Set(key, strings.Bytes(value), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	testGet(t, memc, value)
	err = memc.Delete("foo")
	if err != nil {
		t.Error(err.String())
	}
	_, _, err = memc.Get("foo")
	if err == nil {
		t.Error("Data not removed from memcache")
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

