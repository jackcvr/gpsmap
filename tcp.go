package main

import (
	"encoding/binary"
	"github.com/goccy/go-json"
	"github.com/jackcvr/gpsmap/orm"
	"gorm.io/gorm"
	"io"
	"log"
	"net"
	"time"
)

const BufferSize = 1024

type Payload struct {
	State struct {
		Reported struct {
			Evt int `json:"evt"`
		} `json:"reported"`
	} `json:"state"`
}

type Values = map[int]string

type EvtInfo struct {
	Name   string
	Values Values
}

var EvtMap = map[int]EvtInfo{
	239: {"Ignition", Values{0: "Off", 1: "On"}},
	175: {"Auto Geofence", Values{0: "target left zone", 1: "target entered zone"}},
	252: {"Unplug", Values{0: "battery present", 1: "battery unplugged"}},
}

func ServeTCP(config Config, db *gorm.DB, pubsub *PubSub, bot *TGBot) {
	var err error
	var ln net.Listener
	log.Printf("TCP listening on %s", config.GPRS.Bind)
	ln, err = net.Listen("tcp", config.GPRS.Bind)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	var timer *time.Timer
	var recvTimeout time.Duration
	if bot != nil && config.TGBot.Timeout > 0 {
		recvTimeout = time.Second * time.Duration(config.TGBot.Timeout)
		timer = time.NewTimer(recvTimeout)
		defer timer.Stop()
		go func() {
			for {
				<-timer.C
				timer.Reset(recvTimeout)
				if err = bot.NotifyUsers("Backend have not received a message in time! Check your car!"); err != nil {
					errLog.Print(err)
				} else {
					log.Printf("timeout occured: users were notified")
				}
			}
		}()
		log.Printf("monitor started with timeout: %s", recvTimeout)
	}

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
					log.Printf("received: %s", payload)
					r := orm.Record{
						Imei:    imei,
						Payload: payload,
					}
					go func(r orm.Record) {
						if err := db.Create(&r).Error; err != nil {
							errLog.Print(err)
						}
						pubsub.Publish(r)
					}(r)
					if timer != nil {
						timer.Reset(recvTimeout)
						go func() {
							var p Payload
							if err = json.Unmarshal([]byte(payload), &p); err != nil {
								errLog.Print(err)
							} else if evt, ok := EvtMap[p.State.Reported.Evt]; ok {
								if err = bot.NotifyUsers(evt.Name); err != nil {
									errLog.Print(err)
								}
							}
						}()
					}
				}
			}
		}(conn)
	}
}
