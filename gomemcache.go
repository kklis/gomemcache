//
// Go memcachedb client package
//
// This is a memcachedb (https://github.com/stvchu/memcachedb) client package for the Go programming language. Originally based on kklis's memcache library https://github.com/kklis/gomemcache .
// This package adds a few commands for memcachedb.
//
// The way to comminucate between memcachedb is defined by a protocol document.
// https://github.com/stvchu/memcachedb/blob/master/doc/protocol.txt
//
// Author: Krzysztof Kliś <krzysztof.klis@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version with the following modification:
//
// As a special exception, the copyright holders of this library give you
// permission to link this library with independent modules to produce an
// executable, regardless of the license terms of these independent modules,
// and to copy and distribute the resulting executable under terms of your choice,
// provided that you also meet, for each linked independent module, the terms
// and conditions of the license of that module. An independent module is a
// module which is not derived from or based on this library. If you modify this
// library, you may extend this exception to your version of the library, but
// you are not obligated to do so. If you do not wish to do so, delete this
// exception statement from your version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
///
package gomemcachedb

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// memcachedb client object
type MemcacheDB struct {
	// persistent connection between memcachedb
	conn net.Conn
}

// fetched result
type Result struct {
	Value []uint8
	Flags int
}

var (
	ConnectionError = errors.New("memcachedb: not connected")
	ReadError       = errors.New("memcachedb: read error")
	DeleteError     = errors.New("memcachedb: delete error")
	FlushAllError   = errors.New("memcachedb: flush_all error")
	NotFoundError   = errors.New("memcachedb: not found")
)

// Connect create connection to memcachedb.
// When the port specifies as zero, using unix domain socket.
func Connect(host string, port int) (*MemcacheDB, error) {
	var network, addr string
	if port == 0 {
		network = "unix"
		addr = host
	} else {
		network = "tcp"
		addr = host + ":" + strconv.Itoa(port)
	}
	return Dial(network, addr)
}

// dial memcachedb
func Dial(network, addr string) (memc *MemcacheDB, err error) {
	memc = new(MemcacheDB)
	conn, err := net.Dial(network, addr)
	if err != nil {
		return
	}
	memc.conn = conn
	return
}

// close connection
func (memc *MemcacheDB) Close() (err error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	return memc.conn.Close()
}

// flushing items
func (memc *MemcacheDB) FlushAll() (err error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	cmd := "flush_all\r\n"
	_, err1 := memc.conn.Write([]uint8(cmd))
	if err1 != nil {
		err = err1
		return err
	}
	reader := bufio.NewReader(memc.conn)
	line, err1 := reader.ReadString('\n')
	if err1 != nil {
		err = err1
		return err
	}
	if line != "OK\r\n" {
		return FlushAllError
	}
	return nil
}

// Get results from database.
func (memc *MemcacheDB) Get(key string) (value []byte, flags int, err error) {
	if memc == nil || memc.conn == nil {
		err = ConnectionError
		return
	}
	cmd := "get " + key + "\r\n"
	_, err = memc.conn.Write([]uint8(cmd))
	if err != nil {
		return
	}
	reader := bufio.NewReader(memc.conn)
	return memc.readValue(reader, key)
}

// GetMulti retrives results by using multiple `get` request to memcachedb.
//
// TODO Consider using multiple keys in get request for decreasing cost of network.
// `get k1, k2 ...` commands are available in memcachedb.
func (memc *MemcacheDB) GetMulti(keys ...string) (results map[string]Result, err error) {
	results = map[string]Result{}
	for _, key := range keys {
		value, flags, err1 := memc.Get(key)
		if err1 == nil {
			results[key] = Result{Value: value, Flags: flags}
		} else if err1 != NotFoundError {
			err = err1
			return
		}
	}
	return
}

func (memc *MemcacheDB) readValue(reader *bufio.Reader, key string) (value []byte, flags int, err error) {
	line, err1 := reader.ReadString('\n')
	if err1 != nil {
		err = err1
		return
	}
	a := strings.Split(strings.TrimSpace(line), " ")
	if len(a) != 4 || a[0] != "VALUE" || a[1] != key {
		if line == "END\r\n" {
			err = NotFoundError
		} else {
			err = ReadError
		}
		return
	}
	flags, _ = strconv.Atoi(a[2])
	l, _ := strconv.Atoi(a[3])
	value = make([]byte, l)
	n := 0
	for {
		i, err1 := reader.Read(value[n:])
		if i == 0 && err == io.EOF {
			break
		}
		if err1 != nil {
			err = err1
			return
		}
		n += i
		if n >= l {
			break
		}
	}
	if n != l {
		err = ReadError
		return
	}
	line, err = reader.ReadString('\n')
	if err != nil {
		return
	}
	if line != "\r\n" {
		err = ReadError
		return
	}
	return
}

func (memc *MemcacheDB) store(cmd string, key string, value []byte, flags int, exptime int64) (err error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	l := len(value)
	s := cmd + " " + key + " " + strconv.Itoa(flags) + " " + strconv.FormatInt(exptime, 10) + " " + strconv.Itoa(l) + "\r\n"
	writer := bufio.NewWriter(memc.conn)
	_, err = writer.WriteString(s)
	if err != nil {
		return err
	}
	_, err = writer.Write(value)
	if err != nil {
		return err
	}
	_, err = writer.WriteString("\r\n")
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(memc.conn)
	line, err1 := reader.ReadString('\n')
	if err1 != nil {
		err = err1
		return err
	}
	if line != "STORED\r\n" {
		WriteError := errors.New("memcachedb: " + strings.TrimSpace(line))
		return WriteError
	}
	return nil
}

// "set" means "store this data".
func (memc *MemcacheDB) Set(key string, value []byte, flags int, exptime int64) (err error) {
	return memc.store("set", key, value, flags, exptime)
}

// "add" means "store this data, but only if the server *doesn't* already
// hold data for this key". (from protocol of memcachedb
func (memc *MemcacheDB) Add(key string, value []byte, flags int, exptime int64) (err error) {
	return memc.store("add", key, value, flags, exptime)
}

// "replace" means "store this data, but only if the server *does*
// already hold data for this key".
func (memc *MemcacheDB) Replace(key string, value []byte, flags int, exptime int64) (err error) {
	return memc.store("replace", key, value, flags, exptime)
}

// "append" means "add this data to an existing key after existing data".
func (memc *MemcacheDB) Append(key string, value []byte, flags int, exptime int64) (err error) {
	return memc.store("append", key, value, flags, exptime)
}

// "prepend" means "add this data to an existing key before existing data".
func (memc *MemcacheDB) Prepend(key string, value []byte, flags int, exptime int64) (err error) {
	return memc.store("prepend", key, value, flags, exptime)
}

// Delete item using `delete` command
func (memc *MemcacheDB) Delete(key string) (err error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	cmd := "delete " + key + "\r\n"
	_, err1 := memc.conn.Write([]uint8(cmd))
	if err1 != nil {
		err = err1
		return err
	}
	reader := bufio.NewReader(memc.conn)
	line, err1 := reader.ReadString('\n')
	if err1 != nil {
		err = err1
		return err
	}
	if line != "DELETED\r\n" {
		return DeleteError
	}
	return nil
}

func (memc *MemcacheDB) incdec(cmd string, key string, value uint64) (i uint64, err error) {
	if memc == nil || memc.conn == nil {
		err = ConnectionError
		return
	}
	s := cmd + " " + key + " " + strconv.FormatUint(value, 10) + "\r\n"
	_, err = memc.conn.Write([]uint8(s))
	if err != nil {
		return
	}
	reader := bufio.NewReader(memc.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	if line == "NOT_FOUND\r\n" {
		err = NotFoundError
		return
	}
	i, err = strconv.ParseUint(strings.TrimSpace(line), 10, 64)
	return
}

// Commands "incr" and "decr" are used to change data for some item
// in-place, incrementing or decrementing it. The data for the item is
// treated as decimal representation of a 64-bit unsigned integer. If the
// current data value does not conform to such a representation, the
// commands behave as if the value were 0. Also, the item must already
// exist for incr/decr to work; these commands won't pretend that a
// non-existent key exists with value 0; instead, they will fail.
//
// see detail: https://github.com/stvchu/memcachedb/blob/master/doc/protocol.txt
func (memc *MemcacheDB) Incr(key string, value uint64) (i uint64, err error) {
	i, err = memc.incdec("incr", key, value)
	return
}

// Decrement values.  Please see Incr godoc.
func (memc *MemcacheDB) Decr(key string, value uint64) (i uint64, err error) {
	i, err = memc.incdec("decr", key, value)
	return
}

// setting timeout for read operation
func (memc *MemcacheDB) SetReadTimeout(nsec int64) (err error) {
	return memc.conn.SetReadDeadline(time.Now().Add(time.Duration(nsec)))
}

// setting timeout for write operation
func (memc *MemcacheDB) SetWriteTimeout(nsec int64) (err error) {
	return memc.conn.SetWriteDeadline(time.Now().Add(time.Duration(nsec)))
}

// statistics of memcachedb
type Stat struct {
	Pid                  int     // Process id of this server process
	Uptime               int     // Number of seconds this server has been running
	Time                 int     // current UNIX time according to the server
	Version              string  // Version string of this server
	PointerSize          int     // Default size of pointers on the host OS (generally 32 or 64)
	RusageUser           float32 // Accumulated user time for this process (seconds:microseconds)
	RusageSystem         float32 // Accumulated system time for this process (seconds:microseconds)
	CurrItems            int     // Current number of items stored by the server
	TotalItems           int     // Total number of items stored by this server ever since it started
	Bytes                int     // Current number of bytes used by this server to store items
	CurrConnections      int     // Number of open connections
	TotalConnections     int     // Total number of connections opened since the server started running
	ConnectionStructures int     // Number of connection structures allocated by the server
	CmdGet               int     // Cumulative number of retrieval requests
	CmdSet               int     // Cumulative number of storage requests
	GetHits              int     // Number of keys that have been requested and found present
	GetMisses            int     // Number of items that have been requested and not found
	Evictions            int     // Number of valid items removed from cache to free memory for new items
	BytesRead            int     // Total number of bytes read by this server from network
	BytesWritten         int     // Total number of bytes sent by this server to network
	Threads              int     // Number of worker threads requested.
}

// Stats fetches statistics from memcachedb.
func (memc *MemcacheDB) Stats() (v map[string]interface{}, err error) {
	if memc == nil || memc.conn == nil {
		err = ConnectionError
		return
	}
	cmd := "stats\r\n"
	if _, err = memc.conn.Write([]uint8(cmd)); err != nil {
		return
	}
	reader := bufio.NewReader(memc.conn)
	return memc.readStats(reader)
}

// read stats response
func (memc *MemcacheDB) readStats(reader *bufio.Reader) (v map[string]interface{}, err error) {
	r := map[string]interface{}{}
	for {
		line, err1 := reader.ReadString('\n')
		if err1 != nil {
			err = err1
			return r, err
		}
		a := strings.Split(strings.TrimSpace(line), " ")
		if len(a) != 3 || a[0] != "STAT" {
			if line == "END\r\n" {
				err = NotFoundError
			} else {
				err = ReadError
			}
			return r, err
		}
		// FIXME not all fields are int. defined type by each fields.
		// using struct tag or else.
		r[a[1]], _ = strconv.Atoi(a[2])
	}
	return r, nil
}
