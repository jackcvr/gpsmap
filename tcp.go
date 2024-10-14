package main

import (
	"encoding/binary"
	"github.com/jackcvr/gpstrack/orm"
	"io"
	"log"
	"net"
)

func StartTCPServer(addr string) {
	var err error
	var ln net.Listener
	ln, err = net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	var conn net.Conn
	for {
		conn, err = ln.Accept()
		if err != nil {
			errLog.Print(err)
			continue
		}
		log.Printf("accepted: %s", conn.RemoteAddr())
		go func(c net.Conn) {
			defer c.Close()
			defer func() {
				if r := recover(); r != nil {
					errLog.Print(r)
				}
			}()
			buf := make([]byte, BufferSize)
			var n int
			var imei string
			if n, err = c.Read(buf); err != nil {
				errLog.Print(err)
				return
			} else {
				size := binary.BigEndian.Uint16(buf[:2])
				imei = string(buf[2:size])
			}
			log.Printf("IMEI: %s", imei)
			for {
				if n, err = c.Read(buf); err != nil {
					if err != io.EOF {
						errLog.Print(err)
					}
					return
				} else {
					payload := string(buf[:n])
					log.Print(payload)
					r := orm.Record{
						Imei:    imei,
						Payload: payload,
					}
					go func(r orm.Record) {
						if err := orm.Client.Create(&r).Error; err != nil {
							errLog.Print(err)
						}
					}(r)
				}
			}
		}(conn)
	}
}
