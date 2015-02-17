// Package ojo a proof concept network observer
package ojo

import (
	"crypto/tls"
	"github.com/docker/libchan/spdy"
	"log"
	"net"
)

// Handler for react at notifications
type Handler interface {
	OnNotify(observable interface{}, notifications chan interface{}, feedback Feedback)
}

// Server where listener client touchers
type Server struct {
	Addr    string
	Handler Handler
}

// ServerClose it's a message for notify that server goes down
type Down struct{}

// ListenAndServe listen on the TCP and call Handler to react a Touch
func ListenAndServe(addr string, handler Handler) error {
	server := &Server{Addr: addr, Handler: handler}
	return server.ListenAndServe()
}

// ListenAndServeTLS listen on the TCP crypto and call Handler to react a Touch
func ListenAndServeTLS(addr string, handler Handler, certFile string, keyFile string) error {
	server := &Server{Addr: addr, Handler: handler}
	return server.ListenAndServeTLS(certFile, keyFile)
}

// Server listen on TCP handle the connections
func (srv *Server) Serve(listener net.Listener) error {
	tl, err := spdy.NewTransportListener(listener, spdy.NoAuthenticator)
	if err != nil {
		return err
	}
	observers := make(map[int64]EventObservable)
	lock := make(chan struct{})

	for {
		t, err := tl.AcceptTransport()
		if err != nil {
			return err
		}

		go func() {
			for {
				receiver, err := t.WaitReceiveChannel()
				if err != nil {
					log.Println(err)
					break
				}

				go func() {
					for {
						event := &Event{}
						err := receiver.Receive(event)
						if err != nil {
							log.Println(err)
							break
						}
						if _, ok := observers[event.RequestID]; !ok {

							observers[event.RequestID] = EventObservable{
								event,
								make(chan interface{}),
							}
							lock <- struct{}{}

							go func() {
								<-lock
								srv.Handler.OnNotify(
									observers[event.RequestID].Data,
									observers[event.RequestID].Notification,
									observers[event.RequestID].RemoteSender,
								)
								close(observers[event.RequestID].Notification)
								delete(observers, event.RequestID)
							}()
						} else {
							observers[event.RequestID].Notification <- event.Data
						}
					}
				}()

			}
		}()
	}

	// Notify we go down
	for _, observable := range observers {
		observable.Notification <- Down{}
	}

	return nil
}

func (srv *Server) ListenAndServe() error {
	var listener net.Listener
	var err error

	listener, err = net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatal(err)
	}
	return srv.Serve(listener)
}

func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	var listener net.Listener
	var err error

	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{tlsCert},
	}

	listener, err = tls.Listen("tcp", srv.Addr, tlsConfig)
	if err != nil {
		return err
	}
	return srv.Serve(listener)
}
