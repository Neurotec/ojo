// Package ojo notify to central
package ojo

import (
	"github.com/docker/libchan"
	"log"
)

type Feedback interface {
	libchan.Sender
}

type Event struct {
	RequestID    int64
	Data         interface{}
	RemoteSender libchan.Sender
}

type EventObservable struct {
	*Event
	Notification chan interface{}
}

// Feedback from central to node/observer
func (msg *Event) Feedback(data interface{}) {
	event := &Event{-1, data, msg.RemoteSender}
	err := msg.RemoteSender.Send(event)
	if err != nil {
		log.Fatal(err)
	}
}
