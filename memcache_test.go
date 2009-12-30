package memcache
 
import (
	"strings";
	"testing";
)

var key string = "foo";
var value string = "bar123";
var flags int = 1;

func TestClient(t *testing.T) {
	memc, err := Connect("127.0.0.1", 11211);
	if err != nil {
		t.Error("Failed to connect to memcache: " + err.String());
	}
	stat := memc.Set(key, strings.Bytes(value), flags, 0);
	if stat == false {
		t.Error("Failed to store data in memcache");
	}
	val, fl := memc.Get("foo");
	if string(val) != value || fl != flags {
		t.Error("Failed to get data from memcache");
	}
	stat = memc.Delete("foo");
	if stat == false {
		t.Error("Failed to delete data from memcache");
	}
}

