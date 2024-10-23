package main

import (
	"github.com/jackcvr/gpsmap/orm"
	"net"
	"sync"
)

type PubSub struct {
	Mux         sync.Mutex
	Subscribers map[net.Conn]chan orm.Record
}

func NewPubSub() *PubSub {
	return &PubSub{
		Subscribers: make(map[net.Conn]chan orm.Record),
	}
}

func (ps *PubSub) Subscribe(conn net.Conn) chan orm.Record {
	ps.Mux.Lock()
	defer ps.Mux.Unlock()
	recv := make(chan orm.Record, 1)
	ps.Subscribers[conn] = recv
	return recv
}

func (ps *PubSub) Unsubscribe(conn net.Conn) {
	ps.Mux.Lock()
	defer ps.Mux.Unlock()
	sub := ps.Subscribers[conn]
	close(sub)
	delete(ps.Subscribers, conn)
}

func (ps *PubSub) Publish(r orm.Record) {
	for _, sub := range ps.Subscribers {
		sub <- r
	}
}
