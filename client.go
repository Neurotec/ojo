// Package touch on network
package ojo

import (
	"crypto/tls"
	"errors"
	"github.com/docker/libchan"
	"github.com/docker/libchan/spdy"
	"log"
	"net"
)

var ErrNotHandler = errors.New("Not a handler")

var lastRequestID = int64(0)

// Handler react on feedback
type HandlerEvent interface {
	Feedback(interface{})
}

type Observable struct {
	sender       libchan.Sender
	remoteSender libchan.Sender
}

type ObservableClose struct{}

type Channel struct {
	transport libchan.Transport
}

func Dial(addr string) (*Channel, error) {
	var client net.Conn
	var err error

	client, err = net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	transport, err := spdy.NewClientTransport(client)
	if err != nil {
		return nil, err
	}

	return &Channel{transport}, nil
}

func DialTLS(addr string) (*Channel, error) {
	var client net.Conn
	var err error

	client, err = tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	transport, err := spdy.NewClientTransport(client)
	if err != nil {
		return nil, err
	}

	return &Channel{transport}, nil
}

func (channel *Channel) Observe(data interface{}) (*Observable, error) {
	var err error

	if _, ok := data.(HandlerEvent); !ok {
		return nil, ErrNotHandler
	}

	sender, err := channel.transport.NewSendChannel()
	if err != nil {
		return nil, err
	}

	receiver, remoteSender := libchan.Pipe()

	if err := sender.Send(&Event{lastRequestID, data, remoteSender}); err != nil {
		return nil, err
	}

	lastRequestID += 1
	if lastRequestID == ^int64(0) {
		lastRequestID = 1
	}

	go func() {
		for {
			event := &Event{}
			err := receiver.Receive(event)
			if err != nil {
				log.Println(err)
				break
			}
			if data != nil {
				data.(HandlerEvent).Feedback(event.Data)
			}

		}
	}()

	return &Observable{sender, remoteSender}, nil
}

func (obs *Observable) Notify(data interface{}) {
	event := &Event{lastRequestID, data, obs.remoteSender}
	obs.sender.Send(event)
}

func (obs *Observable) Close() {
	obs.sender.Send(&Event{0, &ObservableClose{}, nil})
	obs.sender.Close()
}
