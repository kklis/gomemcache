## Description
This is a [memcachedb](https://github.com/stvchu/memcachedb) client package for the Go programming language. Originally based on kklis's memcache library [kklis/gomemcache](https://github.com/kklis/gomemcache).

The following commands are implemented:

* get (single key)
* set, add, replace, append, prepend
* delete
* incr, decr

## Installation

```
go get github.com/suzuken/gomemcachedb
```

Depending on your environment configuration, you may need root (Linux) or administrator (Windows) access rights to run the above command.

## Testing

* Install gomemcachedb package (as described above).
* Start memcachedb at 127.0.0.1:21201 before running the test.
* On Unix start memcache socket listener: `memcachedb -f /tmp/sample.db -p 21201`
* Run command: `go test github.com/suzuken/gomemcachedb`

**Warning**: Test suite includes a test that flushes all memcache content.

**Note**: On systems that don't support Unix sockets (like Microsoft Windows) TestDial_UNIX will fail.

## Example usage

* Go to $GOPATH/src/github.com/suzuken/gomemcachedb/example/
* Compile example with: `go build example.go`
* Run the binary

## TODO

* support memcachedb private commands
