/*
 * Go memcache client package
 *
 * Author: Krzysztof Kliś <krzysztof.klis@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
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
	"regexp"
	"strconv"
	"strings"
)

type Memcache struct {
	conn net.Conn
}

type Error struct {
	os.ErrorString
}
        
var (
	ConnectionError	os.Error = &Error{"memcache: not connected"}
	ReadError		os.Error = &Error{"memcache: read error"}
	WriteError		os.Error = &Error{"memcache: write error"}
	DeleteError		os.Error = &Error{"memcache: delete error"}
)

func Connect(host string, port int) (*Memcache, os.Error) {
	memc := new(Memcache)
	addr := host + ":" + strconv.Itoa(port)
	conn, err := net.Dial("tcp", "", addr)
	if err != nil {
		return nil, err
	}
	memc.conn = conn
	return memc, nil
}

func (memc *Memcache) Close() (err os.Error) {
	if memc == nil || memc.conn == nil {
		err = ConnectionError
		return
	}
	memc.conn.Close()
	return
}

func (memc *Memcache) Get(key string) (value []byte, flags int, err os.Error) {
	if memc == nil || memc.conn == nil {
		err = ConnectionError
		return
	}
	cmd := "get " + key + "\r\n"
	n, err := memc.conn.Write(strings.Bytes(cmd))
	if err != nil  {
		return
	}
	reader := bufio.NewReader(memc.conn)
	line, err := reader.ReadString('\n')
	re, _ := regexp.Compile("VALUE " + key + " ([0-9]+) ([0-9]+)")
	a := re.MatchStrings(line)
	if len(a) != 3 {
		err = ReadError
		return
	}
	flags, _ = strconv.Atoi(a[1])
	l, _ := strconv.Atoi(a[2])
	value = make([]byte, l)
	n, err = reader.Read(value)
	if err != nil {
		return
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

func (memc *Memcache) Set(key string, value []byte, flags int, exptime int64) (os.Error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	l := len(value)
	cmd := "set " + key + " " + strconv.Itoa(flags) + " " + strconv.Itoa64(exptime) + " " + strconv.Itoa(l) + "\r\n"
	_, err := memc.conn.Write(strings.Bytes(cmd))
	if err != nil {
		return err
	}
	_, err = memc.conn.Write(value)
	if err != nil {
		return err
	}
	_, err = memc.conn.Write(strings.Bytes("\r\n"))
	if err != nil {
		return err
	}
	reader := bufio.NewReader(memc.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if line != "STORED\r\n" {
		return WriteError
	}
	return nil
}

func (memc *Memcache) Delete(key string) (os.Error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	cmd := "delete " + key + "\r\n"
	_, err := memc.conn.Write(strings.Bytes(cmd))
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

