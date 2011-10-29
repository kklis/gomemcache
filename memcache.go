/*
 * Go memcache client package
 *
 * Author: Krzysztof Kliś <krzysztof.klis@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version with the following modification:
 *
 * As a special exception, the copyright holders of this library give you
 * permission to link this library with independent modules to produce an
 * executable, regardless of the license terms of these independent modules,
 * and to copy and distribute the resulting executable under terms of your choice,
 * provided that you also meet, for each linked independent module, the terms
 * and conditions of the license of that module. An independent module is a
 * module which is not derived from or based on this library. If you modify this
 * library, you may extend this exception to your version of the library, but
 * you are not obligated to do so. If you do not wish to do so, delete this
 * exception statement from your version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package memcache

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
)

type Memcache struct {
	conn net.Conn
}

var (
	ConnectionError	os.Error = os.NewError("memcache: not connected")
	ReadError	os.Error = os.NewError("memcache: read error")
	DeleteError	os.Error = os.NewError("memcache: delete error")
	NotFoundError	os.Error = os.NewError("memcache: not found")
)

func Connect(host string, port int) (memc *Memcache, err os.Error) {
	memc = new(Memcache)
	addr := host + ":" + strconv.Itoa(port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	memc.conn = conn
	return
}

func (memc *Memcache) Close() (os.Error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	err := memc.conn.Close()
	return err
}

func (memc *Memcache) Get(key string) (value []byte, flags int, err os.Error) {
	if memc == nil || memc.conn == nil {
		err = ConnectionError
		return
	}
	cmd := "get " + key + "\r\n"
	_, err = memc.conn.Write([]uint8(cmd))
	if err != nil  {
		return
	}
	reader := bufio.NewReader(memc.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
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
		i, err := reader.Read(value[n:])
		if i == 0 && err == os.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
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
	line, err = reader.ReadString('\n')
	if err != nil {
		return
	}
	if line != "END\r\n" {
		err = ReadError
		return
	}
	return
}

func (memc *Memcache) store(cmd string,key string, value []byte, flags int, exptime int64) (os.Error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	l := len(value)
	s := cmd + " " + key + " " + strconv.Itoa(flags) + " " + strconv.Itoa64(exptime) + " " + strconv.Itoa(l) + "\r\n"
	writer := bufio.NewWriter(memc.conn)
	_, err := writer.WriteString(s)
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
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if line != "STORED\r\n" {
		WriteError := os.NewError("memcache: " + strings.TrimSpace(line))
		return WriteError
	}
	return nil
}

func (memc *Memcache) Set(key string, value []byte, flags int, exptime int64) (err os.Error) {
	err = memc.store("set", key, value, flags, exptime)
	return
}

func (memc *Memcache) Add(key string, value []byte, flags int, exptime int64) (err os.Error) {
	err = memc.store("add", key, value, flags, exptime)
	return
}

func (memc *Memcache) Replace(key string, value []byte, flags int, exptime int64) (err os.Error) {
	err = memc.store("replace", key, value, flags, exptime)
	return
}

func (memc *Memcache) Append(key string, value []byte, flags int, exptime int64) (err os.Error) {
	err = memc.store("append", key, value, flags, exptime)
	return
}

func (memc *Memcache) Prepend(key string, value []byte, flags int, exptime int64) (err os.Error) {
	err = memc.store("prepend", key, value, flags, exptime)
	return
}

func (memc *Memcache) Delete(key string) (os.Error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	cmd := "delete " + key + "\r\n"
	_, err := memc.conn.Write([]uint8(cmd))
	if err != nil {
		return err
	}
	reader := bufio.NewReader(memc.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if line != "DELETED\r\n"  {
		return DeleteError
	}
	return nil
}

func (memc *Memcache) incdec(cmd string, key string, value uint64) (i uint64, err os.Error) {
	if memc == nil || memc.conn == nil {
		err = ConnectionError
		return
	}
	s := cmd + " " + key + " " + strconv.Uitoa64(value) + "\r\n"
	_, err = memc.conn.Write([]uint8(s))
	if err != nil {
		return
	}
	reader := bufio.NewReader(memc.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	if line == "NOT_FOUND\r\n"  {
		err = NotFoundError
		return
	}
	i, err = strconv.Atoui64(strings.TrimSpace(line))
	return
}

func (memc *Memcache) Incr(key string, value uint64) (i uint64, err os.Error) {
	i, err = memc.incdec("incr", key, value)
	return
}

func (memc *Memcache) Decr(key string, value uint64) (i uint64, err os.Error) {
	i, err = memc.incdec("decr", key, value)
	return
}

func (memc *Memcache) SetReadTimeout(nsec int64) (err os.Error) {
	err = memc.conn.SetReadTimeout(nsec)
	return
}

func (memc *Memcache) SetWriteTimeout(nsec int64) (err os.Error) {
	err = memc.conn.SetWriteTimeout(nsec)
	return
}

