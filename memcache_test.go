package memcache

import (
	"strconv"
	"testing"
	"os"
)

var key string = "foo"
var value string = "bar"
var flags int = 1
var memc *Memcache

func TestSet(t *testing.T) {
	connect(t)
	err := memc.Set(key, []uint8(value), flags, 0)
	assertNoError(t, err)
	assertGet(t, value)
	cleanUp()
}

func TestAdd(t *testing.T) {
	connect(t)
	err := memc.Add(key, []uint8(value), flags, 0)
	assertNoError(t, err)
	assertGet(t, value)
	cleanUp()
}

func TestAddingPresentKey(t *testing.T) {
	connect(t)
	err := memc.Set(key, []uint8(value), flags, 0)
	assertNoError(t, err)
	assertGet(t, value)
	err = memc.Add(key, []uint8(value), flags, 0)
	if err == nil {
		t.Error("Adding should fail because key is already present")
	}
	cleanUp()
}


func TestReplace(t *testing.T) {
	connect(t)
	err := memc.Add(key, []uint8(value), flags, 0)
	assertNoError(t, err)
	newValue := "new value"
	err = memc.Replace(key, []uint8(newValue), flags, 0)
	assertNoError(t, err)
	assertGet(t, newValue)
	cleanUp()
}

func TestPrepend(t *testing.T) {
	connect(t)
	err := memc.Add(key, []uint8(value), flags, 0)
	assertNoError(t, err)
	prefix := "prefix"
	err = memc.Prepend(key, []uint8(prefix), flags, 0)
	assertNoError(t, err)
	assertGet(t, prefix+value)
	cleanUp()
}

func TestAppend(t *testing.T) {
	connect(t)
	err := memc.Add(key, []uint8(value), flags, 0)
	assertNoError(t, err)
	suffix := "suffix"
	err = memc.Append(key, []uint8(suffix), flags, 0)
	assertNoError(t, err)
	assertGet(t, value+suffix)
	cleanUp()
}

func TestDelete(t *testing.T) {
	connect(t)
	err := memc.Add(key, []uint8(value), flags, 0)
	assertNoError(t, err)
	err = memc.Delete(key)
	assertNoError(t, err)
	_, _, err = memc.Get(key)
	if err == nil {
		t.Error("Data not removed from memcache")
	}
	cleanUp()
}

func TestIncr(t *testing.T) {
	connect(t)
	err := memc.Add(key, []uint8("1234"), flags, 0)
	assertNoError(t, err)
	i, err := memc.Incr(key, 9)
	assertNoError(t, err)
	if i != 1243 {
		t.Error("Value expexcted: 1243\nValue received: " + strconv.Uitoa64(i))
	}
	cleanUp()
}

func TestDecr(t *testing.T) {
	connect(t)
	err := memc.Add(key, []uint8("1243"), flags, 0)
	assertNoError(t, err)
	i, err := memc.Decr(key, 9)
	assertNoError(t, err)
	if i != 1234 {
		t.Error("Value expexcted: 1234\nValue received: " + strconv.Uitoa64(i))
	}
	cleanUp()
}

func assertGet(t *testing.T, expectedValue string) {
	receivedValue, receivedFlags, err := memc.Get("foo")
	assertNoError(t, err)
	if string(receivedValue) != expectedValue {
		t.Error("Value expexcted: " + expectedValue + "\nValue received: " + string(receivedValue))
	}
	if receivedFlags != flags {
		t.Error("Flags expected: " + strconv.Itoa(flags) + "\nFlags received: " + strconv.Itoa(receivedFlags))
	}
}

func connect(t *testing.T) {
	connection, err := Connect("127.0.0.1", 11211)
	assertNoError(t, err)
	memc = connection
}

func cleanUp() {
	memc.Delete(key)
}

func assertNoError(t *testing.T, err os.Error) {
	if err != nil {
		t.Error(err.String())
	}
}
