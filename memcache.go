/*
 * Go memcache client package
 *
 * Author: Krzysztof Kliś <krzysztof.klis@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package memcache

import (
	"bufio";
	"bytes";
	"net";
	"os";
	"regexp";
	"strconv";
	"strings";
)

type Memcache struct {
	conn net.Conn;
}

func Connect(host string, port int) (*Memcache, os.Error) {
	memc := new(Memcache);
	addr := host + ":" + strconv.Itoa(port);
	conn, err := net.Dial("tcp", "", addr);
	if err != nil {
		return nil, err;
	}
	memc.conn = conn;
	return memc, nil;
}

func (memc *Memcache) Close() {
	if memc != nil && memc.conn != nil {
		memc.conn.Close();
	}
}

func (memc *Memcache) Get(key string) (value []byte, flags int) {
	if memc == nil || memc.conn == nil {
		return nil, 0;
	}
	cmd := "get " + key + "\r\n";
	n, err := memc.conn.Write(strings.Bytes(cmd));
	if err != nil || n != len(cmd) {
		return nil, 0;
	}
	line := make([]byte, 0);
	buf := make([]byte, 1);
	for {
		n, err = memc.conn.Read(buf);
		if err != nil || n != 1 {
			return nil, 0;
		}
		line = bytes.Add(line, buf);
		l := len(line);
		if l > 1 && line[l-2] == '\r' && line[l-1] == '\n' {
			break;
		}
	}
	re, _ := regexp.Compile("VALUE " + key + " ([0-9]+) ([0-9]+)");
	a := re.MatchStrings(string(line));
	if len(a) != 3 {
		return nil, 0;
	}
	flags, _ = strconv.Atoi(a[1]);
	l, _ := strconv.Atoi(a[2]);
	value = make([]byte, l);
	n, err = memc.conn.Read(value);
	if err != nil || n != l {
		return nil, 0;
	}
	buf = make([]byte, 7);
	n, err = memc.conn.Read(buf);
	if err != nil || string(buf) != "\r\nEND\r\n" {
		return nil, 0;
	}
	return value, flags;
}

func (memc *Memcache) Set(key string, value []byte, flags int, exptime int64) (bool) {
	if memc == nil || memc.conn == nil {
		return false;
	}
	l := len(value);
	cmd := "set " + key + " " + strconv.Itoa(flags) + " " + strconv.Itoa64(exptime) + " " + strconv.Itoa(l) + "\r\n";
	n, err := memc.conn.Write(strings.Bytes(cmd));
	if err != nil || n != len(cmd) {
		return false;
	}
	n, err = memc.conn.Write(value);
	if err != nil || n != l {
		return false;
	}
	n, err = memc.conn.Write(strings.Bytes("\r\n"));
	if err != nil || n != 2 {
		return false;
	}
	reader := bufio.NewReader(memc.conn);
	line, err := reader.ReadString('\n');
	if err != nil || line != "STORED\r\n" {
		return false;
	}
	return true;
}

func (memc *Memcache) Delete(key string) (bool) {
	if memc == nil || memc.conn == nil {
		return false;
	}
	cmd := "delete " + key + "\r\n";
	n, err := memc.conn.Write(strings.Bytes(cmd));
	if err != nil || n != len(cmd) {
		return false;
	}
	reader := bufio.NewReader(memc.conn);
	line, err := reader.ReadString('\n');
	if err != nil || line != "DELETED\r\n" {
		return false;
	}
	return true;
}

