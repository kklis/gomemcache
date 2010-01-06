package memcache
 
import (
	"strconv"
	"strings"
	"testing"
)

var key string = "foo"
var value string = "bar123"
var flags int = 1

func TestClient(t *testing.T) {
	memc, err := Connect("127.0.0.1", 11211)
	if err != nil {
		t.Error(err.String())
	}
	err = memc.Set(key, strings.Bytes(value), flags, 0)
	if err != nil {
		t.Error(err.String())
	}
	val, fl, err := memc.Get("foo")
	if err != nil {
		t.Error(err.String())
	}
	if string(val) != value {
		t.Error("Value stored: " + value + "\nValue received: " + string(val))
	}
	if fl != flags {
		t.Error("Flags stored: " + strconv.Itoa(flags) + "\nFlags received: " + strconv.Itoa(fl))
	}
	err = memc.Delete("foo")
	if err != nil {
		t.Error(err.String())
	}
}

