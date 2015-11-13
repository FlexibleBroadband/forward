// Copyright © 2015 FlexibleBroadband Team.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//	      ___ _           _ _     _
//	     / __\ | _____  _(_) |__ | | ___
//	    / _\ | |/ _ \ \/ / | '_ \| |/ _ \
//	   / /   | |  __/>  <| | |_) | |  __/
//	   \/    |_|\___/_/\_\_|_.__/|_|\___|

package forward

import (
	"fmt"
	"io"
	"log"
	"net"
)

const (
	MaxUDPLen = 4096 * 2
)

func Forward(network, laddr, daddr string) error {
	switch network {
	case "tcp":
		return forwardTCP(laddr, daddr)
	case "udp":
		return forwardUDP(laddr, daddr)
	default:
		return fmt.Errorf("not support network:%v", network)
	}
}

// 			0        |              |
// client addr length|  client addr |  playload
func forwardUDP(laddr, daddr string) error {
	dst, err := net.ResolveUDPAddr("udp", daddr)
	if err != nil {
		return err
	}
	listenAddr, err := net.ResolveUDPAddr("udp", laddr)
	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return err
	}
	buf := make([]byte, MaxUDPLen)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		// forward buf data.
		if addr.String() != dst.String() {
			// forward to client.
			toAddr, err := net.ResolveUDPAddr("udp", string(buf[1:int(buf[0])+1]))
			if err != nil {
				log.Println("read client dst addr error:", err)
				continue
			}
			if _, err = conn.WriteTo(buf[1+int(buf[0]):n], toAddr); err != nil {
				log.Println("write udp to client error:", err)
				return err
			}
		} else {
			// forward to tunnel.
			if _, err = conn.WriteTo(buf[:n], dst); err != nil {
				log.Println("write udp to client error:", err)
				return err
			}
		}
	}
}

func forwardTCP(laddr, daddr string) error {
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		return nil
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			dstConn, err := net.Dial("tcp", daddr)
			if err != nil {
				log.Println("dial tcp connection to remote server error:", err)
				return
			}
			go io.Copy(dstConn, c)
			io.Copy(c, dstConn)
		}(conn)
	}
	return nil
}
