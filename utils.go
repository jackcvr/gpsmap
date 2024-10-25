package main

import (
	"github.com/jackcvr/gpsmap/orm"
	"net"
	"sync"
	"time"
)

type PubSub struct {
	mu          sync.Mutex
	Subscribers map[net.Conn]chan orm.Record
}

func NewPubSub() *PubSub {
	return &PubSub{
		Subscribers: make(map[net.Conn]chan orm.Record),
	}
}

func (ps *PubSub) Subscribe(conn net.Conn) chan orm.Record {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	sub := make(chan orm.Record, 1)
	ps.Subscribers[conn] = sub
	return sub
}

func (ps *PubSub) Unsubscribe(conn net.Conn) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	sub := ps.Subscribers[conn]
	close(sub)
	delete(ps.Subscribers, conn)
}

func (ps *PubSub) Publish(r orm.Record) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, sub := range ps.Subscribers {
		sub <- r
	}
}

func RunPeriodic(d time.Duration, f func()) {
	t := time.NewTicker(d)
	for {
		<-t.C
		f()
	}
}
